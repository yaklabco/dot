package progress

import (
	"os"

	"golang.org/x/term"

	"github.com/yaklabco/dot/internal/cli/terminal"
)

// Indicator provides progress feedback.
type Indicator interface {
	Start(message string)
	Update(current, total int, message string)
	Stop(message string)
	Fail(message string)
}

// Config for progress indicators.
type Config struct {
	Enabled     bool
	Interactive bool // Terminal supports cursor control
	Width       int
}

// New creates appropriate indicator for terminal.
func New(cfg Config) Indicator {
	if !cfg.Enabled {
		return &NoOpIndicator{}
	}

	if !cfg.Interactive {
		return &NoOpIndicator{}
	}

	// Default to spinner for indeterminate progress
	return NewSpinner(cfg)
}

// NewBar creates a progress bar indicator.
func NewBar(cfg Config) Indicator {
	if !cfg.Enabled || !cfg.Interactive {
		return &NoOpIndicator{}
	}
	return &Bar{
		width: cfg.Width,
	}
}

// NewSpinner creates a spinner indicator.
func NewSpinner(cfg Config) Indicator {
	if !cfg.Enabled || !cfg.Interactive {
		return &NoOpIndicator{}
	}
	return NewSpinnerWithStyle(cfg, SpinnerDots)
}

// IsInteractive checks if the terminal is interactive.
// Checks if stdout or stderr is a terminal (not stdin, which may be piped).
func IsInteractive() bool {
	// Check stdout first
	if term.IsTerminal(terminal.FdInt(os.Stdout.Fd())) {
		return true
	}
	// Fallback to stderr
	return term.IsTerminal(terminal.FdInt(os.Stderr.Fd()))
}

// NoOpIndicator does nothing (for non-interactive terminals).
type NoOpIndicator struct{}

// Start does nothing.
func (n *NoOpIndicator) Start(message string) {}

// Update does nothing.
func (n *NoOpIndicator) Update(current, total int, message string) {}

// Stop does nothing.
func (n *NoOpIndicator) Stop(message string) {}

// Fail does nothing.
func (n *NoOpIndicator) Fail(message string) {}
