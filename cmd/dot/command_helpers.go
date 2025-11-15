package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/doctor"
	"github.com/jamesainslie/dot/pkg/dot"
)

// packageCommandFunc is a function that executes a package operation.
type packageCommandFunc func(*dot.Client, context.Context, []string) error

// executePackageCommand is a helper that handles the common pattern for package commands.
// It builds the config, creates a client, executes the provided function, and prints success message.
func executePackageCommand(cmd *cobra.Command, args []string, fn packageCommandFunc, actionVerb string) error {
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	packages := args

	if err := fn(client, ctx, packages); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	if !cfg.DryRun {
		fmt.Printf("%s %s\n", actionVerb, formatCount(len(packages), "package", "packages"))
	}

	return nil
}

// getAvailablePackages returns list of available packages from the package directory.
func getAvailablePackages() []string {
	packageDir := globalCfg.packageDir
	if packageDir == "" {
		packageDir = "."
	}

	absDir, err := filepath.Abs(packageDir)
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil
	}

	packages := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && !isHiddenOrIgnored(entry.Name()) {
			packages = append(packages, entry.Name())
		}
	}

	return packages
}

// getInstalledPackages returns list of installed packages from the manifest.
func getInstalledPackages() []string {
	cfg, err := buildConfigWithCmd(nil)
	if err != nil {
		return nil
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		return nil
	}

	ctx := context.Background()
	pkgList, err := client.List(ctx)
	if err != nil {
		return nil
	}

	packages := make([]string, 0, len(pkgList))
	for _, pkg := range pkgList {
		packages = append(packages, pkg.Name)
	}

	return packages
}

// isHiddenOrIgnored checks if a directory name should be ignored for completion.
func isHiddenOrIgnored(name string) bool {
	if len(name) == 0 {
		return true
	}
	// Ignore hidden directories (starting with .)
	if name[0] == '.' {
		return true
	}
	// Ignore common non-package directories
	ignoredDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		".svn":         true,
	}
	return ignoredDirs[name]
}

// packageCompletion returns a completion function for package names.
// If installed is true, completes with installed packages, otherwise available packages.
func packageCompletion(installed bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var packages []string
		if installed {
			packages = getInstalledPackages()
		} else {
			packages = getAvailablePackages()
		}
		return packages, cobra.ShellCompDirectiveNoFileComp
	}
}

// derivePackageName derives a package name from a file or directory path.
// Preserves leading dots - scanner will translate to "dot-" prefix.
// Examples:
//
//	.ssh -> .ssh (scanner translates to dot-ssh)
//	.vimrc -> .vimrc (scanner translates to dot-vimrc)
//	.config/nvim -> nvim (base name, no leading dot)
//	README.md -> README.md (no leading dot, no translation)
func derivePackageName(path string) string {
	// Get the base name
	base := filepath.Base(path)

	// Handle special cases
	if base == "." || base == ".." {
		return ""
	}

	// Keep leading dot - scanner.UntranslateDotfile will handle translation
	// ".ssh" stays as ".ssh", which scanner converts to "dot-ssh" for package name
	return base
}

// pluralize returns the singular or plural form based on count.
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// formatCount formats a count with the appropriate singular or plural form.
// Examples:
//
//	formatCount(1, "package", "packages") -> "1 package"
//	formatCount(3, "package", "packages") -> "3 packages"
//	formatCount(0, "file", "files") -> "0 files"
func formatCount(count int, singular, plural string) string {
	return fmt.Sprintf("%d %s", count, pluralize(count, singular, plural))
}

// secretWarning represents a warning about a potential secret file.
type secretWarning struct {
	Path   string
	Reason string
}

// checkPackagesForSecrets scans packages for potential secrets before managing them.
// Returns a list of warnings for files that might contain sensitive information.
func checkPackagesForSecrets(ctx context.Context, client *dot.Client, packages []string) []secretWarning {
	warnings := make([]secretWarning, 0)
	patterns := doctor.DefaultSensitivePatterns()

	// For each package, scan its files for potential secrets
	for _, pkgName := range packages {
		// Get package directory
		pkgDir := filepath.Join(globalCfg.packageDir, pkgName)

		// Get all files in package (recursively)
		files, err := getPackageFiles(pkgDir)
		if err != nil {
			continue
		}

		// Detect secrets in the file list
		detections := doctor.DetectSecrets(files, patterns)

		// Convert detections to warnings
		for _, detection := range detections {
			warnings = append(warnings, secretWarning{
				Path:   detection.Path,
				Reason: detection.Pattern.Description,
			})
		}
	}

	return warnings
}

// checkFilesForSecrets scans individual files for potential secrets.
// Used by the adopt command to warn before adopting sensitive files.
func checkFilesForSecrets(files []string) []secretWarning {
	warnings := make([]secretWarning, 0)
	patterns := doctor.DefaultSensitivePatterns()

	// Detect secrets in the file list
	detections := doctor.DetectSecrets(files, patterns)

	// Convert detections to warnings
	for _, detection := range detections {
		warnings = append(warnings, secretWarning{
			Path:   detection.Path,
			Reason: detection.Pattern.Description,
		})
	}

	return warnings
}

// getPackageFiles returns all files in a package directory recursively.
func getPackageFiles(dir string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
