package dot

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/manifest"
)

// FixOptions configures fix behavior.
type FixOptions struct {
	DryRun      bool
	AutoConfirm bool // --yes flag
	Interactive bool // Prompt user for decisions
}

// FixResult contains the results of a fix operation.
type FixResult struct {
	Fixed   []string
	Skipped []string
	Errors  map[string]error
}

// issueGroup groups issues by category for batch processing.
type issueGroup struct {
	Category string
	Issues   []Issue
}

// Fix repairs broken symlinks found during doctor scan.
func (s *DoctorService) Fix(ctx context.Context, scanCfg ScanConfig, opts FixOptions) (FixResult, error) {
	// Run doctor to get issues
	report, err := s.DoctorWithScan(ctx, scanCfg)
	if err != nil {
		return FixResult{}, err
	}

	result := FixResult{
		Errors: make(map[string]error),
	}

	// Load manifest
	targetPath, err := s.getTargetPath()
	if err != nil {
		return result, err
	}

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return result, manifestResult.UnwrapErr()
	}
	m := manifestResult.Unwrap()

	// Group issues for batch prompting
	groupedIssues := s.groupIssuesForFix(report.Issues, &m)

	// Process each group
	for _, group := range groupedIssues {
		s.processFixGroup(ctx, &m, group, opts, &result)
	}

	// Save manifest if changes made
	if len(result.Fixed) > 0 && !opts.DryRun {
		if err := s.manifestSvc.Save(ctx, targetPath, m); err != nil {
			return result, fmt.Errorf("failed to save manifest: %w", err)
		}
	}

	return result, nil
}

// groupIssuesForFix groups issues by type and managed status for batch processing.
func (s *DoctorService) groupIssuesForFix(issues []Issue, m *manifest.Manifest) []issueGroup {
	groups := []issueGroup{}

	// Group managed broken links
	managedBroken := []Issue{}
	// Group unmanaged broken links
	unmanagedBroken := []Issue{}

	for _, issue := range issues {
		if issue.Type == IssueBrokenLink {
			// Check if link is managed
			if s.isManagedLink(issue.Path, m) {
				managedBroken = append(managedBroken, issue)
			} else {
				unmanagedBroken = append(unmanagedBroken, issue)
			}
		}
	}

	if len(managedBroken) > 0 {
		groups = append(groups, issueGroup{
			Category: "Managed broken links",
			Issues:   managedBroken,
		})
	}

	if len(unmanagedBroken) > 0 {
		groups = append(groups, issueGroup{
			Category: "Unmanaged broken links",
			Issues:   unmanagedBroken,
		})
	}

	return groups
}

// isManagedLink checks if a link path is managed by any package.
func (s *DoctorService) isManagedLink(linkPath string, m *manifest.Manifest) bool {
	for _, pkg := range m.Packages {
		for _, link := range pkg.Links {
			if link == linkPath {
				return true
			}
		}
	}
	return false
}

// processFixGroup processes a group of issues with batched user prompts.
func (s *DoctorService) processFixGroup(ctx context.Context, m *manifest.Manifest, group issueGroup, opts FixOptions, result *FixResult) {
	applyToAll := false
	applyToAllDecision := false

	for _, issue := range group.Issues {
		// Skip if apply-to-all was set
		if applyToAll {
			if applyToAllDecision {
				if err := s.fixIssue(ctx, issue, m, opts); err != nil {
					result.Errors[issue.Path] = err
				} else {
					result.Fixed = append(result.Fixed, issue.Path)
				}
			} else {
				result.Skipped = append(result.Skipped, issue.Path)
			}
			continue
		}

		// Auto-confirm if requested
		if opts.AutoConfirm {
			if err := s.fixIssue(ctx, issue, m, opts); err != nil {
				result.Errors[issue.Path] = err
			} else {
				result.Fixed = append(result.Fixed, issue.Path)
			}
			continue
		}

		// Interactive prompt (default behavior when Interactive=true or both flags are false)
		// Default to interactive mode to prevent silently dropping issues
		decision, all := s.promptFixDecision(ctx, issue, group.Category, m)
		if all {
			applyToAll = true
			applyToAllDecision = decision
		}

		if decision {
			if err := s.fixIssue(ctx, issue, m, opts); err != nil {
				result.Errors[issue.Path] = err
			} else {
				result.Fixed = append(result.Fixed, issue.Path)
			}
		} else {
			result.Skipped = append(result.Skipped, issue.Path)
		}
	}
}

