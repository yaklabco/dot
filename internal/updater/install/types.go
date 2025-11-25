package install

import (
	"context"
)

// Source identifies how dot was installed.
type Source string

const (
	// SourceHomebrew indicates installation via Homebrew (macOS/Linux).
	SourceHomebrew Source = "homebrew"

	// SourceApt indicates installation via APT/dpkg (Debian/Ubuntu).
	SourceApt Source = "apt"

	// SourcePacman indicates installation via Pacman (Arch Linux).
	SourcePacman Source = "pacman"

	// SourceChocolatey indicates installation via Chocolatey (Windows).
	SourceChocolatey Source = "chocolatey"

	// SourceGoInstall indicates installation via go install.
	SourceGoInstall Source = "go-install"

	// SourceBuild indicates a development/source build.
	SourceBuild Source = "source"

	// SourceManual indicates manual installation or unknown source.
	SourceManual Source = "manual"
)

// String returns the string representation of the source.
func (s Source) String() string {
	return string(s)
}

// Info contains installation detection results.
type Info struct {
	// Source identifies the installation method.
	Source Source

	// Version is the installed version if detectable.
	// Empty string if version cannot be determined.
	Version string

	// ExecutablePath is the resolved path to the dot binary.
	ExecutablePath string

	// Metadata contains source-specific information.
	// For Homebrew: formula, cellar, tap
	// For dpkg: package, architecture, status
	// For Go install: module, goVersion
	Metadata map[string]string

	// CanAutoUpgrade indicates if automatic upgrade is possible.
	CanAutoUpgrade bool

	// UpgradeInstructions provides human-readable upgrade guidance.
	UpgradeInstructions string
}

// Probe detects a specific installation source.
type Probe interface {
	// Name returns the probe identifier.
	Name() string

	// Platforms returns the platforms this probe supports.
	// Empty slice means all platforms.
	Platforms() []string

	// Detect checks if this probe matches the installation.
	// Returns nil, nil if this probe does not match.
	// Returns nil, error if detection failed with an error.
	Detect(ctx context.Context, execPath string) (*Info, error)
}

// Detector discovers how dot was installed.
type Detector interface {
	// Detect returns installation information.
	Detect(ctx context.Context) (*Info, error)
}

// UpgradeResult contains the outcome of an upgrade attempt.
type UpgradeResult struct {
	// Success indicates whether the upgrade completed successfully.
	Success bool

	// PreviousVersion is the version before the upgrade.
	PreviousVersion string

	// NewVersion is the version after the upgrade (if verification succeeded).
	NewVersion string

	// Output contains any output from the upgrade command.
	Output string

	// Error contains any error that occurred during upgrade.
	Error error
}

// Upgrader executes upgrades for a specific installation source.
type Upgrader interface {
	// CanUpgrade returns true if this upgrader can handle the installation.
	CanUpgrade(info *Info) bool

	// Upgrade executes the upgrade and returns the result.
	Upgrade(ctx context.Context, info *Info) (*UpgradeResult, error)

	// VerifyUpgrade checks if the upgrade was successful by comparing versions.
	VerifyUpgrade(ctx context.Context, expectedVersion string) (bool, error)
}
