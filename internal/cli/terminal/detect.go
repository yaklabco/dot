// Package terminal provides terminal detection utilities.
package terminal

import (
	"math"
	"os"

	"golang.org/x/term"
)

// FdInt safely converts a file descriptor from uintptr to int.
// File descriptors are small non-negative integers on all platforms,
// so overflow is not possible in practice, but this satisfies gosec G115.
func FdInt(fd uintptr) int {
	if fd > math.MaxInt {
		return -1
	}
	return int(fd)
}

// IsInteractive determines if the current process is running in an interactive terminal.
//
// Returns true if both stdin and stdout are connected to a terminal (TTY).
// Returns false if either is redirected to a file or pipe.
//
// This is useful for deciding whether to prompt the user for input or
// fall back to non-interactive behavior.
func IsInteractive() bool {
	// Check if stdin is a terminal
	if !term.IsTerminal(FdInt(os.Stdin.Fd())) {
		return false
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(FdInt(os.Stdout.Fd())) {
		return false
	}

	return true
}
