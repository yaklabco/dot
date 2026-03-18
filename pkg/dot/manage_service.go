package dot

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/pipeline"
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
		return ErrConflict{
			Path:   plan.Metadata.Conflicts[0].Path,
			Reason: conflictMsg,
		}
	}

	// If plan is empty (no operations needed), validate manifest before returning.
	// A corrupt manifest could cause the pipeline to produce zero operations
	// (symlinks exist on disk but manifest is unreadable), masking data integrity issues.
	if len(plan.Operations) == 0 {
		if err := s.validateManifestReadable(ctx); err != nil {
			return err
		}

		// Reconciliation: if symlinks exist on disk but the package is not in the manifest
		// (e.g., after manifest loss), re-register the package.
		reconciled, err := s.reconcileManifest(ctx, packages, plan)
		if err != nil {
			return err
		}
		if reconciled {
			return nil
		}

		s.logger.Info(ctx, "no_operations_required", "packages", packages)
		return ErrNoChanges{Packages: packages}
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
		return fmt.Errorf("manifest update failed: %w", err)
	}
	return nil
}

// PlanManage computes the execution plan for managing packages without applying changes.
func (s *ManageService) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
	// Validate packages - filter out reserved names
	validPackages := make([]string, 0, len(packages))
	var reservedNames []string
	for _, pkg := range packages {
		// Check reserved name
		if isReservedPackageName(pkg) {
			s.logger.Warn(ctx, "skipping_reserved_package", "package", pkg)
			fmt.Fprintf(os.Stderr,
				"Warning: Package %q is reserved for dot's internal use. Skipping.\n", pkg)
			reservedNames = append(reservedNames, pkg)
			continue
		}
		validPackages = append(validPackages, pkg)
	}

	if len(validPackages) == 0 {
		if len(reservedNames) > 0 {
			if len(reservedNames) == 1 {
				return Plan{}, fmt.Errorf("package %q is reserved for dot's internal use", reservedNames[0])
			}
			return Plan{}, fmt.Errorf("all specified packages are reserved for dot's internal use")
		}
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
		return ErrNoChanges{Packages: packages}
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
			return fmt.Errorf("manifest update failed for %s: %w", pkg, err)
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

	// Load manifest - missing manifests return Ok(empty), so errors here
	// mean the file exists but is corrupt or unreadable.
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return Plan{}, fmt.Errorf("failed to load manifest: %w", manifestResult.UnwrapErr())
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

	// Remove existing symlinks before planning manage operations.
	// This prevents the scanner from skipping recreation of links that will be deleted.
	if err := s.removeSymlinksOnly(ctx, unmanagePlan.Operations); err != nil {
		return nil, nil, err
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

// planAdoptedPackageRemanage plans remanage for an adopted package.
// Instead of pointing to the package root directory (which breaks single-file
// packages), it deletes existing links and re-runs the normal manage pipeline
// to create correct file-level symlinks.
func (s *ManageService) planAdoptedPackageRemanage(ctx context.Context, pkg string, m manifest.Manifest) ([]Operation, map[string][]OperationID, error) {
	pkgInfo, exists := m.GetPackage(pkg)
	if !exists {
		return nil, nil, fmt.Errorf("package %s not found in manifest", pkg)
	}

	// Delete existing symlinks
	var ops []Operation
	var opIDs []OperationID

	for _, link := range pkgInfo.Links {
		targetPath := filepath.Join(s.targetDir, link)
		targetPathResult := NewTargetPath(targetPath)
		if targetPathResult.IsOk() {
			delID := OperationID(fmt.Sprintf("remanage-del-%s", link))
			ops = append(ops, NewLinkDelete(delID, targetPathResult.Unwrap()))
			opIDs = append(opIDs, delID)
		}
	}

	// Remove existing symlinks so the manage pipeline sees a clean target.
	if err := s.removeSymlinksOnly(ctx, ops); err != nil {
		return nil, nil, err
	}

	// Re-run the normal manage pipeline to create file-level symlinks
	managePlan, err := s.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	ops = append(ops, managePlan.Operations...)
	for _, op := range managePlan.Operations {
		opIDs = append(opIDs, op.ID())
	}

	packageOps := map[string][]OperationID{pkg: opIDs}
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

// verifyLinksExist checks if all links in the manifest still exist as symlinks in the filesystem.
// It uses IsSymlink rather than Stat to detect when a managed symlink has been replaced by a regular file.
func (s *ManageService) verifyLinksExist(ctx context.Context, pkg string, m *manifest.Manifest) (bool, error) {
	pkgInfo, exists := m.GetPackage(pkg)
	if !exists {
		return false, nil
	}

	// Check each link from the manifest is still a symlink
	for _, link := range pkgInfo.Links {
		linkPath := filepath.Join(s.targetDir, link)
		isLink, err := s.fs.IsSymlink(ctx, linkPath)
		if err != nil {
			// Link doesn't exist or can't be accessed
			return false, nil
		}
		if !isLink {
			// Path exists but is not a symlink (e.g., replaced by a regular file)
			return false, nil
		}
	}

	return true, nil
}

// validateManifestReadable checks that the manifest can be loaded without errors.
// Returns nil if the manifest is valid or doesn't exist; returns an error if corrupt.
func (s *ManageService) validateManifestReadable(ctx context.Context) error {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return nil // Can't validate, don't block
	}
	manifestResult := s.manifestSvc.Load(ctx, targetPathResult.Unwrap())
	if !manifestResult.IsOk() {
		return fmt.Errorf("manifest is corrupt: %w", manifestResult.UnwrapErr())
	}
	return nil
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

// reconcileManifest checks if packages exist on disk as symlinks but are missing
// from the manifest (e.g., after manifest loss). If so, it scans the target dir
// for symlinks pointing into the package dir and registers them.
// Returns true if any packages were reconciled.
func (s *ManageService) reconcileManifest(ctx context.Context, packages []string, _ Plan) (bool, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return false, nil
	}
	targetPath := targetPathResult.Unwrap()

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return false, nil
	}
	m := manifestResult.Unwrap()

	reconciled := false
	for _, pkg := range packages {
		if _, exists := m.GetPackage(pkg); exists {
			continue
		}
		// Package not in manifest — scan target for symlinks pointing into this package
		pkgDir := filepath.Join(s.packageDir, pkg)
		if !s.fs.Exists(ctx, pkgDir) {
			continue
		}

		links := s.findSymlinksForPackage(ctx, pkgDir)
		if len(links) == 0 {
			continue
		}

		m.AddPackage(manifest.PackageInfo{
			Name:      pkg,
			LinkCount: len(links),
			Links:     links,
			Source:    manifest.SourceManaged,
			TargetDir: s.targetDir,
			PackageDir: pkgDir,
		})
		s.logger.Info(ctx, "reconciled_package_in_manifest", "package", pkg, "links", len(links))
		reconciled = true
	}

	if reconciled {
		if err := s.manifestSvc.Save(ctx, targetPath, m); err != nil {
			return false, fmt.Errorf("save reconciled manifest: %w", err)
		}
	}
	return reconciled, nil
}

// findSymlinksForPackage scans the target directory for symlinks that point
// into the given package directory, returning their relative paths.
func (s *ManageService) findSymlinksForPackage(ctx context.Context, pkgDir string) []string {
	var links []string
	entries, err := s.fs.ReadDir(ctx, s.targetDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		entryPath := filepath.Join(s.targetDir, entry.Name())
		s.findSymlinksRecursive(ctx, entryPath, pkgDir, "", &links)
	}
	return links
}

// findSymlinksRecursive recursively finds symlinks pointing into pkgDir.
func (s *ManageService) findSymlinksRecursive(ctx context.Context, path, pkgDir, relPrefix string, links *[]string) {
	isLink, err := s.fs.IsSymlink(ctx, path)
	if err != nil {
		return
	}

	name := filepath.Base(path)
	rel := name
	if relPrefix != "" {
		rel = filepath.Join(relPrefix, name)
	}

	if isLink {
		target, err := s.fs.ReadLink(ctx, path)
		if err != nil {
			return
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		target = filepath.Clean(target)
		if strings.HasPrefix(target, filepath.Clean(pkgDir)+string(os.PathSeparator)) || target == filepath.Clean(pkgDir) {
			*links = append(*links, rel)
		}
		return
	}

	// If it's a directory, recurse into it
	isDir, err := s.fs.IsDir(ctx, path)
	if err != nil || !isDir {
		return
	}
	entries, err := s.fs.ReadDir(ctx, path)
	if err != nil {
		return
	}
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		s.findSymlinksRecursive(ctx, childPath, pkgDir, rel, links)
	}
}

// removeSymlinksOnly removes symlink targets from LinkDelete operations.
// If any target is a regular file (not a symlink), it returns ErrConflict
// to prevent data loss. Missing targets are silently skipped.
func (s *ManageService) removeSymlinksOnly(ctx context.Context, ops []Operation) error {
	for _, op := range ops {
		linkDel, ok := op.(LinkDelete)
		if !ok {
			continue
		}
		target := linkDel.Target.String()
		isLink, err := s.fs.IsSymlink(ctx, target)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("checking symlink status of %s: %w", target, err)
		}
		if !isLink {
			return ErrConflict{
				Path:   target,
				Reason: fmt.Sprintf("expected symlink at %s but found a regular file; remove or back up the file before remanaging", target),
			}
		}
		_ = s.fs.Remove(ctx, target)
	}
	return nil
}
