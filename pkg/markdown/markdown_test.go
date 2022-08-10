package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Render(t *testing.T) {
	t.Setenv("GLAMOUR_STYLE", "")
	tests := []struct {
		name  string
		text  string
		theme string
	}{
		{
			name:  "light style",
			text:  "some text",
			theme: "light",
		},
		{
			name:  "dark style",
			text:  "some text",
			theme: "dark",
		},
		{
			name:  "notty style",
			text:  "some text",
			theme: "none",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Render(tt.text, WithTheme(tt.theme))
			assert.NoError(t, err)
		})
	}
}
