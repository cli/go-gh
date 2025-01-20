package jq

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
)

func ExampleEvaluate() {
	now := time.Now()
	input := strings.NewReader(fmt.Sprintf(`[
		{
			"event": "first event",
			"time": "%s"
		},
		{
			"event": "second event",
			"time": "%s"
		}
	]`,
		now.Add(-10*time.Minute).Format(time.RFC3339),
		now.Add(-5*time.Minute).Format(time.RFC3339),
	))

	output := bytes.Buffer{}
	err := Evaluate(input, &output, "map(.time |= timeago) | .[]", WithTemplateFunctions())
	if err != nil {
		panic(err)
	}

	if _, err := io.Copy(os.Stdout, &output); err != nil {
		panic(err)
	}

	// Output:
	// {"event":"first event","time":"10 minutes ago"}
	// {"event":"second event","time":"5 minutes ago"}
}

func TestEvaluateFormatted(t *testing.T) {
	t.Setenv("CODE", "code_c")
	type args struct {
		json     io.Reader
		expr     string
		indent   string
		colorize bool
		options  []EvaluateOption
	}
	tests := []struct {
		name       string
		args       args
		wantW      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "simple",
			args: args{
				json:     strings.NewReader(`{"name":"Mona", "arms":8}`),
				expr:     `.name`,
				indent:   "",
				colorize: false,
			},
			wantW: "Mona\n",
		},
		{
			name: "multiple queries",
			args: args{
				json:     strings.NewReader(`{"name":"Mona", "arms":8}`),
				expr:     `.name,.arms`,
				indent:   "",
				colorize: false,
			},
			wantW: "Mona\n8\n",
		},
		{
			name: "object as JSON",
			args: args{
				json:     strings.NewReader(`{"user":{"login":"monalisa"}}`),
				expr:     `.user`,
				indent:   "",
				colorize: false,
			},
			wantW: "{\"login\":\"monalisa\"}\n",
		},
		{
			name: "object as JSON, indented",
			args: args{
				json:     strings.NewReader(`{"user":{"login":"monalisa"}}`),
				expr:     `.user`,
				indent:   "  ",
				colorize: false,
			},
			wantW: "{\n  \"login\": \"monalisa\"\n}\n",
		},
		{
			name: "object as JSON, indented & colorized",
			args: args{
				json:     strings.NewReader(`{"user":{"login":"monalisa"}}`),
				expr:     `.user`,
				indent:   "  ",
				colorize: true,
			},
			wantW: "\x1b[1;38m{\x1b[m\n" +
				"  \x1b[1;34m\"login\"\x1b[m\x1b[1;38m:\x1b[m" +
				" \x1b[32m\"monalisa\"\x1b[m\n" +
				"\x1b[1;38m}\x1b[m\n",
		},
		{
			name: "empty array",
			args: args{
				json:     strings.NewReader(`[]`),
				expr:     `., [], unique`,
				indent:   "",
				colorize: false,
			},
			wantW: "[]\n[]\n[]\n",
		},
		{
			name: "empty array, colorized",
			args: args{
				json:     strings.NewReader(`[]`),
				expr:     `.`,
				indent:   "",
				colorize: true,
			},
			wantW: "\x1b[1;38m[\x1b[m\x1b[1;38m]\x1b[m\n",
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
				expr:     `.[] | [.title,(.labels | map(.name) | join(","))] | @tsv`,
				indent:   "",
				colorize: false,
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
				expr:     `.[] | select(.title == env.CODE) | .labels`,
				indent:   "  ",
				colorize: false,
			},
			wantW: "[\n  {},\n  {\n    \"name\": \"feature\"\n  }\n]\n",
		},
		{
			name: "mixing scalars, arrays and objects",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					"foo",
					true,
					42,
					[17, 23],
					{"foo": "bar"}
				]`)),
				expr:     `.[]`,
				indent:   "  ",
				colorize: true,
			},
			wantW: "foo\ntrue\n42\n" +
				"\x1b[1;38m[\x1b[m\n" +
				"  17\x1b[1;38m,\x1b[m\n" +
				"  23\n" +
				"\x1b[1;38m]\x1b[m\n" +
				"\x1b[1;38m{\x1b[m\n" +
				"  \x1b[1;34m\"foo\"\x1b[m\x1b[1;38m:\x1b[m" +
				" \x1b[32m\"bar\"\x1b[m\n" +
				"\x1b[1;38m}\x1b[m\n",
		},
		{
			name: "halt function",
			args: args{
				json: strings.NewReader("{}"),
				expr: `1,halt,2`,
			},
			wantW: "1\n",
		},
		{
			name: "halt_error function",
			args: args{
				json: strings.NewReader("{}"),
				expr: `1,halt_error,2`,
			},
			wantW:      "1\n",
			wantErr:    true,
			wantErrMsg: "halt error: {}",
		},
		{
			name: "invalid one-line query",
			args: args{
				json: strings.NewReader("{}"),
				expr: `[1,2,,3]`,
			},
			wantErr: true,
			wantErrMsg: `failed to parse jq expression (line 1, column 6)
    [1,2,,3]
         ^  unexpected token ","`,
		},
		{
			name: "invalid multi-line query",
			args: args{
				json: strings.NewReader("{}"),
				expr: `[
  1,,2
  ,3]`,
			},
			wantErr: true,
			wantErrMsg: `failed to parse jq expression (line 2, column 5)
      1,,2
        ^  unexpected token ","`,
		},
		{
			name: "invalid unterminated query",
			args: args{
				json: strings.NewReader("{}"),
				expr: `[1,`,
			},
			wantErr: true,
			wantErrMsg: `failed to parse jq expression (line 1, column 4)
    [1,
       ^  unexpected EOF`,
		},
		{
			name: "with module path",
			args: args{
				json: strings.NewReader(`[1,2]`),
				expr: `import "mod" as m; map(m::inc)`,
				options: []EvaluateOption{
					WithModulePaths([]string{"testdata"}),
				},
			},
			wantW: "[2,3]\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := EvaluateFormatted(
				tt.args.json,
				w,
				tt.args.expr,
				tt.args.indent,
				tt.args.colorize,
				tt.args.options...,
			)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}
