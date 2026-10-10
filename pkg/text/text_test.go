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
		{
			// A ZWJ sequence occupies two columns however many people are in it,
			// so eight of the eleven are still available for the text after it.
			name: "Emoji ZWJ sequence",
			args: args{
				max: 11,
				s:   "👨‍👩‍👧‍👦 family names here",
			},
			want: "👨‍👩‍👧‍👦 famil...",
		},
		{
			name: "Regional indicator pair",
			args: args{
				max: 11,
				s:   "🇯🇵 Japanese README",
			},
			want: "🇯🇵 Japan...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.args.max, tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateFillsTheBudget(t *testing.T) {
	// Truncate measures with DisplayWidth to decide whether a string is too
	// wide, so the result has to be measured the same way: never wider than the
	// budget, and never narrower than it once the string really was too long.
	// Mixing two width models silently loses columns, which shows up as ragged
	// table output for anyone whose text is not ASCII.
	corpus := []string{
		"hello world, a plain ascii string",
		"テストテストテストテスト",
		"幫新舉報違章工廠新增編號",
		"프로젝트 내의",
		"💡💡💡💡💡💡💡💡💡💡💡💡",
		"👨‍👩‍👧‍👦 family names here",
		"🇯🇵 Japanese README",
		"🏳️‍🌈 pride flag entry",
		"👍🏽 thumbs up with a skin tone",
		"❤️ reactions on issues",
		"é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́é́́",
		"\x1b[0;31mcoloured text that is long\x1b[0m",
	}
	for _, s := range corpus {
		for maxWidth := 1; maxWidth <= 24; maxWidth++ {
			got := Truncate(maxWidth, s)
			gotWidth := DisplayWidth(got)
			assert.LessOrEqual(t, gotWidth, maxWidth,
				"Truncate(%d, %q) = %q is wider than the budget", maxWidth, s, got)
			if DisplayWidth(s) > maxWidth {
				assert.Equal(t, maxWidth, gotWidth,
					"Truncate(%d, %q) = %q leaves columns unused", maxWidth, s, got)
			}
		}
	}
}

func TestPadRight(t *testing.T) {
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
				max: 5,
			},
			want: "     ",
		},
		{
			name: "short",
			args: args{
				s:   "hello",
				max: 7,
			},
			want: "hello  ",
		},
		{
			name: "long",
			args: args{
				s:   "hello world",
				max: 5,
			},
			want: "hello world",
		},
		{
			name: "exact",
			args: args{
				s:   "hello world",
				max: 11,
			},
			want: "hello world",
		},
		{
			name: "Japanese",
			args: args{
				s:   "テストテスト",
				max: 13,
			},
			want: "テストテスト ",
		},
		{
			name: "Japanese filled",
			args: args{
				s:   "aテスト",
				max: 9,
			},
			want: "aテスト  ",
		},
		{
			name: "Chinese",
			args: args{
				s:   "幫新舉報違章工廠新增編號",
				max: 26,
			},
			want: "幫新舉報違章工廠新增編號  ",
		},
		{
			name: "Chinese filled",
			args: args{
				s:   "a幫新舉報違章工廠新增編號",
				max: 26,
			},
			want: "a幫新舉報違章工廠新增編號 ",
		},
		{
			name: "Korean",
			args: args{
				s:   "프로젝트 내의",
				max: 15,
			},
			want: "프로젝트 내의  ",
		},
		{
			name: "Korean filled",
			args: args{
				s:   "a프로젝트 내의",
				max: 15,
			},
			want: "a프로젝트 내의 ",
		},
		{
			name: "Emoji",
			args: args{
				s:   "💡💡💡💡",
				max: 10,
			},
			want: "💡💡💡💡  ",
		},
		{
			name: "Accented characters",
			args: args{
				s:   "é́́é́́é́́é́́é́́",
				max: 7,
			},
			want: "é́́é́́é́́é́́é́́  ",
		},
		{
			name: "Red accented characters",
			args: args{
				s:   "\x1b[0;31mé́́é́́é́́é́́é́́\x1b[0m",
				max: 7,
			},
			want: "\x1b[0;31mé́́é́́é́́é́́é́́\x1b[0m  ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PadRight(tt.args.max, tt.args.s)
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
			name: "multi-byte emoji",
			text: `👨‍💻`,
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

func TestRemoveDiacritics(t *testing.T) {
	tests := [][]string{
		// no diacritics
		{"e", "e"},
		{"و", "و"},
		{"И", "И"},
		{"ж", "ж"},
		{"私", "私"},
		{"万", "万"},

		// diacritics test sets
		{"à", "a"},
		{"é", "e"},
		{"è", "e"},
		{"ô", "o"},
		{"ᾳ", "α"},
		{"εͅ", "ε"},
		{"ῃ", "η"},
		{"ιͅ", "ι"},

		{"ؤ", "و"},

		{"ā", "a"},
		{"č", "c"},
		{"ģ", "g"},
		{"ķ", "k"},
		{"ņ", "n"},
		{"š", "s"},
		{"ž", "z"},

		{"ŵ", "w"},
		{"ŷ", "y"},
		{"ä", "a"},
		{"ÿ", "y"},
		{"á", "a"},
		{"ẁ", "w"},
		{"ỳ", "y"},
		{"ō", "o"},

		// full words
		{"Miķelis", "Mikelis"},
		{"François", "Francois"},
		{"žluťoučký", "zlutoucky"},
		{"învățătorița", "invatatorita"},
		{"Kękę przy łóżku", "Keke przy łozku"},
	}

	for _, tt := range tests {
		t.Run(RemoveDiacritics(tt[0]), func(t *testing.T) {
			assert.Equal(t, tt[1], RemoveDiacritics(tt[0]))
		})
	}
}
