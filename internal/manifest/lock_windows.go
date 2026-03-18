//go:build windows

package manifest

import (
	"fmt"
	"time"
)

// Lock is a no-op on Windows. Advisory locking is not supported.
// Returns an error so callers can fall through to best-effort behavior.
func (l *FileLock) Lock(timeout time.Duration) error {
	return fmt.Errorf("advisory file locking not supported on Windows")
}

// Unlock is a no-op on Windows.
func (l *FileLock) Unlock() error {
	return nil
}
