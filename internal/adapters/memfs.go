package adapters

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/yaklabco/dot/internal/domain"
)

// MemFS implements an in-memory filesystem for testing.
// It is not thread-safe and should only be used in tests.
type MemFS struct {
	files map[string]*memFile
	mu    sync.RWMutex
}

type memFile struct {
	data    []byte
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	symlink string // If not empty, this is a symlink
}

// NewMemFS creates a new in-memory filesystem.
func NewMemFS() *MemFS {
	return &MemFS{
		files: make(map[string]*memFile),
	}
}

func (f *MemFS) Stat(ctx context.Context, name string) (domain.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, exists := f.files[name]
	if !exists {
		return nil, fs.ErrNotExist
	}

	return &memFileInfo{
		name:    filepath.Base(name),
		size:    int64(len(file.data)),
		mode:    file.mode,
		modTime: file.modTime,
		isDir:   file.isDir,
	}, nil
}

func (f *MemFS) Lstat(ctx context.Context, name string) (domain.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, exists := f.files[name]
	if !exists {
		return nil, fs.ErrNotExist
	}

	// For Lstat, if it's a symlink, report it as such
	mode := file.mode
	if file.symlink != "" {
		mode |= fs.ModeSymlink
	}

	return &memFileInfo{
		name:    filepath.Base(name),
		size:    int64(len(file.data)),
		mode:    mode,
		modTime: file.modTime,
		isDir:   file.isDir,
	}, nil
}

func (f *MemFS) ReadDir(ctx context.Context, name string) ([]domain.DirEntry, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Check if directory exists
	dir, exists := f.files[name]
	if !exists {
		return nil, fs.ErrNotExist
	}
	if !dir.isDir {
		return nil, errors.New("not a directory")
	}

	// Find all direct children
	var entries []domain.DirEntry
	prefix := filepath.Clean(name) + string(filepath.Separator)

	for path, file := range f.files {
		if path == name {
			continue
		}
		// Check if this is a direct child
		if filepath.Dir(path) == name {
			entries = append(entries, &memDirEntry{
				name:  filepath.Base(path),
				isDir: file.isDir,
				mode:  file.mode,
			})
		} else if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			// This is a child in a subdirectory, create directory entry if needed
			rel := path[len(prefix):]
			if idx := filepath.Dir(rel); idx != "." {
				// Skip nested entries
				continue
			}
		}
	}

	return entries, nil
}

func (f *MemFS) ReadLink(ctx context.Context, name string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, exists := f.files[name]
	if !exists {
		return "", fs.ErrNotExist
	}
	if file.symlink == "" {
		return "", errors.New("not a symlink")
	}

	return file.symlink, nil
}

func (f *MemFS) ReadFile(ctx context.Context, name string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, exists := f.files[name]
	if !exists {
		return nil, fs.ErrNotExist
	}
	if file.isDir {
		return nil, errors.New("is a directory")
	}

	return file.data, nil
}

func (f *MemFS) WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Ensure parent directory exists
	parent := filepath.Dir(name)
	if parent != "." && parent != "/" {
		if _, exists := f.files[parent]; !exists {
			return fs.ErrNotExist
		}
	}

	f.files[name] = &memFile{
		data:    data,
		mode:    perm,
		modTime: time.Now(),
		isDir:   false,
	}

	return nil
}

func (f *MemFS) Mkdir(ctx context.Context, name string, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.files[name]; exists {
		return fs.ErrExist
	}

	f.files[name] = &memFile{
		mode:    perm | fs.ModeDir,
		modTime: time.Now(),
		isDir:   true,
	}

	return nil
}

