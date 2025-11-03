package dot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
)

// ManageService handles package installation (manage and remanage operations).
type ManageService struct {
	fs          FS
	logger      Logger
	managePipe  *pipeline.ManagePipeline
	executor    *executor.Executor
	manifestSvc *ManifestService
	unmanageSvc *UnmanageService
	packageDir  string
	targetDir   string
	dryRun      bool
}

// newManageService creates a new manage service.
func newManageService(
	fs FS,
	logger Logger,
	managePipe *pipeline.ManagePipeline,
	exec *executor.Executor,
	manifestSvc *ManifestService,
	unmanageSvc *UnmanageService,
	packageDir string,
	targetDir string,
	dryRun bool,
) *ManageService {
	return &ManageService{
		fs:          fs,
		logger:      logger,
		managePipe:  managePipe,
		executor:    exec,
		manifestSvc: manifestSvc,
		unmanageSvc: unmanageSvc,
		packageDir:  packageDir,
		targetDir:   targetDir,
		dryRun:      dryRun,
	}
}

// Manage installs the specified packages by creating symlinks.
func (s *ManageService) Manage(ctx context.Context, packages ...string) error {
	// Validate package names
	for _, pkg := range packages {
		if pkg == "" {
			return fmt.Errorf("package name cannot be empty")
		}
	}

	plan, err := s.PlanManage(ctx, packages...)
	if err != nil {
		return err
	}

	// Check for conflicts before execution
	if len(plan.Metadata.Conflicts) > 0 {
		// Build error message with conflict details
		conflictMsg := fmt.Sprintf("cannot manage packages: %d conflict(s) detected", len(plan.Metadata.Conflicts))
		for i, conflict := range plan.Metadata.Conflicts {
			if i < 3 { // Show first 3 conflicts
				conflictMsg += fmt.Sprintf("\n  - %s at %s: %s", conflict.Type, conflict.Path, conflict.Details)
			}
		}
		if len(plan.Metadata.Conflicts) > 3 {
			conflictMsg += fmt.Sprintf("\n  ... and %d more", len(plan.Metadata.Conflicts)-3)
		}
		return errors.New(conflictMsg)
	}

	// If plan is empty (no operations needed), consider it success
	if len(plan.Operations) == 0 {
		s.logger.Info(ctx, "no_operations_required", "packages", packages)
		return nil
	}

	if s.dryRun {
		return nil
	}
	result := s.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}
	execResult := result.Unwrap()
	if !execResult.Success() {
		return fmt.Errorf("execution failed: %d operations failed", len(execResult.Failed))
	}
	// Update manifest
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	if err := s.manifestSvc.Update(ctx, targetPathResult.Unwrap(), s.packageDir, packages, plan); err != nil {
		s.logger.Warn(ctx, "manifest_update_failed", "error", err)
	}
	return nil
}

// PlanManage computes the execution plan for managing packages without applying changes.
func (s *ManageService) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
	// Validate packages - filter out reserved names
	validPackages := make([]string, 0, len(packages))
	for _, pkg := range packages {
		// Check reserved name
		if isReservedPackageName(pkg) {
			s.logger.Warn(ctx, "skipping_reserved_package", "package", pkg)
			fmt.Fprintf(os.Stderr,
				"Warning: Package %q is reserved for dot's internal use. Skipping.\n", pkg)
			continue
		}
		validPackages = append(validPackages, pkg)
	}

	if len(validPackages) == 0 {
		return Plan{}, fmt.Errorf("no valid packages to manage")
	}

	packages = validPackages

	packagePathResult := NewPackagePath(s.packageDir)
	if !packagePathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid package directory: %w", packagePathResult.UnwrapErr())
	}
	packagePath := packagePathResult.Unwrap()

	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid target directory: %w", targetPathResult.UnwrapErr())
	}
	targetPath := targetPathResult.Unwrap()

	input := pipeline.ManageInput{
		PackageDir: packagePath,
		TargetDir:  targetPath,
		Packages:   packages,
	}
	planResult := s.managePipe.Execute(ctx, input)
	if !planResult.IsOk() {
		return Plan{}, planResult.UnwrapErr()
	}
	return planResult.Unwrap(), nil
}

