package doctor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaklabco/dot/internal/domain"
)

// ConflictCheck detects filesystem conflicts before operations.
type ConflictCheck struct {
	fs        FS
	targetDir string
	// Packages to check. If empty, check all.
	packages []string
	// Map of package name to link paths relative to targetDir
	packageLinks map[string][]string
}

func NewConflictCheck(fs FS, targetDir string, packageLinks map[string][]string) *ConflictCheck {
	return &ConflictCheck{
		fs:           fs,
		targetDir:    targetDir,
		packageLinks: packageLinks,
	}
}

func (c *ConflictCheck) Name() string {
	return "conflicts"
}

func (c *ConflictCheck) Description() string {
	return "Detects filesystem conflicts before operations"
}

func (c *ConflictCheck) Run(ctx context.Context) (domain.CheckResult, error) {
	result := domain.CheckResult{
		CheckName: c.Name(),
		Status:    domain.CheckStatusPass,
		Issues:    make([]domain.Issue, 0),
		Stats:     make(map[string]any),
	}

	conflicts := 0
	for pkgName, links := range c.packageLinks {
		for _, link := range links {
			fullPath := filepath.Join(c.targetDir, link)

			// Check if something exists at path
			info, err := c.fs.Lstat(ctx, fullPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue // Path is clear
				}
				// Access error
				result.Status = domain.CheckStatusWarning
				result.Issues = append(result.Issues, domain.Issue{
					Code:     "ACCESS_ERROR",
					Message:  fmt.Sprintf("Cannot check path '%s': %v", link, err),
					Severity: domain.IssueSeverityError,
					Path:     link,
				})
				continue
			}

			// Something exists. Is it our symlink?
			if info.Mode()&os.ModeSymlink != 0 {
				// It's a symlink. Ideally check if it points to correct location.
				// For now, just note it exists.
				// If we were doing a strict install check, existing correct symlink is fine.
				// But if we are checking for *conflicts* (i.e. user file vs managed file),
				// existing symlink might be fine or conflict depending on intent.

				// For this check, let's assume any existing file/dir/link is a conflict
				// UNLESS we verify it's already managed correctly.
				// But `packageLinks` typically comes from what we WANT to install.

				// If it's a directory, it might be okay if we are installing into it,
				// but if we want to put a file there, it's a conflict.
				// This logic duplicates planner somewhat.

				// Let's classify carefully.
				continue
			}

			// It exists and is NOT a symlink (file or dir)
			conflicts++
			fileType := "file"
			if info.IsDir() {
				fileType = "directory"
			}

			result.Issues = append(result.Issues, domain.Issue{
				Code:     "CONFLICT_DETECTED",
				Message:  fmt.Sprintf("Path '%s' is occupied by a %s", link, fileType),
				Severity: domain.IssueSeverityError,
				Path:     link,
				Context: map[string]any{
					"package":   pkgName,
					"file_type": fileType,
				},
			})
		}
	}

	result.Stats["conflicts"] = conflicts
	if conflicts > 0 {
		result.Status = domain.CheckStatusFail
	}

	return result, nil
}

// ConflictPermissionCheck verifies write permissions for target directories.
type ConflictPermissionCheck struct {
	fs        FS
	targetDir string
}

func NewConflictPermissionCheck(fs FS, targetDir string) *ConflictPermissionCheck {
	return &ConflictPermissionCheck{
		fs:        fs,
		targetDir: targetDir,
	}
}

func (c *ConflictPermissionCheck) Name() string {
	return "permissions"
}

func (c *ConflictPermissionCheck) Description() string {
	return "Verifies write permissions in target directory"
}

func (c *ConflictPermissionCheck) Run(ctx context.Context) (domain.CheckResult, error) {
	result := domain.CheckResult{
		CheckName: c.Name(),
		Status:    domain.CheckStatusPass,
		Issues:    make([]domain.Issue, 0),
	}

	// Simple check: try to create a temp file in target dir
	testFile := filepath.Join(c.targetDir, ".dot-perm-test")
	err := c.fs.WriteFile(ctx, testFile, []byte("test"), 0644)
	if err != nil {
		if os.IsPermission(err) {
			result.Status = domain.CheckStatusFail
			result.Issues = append(result.Issues, domain.Issue{
				Code:     "PERMISSION_DENIED",
				Message:  "Target directory is not writable",
				Severity: domain.IssueSeverityError,
				Path:     c.targetDir,
			})
			return result, nil
		}
		// Other error
		result.Status = domain.CheckStatusWarning
		result.Issues = append(result.Issues, domain.Issue{
			Code:     "WRITE_TEST_FAILED",
			Message:  fmt.Sprintf("Failed to verify write permissions: %v", err),
			Severity: domain.IssueSeverityWarning,
			Path:     c.targetDir,
		})
		return result, nil
	}
	_ = c.fs.Remove(ctx, testFile)

	return result, nil
}
