package term

import (
	"strings"
	"testing"

	"github.com/mgutz/ansi"
)

func TestColorSchemeMuted(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
		want string
	}{
		{name: "disabled returns plain text", cs: ColorScheme{ColorEnabled: false}, want: "plain"},
		{name: "default theme no color", cs: ColorScheme{ColorEnabled: true, Theme: NoTheme}, want: "plain"},
		{name: "light theme colorizes", cs: ColorScheme{Accessible: true, ColorEnabled: true, Theme: LightTheme}, want: "plain"},
		{name: "dark theme colorizes", cs: ColorScheme{Accessible: true, ColorEnabled: true, Theme: DarkTheme}, want: "plain"},
		{name: "inaccessible uses gray path with 256", cs: ColorScheme{Accessible: false, ColorEnabled: true, Is256Enabled: true}, want: "plain"},
		{name: "inaccessible uses gray path without 256", cs: ColorScheme{Accessible: false, ColorEnabled: true, Is256Enabled: false}, want: "plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cs.Muted("plain")
			if !tt.cs.ColorEnabled {
				if got != tt.want {
					t.Fatalf("expected plain text when colors are disabled, got %q", got)
				}
				return
			}

			switch {
			case tt.cs.Accessible && tt.cs.Theme == LightTheme:
				if got == tt.want || !strings.Contains(got, "\x1b[") {
					t.Fatalf("expected light-theme muted output with ANSI codes, got %q", got)
				}
			case tt.cs.Accessible && tt.cs.Theme == DarkTheme:
				if got == tt.want || !strings.Contains(got, "\x1b[") {
					t.Fatalf("expected dark-theme muted output with ANSI codes, got %q", got)
				}
			case !tt.cs.Accessible && tt.cs.Is256Enabled:
				if got == tt.want || !strings.Contains(got, "\x1b[") {
					t.Fatalf("expected 256-color gray output, got %q", got)
				}
			case !tt.cs.Accessible:
				if got == tt.want || !strings.Contains(got, "\x1b[") {
					t.Fatalf("expected gray output, got %q", got)
				}
			default:
				if got != tt.want {
					t.Fatalf("expected plain text for default theme without color changes, got %q", got)
				}
			}
		})
	}
}

func TestColorSchemeHighlight(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
		want string
	}{
		{name: "disabled returns plain text", cs: ColorScheme{ColorEnabled: false}, want: "plain"},
		{name: "enabled returns highlighted text", cs: ColorScheme{ColorEnabled: true}, want: "plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cs.Highlight("plain")
			if !tt.cs.ColorEnabled {
				if got != tt.want {
					t.Fatalf("expected plain text with colors disabled, got %q", got)
				}
				return
			}

			if got == tt.want || !strings.Contains(got, "\x1b[") {
				t.Fatalf("expected highlighted text to contain ANSI escapes, got %q", got)
			}
		})
	}
}

func TestColorSchemeHighlightStart(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
		want string
	}{
		{name: "disabled returns empty", cs: ColorScheme{ColorEnabled: false}, want: ""},
		{name: "enabled returns highlight start sequence", cs: ColorScheme{ColorEnabled: true}, want: highlightStart},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.HighlightStart(); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestColorSchemeSuccessIcon(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
	}{
		{name: "disabled", cs: ColorScheme{ColorEnabled: false}},
		{name: "enabled", cs: ColorScheme{ColorEnabled: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := "✓"
			if tt.cs.ColorEnabled {
				want = tt.cs.Green("✓")
			}
			if got := tt.cs.SuccessIcon(); got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}
		})
	}
}

func TestColorSchemeSuccessIconWithColor(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
		fn   func(string) string
		want string
	}{
		{name: "custom formatter", cs: ColorScheme{ColorEnabled: true}, fn: func(s string) string { return "[green]" + s }, want: "[green]✓"},
		{name: "empty formatter", cs: ColorScheme{ColorEnabled: false}, fn: func(s string) string { return "" }, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.SuccessIconWithColor(tt.fn); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestColorSchemeWarningIcon(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
	}{
		{name: "disabled", cs: ColorScheme{ColorEnabled: false}},
		{name: "enabled", cs: ColorScheme{ColorEnabled: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := "!"
			if tt.cs.ColorEnabled {
				want = tt.cs.Yellow("!")
			}
			if got := tt.cs.WarningIcon(); got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}
		})
	}
}

func TestColorSchemeWarningIconWithColor(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
		fn   func(string) string
		want string
	}{
		{name: "custom formatter", cs: ColorScheme{ColorEnabled: true}, fn: func(s string) string { return "[yellow]" + s }, want: "[yellow]!"},
		{name: "empty formatter", cs: ColorScheme{ColorEnabled: false}, fn: func(s string) string { return "" }, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.WarningIconWithColor(tt.fn); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestColorSchemeFailureIcon(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
	}{
		{name: "disabled", cs: ColorScheme{ColorEnabled: false}},
		{name: "enabled", cs: ColorScheme{ColorEnabled: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := "X"
			if tt.cs.ColorEnabled {
				want = tt.cs.Red("X")
			}
			if got := tt.cs.FailureIcon(); got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}
		})
	}
}

func TestColorSchemeFailureIconWithColor(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
		fn   func(string) string
		want string
	}{
		{name: "custom formatter", cs: ColorScheme{ColorEnabled: true}, fn: func(s string) string { return "[red]" + s }, want: "[red]X"},
		{name: "empty formatter", cs: ColorScheme{ColorEnabled: false}, fn: func(s string) string { return "" }, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.FailureIconWithColor(tt.fn); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestColorSchemeReset(t *testing.T) {
	tests := []struct {
		name string
		cs   ColorScheme
		want string
	}{
		{name: "disabled returns empty", cs: ColorScheme{ColorEnabled: false}, want: ""},
		{name: "enabled returns ANSI reset", cs: ColorScheme{ColorEnabled: true}, want: ansi.Reset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.Reset(); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
