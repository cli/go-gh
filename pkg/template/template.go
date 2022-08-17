// Package template facilitates processing of JSON strings using Go templates.
// Provides additional functions not available using basic Go templates, such as coloring,
// and table rendering.
package template

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cli/go-gh/pkg/tableprinter"
	color "github.com/mgutz/ansi"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
)

// Template is the representation of a template.
type Template struct {
	colorEnabled bool
	output       io.Writer
	tmpl         *template.Template
	tp           tableprinter.TablePrinter
	width        int
}

// New initializes a Template.
func New(w io.Writer, width int, colorEnabled bool) Template {
	return Template{
		colorEnabled: colorEnabled,
		output:       w,
		tp:           tableprinter.New(w, true, width),
		width:        width,
	}
}

// Parse the given template string for use with Execute.
func (t *Template) Parse(tmpl string) error {
	now := time.Now()
	templateFuncs := map[string]interface{}{
		"autocolor": colorFunc,
		"color":     colorFunc,
		"join":      joinFunc,
		"pluck":     pluckFunc,
		"tablerender": func() (string, error) {
			// After rendering a table, prepare a new table printer incase user wants to output
			// another table.
			defer func() {
				t.tp = tableprinter.New(t.output, true, t.width)
			}()
			return tableRenderFunc(t.tp)
		},
		"tablerow": func(fields ...interface{}) (string, error) {
			return tableRowFunc(t.tp, fields...)
		},
		"timeago": func(input string) (string, error) {
			return timeAgoFunc(now, input)
		},
		"timefmt":  timeFormatFunc,
		"truncate": truncateFunc,
	}
	if !t.colorEnabled {
		templateFuncs["autocolor"] = autoColorFunc
	}
	var err error
	t.tmpl, err = template.New("").Funcs(templateFuncs).Parse(tmpl)
	return err
}

// Execute applies the parsed template to the input and writes result to the writer
// the template was initialized with.
func (t *Template) Execute(input io.Reader) error {
	jsonData, err := io.ReadAll(input)
	if err != nil {
		return err
	}

	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	return t.tmpl.Execute(t.output, data)
}

// Flush writes any remaining data to the writer. This is mostly useful
// when a templates uses the tablerow function but does not include the
// tablerender function at the end.
// If a template did not use the table functionality this is a noop.
func (t *Template) Flush() error {
	if _, err := tableRenderFunc(t.tp); err != nil {
		return err
	}
	return nil
}

func colorFunc(colorName string, input interface{}) (string, error) {
	text, err := jsonScalarToString(input)
	if err != nil {
		return "", err
	}
	return color.Color(text, colorName), nil
}

func pluckFunc(field string, input []interface{}) []interface{} {
	var results []interface{}
	for _, item := range input {
		obj := item.(map[string]interface{})
		results = append(results, obj[field])
	}
	return results
}

func joinFunc(sep string, input []interface{}) (string, error) {
	var results []string
	for _, item := range input {
		text, err := jsonScalarToString(item)
		if err != nil {
			return "", err
		}
		results = append(results, text)
	}
	return strings.Join(results, sep), nil
}

func timeFormatFunc(format, input string) (string, error) {
	t, err := time.Parse(time.RFC3339, input)
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

func timeAgoFunc(now time.Time, input string) (string, error) {
	t, err := time.Parse(time.RFC3339, input)
	if err != nil {
		return "", err
	}
	return timeAgo(now.Sub(t)), nil
}

func truncateFunc(maxWidth int, v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	if s, ok := v.(string); ok {
		return truncateText(maxWidth, s), nil
	}
	return "", fmt.Errorf("invalid value; expected string, got %T", v)
}

func autoColorFunc(colorName string, input interface{}) (string, error) {
	return jsonScalarToString(input)
}

func tableRowFunc(tp tableprinter.TablePrinter, fields ...interface{}) (string, error) {
	if tp == nil {
		return "", fmt.Errorf("failed to write table row: no table printer")
	}
	for _, e := range fields {
		s, err := jsonScalarToString(e)
		if err != nil {
			return "", fmt.Errorf("failed to write table row: %v", err)
		}
		tp.AddField(s, tableprinter.WithTruncate(truncateColumn))
	}
	tp.EndRow()
	return "", nil
}

func tableRenderFunc(tp tableprinter.TablePrinter) (string, error) {
	if tp == nil {
		return "", fmt.Errorf("failed to render table: no table printer")
	}
	err := tp.Render()
	if err != nil {
		return "", fmt.Errorf("failed to render table: %v", err)
	}
	return "", nil
}

func jsonScalarToString(input interface{}) (string, error) {
	switch tt := input.(type) {
	case string:
		return tt, nil
	case float64:
		if math.Trunc(tt) == tt {
			return strconv.FormatFloat(tt, 'f', 0, 64), nil
		} else {
			return strconv.FormatFloat(tt, 'f', 2, 64), nil
		}
	case nil:
		return "", nil
	case bool:
		return fmt.Sprintf("%v", tt), nil
	default:
		return "", fmt.Errorf("cannot convert type to string: %v", tt)
	}
}

func timeAgo(ago time.Duration) string {
	if ago < time.Minute {
		return "just now"
	}
	if ago < time.Hour {
		return pluralize(int(ago.Minutes()), "minute") + " ago"
	}
	if ago < 24*time.Hour {
		return pluralize(int(ago.Hours()), "hour") + " ago"
	}
	if ago < 30*24*time.Hour {
		return pluralize(int(ago.Hours())/24, "day") + " ago"
	}
	if ago < 365*24*time.Hour {
		return pluralize(int(ago.Hours())/24/30, "month") + " ago"
	}
	return pluralize(int(ago.Hours()/24/365), "year") + " ago"
}

func pluralize(num int, thing string) string {
	if num == 1 {
		return fmt.Sprintf("%d %s", num, thing)
	}
	return fmt.Sprintf("%d %ss", num, thing)
}

// TruncateColumn replaces the first new line character with an ellipsis
// and shortens a string to fit the maximum display width.
func truncateColumn(maxWidth int, s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i] + "..."
	}
	return truncateText(maxWidth, s)
}

func truncateText(maxWidth int, s string) string {
	rw := ansi.PrintableRuneWidth(s)
	if rw <= maxWidth {
		return s
	}
	if maxWidth < 5 {
		return truncate.String(s, uint(maxWidth))
	}
	return truncate.StringWithTail(s, uint(maxWidth), "...")
}
