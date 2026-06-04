package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/cli/go-gh/v2/pkg/x/color"
	ansi "github.com/leaanthony/go-ansi-parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// glamour theme colors found at https://github.com/charmbracelet/glamour/tree/master/styles
const (
	glamourLightH2_8bitColorSeq = "\x1b[38;5;27;"
	glamourDarkH2_8bitColorSeq  = "\x1b[38;5;39;"
	customH2_8bitColorSeq       = "\x1b[38;5;61;"
	magenta_4bitColorSeq        = "\x1b[35;"
	brightMagenta_4bitColorSeq  = "\x1b[95;"
)

// Test_RenderAccessible tests rendered markdown for accessibility concerns such as color mode / depth and other display attributes.
// It works by parsing the rendered markdown for ANSI escape sequences and checking their display attributes.
// Test scenarios allow multiple color mode / depths because `ansi.Parse()` considers `\x1b[0m` sequence as part of `ansi.Default`.
func Test_RenderAccessible(t *testing.T) {
	anchor := heredoc.Doc(`
		[GitHub CLI repository](https://github.com/cli/cli)
	`)

	img := heredoc.Doc(`
		[Animated Mona for loading screen](https://github.com/user-attachments/assets/a43e7ce6-8360-466c-b2d5-0c74a60c30a4)
	`)

	goCodeBlock := heredoc.Docf(`
		%[1]s%[1]s%[1]sgo
		package main

		import (
			"fmt"
		)

		func main() {
			fmt.Println("Hello, world!")
		}
		%[1]s%[1]s%[1]s
	`, "`")

	shellCodeBlock := heredoc.Docf(`
		%[1]s%[1]s%[1]sshell
		# list all repositories for a user
		$ gh api graphql --paginate -f query='
			query($endCursor: String) {
				viewer {
					repositories(first: 100, after: $endCursor) {
						nodes { nameWithOwner }
						pageInfo {
							hasNextPage
							endCursor
						}
					}
				}
			}
		'
		%[1]s%[1]s%[1]s
	`, "`")

	tests := []struct {
		name              string
		text              string
		theme             string
		accessible        bool
		wantColourModes   []ansi.ColourMode
		allowDimFaintText bool
	}{
		// Go block
		{
			name:              "when the light theme is selected, the Go codeblock renders using 8-bit colors",
			text:              goCodeBlock,
			theme:             "light",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the dark theme is selected, the Go codeblock renders using 8-bit colors",
			text:              goCodeBlock,
			theme:             "dark",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the accessible env var is set and the light theme is selected, the Go codeblock renders using 4-bit colors without dim/faint text",
			text:              goCodeBlock,
			theme:             "light",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
		{
			name:              "when the accessible env var is set and the dark theme is selected, the Go codeblock renders using 4-bit colors without dim/faint text",
			text:              goCodeBlock,
			theme:             "dark",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
		// shell block
		{
			name:              "when the light theme is selected, the Shell codeblock renders using 8-bit colors",
			text:              shellCodeBlock,
			theme:             "light",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the dark theme is selected, the Shell codeblock renders using 8-bit colors",
			text:              shellCodeBlock,
			theme:             "dark",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the accessible env var is set and the light theme is selected, the Shell codeblock renders using 4-bit colors without dim/faint text",
			text:              shellCodeBlock,
			theme:             "light",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
		{
			name:              "when the accessible env var is set and the dark theme is selected, the Shell codeblock renders using 4-bit colors without dim/faint text",
			text:              shellCodeBlock,
			theme:             "dark",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
		// image text and link
		{
			name:              "when the light theme is selected, the image text and link rendered using 8-bit colors",
			text:              img,
			theme:             "light",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the dark theme is selected, the image text and link rendered using 8-bit colors",
			text:              img,
			theme:             "dark",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the accessible env var is set and the light theme is selected, the image text and link render using 4-bit colors without dim/faint text",
			text:              img,
			theme:             "light",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
		{
			name:              "when the accessible env var is set and the dark theme is selected, the image text and link render using 4-bit colors without dim/faint text",
			text:              img,
			theme:             "dark",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
		// anchor text and link
		{
			name:              "when the light theme is selected, the anchor text and link rendered using 8-bit colors",
			text:              anchor,
			theme:             "light",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the dark theme is selected, the anchor text and link rendered using 8-bit colors",
			text:              anchor,
			theme:             "dark",
			wantColourModes:   []ansi.ColourMode{ansi.Default, ansi.TwoFiveSix},
			allowDimFaintText: true,
		},
		{
			name:              "when the accessible env var is set and the light theme is selected, the anchor text and link render using 4-bit colors without dim/faint text",
			text:              anchor,
			theme:             "light",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
		{
			name:              "when the accessible env var is set and the dark theme is selected, the anchor text and link render using 4-bit colors without dim/faint text",
			text:              anchor,
			theme:             "dark",
			accessible:        true,
			wantColourModes:   []ansi.ColourMode{ansi.Default},
			allowDimFaintText: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.accessible {
				t.Setenv(color.AccessibleColorsEnv, "true")
			}

			out, err := Render(tt.text, WithTheme(tt.theme))
			require.NoError(t, err)

			// Parse and test ANSI escape sequences from rendered markdown for inaccessible display attributes
			styledText, err := ansi.Parse(out)
			require.NoError(t, err)

			for _, st := range styledText {
				require.Containsf(t, tt.wantColourModes, st.ColourMode, "Unexpected color mode detected in '%s' at %d", st, st.Offset)

				if st.Faint() {
					require.Truef(t, tt.allowDimFaintText, "Unexpected dim/faint text detected in '%s' at %d", st, st.Offset)
				}
			}
		})
	}
}

// Test_RenderColor verifies that the proper ANSI color codes are applied to the rendered
// markdown by examining the ANSI escape sequences in the output for the correct color
// match. For more information on ANSI color codes, see
// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
func Test_RenderColor(t *testing.T) {
	t.Setenv("GLAMOUR_STYLE", "")

	codeBlock := heredoc.Docf(`
		%[1]s%[1]s%[1]sgo
		fmt.Println("Hello, world!")
		%[1]s%[1]s%[1]s
	`, "`")

	tests := []struct {
		name             string
		text             string
		theme            string
		styleEnvVar      string
		accessibleEnvVar string
		wantOut          string
	}{
		{
			name:    "when the light theme is selected, the h2 renders using the 8-bit blue 27 provided by glamour",
			text:    "## h2",
			theme:   "light",
			wantOut: fmt.Sprintf("%s1mh2", glamourLightH2_8bitColorSeq),
		},
		{
			name:    "when the dark theme is selected, the h2 renders using the 8-bit blue 39 provided by glamour",
			text:    "## h2",
			theme:   "dark",
			wantOut: fmt.Sprintf("%s1mh2", glamourDarkH2_8bitColorSeq),
		},
		{
			name:    "when no theme is selected, the h2 renders in plain text without ansi coloring",
			text:    "## h2",
			theme:   "none",
			wantOut: "## h2",
		},
		{
			name:        "when the style env var is set, we override the theme with that style",
			text:        "## h2",
			theme:       "light",
			styleEnvVar: "customStyle",
			wantOut:     fmt.Sprintf("%s1mh2", customH2_8bitColorSeq),
		},
		{
			name:             "when the accessible env var is set and the light theme is selected, the h2 renders using the 4-bit magenta provided by the light accessible style",
			text:             "## h2",
			theme:            "light",
			accessibleEnvVar: "true",
			wantOut:          fmt.Sprintf("%s1mh2", magenta_4bitColorSeq),
		},
		{
			name:             "when the accessible env var is set and the dark theme is selected, the h2 renders using the 4-bit bright magenta provided by the dark accessible style",
			text:             "## h2",
			theme:            "dark",
			accessibleEnvVar: "true",
			wantOut:          fmt.Sprintf("%s1mh2", brightMagenta_4bitColorSeq),
		},
		{
			name:    "when the light theme is selected, the codeblock renders using 8-bit colors",
			text:    codeBlock,
			theme:   "light",
			wantOut: "\x1b[0m\x1b[38;5;235mfmt\x1b[0m\x1b[38;5;210m.\x1b[0m\x1b[38;5;35mPrintln\x1b[0m\x1b[38;5;210m(\x1b[0m\x1b[38;5;95m\"Hello, world!\"\x1b[0m\x1b[38;5;210m)\x1b[0m",
		},
		{
			name:    "when the dark theme is selected, the codeblock renders using 8-bit colors",
			text:    codeBlock,
			theme:   "dark",
			wantOut: "\x1b[0m\x1b[38;5;251mfmt\x1b[0m\x1b[38;5;187m.\x1b[0m\x1b[38;5;42mPrintln\x1b[0m\x1b[38;5;187m(\x1b[0m\x1b[38;5;173m\"Hello, world!\"\x1b[0m\x1b[38;5;187m)\x1b[0m",
		},
		{
			name:             "when the accessible env var is set and the light theme is selected, the codeblock renders using 4-bit colors",
			text:             codeBlock,
			theme:            "light",
			accessibleEnvVar: "true",
			wantOut:          "\x1b[0m\x1b[30mfmt\x1b[0m\x1b[33m.\x1b[0m\x1b[36mPrintln\x1b[0m\x1b[33m(\x1b[0m\x1b[90m\"Hello, world!\"\x1b[0m\x1b[33m)\x1b[0m",
		},
		{
			name:             "when the accessible env var is set and the dark theme is selected, the codeblock renders using 4-bit colors",
			text:             codeBlock,
			theme:            "dark",
			accessibleEnvVar: "true",
			wantOut:          "\x1b[0m\x1b[37mfmt\x1b[0m\x1b[37m.\x1b[0m\x1b[36mPrintln\x1b[0m\x1b[37m(\x1b[0m\x1b[33m\"Hello, world!\"\x1b[0m\x1b[37m)\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				// Chroma caches charm style used to render codeblocks, it must be unregistered to avoid previously used style being reused.
				delete(styles.Registry, "charm")
			})
			t.Setenv("GH_ACCESSIBLE_COLORS", tt.accessibleEnvVar)

			if tt.styleEnvVar != "" {
				path := filepath.Join(t.TempDir(), fmt.Sprintf("%s.json", tt.styleEnvVar))
				err := os.WriteFile(path, []byte(customGlamourStyle(t)), 0644)
				require.NoError(t, err)
				t.Setenv("GLAMOUR_STYLE", path)
			}

			out, err := Render(tt.text, WithTheme(tt.theme))
			require.NoError(t, err)
			assert.Contains(t, out, tt.wantOut)
		})
	}
}

func customGlamourStyle(t *testing.T) string {
	t.Helper()
	colorCode := strings.Split(customH2_8bitColorSeq, ";")[2]
	return fmt.Sprintf(`
{
	"heading": {
		"block_suffix": "\n",
		"color": "%s",
		"bold": true
	},
	"h2": {
		"prefix": "## "
	}
}`, colorCode)
}
