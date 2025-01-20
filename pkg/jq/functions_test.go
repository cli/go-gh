package jq

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTemplateFunctions(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		filter    string
		wantW     string
		wantError bool
	}{
		{
			name:   "timeago",
			input:  fmt.Sprintf(`{"time":"%s"}`, time.Now().Add(-5*time.Minute).Format(time.RFC3339)),
			filter: `.time | timeago`,
			wantW:  "5 minutes ago\n",
		},
		{
			name:      "timeago with int",
			input:     `{"time":42}`,
			filter:    `.time | timeago`,
			wantError: true,
		},
		{
			name:      "timeago with non-date string",
			input:     `{"time":"not a date-time"}`,
			filter:    `.time | timeago`,
			wantError: true,
		},
		{
			name:   "timefmt",
			input:  `{"time":"2025-01-20T01:08:15Z"}`,
			filter: `.time | timefmt("Mon, 02 Jan 2006 15:04:05 MST")`,
			wantW:  "Mon, 20 Jan 2025 01:08:15 UTC\n",
		},
		{
			name:      "timeago with int",
			input:     `{"time":42}`,
			filter:    `.time | timefmt("Mon, 02 Jan 2006 15:04:05 MST")`,
			wantError: true,
		},
		{
			name:      "timeago with invalid date-time string",
			input:     `{"time":"not a date-time"}`,
			filter:    `.time | timefmt("Mon, 02 Jan 2006 15:04:05 MST")`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			err := Evaluate(strings.NewReader(tt.input), &buf, tt.filter, WithTemplateFunctions())
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantW, buf.String())
		})
	}
}
