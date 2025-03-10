package markdown

import (
	"testing"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/stretchr/testify/assert"
)

func TestANSIColorCodeString(t *testing.T) {
	tests := []struct {
		name string
		c    ANSIColorCode
		want string
	}{
		{
			name: "red",
			c:    Red,
			want: "1",
		},
		{
			name: "bright red",
			c:    BrightRed,
			want: "9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.String())
		})
	}
}

func TestANSIColorCodeStrPtr(t *testing.T) {
	tests := []struct {
		name string
		c    ANSIColorCode
		want *string
	}{
		{
			name: "green",
			c:    Green,
			want: strPtr("2"),
		},
		{
			name: "bright green",
			c:    BrightGreen,
			want: strPtr("10"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.StrPtr())
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
			assert.Equal(t, tt.want, AccessibleStyleConfig(tt.theme))
		})
	}
}

func Test_accessibleDarkStyleConfig(t *testing.T) {
	cfg := accessibleDarkStyleConfig()
	assert.Equal(t, White.StrPtr(), cfg.Document.StylePrimitive.Color)
	assert.Equal(t, BrightCyan.StrPtr(), cfg.Link.Color)
	assert.Equal(t, BrightMagenta.StrPtr(), cfg.Heading.StylePrimitive.Color)
	assert.Equal(t, BrightWhite.StrPtr(), cfg.H1.StylePrimitive.Color)
	assert.Equal(t, BrightBlue.StrPtr(), cfg.H1.StylePrimitive.BackgroundColor)
	assert.Equal(t, BrightWhite.StrPtr(), cfg.Code.BackgroundColor)
	assert.Equal(t, Red.StrPtr(), cfg.Code.Color)
	assert.Equal(t, BrightMagenta.StrPtr(), cfg.Image.Color)
	assert.Equal(t, White.StrPtr(), cfg.HorizontalRule.Color)

	// Test that we haven't changed the original style
	assert.Equal(t, styles.DarkStyleConfig.H2, cfg.H2)
}

func Test_accessibleLightStyleConfig(t *testing.T) {
	cfg := accessibleLightStyleConfig()
	assert.Equal(t, Black.StrPtr(), cfg.Document.StylePrimitive.Color)
	assert.Equal(t, BrightBlue.StrPtr(), cfg.Link.Color)
	assert.Equal(t, Magenta.StrPtr(), cfg.Heading.StylePrimitive.Color)
	assert.Equal(t, BrightWhite.StrPtr(), cfg.H1.StylePrimitive.Color)
	assert.Equal(t, Blue.StrPtr(), cfg.H1.StylePrimitive.BackgroundColor)
	assert.Equal(t, BrightWhite.StrPtr(), cfg.Code.BackgroundColor)
	assert.Equal(t, Red.StrPtr(), cfg.Code.Color)
	assert.Equal(t, Magenta.StrPtr(), cfg.Image.Color)
	assert.Equal(t, White.StrPtr(), cfg.HorizontalRule.Color)

	// Test that we haven't changed the original style
	assert.Equal(t, styles.LightStyleConfig.H2, cfg.H2)
}
