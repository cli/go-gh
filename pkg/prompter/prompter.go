// Package prompter provides various methods for prompting the user with
// questions for input.
package prompter

import (
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh/v2/pkg/text"
)

// Prompter provides methods for prompting the user.
type Prompter struct {
	stdin  FileReader
	stdout FileWriter
	stderr FileWriter
}

// FileWriter provides a minimal writable interface for stdout and stderr.
type FileWriter interface {
	io.Writer
	Fd() uintptr
}

// FileReader provides a minimal readable interface for stdin.
type FileReader interface {
	io.Reader
	Fd() uintptr
}

// New instantiates a new Prompter.
func New(stdin FileReader, stdout FileWriter, stderr FileWriter) *Prompter {
	return &Prompter{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}
}

// Select prompts the user to select an option from a list of options.
func (p *Prompter) Select(prompt, defaultValue string, options []string) (int, error) {
	var result int
	q := &survey.Select{
		Message:  prompt,
		Options:  options,
		PageSize: 20,
		Filter:   latinMatchingFilter,
	}
	if defaultValue != "" {
		for _, o := range options {
			if o == defaultValue {
				q.Default = defaultValue
				break
			}
		}
	}
	err := p.ask(q, &result)
	return result, err
}

// MultiSelect prompts the user to select multiple options from a list of options.
func (p *Prompter) MultiSelect(prompt string, defaultValues, options []string) ([]int, error) {
	var result []int
	q := &survey.MultiSelect{
		Message:  prompt,
		Options:  options,
		PageSize: 20,
		Filter:   latinMatchingFilter,
	}
	if len(defaultValues) > 0 {
		validatedDefault := []string{}
		for _, x := range defaultValues {
			for _, y := range options {
				if x == y {
					validatedDefault = append(validatedDefault, x)
				}
			}
		}
		q.Default = validatedDefault
	}
	err := p.ask(q, &result)
	return result, err
}

// Input prompts the user to input a single-line string.
func (p *Prompter) Input(prompt, defaultValue string) (string, error) {
	var result string
	err := p.ask(&survey.Input{
		Message: prompt,
		Default: defaultValue,
	}, &result)
	return result, err
}

// Password prompts the user to input a single-line string without echoing the input.
func (p *Prompter) Password(prompt string) (string, error) {
	var result string
	err := p.ask(&survey.Password{
		Message: prompt,
	}, &result)
	return result, err
}

// Confirm prompts the user to confirm a yes/no question.
func (p *Prompter) Confirm(prompt string, defaultValue bool) (bool, error) {
	var result bool
	err := p.ask(&survey.Confirm{
		Message: prompt,
		Default: defaultValue,
	}, &result)
	return result, err
}

func (p *Prompter) ask(q survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	opts = append(opts, survey.WithStdio(p.stdin, p.stdout, p.stderr))
	err := survey.AskOne(q, response, opts...)
	if err == nil {
		return nil
	}
	return fmt.Errorf("could not prompt: %w", err)
}

// latinMatchingFilter returns whether the value matches the input filter.
// The strings are compared normalized in case.
// The filter's diactritics are kept as-is, but the value's are normalized,
// so that a missing diactritic in the filter still returns a result.
func latinMatchingFilter(filter, value string, index int) bool {
	filter = strings.ToLower(filter)
	value = strings.ToLower(value)
	// include this option if it matches.
	return strings.Contains(value, filter) || strings.Contains(text.RemoveDiacritics(value), filter)
}
