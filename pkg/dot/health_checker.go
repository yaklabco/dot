package dot

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LinkHealthResult contains detailed health information for a single link.
type LinkHealthResult struct {
	IsHealthy  bool
	IssueType  IssueType
	Severity   IssueSeverity
	Message    string
	Suggestion string
}

// HealthChecker provides unified health checking logic for symlinks.
type HealthChecker struct {
	fs        FS
	targetDir string
}

// newHealthChecker creates a new health checker instance.
func newHealthChecker(fs FS, targetDir string) *HealthChecker {
	return &HealthChecker{
		fs:        fs,
		targetDir: targetDir,
	}
}

// CheckLink validates a single symlink and returns detailed health information.
// This is the single source of truth for link health checking.
func (h *HealthChecker) CheckLink(ctx context.Context, pkgName, linkPath, packageDir string) LinkHealthResult {
	fullPath := filepath.Join(h.targetDir, linkPath)

	// Check if symlink exists using Lstat (doesn't follow symlink)
	linkInfo, err := h.fs.Lstat(ctx, fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
			return LinkHealthResult{
				IsHealthy:  false,
				IssueType:  IssueBrokenLink,
				Severity:   SeverityError,
				Message:    "Link does not exist",
				Suggestion: "Run 'dot remanage " + pkgName + "' to restore link",
			}
		}
		// Permission or other filesystem error
		return LinkHealthResult{
			IsHealthy:  false,
			IssueType:  IssuePermission,
			Severity:   SeverityError,
			Message:    "Cannot access link: " + err.Error(),
			Suggestion: "Check filesystem permissions",
		}
	}

	// Verify it's actually a symlink
	if linkInfo.Mode()&fs.ModeSymlink == 0 {
		return LinkHealthResult{
			IsHealthy:  false,
			IssueType:  IssueWrongTarget,
			Severity:   SeverityError,
			Message:    "Expected symlink but found regular file",
			Suggestion: "Run 'dot unmanage " + pkgName + "' then 'dot manage " + pkgName + "'",
		}
	}

	// Read symlink target
	target, err := h.fs.ReadLink(ctx, fullPath)
	if err != nil {
		return LinkHealthResult{
			IsHealthy:  false,
			IssueType:  IssuePermission,
			Severity:   SeverityError,
			Message:    "Cannot read link target: " + err.Error(),
			Suggestion: "Check filesystem permissions",
		}
	}

	// Resolve target to absolute path
	var absTarget string
	if filepath.IsAbs(target) {
		absTarget = target
	} else {
		absTarget = filepath.Join(filepath.Dir(fullPath), target)
	}

	// Check if target exists using Stat (follows symlink)
	_, err = h.fs.Stat(ctx, absTarget)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
			return LinkHealthResult{
				IsHealthy:  false,
				IssueType:  IssueBrokenLink,
				Severity:   SeverityError,
				Message:    "Link target does not exist: " + target,
				Suggestion: "Run 'dot remanage " + pkgName + "' to fix broken link",
			}
		}
		// Permission or other filesystem error
		return LinkHealthResult{
			IsHealthy:  false,
			IssueType:  IssuePermission,
			Severity:   SeverityError,
			Message:    "Cannot access link target: " + err.Error(),
			Suggestion: "Check file permissions for target path or run with appropriate permissions",
		}
	}

	// Verify target is in package directory (only if package_dir is set)
	// Old adopted packages and some legacy packages don't have package_dir set
	if packageDir != "" && !isInPackageDir(absTarget, packageDir) {
		return LinkHealthResult{
			IsHealthy:  false,
			IssueType:  IssueWrongTarget,
			Severity:   SeverityError,
			Message:    "Link target is outside package directory",
			Suggestion: "Run 'dot remanage " + pkgName + "' to fix target location",
		}
	}

	// All checks passed
	return LinkHealthResult{
		IsHealthy: true,
	}
}

// CheckPackage validates all symlinks for a package and returns aggregated health status.
// Returns healthy status and issue type if problems are found.
func (h *HealthChecker) CheckPackage(ctx context.Context, pkgName string, links []string, packageDir string) (bool, string) {
	brokenLinks := 0
	wrongTargets := 0
	missingLinks := 0
	permissionIssues := 0

	for _, linkPath := range links {
		result := h.CheckLink(ctx, pkgName, linkPath, packageDir)
		if !result.IsHealthy {
			switch result.IssueType {
			case IssueBrokenLink:
				if strings.Contains(result.Message, "does not exist") && !strings.Contains(result.Message, "target") {
					missingLinks++
				} else {
					brokenLinks++
				}
			case IssueWrongTarget:
				wrongTargets++
			case IssuePermission:
				permissionIssues++
			}
		}
	}

	// Determine health status and issue type
	totalIssues := brokenLinks + wrongTargets + missingLinks + permissionIssues
	if totalIssues == 0 {
		return true, ""
	}

	// Return most specific issue type (prioritize by severity)
	if brokenLinks > 0 {
		return false, "broken links"
	}
	if wrongTargets > 0 {
		return false, "wrong target"
	}
	if missingLinks > 0 {
		return false, "missing links"
	}
	if permissionIssues > 0 {
		return false, "permission issues"
	}

	return false, "unknown issue"
}

// isInPackageDir checks if target path is within package directory.
func isInPackageDir(targetPath, packageDir string) bool {
	cleanTarget := filepath.Clean(targetPath)
	cleanPackageDir := filepath.Clean(packageDir)
	return strings.HasPrefix(cleanTarget, cleanPackageDir+string(filepath.Separator)) ||
		cleanTarget == cleanPackageDir
}
