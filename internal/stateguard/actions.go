package stateguard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/manifest"
)

// ExistingState describes the detected state of a previous dot installation.
type ExistingState struct {
	ManifestPath string
	PackageCount int
	LinkCount    int
}

// DetectExistingState loads the manifest and counts packages/links.
// Returns nil if no manifest exists or the manifest is empty.
func DetectExistingState(ctx context.Context, manifestDir, targetDir string) (*ExistingState, error) {
	store := manifest.NewFSManifestStoreWithDir(adapters.NewOSFilesystem(), manifestDir)
	targetPath := domain.NewTargetPath(targetDir)
	if targetPath.IsErr() {
		return nil, targetPath.UnwrapErr()
	}

	result := store.Load(ctx, targetPath.Unwrap())
	if result.IsErr() {
		return nil, result.UnwrapErr()
	}

	m := result.Unwrap()
	if len(m.Packages) == 0 {
		return nil, nil
	}

	linkCount := 0
	for _, pkg := range m.Packages {
		linkCount += len(pkg.Links)
	}

	return &ExistingState{
		ManifestPath: filepath.Join(manifestDir, ".dot-manifest.json"),
		PackageCount: len(m.Packages),
		LinkCount:    linkCount,
	}, nil
}

// ActionContinue writes the marker and returns.
func ActionContinue() error {
	return WriteMarker()
}

// ActionFresh removes managed symlinks, deletes the manifest, and writes the marker.
// Regular files referenced in the manifest are preserved (only symlinks are removed).
func ActionFresh(manifestDir, targetDir string) error {
	ctx := context.Background()
	store := manifest.NewFSManifestStoreWithDir(adapters.NewOSFilesystem(), manifestDir)
	targetPath := domain.NewTargetPath(targetDir)
	if targetPath.IsErr() {
		return targetPath.UnwrapErr()
	}

	result := store.Load(ctx, targetPath.Unwrap())
	if result.IsErr() {
		return result.UnwrapErr()
	}

	m := result.Unwrap()
	var errs []error
	for _, pkg := range m.Packages {
		pkgTargetDir := targetDir
		if pkg.TargetDir != "" {
			pkgTargetDir = pkg.TargetDir
		}
		for _, link := range pkg.Links {
			linkPath := filepath.Join(pkgTargetDir, link)
			fi, err := os.Lstat(linkPath)
			if err != nil {
				continue // Skip non-existent links
			}
			if fi.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(linkPath); err != nil {
					errs = append(errs, fmt.Errorf("remove symlink %s: %w", linkPath, err))
				}
			}
		}
	}

	// Remove manifest file
	manifestPath := filepath.Join(manifestDir, ".dot-manifest.json")
	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove manifest: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return WriteMarker()
}

// ActionBackupAndFresh copies the manifest and config to a backup directory,
// then delegates to ActionFresh. Returns the backup directory path.
func ActionBackupAndFresh(manifestDir, targetDir, configPath, homeDir string) (string, error) {
	backupDir, err := resolveBackupDir(homeDir)
	if err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	// Copy manifest
	manifestSrc := filepath.Join(manifestDir, ".dot-manifest.json")
	if _, err := os.Stat(manifestSrc); err == nil {
		manifestBackupDir := filepath.Join(backupDir, "manifest")
		if err := os.MkdirAll(manifestBackupDir, 0755); err != nil {
			return "", fmt.Errorf("create manifest backup dir: %w", err)
		}
		if err := copyFile(manifestSrc, filepath.Join(manifestBackupDir, ".dot-manifest.json")); err != nil {
			return "", fmt.Errorf("backup manifest: %w", err)
		}
	}

	// Copy config
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			if err := copyFile(configPath, filepath.Join(backupDir, filepath.Base(configPath))); err != nil {
				return "", fmt.Errorf("backup config: %w", err)
			}
		}
	}

	if err := ActionFresh(manifestDir, targetDir); err != nil {
		return backupDir, err
	}

	return backupDir, nil
}

// PrintSummary writes the detected state summary to the writer.
func PrintSummary(w io.Writer, state *ExistingState) {
	fmt.Fprintln(w, "Existing dot installation detected.")
	fmt.Fprintf(w, "  %d packages · %d symlinks · manifest: %s\n\n",
		state.PackageCount, state.LinkCount, state.ManifestPath)
}

func resolveBackupDir(homeDir string) (string, error) {
	today := time.Now().Format("20060102")
	base := filepath.Join(homeDir, ".dot-backup-"+today)

	// Try the base name first
	if _, err := os.Stat(base); os.IsNotExist(err) {
		if err := os.MkdirAll(base, 0755); err != nil {
			return "", err
		}
		return base, nil
	}

	// Append -2, -3, etc. on collision
	for i := 2; i <= 99; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			if err := os.MkdirAll(candidate, 0755); err != nil {
				return "", err
			}
			return candidate, nil
		}
	}

	return "", fmt.Errorf("too many backup directories for %s", base)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
