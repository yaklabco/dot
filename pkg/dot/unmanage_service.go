package dot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/scanner"
)

// UnmanageOptions configures unmanage behavior.
type UnmanageOptions struct {
	// Purge deletes the package directory instead of restoring files
	Purge bool
	// Restore moves adopted files back to target directory (default for adopted packages)
	Restore bool
	// Cleanup removes orphaned manifest entries (packages with no links or missing directories)
	Cleanup bool
}

// DefaultUnmanageOptions returns default unmanage options.
func DefaultUnmanageOptions() UnmanageOptions {
	return UnmanageOptions{
		Purge:   false,
		Restore: true,  // Restore adopted packages by default
		Cleanup: false, // Manual opt-in for cleanup
	}
}

// UnmanageService handles package removal (unmanage operations).
type UnmanageService struct {
	fs          FS
	logger      Logger
	executor    *executor.Executor
	manifestSvc *ManifestService
	packageDir  string
	targetDir   string
	dryRun      bool
}

// newUnmanageService creates a new UnmanageService instance.
func newUnmanageService(
	fs FS,
	logger Logger,
	exec *executor.Executor,
	manifestSvc *ManifestService,
	packageDir string,
	targetDir string,
	dryRun bool,
) *UnmanageService {
	return &UnmanageService{
		fs:          fs,
		logger:      logger,
		executor:    exec,
		manifestSvc: manifestSvc,
		packageDir:  packageDir,
		targetDir:   targetDir,
		dryRun:      dryRun,
	}
}

// Unmanage removes the specified packages by deleting symlinks.
// Uses default options (restore adopted packages, don't purge).
func (s *UnmanageService) Unmanage(ctx context.Context, packages ...string) error {
	return s.UnmanageWithOptions(ctx, DefaultUnmanageOptions(), packages...)
}

// UnmanageWithOptions removes packages with specified options.
func (s *UnmanageService) UnmanageWithOptions(ctx context.Context, opts UnmanageOptions, packages ...string) error {
	s.logger.Info(ctx, "unmanaging_packages", "count", len(packages), "packages", packages)

	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest to check package sources
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		if isManifestNotFoundError(err) {
			s.logger.Info(ctx, "no_manifest_nothing_to_unmanage")
			return nil
		}
		return err
	}
	m := manifestResult.Unwrap()

	// Plan unmanage and restoration operations
	s.logger.Debug(ctx, "planning_unmanage", "packages", packages)
	plan, err := s.planUnmanageWithOptions(ctx, m, packages, opts)
	if err != nil {
		s.logger.Error(ctx, "plan_failed", "error", err)
		return err
	}

	// In cleanup mode, empty operations are expected for orphaned packages
	// Skip early return to allow manifest cleanup
	if len(plan.Operations) == 0 && !opts.Cleanup {
		s.logger.Info(ctx, "nothing_to_unmanage", "packages", packages)
		return nil
	}

	// Execute operations if any exist
	if len(plan.Operations) > 0 {
		s.logger.Info(ctx, "plan_created", "operations", len(plan.Operations))

		if s.dryRun {
			s.logger.Info(ctx, "dry_run_plan", "operations", len(plan.Operations))
			return nil
		}

		s.logger.Debug(ctx, "executing_plan", "operation_count", len(plan.Operations))
		result := s.executor.Execute(ctx, plan)
		if !result.IsOk() {
			s.logger.Error(ctx, "execution_error", "error", result.UnwrapErr())
			return result.UnwrapErr()
		}
		execResult := result.Unwrap()
		if !execResult.Success() {
			s.logger.Error(ctx, "execution_failed", "failed_count", len(execResult.Failed))
			return ErrMultiple{Errors: execResult.Errors}
		}

		s.logger.Info(ctx, "execution_successful", "operations", len(execResult.Executed))
	}

	// Update manifest to remove packages
	// In cleanup mode, only remove packages that are actually orphaned
	packagesToRemove := packages
	if opts.Cleanup {
		packagesToRemove = s.filterOrphanedPackages(ctx, m, packages)
		s.logger.Debug(ctx, "cleanup_mode_filtered", "total", len(packages), "orphaned", len(packagesToRemove))
	}

	s.logger.Debug(ctx, "removing_packages_from_manifest", "packages", packagesToRemove)

	for _, pkg := range packagesToRemove {
		if err := s.manifestSvc.RemovePackage(ctx, targetPath, pkg); err != nil {
			s.logger.Warn(ctx, "failed_to_update_manifest", "package", pkg, "error", err)
			return err
		}
		s.logger.Debug(ctx, "package_removed_from_manifest", "package", pkg)
	}

	s.logger.Debug(ctx, "manifest_updated")
	return nil
}

