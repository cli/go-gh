package tableprinter

import (
	"bytes"
	"log"
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
