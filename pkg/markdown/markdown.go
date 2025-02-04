// Package markdown facilitates rendering markdown in the terminal.
package markdown

import (
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

// WithoutIndentation is a rendering option that removes indentation from the markdown rendering.
func WithoutIndentation() glamour.TermRendererOption {
	overrides := []byte(`
	  {
			"document": {
				"margin": 0
			},
			"code_block": {
				"margin": 0
			}
	  }`)

	return glamour.WithStylesFromJSONBytes(overrides)
}

// WithoutWrap is a rendering option that set the character limit for soft wraping the markdown rendering.
func WithWrap(w int) glamour.TermRendererOption {
	return glamour.WithWordWrap(w)
}

// WithTheme is a rendering option that sets the theme to use while rendering the markdown.
// It can be used in conjunction with [term.Theme].
// If the environment variable GLAMOUR_STYLE is set, it will take precedence over the provided theme.
func WithTheme(theme string) glamour.TermRendererOption {
	style := os.Getenv("GLAMOUR_STYLE")
	if style == "" || style == "auto" {
		switch theme {
		case "light", "dark":
			style = theme
		default:
			style = "notty"
		}
	}
	return glamour.WithStylePath(style)
}

// WithBaseURL is a rendering option that sets the base URL to use when rendering relative URLs.
func WithBaseURL(u string) glamour.TermRendererOption {
	return glamour.WithBaseURL(u)
}

// Render the markdown string according to the specified rendering options.
// By default emoji are rendered and new lines are preserved.
func Render(text string, opts ...glamour.TermRendererOption) (string, error) {
	if accessible := os.Getenv("ACCESSIBLE"); accessible != "" {
		makeDefaultStyleColorsAccessible()
	}

	// Glamour rendering preserves carriage return characters in code blocks, but
	// we need to ensure that no such characters are present in the output.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	opts = append(opts, glamour.WithEmoji(), glamour.WithPreservedNewLines())
	tr, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return "", err
	}
	return tr.Render(text)
}

var (
	ansiBlack         = stringPtr("0")
	ansiRed           = stringPtr("1")
	ansiGreen         = stringPtr("2")
	ansiYellow        = stringPtr("3")
	ansiBlue          = stringPtr("4")
	ansiMagenta       = stringPtr("5")
	ansiCyan          = stringPtr("6")
	ansiWhite         = stringPtr("7")
	ansiBrightBlack   = stringPtr("8")
	ansiBrightRed     = stringPtr("9")
	ansiBrightGreen   = stringPtr("10")
	ansiBrightYellow  = stringPtr("11")
	ansiBrightBlue    = stringPtr("12")
	ansiBrightMagenta = stringPtr("13")
	ansiBrightCyan    = stringPtr("14")
	ansiBrightWhite   = stringPtr("15")
)

func makeDefaultStyleColorsAccessible() {
	// Changes to the `dark` style for users with low vision.
	glamour.DarkStyleConfig.Document.StylePrimitive.Color = ansiWhite             // Original value of 252
	glamour.DarkStyleConfig.Heading.StylePrimitive.Color = ansiBrightCyan         // Original value of 39
	glamour.DarkStyleConfig.H1.StylePrimitive.Color = ansiBrightYellow            // Original value of 228
	glamour.DarkStyleConfig.H1.StylePrimitive.BackgroundColor = ansiBrightBlue    // Original value of 63
	glamour.DarkStyleConfig.H6.StylePrimitive.Color = ansiGreen                   // Original value of 35
	glamour.DarkStyleConfig.HorizontalRule.Color = ansiBlack                      // Original value of 240
	glamour.DarkStyleConfig.Link.Color = ansiBrightBlue                           // Original value of 30
	glamour.DarkStyleConfig.LinkText.Color = ansiBlue                             // Original value of 35
	glamour.DarkStyleConfig.Image.Color = ansiBrightMagenta                       // Original value of 212
	glamour.DarkStyleConfig.ImageText.Color = ansiMagenta                         // Original value of 243
	glamour.DarkStyleConfig.Code.StylePrimitive.Color = ansiBrightRed             // Original value of 203
	glamour.DarkStyleConfig.Code.StylePrimitive.BackgroundColor = ansiWhite       // Original value of 236
	glamour.DarkStyleConfig.CodeBlock.StyleBlock.StylePrimitive.Color = ansiWhite // Original value of 244

	// Changes to the `light` style for users with low vision.
	glamour.LightStyleConfig.Document.StylePrimitive.Color = ansiBlack             // Original value of 234
	glamour.LightStyleConfig.Heading.StylePrimitive.Color = ansiBrightBlue         // Original value of 27
	glamour.LightStyleConfig.H1.StylePrimitive.Color = ansiBrightYellow            // Original value of 228
	glamour.LightStyleConfig.H1.StylePrimitive.BackgroundColor = ansiBrightBlue    // Original value of 63
	glamour.LightStyleConfig.HorizontalRule.Color = ansiWhite                      // Original value of 249
	glamour.LightStyleConfig.Link.Color = ansiGreen                                // Original value of 36
	glamour.LightStyleConfig.LinkText.Color = ansiBlack                            // Original value of 29
	glamour.LightStyleConfig.Image.Color = ansiBrightMagenta                       // Original value of 205
	glamour.LightStyleConfig.ImageText.Color = ansiBlack                           // Original value of 243
	glamour.LightStyleConfig.Code.StylePrimitive.Color = ansiBrightRed             // Original value of 203
	glamour.LightStyleConfig.Code.StylePrimitive.BackgroundColor = ansiWhite       // Original value of 254
	glamour.LightStyleConfig.CodeBlock.StyleBlock.StylePrimitive.Color = ansiBlack // Original value of 242
}

// stringPtr is a helper for transforming string values into form needed to customize charmbracelet/glamour ANSI styles.
func stringPtr(s string) *string {
	return &s
}
