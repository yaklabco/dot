package dot

import (
	"context"
)

// StatusService handles status and listing operations.
type StatusService struct {
	manifestSvc *ManifestService
	targetDir   string
}

// newStatusService creates a new status service.
func newStatusService(manifestSvc *ManifestService, targetDir string) *StatusService {
	return &StatusService{
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
			pkgInfos = append(pkgInfos, PackageInfo{
				Name:        info.Name,
				Source:      string(info.Source),
				InstalledAt: info.InstalledAt,
				LinkCount:   info.LinkCount,
				Links:       info.Links,
				TargetDir:   info.TargetDir,
				PackageDir:  info.PackageDir,
			})
		}
	} else {
		// Return only specified packages
		for _, pkg := range packages {
			if info, exists := m.GetPackage(pkg); exists {
				pkgInfos = append(pkgInfos, PackageInfo{
					Name:        info.Name,
					Source:      string(info.Source),
					InstalledAt: info.InstalledAt,
					LinkCount:   info.LinkCount,
					Links:       info.Links,
					TargetDir:   info.TargetDir,
					PackageDir:  info.PackageDir,
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
