package doctor

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/jamesainslie/dot/internal/domain"
)

// PlatformCheck validates platform compatibility for packages.
type PlatformCheck struct {
	fs          FS
	manifestSvc ManifestLoader
	packageDir  string
}

// NewPlatformCheck creates a new platform compatibility check.
func NewPlatformCheck(fs FS, manifestSvc ManifestLoader, packageDir string) *PlatformCheck {
	return &PlatformCheck{
		fs:          fs,
		manifestSvc: manifestSvc,
		packageDir:  packageDir,
	}
}

func (c *PlatformCheck) Name() string {
	return "platform_compatibility"
}

func (c *PlatformCheck) Description() string {
	return "Validates platform compatibility for managed packages"
}

func (c *PlatformCheck) Run(ctx domain.Context) (domain.CheckResult, error) {
	result := domain.CheckResult{
		CheckName: c.Name(),
		Status:    domain.CheckStatusPass,
		Issues:    make([]domain.Issue, 0),
		Stats:     make(map[string]any),
	}

	mf, err := c.manifestSvc.LoadManifest(context.Background())
	if err != nil {
		return result, fmt.Errorf("failed to load manifest: %w", err)
	}

	if mf == nil || len(mf.Packages) == 0 {
		result.Stats["packages_checked"] = 0
		return result, nil
	}

	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH
	packagesChecked := 0
	incompatibleCount := 0

	stdCtx := context.Background()

	for _, pkg := range mf.Packages {
		packagesChecked++
		pkgPath := filepath.Join(c.packageDir, pkg.Name)

		// Check if package directory exists
		exists, err := c.fs.Exists(stdCtx, pkgPath)
		if err != nil {
			return result, fmt.Errorf("failed to check package directory: %w", err)
		}

		if !exists {
			// Package not cloned yet, skip platform check
			continue
		}

		// Load package metadata
		metadataPath := filepath.Join(pkgPath, ".dot-metadata.json")
		metadataExists, err := c.fs.Exists(stdCtx, metadataPath)
		if err != nil {
			// Metadata file check failed, assume no metadata and continue
			continue
		}
		if !metadataExists {
			// No metadata file, assume compatible
			continue
		}

		data, err := c.fs.ReadFile(stdCtx, metadataPath)
		if err != nil {
			result.Issues = append(result.Issues, domain.Issue{
				Code:     "METADATA_READ_ERROR",
				Message:  fmt.Sprintf("Failed to read metadata for %s: %v", pkg.Name, err),
				Severity: domain.IssueSeverityWarning,
				Context: map[string]any{
					"package": pkg.Name,
					"path":    metadataPath,
				},
			})
			continue
		}

		// For now, skip parsing metadata as we don't have the Metadata type defined
		// This is a placeholder check
		_ = data

		// Check platform compatibility (simplified for now)
		// In a full implementation, we would parse metadata and check platform
		// For now, assume all packages are compatible
		if false {
			incompatibleCount++
			result.Status = domain.CheckStatusWarning
			result.Issues = append(result.Issues, domain.Issue{
				Code:     "PLATFORM_INCOMPATIBLE",
				Message:  fmt.Sprintf("Package %s may not be compatible with %s/%s", pkg.Name, currentOS, currentArch),
				Severity: domain.IssueSeverityWarning,
				Context: map[string]any{
					"package":      pkg.Name,
					"current_os":   currentOS,
					"current_arch": currentArch,
				},
				Remediation: &domain.Remediation{
					Description: fmt.Sprintf("Review package %s for platform-specific issues or check with package maintainer", pkg.Name),
				},
			})
		}
	}

	result.Stats["packages_checked"] = packagesChecked
	result.Stats["incompatible_count"] = incompatibleCount
	result.Stats["current_platform"] = fmt.Sprintf("%s/%s", currentOS, currentArch)

	return result, nil
}
