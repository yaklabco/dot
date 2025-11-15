package ignore

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/domain"
)

// LoadDotignoreFile loads patterns from a .dotignore file.
// Returns nil patterns (no error) if the file does not exist.
// Empty lines and lines starting with # are treated as comments and skipped.
func LoadDotignoreFile(ctx context.Context, fs domain.FS, path string) ([]string, error) {
	// File not existing is acceptable
	if !fs.Exists(ctx, path) {
		return nil, nil
	}

	content, err := fs.ReadFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("read .dotignore: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	patterns := make([]string, 0, len(lines))

	for lineNum, line := range lines {
		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Validate pattern is not just whitespace after trim
		if len(line) == 0 {
			continue
		}

		// Check for invalid patterns (multiple ! prefixes, etc.)
		if strings.HasPrefix(line, "!!") {
			return nil, fmt.Errorf("invalid pattern at line %d: multiple ! prefixes not allowed", lineNum+1)
		}

		patterns = append(patterns, line)
	}

	return patterns, nil
}

// LoadDotignoreWithInheritance loads .dotignore files from startPath up to rootPath.
// Files closer to startPath have higher priority (patterns are prepended).
// This implements subdirectory inheritance similar to .gitignore behavior.
//
// Example: For path /packages/vim/colors, it checks:
//  1. /packages/vim/colors/.dotignore (highest priority)
//  2. /packages/vim/.dotignore
//  3. /packages/.dotignore (lowest priority, assuming rootPath=/packages)
func LoadDotignoreWithInheritance(ctx context.Context, fs domain.FS, startPath, rootPath string) ([]string, error) {
	var allPatterns []string
	currentPath := startPath

	// Normalize paths for consistent comparison
	rootPath = filepath.Clean(rootPath)
	currentPath = filepath.Clean(currentPath)

	// Track visited paths to prevent infinite loops
	visited := make(map[string]bool)

	for {
		// Prevent infinite loops
		if visited[currentPath] {
			break
		}
		visited[currentPath] = true

		// Load .dotignore from current directory
		dotignorePath := filepath.Join(currentPath, ".dotignore")
		patterns, err := LoadDotignoreFile(ctx, fs, dotignorePath)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", dotignorePath, err)
		}

		// Prepend patterns (closest file has highest priority)
		if len(patterns) > 0 {
			allPatterns = append(patterns, allPatterns...)
		}

		// Stop at root or filesystem root
		if currentPath == rootPath || currentPath == "/" || currentPath == "." {
			break
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)

		// Detect if we've reached the top (Dir returns same path)
		if parentPath == currentPath {
			break
		}

		currentPath = parentPath
	}

	return allPatterns, nil
}
