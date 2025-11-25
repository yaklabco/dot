package install

import (
	"io/fs"
	"os"
)

// FileSystem abstracts filesystem operations for testing.
type FileSystem interface {
	// ReadFile reads the named file and returns its contents.
	ReadFile(path string) ([]byte, error)

	// ReadDir reads the named directory and returns a list of directory entries.
	ReadDir(path string) ([]fs.DirEntry, error)

	// Stat returns file info for the named file.
	Stat(path string) (fs.FileInfo, error)
}

// OSFileSystem is the real filesystem implementation using os package.
type OSFileSystem struct{}

// ReadFile reads the named file and returns its contents.
func (OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadDir reads the named directory and returns a list of directory entries.
func (OSFileSystem) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

// Stat returns file info for the named file.
func (OSFileSystem) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

// Ensure OSFileSystem implements FileSystem.
var _ FileSystem = OSFileSystem{}
