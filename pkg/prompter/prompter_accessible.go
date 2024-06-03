package prompter

import (
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type AccessiblePrompter struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewAccessiblePrompter(stdin io.Reader, stdout io.Writer, stderr io.Writer) *AccessiblePrompter {
	return &AccessiblePrompter{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *AccessiblePrompter) newForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).
		WithTheme(huh.ThemeBase16()).
		WithAccessible(os.Getenv("ACCESSIBLE") != "").
		WithProgramOptions(tea.WithOutput(p.stdout), tea.WithInput(p.stdin))
}

// Select prompts the user to select an option from a list of options.
func (p *AccessiblePrompter) Select(prompt, defaultValue string, options []string) (int, error) {
	var result int
	formOptions := []huh.Option[int]{}
	for i, o := range options {
		formOptions = append(formOptions, huh.NewOption(o, i))

		if o == defaultValue {
			result = i
		}
	}

	form := p.newForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title(prompt).
				Value(&result).
				Options(formOptions...),
		),
	)

	err := form.Run()
	return result, err
}

// MultiSelect prompts the user to select multiple options from a list of options.
func (p *AccessiblePrompter) MultiSelect(prompt string, defaultValues, options []string) ([]int, error) {
	var result []int
	formOptions := []huh.Option[int]{}
	for i, o := range options {
		formOptions = append(formOptions, huh.NewOption(o, i))

		for _, d := range defaultValues {
			if d == o {
				result = append(result, i)
			}
		}
	}

	form := p.newForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title(prompt).
				Value(&result).
				Options(formOptions...),
		),
	)

	err := form.Run()
	return result, err
}

// Input prompts the user to input a single-line string.
func (p *AccessiblePrompter) Input(prompt, defaultValue string) (string, error) {
	result := defaultValue
	form := p.newForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Value(&result),
		),
	)

	err := form.Run()
	return result, err
}

// Password prompts the user to input a single-line string without echoing the input.
func (p *AccessiblePrompter) Password(prompt string) (string, error) {
	var result string
	form := p.newForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Value(&result).
				EchoMode(huh.EchoModePassword),
		),
	)

	err := form.Run()
	return result, err
}

// Confirm prompts the user to confirm a yes/no question.
func (p *AccessiblePrompter) Confirm(prompt string, defaultValue bool) (bool, error) {
	result := defaultValue
	form := p.newForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Value(&result),
		),
	)
	err := form.Run()
	return result, err
}
