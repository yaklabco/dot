package planner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DotOperationalPaths returns the paths that dot uses for its own operation.
// Packages cannot create symlinks into these directories.
func DotOperationalPaths() []string {
	homeDir, _ := os.UserHomeDir()

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
	}
}

// ValidateNoSelfManagement validates that a package does not attempt to manage
// dot's own operational directories.
//
// Returns an error if:
// - Any link would be created inside dot's config directory (~/.config/dot/)
// - Any link would be created inside dot's data directory (~/.local/share/dot/)
// - Any link would override dot's operational directories
func ValidateNoSelfManagement(packageName string, desired DesiredState) error {
	protectedPaths := DotOperationalPaths()

	// Check all desired links
	for linkPath := range desired.Links {
		absPath, err := filepath.Abs(linkPath)
		if err != nil {
			continue // Skip if we can't resolve path
		}

		for _, protected := range protectedPaths {
			absProtected, err := filepath.Abs(protected)
			if err != nil {
				continue
			}

			// Check if link would be inside protected path
			if strings.HasPrefix(absPath, absProtected+string(filepath.Separator)) ||
				absPath == absProtected {
				return fmt.Errorf(
					"package %q attempts to create symlinks in dot's operational directory: %s\n"+
						"Packages cannot manage dot's configuration or data directories.\n"+
						"Add these paths to .dotignore or remove them from the package",
					packageName, protected,
				)
			}

			// Check if protected path would be inside link path
			if strings.HasPrefix(absProtected, absPath+string(filepath.Separator)) {
				return fmt.Errorf(
					"package %q would override dot's operational directory: %s",
					packageName, protected,
				)
			}
		}
	}

	// Also check directories
	for dirPath := range desired.Dirs {
		absPath, err := filepath.Abs(dirPath)
		if err != nil {
			continue
		}

		for _, protected := range protectedPaths {
			absProtected, err := filepath.Abs(protected)
			if err != nil {
				continue
			}

			// Check if directory would be inside protected path
			if strings.HasPrefix(absPath, absProtected+string(filepath.Separator)) ||
				absPath == absProtected {
				return fmt.Errorf(
					"package %q attempts to create directories in dot's operational path: %s",
					packageName, protected,
				)
			}

			// Check if protected path would be inside directory path
			if strings.HasPrefix(absProtected, absPath+string(filepath.Separator)) {
				return fmt.Errorf(
					"package %q would override dot's operational directory: %s",
					packageName, protected,
				)
			}
		}
	}

	return nil
}