// UnmanageAll removes all installed packages with specified options.
// Returns the count of packages unmanaged.
func (s *UnmanageService) UnmanageAll(ctx context.Context, opts UnmanageOptions) (int, error) {
	s.logger.Info(ctx, "unmanaging_all_packages")

	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return 0, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest to get all packages
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		if isManifestNotFoundError(err) {
			s.logger.Info(ctx, "no_manifest_nothing_to_unmanage")
			return 0, nil
		}
		return 0, err
	}
	m := manifestResult.Unwrap()

	// Get all package names
	packages := make([]string, 0, len(m.Packages))
	for pkgName := range m.Packages {
		packages = append(packages, pkgName)
	}

	if len(packages) == 0 {
		s.logger.Info(ctx, "no_packages_to_unmanage")
		return 0, nil
	}

	// Use UnmanageWithOptions for the actual work
	err := s.UnmanageWithOptions(ctx, opts, packages...)
	if err != nil {
		return 0, err
	}

	return len(packages), nil
}

// filterOrphanedPackages returns only the packages that are orphaned.
func (s *UnmanageService) filterOrphanedPackages(ctx context.Context, m manifest.Manifest, packages []string) []string {
	orphaned := make([]string, 0, len(packages))
	for _, pkg := range packages {
		pkgInfo, exists := m.GetPackage(pkg)
		if !exists {
			continue
		}
		if s.isPackageOrphaned(ctx, pkg, pkgInfo) {
			orphaned = append(orphaned, pkg)
		}
	}
	return orphaned
}

// PlanUnmanage computes the execution plan for unmanaging packages.
func (s *UnmanageService) PlanUnmanage(ctx context.Context, packages ...string) (Plan, error) {
	s.logger.Debug(ctx, "plan_unmanage_started", "packages", packages)

	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	s.logger.Debug(ctx, "loading_manifest")
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		// Check if this is a "file not found" error
		if isManifestNotFoundError(err) {
			s.logger.Debug(ctx, "no_manifest_found_nothing_to_unmanage")
			return Plan{
				Operations: []Operation{},
				Metadata:   PlanMetadata{},
			}, nil
		}
		return Plan{}, err
	}

	m := manifestResult.Unwrap()
	return s.planUnmanageWithOptions(ctx, m, packages, DefaultUnmanageOptions())
}

