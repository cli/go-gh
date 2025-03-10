package markdown

import (
	"strconv"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

type ANSIColorCode int

const (
	Black ANSIColorCode = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

func (a ANSIColorCode) String() string {
	return strconv.Itoa(int(a))
}

func (a ANSIColorCode) StrPtr() *string {
	return strPtr(a.String())
}

func AccessibleStyleConfig(theme string) ansi.StyleConfig {
	switch theme {
	case "light":
		return accessibleLightStyleConfig()
	case "dark":
		return accessibleDarkStyleConfig()
	default:
		return ansi.StyleConfig{}
	}
}

func accessibleDarkStyleConfig() ansi.StyleConfig {
	cfg := styles.DarkStyleConfig

	// Text color
	cfg.Document.StylePrimitive.Color = White.StrPtr()

	// Link colors
	cfg.Link.Color = BrightCyan.StrPtr()

	// Heading colors
	cfg.Heading.StylePrimitive.Color = BrightMagenta.StrPtr()
	cfg.H1.StylePrimitive.Color = BrightWhite.StrPtr()
	cfg.H1.StylePrimitive.BackgroundColor = BrightBlue.StrPtr()

	// Code colors
	cfg.Code.BackgroundColor = BrightWhite.StrPtr()
	cfg.Code.Color = Red.StrPtr()

	// Image colors
	cfg.Image.Color = BrightMagenta.StrPtr()

	// Horizontal rule colors
	cfg.HorizontalRule.Color = White.StrPtr()

	return cfg
}

func accessibleLightStyleConfig() ansi.StyleConfig {
	cfg := styles.LightStyleConfig

	// Text color
	cfg.Document.StylePrimitive.Color = Black.StrPtr()

	// Link colors
	cfg.Link.Color = BrightBlue.StrPtr()

	// Heading colors
	cfg.Heading.StylePrimitive.Color = Magenta.StrPtr()
	cfg.H1.StylePrimitive.Color = BrightWhite.StrPtr()
	cfg.H1.StylePrimitive.BackgroundColor = Blue.StrPtr()

	// Code colors
	cfg.Code.BackgroundColor = BrightWhite.StrPtr()
	cfg.Code.Color = Red.StrPtr()

	// Image colors
	cfg.Image.Color = Magenta.StrPtr()

	// Horizontal rule colors
	cfg.HorizontalRule.Color = White.StrPtr()

	return cfg
}

func strPtr(s string) *string {
	return &s
}
