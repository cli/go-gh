package tableprinter

import (
	"io"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/evertras/bubble-table/table"
)

type tableOptions struct {
	out         io.Writer
	isTTY       *bool
	width       *int
	interactive bool
}

type tableOption func(*tableOptions)

func WithTableWriter(w io.Writer) tableOption {
	return func(opt *tableOptions) {
		opt.out = w
	}
}

func WithTableTTY(isTTY bool) tableOption {
	return func(opt *tableOptions) {
		opt.isTTY = new(bool)
		*opt.isTTY = isTTY
	}
}

func WithTableWidth(viewportWidth int) tableOption {
	return func(opt *tableOptions) {
		opt.width = new(int)
		*opt.width = viewportWidth
	}
}

func WithTableInteractive(interactive bool) tableOption {
	return func(opt *tableOptions) {
		opt.interactive = interactive
	}
}

func FromTerm(term term.Term, opts ...tableOption) TablePrinter {
	var options tableOptions
	for _, opt := range opts {
		opt(&options)
	}

	out := options.out
	if out == nil {
		out = term.Out()
	}

	isTTY := options.isTTY
	if isTTY == nil {
		isTTY = new(bool)
		*isTTY = term.IsTerminalOutput()
	}

	width := options.width
	if width == nil {
		width = new(int)
		if _width, _, err := term.Size(); err != nil {
			*width = _width
		} else {
			*width = 80
		}
	}

	if *isTTY && options.interactive {
		return &interactiveTablePrinter{
			out:           out,
			viewportWidth: *width,
		}
	}

	return New(out, *isTTY, *width)
}

type interactiveTablePrinter struct {
	out           io.Writer
	viewportWidth int
	headers       []interactiveHeaderField
	rows          []table.Row
	column        int
}

func (t *interactiveTablePrinter) AddHeader(columns []string, opts ...fieldOption) {
	if len(t.headers) > 0 {
		return
	}

	t.headers = make([]interactiveHeaderField, len(columns))
	for i, column := range columns {
		header := interactiveHeaderField{
			tableField: tableField{
				text: column,
			},
			width: len(column),
		}
		for _, opt := range opts {
			opt(&header.tableField)
		}
		// TODO: Adapt styles, truncation.
		t.headers[i] = header
	}
}

func (t *interactiveTablePrinter) AddField(s string, opts ...fieldOption) {
	if len(t.rows) == 0 {
		t.rows = make([]table.Row, 1)
		t.rows[0] = table.NewRow(make(table.RowData))
	}

	// TODO: Figure out a better way to ignore pre-computed styles because this is way too slow.
	length, _ := lipgloss.Size(s)

	// TODO: Figure out max column width or at least define const.
	t.headers[t.column].width = min(t.viewportWidth/2, max(t.headers[t.column].width, length))

	row := len(t.rows) - 1
	t.rows[row].Data[strconv.Itoa(t.column)] = s
	t.column++

}

func (t *interactiveTablePrinter) EndRow() {
	t.rows = append(t.rows, table.NewRow(make(table.RowData)))
	t.column = 0
}

func (t *interactiveTablePrinter) Render() error {
	columnStyle := lipgloss.NewStyle().AlignHorizontal(lipgloss.Left)

	headers := make([]table.Column, len(t.headers))
	for i := range t.headers {
		headers[i] = table.NewColumn(strconv.Itoa(i), t.headers[i].text, t.headers[i].width).WithStyle(columnStyle)
	}

	last := len(t.rows) - 1
	rows := t.rows
	if len(t.rows[last].Data) == 0 {
		rows = t.rows[:last]
	}
	m := interactiveTableModel{
		model: table.New(headers).
			WithMaxTotalWidth(t.viewportWidth).
			// WithTargetWidth(t.viewportWidth).
			WithRows(rows).
			WithStaticFooter("Shit+Left, Shift+Right to scroll horizontally; Ctrl+C, ESC to quit").
			Focused(true),
	}
	p := tea.NewProgram(m, tea.WithOutput(t.out), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

type interactiveHeaderField struct {
	tableField
	width int
}

type interactiveTableModel struct {
	model table.Model
}

func (m interactiveTableModel) Init() tea.Cmd {
	return nil
}

func (m interactiveTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.model, cmd = m.model.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			cmds = append(cmds, tea.Quit)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m interactiveTableModel) View() string {
	return m.model.View()
}
