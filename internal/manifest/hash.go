package manifest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/yaklabco/dot/internal/domain"
)

// ContentHasher computes content hashes for packages
type ContentHasher struct {
	fs domain.FS
}

// NewContentHasher creates a new content hasher
func NewContentHasher(fs domain.FS) *ContentHasher {
	return &ContentHasher{fs: fs}
}

// HashPackage computes content hash for entire package
// Hash is deterministic and based on file contents and paths
func (h *ContentHasher) HashPackage(ctx context.Context, pkgPath domain.PackagePath) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	hasher := sha256.New()

	// Collect all files in sorted order for determinism
	var files []string
	err := h.walkPackage(ctx, pkgPath.String(), pkgPath.String(), &files)
	if err != nil {
		return "", fmt.Errorf("failed to walk package: %w", err)
	}

	sort.Strings(files)

	// Hash each file's path and content with delimiter to prevent collisions
	// Delimiter prevents ambiguous concatenations like:
	// path="a", content="bc" vs path="ab", content="c"
	delimiter := []byte{0} // null byte separator

	for _, relPath := range files {
		fullPath := filepath.Join(pkgPath.String(), relPath)

		// Write path to hash
		if _, err := hasher.Write([]byte(relPath)); err != nil {
			return "", fmt.Errorf("failed to hash path: %w", err)
		}

		// Write delimiter
		if _, err := hasher.Write(delimiter); err != nil {
			return "", fmt.Errorf("failed to write delimiter: %w", err)
		}

		// Write content to hash
		data, err := h.fs.ReadFile(ctx, fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
		}

		if _, err := hasher.Write(data); err != nil {
			return "", fmt.Errorf("failed to hash content: %w", err)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// walkPackage collects regular files recursively
func (h *ContentHasher) walkPackage(ctx context.Context, root, current string, files *[]string) error {
	entries, err := h.fs.ReadDir(ctx, current)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fullPath := filepath.Join(current, entry.Name())

		if entry.IsDir() {
			if err := h.walkPackage(ctx, root, fullPath, files); err != nil {
				return err
			}
		} else if entry.Type().IsRegular() {
			// Store relative path for determinism
			relPath, err := filepath.Rel(root, fullPath)
			if err != nil {
				return err
			}
			*files = append(*files, relPath)
		}
		// Skip symlinks and other non-regular files
	}

	return nil
}
