package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const lockFileName = ".dot-manifest.lock"

// FileLock provides advisory file locking for manifest operations.
// It uses flock(2) on Unix systems to prevent concurrent manifest writes.
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock creates a new file lock for the given manifest directory.
func NewFileLock(manifestDir string) *FileLock {
	return &FileLock{
		path: filepath.Join(manifestDir, lockFileName),
	}
}

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
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			l.file = nil
			return fmt.Errorf("lock timeout after %v: another dot process may be running", timeout)
		}
	}

	return fmt.Errorf("lock acquisition failed")
}

// Unlock releases the advisory lock and removes the lock file.
func (l *FileLock) Unlock() error {
	if l.file == nil {
		return nil
	}

	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		_ = l.file.Close()
		l.file = nil
		return fmt.Errorf("unlock: %w", err)
	}

	name := l.file.Name()
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("close lock file: %w", err)
	}
	l.file = nil

	// Best-effort removal of lock file
	_ = os.Remove(name)
	return nil
}
