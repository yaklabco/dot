package dot

import (
	"github.com/yaklabco/dot/internal/updater/install"
)

// Installation source type aliases for external use.
type (
	// InstallSource identifies how dot was installed.
	InstallSource = install.Source

	// InstallInfo contains installation detection results.
	InstallInfo = install.Info

	// UpgradeResult contains the outcome of an upgrade attempt.
	UpgradeResult = install.UpgradeResult

	// UpgradeOptions configures upgrade behavior.
	UpgradeOptions = install.UpgradeOptions

	// Detector discovers how dot was installed.
	InstallDetector = install.Detector

	// Upgrader executes upgrades for a specific installation source.
	InstallUpgrader = install.Upgrader
)

// Installation source constants for external use.
const (
	InstallSourceHomebrew   = install.SourceHomebrew
	InstallSourceApt        = install.SourceApt
	InstallSourcePacman     = install.SourcePacman
	InstallSourceChocolatey = install.SourceChocolatey
	InstallSourceGoInstall  = install.SourceGoInstall
	InstallSourceBuild      = install.SourceBuild
	InstallSourceManual     = install.SourceManual
)

// NewInstallDetector creates a new installation detector with the given version.
func NewInstallDetector(version string) InstallDetector {
	return install.NewDetector(install.WithVersion(version))
}

// NewUpgradeOrchestrator creates an upgrade orchestrator.
func NewUpgradeOrchestrator(currentVersion string) *install.UpgradeOrchestrator {
	return install.NewUpgradeOrchestrator(currentVersion)
}

// DefaultUpgradeOptions returns default upgrade options.
func DefaultUpgradeOptions() UpgradeOptions {
	return install.DefaultUpgradeOptions()
}
