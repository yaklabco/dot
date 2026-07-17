package planner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DotOperationalPaths returns the paths that dot uses for its own operation.
// Packages cannot create symlinks into these directories.
// Returns an error if the home directory cannot be determined.
func DotOperationalPaths() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determine home directory: %w", err)
	}

	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(homeDir, ".config")
	}

	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData == "" {
		xdgData = filepath.Join(homeDir, ".local", "share")
	}

	return []string{
		filepath.Join(xdgConfig, "dot"),
		filepath.Join(xdgData, "dot"),
	}, nil
}

// checkPathAgainstProtected checks if a path conflicts with a protected path.
// Returns an error describing the conflict, or nil if no conflict.
func checkPathAgainstProtected(absPath, absProtected string) (insideProtected, overridesProtected bool) {
	// Check if path would be inside protected path
	if strings.HasPrefix(absPath, absProtected+string(filepath.Separator)) || absPath == absProtected {
		return true, false
	}
	// Check if protected path would be inside path
	if strings.HasPrefix(absProtected, absPath+string(filepath.Separator)) {
		return false, true
	}
	return false, false
}

// resourceKind identifies the kind of desired-state resource being validated.
// It gates the parent-override rejection below, so it is a closed enum rather
// than a free-form string.
type resourceKind int

const (
	resourceSymlink resourceKind = iota
	resourceDirectory
)

// ValidateNoSelfManagement validates that a package does not attempt to manage
// dot's own operational directories.
//
// Returns an error if:
// - Any link would be created inside dot's config directory (~/.config/dot/)
// - Any link would be created inside dot's data directory (~/.local/share/dot/)
// - Any link would override dot's operational directories
func ValidateNoSelfManagement(packageName string, desired DesiredState) error {
	protectedPaths, err := DotOperationalPaths()
	if err != nil {
		// If we can't determine protected paths, skip validation rather than block
		// This prevents users from being locked out if $HOME is not set
		return nil
	}

	// Check all desired links
	for linkPath := range desired.Links {
		if err := checkPathConflicts(packageName, linkPath, protectedPaths, resourceSymlink); err != nil {
			return err
		}
	}

	// Also check directories
	for dirPath := range desired.Dirs {
		if err := checkPathConflicts(packageName, dirPath, protectedPaths, resourceDirectory); err != nil {
			return err
		}
	}

	return nil
}

// checkPathConflicts checks if a path conflicts with any protected paths.
func checkPathConflicts(packageName, path string, protectedPaths []string, kind resourceKind) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil // Skip if we can't resolve path
	}

	for _, protected := range protectedPaths {
		absProtected, err := filepath.Abs(protected)
		if err != nil {
			continue
		}

		insideProtected, overridesProtected := checkPathAgainstProtected(absPath, absProtected)

		if insideProtected {
			if kind == resourceSymlink {
				return fmt.Errorf(
					"package %q attempts to create symlinks in dot's operational directory: %s\n"+
						"Packages cannot manage dot's configuration or data directories.\n"+
						"Add these paths to .dotignore or remove them from the package",
					packageName, protected,
				)
			}
			return fmt.Errorf(
				"package %q attempts to create directories in dot's operational path: %s",
				packageName, protected,
			)
		}

		// A symlink at a parent of a protected path would make the protected
		// path resolve into package-controlled space, so it stays rejected.
		// Creating a plain directory at a parent (e.g. ~/.config when
		// ~/.config/dot is protected) is safe: the protected path continues
		// to exist inside it untouched.
		if overridesProtected && kind == resourceSymlink {
			return fmt.Errorf(
				"package %q would override dot's operational directory: %s",
				packageName, protected,
			)
		}
	}

	return nil
}
