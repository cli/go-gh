package template

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/v2/pkg/text"
	"github.com/stretchr/testify/assert"
)

func ExampleTemplate() {
	// Information about the terminal can be obtained using the [pkg/term] package.
	colorEnabled := true
	termWidth := 14
	json := strings.NewReader(heredoc.Doc(`[
		{"number": 1, "title": "One"},
		{"number": 2, "title": "Two"}
	]`))
	template := "HEADER\n\n{{range .}}{{tablerow .number .title}}{{end}}{{tablerender}}\nFOOTER"
	tmpl := New(os.Stdout, termWidth, colorEnabled)
	if err := tmpl.Parse(template); err != nil {
		log.Fatal(err)
	}
	if err := tmpl.Execute(json); err != nil {
		log.Fatal(err)
	}
	// Output:
	// HEADER
	//
	// 1  One
	// 2  Two
	//
	// FOOTER
}

func ExampleTemplate_Funcs() {
	// Information about the terminal can be obtained using the [pkg/term] package.
	colorEnabled := true
	termWidth := 14
	json := strings.NewReader(heredoc.Doc(`[
		{"num": 1, "thing": "apple"},
		{"num": 2, "thing": "orange"}
	]`))
	template := "{{range .}}* {{pluralize .num .thing}}\n{{end}}"
	tmpl := New(os.Stdout, termWidth, colorEnabled)
	tmpl.Funcs(map[string]interface{}{
		"pluralize": func(fields ...interface{}) (string, error) {
			if l := len(fields); l != 2 {
				return "", fmt.Errorf("wrong number of args for pluralize: want 2 got %d", l)
			}
			var ok bool
			var num float64
			var thing string
			if num, ok = fields[0].(float64); !ok && num == float64(int(num)) {
				return "", fmt.Errorf("invalid value; expected int")
			}
			if thing, ok = fields[1].(string); !ok {
				return "", fmt.Errorf("invalid value; expected string")
			}
			return text.Pluralize(int(num), thing), nil
		},
	})
	if err := tmpl.Parse(template); err != nil {
		log.Fatal(err)
	}
	if err := tmpl.Execute(json); err != nil {
		log.Fatal(err)
	}
	// Output:
	// * 1 apple
	// * 2 oranges
}

