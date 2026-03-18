//go:build unix

package manifest

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Lock acquires an exclusive advisory lock with a timeout.
// Returns an error if the lock cannot be acquired within the timeout.
func (l *FileLock) Lock(timeout time.Duration) error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	l.file = f

	// Try non-blocking lock first
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == nil {
		return nil
	}

	// Fall back to polling with timeout
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if time.Now().After(deadline) {
			_ = f.Close()
			l.file = nil
			return fmt.Errorf("lock timeout after %v: another dot process may be running", timeout)
		}
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("lock acquisition failed")
}

// Unlock releases the advisory lock.
func (l *FileLock) Unlock() error {
	if l.file == nil {
		return nil
	}

	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		_ = l.file.Close()
		l.file = nil
		return fmt.Errorf("unlock: %w", err)
	}

	if err := l.file.Close(); err != nil {
		l.file = nil
		return fmt.Errorf("close lock file: %w", err)
	}
	l.file = nil
	return nil
}
