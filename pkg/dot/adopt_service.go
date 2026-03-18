package dot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/scanner"
)

// AdoptService handles file adoption operations.
type AdoptService struct {
	fs          FS
	logger      Logger
	executor    *executor.Executor
	manifestSvc *ManifestService
	packageDir  string
	targetDir   string
	dryRun      bool
}

// newAdoptService creates a new adopt service.
func newAdoptService(
	fs FS,
	logger Logger,
	exec *executor.Executor,
	manifestSvc *ManifestService,
	packageDir string,
	targetDir string,
	dryRun bool,
) *AdoptService {
	return &AdoptService{
		fs:          fs,
		logger:      logger,
		executor:    exec,
		manifestSvc: manifestSvc,
		packageDir:  packageDir,
		targetDir:   targetDir,
		dryRun:      dryRun,
	}
}

// GetManagedPaths returns a map of all paths currently managed by dot.
// The map keys are absolute paths that are already symlinked.
// This is useful for filtering out already-managed files during discovery.
func (s *AdoptService) GetManagedPaths(ctx context.Context) (map[string]struct{}, error) {
	// Load manifest
	targetPath := NewTargetPath(s.targetDir)
	if targetPath.IsErr() {
		return nil, targetPath.UnwrapErr()
	}

	result := s.manifestSvc.Load(ctx, targetPath.Unwrap())
	if result.IsErr() {
		err := result.UnwrapErr()
		// If manifest doesn't exist, return empty map
		if os.IsNotExist(err) {
			return make(map[string]struct{}), nil
		}
		return nil, fmt.Errorf("load manifest: %w", err)
	}

	m := result.Unwrap()

	// Extract all managed paths
	managedPaths := make(map[string]struct{})
	for _, pkg := range m.Packages {
		for _, link := range pkg.Links {
			// Convert relative link to absolute path
			absPath := filepath.Join(s.targetDir, link)
			managedPaths[absPath] = struct{}{}
		}
	}

	return managedPaths, nil
}

// resolveAdoptPath resolves a file path for adoption based on context.
// Resolution rules:
//   - Absolute paths (starting with / or ~): used as-is
//   - Relative paths starting with ./: resolved from current working directory
//   - Bare paths: resolved from target directory (default behavior)
func (s *AdoptService) resolveAdoptPath(ctx context.Context, file string) (string, error) {
	// Already absolute path
	if filepath.IsAbs(file) {
		return file, nil
	}

	// Tilde expansion
	if strings.HasPrefix(file, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot expand ~: %w", err)
		}
		if file == "~" {
			return home, nil
		}
		if strings.HasPrefix(file, "~/") {
			return filepath.Join(home, file[2:]), nil
		}
		// Reject malformed tilde paths like ~user/path or ~abc
		return "", fmt.Errorf("unsupported tilde expansion: %s (only ~ and ~/ are supported)", file)
	}

	// Explicit relative path from pwd (starts with ./ or ../)
	if strings.HasPrefix(file, "./") || strings.HasPrefix(file, "../") {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot get working directory: %w", err)
		}
		absPath, err := filepath.Abs(filepath.Join(cwd, file))
		if err != nil {
			return "", fmt.Errorf("cannot resolve path: %w", err)
		}
		return absPath, nil
	}

	// Bare path: resolve from target directory (backward compatible default)
	return filepath.Join(s.targetDir, file), nil
}

// Adopt moves existing files from target into package then creates symlinks.
func (s *AdoptService) Adopt(ctx context.Context, files []string, pkg string) error {
	plan, err := s.PlanAdopt(ctx, files, pkg)
	if err != nil {
		return err
	}
	if len(plan.Operations) == 0 {
		s.logger.Info(ctx, "nothing_to_adopt")
		return nil
	}
	if s.dryRun {
		s.logger.Info(ctx, "dry_run_plan", "operations", len(plan.Operations))
		return nil
	}
	result := s.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}
	execResult := result.Unwrap()
	if !execResult.Success() {
		return ErrMultiple{Errors: execResult.Errors}
	}
	// Update manifest with source=adopted
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	if err := s.manifestSvc.UpdateWithSource(ctx, targetPathResult.Unwrap(), s.packageDir, []string{pkg}, plan, manifest.SourceAdopted); err != nil {
		s.logger.Warn(ctx, "failed_to_update_manifest", "error", err)
	}
	return nil
}

