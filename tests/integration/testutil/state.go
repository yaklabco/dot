package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FileState represents the state of a file or directory.
type FileState struct {
	Path      string
	IsDir     bool
	IsSymlink bool
	Target    string // for symlinks
	Mode      os.FileMode
	Size      int64
}

// StateSnapshot captures the state of a directory tree.
type StateSnapshot struct {
	Root  string
	Files []FileState
}

// CaptureState captures the current state of a directory tree.
func CaptureState(t *testing.T, root string) *StateSnapshot {
	t.Helper()

	var files []FileState

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip root directory itself
		if path == root {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		state := FileState{
			Path:  relPath,
			IsDir: info.IsDir(),
			Mode:  info.Mode(),
			Size:  info.Size(),
		}

		if info.Mode()&os.ModeSymlink != 0 {
			state.IsSymlink = true
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			state.Target = target
		}

		files = append(files, state)
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to capture state: %v", err)
	}

	// Sort for consistent comparison
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return &StateSnapshot{
		Root:  root,
		Files: files,
	}
}

// CompareStates compares two state snapshots and returns differences.
func CompareStates(t *testing.T, before, after *StateSnapshot) []string {
	t.Helper()

	beforeMap := createStateMap(before)
	afterMap := createStateMap(after)

	// Estimate capacity: worst case each file could be added, removed, or modified
	estimatedCap := len(beforeMap) + len(afterMap)
	diffs := make([]string, 0, estimatedCap)
	diffs = append(diffs, findAddedFiles(beforeMap, afterMap)...)
	diffs = append(diffs, findRemovedFiles(beforeMap, afterMap)...)
	diffs = append(diffs, findModifiedFiles(beforeMap, afterMap)...)

	sort.Strings(diffs)
	return diffs
}

func createStateMap(snapshot *StateSnapshot) map[string]FileState {
	stateMap := make(map[string]FileState)
	for _, f := range snapshot.Files {
		stateMap[f.Path] = f
	}
	return stateMap
}

func findAddedFiles(beforeMap, afterMap map[string]FileState) []string {
	var diffs []string
	for path, afterState := range afterMap {
		if _, exists := beforeMap[path]; !exists {
			if afterState.IsSymlink {
				diffs = append(diffs, fmt.Sprintf("+ symlink %s -> %s", path, afterState.Target))
			} else if afterState.IsDir {
				diffs = append(diffs, fmt.Sprintf("+ dir %s", path))
			} else {
				diffs = append(diffs, fmt.Sprintf("+ file %s", path))
			}
		}
	}
	return diffs
}

func findRemovedFiles(beforeMap, afterMap map[string]FileState) []string {
	var diffs []string
	for path, beforeState := range beforeMap {
		if _, exists := afterMap[path]; !exists {
			if beforeState.IsSymlink {
				diffs = append(diffs, fmt.Sprintf("- symlink %s", path))
			} else if beforeState.IsDir {
				diffs = append(diffs, fmt.Sprintf("- dir %s", path))
			} else {
				diffs = append(diffs, fmt.Sprintf("- file %s", path))
			}
		}
	}
	return diffs
}

func findModifiedFiles(beforeMap, afterMap map[string]FileState) []string {
	var diffs []string
	for path, beforeState := range beforeMap {
		if afterState, exists := afterMap[path]; exists {
			if beforeState.IsSymlink != afterState.IsSymlink {
				diffs = append(diffs, fmt.Sprintf("~ %s: type changed", path))
			} else if beforeState.IsSymlink && beforeState.Target != afterState.Target {
				diffs = append(diffs, fmt.Sprintf("~ symlink %s: %s -> %s", path, beforeState.Target, afterState.Target))
			} else if !beforeState.IsDir && !afterState.IsDir && beforeState.Size != afterState.Size {
				diffs = append(diffs, fmt.Sprintf("~ file %s: size changed", path))
			}
		}
	}
	return diffs
}

// AssertStateUnchanged verifies that state has not changed.
func AssertStateUnchanged(t *testing.T, before, after *StateSnapshot) {
	t.Helper()
	diffs := CompareStates(t, before, after)
	assert.Empty(t, diffs, "state should be unchanged, but found differences")
}

// AssertStateChanges verifies expected state changes occurred.
func AssertStateChanges(t *testing.T, before, after *StateSnapshot, expectedChanges []string) {
	t.Helper()
	diffs := CompareStates(t, before, after)

	sort.Strings(expectedChanges)
	sort.Strings(diffs)

	assert.Equal(t, expectedChanges, diffs, "state changes mismatch")
}

// CountFiles counts the number of files (not directories) in a snapshot.
func (s *StateSnapshot) CountFiles() int {
	count := 0
	for _, f := range s.Files {
		if !f.IsDir {
			count++
		}
	}
	return count
}

// CountSymlinks counts the number of symlinks in a snapshot.
func (s *StateSnapshot) CountSymlinks() int {
	count := 0
	for _, f := range s.Files {
		if f.IsSymlink {
			count++
		}
	}
	return count
}

// CountDirs counts the number of directories in a snapshot.
func (s *StateSnapshot) CountDirs() int {
	count := 0
	for _, f := range s.Files {
		if f.IsDir {
			count++
		}
	}
	return count
}

// HasPath checks if a path exists in the snapshot.
func (s *StateSnapshot) HasPath(path string) bool {
	for _, f := range s.Files {
		if f.Path == path {
			return true
		}
	}
	return false
}

// GetState returns the state of a specific path.
func (s *StateSnapshot) GetState(path string) (FileState, bool) {
	for _, f := range s.Files {
		if f.Path == path {
			return f, true
		}
	}
	return FileState{}, false
}
