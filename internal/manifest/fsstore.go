package manifest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yaklabco/dot/internal/domain"
)

const manifestFileName = ".dot-manifest.json"

// FSManifestStore implements ManifestStore using filesystem
type FSManifestStore struct {
	fs          domain.FS
	manifestDir string // Directory to store manifest (empty means use target directory)
}

// NewFSManifestStore creates filesystem-based manifest store.
// Manifest is stored in the target directory for backward compatibility.
func NewFSManifestStore(fs domain.FS) *FSManifestStore {
	return &FSManifestStore{
		fs:          fs,
		manifestDir: "", // Empty means use target directory
	}
}

// NewFSManifestStoreWithDir creates filesystem-based manifest store with custom directory.
// Manifest is stored in the specified manifestDir instead of target directory.
func NewFSManifestStoreWithDir(fs domain.FS, manifestDir string) *FSManifestStore {
	return &FSManifestStore{
		fs:          fs,
		manifestDir: manifestDir,
	}
}

// Load retrieves manifest from configured directory
func (s *FSManifestStore) Load(ctx context.Context, targetDir domain.TargetPath) domain.Result[Manifest] {
	if ctx.Err() != nil {
		return domain.Err[Manifest](ctx.Err())
	}

	manifestPath := s.getManifestPath(targetDir)

	data, err := s.fs.ReadFile(ctx, manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Missing manifest is not an error - return empty manifest
			return domain.Ok(New())
		}
		return domain.Err[Manifest](fmt.Errorf("failed to read manifest: %w", err))
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return domain.Err[Manifest](fmt.Errorf("failed to parse manifest: %w", err))
	}

	return domain.Ok(m)
}

// getManifestPath returns the full path to the manifest file.
// Uses manifestDir if configured, otherwise falls back to targetDir.
func (s *FSManifestStore) getManifestPath(targetDir domain.TargetPath) string {
	if s.manifestDir != "" {
		return filepath.Join(s.manifestDir, manifestFileName)
	}
	return filepath.Join(targetDir.String(), manifestFileName)
}

// Save persists manifest to configured directory
func (s *FSManifestStore) Save(ctx context.Context, targetDir domain.TargetPath, manifest Manifest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Update timestamp
	manifest.UpdatedAt = time.Now()

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestPath := s.getManifestPath(targetDir)

	// Ensure manifest directory exists
	manifestDir := filepath.Dir(manifestPath)
	if !s.fs.Exists(ctx, manifestDir) {
		if err := s.fs.MkdirAll(ctx, manifestDir, 0755); err != nil {
			return fmt.Errorf("failed to create manifest directory: %w", err)
		}
	}

	// Atomic write via temp file and rename
	tempPath := manifestPath + ".tmp"

	// Write to temp file
	if err := s.fs.WriteFile(ctx, tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp manifest: %w", err)
	}

	// Atomic rename
	if err := s.fs.Rename(ctx, tempPath, manifestPath); err != nil {
		// Ignore cleanup error: best-effort during error recovery.
		// Temp file (.dot-manifest.json.tmp) is harmless and will be
		// overwritten on next successful write operation.
		_ = s.fs.Remove(ctx, tempPath)
		return fmt.Errorf("failed to rename manifest: %w", err)
	}

	return nil
}
