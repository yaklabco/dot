package dot

import (
	"context"
	"path/filepath"
	"time"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/manifest"
)

// ManifestService manages manifest operations.
type ManifestService struct {
	fs     FS
	logger Logger
	store  manifest.ManifestStore
}

// newManifestService creates a new manifest service.
func newManifestService(fs FS, logger Logger, store manifest.ManifestStore) *ManifestService {
	return &ManifestService{
		fs:     fs,
		logger: logger,
		store:  store,
	}
}

// Load loads the manifest from the target directory.
func (s *ManifestService) Load(ctx context.Context, targetPath TargetPath) domain.Result[manifest.Manifest] {
	return s.store.Load(ctx, targetPath)
}

// Save saves the manifest to the target directory.
func (s *ManifestService) Save(ctx context.Context, targetPath TargetPath, m manifest.Manifest) error {
	return s.store.Save(ctx, targetPath, m)
}

// Update updates the manifest with package information from a plan.
func (s *ManifestService) Update(ctx context.Context, targetPath TargetPath, packageDir string, packages []string, plan Plan) error {
	return s.UpdateWithSource(ctx, targetPath, packageDir, packages, plan, manifest.SourceManaged)
}

// UpdateWithSource updates the manifest with package information and source type.
func (s *ManifestService) UpdateWithSource(ctx context.Context, targetPath TargetPath, packageDir string, packages []string, plan Plan, source manifest.PackageSource) error {
	// Load existing manifest (Load returns new manifest for not found case)
	manifestResult := s.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// Propagate non-not-found errors (Load already handles not found by returning new manifest)
		return manifestResult.UnwrapErr()
	}
	m := manifestResult.Unwrap()

	// Update package entries
	hasher := manifest.NewContentHasher(s.fs)

	// If packages slice is empty, populate from plan
	packagesToUpdate := packages
	if len(packagesToUpdate) == 0 && plan.PackageOperations != nil {
		packagesToUpdate = plan.PackageNames()
	}

	for _, pkg := range packagesToUpdate {
		// Extract links and backups from package operations
		ops := plan.OperationsForPackage(pkg)
		links := s.extractLinksFromOperations(ops, targetPath.String())
		backups := s.extractBackupsFromOperations(ops)

		m.AddPackage(manifest.PackageInfo{
			Name:        pkg,
			InstalledAt: time.Now(),
			LinkCount:   len(links),
			Links:       links,
			Backups:     backups,
			Source:      source,
			TargetDir:   targetPath.String(),
			PackageDir:  filepath.Join(packageDir, pkg),
		})

		// Compute and store package hash
		pkgPathStr := filepath.Join(packageDir, pkg)
		pkgPathResult := NewPackagePath(pkgPathStr)
		if pkgPathResult.IsOk() {
			pkgPath := pkgPathResult.Unwrap()
			hash, err := hasher.HashPackage(ctx, pkgPath)
			if err != nil {
				s.logger.Warn(ctx, "failed_to_compute_hash", "package", pkg, "error", err)
			} else {
				m.SetHash(pkg, hash)
			}
		}
	}

	// Save manifest
	return s.Save(ctx, targetPath, m)
}

// RemovePackage removes a package from the manifest.
func (s *ManifestService) RemovePackage(ctx context.Context, targetPath TargetPath, pkg string) error {
	manifestResult := s.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return manifestResult.UnwrapErr()
	}

	m := manifestResult.Unwrap()
	m.RemovePackage(pkg)

	return s.Save(ctx, targetPath, m)
}

// extractLinksFromOperations extracts link paths from LinkCreate operations.
func (s *ManifestService) extractLinksFromOperations(ops []Operation, targetDir string) []string {
	links := make([]string, 0, len(ops))
	for _, op := range ops {
		if linkOp, ok := op.(LinkCreate); ok {
			targetPath := linkOp.Target.String()
			relPath, err := filepath.Rel(targetDir, targetPath)
			if err != nil {
				relPath = targetPath
			}
			links = append(links, relPath)
		}
	}
	return links
}

func (s *ManifestService) extractBackupsFromOperations(ops []Operation) map[string]string {
	backups := make(map[string]string)
	for _, op := range ops {
		if backupOp, ok := op.(FileBackup); ok {
			// Map source (original file location) to backup location
			backups[backupOp.Source.String()] = backupOp.Backup.String()
		}
	}
	return backups
}
