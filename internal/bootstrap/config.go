// Package bootstrap provides configuration schema for repository setup.
//
// The bootstrap configuration allows repository owners to specify which
// packages should be installed, installation profiles, platform-specific
// packages, and conflict resolution policies.
package bootstrap

import (
	"fmt"
)

// Config represents the bootstrap configuration for a dotfiles repository.
type Config struct {
	// Version specifies the bootstrap config schema version.
	Version string `yaml:"version"`

	// Packages lists all available packages in the repository.
	Packages []PackageSpec `yaml:"packages"`

	// Profiles defines named sets of packages for different use cases.
	Profiles map[string]Profile `yaml:"profiles,omitempty"`

	// Defaults specifies default settings for installation.
	Defaults Defaults `yaml:"defaults,omitempty"`
}

// PackageSpec defines a package and its installation requirements.
type PackageSpec struct {
	// Name is the package directory name.
	Name string `yaml:"name"`

	// Required indicates if this package must be installed.
	Required bool `yaml:"required"`

	// Platform restricts installation to specific operating systems.
	// Valid values: linux, darwin, windows, freebsd
	Platform []string `yaml:"platform,omitempty"`

	// ConflictPolicy specifies how to handle conflicts for this package.
	// Valid values: fail, backup, overwrite, skip
	ConflictPolicy string `yaml:"on_conflict,omitempty"`
}

// Profile represents a named set of packages.
type Profile struct {
	// Description provides human-readable explanation of the profile.
	Description string `yaml:"description"`

	// Packages lists the package names included in this profile.
	Packages []string `yaml:"packages"`
}

// Defaults specifies default configuration values.
type Defaults struct {
	// ConflictPolicy is the default conflict resolution strategy.
	// Valid values: fail, backup, overwrite, skip
	ConflictPolicy string `yaml:"on_conflict"`

	// Profile is the default profile to use if none specified.
	Profile string `yaml:"profile"`
}

// Validate checks the configuration for errors.
//
// Returns an error if:
//   - Version is missing or empty
//   - No packages are defined
//   - Package names are empty or duplicated
//   - Invalid platform names are used
//   - Invalid conflict policies are specified
//   - Profiles reference non-existent packages
//   - Default profile does not exist
func (c Config) Validate() error {
	// Check version
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}

	// Validate packages and build name set (empty package list is allowed)
	packageNames, err := c.validatePackages()
	if err != nil {
		return err
	}

	// Validate defaults
	if err := c.validateDefaults(); err != nil {
		return err
	}

	// Validate profiles reference valid packages
	if err := c.validateProfiles(packageNames); err != nil {
		return err
	}

	return nil
}

// validatePackages validates all packages and returns their names.
func (c Config) validatePackages() (map[string]struct{}, error) {
	packageNames := make(map[string]struct{})

	for _, pkg := range c.Packages {
		// Check package name
		if pkg.Name == "" {
			return nil, fmt.Errorf("package name cannot be empty")
		}

		// Check for duplicates
		if _, exists := packageNames[pkg.Name]; exists {
			return nil, fmt.Errorf("duplicate package name: %s", pkg.Name)
		}
		packageNames[pkg.Name] = struct{}{}

		// Validate platforms
		for _, platform := range pkg.Platform {
			if !isValidPlatform(platform) {
				return nil, fmt.Errorf("invalid platform %q for package %s", platform, pkg.Name)
			}
		}

		// Validate conflict policy
		if pkg.ConflictPolicy != "" && !isValidConflictPolicy(pkg.ConflictPolicy) {
			return nil, fmt.Errorf("invalid conflict policy %q for package %s", pkg.ConflictPolicy, pkg.Name)
		}
	}

	return packageNames, nil
}

// validateDefaults validates default configuration.
func (c Config) validateDefaults() error {
	// Validate default conflict policy
	if c.Defaults.ConflictPolicy != "" && !isValidConflictPolicy(c.Defaults.ConflictPolicy) {
		return fmt.Errorf("invalid conflict policy in defaults: %s", c.Defaults.ConflictPolicy)
	}

	// Validate default profile exists
	if c.Defaults.Profile != "" {
		if _, exists := c.Profiles[c.Defaults.Profile]; !exists {
			return fmt.Errorf("default profile %q does not exist", c.Defaults.Profile)
		}
	}

	return nil
}

// validateProfiles validates that profiles reference valid packages.
func (c Config) validateProfiles(packageNames map[string]struct{}) error {
	for profileName, profile := range c.Profiles {
		for _, pkgName := range profile.Packages {
			if _, exists := packageNames[pkgName]; !exists {
				return fmt.Errorf("profile %q references unknown package: %s", profileName, pkgName)
			}
		}
	}
	return nil
}

// isValidPlatform checks if a platform name is supported.
func isValidPlatform(platform string) bool {
	switch platform {
	case "linux", "darwin", "windows", "freebsd":
		return true
	default:
		return false
	}
}

// isValidConflictPolicy checks if a conflict policy is supported.
func isValidConflictPolicy(policy string) bool {
	switch policy {
	case "fail", "backup", "overwrite", "skip":
		return true
	default:
		return false
	}
}
