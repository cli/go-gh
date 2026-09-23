package term

import (
	"bytes"
	"errors"
	"testing"
)

func TestTestTerminal(t *testing.T) {
	input := bytes.NewBufferString("input")
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	sizeErr := errors.New("size unavailable")
	terminal := TestTerminal{
		Input:         input,
		Output:        output,
		ErrorOutput:   errOutput,
		TTY:           true,
		Color:         true,
		Color256:      true,
		TrueColor:     true,
		Width:         120,
		Height:        40,
		SizeError:     sizeErr,
		TerminalTheme: "dark",
	}

	if got := terminal.In(); got != input {
		t.Fatalf("expected input reader to be preserved")
	}
	if got := terminal.Out(); got != output {
		t.Fatalf("expected output writer to be preserved")
	}
	if got := terminal.ErrOut(); got != errOutput {
		t.Fatalf("expected error writer to be preserved")
	}
	if !terminal.IsTerminalOutput() || !terminal.IsColorEnabled() || !terminal.Is256ColorSupported() || !terminal.IsTrueColorSupported() {
		t.Fatalf("expected configured terminal capabilities")
	}
	if width, height, err := terminal.Size(); width != 120 || height != 40 || !errors.Is(err, sizeErr) {
		t.Fatalf("expected configured terminal size and error, got %d x %d, %v", width, height, err)
	}
	if got := terminal.Theme(); got != "dark" {
		t.Fatalf("expected configured terminal theme, got %q", got)
	}
}

func TestTestTerminalZeroValue(t *testing.T) {
	var terminal TestTerminal
	if got := terminal.In(); got == nil {
		t.Fatal("expected zero-value terminal to provide an input reader")
	}
	if width, height, err := terminal.Size(); width != 0 || height != 0 || err != nil {
		t.Fatalf("expected zero-value size, got %d x %d, %v", width, height, err)
	}
}

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		wantTerminal  bool
		wantColor     bool
		want256Color  bool
		wantTrueColor bool
	}{
		{
			name: "default",
			env: map[string]string{
				"GH_FORCE_TTY":   "",
				"CLICOLOR":       "",
				"CLICOLOR_FORCE": "",
				"NO_COLOR":       "",
				"TERM":           "",
				"COLORTERM":      "",
			},
			wantTerminal:  false,
			wantColor:     false,
			want256Color:  false,
			wantTrueColor: false,
		},
		{
			name: "force color",
			env: map[string]string{
				"GH_FORCE_TTY":   "",
				"CLICOLOR":       "",
				"CLICOLOR_FORCE": "1",
				"NO_COLOR":       "",
				"TERM":           "",
				"COLORTERM":      "",
			},
			wantTerminal:  false,
			wantColor:     true,
			want256Color:  false,
			wantTrueColor: false,
		},
		{
			name: "force tty",
			env: map[string]string{
				"GH_FORCE_TTY":   "true",
				"CLICOLOR":       "",
				"CLICOLOR_FORCE": "",
				"NO_COLOR":       "",
				"TERM":           "",
				"COLORTERM":      "",
			},
			wantTerminal:  true,
			wantColor:     true,
			want256Color:  false,
			wantTrueColor: false,
		},
		{
			name: "has 256-color support",
			env: map[string]string{
				"GH_FORCE_TTY":   "true",
				"CLICOLOR":       "",
				"CLICOLOR_FORCE": "",
				"NO_COLOR":       "",
				"TERM":           "256-color",
				"COLORTERM":      "",
			},
			wantTerminal:  true,
			wantColor:     true,
			want256Color:  true,
			wantTrueColor: false,
		},
		{
			name: "has truecolor support",
			env: map[string]string{
				"GH_FORCE_TTY":   "true",
				"CLICOLOR":       "",
				"CLICOLOR_FORCE": "",
				"NO_COLOR":       "",
				"TERM":           "truecolor",
				"COLORTERM":      "",
			},
			wantTerminal:  true,
			wantColor:     true,
			want256Color:  true,
			wantTrueColor: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.env {
				t.Setenv(key, value)
			}
			terminal := FromEnv()
			if got := terminal.IsTerminalOutput(); got != tt.wantTerminal {
				t.Errorf("expected terminal %v, got %v", tt.wantTerminal, got)
			}
			if got := terminal.IsColorEnabled(); got != tt.wantColor {
				t.Errorf("expected color %v, got %v", tt.wantColor, got)
			}
			if got := terminal.Is256ColorSupported(); got != tt.want256Color {
				t.Errorf("expected 256-color %v, got %v", tt.want256Color, got)
			}
			if got := terminal.IsTrueColorSupported(); got != tt.wantTrueColor {
				t.Errorf("expected truecolor %v, got %v", tt.wantTrueColor, got)
			}
		})
	}
}
