package pipeline

import (
	"context"
	"path/filepath"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/planner"
	"github.com/yaklabco/dot/internal/scanner"
)

// scanCurrentState scans only the specific paths relevant to the desired state.
// This is vastly more efficient than recursively scanning the entire target directory,
// especially when the target is a home directory with large subdirectories like node_modules.
func scanCurrentState(ctx context.Context, fs domain.FS, desired planner.DesiredState) planner.CurrentState {
	current := planner.CurrentState{
		Files: make(map[string]planner.FileInfo),
		Links: make(map[string]planner.LinkTarget),
		Dirs:  make(map[string]struct{}),
	}

	// Collect all paths we need to check
	pathsToCheck := make(map[string]struct{})

	// Add all desired link paths
	for path := range desired.Links {
		pathsToCheck[path] = struct{}{}
		// Also check parent directories
		addParentPaths(path, pathsToCheck)
	}

	// Add all desired directory paths
	for path := range desired.Dirs {
		pathsToCheck[path] = struct{}{}
		// Also check parent directories
		addParentPaths(path, pathsToCheck)
	}

	// Check each path
	for path := range pathsToCheck {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return current
		default:
		}

		// Check if path exists
		if !fs.Exists(ctx, path) {
			continue
		}

		// Check if it's a symlink
		if isLink, _ := fs.IsSymlink(ctx, path); isLink {
			if linkTarget, err := fs.ReadLink(ctx, path); err == nil {
				current.Links[path] = planner.LinkTarget{
					Target: linkTarget,
				}
			}
			continue
		}

		// Check if it's a directory
		if isDir, _ := fs.IsDir(ctx, path); isDir {
			current.Dirs[path] = struct{}{}
			continue
		}

		// It's a regular file
		if info, err := fs.Stat(ctx, path); err == nil {
			current.Files[path] = planner.FileInfo{
				Size: info.Size(),
			}
		}
	}

	return current
}

// addParentPaths adds all parent directory paths to the set
func addParentPaths(path string, paths map[string]struct{}) {
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" && dir != "" {
		if _, exists := paths[dir]; exists {
			break // Already added this and all its parents
		}
		paths[dir] = struct{}{}
		dir = filepath.Dir(dir)
	}
}

// ScanInput contains the input for scanning packages
type ScanInput struct {
	PackageDir domain.PackagePath
	TargetDir  domain.TargetPath
	Packages   []string
	IgnoreSet  *ignore.IgnoreSet
	ScanConfig scanner.ScanConfig
	FS         domain.FS
}

// ScanStage creates a pipeline stage that scans packages.
// Returns a slice of scanned packages with their file trees.
func ScanStage() Pipeline[ScanInput, []domain.Package] {
	return func(ctx context.Context, input ScanInput) domain.Result[[]domain.Package] {
		// Early cancellation check
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Package](ctx.Err())
		default:
		}

		packages := make([]domain.Package, 0, len(input.Packages))

		for _, pkgName := range input.Packages {
			// Check for cancellation before processing each package
			select {
			case <-ctx.Done():
				return domain.Err[[]domain.Package](ctx.Err())
			default:
			}

			// Create package path by joining package dir with package name
			pkgPathStr := filepath.Join(input.PackageDir.String(), pkgName)
			pkgPathResult := domain.NewPackagePath(pkgPathStr)
			if pkgPathResult.IsErr() {
				return domain.Err[[]domain.Package](pkgPathResult.UnwrapErr())
			}
			pkgPath := pkgPathResult.Unwrap()

			// Use ScanPackageWithConfig if any advanced features are enabled
			var pkgResult domain.Result[domain.Package]
			if input.ScanConfig.PerPackageIgnore || input.ScanConfig.MaxFileSize > 0 {
				pkgResult = scanner.ScanPackageWithConfig(ctx, input.FS, pkgPath, pkgName, input.IgnoreSet, input.ScanConfig)
			} else {
				// Use standard scan for backward compatibility
				pkgResult = scanner.ScanPackage(ctx, input.FS, pkgPath, pkgName, input.IgnoreSet)
			}

			if pkgResult.IsErr() {
				return domain.Err[[]domain.Package](pkgResult.UnwrapErr())
			}

			packages = append(packages, pkgResult.Unwrap())
		}

		return domain.Ok(packages)
	}
}

// PlanInput contains the input for planning operations
type PlanInput struct {
	Packages           []domain.Package
	TargetDir          domain.TargetPath
	PackageNameMapping bool
}

// PlanStage creates a pipeline stage that computes desired state.
// Takes scanned packages and computes what links should exist.
func PlanStage() Pipeline[PlanInput, planner.DesiredState] {
	return func(ctx context.Context, input PlanInput) domain.Result[planner.DesiredState] {
		// Early cancellation check before potentially long-running planning
		select {
		case <-ctx.Done():
			return domain.Err[planner.DesiredState](ctx.Err())
		default:
		}

		return planner.ComputeDesiredState(input.Packages, input.TargetDir, input.PackageNameMapping)
	}
}

// ResolveInput contains the input for conflict resolution
type ResolveInput struct {
	Desired   planner.DesiredState
	TargetDir domain.TargetPath
	FS        domain.FS
	Policies  planner.ResolutionPolicies
	BackupDir string
}

// ResolveStage creates a pipeline stage that resolves conflicts.
// Takes desired state and current filesystem state to generate operations.
func ResolveStage() Pipeline[ResolveInput, planner.ResolveResult] {
	return func(ctx context.Context, input ResolveInput) domain.Result[planner.ResolveResult] {
		// Early cancellation check
		select {
		case <-ctx.Done():
			return domain.Err[planner.ResolveResult](ctx.Err())
		default:
		}

		// Convert desired state to operations
		operations := planner.ComputeOperationsFromDesiredState(input.Desired)

		// Check for cancellation before building current state
		select {
		case <-ctx.Done():
			return domain.Err[planner.ResolveResult](ctx.Err())
		default:
		}

		// Scan only the specific paths we care about for conflict detection
		// This is much more efficient than scanning the entire target directory
		current := scanCurrentState(ctx, input.FS, input.Desired)

		// Check for cancellation before potentially long-running conflict resolution
		select {
		case <-ctx.Done():
			return domain.Err[planner.ResolveResult](ctx.Err())
		default:
		}

		// Resolve conflicts
		result := planner.Resolve(operations, current, input.Policies, input.BackupDir)
		return domain.Ok(result)
	}
}

// SortInput contains the input for topological sorting
type SortInput struct {
	Operations []domain.Operation
}

// SortStage creates a pipeline stage that sorts operations.
// Takes operations and returns them in dependency order.
func SortStage() Pipeline[SortInput, []domain.Operation] {
	return func(ctx context.Context, input SortInput) domain.Result[[]domain.Operation] {
		// Early cancellation check
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Operation](ctx.Err())
		default:
		}

		if len(input.Operations) == 0 {
			return domain.Ok([]domain.Operation{})
		}

		// Check for cancellation before building dependency graph
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Operation](ctx.Err())
		default:
		}

		graph := planner.BuildGraph(input.Operations)

		// Check for cancellation before potentially long-running topological sort
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Operation](ctx.Err())
		default:
		}

		sorted, err := graph.TopologicalSort()
		if err != nil {
			return domain.Err[[]domain.Operation](err)
		}
		return domain.Ok(sorted)
	}
}
