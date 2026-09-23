package term

import (
	"io"
	"strings"
)

// TestTerminal is a configurable terminal implementation for tests.
//
// Its zero value is useful for non-interactive tests. Set the exported fields
// when a test needs to model terminal capabilities or a particular size.
type TestTerminal struct {
	Input         io.Reader
	Output        io.Writer
	ErrorOutput   io.Writer
	TTY           bool
	Color         bool
	Color256      bool
	TrueColor     bool
	Width         int
	Height        int
	SizeError     error
	TerminalTheme string
}

var _ Terminal = TestTerminal{}

func (t TestTerminal) In() io.Reader {
	if t.Input == nil {
		return strings.NewReader("")
	}
	return t.Input
}

func (t TestTerminal) Out() io.Writer {
	return t.Output
}

func (t TestTerminal) ErrOut() io.Writer {
	return t.ErrorOutput
}

func (t TestTerminal) IsTerminalOutput() bool {
	return t.TTY
}

func (t TestTerminal) IsColorEnabled() bool {
	return t.Color
}

func (t TestTerminal) Is256ColorSupported() bool {
	return t.Color256
}

func (t TestTerminal) IsTrueColorSupported() bool {
	return t.TrueColor
}

func (t TestTerminal) Size() (int, int, error) {
	return t.Width, t.Height, t.SizeError
}

func (t TestTerminal) Theme() string {
	return t.TerminalTheme
}
