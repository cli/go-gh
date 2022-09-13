package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRelativeTimeAgo(t *testing.T) {
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
		relative := RelativeTimeAgo(now, d)
		assert.Equal(t, expected, relative)
	}
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