// PlanAdopt computes the execution plan for adopting files.
func (s *AdoptService) PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error) {
	packagePathResult := NewPackagePath(s.packageDir)
	if !packagePathResult.IsOk() {
		return Plan{}, packagePathResult.UnwrapErr()
	}
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}

	// Check if package directory exists, create if not
	pkgPath := filepath.Join(s.packageDir, pkg)
	operations := make([]Operation, 0, len(files)*2+1)

	if !s.fs.Exists(ctx, pkgPath) {
		// Add operation to create package directory
		pkgPathResult := NewFilePath(pkgPath)
		if pkgPathResult.IsErr() {
			return Plan{}, fmt.Errorf("invalid package path %s: %w", pkgPath, pkgPathResult.UnwrapErr())
		}
		dirID := OperationID(fmt.Sprintf("adopt-create-pkg-%s", pkg))
		operations = append(operations, NewDirCreate(dirID, pkgPathResult.Unwrap()))
	}

	for _, file := range files {
		fileOps, err := s.planAdoptFile(ctx, file, pkgPath)
		if err != nil {
			return Plan{}, err
		}
		operations = append(operations, fileOps...)
	}

	// Build PackageOperations map for manifest tracking
	packageOps := make(map[string][]OperationID)
	opIDs := make([]OperationID, 0, len(operations))
	for _, op := range operations {
		opIDs = append(opIDs, op.ID())
	}
	packageOps[pkg] = opIDs

	return Plan{
		Operations:        operations,
		PackageOperations: packageOps,
		Metadata: PlanMetadata{
			PackageCount:   1,
			OperationCount: len(operations),
		},
	}, nil
}

// planAdoptFile plans the operations for adopting a single file or directory.
func (s *AdoptService) planAdoptFile(ctx context.Context, file, pkgPath string) ([]Operation, error) {
	sourceFile, err := s.resolveAdoptPath(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %s: %w", file, err)
	}

	if !s.fs.Exists(ctx, sourceFile) {
		return nil, ErrSourceNotFound{Path: sourceFile}
	}

	if err := s.validateAdoptSource(ctx, file, sourceFile); err != nil {
		return nil, err
	}

	isDir, err := s.fs.IsDir(ctx, sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to check if directory: %w", err)
	}

	if isDir {
		return s.createDirectoryAdoptOperations(ctx, sourceFile, pkgPath, file)
	}

	// Compute relative path from target dir to preserve nested directory structure.
	// For a file at .config/nvim/init.vim, we translate each path component
	// (e.g., .config -> dot-config) and use the full relative path in the package.
	relPath, err := filepath.Rel(s.targetDir, sourceFile)
	if err != nil {
		relPath = filepath.Base(file)
	}
	adoptedRelPath := translatePathComponents(relPath)
	destFile := filepath.Join(pkgPath, adoptedRelPath)

	if s.fs.Exists(ctx, destFile) {
		return nil, fmt.Errorf("cannot adopt %s: file %q already exists in package %q (use 'dot unmanage %s --purge' first to remove the existing package)", file, adoptedRelPath, filepath.Base(pkgPath), filepath.Base(pkgPath))
	}

	operations := s.planIntermediateDirs(ctx, adoptedRelPath, pkgPath)

	sourceLinkPathResult := NewTargetPath(sourceFile)
	if !sourceLinkPathResult.IsOk() {
		return nil, sourceLinkPathResult.UnwrapErr()
	}
	destPathResult := NewFilePath(destFile)
	if !destPathResult.IsOk() {
		return nil, destPathResult.UnwrapErr()
	}

	moveID := OperationID(fmt.Sprintf("adopt-move-%s", file))
	linkID := OperationID(fmt.Sprintf("adopt-link-%s", file))

	operations = append(operations,
		FileMove{
			OpID:   moveID,
			Source: sourceLinkPathResult.Unwrap(),
			Dest:   destPathResult.Unwrap(),
		},
		NewLinkCreate(linkID, destPathResult.Unwrap(), sourceLinkPathResult.Unwrap()),
	)
	return operations, nil
}

