// Package tableprinter facilitates rendering column-formatted data to a terminal and TSV-formatted data to
// a script or a file. It is suitable for presenting tabular data in a human-readable format that is
// guaranteed to fit within the given viewport, while at the same time offering the same data in a
// machine-readable format for scripts.
package tableprinter

import (
	"fmt"
	"io"

	"github.com/cli/go-gh/v2/pkg/text"
)

type fieldOption func(*tableField)

type TablePrinter interface {
	AddHeader([]string, ...fieldOption)
	AddField(string, ...fieldOption)
	EndRow()
	Render() error
}

// WithTruncate overrides the truncation function for the field. The function should transform a string
// argument into a string that fits within the given display width. The default behavior is to truncate the
// value by adding "..." in the end. The truncation function will be called before padding and coloring.
// Pass nil to disable truncation for this value.
func WithTruncate(fn func(int, string) string) fieldOption {
	return func(f *tableField) {
		f.truncateFunc = fn
	}
}

// WithPadding overrides the padding function for the field. The function should transform a string argument
// into a string that is padded to fit within the given display width. The default behavior is to pad fields
// with spaces except for the last field. The padding function will be called after truncation and before coloring.
// Pass nil to disable padding for this value.
func WithPadding(fn func(int, string) string) fieldOption {
	return func(f *tableField) {
		f.paddingFunc = fn
	}
}

// WithColor sets the color function for the field. The function should transform a string value by wrapping
// it in ANSI escape codes. The color function will not be used if the table was initialized in non-terminal mode.
// The color function will be called before truncation and padding.
func WithColor(fn func(string) string) fieldOption {
	return func(f *tableField) {
		f.colorFunc = fn
	}
}

// New initializes a table printer with terminal mode and terminal width. When terminal mode is enabled, the
// output will be human-readable, column-formatted to fit available width, and rendered with color support.
// In non-terminal mode, the output is tab-separated and all truncation of values is disabled.
func New(w io.Writer, isTTY bool, maxWidth int) TablePrinter {
	if isTTY {
		return &ttyTablePrinter{
			out:      w,
			maxWidth: maxWidth,
		}
	}

	return &tsvTablePrinter{
		out: w,
	}
}

type tableField struct {
	text         string
	truncateFunc func(int, string) string
	paddingFunc  func(int, string) string
	colorFunc    func(string) string
}

type ttyTablePrinter struct {
	out        io.Writer
	maxWidth   int
	hasHeaders bool
	rows       [][]tableField
}

func (t *ttyTablePrinter) AddHeader(columns []string, opts ...fieldOption) {
	if t.hasHeaders {
		return
	}

	t.hasHeaders = true
	for _, column := range columns {
		t.AddField(column, opts...)
	}
	t.EndRow()
}

func (t *ttyTablePrinter) AddField(s string, opts ...fieldOption) {
	if t.rows == nil {
		t.rows = make([][]tableField, 1)
	}
	rowI := len(t.rows) - 1
	field := tableField{
		text:         s,
		truncateFunc: text.Truncate,
	}
	for _, opt := range opts {
		opt(&field)
	}
	t.rows[rowI] = append(t.rows[rowI], field)
}

func (t *ttyTablePrinter) EndRow() {
	t.rows = append(t.rows, []tableField{})
}

func (t *ttyTablePrinter) Render() error {
	if len(t.rows) == 0 {
		return nil
	}

	delim := "  "
	numCols := len(t.rows[0])
	colWidths := t.calculateColumnWidths(len(delim))

	for _, row := range t.rows {
		for col, field := range row {
			if col > 0 {
				_, err := fmt.Fprint(t.out, delim)
				if err != nil {
					return err
				}
			}
			truncVal := field.text
			if field.truncateFunc != nil {
				truncVal = field.truncateFunc(colWidths[col], field.text)
			}
			if field.paddingFunc != nil {
				truncVal = field.paddingFunc(colWidths[col], truncVal)
			} else if col < numCols-1 {
				truncVal = text.PadRight(colWidths[col], truncVal)
			}
			if field.colorFunc != nil {
				truncVal = field.colorFunc(truncVal)
			}
			_, err := fmt.Fprint(t.out, truncVal)
			if err != nil {
				return err
			}
		}
		if len(row) > 0 {
			_, err := fmt.Fprint(t.out, "\n")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *ttyTablePrinter) calculateColumnWidths(delimSize int) []int {
	numCols := len(t.rows[0])
	maxColWidths := make([]int, numCols)
	colWidths := make([]int, numCols)

	for _, row := range t.rows {
		for col, field := range row {
			w := text.DisplayWidth(field.text)
			if w > maxColWidths[col] {
				maxColWidths[col] = w
			}
			// if this field has disabled truncating, ensure that the column is wide enough
			if field.truncateFunc == nil && w > colWidths[col] {
				colWidths[col] = w
			}
		}
	}

	availWidth := func() int {
		setWidths := 0
		for col := 0; col < numCols; col++ {
			setWidths += colWidths[col]
		}
		return t.maxWidth - delimSize*(numCols-1) - setWidths
	}
	numFixedCols := func() int {
		fixedCols := 0
		for col := 0; col < numCols; col++ {
			if colWidths[col] > 0 {
				fixedCols++
			}
		}
		return fixedCols
	}

	// set the widths of short columns
	if w := availWidth(); w > 0 {
		if numFlexColumns := numCols - numFixedCols(); numFlexColumns > 0 {
			perColumn := w / numFlexColumns
			for col := 0; col < numCols; col++ {
				if max := maxColWidths[col]; max < perColumn {
					colWidths[col] = max
				}
			}
		}
	}

	// truncate long columns to the remaining available width
	if numFlexColumns := numCols - numFixedCols(); numFlexColumns > 0 {
		perColumn := availWidth() / numFlexColumns
		for col := 0; col < numCols; col++ {
			if colWidths[col] == 0 {
				if max := maxColWidths[col]; max < perColumn {
					colWidths[col] = max
				} else if perColumn > 0 {
					colWidths[col] = perColumn
				}
			}
		}
	}

	// add the remainder to truncated columns
	if w := availWidth(); w > 0 {
		for col := 0; col < numCols; col++ {
			d := maxColWidths[col] - colWidths[col]
			toAdd := w
			if d < toAdd {
				toAdd = d
			}
			colWidths[col] += toAdd
			w -= toAdd
			if w <= 0 {
				break
			}
		}
	}

	return colWidths
}

type tsvTablePrinter struct {
	out        io.Writer
	currentCol int
}

func (t *tsvTablePrinter) AddHeader(_ []string, _ ...fieldOption) {}

func (t *tsvTablePrinter) AddField(text string, _ ...fieldOption) {
	if t.currentCol > 0 {
		fmt.Fprint(t.out, "\t")
	}
	fmt.Fprint(t.out, text)
	t.currentCol++
}

func (t *tsvTablePrinter) EndRow() {
	fmt.Fprint(t.out, "\n")
	t.currentCol = 0
}

func (t *tsvTablePrinter) Render() error {
	return nil
}
