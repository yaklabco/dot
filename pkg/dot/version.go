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
type PackageManager = updater.PackageManager

// ResolvePackageManager resolves the package manager based on configuration.
func ResolvePackageManager(configured string) (PackageManager, error) {
	return updater.ResolvePackageManager(configured)
}
