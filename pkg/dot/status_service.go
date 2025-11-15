package dot

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

// StatusService handles status and listing operations.
type StatusService struct {
	fs          FS
	logger      Logger
	manifestSvc *ManifestService
	targetDir   string
}

// newStatusService creates a new status service.
func newStatusService(fs FS, logger Logger, manifestSvc *ManifestService, targetDir string) *StatusService {
	return &StatusService{
		fs:          fs,
		logger:      logger,
		manifestSvc: manifestSvc,
		targetDir:   targetDir,
	}
}

// Status reports the current installation state for packages.
func (s *StatusService) Status(ctx context.Context, packages ...string) (Status, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Status{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		if isManifestNotFoundError(err) {
			// No manifest means nothing installed
			return Status{Packages: []PackageInfo{}}, nil
		}
		return Status{}, err
	}

	m := manifestResult.Unwrap()

	// Filter to requested packages if specified
	pkgInfos := make([]PackageInfo, 0)
	if len(packages) == 0 {
		// Return all packages
		for _, info := range m.Packages {
			isHealthy, issueType := s.checkPackageHealth(ctx, info.Name, info.Links, info.PackageDir)
			pkgInfos = append(pkgInfos, PackageInfo{
				Name:        info.Name,
				Source:      string(info.Source),
				InstalledAt: info.InstalledAt,
				LinkCount:   info.LinkCount,
				Links:       info.Links,
				TargetDir:   info.TargetDir,
				PackageDir:  info.PackageDir,
				IsHealthy:   isHealthy,
				IssueType:   issueType,
			})
		}
	} else {
		// Return only specified packages
		for _, pkg := range packages {
			if info, exists := m.GetPackage(pkg); exists {
				isHealthy, issueType := s.checkPackageHealth(ctx, info.Name, info.Links, info.PackageDir)
				pkgInfos = append(pkgInfos, PackageInfo{
					Name:        info.Name,
					Source:      string(info.Source),
					InstalledAt: info.InstalledAt,
					LinkCount:   info.LinkCount,
					Links:       info.Links,
					TargetDir:   info.TargetDir,
					PackageDir:  info.PackageDir,
					IsHealthy:   isHealthy,
					IssueType:   issueType,
				})
			}
		}
	}
	return Status{
		Packages: pkgInfos,
	}, nil
}

// List returns all installed packages from the manifest.
func (s *StatusService) List(ctx context.Context) ([]PackageInfo, error) {
	status, err := s.Status(ctx)
	if err != nil {
		return nil, err
	}
	return status.Packages, nil
}

// resolveLinkPath converts relative link paths to absolute.
func (s *StatusService) resolveLinkPath(linkPath string) string {
	if filepath.IsAbs(linkPath) {
		return linkPath
	}
	return filepath.Join(s.targetDir, linkPath)
}

// resolveTargetPath resolves symlink target to absolute path.
func resolveTargetPath(symlinkPath, target string) string {
	if filepath.IsAbs(target) {
		return target
	}
	return filepath.Join(filepath.Dir(symlinkPath), target)
}

// isInPackageDir checks if target path is within package directory.
func isInPackageDir(targetPath, packageDir string) bool {
	cleanTarget := filepath.Clean(targetPath)
	cleanPackageDir := filepath.Clean(packageDir)
	return strings.HasPrefix(cleanTarget, cleanPackageDir+string(filepath.Separator)) ||
		cleanTarget == cleanPackageDir
}

// checkPackageHealth validates all symlinks for a package.
// Returns healthy status and issue type if problems are found.
func (s *StatusService) checkPackageHealth(ctx context.Context, pkgName string, links []string, packageDir string) (bool, string) {
	brokenLinks := 0
	wrongTargets := 0
	missingLinks := 0

	for _, linkPath := range links {
		absLinkPath := s.resolveLinkPath(linkPath)

		// Check if symlink exists and is valid
		linkInfo, err := s.fs.Lstat(ctx, absLinkPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				missingLinks++
			} else {
				brokenLinks++
			}
			continue
		}

		// Verify it's a symlink
		if linkInfo.Mode()&fs.ModeSymlink == 0 {
			wrongTargets++
			continue
		}

		// Read and resolve target
		target, err := s.fs.ReadLink(ctx, absLinkPath)
		if err != nil {
			brokenLinks++
			continue
		}

		resolvedTarget := resolveTargetPath(absLinkPath, target)

		// Check if target exists
		if _, err := s.fs.Stat(ctx, resolvedTarget); err != nil {
			brokenLinks++
			continue
		}

		// Verify target is in package directory (only if package_dir is set)
		// Old adopted packages and some legacy packages don't have package_dir set
		if packageDir != "" && !isInPackageDir(resolvedTarget, packageDir) {
			wrongTargets++
		}
	}

	// Determine health status and issue type
	if brokenLinks+wrongTargets+missingLinks == 0 {
		return true, ""
	}

	// Return most specific issue type
	if brokenLinks > 0 {
		return false, "broken links"
	}
	if wrongTargets > 0 {
		return false, "wrong target"
	}
	if missingLinks > 0 {
		return false, "missing links"
	}

	return false, "unknown issue"
}