// promptFixDecision prompts user for fix decision.
// Returns (shouldFix, applyToAll).
func (s *DoctorService) promptFixDecision(ctx context.Context, issue Issue, category string, m *manifest.Manifest) (bool, bool) {
	fmt.Printf("\nFix issue at %s?\n", issue.Path)
	fmt.Printf("  %s\n", issue.Message)
	fmt.Printf("  Category: %s\n", category)

	// Explain what action will be taken
	pkgName := s.findPackageForLink(issue.Path, m)
	if pkgName != "" {
		// Managed link - check if source exists
		sourcePath := s.constructSourcePath(pkgName, issue.Path)
		if s.fs.Exists(ctx, sourcePath) {
			fmt.Printf("\n  Action: Recreate symlink from package source\n")
			fmt.Printf("  Source: %s\n", sourcePath)
		} else {
			fmt.Printf("\n  Action: Remove broken link (source no longer exists)\n")
			fmt.Printf("  Package: %s\n", pkgName)
		}
	} else {
		// Unmanaged link
		fmt.Printf("\n  Action: Remove broken symlink\n")
	}

	fmt.Printf("\nOptions:\n")
	fmt.Printf("  y - Yes, fix this\n")
	fmt.Printf("  n - No, skip this\n")
	fmt.Printf("  a - Yes to all in this category\n")
	fmt.Printf("  x - No to all in this category\n")
	fmt.Printf("\nChoice [y/n/a/x]: ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Handle EOF or input error by defaulting to skip
		return false, false
	}

	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "y", "yes", "":
		return true, false
	case "n", "no":
		return false, false
	case "a", "all":
		return true, true
	case "x":
		return false, true
	default:
		return false, false
	}
}

// fixIssue fixes a single issue based on its type.
func (s *DoctorService) fixIssue(ctx context.Context, issue Issue, m *manifest.Manifest, opts FixOptions) error {
	if opts.DryRun {
		s.logger.Info(ctx, "dry_run_fix", "path", issue.Path, "type", issue.Type)
		return nil
	}

	switch issue.Type {
	case IssueBrokenLink:
		// Find which package this link belongs to
		pkgName := s.findPackageForLink(issue.Path, m)
		if pkgName != "" {
			// Managed link - try to recreate from source
			return s.fixBrokenManagedLink(ctx, pkgName, issue.Path, m)
		}
		// Unmanaged link - just remove it
		return s.fixBrokenUnmanagedLink(ctx, issue.Path)
	default:
		return fmt.Errorf("unsupported issue type for fix: %v", issue.Type)
	}
}

// findPackageForLink finds which package manages a given link path.
func (s *DoctorService) findPackageForLink(linkPath string, m *manifest.Manifest) string {
	for pkgName, pkg := range m.Packages {
		for _, link := range pkg.Links {
			if link == linkPath {
				return pkgName
			}
		}
	}
	return ""
}

// fixBrokenManagedLink tries to recreate from package source, falls back to removal.
func (s *DoctorService) fixBrokenManagedLink(ctx context.Context, pkgName, linkPath string, m *manifest.Manifest) error {
	pkg, exists := m.GetPackage(pkgName)
	if !exists {
		return fmt.Errorf("package not found in manifest: %s", pkgName)
	}

	// Construct source path
	sourcePath := s.constructSourcePath(pkgName, linkPath)

	// Check if source exists
	if s.fs.Exists(ctx, sourcePath) {
		// Source exists - recreate symlink
		fullPath := filepath.Join(s.targetDir, linkPath)

		// Remove broken link if exists
		_ = s.fs.Remove(ctx, fullPath)

		// Ensure parent directory exists
		parentDir := filepath.Dir(fullPath)
		if err := s.fs.MkdirAll(ctx, parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Create new symlink
		if err := s.fs.Symlink(ctx, sourcePath, fullPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		s.logger.Info(ctx, "recreated_symlink", "path", linkPath, "source", sourcePath)
		return nil
	}

	// Source doesn't exist - remove from manifest
	links := pkg.Links
	for i, link := range links {
		if link == linkPath {
			links = append(links[:i], links[i+1:]...)
			break
		}
	}
	pkg.Links = links
	pkg.LinkCount = len(links)

	if len(links) == 0 {
		// No links left - remove package
		m.RemovePackage(pkgName)
		s.logger.Info(ctx, "removed_empty_package", "package", pkgName)
	} else {
		m.AddPackage(pkg)
	}

	// Remove the broken symlink from filesystem
	fullPath := filepath.Join(s.targetDir, linkPath)
	if err := s.fs.Remove(ctx, fullPath); err != nil {
		s.logger.Warn(ctx, "failed_to_remove_broken_link", "path", fullPath, "error", err)
	}

	s.logger.Info(ctx, "removed_broken_link_no_source", "path", linkPath, "package", pkgName)
	return nil
}

// fixBrokenUnmanagedLink removes an unmanaged broken symlink.
func (s *DoctorService) fixBrokenUnmanagedLink(ctx context.Context, linkPath string) error {
	fullPath := filepath.Join(s.targetDir, linkPath)

	if err := s.fs.Remove(ctx, fullPath); err != nil {
		return fmt.Errorf("failed to remove broken link: %w", err)
	}

	s.logger.Info(ctx, "removed_unmanaged_broken_link", "path", linkPath)
	return nil
}

// constructSourcePath builds the expected source path for a link.
func (s *DoctorService) constructSourcePath(pkgName, linkPath string) string {
	// Remove dot- prefix if present for directory name mapping
	targetName := linkPath
	if strings.HasPrefix(filepath.Base(targetName), ".") {
		// It's a hidden file - need to find corresponding source file with dot- prefix
		baseWithoutDot := strings.TrimPrefix(filepath.Base(targetName), ".")
		dirPart := filepath.Dir(targetName)
		if dirPart == "." {
			targetName = "dot-" + baseWithoutDot
		} else {
			targetName = filepath.Join(dirPart, "dot-"+baseWithoutDot)
		}
	}

	return filepath.Join(s.packageDir, pkgName, targetName)
}
