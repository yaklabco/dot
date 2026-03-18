package dot

import (
	"context"
	"path/filepath"
	"time"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/manifest"
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
		newLinks := s.extractLinksFromOperations(ops, targetPath.String())
		deletedLinks := s.extractDeletedLinksFromOperations(ops, targetPath.String())
		backups := s.extractBackupsFromOperations(ops)

		// Merge with existing links: start from existing, remove deleted, add new
		links := s.mergeLinks(m, pkg, newLinks, deletedLinks)

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
	return s.RemovePackages(ctx, targetPath, []string{pkg})
}

// RemovePackages removes multiple packages from the manifest in a single load-save cycle.
func (s *ManifestService) RemovePackages(ctx context.Context, targetPath TargetPath, pkgs []string) error {
	manifestResult := s.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return manifestResult.UnwrapErr()
	}

	m := manifestResult.Unwrap()
	for _, pkg := range pkgs {
		m.RemovePackage(pkg)
	}

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

// extractDeletedLinksFromOperations extracts link paths from LinkDelete operations.
func (s *ManifestService) extractDeletedLinksFromOperations(ops []Operation, targetDir string) []string {
	var links []string
	for _, op := range ops {
		if linkOp, ok := op.(LinkDelete); ok {
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

// mergeLinks merges existing manifest links with plan deltas.
// It starts from existing links, removes deleted ones, and adds new ones.
func (s *ManifestService) mergeLinks(m manifest.Manifest, pkg string, newLinks, deletedLinks []string) []string {
	existing, hasExisting := m.GetPackage(pkg)
	if !hasExisting {
		// No existing entry — just use new links
		return newLinks
	}

	// Build set of deleted links
	deletedSet := make(map[string]struct{}, len(deletedLinks))
	for _, l := range deletedLinks {
		deletedSet[l] = struct{}{}
	}

	// Build set of new links for dedup
	newSet := make(map[string]struct{}, len(newLinks))
	for _, l := range newLinks {
		newSet[l] = struct{}{}
	}

	// Start with existing links, removing deleted ones
	merged := make([]string, 0, len(existing.Links)+len(newLinks))
	for _, l := range existing.Links {
		if _, isDeleted := deletedSet[l]; isDeleted {
			continue
		}
		if _, isNew := newSet[l]; isNew {
			continue // Will be added from newLinks
		}
		merged = append(merged, l)
	}

	// Add new links
	merged = append(merged, newLinks...)
	return merged
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
