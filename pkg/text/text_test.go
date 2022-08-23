package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFuzzyAgo(t *testing.T) {
	const form = "2006-Jan-02 15:04:05"
	now, _ := time.Parse(form, "2020-Nov-22 14:00:00")
	cases := map[string]string{
		"2020-Nov-22 14:00:00": "less than a minute ago",
		"2020-Nov-22 13:59:30": "less than a minute ago",
		"2020-Nov-22 13:59:00": "about 1 minute ago",
		"2020-Nov-22 13:30:00": "about 30 minutes ago",
		"2020-Nov-22 13:00:00": "about 1 hour ago",
		"2020-Nov-22 02:00:00": "about 12 hours ago",
		"2020-Nov-21 14:00:00": "about 1 day ago",
		"2020-Nov-07 14:00:00": "about 15 days ago",
		"2020-Oct-24 14:00:00": "about 29 days ago",
		"2020-Oct-23 14:00:00": "about 1 month ago",
		"2020-Sep-23 14:00:00": "about 2 months ago",
		"2019-Nov-22 14:00:00": "about 1 year ago",
		"2018-Nov-22 14:00:00": "about 2 years ago",
	}
	for createdAt, expected := range cases {
		d, err := time.Parse(form, createdAt)
		assert.NoError(t, err)
		fuzzy := FuzzyAgo(now, d)
		assert.Equal(t, expected, fuzzy)
	}
}

func TestFuzzyAgoAbbr(t *testing.T) {
	const form = "2006-Jan-02 15:04:05"
	now, _ := time.Parse(form, "2020-Nov-22 14:00:00")
	cases := map[string]string{
		"2020-Nov-22 14:00:00": "0m",
		"2020-Nov-22 13:59:00": "1m",
		"2020-Nov-22 13:30:00": "30m",
		"2020-Nov-22 13:00:00": "1h",
		"2020-Nov-22 02:00:00": "12h",
		"2020-Nov-21 14:00:00": "1d",
		"2020-Nov-07 14:00:00": "15d",
		"2020-Oct-24 14:00:00": "29d",
		"2020-Oct-23 14:00:00": "Oct 23, 2020",
		"2019-Nov-22 14:00:00": "Nov 22, 2019",
	}
	for createdAt, expected := range cases {
		d, err := time.Parse(form, createdAt)
		assert.NoError(t, err)
		fuzzy := FuzzyAgoAbbr(now, d)
		assert.Equal(t, expected, fuzzy)
	}
}

func ExampleTruncate() {

}

