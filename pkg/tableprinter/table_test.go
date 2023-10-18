package tableprinter

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
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
		log.Fatal(err)
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

func Test_ttyTablePrinter_AddHeader(t *testing.T) {
	buf := bytes.Buffer{}
	tp := New(&buf, true, 80)

	tp.AddHeader([]string{"ONE", "TWO", "THREE"}, WithColor(func(s string) string {
		return fmt.Sprintf("\x1b[4m%s\x1b[m", s)
	}))
	tp.AddHeader([]string{"SHOULD", "NOT", "EXIST"})

	tp.AddField("hello")
	tp.AddField("beautiful")
	tp.AddField("people")
	tp.EndRow()

	err := tp.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := heredoc.Docf(`
		%[1]s[4mONE  %[1]s[m  %[1]s[4mTWO      %[1]s[m  %[1]s[4mTHREE%[1]s[m
		hello  beautiful  people
	`, "\x1b")
	if buf.String() != expected {
		t.Errorf("expected: %q, got: %q", expected, buf.String())
	}
}

func Test_ttyTablePrinter_WithPadding(t *testing.T) {
	buf := bytes.Buffer{}
	tp := New(&buf, true, 80)

	// Center the headers.
	tp.AddHeader([]string{"A", "B", "C"}, WithPadding(func(width int, s string) string {
		left := (width - len(s)) / 2
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", width-left-len(s))
	}))

	tp.AddField("hello")
	tp.AddField("beautiful")
	tp.AddField("people")
	tp.EndRow()

	err := tp.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := heredoc.Doc(`
		  A        B        C   
		hello  beautiful  people
	`)
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
