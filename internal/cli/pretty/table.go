// Package pretty provides consistent, professional CLI output formatting using lipgloss.
package pretty

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"

	"github.com/yaklabco/dot/internal/cli/terminal"
)

// TableStyle defines the visual style for tables.
type TableStyle int

const (
	// StyleBordered uses rounded borders with full table structure.
	StyleBordered TableStyle = iota
	// StyleLight uses light borders for a clean look.
	StyleLight
	// StyleMinimal uses no borders, just spacing (default).
	StyleMinimal
	// StyleCompact uses dense formatting for large datasets.
	StyleCompact
)

// TableConfig holds configuration for table rendering.
type TableConfig struct {
	// MaxWidth is the maximum table width (0 = auto-detect from terminal).
	MaxWidth int
	// ColorEnabled controls whether to use colors in output.
	ColorEnabled bool
	// AutoWrap enables automatic text wrapping in columns.
	AutoWrap bool
	// SortColumn is the column index to sort by (-1 = no sorting).
	SortColumn int
	// SortAsc controls sort direction (true = ascending).
	SortAsc bool
}

// DefaultTableConfig returns sensible defaults for table rendering.
func DefaultTableConfig() TableConfig {
	return TableConfig{
		MaxWidth:     GetTerminalWidth(),
		ColorEnabled: ShouldUseColor(),
		AutoWrap:     true,
		SortColumn:   -1,
		SortAsc:      true,
	}
}

// TableWriter provides table rendering with lipgloss styling.
type TableWriter struct {
	headers []string
	rows    [][]string
	config  TableConfig
	style   TableStyle
}

// NewTableWriter creates a new table writer with the given style and config.
func NewTableWriter(style TableStyle, config TableConfig) *TableWriter {
	return &TableWriter{
		headers: []string{},
		rows:    [][]string{},
		config:  config,
		style:   style,
	}
}

// SetHeader sets the table header row.
func (w *TableWriter) SetHeader(headers ...interface{}) {
	w.headers = make([]string, len(headers))
	for i, h := range headers {
		// Uppercase headers to match go-pretty behavior
		w.headers[i] = strings.ToUpper(fmt.Sprintf("%v", h))
	}
}

// AppendRow adds a data row to the table.
func (w *TableWriter) AppendRow(row ...interface{}) {
	strRow := make([]string, len(row))
	for i, cell := range row {
		strRow[i] = fmt.Sprintf("%v", cell)
	}
	w.rows = append(w.rows, strRow)
}

// AppendRows adds multiple data rows.
func (w *TableWriter) AppendRows(rows [][]interface{}) {
	for _, row := range rows {
		w.AppendRow(row...)
	}
}

// AppendSeparator adds a visual separator line (no-op for lipgloss implementation).
func (w *TableWriter) AppendSeparator() {
	// Separator not supported in this implementation
	_ = w
}

// SetAutoIndex enables row numbering (no-op for lipgloss implementation).
func (w *TableWriter) SetAutoIndex(enabled bool) {
	// Auto-index not supported in this implementation
	_ = enabled
}

// SetColumnConfig sets configuration for a specific column (no-op for lipgloss implementation).
func (w *TableWriter) SetColumnConfig(columnNumber int, config interface{}) {
	// Column config not supported in this implementation
	_, _ = columnNumber, config
}

// SortBy sorts the table by the configured column (no-op for lipgloss implementation).
func (w *TableWriter) SortBy(columnNumber int, ascending bool) {
	// Sorting not supported in this implementation
	_, _ = columnNumber, ascending
}

// Render outputs the table to the given writer.
func (w *TableWriter) Render(out io.Writer) {
	fmt.Fprint(out, w.RenderString())
}

// RenderString returns the rendered table as a string.
func (w *TableWriter) RenderString() string {
	// Create lipgloss table
	tbl := table.New()

	// Apply border style
	switch w.style {
	case StyleMinimal:
		tbl.Border(lipgloss.HiddenBorder()).BorderTop(false).BorderBottom(false)
	case StyleLight:
		tbl.Border(lipgloss.RoundedBorder())
	case StyleBordered:
		tbl.Border(lipgloss.RoundedBorder())
	case StyleCompact:
		tbl.Border(lipgloss.HiddenBorder()).BorderTop(false).BorderBottom(false)
	}

	// Apply styling function
	tbl.StyleFunc(func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)

		if w.config.ColorEnabled {
			switch row {
			case table.HeaderRow:
				// Header styling: bold and muted gray
				style = style.Bold(true).Foreground(lipgloss.Color("245")).Align(lipgloss.Center)
			default:
				// Data row styling: left-aligned
				style = style.Align(lipgloss.Left)
			}
		} else {
			// Without colors, center headers and left-align data
			if row == table.HeaderRow {
				style = style.Align(lipgloss.Center)
			} else {
				style = style.Align(lipgloss.Left)
			}
		}

		return style
	})

	// Set headers
	if len(w.headers) > 0 {
		tbl.Headers(w.headers...)
	}

	// Add rows
	data := table.NewStringData()
	for _, row := range w.rows {
		data.Append(row)
	}
	tbl.Data(data)

	// Render and return
	return tbl.Render()
}

// ShouldUseColor determines if color output should be enabled.
func ShouldUseColor() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	fd := terminal.FdInt(os.Stdout.Fd())
	return term.IsTerminal(fd)
}

// GetTerminalWidth returns the width of the terminal.
func GetTerminalWidth() int {
	fd := terminal.FdInt(os.Stdout.Fd())
	width, _, err := term.GetSize(fd)
	if err != nil || width == 0 {
		return 80 // Default fallback
	}
	return width
}

// GetTerminalHeight returns the height of the terminal.
func GetTerminalHeight() int {
	fd := terminal.FdInt(os.Stdout.Fd())
	_, height, err := term.GetSize(fd)
	if err != nil || height == 0 {
		return 24 // Default fallback
	}
	return height
}

// IsInteractive returns true if the output is an interactive terminal.
func IsInteractive() bool {
	return term.IsTerminal(terminal.FdInt(os.Stdout.Fd()))
}