func TestJsonScalarToString(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "string",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "int",
			input: float64(1234),
			want:  "1234",
		},
		{
			name:  "float",
			input: float64(12.34),
			want:  "12.34",
		},
		{
			name:  "null",
			input: nil,
			want:  "",
		},
		{
			name:  "true",
			input: true,
			want:  "true",
		},
		{
			name:  "false",
			input: false,
			want:  "false",
		},
		{
			name:    "object",
			input:   map[string]interface{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonScalarToString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExecute(t *testing.T) {
	type args struct {
		json     io.Reader
		template string
		colorize bool
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "color",
			args: args{
				json:     strings.NewReader(`{}`),
				template: `{{color "blue+h" "songs are like tattoos"}}`,
			},
			wantW: "\x1b[0;94msongs are like tattoos\x1b[0m",
		},
		{
			name: "autocolor enabled",
			args: args{
				json:     strings.NewReader(`{}`),
				template: `{{autocolor "red" "stop"}}`,
				colorize: true,
			},
			wantW: "\x1b[0;31mstop\x1b[0m",
		},
		{
			name: "autocolor disabled",
			args: args{
				json:     strings.NewReader(`{}`),
				template: `{{autocolor "red" "go"}}`,
			},
			wantW: "go",
		},
		{
			name: "timefmt",
			args: args{
				json:     strings.NewReader(`{"created_at":"2008-02-25T20:18:33Z"}`),
				template: `{{.created_at | timefmt "Mon Jan 2, 2006"}}`,
			},
			wantW: "Mon Feb 25, 2008",
		},
		{
			name: "timeago",
			args: args{
				json:     strings.NewReader(fmt.Sprintf(`{"created_at":"%s"}`, time.Now().Add(-5*time.Minute).Format(time.RFC3339))),
				template: `{{.created_at | timeago}}`,
			},
			wantW: "5 minutes ago",
		},
		{
			name: "pluck",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{"name": "bug"},
					{"name": "feature request"},
					{"name": "chore"}
				]`)),
				template: `{{range(pluck "name" .)}}{{. | printf "%s\n"}}{{end}}`,
			},
			wantW: "bug\nfeature request\nchore\n",
		},
		{
			name: "join",
			args: args{
				json:     strings.NewReader(`[ "bug", "feature request", "chore" ]`),
				template: `{{join "\t" .}}`,
			},
			wantW: "bug\tfeature request\tchore",
		},
		{
			name: "table",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{"number": 1, "title": "One"},
					{"number": 20, "title": "Twenty"},
					{"number": 3000, "title": "Three thousand"}
				]`)),
				template: `{{range .}}{{tablerow (.number | printf "#%v") .title}}{{end}}`,
			},
			wantW: heredoc.Doc(`#1     One
			#20    Twenty
			#3000  Three thousand
			`),
		},
		{
			name: "table with multiline text",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{"number": 1, "title": "One\ranother line of text"},
					{"number": 20, "title": "Twenty\nanother line of text"},
					{"number": 3000, "title": "Three thousand\r\nanother line of text"}
				]`)),
				template: `{{range .}}{{tablerow (.number | printf "#%v") .title}}{{end}}`,
			},
			wantW: heredoc.Doc(`#1     One...
			#20    Twenty...
			#3000  Three thousand...
			`),
		},
		{
			name: "table with mixed value types",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{"number": 1, "title": null, "float": false},
					{"number": 20.1, "title": "Twenty-ish", "float": true},
					{"number": 3000, "title": "Three thousand", "float": false}
				]`)),
				template: `{{range .}}{{tablerow .number .title .float}}{{end}}`,
			},
			wantW: heredoc.Doc(`1                      false
			20.10  Twenty-ish      true
			3000   Three thousand  false
			`),
		},
		{
			name: "table with color",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{"number": 1, "title": "One"}
				]`)),
				template: `{{range .}}{{tablerow (.number | color "green") .title}}{{end}}`,
			},
			wantW: "\x1b[0;32m1\x1b[0m  One\n",
		},
		{
			name: "table with header and footer",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{"number": 1, "title": "One"},
					{"number": 2, "title": "Two"}
				]`)),
				template: heredoc.Doc(`HEADER
				{{range .}}{{tablerow .number .title}}{{end}}FOOTER
				`),
			},
			wantW: heredoc.Doc(`HEADER
			FOOTER
			1  One
			2  Two
			`),
		},
		{
			name: "table with header and footer using endtable",
			args: args{
				json: strings.NewReader(heredoc.Doc(`[
					{"number": 1, "title": "One"},
					{"number": 2, "title": "Two"}
				]`)),
				template: heredoc.Doc(`HEADER
				{{range .}}{{tablerow .number .title}}{{end}}{{tablerender}}FOOTER
				`),
			},
			wantW: heredoc.Doc(`HEADER
			1  One
			2  Two
			FOOTER
			`),
		},
		{
			name: "multiple tables with different columns",
			args: args{
				json: strings.NewReader(heredoc.Doc(`{
					"issues": [
						{"number": 1, "title": "One"},
						{"number": 2, "title": "Two"}
					],
					"prs": [
						{"number": 3, "title": "Three", "reviewDecision": "REVIEW_REQUESTED"},
						{"number": 4, "title": "Four", "reviewDecision": "CHANGES_REQUESTED"}
					]
				}`)),
				template: heredoc.Doc(`{{tablerow "ISSUE" "TITLE"}}{{range .issues}}{{tablerow .number .title}}{{end}}{{tablerender}}
				{{tablerow "PR" "TITLE" "DECISION"}}{{range .prs}}{{tablerow .number .title .reviewDecision}}{{end}}`),
			},
			wantW: heredoc.Docf(`ISSUE  TITLE
			1      One
			2      Two

			PR  TITLE  DECISION
			3   Three  REVIEW_REQUESTED
			4   Four   CHANGES_REQUESTED
			`),
		},
		{
			name: "truncate",
			args: args{
				json:     strings.NewReader(`{"title": "This is a long title"}`),
				template: `{{truncate 13 .title}}`,
			},
			wantW: "This is a ...",
		},
		{
			name: "truncate with JSON null",
			args: args{
				json:     strings.NewReader(`{}`),
				template: `{{ truncate 13 .title }}`,
			},
			wantW: "",
		},
		{
			name: "truncate with piped JSON null",
			args: args{
				json:     strings.NewReader(`{}`),
				template: `{{ .title | truncate 13 }}`,
			},
			wantW: "",
		},
		{
			name: "truncate with piped JSON null in parenthetical",
			args: args{
				json:     strings.NewReader(`{}`),
				template: `{{ (.title | truncate 13) }}`,
			},
			wantW: "",
		},
		{
			name: "truncate invalid type",
			args: args{
				json:     strings.NewReader(`{"title": 42}`),
				template: `{{ (.title | truncate 13) }}`,
			},
			wantErr: true,
		},
		{
			name: "hyperlink enabled",
			args: args{
				json:     strings.NewReader(`{"link":"https://github.com"}`),
				template: `{{ hyperlink .link "" }}`,
			},
			wantW: "\x1b]8;;https://github.com\x1b\\https://github.com\x1b]8;;\x1b\\",
		},
		{
			name: "hyperlink with text enabled",
			args: args{
				json:     strings.NewReader(`{"link":"https://github.com","text":"GitHub"}`),
				template: `{{ hyperlink .link .text }}`,
			},
			wantW: "\x1b]8;;https://github.com\x1b\\GitHub\x1b]8;;\x1b\\",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tmpl := New(w, 80, tt.args.colorize)
			err := tmpl.Parse(tt.args.template)
			assert.NoError(t, err)
			err = tmpl.Execute(tt.args.json)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			err = tmpl.Flush()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}