// planUnmanageWithOptions creates an unmanage plan with restoration/purge/cleanup logic.
func (s *UnmanageService) planUnmanageWithOptions(ctx context.Context, m manifest.Manifest, packages []string, opts UnmanageOptions) (Plan, error) {
	s.logger.Debug(ctx, "manifest_loaded", "installed_packages", len(m.Packages))

	// Build operations for each package
	var operations []Operation
	for _, pkg := range packages {
		pkgInfo, exists := m.GetPackage(pkg)
		if !exists {
			s.logger.Warn(ctx, "package_not_installed", "package", pkg)
			continue
		}

		s.logger.Debug(ctx, "planning_package", "package", pkg, "source", pkgInfo.Source, "links", len(pkgInfo.Links))

		// Cleanup mode: skip operations for orphaned packages (manifest will be cleaned up regardless)
		if opts.Cleanup {
			isOrphaned := s.isPackageOrphaned(ctx, pkg, pkgInfo)
			if isOrphaned {
				s.logger.Info(ctx, "orphaned_package_will_be_cleaned", "package", pkg)
				// Skip filesystem operations for orphaned packages
				// Manifest cleanup happens after execution
				continue
			} else {
				s.logger.Info(ctx, "package_not_orphaned_skipping_cleanup", "package", pkg)
				// In cleanup mode, only process orphaned packages
				continue
			}
		}

		// Delete symlinks
		for _, link := range pkgInfo.Links {
			targetFilePath := s.targetDir + "/" + link
			targetPathResult := NewTargetPath(targetFilePath)
			if !targetPathResult.IsOk() {
				continue
			}
			id := OperationID(fmt.Sprintf("unmanage-link-%s", link))
			operations = append(operations, NewLinkDelete(id, targetPathResult.Unwrap()))
		}

		// Handle adopted packages
		if pkgInfo.Source == manifest.SourceAdopted && opts.Restore && !opts.Purge {
			// Restore files from package back to target
			s.logger.Debug(ctx, "adding_restore_operations", "package", pkg)
			restoreOps, err := s.createRestoreOperations(ctx, pkg, pkgInfo.Links)
			if err != nil {
				s.logger.Warn(ctx, "failed_to_create_restore_operations", "package", pkg, "error", err)
			} else {
				operations = append(operations, restoreOps...)
			}
		} else if opts.Purge {
			// Delete package directory recursively
			s.logger.Debug(ctx, "adding_purge_operations", "package", pkg)
			pkgPath := filepath.Join(s.packageDir, pkg)
			pkgPathResult := NewFilePath(pkgPath)
			if pkgPathResult.IsErr() {
				return Plan{}, fmt.Errorf("invalid package path %s: %w", pkgPath, pkgPathResult.UnwrapErr())
			}
			id := OperationID(fmt.Sprintf("unmanage-purge-%s", pkg))
			operations = append(operations, NewDirRemoveAll(id, pkgPathResult.Unwrap()))
		}
	}

	s.logger.Debug(ctx, "plan_unmanage_completed", "operations", len(operations))

	return Plan{
		Operations: operations,
		Metadata: PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(operations),
		},
	}, nil
}

// isPackageOrphaned checks if a package is orphaned (has no valid links or missing package directory).
func (s *UnmanageService) isPackageOrphaned(ctx context.Context, pkg string, pkgInfo manifest.PackageInfo) bool {
	// Check if package directory exists
	pkgPath := filepath.Join(s.packageDir, pkg)
	if !s.fs.Exists(ctx, pkgPath) {
		s.logger.Debug(ctx, "package_directory_missing", "package", pkg)
		return true
	}

	// Check if any symlinks actually exist
	hasValidLinks := false
	for _, link := range pkgInfo.Links {
		linkPath := filepath.Join(s.targetDir, link)
		if s.fs.Exists(ctx, linkPath) {
			hasValidLinks = true
			break
		}
	}

	if !hasValidLinks && len(pkgInfo.Links) > 0 {
		s.logger.Debug(ctx, "no_valid_links_found", "package", pkg)
		return true
	}

	return false
}

// createRestoreOperations creates operations to restore adopted files back to target.
// Files are copied (not moved) so they remain in the package directory.
func (s *UnmanageService) createRestoreOperations(ctx context.Context, pkg string, links []string) ([]Operation, error) {
	operations := make([]Operation, 0, len(links))

	for _, link := range links {
		// The link in manifest is the target path (e.g., ".ssh")
		// With flat structure, package root contains the directory contents
		// So for link ".ssh", we copy from package root to target ".ssh"

		targetFilePath := filepath.Join(s.targetDir, link)
		pkgRootPath := filepath.Join(s.packageDir, pkg)

		// Check if the target link was a directory
		// If manifest has link ".ssh", check if package root is the directory
		if !s.fs.Exists(ctx, pkgRootPath) {
			s.logger.Warn(ctx, "package_directory_not_found", "package", pkg, "path", pkgRootPath)
			continue
		}

		// Check if package root is a directory with contents
		isDir, err := s.fs.IsDir(ctx, pkgRootPath)
		if err != nil {
			s.logger.Warn(ctx, "failed_to_check_package_type", "path", pkgRootPath, "error", err)
			continue
		}

		if isDir {
			// Handle directory package restoration (including corrupted structures)
			dirOps := s.createDirectoryRestoreOperations(ctx, pkg, link, pkgRootPath, targetFilePath)
			operations = append(operations, dirOps...)
		} else {
			// Single file adoption - old behavior
			// For files, the package structure is pkg/dot-filename
			linkDir := filepath.Dir(link)
			linkBase := filepath.Base(link)

			pkgFileName := scanner.UntranslateDotfile(linkBase)

			var pkgFilePath string
			if linkDir == "." {
				pkgFilePath = filepath.Join(s.packageDir, pkg, pkgFileName)
			} else {
				pkgFilePath = filepath.Join(s.packageDir, pkg, linkDir, pkgFileName)
			}

			if !s.fs.Exists(ctx, pkgFilePath) {
				s.logger.Warn(ctx, "package_file_not_found", "package", pkg, "file", pkgFilePath)
				continue
			}

			sourceResult := NewFilePath(pkgFilePath)
			if !sourceResult.IsOk() {
				continue
			}

			destResult := NewFilePath(targetFilePath)
			if !destResult.IsOk() {
				continue
			}

			id := OperationID(fmt.Sprintf("restore-copy-%s-%s", pkg, link))
			operations = append(operations, NewFileBackup(id, sourceResult.Unwrap(), destResult.Unwrap()))
		}
	}

	return operations, nil
}

