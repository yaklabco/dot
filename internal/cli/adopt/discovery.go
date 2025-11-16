// Package adopt provides interactive file adoption.
//
// This file contains discovery logic for the interactive adopt workflow that is
// tightly coupled to the interactive UI and filesystem operations. It is excluded
// from coverage requirements as it requires integration testing.
package adopt

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/pkg/dot"
)

// DefaultScanDirs returns the default directories to scan for dotfiles.
func DefaultScanDirs(homeDir string) []string {
	return []string{
		homeDir,
		filepath.Join(homeDir, ".config"),
	}
}

// DefaultExcludeDirs returns commonly excluded directories.
func DefaultExcludeDirs() []string {
	return []string{
		".git",
		".svn",
		".hg",
		"node_modules",
		".cache",
		".cargo",
		".rustup",
		".npm",
		".yarn",
		".gradle",
		".m2",
		"Library",
		"Applications",
		"Documents",
		"Downloads",
		"Desktop",
		"Pictures",
		"Music",
		"Videos",
		"Movies",
		"Public",
		".Trash",
		"bin",
		"opt",
		"tmp",
		".local",
		".mozilla",
		".thunderbird",
	}
}

// DiscoverDotfiles scans directories for adoptable dotfiles.
// Returns candidates excluding already-managed files.
func DiscoverDotfiles(
	ctx context.Context,
	fs domain.FS,
	opts DiscoveryOptions,
	client *dot.Client,
	targetDir string,
) ([]DotfileCandidate, error) {
	var candidates []DotfileCandidate
	excludeDirs := makeExcludeMap(opts.ExcludeDirs)

	// Get managed paths from client
	managedPaths, err := getManagedPaths(ctx, client, targetDir)
	if err != nil {
		// Log warning but continue - we'll just not filter managed paths
		managedPaths = make(map[string]bool)
	}

	// Get home directory to check if scanDir is $HOME or .config
	homeDir := filepath.Dir(filepath.Join(targetDir, ".config"))

	// Scan each directory
	for _, scanDir := range opts.ScanDirs {
		// Check if directory is .config or $HOME
		isConfigDir := strings.HasSuffix(scanDir, ".config")
		isHomeDir := scanDir == homeDir || scanDir == targetDir

		entries, err := fs.ReadDir(ctx, scanDir)
		if err != nil {
			continue // Skip inaccessible directories
		}

		for _, entry := range entries {
			name := entry.Name()
			fullPath := filepath.Join(scanDir, name)

			// Skip files that shouldn't be discovered
			if shouldSkipFile(name, fullPath, isHomeDir, isConfigDir, managedPaths, excludeDirs) {
				continue
			}

			// Get file info
			info, err := fs.Stat(ctx, fullPath)
			if err != nil {
				continue
			}

			// Calculate actual size (including directory contents)
			actualSize := info.Size()
			if entry.IsDir() {
				actualSize = calculateDirectorySize(ctx, fs, fullPath)
			}

			// Skip if over size limit
			if opts.MaxFileSize > 0 && actualSize > opts.MaxFileSize {
				continue
			}

			// Create candidate with actual size
			candidate := createCandidateWithSize(name, fullPath, targetDir, info, entry.IsDir(), actualSize)
			candidates = append(candidates, candidate)
		}
	}

	return candidates, nil
}

// getManagedPaths retrieves all paths currently managed by dot.
func getManagedPaths(ctx context.Context, client *dot.Client, targetDir string) (map[string]bool, error) {
	status, err := client.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("get status: %w", err)
	}

	managedPaths := make(map[string]bool)
	for _, pkg := range status.Packages {
		for _, link := range pkg.Links {
			absPath := filepath.Join(targetDir, link)
			managedPaths[absPath] = true
		}
	}

	return managedPaths, nil
}

// shouldSkipFile determines if a file should be skipped during discovery.
func shouldSkipFile(name, fullPath string, isHomeDir, isConfigDir bool, managedPaths, excludeDirs map[string]bool) bool {
	// Skip if already managed
	if managedPaths[fullPath] {
		return true
	}

	// Skip excluded directories
	if excludeDirs[name] {
		return true
	}

	// Skip .config directory itself when scanning $HOME
	if isHomeDir && name == ".config" {
		return true
	}

	// Skip .dot-manifest.json (dot's own manifest file)
	if name == ".dot-manifest.json" {
		return true
	}

	// For $HOME: only include dotfiles (starting with .)
	if !isConfigDir && !strings.HasPrefix(name, ".") {
		return true
	}

	return false
}

// createCandidate creates a DotfileCandidate from file information.
func createCandidate(name, fullPath, targetDir string, info domain.FileInfo, isDir bool) DotfileCandidate {
	return createCandidateWithSize(name, fullPath, targetDir, info, isDir, info.Size())
}

// createCandidateWithSize creates a DotfileCandidate with a specific size.
func createCandidateWithSize(name, fullPath, targetDir string, info domain.FileInfo, isDir bool, actualSize int64) DotfileCandidate {
	relPath := strings.TrimPrefix(fullPath, targetDir+string(filepath.Separator))
	modTime, ok := info.ModTime().(time.Time)
	if !ok {
		modTime = time.Time{} // Zero time if assertion fails
	}

	candidate := DotfileCandidate{
		Path:    fullPath,
		RelPath: relPath,
		Size:    actualSize,
		ModTime: modTime,
		IsDir:   isDir,
	}

	// Categorize and suggest package name
	candidate.Category = categorizeFile(name)
	candidate.SuggestedPkg = suggestPackageName(name, candidate.Category)

	return candidate
}

// calculateDirectorySize recursively calculates the total size of a directory.
func calculateDirectorySize(ctx context.Context, fs domain.FS, dirPath string) int64 {
	var totalSize int64

	entries, err := fs.ReadDir(ctx, dirPath)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		info, err := fs.Stat(ctx, fullPath)
		if err != nil {
			continue
		}

		if entry.IsDir() {
			// Recursively calculate subdirectory size
			totalSize += calculateDirectorySize(ctx, fs, fullPath)
		} else {
			totalSize += info.Size()
		}
	}

	return totalSize
}

// makeExcludeMap converts slice to map for O(1) lookup.
func makeExcludeMap(excludeDirs []string) map[string]bool {
	m := make(map[string]bool, len(excludeDirs))
	for _, dir := range excludeDirs {
		m[dir] = true
	}
	return m
}
