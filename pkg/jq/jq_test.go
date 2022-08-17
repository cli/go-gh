package jq

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	t.Setenv("CODE", "code_c")
	type args struct {
		json io.Reader
		expr string
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				json: strings.NewReader(`{"name":"Mona", "arms":8}`),
				expr: `.name`,
			},
			wantW: "Mona\n",
		},
		{
			name: "multiple queries",
			args: args{
				json: strings.NewReader(`{"name":"Mona", "arms":8}`),
				expr: `.name,.arms`,
			},
			wantW: "Mona\n8\n",
		},
		{
			name: "object as JSON",
			args: args{
				json: strings.NewReader(`{"user":{"login":"monalisa"}}`),
				expr: `.user`,
			},
			wantW: "{\"login\":\"monalisa\"}\n",
		},
		{
			name: "complex",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{
						"title": "First title",
						"labels": [{"name":"bug"}, {"name":"help wanted"}]
					},
					{
						"title": "Second but not last",
						"labels": []
					},
					{
						"title": "Alas, tis' the end",
						"labels": [{}, {"name":"feature"}]
					}
				]`)),
				expr: `.[] | [.title,(.labels | map(.name) | join(","))] | @tsv`,
			},
			wantW: heredoc.Doc(`
				First title	bug,help wanted
				Second but not last	
				Alas, tis' the end	,feature
			`),
		},
		{
			name: "with env var",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{
						"title": "code_a",
						"labels": [{"name":"bug"}, {"name":"help wanted"}]
					},
					{
						"title": "code_b",
						"labels": []
					},
					{
						"title": "code_c",
						"labels": [{}, {"name":"feature"}]
					}
				]`)),
				expr: `.[] | select(.title == env.CODE) | .labels`,
			},
			wantW: "[{},{\"name\":\"feature\"}]\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := Evaluate(tt.args.json, w, tt.args.expr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}