func TestExecuteFiltered(t *testing.T) {
	type args struct {
		json     io.Reader
		template string
		filter   string
		colorize bool
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "without filter",
			args: args{
				json:     strings.NewReader(`[{"name":"b"},{"name":"a"}]`),
				template: `{{range .}}{{.name}}{{end}}`,
			},
			wantW: "ba",
		},
		{
			name: "with filter",
			args: args{
				json:     strings.NewReader(`[{"name":"b"},{"name":"a"}]`),
				filter:   `sort_by(.name)`,
				template: `{{range .}}{{.name}}{{end}}`,
			},
			wantW: "ab",
		},
		{
			name: "color with filter",
			args: args{
				json:     strings.NewReader(`{"error":"an error occurred"}`),
				filter:   `{error: .error | ascii_upcase}`,
				template: `{{color "red" .error}}`,
			},
			wantW: "\x1b[0;31mAN ERROR OCCURRED\x1b[0m",
		},
		{
			name: "filter error",
			args: args{
				json:     strings.NewReader(`{}`),
				filter:   "nofunc",
				template: "{{.}}",
			},
			wantErr: true,
		},
		{
			name: "template error",
			args: args{
				json:     strings.NewReader(`{}`),
				template: `{{ truncate }}`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tmpl := New(w, 80, tt.args.colorize)
			err := tmpl.Parse(tt.args.template)
			assert.NoError(t, err)
			err = tmpl.ExecuteFiltered(tt.args.json, tt.args.filter)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			err = tmpl.Flush()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}

func TestTruncateMultiline(t *testing.T) {
	type args struct {
		max int
		s   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "exactly minimum width",
			args: args{
				max: 5,
				s:   "short",
			},
			want: "short",
		},
		{
			name: "exactly minimum width with new line",
			args: args{
				max: 5,
				s:   "short\n",
			},
			want: "sh...",
		},
		{
			name: "less than minimum width",
			args: args{
				max: 4,
				s:   "short",
			},
			want: "shor",
		},
		{
			name: "less than minimum width with new line",
			args: args{
				max: 4,
				s:   "short\n",
			},
			want: "shor",
		},
		{
			name: "first line of multiple is short enough",
			args: args{
				max: 80,
				s:   "short\n\nthis is a new line",
			},
			want: "short...",
		},
		{
			name: "using Windows line endings",
			args: args{
				max: 80,
				s:   "short\r\n\r\nthis is a new line",
			},
			want: "short...",
		},
		{
			name: "using older MacOS line endings",
			args: args{
				max: 80,
				s:   "short\r\rthis is a new line",
			},
			want: "short...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateMultiline(tt.args.max, tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFuncs(t *testing.T) {
	w := &bytes.Buffer{}
	tmpl := New(w, 80, false)

	// Override "truncate" and define a new "foo" function.
	tmpl.Funcs(map[string]interface{}{
		"truncate": func(fields ...interface{}) (string, error) {
			if l := len(fields); l != 2 {
				return "", fmt.Errorf("wrong number of args for truncate: want 2 got %d", l)
			}
			var ok bool
			var width int
			var input string
			if width, ok = fields[0].(int); !ok {
				return "", fmt.Errorf("invalid value; expected int")
			}
			if input, ok = fields[1].(string); !ok {
				return "", fmt.Errorf("invalid value; expected string")
			}
			return input[:width], nil
		},
		"foo": func(fields ...interface{}) (string, error) {
			return "test", nil
		},
	})

	err := tmpl.Parse(`{{ .text | truncate 5 }} {{ .status | color "green" }} {{ foo }}`)
	assert.NoError(t, err)

	r := strings.NewReader(`{"text":"truncated","status":"open"}`)
	err = tmpl.Execute(r)
	assert.NoError(t, err)

	err = tmpl.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "trunc \x1b[0;32mopen\x1b[0m test", w.String())
}