func TestTruncate(t *testing.T) {
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
			name: "empty",
			args: args{
				s:   "",
				max: 10,
			},
			want: "",
		},
		{
			name: "short",
			args: args{
				s:   "hello",
				max: 3,
			},
			want: "hel",
		},
		{
			name: "long",
			args: args{
				s:   "hello world",
				max: 5,
			},
			want: "he...",
		},
		{
			name: "no truncate",
			args: args{
				s:   "hello world",
				max: 11,
			},
			want: "hello world",
		},
		{
			name: "Short enough",
			args: args{
				max: 5,
				s:   "short",
			},
			want: "short",
		},
		{
			name: "Too short",
			args: args{
				max: 4,
				s:   "short",
			},
			want: "shor",
		},
		{
			name: "Japanese",
			args: args{
				max: 11,
				s:   "テストテストテストテスト",
			},
			want: "テストテ...",
		},
		{
			name: "Japanese filled",
			args: args{
				max: 11,
				s:   "aテストテストテストテスト",
			},
			want: "aテスト... ",
		},
		{
			name: "Chinese",
			args: args{
				max: 11,
				s:   "幫新舉報違章工廠新增編號",
			},
			want: "幫新舉報...",
		},
		{
			name: "Chinese filled",
			args: args{
				max: 11,
				s:   "a幫新舉報違章工廠新增編號",
			},
			want: "a幫新舉... ",
		},
		{
			name: "Korean",
			args: args{
				max: 11,
				s:   "프로젝트 내의",
			},
			want: "프로젝트...",
		},
		{
			name: "Korean filled",
			args: args{
				max: 11,
				s:   "a프로젝트 내의",
			},
			want: "a프로젝... ",
		},
		{
			name: "Emoji",
			args: args{
				max: 11,
				s:   "💡💡💡💡💡💡💡💡💡💡💡💡",
			},
			want: "💡💡💡💡...",
		},
		{
			name: "Accented characters",
			args: args{
				max: 11,
				s:   "é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́",
			},
			want: "é́́é́́é́́é́́é́́é́́é́́é́́...",
		},
		{
			name: "Red accented characters",
			args: args{
				max: 11,
				s:   "\x1b[0;31mé́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́\x1b[0m",
			},
			want: "\x1b[0;31mé́́é́́é́́é́́é́́é́́é́́é́́...\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.args.max, tt.args.s)
			assert.Equal(t, tt.want, got)
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
			got := TruncateMultiline(tt.args.max, tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "check mark",
			text: `✓`,
			want: 1,
		},
		{
			name: "bullet icon",
			text: `•`,
			want: 1,
		},
		{
			name: "middle dot",
			text: `·`,
			want: 1,
		},
		{
			name: "ellipsis",
			text: `…`,
			want: 1,
		},
		{
			name: "right arrow",
			text: `→`,
			want: 1,
		},
		{
			name: "smart double quotes",
			text: `“”`,
			want: 2,
		},
		{
			name: "smart single quotes",
			text: `‘’`,
			want: 2,
		},
		{
			name: "em dash",
			text: `—`,
			want: 1,
		},
		{
			name: "en dash",
			text: `–`,
			want: 1,
		},
		{
			name: "emoji",
			text: `👍`,
			want: 2,
		},
		{
			name: "accent character",
			text: `é́́`,
			want: 1,
		},
		{
			name: "color codes",
			text: "\x1b[0;31mred\x1b[0m",
			want: 3,
		},
		{
			name: "empty",
			text: "",
			want: 0,
		},
		{
			name: "Latin",
			text: "hello world 123$#!",
			want: 18,
		},
		{
			name: "Asian",
			text: "つのだ☆HIRO",
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DisplayWidth(tt.text)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveExcessiveWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "nothing to remove",
			input: "one two three",
			want:  "one two three",
		},
		{
			name:  "whitespace b-gone",
			input: "\n  one\n\t  two  three\r\n  ",
			want:  "one two three",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveExcessiveWhitespace(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCamelToKebab(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "single lowercase word",
			in:   "test",
			out:  "test",
		},
		{
			name: "multiple mixed words",
			in:   "testTestTest",
			out:  "test-test-test",
		},
		{
			name: "multiple uppercase words",
			in:   "TestTest",
			out:  "test-test",
		},
		{
			name: "multiple lowercase words",
			in:   "testtest",
			out:  "testtest",
		},
		{
			name: "multiple mixed words with number",
			in:   "test2Test",
			out:  "test2-test",
		},
		{
			name: "multiple lowercase words with number",
			in:   "test2test",
			out:  "test2test",
		},
		{
			name: "multiple lowercase words with dash",
			in:   "test-test",
			out:  "test-test",
		},
		{
			name: "multiple uppercase words with dash",
			in:   "Test-Test",
			out:  "test--test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, CamelToKebab(tt.in))
		})
	}
}

func TestIndent(t *testing.T) {
	type args struct {
		s      string
		indent string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				s:      "",
				indent: "--",
			},
			want: "",
		},
		{
			name: "blank",
			args: args{
				s:      "\n",
				indent: "--",
			},
			want: "\n",
		},
		{
			name: "indent",
			args: args{
				s:      "one\ntwo\nthree",
				indent: "--",
			},
			want: "--one\n--two\n--three",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Indent(tt.args.s, tt.args.indent)
			assert.Equal(t, tt.want, got)
		})
	}
}