// Remanage reinstalls packages using incremental hash-based change detection.
func (s *ManageService) Remanage(ctx context.Context, packages ...string) error {
	plan, err := s.PlanRemanage(ctx, packages...)
	if err != nil {
		return err
	}
	if len(plan.Operations) == 0 {
		s.logger.Info(ctx, "no_changes_detected", "packages", packages)
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
	// Update manifest, preserving source type for each package
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}

	// Load manifest to check source types
	manifestResult := s.manifestSvc.Load(ctx, targetPathResult.Unwrap())
	for _, pkg := range packages {
		source := manifest.SourceManaged // Default
		if manifestResult.IsOk() {
			m := manifestResult.Unwrap()
			if pkgInfo, exists := m.GetPackage(pkg); exists {
				source = pkgInfo.Source
			}
		}
		if err := s.manifestSvc.UpdateWithSource(ctx, targetPathResult.Unwrap(), s.packageDir, []string{pkg}, plan, source); err != nil {
			s.logger.Warn(ctx, "manifest_update_failed", "package", pkg, "error", err)
		}
	}
	return nil
}

// PlanRemanage computes incremental execution plan using hash-based change detection.
func (s *ManageService) PlanRemanage(ctx context.Context, packages ...string) (Plan, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid target directory: %w", targetPathResult.UnwrapErr())
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// No manifest - fall back to full manage
		return s.PlanManage(ctx, packages...)
	}

	m := manifestResult.Unwrap()
	hasher := manifest.NewContentHasher(s.fs)
	allOperations := make([]Operation, 0)
	packageOps := make(map[string][]OperationID)

	for _, pkg := range packages {
		ops, pkgOpsMap, err := s.planSinglePackageRemanage(ctx, pkg, &m, hasher)
		if err != nil {
			return Plan{}, err
		}
		allOperations = append(allOperations, ops...)
		for k, v := range pkgOpsMap {
			packageOps[k] = v
		}
	}

	return Plan{
		Operations: allOperations,
		Metadata: PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(allOperations),
		},
		PackageOperations: packageOps,
	}, nil
}

// planSinglePackageRemanage plans remanage for a single package using hash comparison.
func (s *ManageService) planSinglePackageRemanage(
	ctx context.Context,
	pkg string,
	m *manifest.Manifest,
	hasher *manifest.ContentHasher,
) ([]Operation, map[string][]OperationID, error) {
	_, exists := m.GetPackage(pkg)
	if !exists {
		return s.planNewPackageInstall(ctx, pkg)
	}

	pkgPath, err := s.getPackagePath(pkg)
	if err != nil {
		return nil, nil, err
	}
	currentHash, err := hasher.HashPackage(ctx, pkgPath)
	if err != nil {
		s.logger.Warn(ctx, "hash_computation_failed", "package", pkg, "error", err)
		return s.planFullRemanage(ctx, pkg)
	}

	storedHash, hasHash := m.GetHash(pkg)
	if !hasHash || storedHash != currentHash {
		return s.planFullRemanage(ctx, pkg)
	}

	// Check if all links still exist - recreate if any are missing
	if linksExist, err := s.verifyLinksExist(ctx, pkg, m); err != nil || !linksExist {
		if err != nil {
			s.logger.Warn(ctx, "link_verification_failed", "package", pkg, "error", err)
		} else {
			s.logger.Info(ctx, "missing_links_detected", "package", pkg)
		}
		return s.planFullRemanage(ctx, pkg)
	}

	s.logger.Info(ctx, "package_unchanged", "package", pkg)
	return []Operation{}, map[string][]OperationID{}, nil
}

// planNewPackageInstall plans installation of a package not yet in manifest.
func (s *ManageService) planNewPackageInstall(ctx context.Context, pkg string) ([]Operation, map[string][]OperationID, error) {
	pkgPlan, err := s.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}
	packageOps := make(map[string][]OperationID)
	if pkgPlan.PackageOperations != nil {
		if pkgOps, hasPkg := pkgPlan.PackageOperations[pkg]; hasPkg {
			packageOps[pkg] = pkgOps
		}
	}
	return pkgPlan.Operations, packageOps, nil
}