// createDirectoryRestoreOperations handles restoration of directory packages.
// Detects and handles corrupted nested structures.
func (s *UnmanageService) createDirectoryRestoreOperations(ctx context.Context, pkg, link, pkgRootPath, targetFilePath string) []Operation {
	// Improved detection: check what actually exists in package
	linkBase := filepath.Base(link)
	translatedName := scanner.UntranslateDotfile(linkBase)
	fileInPackage := filepath.Join(pkgRootPath, translatedName)

	// Check if fileInPackage exists AND is a file (not directory)
	if s.fs.Exists(ctx, fileInPackage) {
		fileInfo, err := s.fs.Stat(ctx, fileInPackage)
		if err == nil && !fileInfo.IsDir() {
			// Single file adoption - restore the file
			s.logger.Info(ctx, "restoring_adopted_file", "package", pkg, "file", translatedName)

			sourceResult := NewFilePath(fileInPackage)
			if !sourceResult.IsOk() {
				return nil
			}

			destResult := NewFilePath(targetFilePath)
			if !destResult.IsOk() {
				return nil
			}

			id := OperationID(fmt.Sprintf("restore-file-%s", link))
			return []Operation{NewFileBackup(id, sourceResult.Unwrap(), destResult.Unwrap())}
		}

		// If fileInPackage is a directory, it's likely corruption
		if err == nil && fileInfo.IsDir() {
			s.logger.Warn(ctx, "detected_nested_structure", "package", pkg, "nested_dir", translatedName)
			return s.createCorruptedStructureRepair(ctx, pkg, fileInPackage, targetFilePath)
		}
	}

	// Standard directory adoption - package root IS the adopted directory
	s.logger.Info(ctx, "restoring_adopted_directory", "package", pkg, "link", link)

	sourceResult := NewFilePath(pkgRootPath)
	if !sourceResult.IsOk() {
		return nil
	}

	destResult := NewFilePath(targetFilePath)
	if !destResult.IsOk() {
		return nil
	}

	id := OperationID(fmt.Sprintf("restore-dir-%s", pkg))
	return []Operation{NewDirCopy(id, sourceResult.Unwrap(), destResult.Unwrap())}
}

// createCorruptedStructureRepair attempts to repair a nested directory structure.
// This handles the case where a package has a corrupted nested structure like dot-ssh/dot-ssh/files.
func (s *UnmanageService) createCorruptedStructureRepair(ctx context.Context, pkg, nestedPath, targetPath string) []Operation {
	// Copy from the nested incorrect location
	sourceResult := NewFilePath(nestedPath)
	if !sourceResult.IsOk() {
		return nil
	}

	destResult := NewFilePath(targetPath)
	if !destResult.IsOk() {
		return nil
	}

	s.logger.Info(ctx, "repairing_corrupted_structure", "package", pkg, "nested", nestedPath)

	id := OperationID(fmt.Sprintf("repair-nested-%s", pkg))
	return []Operation{NewDirCopy(id, sourceResult.Unwrap(), destResult.Unwrap())}
}
