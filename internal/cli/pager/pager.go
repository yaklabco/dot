package pager

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"
)

// debugLogger is an optional logger for debugging pager internals.
// Set via WithDebugLogger option.
var debugLogger *slog.Logger

// Pager displays content with pagination support.
type Pager struct {
	output           io.Writer
	input            io.Reader
	height           int
	width            int
	interactive      bool
	forceInteractive bool // Skip terminal detection
}

// Option configures a Pager.
type Option func(*Pager)

// WithOutput sets the output writer.
func WithOutput(w io.Writer) Option {
	return func(p *Pager) {
		p.output = w
	}
}

// WithInput sets the input reader for pagination controls.
func WithInput(r io.Reader) Option {
	return func(p *Pager) {
		p.input = r
	}
}

// WithSize sets explicit terminal dimensions.
func WithSize(width, height int) Option {
	return func(p *Pager) {
		p.width = width
		p.height = height
	}
}

// WithInteractive sets whether the pager should be interactive.
// When set to true, this forces interactive mode regardless of terminal detection.
func WithInteractive(interactive bool) Option {
	return func(p *Pager) {
		p.interactive = interactive
		p.forceInteractive = interactive // Force skips terminal detection
	}
}

// New creates a new Pager with automatic terminal detection.
func New(opts ...Option) *Pager {
	if debugLogger != nil {
		debugLogger.Debug("pager.New: starting")
	}

	p := &Pager{
		output:      os.Stdout,
		input:       os.Stdin,
		interactive: true,
	}

	for _, opt := range opts {
		opt(p)
	}

	if debugLogger != nil {
		debugLogger.Debug("pager.New: options applied")
	}

	// Detect terminal size if not set
	if p.width == 0 || p.height == 0 {
		if debugLogger != nil {
			debugLogger.Debug("pager.New: detecting terminal size")
		}
		p.detectTerminalSize()
		if debugLogger != nil {
			debugLogger.Debug("pager.New: terminal size detected", "width", p.width, "height", p.height)
		}
	}

	// Check if both stdin AND stdout are terminals for interactivity (unless forced)
	// Both must be TTYs for interactive pagination to work properly
	if p.interactive && !p.forceInteractive {
		if debugLogger != nil {
			debugLogger.Debug("pager.New: checking terminal status")
		}
		inputIsTTY := p.isInputTerminal()
		if debugLogger != nil {
			debugLogger.Debug("pager.New: input terminal check complete", "is_tty", inputIsTTY)
		}
		outputIsTTY := p.isOutputTerminal()
		if debugLogger != nil {
			debugLogger.Debug("pager.New: output terminal check complete", "is_tty", outputIsTTY)
		}
		p.interactive = inputIsTTY && outputIsTTY
	}

	if debugLogger != nil {
		debugLogger.Debug("pager.New: complete", "interactive", p.interactive)
	}

	return p
}

// SetDebugLogger sets the debug logger for the pager package.
// Pass nil to disable debug logging.
func SetDebugLogger(logger *slog.Logger) {
	debugLogger = logger
}

// isInputTerminal checks if the input is a terminal.
func (p *Pager) isInputTerminal() bool {
	if f, ok := p.input.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// isOutputTerminal checks if the output is a terminal.
func (p *Pager) isOutputTerminal() bool {
	if f, ok := p.output.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// IsInteractive returns whether the pager is in interactive mode.
func (p *Pager) IsInteractive() bool {
	return p.interactive
}

func (p *Pager) detectTerminalSize() {
	if f, ok := p.output.(*os.File); ok {
		width, height, err := term.GetSize(int(f.Fd()))
		if err == nil {
			p.width = width
			p.height = height
			return
		}
	}
	// Default fallback
	p.width = 80
	p.height = 24
}

// Height returns the terminal height.
func (p *Pager) Height() int {
	return p.height
}

// Width returns the terminal width.
func (p *Pager) Width() int {
	return p.width
}

// NeedsPaging returns true if the content would exceed one page.
func (p *Pager) NeedsPaging(lines int) bool {
	// Reserve 2 lines for prompt
	return lines > p.height-2
}

// Display shows content with pagination if needed.
func (p *Pager) Display(content string) error {
	lines := strings.Split(content, "\n")
	return p.DisplayLines(lines)
}

// DisplayLines shows lines with pagination if needed.
func (p *Pager) DisplayLines(lines []string) error {
	if !p.interactive || !p.NeedsPaging(len(lines)) {
		// No pagination needed - just print
		for _, line := range lines {
			fmt.Fprintln(p.output, line)
		}
		return nil
	}

	return p.paginateLines(lines)
}

func (p *Pager) paginateLines(lines []string) error {
	pageSize := p.height - 2 // Reserve space for prompt
	totalLines := len(lines)
	currentLine := 0

	reader := bufio.NewReader(p.input)

	for currentLine < totalLines {
		// Display one page
		endLine := currentLine + pageSize
		if endLine > totalLines {
			endLine = totalLines
		}

		for i := currentLine; i < endLine; i++ {
			fmt.Fprintln(p.output, lines[i])
		}

		currentLine = endLine

		// Check if there's more content
		if currentLine >= totalLines {
			break
		}

		// Show prompt
		remaining := totalLines - currentLine
		fmt.Fprintf(p.output, "\n-- %d more line(s) -- Press Enter to continue, q to quit: ", remaining)

		// Wait for input
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read input: %w", err)
		}

		// Check for quit
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "q" || input == "quit" {
			fmt.Fprintln(p.output)
			return nil
		}

		// Clear the prompt line (move up and clear)
		fmt.Fprint(p.output, "\033[1A\033[2K")
	}

	return nil
}

// PagedWriter wraps a writer and provides pagination.
type PagedWriter struct {
	pager *Pager
	lines []string
}

// NewPagedWriter creates a writer that collects output for pagination.
func NewPagedWriter(pager *Pager) *PagedWriter {
	return &PagedWriter{
		pager: pager,
		lines: make([]string, 0),
	}
}

// Write implements io.Writer.
func (w *PagedWriter) Write(p []byte) (n int, err error) {
	content := string(p)
	newLines := strings.Split(content, "\n")

	// Handle partial line from previous write
	if len(w.lines) > 0 && !strings.HasSuffix(content, "\n") {
		lastIdx := len(w.lines) - 1
		w.lines[lastIdx] += newLines[0]
		newLines = newLines[1:]
	}

	w.lines = append(w.lines, newLines...)
	return len(p), nil
}

// Flush displays all collected content with pagination.
func (w *PagedWriter) Flush() error {
	return w.pager.DisplayLines(w.lines)
}

// LineCount returns the number of lines collected.
func (w *PagedWriter) LineCount() int {
	return len(w.lines)
}
