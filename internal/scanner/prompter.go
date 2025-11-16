package scanner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// LargeFilePrompter determines whether large files should be included during scanning.
type LargeFilePrompter interface {
	// ShouldInclude asks whether a large file should be included.
	// Returns true to include the file, false to skip it.
	ShouldInclude(path string, size int64, limit int64) bool
}

// InteractivePrompter prompts the user for each large file in TTY mode.
type InteractivePrompter struct {
	input   io.Reader
	output  io.Writer
	skipAll bool
}

// NewInteractivePrompter creates a new interactive prompter using stdin/stderr.
func NewInteractivePrompter() *InteractivePrompter {
	return &InteractivePrompter{
		input:  os.Stdin,
		output: os.Stderr,
	}
}

// NewInteractivePrompterWithIO creates a new interactive prompter with custom I/O.
func NewInteractivePrompterWithIO(input io.Reader, output io.Writer) *InteractivePrompter {
	return &InteractivePrompter{
		input:  input,
		output: output,
	}
}

// ShouldInclude prompts the user whether to include a large file.
func (p *InteractivePrompter) ShouldInclude(path string, size int64, limit int64) bool {
	if p.skipAll {
		return false
	}

	fmt.Fprintf(p.output, "\nLarge file detected:\n")
	fmt.Fprintf(p.output, "  Path: %s\n", path)
	fmt.Fprintf(p.output, "  Size: %s (limit: %s)\n", formatSize(size), formatSize(limit))
	fmt.Fprintf(p.output, "\nOptions:\n")
	fmt.Fprintf(p.output, "  i) Include this file\n")
	fmt.Fprintf(p.output, "  s) Skip this file\n")
	fmt.Fprintf(p.output, "  a) Skip all large files\n")
	fmt.Fprintf(p.output, "Choice [s]: ")

	reader := bufio.NewReader(p.input)
	choice, err := reader.ReadString('\n')
	if err != nil {
		// On error, default to skip
		return false
	}

	choice = strings.ToLower(strings.TrimSpace(choice))

	switch choice {
	case "i":
		return true
	case "a":
		p.skipAll = true
		return false
	default: // "s" or empty
		return false
	}
}

// BatchPrompter automatically skips large files in non-interactive mode.
type BatchPrompter struct{}

// NewBatchPrompter creates a new batch prompter.
func NewBatchPrompter() *BatchPrompter {
	return &BatchPrompter{}
}

// ShouldInclude always returns false in batch mode.
func (p *BatchPrompter) ShouldInclude(path string, size int64, limit int64) bool {
	return false
}

// formatSize formats a byte count as a human-readable size string.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsInteractive checks if the program is running in an interactive terminal.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
