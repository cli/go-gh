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