// planFullRemanage plans full unmanage + manage for a package.
func (s *ManageService) planFullRemanage(ctx context.Context, pkg string) ([]Operation, map[string][]OperationID, error) {
	// Check if this is an adopted package
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return nil, nil, targetPathResult.UnwrapErr()
	}

	manifestResult := s.manifestSvc.Load(ctx, targetPathResult.Unwrap())
	var isAdopted bool
	if manifestResult.IsOk() {
		m := manifestResult.Unwrap()
		if pkgInfo, exists := m.GetPackage(pkg); exists && pkgInfo.Source == manifest.SourceAdopted {
			isAdopted = true
		}
	}

	// For adopted packages, we need to recreate the original adoption pattern
	// (single symlink from target to package root), not re-scan as a normal package
	if isAdopted {
		return s.planAdoptedPackageRemanage(ctx, pkg, manifestResult.Unwrap())
	}

	// Get unmanage operations first
	unmanagePlan, err := s.unmanageSvc.PlanUnmanage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	// Remove existing symlinks before planning manage operations
	// This prevents the scanner from skipping recreation of links that will be deleted
	for _, op := range unmanagePlan.Operations {
		if linkDel, ok := op.(LinkDelete); ok {
			_ = s.fs.Remove(ctx, linkDel.Target.String())
		}
	}

	// Get manage operations (scanner will now not see the old symlinks)
	managePlan, err := s.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	// Concatenate operations (unmanage first, then manage)
	ops := make([]Operation, 0, len(unmanagePlan.Operations)+len(managePlan.Operations))
	ops = append(ops, unmanagePlan.Operations...)
	ops = append(ops, managePlan.Operations...)

	// Merge package operations
	packageOps := make(map[string][]OperationID)
	unmanageOps := unmanagePlan.PackageOperations[pkg]
	manageOps := managePlan.PackageOperations[pkg]
	mergedOps := make([]OperationID, 0, len(unmanageOps)+len(manageOps))
	mergedOps = append(mergedOps, unmanageOps...)
	mergedOps = append(mergedOps, manageOps...)
	packageOps[pkg] = mergedOps

	return ops, packageOps, nil
}

// planAdoptedPackageRemanage plans remanage for an adopted package by recreating the original symlink.
func (s *ManageService) planAdoptedPackageRemanage(ctx context.Context, pkg string, m manifest.Manifest) ([]Operation, map[string][]OperationID, error) {
	pkgInfo, exists := m.GetPackage(pkg)
	if !exists {
		return nil, nil, fmt.Errorf("package %s not found in manifest", pkg)
	}

	// Adopted packages should have exactly one link (the original target path)
	if len(pkgInfo.Links) != 1 {
		s.logger.Warn(ctx, "adopted_package_unexpected_links", "package", pkg, "link_count", len(pkgInfo.Links))
	}

	// Create operations to recreate the symlink
	var ops []Operation
	packageOps := make(map[string][]OperationID)
	var opIDs []OperationID

	for _, link := range pkgInfo.Links {
		// Delete existing symlink if it exists
		targetPath := filepath.Join(s.targetDir, link)
		targetPathResult := NewTargetPath(targetPath)
		if targetPathResult.IsOk() {
			delID := OperationID(fmt.Sprintf("remanage-del-%s", link))
			ops = append(ops, NewLinkDelete(delID, targetPathResult.Unwrap()))
			opIDs = append(opIDs, delID)
		}

		// Recreate symlink from target to package root
		pkgPath := filepath.Join(s.packageDir, pkg)
		sourcePathResult := NewFilePath(pkgPath)
		if !sourcePathResult.IsOk() {
			return nil, nil, fmt.Errorf("invalid package path: %w", sourcePathResult.UnwrapErr())
		}

		if targetPathResult.IsOk() {
			linkID := OperationID(fmt.Sprintf("remanage-link-%s", link))
			ops = append(ops, NewLinkCreate(linkID, sourcePathResult.Unwrap(), targetPathResult.Unwrap()))
			opIDs = append(opIDs, linkID)
		}
	}

	packageOps[pkg] = opIDs
	return ops, packageOps, nil
}

// getPackagePath constructs and validates package path.
func (s *ManageService) getPackagePath(pkg string) (PackagePath, error) {
	pkgPathStr := filepath.Join(s.packageDir, pkg)
	pkgPathResult := NewPackagePath(pkgPathStr)
	if !pkgPathResult.IsOk() {
		return PackagePath{}, pkgPathResult.UnwrapErr()
	}
	return pkgPathResult.Unwrap(), nil
}

// verifyLinksExist checks if all links in the manifest still exist in the filesystem.
func (s *ManageService) verifyLinksExist(ctx context.Context, pkg string, m *manifest.Manifest) (bool, error) {
	pkgInfo, exists := m.GetPackage(pkg)
	if !exists {
		return false, nil
	}

	// Check each link from the manifest
	for _, link := range pkgInfo.Links {
		linkPath := filepath.Join(s.targetDir, link)
		_, err := s.fs.Stat(ctx, linkPath)
		if err != nil {
			// Link doesn't exist or can't be accessed
			return false, nil
		}
	}

	return true, nil
}

// isReservedPackageName checks if the given package name is reserved for dot's internal use.
func isReservedPackageName(name string) bool {
	reserved := []string{
		"dot",
		".dot",
		"dot-config",
	}

	nameLower := strings.ToLower(name)
	for _, r := range reserved {
		if nameLower == r {
			return true
		}
	}

	return false
}