func (f *MemFS) MkdirAll(ctx context.Context, name string, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clean the path
	cleanPath := filepath.Clean(name)

	// Ensure root exists
	if _, exists := f.files["/"]; !exists {
		f.files["/"] = &memFile{
			mode:    0755 | fs.ModeDir,
			modTime: time.Now(),
			isDir:   true,
		}
	}

	// If requesting root, we're done
	if cleanPath == "/" || cleanPath == "." {
		return nil
	}

	// Create all ancestor directories by walking up the path
	ancestors := []string{}
	current := cleanPath
	for current != "/" && current != "." {
		ancestors = append(ancestors, current)
		current = filepath.Dir(current)
	}

	// Create directories from root to leaf
	for i := len(ancestors) - 1; i >= 0; i-- {
		path := ancestors[i]
		if _, exists := f.files[path]; !exists {
			f.files[path] = &memFile{
				mode:    perm | fs.ModeDir,
				modTime: time.Now(),
				isDir:   true,
			}
		}
	}

	return nil
}

func (f *MemFS) Remove(ctx context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.files[name]; !exists {
		return fs.ErrNotExist
	}

	delete(f.files, name)
	return nil
}

func (f *MemFS) RemoveAll(ctx context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Remove the path and all children
	prefix := filepath.Clean(name)
	toRemove := []string{}

	for path := range f.files {
		if path == prefix || filepath.HasPrefix(path, prefix+string(filepath.Separator)) {
			toRemove = append(toRemove, path)
		}
	}

	for _, path := range toRemove {
		delete(f.files, path)
	}

	return nil
}

func (f *MemFS) Symlink(ctx context.Context, oldname, newname string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if parent directory exists
	parent := filepath.Dir(newname)
	if parent != "." && parent != "/" {
		if _, exists := f.files[parent]; !exists {
			return fs.ErrNotExist
		}
	}

	f.files[newname] = &memFile{
		mode:    fs.ModeSymlink | 0777,
		modTime: time.Now(),
		symlink: oldname,
	}

	return nil
}

func (f *MemFS) Rename(ctx context.Context, oldname, newname string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, exists := f.files[oldname]
	if !exists {
		return fs.ErrNotExist
	}

	// Rename the file/directory itself
	f.files[newname] = file
	delete(f.files, oldname)

	// If it's a directory, also rename all children
	if file.isDir {
		oldPrefix := filepath.Clean(oldname) + string(filepath.Separator)
		newPrefix := filepath.Clean(newname) + string(filepath.Separator)

		// Find all children and rename them
		toRename := make(map[string]string)
		for path := range f.files {
			if len(path) > len(oldPrefix) && path[:len(oldPrefix)] == oldPrefix {
				// This is a child of the old directory
				newPath := newPrefix + path[len(oldPrefix):]
				toRename[path] = newPath
			}
		}

		// Perform the renames
		for oldPath, newPath := range toRename {
			f.files[newPath] = f.files[oldPath]
			delete(f.files, oldPath)
		}
	}

	return nil
}

func (f *MemFS) Exists(ctx context.Context, name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, exists := f.files[name]
	return exists
}

func (f *MemFS) IsDir(ctx context.Context, name string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, exists := f.files[name]
	if !exists {
		return false, fs.ErrNotExist
	}

	return file.isDir, nil
}

func (f *MemFS) IsSymlink(ctx context.Context, name string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, exists := f.files[name]
	if !exists {
		return false, fs.ErrNotExist
	}

	return file.symlink != "", nil
}

// memFileInfo implements domain.FileInfo
type memFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (i *memFileInfo) Name() string      { return i.name }
func (i *memFileInfo) Size() int64       { return i.size }
func (i *memFileInfo) Mode() fs.FileMode { return i.mode }
func (i *memFileInfo) ModTime() any      { return i.modTime }
func (i *memFileInfo) IsDir() bool       { return i.isDir }
func (i *memFileInfo) Sys() any          { return nil }

// memDirEntry implements domain.DirEntry
type memDirEntry struct {
	name  string
	isDir bool
	mode  fs.FileMode
}

func (e *memDirEntry) Name() string      { return e.name }
func (e *memDirEntry) IsDir() bool       { return e.isDir }
func (e *memDirEntry) Type() fs.FileMode { return e.mode }
func (e *memDirEntry) Info() (domain.FileInfo, error) {
	return &memFileInfo{
		name:  e.name,
		mode:  e.mode,
		isDir: e.isDir,
	}, nil
}
