package prompter

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/stretchr/testify/assert"
)

func ExamplePrompter() {
	term := term.FromEnv()
	in, ok := term.In().(*os.File)
	if !ok {
		log.Fatal("error casting to file")
	}
	out, ok := term.Out().(*os.File)
	if !ok {
		log.Fatal("error casting to file")
	}
	errOut, ok := term.ErrOut().(*os.File)
	if !ok {
		log.Fatal("error casting to file")
	}
	prompter := New(in, out, errOut)
	response, err := prompter.Confirm("Shall we play a game", true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}

func TestLatinMatchingFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		value  string
		want   bool
	}{
		{
			name:   "exact match no diacritics",
			filter: "Mikelis",
			value:  "Mikelis",
			want:   true,
		},
		{
			name:   "exact match no diacritics",
			filter: "Mikelis",
			value:  "Mikelis",
			want:   true,
		},
		{
			name:   "exact match diacritics",
			filter: "Miķelis",
			value:  "Miķelis",
			want:   true,
		},
		{
			name:   "partial match diacritics",
			filter: "Miķe",
			value:  "Miķelis",
			want:   true,
		},
		{
			name:   "exact match diacritics in value",
			filter: "Mikelis",
			value:  "Miķelis",
			want:   true,
		},
		{
			name:   "partial match diacritics in filter",
			filter: "Miķe",
			value:  "Miķelis",
			want:   true,
		},
		{
			name:   "no match when removing diacritics in filter",
			filter: "Mielis",
			value:  "Mikelis",
			want:   false,
		},
		{
			name:   "no match when removing diacritics in value",
			filter: "Mikelis",
			value:  "Mielis",
			want:   false,
		},
		{
			name:   "no match diacritics in filter",
			filter: "Miķelis",
			value:  "Mikelis",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, latinMatchingFilter(tt.filter, tt.value, 0), tt.want)
		})
	}
}
