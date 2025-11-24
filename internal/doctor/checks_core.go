package doctor

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/internal/domain"
)

// ManagedPackageCheck validates all packages managed by dot.
type ManagedPackageCheck struct {
	fs                 FS
	manifestSvc        ManifestLoader
	healthChecker      LinkHealthChecker
	targetDir          string
	newTargetPath      TargetPathCreator
	isManifestNotFound ManifestNotFoundChecker
}

func NewManagedPackageCheck(
	fs FS,
	manifestSvc ManifestLoader,
	healthChecker LinkHealthChecker,
	targetDir string,
	newTargetPath TargetPathCreator,
	isManifestNotFound ManifestNotFoundChecker,
) *ManagedPackageCheck {
	return &ManagedPackageCheck{
		fs:                 fs,
		manifestSvc:        manifestSvc,
		healthChecker:      healthChecker,
		targetDir:          targetDir,
		newTargetPath:      newTargetPath,
		isManifestNotFound: isManifestNotFound,
	}
}

func (c *ManagedPackageCheck) Name() string {
	return "managed_packages"
}

func (c *ManagedPackageCheck) Description() string {
	return "Validates status of all managed packages and symlinks"
}

func (c *ManagedPackageCheck) Run(ctx domain.Context) (domain.CheckResult, error) {
	result := domain.CheckResult{
		CheckName: c.Name(),
		Status:    domain.CheckStatusPass,
		Issues:    make([]domain.Issue, 0),
		Stats:     make(map[string]any),
	}

	// Load manifest
	// Convert domain.Context to context.Context
	stdCtx, ok := ctx.(context.Context)
	if !ok {
		return result, fmt.Errorf("invalid context type")
	}

	// Construct target path
	targetPathResult := c.newTargetPath.NewTargetPath(c.targetDir)
	if !targetPathResult.IsOk() {
		return result, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	manifestResult := c.manifestSvc.Load(stdCtx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		if c.isManifestNotFound(err) {
			result.Status = domain.CheckStatusSkipped
			result.Issues = append(result.Issues, domain.Issue{
				Code:     "NO_MANIFEST",
				Message:  "No manifest found - no packages managed",
				Severity: domain.IssueSeverityInfo,
			})
			return result, nil
		}
		return result, err
	}

	m := manifestResult.Unwrap()

	totalLinks := 0
	brokenLinks := 0
	managedLinks := 0

	for pkgName, pkgInfo := range m.Packages {
		managedLinks += pkgInfo.LinkCount
		for _, linkPath := range pkgInfo.Links {
			totalLinks++
			healthResult := c.healthChecker.CheckLink(stdCtx, pkgName, linkPath, pkgInfo.PackageDir)

			if !healthResult.IsHealthy {
				brokenLinks++

				severity := domain.IssueSeverityError
				if healthResult.Severity == domain.IssueSeverityWarning {
					severity = domain.IssueSeverityWarning
				}

				result.Issues = append(result.Issues, domain.Issue{
					Code:     string(healthResult.IssueType),
					Message:  healthResult.Message,
					Severity: severity,
					Path:     linkPath,
					Context: map[string]any{
						"package":    pkgName,
						"suggestion": healthResult.Suggestion,
					},
				})
			}
		}
	}

	result.Stats["total_links"] = totalLinks
	result.Stats["broken_links"] = brokenLinks
	result.Stats["managed_links"] = managedLinks

	if brokenLinks > 0 {
		result.Status = domain.CheckStatusFail
	}

	return result, nil
}

// ManifestIntegrityCheck validates the manifest file itself.
type ManifestIntegrityCheck struct {
	fs                 FS
	targetDir          string
	manifestSvc        ManifestLoader
	newTargetPath      TargetPathCreator
	isManifestNotFound ManifestNotFoundChecker
}

func NewManifestIntegrityCheck(
	fs FS,
	manifestSvc ManifestLoader,
	targetDir string,
	newTargetPath TargetPathCreator,
	isManifestNotFound ManifestNotFoundChecker,
) *ManifestIntegrityCheck {
	return &ManifestIntegrityCheck{
		fs:                 fs,
		manifestSvc:        manifestSvc,
		targetDir:          targetDir,
		newTargetPath:      newTargetPath,
		isManifestNotFound: isManifestNotFound,
	}
}

func (c *ManifestIntegrityCheck) Name() string {
	return "manifest_integrity"
}

func (c *ManifestIntegrityCheck) Description() string {
	return "Validates integrity and consistency of the manifest file"
}

func (c *ManifestIntegrityCheck) Run(ctx domain.Context) (domain.CheckResult, error) {
	result := domain.CheckResult{
		CheckName: c.Name(),
		Status:    domain.CheckStatusPass,
		Issues:    make([]domain.Issue, 0),
	}

	stdCtx, ok := ctx.(context.Context)
	if !ok {
		return result, fmt.Errorf("invalid context type")
	}

	targetPathResult := c.newTargetPath.NewTargetPath(c.targetDir)
	if !targetPathResult.IsOk() {
		return result, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// We load the manifest to check validity.
	// Ideally we'd parse raw JSON to check syntax if Load fails,
	// but Load handles unmarshaling errors.
	manifestResult := c.manifestSvc.Load(stdCtx, targetPath)

	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		if c.isManifestNotFound(err) {
			// Missing manifest is valid state for new installs
			return result, nil
		}

		// System-level errors (permission, IO errors) should be returned as errors
		// to signal that the check couldn't complete due to system issues
		return result, fmt.Errorf("cannot access manifest: %w", err)
	}

	m := manifestResult.Unwrap()

	// Consistency checks
	for pkgName, pkg := range m.Packages {
		if pkg.LinkCount != len(pkg.Links) {
			result.Status = domain.CheckStatusWarning
			result.Issues = append(result.Issues, domain.Issue{
				Code:     "MANIFEST_INCONSISTENT",
				Message:  fmt.Sprintf("Package '%s' has link count mismatch (recorded: %d, actual: %d)", pkgName, pkg.LinkCount, len(pkg.Links)),
				Severity: domain.IssueSeverityWarning,
				Context: map[string]any{
					"package": pkgName,
				},
			})
		}
	}

	return result, nil
}
