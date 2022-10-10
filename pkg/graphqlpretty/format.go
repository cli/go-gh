package graphqlpretty

import (
	"fmt"
	"io"
	"strings"
	"text/scanner"
	"unicode"
)

const (
	colorDelim  = "\x1b[1;38m" // bright white
	colorKey    = "\x1b[1;34m" // bright blue
	colorNull   = "\x1b[36m"   // cyan
	colorString = "\x1b[32m"   // green
	colorBool   = "\x1b[33m"   // yellow
	colorReset  = "\x1b[m"
)

// Format reads a GraphQL query from r and writes a prettified version of it to w.
func Format(w io.Writer, r io.Reader, indent string, colorize bool) error {
	s := scanner.Scanner{}
	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanComments | scanner.ScanStrings
	s.IsIdentRune = func(ch rune, i int) bool {
		return i == 0 && ch == '$' ||
			ch == '_' ||
			unicode.IsLetter(ch) ||
			unicode.IsDigit(ch) && i > 0 ||
			i > 0 && ch == '!'
	}

	c := func(ansi string) string {
		if !colorize {
			return ""
		}
		return ansi
	}

	var nesting []rune
	inBrackets := func() bool {
		for i := len(nesting) - 1; i >= 0; i-- {
			if nesting[i] == '(' {
				return true
			}
		}
		return false
	}
	keyNeedsValue := false
	pad := func() string {
		if keyNeedsValue {
			keyNeedsValue = false
			return " "
		}
		if inBrackets() {
			return " "
		}
		if len(nesting) == 0 {
			return " "
		}
		return "\n" + strings.Repeat(indent, len(nesting))
	}

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		switch tok {
		case scanner.String:
			if _, err := fmt.Fprint(w, pad(), c(colorString), s.TokenText(), c(colorReset)); err != nil {
				return err
			}
		case scanner.Ident:
			if _, err := fmt.Fprint(w, pad(), c(colorKey), s.TokenText(), c(colorReset)); err != nil {
				return err
			}
		case scanner.Int, scanner.Float:
			if _, err := fmt.Fprint(w, pad(), s.TokenText()); err != nil {
				return err
			}
		case '{', '}', '[', ']', '(', ')':
			if tok == '}' || tok == ']' || tok == ')' {
				if len(nesting) > 0 {
					nesting = nesting[:len(nesting)-1]
				}
			}
			var padding string
			if tok == '}' || tok == ']' {
				padding = pad()
			}
			if _, err := fmt.Fprint(w, padding, c(colorDelim), s.TokenText(), c(colorReset)); err != nil {
				return err
			}
			if tok == '{' || tok == '[' || tok == '(' {
				nesting = append(nesting, tok)
			}
		case ':':
			if _, err := fmt.Fprint(w, string(tok)); err != nil {
				return err
			}
			keyNeedsValue = true
		case '=':
			if _, err := fmt.Fprint(w, " ", string(tok)); err != nil {
				return err
			}
		}
	}

	_, err := fmt.Fprint(w, "\n")
	return err
}
