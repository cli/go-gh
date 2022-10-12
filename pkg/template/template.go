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
	"github.com/cli/go-gh/pkg/text"
	color "github.com/mgutz/ansi"
)

const (
	ellipsis = "..."
)

// Template is the representation of a template.
type Template struct {
	colorEnabled bool
	output       io.Writer
	tmpl         *template.Template
	tp           tableprinter.TablePrinter
	width        int
	funcs        template.FuncMap
}

// New initializes a Template.
func New(w io.Writer, width int, colorEnabled bool) Template {
	return Template{
		colorEnabled: colorEnabled,
		output:       w,
		tp:           tableprinter.New(w, true, width),
		width:        width,
		funcs:        template.FuncMap{},
	}
}

// Funcs adds the elements of the argument map to the template's function map.
// It must be called before the template is parsed.
// It is legal to overwrite elements of the map including default functions.
// The return value is the template, so calls can be chained.
func (t *Template) Funcs(funcMap map[string]interface{}) *Template {
	for name, f := range funcMap {
		t.funcs[name] = f
	}
	return t
}

// Parse the given template string for use with Execute.
func (t *Template) Parse(tmpl string) error {
	now := time.Now()
	templateFuncs := map[string]interface{}{
		"autocolor": colorFunc,
		"color":     colorFunc,
		"hyperlink": hyperlinkFunc,
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
	for name, f := range t.funcs {
		templateFuncs[name] = f
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
		return text.Truncate(maxWidth, s), nil
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
		tp.AddField(s, tableprinter.WithTruncate(truncateMultiline))
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
		return text.Pluralize(int(ago.Minutes()), "minute") + " ago"
	}
	if ago < 24*time.Hour {
		return text.Pluralize(int(ago.Hours()), "hour") + " ago"
	}
	if ago < 30*24*time.Hour {
		return text.Pluralize(int(ago.Hours())/24, "day") + " ago"
	}
	if ago < 365*24*time.Hour {
		return text.Pluralize(int(ago.Hours())/24/30, "month") + " ago"
	}
	return text.Pluralize(int(ago.Hours()/24/365), "year") + " ago"
}

// TruncateMultiline returns a copy of the string s that has been shortened to fit the maximum
// display width. If string s has multiple lines the first line will be shortened and all others
// removed.
func truncateMultiline(maxWidth int, s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i] + ellipsis
	}
	return text.Truncate(maxWidth, s)
}

func hyperlinkFunc(link, text string) string {
	if text == "" {
		text = link
	}

	// See https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", link, text)
}
