package prompter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// PrompterMock provides stubbed out methods for prompting the user for
// use in tests. PrompterMock has a superset of the methods on Prompter
// so they both can satisfy the same interface.
//
// A basic example of how PrompterMock can be used:
//
//	type ConfirmPrompter interface {
//		Confirm(string, bool) (bool, error)
//	}
//
//	func PlayGame(prompter ConfirmPrompter) (int, error) {
//		confirm, err := prompter.Confirm("Shall we play a game", true)
//		if err != nil {
//			return 0, err
//		}
//		if confirm {
//			return 1, nil
//		}
//		return 2, nil
//	}
//
//	func TestPlayGame(t *testing.T) {
//		expectedOutcome := 1
//		mock := NewMock(t)
//		mock.RegisterConfirm("Shall we play a game", func(prompt string, defaultValue bool) (bool, error) {
//			return true, nil
//		})
//		outcome, err := PlayGame(mock)
//		if err != nil {
//			t.Fatalf("unexpected error: %v", err)
//		}
//		if outcome != expectedOutcome {
//			t.Errorf("expected %q, got %q", expectedOutcome, outcome)
//		}
//	}
type PrompterMock struct {
	t                *testing.T
	selectStubs      []selectStub
	multiSelectStubs []multiSelectStub
	inputStubs       []inputStub
	passwordStubs    []passwordStub
	confirmStubs     []confirmStub
}

type selectStub struct {
	prompt          string
	expectedOptions []string
	fn              func(string, string, []string) (int, error)
}

type multiSelectStub struct {
	prompt          string
	expectedOptions []string
	fn              func(string, []string, []string) ([]int, error)
}

type inputStub struct {
	prompt string
	fn     func(string, string) (string, error)
}

type passwordStub struct {
	prompt string
	fn     func(string) (string, error)
}

type confirmStub struct {
	Prompt string
	Fn     func(string, bool) (bool, error)
}

// NewMock instantiates a new PrompterMock.
func NewMock(t *testing.T) *PrompterMock {
	m := &PrompterMock{
		t:                t,
		selectStubs:      []selectStub{},
		multiSelectStubs: []multiSelectStub{},
		inputStubs:       []inputStub{},
		passwordStubs:    []passwordStub{},
		confirmStubs:     []confirmStub{},
	}
	t.Cleanup(m.verify)
	return m
}

// Select prompts the user to select an option from a list of options.
func (m *PrompterMock) Select(prompt, defaultValue string, options []string) (int, error) {
	var s selectStub
	if len(m.selectStubs) == 0 {
		return -1, noSuchPromptErr(prompt)
	}
	s = m.selectStubs[0]
	m.selectStubs = m.selectStubs[1:len(m.selectStubs)]
	if s.prompt != prompt {
		return -1, noSuchPromptErr(prompt)
	}
	assertOptions(m.t, s.expectedOptions, options)
	return s.fn(prompt, defaultValue, options)
}

// MultiSelect prompts the user to select multiple options from a list of options.
func (m *PrompterMock) MultiSelect(prompt string, defaultValues, options []string) ([]int, error) {
	var s multiSelectStub
	if len(m.multiSelectStubs) == 0 {
		return []int{}, noSuchPromptErr(prompt)
	}
	s = m.multiSelectStubs[0]
	m.multiSelectStubs = m.multiSelectStubs[1:len(m.multiSelectStubs)]
	if s.prompt != prompt {
		return []int{}, noSuchPromptErr(prompt)
	}
	assertOptions(m.t, s.expectedOptions, options)
	return s.fn(prompt, defaultValues, options)
}

// Input prompts the user to input a single-line string.
func (m *PrompterMock) Input(prompt, defaultValue string) (string, error) {
	var s inputStub
	if len(m.inputStubs) == 0 {
		return "", noSuchPromptErr(prompt)
	}
	s = m.inputStubs[0]
	m.inputStubs = m.inputStubs[1:len(m.inputStubs)]
	if s.prompt != prompt {
		return "", noSuchPromptErr(prompt)
	}
	return s.fn(prompt, defaultValue)
}

// Password prompts the user to input a single-line string without echoing the input.
func (m *PrompterMock) Password(prompt string) (string, error) {
	var s passwordStub
	if len(m.passwordStubs) == 0 {
		return "", noSuchPromptErr(prompt)
	}
	s = m.passwordStubs[0]
	m.passwordStubs = m.passwordStubs[1:len(m.passwordStubs)]
	if s.prompt != prompt {
		return "", noSuchPromptErr(prompt)
	}
	return s.fn(prompt)
}

// Confirm prompts the user to confirm a yes/no question.
func (m *PrompterMock) Confirm(prompt string, defaultValue bool) (bool, error) {
	var s confirmStub
	if len(m.confirmStubs) == 0 {
		return false, noSuchPromptErr(prompt)
	}
	s = m.confirmStubs[0]
	m.confirmStubs = m.confirmStubs[1:len(m.confirmStubs)]
	if s.Prompt != prompt {
		return false, noSuchPromptErr(prompt)
	}
	return s.Fn(prompt, defaultValue)
}

// RegisterSelect records that a Select prompt should be called.
func (m *PrompterMock) RegisterSelect(prompt string, opts []string, stub func(_, _ string, _ []string) (int, error)) {
	m.selectStubs = append(m.selectStubs, selectStub{
		prompt:          prompt,
		expectedOptions: opts,
		fn:              stub})
}

// RegisterMultiSelect records that a MultiSelect prompt should be called.
func (m *PrompterMock) RegisterMultiSelect(prompt string, d, opts []string, stub func(_ string, _, _ []string) ([]int, error)) {
	m.multiSelectStubs = append(m.multiSelectStubs, multiSelectStub{
		prompt:          prompt,
		expectedOptions: opts,
		fn:              stub})
}

// RegisterInput records that an Input prompt should be called.
func (m *PrompterMock) RegisterInput(prompt string, stub func(_, _ string) (string, error)) {
	m.inputStubs = append(m.inputStubs, inputStub{prompt: prompt, fn: stub})
}

// RegisterPassword records that a Password prompt should be called.
func (m *PrompterMock) RegisterPassword(prompt string, stub func(string) (string, error)) {
	m.passwordStubs = append(m.passwordStubs, passwordStub{prompt: prompt, fn: stub})
}

// RegisterConfirm records that a Confirm prompt should be called.
func (m *PrompterMock) RegisterConfirm(prompt string, stub func(_ string, _ bool) (bool, error)) {
	m.confirmStubs = append(m.confirmStubs, confirmStub{Prompt: prompt, Fn: stub})
}

func (m *PrompterMock) verify() {
	errs := []string{}
	if len(m.selectStubs) > 0 {
		errs = append(errs, "MultiSelect")
	}
	if len(m.multiSelectStubs) > 0 {
		errs = append(errs, "Select")
	}
	if len(m.inputStubs) > 0 {
		errs = append(errs, "Input")
	}
	if len(m.passwordStubs) > 0 {
		errs = append(errs, "Password")
	}
	if len(m.confirmStubs) > 0 {
		errs = append(errs, "Confirm")
	}
	if len(errs) > 0 {
		m.t.Helper()
		m.t.Errorf("%d unmatched calls to %s", len(errs), strings.Join(errs, ","))
	}
}

func noSuchPromptErr(prompt string) error {
	return fmt.Errorf("no such prompt '%s'", prompt)
}

func assertOptions(t *testing.T, expected, actual []string) {
	assert.Equal(t, expected, actual)
}
