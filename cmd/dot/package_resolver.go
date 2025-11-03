package main

import (
	"os"
	"path/filepath"
)

// resolvePackageDirectory resolves the package directory using hierarchical discovery.
//
// Resolution order (highest to lowest priority):
//  1. Explicit --dir flag (if not ".")
//  2. Environment variable: DOT_PACKAGE_DIR
//  3. Current directory if it contains .dotbootstrap.yaml
//  4. Parent directories up to home (searching for .dotbootstrap.yaml)
//  5. Config file: directories.package
//  6. Default: ~/.dotfiles
func resolvePackageDirectory(explicitDir string) (string, error) {
	// 1. Explicit --dir flag (highest priority)
	if explicitDir != "" && explicitDir != "." {
		return filepath.Abs(explicitDir)
	}

	// 2. Environment variable: DOT_PACKAGE_DIR
	if envDir := os.Getenv("DOT_PACKAGE_DIR"); envDir != "" {
		return filepath.Abs(envDir)
	}

	// 3. Current directory if it contains .dotbootstrap.yaml
	cwd, err := os.Getwd()
	if err == nil && isDotfilesRepo(cwd) {
		return cwd, nil
	}

	// 4. Search parent directories up to home
	if err == nil {
		if repoDir := findDotfilesRepo(cwd); repoDir != "" {
			return repoDir, nil
		}
	}

	// 5. Config file: directories.package
	configPath := getConfigFilePath()
	cfg, _ := loadConfigWithRepoPriority(configPath)
	if cfg != nil && cfg.Directories.Package != "" {
		abs, err := filepath.Abs(cfg.Directories.Package)
		if err == nil {
			return abs, nil
		}
	}

	// 6. Default: ~/.dotfiles
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".dotfiles"), nil
}

// isDotfilesRepo checks if the given directory is a dotfiles repository
// by looking for .dotbootstrap.yaml file.
func isDotfilesRepo(dir string) bool {
	bootstrapPath := filepath.Join(dir, ".dotbootstrap.yaml")
	_, err := os.Stat(bootstrapPath)
	return err == nil
}

// findDotfilesRepo searches parent directories for a dotfiles repository.
// It stops at the home directory or root.
func findDotfilesRepo(startDir string) string {
	homeDir, _ := os.UserHomeDir()
	dir := startDir

	for {
		if isDotfilesRepo(dir) {
			return dir
		}

		parent := filepath.Dir(dir)
		// Stop at root, home, or when parent == dir (reached root)
		if parent == dir || parent == homeDir || parent == "/" {
			break
		}
		dir = parent
	}

	return ""
}
