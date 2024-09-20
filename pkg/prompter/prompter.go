// Package prompter provides various methods for prompting the user with
// questions for input.
package prompter

import (
	"io"
	"os"
)

// Prompter provides methods for prompting the user.
type Prompter struct {
	pClient PrompterClient
}

type PrompterClient interface {
	Select(prompt, defaultValue string, options []string) (int, error)
	MultiSelect(prompt string, defaultValues, options []string) ([]int, error)
	Input(prompt, defaultValue string) (string, error)
	Password(prompt string) (string, error)
	Confirm(prompt string, defaultValue bool) (bool, error)
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
	// TODO: Enhance logic to look at configuration for prompter type.
	prompterType := os.Getenv("GH_PROMPTER")
	if prompterType == "accessible" {
		return &Prompter{
			pClient: NewAccessiblePrompter(stdin, stdout, stderr),
		}
	}

	return &Prompter{
		pClient: NewLegacyPrompter(stdin, stdout, stderr),
	}
}

// Select prompts the user to select an option from a list of options.
func (p *Prompter) Select(prompt, defaultValue string, options []string) (int, error) {
	return p.pClient.Select(prompt, defaultValue, options)
}

// MultiSelect prompts the user to select multiple options from a list of options.
func (p *Prompter) MultiSelect(prompt string, defaultValues, options []string) ([]int, error) {
	return p.pClient.MultiSelect(prompt, defaultValues, options)
}

// Input prompts the user to input a single-line string.
func (p *Prompter) Input(prompt, defaultValue string) (string, error) {
	return p.pClient.Input(prompt, defaultValue)
}

// Password prompts the user to input a single-line string without echoing the input.
func (p *Prompter) Password(prompt string) (string, error) {
	return p.pClient.Password(prompt)
}

// Confirm prompts the user to confirm a yes/no question.
func (p *Prompter) Confirm(prompt string, defaultValue bool) (bool, error) {
	return p.pClient.Confirm(prompt, defaultValue)
}
