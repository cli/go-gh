package term

import (
	"fmt"

	"github.com/mgutz/ansi"
)

const (
	highlightStyle = "black:yellow"
)

var (
	black   = ansi.ColorFunc("black")
	red     = ansi.ColorFunc("red")
	green   = ansi.ColorFunc("green")
	yellow  = ansi.ColorFunc("yellow")
	blue    = ansi.ColorFunc("blue")
	magenta = ansi.ColorFunc("magenta")
	cyan    = ansi.ColorFunc("cyan")
	white   = ansi.ColorFunc("white")
	bold    = ansi.ColorFunc("default+b")

	gray    = ansi.ColorFunc("black+h")
	gray256 = func(t string) string {
		return fmt.Sprintf("\x1b[38;5;242m%s\x1b[0m", t)
	}

	highlight      = ansi.ColorFunc(highlightStyle)
	highlightStart = ansi.ColorCode(highlightStyle)

	darkThemeMuted  = ansi.ColorFunc("white+d")
	lightThemeMuted = ansi.ColorFunc("black+h")
)

// ColorScheme for the current [Term].
type ColorScheme struct {
	Accessible   bool
	ColorEnabled bool
	Is256Enabled bool
	Theme        string
}

func (c ColorScheme) Black(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return black(s)
}

func (c ColorScheme) Red(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return red(s)
}

func (c ColorScheme) Green(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return green(s)
}

func (c ColorScheme) Yellow(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return yellow(s)
}

func (c ColorScheme) Blue(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return blue(s)
}

func (c ColorScheme) Magenta(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return magenta(s)
}

func (c ColorScheme) Cyan(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return cyan(s)
}

func (c ColorScheme) White(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return white(s)
}

func (c ColorScheme) Bold(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return bold(s)
}

func (c ColorScheme) gray(s string) string {
	if !c.ColorEnabled {
		return s
	}
	if c.Is256Enabled {
		return gray256(s)
	}
	return gray(s)
}

func (c ColorScheme) Muted(s string) string {
	if !c.Accessible {
		return c.gray(s)
	}

	if !c.ColorEnabled {
		return s
	}

	switch c.Theme {
	case LightTheme:
		return lightThemeMuted(s)
	case DarkTheme:
		return darkThemeMuted(s)
	default:
		return s
	}
}

func (c ColorScheme) Highlight(s string) string {
	if !c.ColorEnabled {
		return s
	}

	return highlight(s)
}

// HighlightStart starts highlighting text.
// Use [ColorScheme.Reset] to end highlighting.
func (c ColorScheme) HighlightStart() string {
	if !c.ColorEnabled {
		return ""
	}

	return highlightStart
}

func (c ColorScheme) SuccessIcon() string {
	return c.SuccessIconWithColor(c.Green)
}

func (c ColorScheme) SuccessIconWithColor(f func(string) string) string {
	return f("✓")
}

func (c ColorScheme) WarningIcon() string {
	return c.WarningIconWithColor(c.Yellow)
}

func (c ColorScheme) WarningIconWithColor(f func(string) string) string {
	return f("!")
}

func (c ColorScheme) FailureIcon() string {
	return c.FailureIconWithColor(c.Red)
}

func (c ColorScheme) FailureIconWithColor(f func(string) string) string {
	return f("X")
}

func (c ColorScheme) Reset() string {
	if !c.ColorEnabled {
		return ""
	}

	return ansi.Reset
}
