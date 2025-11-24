// Package planner provides pure planning logic for computing operations.
package planner

import (
	"fmt"
	"path/filepath"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/scanner"
)

// LinkSpec specifies a desired symbolic link.
type LinkSpec struct {
	Source domain.FilePath   // Source file in package
	Target domain.TargetPath // Target location
}

// DirSpec specifies a desired directory.
type DirSpec struct {
	Path domain.FilePath
}

// DesiredState represents the desired filesystem state.
type DesiredState struct {
	Links map[string]LinkSpec // Key: target path
	Dirs  map[string]DirSpec  // Key: directory path
}

// PlanResult contains planning results with optional conflict resolution
type PlanResult struct {
	Desired  DesiredState
	Resolved *ResolveResult // Optional resolution results
}

// HasConflicts returns true if there are unresolved conflicts
func (pr PlanResult) HasConflicts() bool {
	return pr.Resolved != nil && pr.Resolved.HasConflicts()
}

// ComputeDesiredState computes desired state from packages.
// This is a pure function that determines what links and directories
// should exist based on the package contents.
//
// For each file in each package:
// 1. Compute relative path from package root
// 2. Apply dotfile translation (dot-vimrc -> .vimrc)
// 3. If packageNameMapping enabled, prepend translated package name
// 4. Join with target to get target path
// 5. Create LinkSpec (source -> target)
// 6. Create DirSpec for parent directories
func ComputeDesiredState(packages []domain.Package, target domain.TargetPath, packageNameMapping bool) domain.Result[DesiredState] {
	state := DesiredState{
		Links: make(map[string]LinkSpec),
		Dirs:  make(map[string]DirSpec),
	}

	for _, pkg := range packages {
		// Skip packages without trees
		if pkg.Tree == nil {
			continue
		}

		// Process all files in the package tree
		if err := processPackageTree(pkg, target, packageNameMapping, &state); err != nil {
			return domain.Err[DesiredState](err)
		}
	}

	return domain.Ok(state)
}

// processPackageTree walks a package tree and adds link/dir specs to state.
func processPackageTree(pkg domain.Package, target domain.TargetPath, packageNameMapping bool, state *DesiredState) error {
	return walkPackageFiles(*pkg.Tree, pkg.Path, pkg.Name, target, packageNameMapping, state)
}

// walkPackageFiles recursively processes files in a package tree.
func walkPackageFiles(node domain.Node, pkgRoot domain.PackagePath, pkgName string, target domain.TargetPath, packageNameMapping bool, state *DesiredState) error {
	// Process files only (not directories or symlinks)
	if node.Type == domain.NodeFile {
		// Compute relative path from package root
		relPathResult := relativePath(pkgRoot, node.Path)
		if relPathResult.IsErr() {
			return relPathResult.UnwrapErr()
		}
		relPath := relPathResult.Unwrap()

		// Apply dotfile translation to the relative path
		translated := translatePath(relPath)

		// Compute target path
		var targetPath domain.TargetPath
		if packageNameMapping {
			// Apply package name translation and prepend to path
			translatedPkgName := scanner.TranslatePackageName(pkgName)
			combinedPath := filepath.Join(translatedPkgName, translated)
			targetPath = target.Join(combinedPath)
		} else {
			// Legacy behavior: no package name mapping
			targetPath = target.Join(translated)
		}

		// Add link spec
		state.Links[targetPath.String()] = LinkSpec{
			Source: node.Path,
			Target: targetPath,
		}

		// Add parent directory specs
		if err := addParentDirs(targetPath, target, state); err != nil {
			return err
		}
	}

	// Recurse on children
	for _, child := range node.Children {
		if err := walkPackageFiles(child, pkgRoot, pkgName, target, packageNameMapping, state); err != nil {
			return err
		}
	}

	return nil
}

// addParentDirs adds directory specs for all parent directories of path.
func addParentDirs(path domain.TargetPath, target domain.TargetPath, state *DesiredState) error {
	current := path
	targetStr := target.String()

	for {
		parentResult := current.Parent()
		if parentResult.IsErr() {
			break
		}

		parent := parentResult.Unwrap()
		parentStr := parent.String()

		// Stop when we reach the target directory
		if parentStr == targetStr {
			break
		}

		// Add directory spec if not already present
		if _, exists := state.Dirs[parentStr]; !exists {
			// Convert TargetPath to FilePath for DirSpec storage
			dirPath := domain.NewFilePath(parentStr).Unwrap()
			state.Dirs[parentStr] = DirSpec{Path: dirPath}
		}

		current = parent
	}

	return nil
}

// Helper functions that will be moved to scanner package

func relativePath(base domain.PackagePath, target domain.FilePath) domain.Result[string] {
	// Simple relative path computation
	baseStr := base.String()
	targetStr := target.String()

	// If target doesn't start with base, error
	if len(targetStr) <= len(baseStr) {
		return domain.Err[string](domain.ErrInvalidPath{Path: targetStr, Reason: "not under base"})
	}

	// Strip base path and leading slash
	rel := targetStr[len(baseStr):]
	if len(rel) > 0 && rel[0] == '/' {
		rel = rel[1:]
	}

	return domain.Ok(rel)
}

func translatePath(path string) string {
	return scanner.TranslatePath(path)
}

// ComputeOperationsFromDesiredState converts desired state into operations
func ComputeOperationsFromDesiredState(desired DesiredState) []domain.Operation {
	// Preallocate slice for directories and links
	ops := make([]domain.Operation, 0, len(desired.Dirs)+len(desired.Links))

	// Create directory operations with content-based IDs for determinism
	for _, dirSpec := range desired.Dirs {
		id := domain.OperationID(fmt.Sprintf("dir-%s", dirSpec.Path.String()))
		ops = append(ops, domain.NewDirCreate(id, dirSpec.Path))
	}

	// Create link operations with content-based IDs for determinism
	for _, linkSpec := range desired.Links {
		id := domain.OperationID(fmt.Sprintf("link-%s->%s", linkSpec.Source.String(), linkSpec.Target.String()))
		ops = append(ops, domain.NewLinkCreate(id, linkSpec.Source, linkSpec.Target))
	}

	return ops
}
