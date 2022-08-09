package tableprinter

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func ExampleTablePrinter() {
	// information about the terminal can be obtained using the [pkg/term] package
	isTTY := true
	termWidth := 14
	red := func(s string) string {
		return "\x1b[31m" + s + "\x1b[m"
	}

	t := New(os.Stdout, isTTY, termWidth)
	t.AddField("9", WithTruncate(nil))
	t.AddField("hello")
	t.EndRow()
	t.AddField("10", WithTruncate(nil))
	t.AddField("long description", WithColor(red))
	t.EndRow()
	if err := t.Render(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// stdout now contains:
	// 9   hello
	// 10  long de...
}

func Test_ttyTablePrinter_autoTruncate(t *testing.T) {
	buf := bytes.Buffer{}
	tp := New(&buf, true, 5)

	tp.AddField("1")
	tp.AddField("hello")
	tp.EndRow()
	tp.AddField("2")
	tp.AddField("world")
	tp.EndRow()

	err := tp.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1  he\n2  wo\n"
	if buf.String() != expected {
		t.Errorf("expected: %q, got: %q", expected, buf.String())
	}
}

func Test_ttyTablePrinter_WithTruncate(t *testing.T) {
	buf := bytes.Buffer{}
	tp := New(&buf, true, 15)

	tp.AddField("long SHA", WithTruncate(nil))
	tp.AddField("hello")
	tp.EndRow()
	tp.AddField("another SHA", WithTruncate(nil))
	tp.AddField("world")
	tp.EndRow()

	err := tp.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "long SHA     he\nanother SHA  wo\n"
	if buf.String() != expected {
		t.Errorf("expected: %q, got: %q", expected, buf.String())
	}
}

func Test_tsvTablePrinter(t *testing.T) {
	buf := bytes.Buffer{}
	tp := New(&buf, false, 0)

	tp.AddField("1")
	tp.AddField("hello")
	tp.EndRow()
	tp.AddField("2")
	tp.AddField("world")
	tp.EndRow()

	err := tp.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1\thello\n2\tworld\n"
	if buf.String() != expected {
		t.Errorf("expected: %q, got: %q", expected, buf.String())
	}
}

func Test_truncateText(t *testing.T) {
	type args struct {
		maxWidth int
		s        string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				s:        "",
				maxWidth: 10,
			},
			want: "",
		},
		{
			name: "short",
			args: args{
				s:        "hello",
				maxWidth: 3,
			},
			want: "hel",
		},
		{
			name: "long",
			args: args{
				s:        "hello world",
				maxWidth: 5,
			},
			want: "he...",
		},
		{
			name: "no truncate",
			args: args{
				s:        "hello world",
				maxWidth: 11,
			},
			want: "hello world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateText(tt.args.maxWidth, tt.args.s); got != tt.want {
				t.Errorf("truncateText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_displayWidth(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "empty",
			args: args{
				s: "",
			},
			want: 0,
		},
		{
			name: "Latin",
			args: args{
				s: "hello world 123$#!",
			},
			want: 18,
		},
		{
			name: "Asian",
			args: args{
				s: "つのだ☆HIRO",
			},
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayWidth(tt.args.s); got != tt.want {
				t.Errorf("displayWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}
