package scanner

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/ignore"
)

// ScanConfig contains configuration options for scanning.
type ScanConfig struct {
	// PerPackageIgnore enables loading .dotignore files from packages
	PerPackageIgnore bool

	// MaxFileSize is the maximum file size in bytes (0 = no limit)
	MaxFileSize int64

	// Interactive enables interactive prompts for large files
	Interactive bool
}

// ScanPackage scans a single package directory.
// Returns a Package containing the package metadata and file tree.
//
// The scanner:
// 1. Verifies package directory exists
// 2. Scans the directory tree
// 3. Applies ignore patterns (filtered during tree scan)
// 4. Returns Package with tree
func ScanPackage(ctx context.Context, fs domain.FS, path domain.PackagePath, name string, ignoreSet *ignore.IgnoreSet) domain.Result[domain.Package] {
	// Check if package exists
	if !fs.Exists(ctx, path.String()) {
		return domain.Err[domain.Package](domain.ErrPackageNotFound{
			Package: name,
		})
	}

	// Scan the package directory tree
	pkgFilePath := domain.NewFilePath(path.String()).Unwrap()
	treeResult := ScanTree(ctx, fs, pkgFilePath)
	if treeResult.IsErr() {
		return domain.Err[domain.Package](treeResult.UnwrapErr())
	}

	tree := treeResult.Unwrap()

	// Filter tree based on ignore patterns
	filtered := filterTree(tree, ignoreSet)

	return domain.Ok(domain.Package{
		Name: name,
		Path: path,
		Tree: &filtered,
	})
}

// ScanPackageWithConfig scans a package with enhanced configuration options.
// Supports per-package .dotignore files, size filtering, and interactive prompts.
func ScanPackageWithConfig(ctx context.Context, fs domain.FS, path domain.PackagePath, name string, globalIgnoreSet *ignore.IgnoreSet, cfg ScanConfig) domain.Result[domain.Package] {
	// Check if package exists
	if !fs.Exists(ctx, path.String()) {
		return domain.Err[domain.Package](domain.ErrPackageNotFound{
			Package: name,
		})
	}

	// Build ignore set for this package by merging global and per-package patterns
	packageIgnoreSet := ignore.NewIgnoreSet()

	// Add all global patterns
	for _, pattern := range globalIgnoreSet.Patterns() {
		packageIgnoreSet.AddPattern(pattern)
	}

	// Load .dotignore if enabled
	if cfg.PerPackageIgnore {
		patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, path.String(), path.String())
		if err != nil {
			return domain.Err[domain.Package](fmt.Errorf("load .dotignore: %w", err))
		}

		// Add per-package patterns (these can include negation patterns)
		for _, pattern := range patterns {
			if err := packageIgnoreSet.Add(pattern); err != nil {
				return domain.Err[domain.Package](fmt.Errorf("invalid pattern %q in .dotignore: %w", pattern, err))
			}
		}
	}

	// Create prompter if size limit is enabled
	var prompter LargeFilePrompter
	if cfg.MaxFileSize > 0 {
		if cfg.Interactive && IsInteractive() {
			prompter = NewInteractivePrompter()
		} else {
			prompter = NewBatchPrompter()
		}
	}

	// Scan the package directory tree with config
	pkgFilePath := domain.NewFilePath(path.String()).Unwrap()
	var treeResult domain.Result[domain.Node]

	if cfg.MaxFileSize > 0 || prompter != nil {
		// Use size-aware scanning
		treeResult = ScanTreeWithConfig(ctx, fs, pkgFilePath, cfg.MaxFileSize, prompter)
	} else {
		// Use standard scanning (backward compatible)
		treeResult = ScanTree(ctx, fs, pkgFilePath)
	}

	if treeResult.IsErr() {
		return domain.Err[domain.Package](treeResult.UnwrapErr())
	}

	tree := treeResult.Unwrap()

	// Filter tree based on ignore patterns
	filtered := filterTree(tree, packageIgnoreSet)

	return domain.Ok(domain.Package{
		Name: name,
		Path: path,
		Tree: &filtered,
	})
}

// filterTree removes ignored files from a tree.
// Returns a new tree with ignored nodes filtered out.
func filterTree(node domain.Node, ignoreSet *ignore.IgnoreSet) domain.Node {
	// Check if this node should be ignored
	if ignoreSet.ShouldIgnore(node.Path.String()) {
		// Return empty node to be filtered by parent
		return domain.Node{}
	}

	// If directory, filter children
	if node.Type == domain.NodeDir {
		var filteredChildren []domain.Node
		for _, child := range node.Children {
			filtered := filterTree(child, ignoreSet)
			// Skip empty nodes (ignored)
			if filtered.Path.String() != "" {
				filteredChildren = append(filteredChildren, filtered)
			}
		}

		return domain.Node{
			Path:     node.Path,
			Type:     node.Type,
			Children: filteredChildren,
		}
	}

	// File or symlink - return as-is
	return node
}

// FilterTreeForTest exports filterTree for testing purposes.
func FilterTreeForTest(node domain.Node, ignoreSet *ignore.IgnoreSet) domain.Node {
	return filterTree(node, ignoreSet)
}
