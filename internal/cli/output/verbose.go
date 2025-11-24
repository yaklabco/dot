package output

import (
	"fmt"

	"github.com/yaklabco/dot/internal/cli/render"
)

// VerboseLogger provides verbosity-aware logging.
type VerboseLogger struct {
	level        int
	colorEnabled bool
	quiet        bool
}

// NewVerboseLogger creates a new verbose logger.
func NewVerboseLogger(level int, colorEnabled bool, quiet bool) *VerboseLogger {
	return &VerboseLogger{
		level:        level,
		colorEnabled: colorEnabled,
		quiet:        quiet,
	}
}

// Debug logs debug information (level 3+).
func (l *VerboseLogger) Debug(format string, args ...interface{}) {
	if l.quiet || l.level < 3 {
		return
	}
	message := fmt.Sprintf(format, args...)
	if l.colorEnabled {
		fmt.Println(render.DimStyle("[DEBUG] " + message))
	} else {
		fmt.Println("[DEBUG] " + message)
	}
}

// Info logs informational messages (level 2+).
func (l *VerboseLogger) Info(format string, args ...interface{}) {
	if l.quiet || l.level < 2 {
		return
	}
	message := fmt.Sprintf(format, args...)
	if l.colorEnabled {
		fmt.Println(render.InfoStyle(message))
	} else {
		fmt.Println(message)
	}
}

// Summary logs summary information (level 1+).
func (l *VerboseLogger) Summary(format string, args ...interface{}) {
	if l.quiet || l.level < 1 {
		return
	}
	message := fmt.Sprintf(format, args...)
	fmt.Println(message)
}

// Always logs messages regardless of verbosity (except quiet mode).
func (l *VerboseLogger) Always(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	message := fmt.Sprintf(format, args...)
	fmt.Println(message)
}

// IsQuiet returns whether quiet mode is enabled.
func (l *VerboseLogger) IsQuiet() bool {
	return l.quiet
}

// Level returns the current verbosity level.
func (l *VerboseLogger) Level() int {
	return l.level
}