// planIntermediateDirs creates DirCreate operations for all missing intermediate
// directories between pkgPath and the file's parent directory.
func (s *AdoptService) planIntermediateDirs(ctx context.Context, adoptedRelPath, pkgPath string) []Operation {
	relDir := filepath.Dir(adoptedRelPath)
	if relDir == "." {
		return nil
	}

	// Collect directories from shallowest to deepest
	var dirsToCreate []string
	cur := relDir
	for cur != "." && cur != "" {
		dirPath := filepath.Join(pkgPath, cur)
		if !s.fs.Exists(ctx, dirPath) {
			dirsToCreate = append([]string{cur}, dirsToCreate...)
		}
		cur = filepath.Dir(cur)
	}

	var ops []Operation
	for _, dir := range dirsToCreate {
		dirFullPath := filepath.Join(pkgPath, dir)
		dirResult := NewFilePath(dirFullPath)
		if dirResult.IsOk() {
			dirID := OperationID(fmt.Sprintf("adopt-create-dir-%s", dir))
			ops = append(ops, NewDirCreate(dirID, dirResult.Unwrap()))
		}
	}
	return ops
}

// createDirectoryAdoptOperations creates operations to adopt a directory's contents.
// Moves directory CONTENTS into package root (flat structure), not the directory itself.
func (s *AdoptService) createDirectoryAdoptOperations(ctx context.Context, sourceDir, pkgPath, originalPath string) ([]Operation, error) {
	var operations []Operation

	// Recursively collect all files in the directory
	filesToMove, err := s.collectDirectoryFiles(ctx, sourceDir, "")
	if err != nil {
		return nil, fmt.Errorf("failed to collect directory files: %w", err)
	}

	// First pass: Create all directories
	for _, relPath := range filesToMove {
		sourcePath := filepath.Join(sourceDir, relPath)

		isDir, _ := s.fs.IsDir(ctx, sourcePath)
		if isDir {
			translatedPath := translatePathComponents(relPath)
			destPath := filepath.Join(pkgPath, translatedPath)

			destResult := NewFilePath(destPath)
			if !destResult.IsOk() {
				continue
			}

			dirID := OperationID(fmt.Sprintf("adopt-create-dir-%s", translatedPath))
			operations = append(operations, NewDirCreate(dirID, destResult.Unwrap()))
		}
	}

	// Second pass: Move all files and track subdirectories
	var subdirs []string
	for _, relPath := range filesToMove {
		sourcePath := filepath.Join(sourceDir, relPath)

		isDir, _ := s.fs.IsDir(ctx, sourcePath)
		if isDir {
			// Track subdirectories for deletion later
			subdirs = append(subdirs, relPath)
		} else {
			translatedPath := translatePathComponents(relPath)
			destPath := filepath.Join(pkgPath, translatedPath)

			sourceResult := NewTargetPath(sourcePath)
			if !sourceResult.IsOk() {
				continue
			}

			destResult := NewFilePath(destPath)
			if !destResult.IsOk() {
				continue
			}

			moveID := OperationID(fmt.Sprintf("adopt-move-content-%s", relPath))
			operations = append(operations, FileMove{
				OpID:   moveID,
				Source: sourceResult.Unwrap(),
				Dest:   destResult.Unwrap(),
			})
		}
	}

	// Third pass: Delete subdirectories in reverse order (deepest first)
	// This ensures child directories are deleted before parents
	for i := len(subdirs) - 1; i >= 0; i-- {
		subdirPath := filepath.Join(sourceDir, subdirs[i])
		subdirResult := NewFilePath(subdirPath)
		if subdirResult.IsOk() {
			delID := OperationID(fmt.Sprintf("adopt-remove-subdir-%s", subdirs[i]))
			operations = append(operations, NewDirDelete(delID, subdirResult.Unwrap()))
		}
	}

	// Remove the now-empty source directory
	sourceDirResult := NewTargetPath(sourceDir)
	if !sourceDirResult.IsOk() {
		return nil, sourceDirResult.UnwrapErr()
	}
	sourceDirPath := sourceDirResult.Unwrap()

	// Delete the original directory (now empty after moving contents)
	delID := OperationID(fmt.Sprintf("adopt-remove-empty-%s", originalPath))
	sourceDirFilePath := NewFilePath(sourceDir)
	if sourceDirFilePath.IsOk() {
		operations = append(operations, NewDirDelete(delID, sourceDirFilePath.Unwrap()))
	}

	// Create symlink from original location to package root
	pkgRootResult := NewFilePath(pkgPath)
	if !pkgRootResult.IsOk() {
		return nil, pkgRootResult.UnwrapErr()
	}

	linkID := OperationID(fmt.Sprintf("adopt-link-%s", originalPath))
	operations = append(operations, NewLinkCreate(linkID, pkgRootResult.Unwrap(), sourceDirPath))

	return operations, nil
}

