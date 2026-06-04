package markdown

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

// glamourStyleColor represents color codes used to customize glamour style elements.
type glamourStyleColor int

// Do not change the order of the following glamour color constants,
// which matches 4-bit colors with their respective color codes.
const (
	black glamourStyleColor = iota
	red
	green
	yellow
	blue
	magenta
	cyan
	white
	brightBlack
	brightRed
	brightGreen
	brightYellow
	brightBlue
	brightMagenta
	brightCyan
	brightWhite
)

func (gsc glamourStyleColor) code() *string {
	s := strconv.Itoa(int(gsc))
	return &s
}

func parseGlamourStyleColor(code string) (glamourStyleColor, error) {
	switch code {
	case "0":
		return black, nil
	case "1":
		return red, nil
	case "2":
		return green, nil
	case "3":
		return yellow, nil
	case "4":
		return blue, nil
	case "5":
		return magenta, nil
	case "6":
		return cyan, nil
	case "7":
		return white, nil
	case "8":
		return brightBlack, nil
	case "9":
		return brightRed, nil
	case "10":
		return brightGreen, nil
	case "11":
		return brightYellow, nil
	case "12":
		return brightBlue, nil
	case "13":
		return brightMagenta, nil
	case "14":
		return brightCyan, nil
	case "15":
		return brightWhite, nil
	default:
		return 0, fmt.Errorf("invalid color code: %s", code)
	}
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
	cfg.Document.StylePrimitive.Color = white.code()

	// Link colors
	cfg.Link.Color = brightCyan.code()
	cfg.LinkText.Color = brightCyan.code()

	// Heading colors
	cfg.Heading.StylePrimitive.Color = brightMagenta.code()
	cfg.H1.StylePrimitive.Color = brightWhite.code()
	cfg.H1.StylePrimitive.BackgroundColor = brightBlue.code()
	cfg.H6.StylePrimitive.Color = brightMagenta.code()

	// Code colors
	cfg.Code.BackgroundColor = brightWhite.code()
	cfg.Code.Color = red.code()

	// Image colors
	cfg.Image.Color = brightMagenta.code()
	cfg.ImageText.Color = brightMagenta.code()

	// Horizontal rule colors
	cfg.HorizontalRule.Color = white.code()

	// Code block colors
	// Unsetting StyleBlock color until we understand what it does versus Chroma style
	cfg.CodeBlock.StyleBlock.StylePrimitive.Color = nil

	return cfg
}

func accessibleLightStyleConfig() ansi.StyleConfig {
	cfg := styles.LightStyleConfig

	// Text color
	cfg.Document.StylePrimitive.Color = black.code()

	// Link colors
	cfg.Link.Color = brightBlue.code()
	cfg.LinkText.Color = brightBlue.code()

	// Heading colors
	cfg.Heading.StylePrimitive.Color = magenta.code()
	cfg.H1.StylePrimitive.Color = brightWhite.code()
	cfg.H1.StylePrimitive.BackgroundColor = blue.code()

	// Code colors
	cfg.Code.BackgroundColor = brightWhite.code()
	cfg.Code.Color = red.code()

	// Image colors
	cfg.Image.Color = magenta.code()
	cfg.ImageText.Color = magenta.code()

	// Horizontal rule colors
	cfg.HorizontalRule.Color = white.code()

	// Code block colors
	// Unsetting StyleBlock color until we understand what it does versus Chroma style
	cfg.CodeBlock.StyleBlock.StylePrimitive.Color = nil

	return cfg
}
