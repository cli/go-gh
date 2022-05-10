package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGQLErrorMatch(t *testing.T) {
	tests := []struct {
		name      string
		error     GQLError
		kind      string
		path      string
		wantMatch bool
	}{
		{
			name: "matches path and type",
			error: GQLError{Errors: []GQLErrorItem{
				{Path: []interface{}{"repository", "issue"}, Type: "NOT_FOUND"},
			}},
			kind:      "NOT_FOUND",
			path:      "repository.issue",
			wantMatch: true,
		},
		{
			name: "matches base path and type",
			error: GQLError{Errors: []GQLErrorItem{
				{Path: []interface{}{"repository", "issue"}, Type: "NOT_FOUND"},
			}},
			kind:      "NOT_FOUND",
			path:      "repository.",
			wantMatch: true,
		},
		{
			name: "does not match path but matches type",
			error: GQLError{Errors: []GQLErrorItem{
				{Path: []interface{}{"repository", "issue"}, Type: "NOT_FOUND"},
			}},
			kind:      "NOT_FOUND",
			path:      "label.title",
			wantMatch: false,
		},
		{
			name: "matches path but not type",
			error: GQLError{Errors: []GQLErrorItem{
				{Path: []interface{}{"repository", "issue"}, Type: "NOT_FOUND"},
			}},
			kind:      "UNKNOWN",
			path:      "repository.issue",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMatch, tt.error.Match(tt.kind, tt.path))
		})
	}
}
