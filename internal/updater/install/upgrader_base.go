package install

import (
	"context"
	"fmt"
)

// baseUpgrader provides common functionality for source-specific upgraders.
type baseUpgrader struct {
	source          Source
	executor        *Executor
	detector        Detector
	metadataKey     string // Key to look up package name in metadata
	defaultPkgName  string // Fallback package name
	buildModuleSpec func(info *Info) string
}

// CanUpgrade returns true if this upgrader can handle the installation.
func (u *baseUpgrader) CanUpgrade(info *Info) bool {
	return info != nil && info.Source == u.source
}

// Upgrade executes the upgrade for this source.
func (u *baseUpgrader) Upgrade(ctx context.Context, info *Info) (*UpgradeResult, error) {
	if !u.CanUpgrade(info) {
		return nil, fmt.Errorf("%s upgrader cannot upgrade source: %s", u.source, info.Source)
	}

	result := &UpgradeResult{
		PreviousVersion: info.Version,
	}

	// Get the package/module specification
	var pkgSpec string
	if u.buildModuleSpec != nil {
		pkgSpec = u.buildModuleSpec(info)
	} else {
		pkgSpec = info.Metadata[u.metadataKey]
		if pkgSpec == "" {
			pkgSpec = u.defaultPkgName
		}
	}

	// Create the upgrade command
	cmd, err := NewCommand(u.source, pkgSpec)
	if err != nil {
		result.Error = fmt.Errorf("failed to create upgrade command: %w", err)
		return result, result.Error
	}

	// Execute the upgrade
	output, err := u.executor.Execute(ctx, cmd)
	result.Output = output
	if err != nil {
		result.Error = fmt.Errorf("upgrade command failed: %w", err)
		return result, result.Error
	}

	result.Success = true
	return result, nil
}

// VerifyUpgrade checks if the upgrade was successful by comparing versions.
func (u *baseUpgrader) VerifyUpgrade(ctx context.Context, expectedVersion string) (bool, error) {
	if u.detector == nil {
		return false, fmt.Errorf("no detector available for verification")
	}

	info, err := u.detector.Detect(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to detect version after upgrade: %w", err)
	}

	return info.Version == expectedVersion, nil
}
