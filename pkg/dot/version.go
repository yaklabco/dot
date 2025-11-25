package dot

import "github.com/yaklabco/dot/internal/updater"

// VersionChecker checks for new versions on GitHub.
type VersionChecker struct {
	checker *updater.VersionChecker
}

// NewVersionChecker creates a new version checker.
func NewVersionChecker(repository string) *VersionChecker {
	return &VersionChecker{
		checker: updater.NewVersionChecker(repository),
	}
}

// CheckForUpdate checks if a new version is available.
func (v *VersionChecker) CheckForUpdate(currentVersion string, includePrerelease bool) (*GitHubRelease, bool, error) {
	return v.checker.CheckForUpdate(currentVersion, includePrerelease)
}

// PackageManager represents a package manager for upgrades.
//
// Deprecated: Use NewUpgradeOrchestrator instead for automatic installation
// detection and upgrade functionality.
type PackageManager = updater.PackageManager

// ResolvePackageManager resolves the package manager based on configuration.
//
// Deprecated: Use NewUpgradeOrchestrator instead for automatic installation
// detection and upgrade functionality.
func ResolvePackageManager(configured string) (PackageManager, error) {
	return updater.ResolvePackageManager(configured)
}
