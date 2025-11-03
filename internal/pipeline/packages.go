package pipeline

import (
	"context"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/planner"
)

// ManagePipelineOpts contains options for the Manage pipeline
type ManagePipelineOpts struct {
	FS                 domain.FS
	IgnoreSet          *ignore.IgnoreSet
	Policies           planner.ResolutionPolicies
	BackupDir          string
	PackageNameMapping bool
}

// ManageInput contains the input for manage operations
type ManageInput struct {
	PackageDir domain.PackagePath
	TargetDir  domain.TargetPath
	Packages   []string
}

// ManagePipeline implements the complete manage workflow.
// It composes scanning, planning, resolution, and sorting stages.
type ManagePipeline struct {
	opts ManagePipelineOpts
}

// NewManagePipeline creates a new Manage pipeline with the given options.
func NewManagePipeline(opts ManagePipelineOpts) *ManagePipeline {
	return &ManagePipeline{
		opts: opts,
	}
}

// Execute runs the complete manage pipeline.
// It performs: scan packages -> compute desired state -> resolve conflicts -> sort operations
func (p *ManagePipeline) Execute(ctx context.Context, input ManageInput) domain.Result[domain.Plan] {
	// Stage 1: Scan packages
	scanInput := ScanInput{
		PackageDir: input.PackageDir,
		TargetDir:  input.TargetDir,
		Packages:   input.Packages,
		IgnoreSet:  p.opts.IgnoreSet,
		FS:         p.opts.FS,
	}

	scanResult := ScanStage()(ctx, scanInput)
	if scanResult.IsErr() {
		return domain.Err[domain.Plan](scanResult.UnwrapErr())
	}
	packages := scanResult.Unwrap()

	// Stage 2: Compute desired state
	planInput := PlanInput{
		Packages:           packages,
		TargetDir:          input.TargetDir,
		PackageNameMapping: p.opts.PackageNameMapping,
	}

	planResult := PlanStage()(ctx, planInput)
	if planResult.IsErr() {
		return domain.Err[domain.Plan](planResult.UnwrapErr())
	}
	desired := planResult.Unwrap()

	// Validate no self-management - check if any package attempts to manage dot's directories
	for _, pkg := range packages {
		// For simplicity, validate the entire desired state against this package name
		// This is conservative but prevents self-management issues
		if err := planner.ValidateNoSelfManagement(pkg.Name, desired); err != nil {
			// Return error plan - this should not be allowed
			return domain.Err[domain.Plan](err)
		}
	}

	// Stage 3: Resolve conflicts and generate operations
	resolveInput := ResolveInput{
		Desired:   desired,
		TargetDir: input.TargetDir,
		FS:        p.opts.FS,
		Policies:  p.opts.Policies,
		BackupDir: p.opts.BackupDir,
	}

	resolveResult := ResolveStage()(ctx, resolveInput)
	if resolveResult.IsErr() {
		return domain.Err[domain.Plan](resolveResult.UnwrapErr())
	}
	resolved := resolveResult.Unwrap()

	// Check for unresolved conflicts
	if resolved.HasConflicts() {
		// Return plan with conflicts for user to handle
		// The caller can inspect the conflicts in the metadata
		return domain.Ok(domain.Plan{
			Operations: resolved.Operations,
			Metadata: domain.PlanMetadata{
				PackageCount:   len(packages),
				OperationCount: len(resolved.Operations),
				LinkCount:      countOperationsByKind(resolved.Operations, domain.OpKindLinkCreate),
				DirCount:       countOperationsByKind(resolved.Operations, domain.OpKindDirCreate),
				Conflicts:      convertConflicts(resolved.Conflicts),
				Warnings:       convertWarnings(resolved.Warnings),
			},
		})
	}

	// Stage 4: Sort operations topologically
	sortInput := SortInput{
		Operations: resolved.Operations,
	}

	sortResult := SortStage()(ctx, sortInput)
	if sortResult.IsErr() {
		return domain.Err[domain.Plan](sortResult.UnwrapErr())
	}
	sorted := sortResult.Unwrap()

	// Build package-operation mapping by matching operations to package source paths
	packageOps := buildPackageOperationMapping(packages, sorted)

	// Build final plan with metadata including any warnings
	plan := domain.Plan{
		Operations: sorted,
		Metadata: domain.PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(sorted),
			LinkCount:      countOperationsByKind(sorted, domain.OpKindLinkCreate),
			DirCount:       countOperationsByKind(sorted, domain.OpKindDirCreate),
			Conflicts:      nil, // No conflicts in success path
			Warnings:       convertWarnings(resolved.Warnings),
		},
		PackageOperations: packageOps,
	}

	return domain.Ok(plan)
}

// countOperationsByKind counts operations of a specific kind
func countOperationsByKind(ops []domain.Operation, kind domain.OperationKind) int {
	count := 0
	for _, op := range ops {
		if op.Kind() == kind {
			count++
		}
	}
	return count
}

// buildPackageOperationMapping creates a mapping from package names to operation IDs
// by matching operation source paths to package paths.
func buildPackageOperationMapping(packages []domain.Package, operations []domain.Operation) map[string][]domain.OperationID {
	packageOps := make(map[string][]domain.OperationID)

	// Build a map of target paths to package names from LinkCreate operations
	targetToPackage := make(map[string]string)
	for _, pkg := range packages {
		pkgPath := pkg.Path.String()
		for _, op := range operations {
			if linkOp, ok := op.(domain.LinkCreate); ok {
				if isUnderPath(linkOp.Source.String(), pkgPath) {
					targetToPackage[linkOp.Target.String()] = pkg.Name
				}
			}
		}
	}

	// For each package, find operations that reference files from that package
	for _, pkg := range packages {
		pkgPath := pkg.Path.String()
		ops := make([]domain.OperationID, 0)

		for _, op := range operations {
			// Check if this operation's source is from this package
			if operationBelongsToPackage(op, pkg.Name, pkgPath, targetToPackage) {
				ops = append(ops, op.ID())
			}
		}

		if len(ops) > 0 {
			packageOps[pkg.Name] = ops
		}
	}

	return packageOps
}

// operationBelongsToPackage checks if an operation belongs to the given package.
// For FileBackup and FileDelete, it checks if they're preparing for a LinkCreate from this package.
func operationBelongsToPackage(op domain.Operation, pkgName string, pkgPath string, targetToPackage map[string]string) bool {
	switch o := op.(type) {
	case domain.LinkCreate:
		// LinkCreate source is the file in the package
		return isUnderPath(o.Source.String(), pkgPath)
	case domain.FileMove:
		// FileMove destination is the file in the package
		return isUnderPath(o.Dest.String(), pkgPath)
	case domain.FileBackup:
		// FileBackup belongs to package that will create a link at this location
		// The Source of FileBackup is the original file being backed up
		targetPkgName, exists := targetToPackage[o.Source.String()]
		return exists && targetPkgName == pkgName
	case domain.FileDelete:
		// FileDelete belongs to package that will create a link at this location
		// The Path of FileDelete is the file being deleted
		targetPkgName, exists := targetToPackage[o.Path.String()]
		return exists && targetPkgName == pkgName
	default:
		// Other operations (DirCreate, LinkDelete, etc.) don't belong to a specific package
		return false
	}
}

// isUnderPath checks if path is under basePath.
func isUnderPath(path, basePath string) bool {
	// Clean both paths for consistent comparison
	cleanPath := filepath.Clean(path)
	cleanBase := filepath.Clean(basePath)

	// Check if path starts with basePath
	rel, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil {
		return false
	}

	// If relative path doesn't go up (..), it's under basePath
	return rel != "." && !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.'
}
