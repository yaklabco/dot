// Package stateguard detects existing dot state on first run and lets
// the user choose how to proceed (continue, reset, or back up and reset).
package stateguard

import (
	"os"
	"path/filepath"

	"github.com/yaklabco/dot/internal/config"
)

const markerFile = ".dot-acknowledged"

// MarkerPath returns the full path to the acknowledgment marker file.
func MarkerPath() string {
	return config.GetXDGStatePath(filepath.Join("dot", markerFile))
}

// MarkerExists returns true if the acknowledgment marker file exists.
func MarkerExists() bool {
	_, err := os.Stat(MarkerPath())
	return err == nil
}

// WriteMarker creates the acknowledgment marker file, including parent directories.
func WriteMarker() error {
	path := MarkerPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, nil, 0644)
}
