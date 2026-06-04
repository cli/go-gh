package markdown

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/stretchr/testify/assert"
)

// TestGlamourStyleColors ensures that the resulting string color codes match the expected values.
func TestGlamourStyleColors(t *testing.T) {
	tests := []struct {
		name string
		c    glamourStyleColor
		want string
	}{
		{
			name: "black",
			c:    black,
			want: "0",
		},
		{
			name: "red",
			c:    red,
			want: "1",
		},
		{
			name: "green",
			c:    green,
			want: "2",
		},
		{
			name: "yellow",
			c:    yellow,
			want: "3",
		},
		{
			name: "blue",
			c:    blue,
			want: "4",
		},
		{
			name: "magenta",
			c:    magenta,
			want: "5",
		},
		{
			name: "cyan",
			c:    cyan,
			want: "6",
		},
		{
			name: "white",
			c:    white,
			want: "7",
		},
		{
			name: "bright black",
			c:    brightBlack,
			want: "8",
		},
		{
			name: "bright red",
			c:    brightRed,
			want: "9",
		},
		{
			name: "bright green",
			c:    brightGreen,
			want: "10",
		},
		{
			name: "bright yellow",
			c:    brightYellow,
			want: "11",
		},
		{
			name: "bright blue",
			c:    brightBlue,
			want: "12",
		},
		{
			name: "bright magenta",
			c:    brightMagenta,
			want: "13",
		},
		{
			name: "bright cyan",
			c:    brightCyan,
			want: "14",
		},
		{
			name: "bright white",
			c:    brightWhite,
			want: "15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, *tt.c.code())
		})
	}
}

func TestAccessibleStyleConfig(t *testing.T) {
	tests := []struct {
		name  string
		theme string
		want  ansi.StyleConfig
	}{
		{
			name:  "light",
			theme: "light",
			want:  accessibleLightStyleConfig(),
		},
		{
			name:  "dark",
			theme: "dark",
			want:  accessibleDarkStyleConfig(),
		},
		{
			name:  "fallback",
			theme: "foo",
			want:  ansi.StyleConfig{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, AccessibleStyleConfig(tt.theme))
		})
	}
}

func TestAccessibleDarkStyleConfig(t *testing.T) {
	cfg := accessibleDarkStyleConfig()
	assert.Equal(t, white.code(), cfg.Document.StylePrimitive.Color)
	assert.Equal(t, brightCyan.code(), cfg.Link.Color)
	assert.Equal(t, brightMagenta.code(), cfg.Heading.StylePrimitive.Color)
	assert.Equal(t, brightWhite.code(), cfg.H1.StylePrimitive.Color)
	assert.Equal(t, brightBlue.code(), cfg.H1.StylePrimitive.BackgroundColor)
	assert.Equal(t, brightWhite.code(), cfg.Code.BackgroundColor)
	assert.Equal(t, red.code(), cfg.Code.Color)
	assert.Equal(t, brightMagenta.code(), cfg.Image.Color)
	assert.Equal(t, white.code(), cfg.HorizontalRule.Color)

	// Test that we haven't changed the original style
	assert.Equal(t, styles.DarkStyleConfig.H2, cfg.H2)
}

func TestAccessibleDarkStyleConfigIs4Bit(t *testing.T) {
	t.Parallel()

	cfg := accessibleDarkStyleConfig()
	validateColors(t, reflect.ValueOf(cfg), "StyleConfig")
}

func TestAccessibleLightStyleConfig(t *testing.T) {
	t.Parallel()

	cfg := accessibleLightStyleConfig()
	assert.Equal(t, black.code(), cfg.Document.StylePrimitive.Color)
	assert.Equal(t, brightBlue.code(), cfg.Link.Color)
	assert.Equal(t, magenta.code(), cfg.Heading.StylePrimitive.Color)
	assert.Equal(t, brightWhite.code(), cfg.H1.StylePrimitive.Color)
	assert.Equal(t, blue.code(), cfg.H1.StylePrimitive.BackgroundColor)
	assert.Equal(t, brightWhite.code(), cfg.Code.BackgroundColor)
	assert.Equal(t, red.code(), cfg.Code.Color)
	assert.Equal(t, magenta.code(), cfg.Image.Color)
	assert.Equal(t, white.code(), cfg.HorizontalRule.Color)

	// Test that we haven't changed the original style
	assert.Equal(t, styles.LightStyleConfig.H2, cfg.H2)
}

func TestAccessibleLightStyleConfigIs4Bit(t *testing.T) {
	t.Parallel()

	cfg := accessibleLightStyleConfig()
	validateColors(t, reflect.ValueOf(cfg), "StyleConfig")
}

// Walk every field in the StyleConfig struct, checking that the Color and
// BackgroundColor fields are valid 4-bit colors.
//
// This test skips Chroma fields because their Color fields are RGB hex values
// that are downsampled to 4-bit colors unlike Glamour, which are 8-bit colors.
// For more information, https://github.com/alecthomas/chroma/blob/0bf0e9f9ae2a81d463afe769cce01ff821bee3ba/formatters/tty_indexed.go#L32-L44
func validateColors(t *testing.T, v reflect.Value, path string) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Field(i)
			fieldType := v.Type().Field(i)

			// Construct path for better error reporting
			fieldPath := path + "." + fieldType.Name

			// Ensure we only check Glamour "Color" and "BackgroundColor"
			if fieldType.Name == "Chroma" {
				continue
			} else if (fieldType.Name == "Color" || fieldType.Name == "BackgroundColor") &&
				fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.String {

				if field.IsNil() {
					continue
				}
				color := field.Elem().String()
				_, err := parseGlamourStyleColor(color)
				assert.NoError(t, err, "Failed to parse color '%s' in %s", color, fieldPath)
			} else {
				// Recurse into nested structs
				validateColors(t, field, fieldPath)
			}
		}
	case reflect.Slice:
		// Handle slices of structs
		for i := range v.Len() {
			validateColors(t, v.Index(i), path+"[]")
		}
	}
}
