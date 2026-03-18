package manifest

import (
	"os"
	"path/filepath"
)

const lockFileName = ".dot-manifest.lock"

// FileLock provides advisory file locking for manifest operations.
// On Unix, it uses flock(2). On Windows, it returns an error (best-effort).
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

