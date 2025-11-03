package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/jamesainslie/dot/internal/cli/render"
)

// Formatter provides consistent formatting for CLI output.
type Formatter struct {
	colorizer *render.Colorizer
	writer    io.Writer
}

// NewFormatter creates a formatter with the given colorization setting.
func NewFormatter(w io.Writer, colorEnabled bool) *Formatter {
	return &Formatter{
		colorizer: render.NewColorizer(colorEnabled),
		writer:    w,
	}
}

// Success formats and prints a success message with icon.
// Format: "✓ [verb] [count] [item]"
// Example: "✓ Managed 2 packages"
func (f *Formatter) Success(verb string, count int, singular, plural string) {
	verb = strings.Title(verb)
	itemText := pluralize(count, singular, plural)
	fmt.Fprintf(f.writer, "%s %s %d %s\n",
		f.colorizer.Success("✓"),
		verb,
		count,
		itemText,
	)
}

// SuccessSimple prints a simple success message.
// Example: "✓ Upgrade completed"
func (f *Formatter) SuccessSimple(message string) {
	fmt.Fprintf(f.writer, "%s %s\n",
		f.colorizer.Success("✓"),
		message,
	)
}

// Error formats and prints an error message with icon.
func (f *Formatter) Error(message string) {
	fmt.Fprintf(f.writer, "%s %s\n",
		f.colorizer.Error("✗"),
		message,
	)
}

// Warning formats and prints a warning message with icon.
func (f *Formatter) Warning(message string) {
	fmt.Fprintf(f.writer, "%s %s\n",
		f.colorizer.Warning("⚠"),
		message,
	)
}

// Info formats and prints an info message with icon.
func (f *Formatter) Info(message string) {
	fmt.Fprintf(f.writer, "%s %s\n",
		f.colorizer.Info("ℹ"),
		message,
	)
}

// Bullet prints a bullet point item.
func (f *Formatter) Bullet(text string) {
	fmt.Fprintf(f.writer, "  %s %s\n",
		f.colorizer.Dim("•"),
		text,
	)
}

// BulletWithDetail prints a bullet point with main text and detail.
// Example: "• main — detail"
func (f *Formatter) BulletWithDetail(main, detail string) {
	fmt.Fprintf(f.writer, "  %s %s %s %s\n",
		f.colorizer.Dim("•"),
		f.colorizer.Bold(main),
		f.colorizer.Dim("—"),
		f.colorizer.Dim(detail),
	)
}

// Header prints a section header.
func (f *Formatter) Header(text string) {
	fmt.Fprintf(f.writer, "%s\n", f.colorizer.Bold(text))
}

// Divider prints a visual divider line.
func (f *Formatter) Divider() {
	fmt.Fprintln(f.writer)
}

// BlankLine prints a single blank line for consistent spacing.
func (f *Formatter) BlankLine() {
	fmt.Fprintln(f.writer)
}

// Indent prints text with 2-space indentation.
func (f *Formatter) Indent(level int, text string) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(f.writer, "%s%s\n", indent, text)
}

// pluralize returns the singular or plural form based on count.
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
