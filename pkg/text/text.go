// Package text is a set of utility functions for text processing and outputting to the terminal.
package text

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	ellipsis            = "..."
	minWidthForEllipsis = len(ellipsis) + 2
)

var indentRE = regexp.MustCompile(`(?m)^`)

// Indent returns a copy of the string s with indent prefixed to it, will apply indent
// to each line of the string.
func Indent(s, indent string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return s
	}
	return indentRE.ReplaceAllLiteralString(s, indent)
}

// DisplayWidth calculates what the rendered width of string s will be.
func DisplayWidth(s string) int {
	return ansi.PrintableRuneWidth(s)
}

// Truncate returns a copy of the string s that has been shortened to fit the maximum display width.
func Truncate(maxWidth int, s string) string {
	w := DisplayWidth(s)
	if w <= maxWidth {
		return s
	}
	tail := ""
	if maxWidth >= minWidthForEllipsis {
		tail = ellipsis
	}
	r := truncate.StringWithTail(s, uint(maxWidth), tail)
	if DisplayWidth(r) < maxWidth {
		r += " "
	}
	return r
}

// PadRight returns a copy of the string s that has been padded on the right with whitespace to fit
// the maximum display width.
func PadRight(maxWidth int, s string) string {
	if padWidth := maxWidth - DisplayWidth(s); padWidth > 0 {
		s += strings.Repeat(" ", padWidth)
	}
	return s
}

// Pluralize returns a concatenated string with num and the plural form of thing if necessary.
func Pluralize(num int, thing string) string {
	if num == 1 {
		return fmt.Sprintf("%d %s", num, thing)
	}
	return fmt.Sprintf("%d %ss", num, thing)
}

func fmtDuration(amount int, unit string) string {
	return fmt.Sprintf("about %s ago", Pluralize(amount, unit))
}

// RelativeTimeAgo returns a human readable string of the time duration between a and b that is estimated
// to the nearest unit of time.
func RelativeTimeAgo(a, b time.Time) string {
	ago := a.Sub(b)

	if ago < time.Minute {
		return "less than a minute ago"
	}
	if ago < time.Hour {
		return fmtDuration(int(ago.Minutes()), "minute")
	}
	if ago < 24*time.Hour {
		return fmtDuration(int(ago.Hours()), "hour")
	}
	if ago < 30*24*time.Hour {
		return fmtDuration(int(ago.Hours())/24, "day")
	}
	if ago < 365*24*time.Hour {
		return fmtDuration(int(ago.Hours())/24/30, "month")
	}

	return fmtDuration(int(ago.Hours()/24/365), "year")
}

// RemoveDiacritics returns the input value without "diacritics", or accent marks.
func RemoveDiacritics(value string) string {
	// Mn = "Mark, nonspacing" unicode character category
	removeMnTransfomer := runes.Remove(runes.In(unicode.Mn))

	// 1. Decompose the text into characters and diacritical marks
	// 2. Remove the diacriticals marks
	// 3. Recompose the text
	t := transform.Chain(norm.NFD, removeMnTransfomer, norm.NFC)
	normalized, _, err := transform.String(t, value)
	if err != nil {
		return value
	}
	return normalized
}
