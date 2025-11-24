package doctor

import (
	"context"
	"fmt"
	"os"

	"github.com/yaklabco/dot/internal/domain"
)

// PermissionCheck validates filesystem permissions for operations.
type PermissionCheck struct {
	fs        FS
	targetDir string
}

// NewPermissionCheck creates a new permission check.
func NewPermissionCheck(fs FS, targetDir string) *PermissionCheck {
	return &PermissionCheck{
		fs:        fs,
		targetDir: targetDir,
	}
}

func (c *PermissionCheck) Name() string {
	return "permissions"
}

func (c *PermissionCheck) Description() string {
	return "Validates filesystem permissions for dot operations"
}

func (c *PermissionCheck) Run(ctx context.Context) (domain.CheckResult, error) {
	result := domain.CheckResult{
		CheckName: c.Name(),
		Status:    domain.CheckStatusPass,
		Issues:    make([]domain.Issue, 0),
		Stats:     make(map[string]any),
	}

	// Check if target directory exists
	exists, err := c.fs.Exists(ctx, c.targetDir)
	if err != nil {
		return result, fmt.Errorf("failed to check target directory: %w", err)
	}

	if !exists {
		result.Status = domain.CheckStatusWarning
		result.Issues = append(result.Issues, domain.Issue{
			Code:     "TARGET_DIR_MISSING",
			Message:  fmt.Sprintf("Target directory does not exist: %s", c.targetDir),
			Severity: domain.IssueSeverityWarning,
			Path:     c.targetDir,
			Remediation: &domain.Remediation{
				Description: "The target directory will be created when you manage packages",
			},
		})
		return result, nil
	}

	// Check write permission to target directory
	testFile := fmt.Sprintf("%s/.dot-permission-test", c.targetDir)
	if err := c.fs.WriteFile(ctx, testFile, []byte("test"), 0600); err != nil {
		result.Status = domain.CheckStatusFail
		result.Issues = append(result.Issues, domain.Issue{
			Code:     "TARGET_DIR_NOT_WRITABLE",
			Message:  fmt.Sprintf("Cannot write to target directory: %s", c.targetDir),
			Severity: domain.IssueSeverityError,
			Path:     c.targetDir,
			Context: map[string]any{
				"error": err.Error(),
			},
			Remediation: &domain.Remediation{
				Description: fmt.Sprintf("Ensure you have write permissions: chmod u+w %s", c.targetDir),
			},
		})
		return result, nil
	}

	// Clean up test file
	if err := c.fs.Remove(ctx, testFile); err != nil {
		// Warning about cleanup failure
		result.Issues = append(result.Issues, domain.Issue{
			Code:     "CLEANUP_FAILED",
			Message:  fmt.Sprintf("Failed to remove test file: %s", testFile),
			Severity: domain.IssueSeverityWarning,
			Path:     testFile,
			Context: map[string]any{
				"error": err.Error(),
			},
		})
		// Don't fail the whole check just for cleanup, but note it.
		if result.Status == domain.CheckStatusPass {
			result.Status = domain.CheckStatusWarning
		}
	}

	// Check read permission
	entries, err := c.fs.ReadDir(ctx, c.targetDir)
	if err != nil {
		if os.IsPermission(err) {
			result.Status = domain.CheckStatusFail
			result.Issues = append(result.Issues, domain.Issue{
				Code:     "TARGET_DIR_NOT_READABLE",
				Message:  fmt.Sprintf("Cannot read target directory: %s", c.targetDir),
				Severity: domain.IssueSeverityError,
				Path:     c.targetDir,
				Context: map[string]any{
					"error": err.Error(),
				},
				Remediation: &domain.Remediation{
					Description: fmt.Sprintf("Ensure you have read permissions: chmod u+r %s", c.targetDir),
				},
			})
			return result, nil
		}
		return result, fmt.Errorf("failed to read target directory: %w", err)
	}

	result.Stats["target_dir"] = c.targetDir
	result.Stats["entries_count"] = len(entries)

	return result, nil
}