// collectDirectoryFiles recursively collects all file paths in a directory.
// Returns paths relative to the root directory.
func (s *AdoptService) collectDirectoryFiles(ctx context.Context, dir, prefix string) ([]string, error) {
	var files []string

	entries, err := s.fs.ReadDir(ctx, dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		relPath := entry.Name()
		if prefix != "" {
			relPath = filepath.Join(prefix, entry.Name())
		}

		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Add directory itself (will be created as empty dir in package)
			files = append(files, relPath)

			// Recursively collect files in subdirectory
			subFiles, err := s.collectDirectoryFiles(ctx, fullPath, relPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			// Regular file
			files = append(files, relPath)
		}
	}

	return files, nil
}

// translatePathComponents applies dotfile translation to each component of a path.
// ".cache/data" → "dot-cache/data"
// "regular/.hidden" → "regular/dot-hidden"
func translatePathComponents(path string) string {
	if path == "" || path == "." {
		return path
	}

	components := splitPath(path)

	for i, comp := range components {
		components[i] = scanner.UntranslateDotfile(comp)
	}

	return filepath.Join(components...)
}

// splitPath splits a file path into components.
func splitPath(path string) []string {
	var components []string
	for {
		dir, file := filepath.Split(path)
		if file != "" {
			components = append([]string{file}, components...)
		}
		if dir == "" || dir == "/" {
			break
		}
		path = filepath.Clean(dir)
	}
	return components
}

// validateAdoptSource validates that the source file can be adopted.
// Checks if it's a symlink and handles accordingly.
func (s *AdoptService) validateAdoptSource(ctx context.Context, originalPath, resolvedPath string) error {
	// Check if source is a symlink (after existence check to avoid lstat errors)
	isSymlink, symlinkTarget, err := s.checkIfSymlink(ctx, resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to check symlink status: %w", err)
	}

	if isSymlink {
		// Check if it points to our package directory
		// Normalize paths to avoid false positives with similar directory names
		pkgRoot := filepath.Clean(s.packageDir)
		target := filepath.Clean(symlinkTarget)

		if target == pkgRoot || strings.HasPrefix(target, pkgRoot+string(os.PathSeparator)) {
			return fmt.Errorf("cannot adopt %s: already managed by dot (symlink to %s)", originalPath, symlinkTarget)
		}
		// Warn if symlink to other location
		s.logger.Warn(ctx, "adopting_symlink", "path", originalPath, "target", symlinkTarget)
	}

	return nil
}

// checkIfSymlink checks if path is a symlink and returns its target.
// Returns (isSymlink, target, error).
func (s *AdoptService) checkIfSymlink(ctx context.Context, path string) (bool, string, error) {
	info, err := s.fs.Lstat(ctx, path)
	if err != nil {
		return false, "", err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false, "", nil
	}

	target, err := s.fs.ReadLink(ctx, path)
	if err != nil {
		return true, "", err
	}

	// Resolve relative paths
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}

	return true, target, nil
}
