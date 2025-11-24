// Package adapters provides concrete implementations of infrastructure ports.
package adapters

import (
	"context"
	"io/fs"
	"os"

	"github.com/yaklabco/dot/internal/domain"
)

// OSFilesystem implements the FS interface using the os package.
type OSFilesystem struct{}

// NewOSFilesystem creates a new OS filesystem adapter.
func NewOSFilesystem() *OSFilesystem {
	return &OSFilesystem{}
}

// Stat returns file information.
func (f *OSFilesystem) Stat(ctx context.Context, name string) (domain.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	info, err := os.Stat(name)
	if err != nil {
		return nil, err
	}

	return WrapFileInfo(info), nil
}

// Lstat returns file information without following symlinks.
func (f *OSFilesystem) Lstat(ctx context.Context, name string) (domain.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}

	return WrapFileInfo(info), nil
}

// ReadDir lists directory contents.
func (f *OSFilesystem) ReadDir(ctx context.Context, name string) ([]domain.DirEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}

	result := make([]domain.DirEntry, len(entries))
	for i, entry := range entries {
		result[i] = WrapDirEntry(entry)
	}

	return result, nil
}

// ReadLink reads the target of a symbolic link.
func (f *OSFilesystem) ReadLink(ctx context.Context, name string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	return os.Readlink(name)
}

// ReadFile reads the entire file.
func (f *OSFilesystem) ReadFile(ctx context.Context, name string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return os.ReadFile(name)
}

// WriteFile writes data to a file.
func (f *OSFilesystem) WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return os.WriteFile(name, data, perm)
}

// Mkdir creates a directory.
func (f *OSFilesystem) Mkdir(ctx context.Context, name string, perm fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return os.Mkdir(name, perm)
}

// MkdirAll creates a directory tree.
func (f *OSFilesystem) MkdirAll(ctx context.Context, name string, perm fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return os.MkdirAll(name, perm)
}

// Remove removes a file or empty directory.
func (f *OSFilesystem) Remove(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return os.Remove(name)
}

// RemoveAll removes a directory tree.
func (f *OSFilesystem) RemoveAll(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return os.RemoveAll(name)
}

// Symlink creates a symbolic link.
func (f *OSFilesystem) Symlink(ctx context.Context, oldname, newname string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return os.Symlink(oldname, newname)
}

// Rename moves or renames a file.
func (f *OSFilesystem) Rename(ctx context.Context, oldname, newname string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return os.Rename(oldname, newname)
}

// Exists checks if a path exists.
func (f *OSFilesystem) Exists(ctx context.Context, name string) bool {
	if err := ctx.Err(); err != nil {
		return false
	}

	_, err := os.Stat(name)
	return err == nil
}

// IsDir checks if a path is a directory.
func (f *OSFilesystem) IsDir(ctx context.Context, name string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	info, err := os.Stat(name)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

// IsSymlink checks if a path is a symbolic link.
func (f *OSFilesystem) IsSymlink(ctx context.Context, name string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	info, err := os.Lstat(name)
	if err != nil {
		return false, err
	}

	return info.Mode()&fs.ModeSymlink != 0, nil
}

// FileInfo adapters

// osFileInfo wraps fs.FileInfo to implement domain.FileInfo.
type osFileInfo struct {
	info fs.FileInfo
}

// WrapFileInfo wraps a standard fs.FileInfo.
func WrapFileInfo(info fs.FileInfo) domain.FileInfo {
	return osFileInfo{info: info}
}

func (i osFileInfo) Name() string      { return i.info.Name() }
func (i osFileInfo) Size() int64       { return i.info.Size() }
func (i osFileInfo) Mode() fs.FileMode { return i.info.Mode() }
func (i osFileInfo) ModTime() any      { return i.info.ModTime() }
func (i osFileInfo) IsDir() bool       { return i.info.IsDir() }
func (i osFileInfo) Sys() any          { return i.info.Sys() }

// DirEntry adapters

// osDirEntry wraps fs.DirEntry to implement domain.DirEntry.
type osDirEntry struct {
	entry fs.DirEntry
}

// WrapDirEntry wraps a standard fs.DirEntry.
func WrapDirEntry(entry fs.DirEntry) domain.DirEntry {
	return osDirEntry{entry: entry}
}

func (e osDirEntry) Name() string { return e.entry.Name() }
func (e osDirEntry) IsDir() bool  { return e.entry.IsDir() }
func (e osDirEntry) Type() fs.FileMode {
	return e.entry.Type()
}

func (e osDirEntry) Info() (domain.FileInfo, error) {
	info, err := e.entry.Info()
	if err != nil {
		return nil, err
	}
	return WrapFileInfo(info), nil
}
