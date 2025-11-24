package manifest

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/yaklabco/dot/internal/domain"
)

// Validator checks manifest consistency with filesystem
type Validator struct {
	fs domain.FS
}

// NewValidator creates a new manifest validator
func NewValidator(fs domain.FS) *Validator {
	return &Validator{fs: fs}
}

// ValidationResult contains validation outcome and issues
type ValidationResult struct {
	IsValid bool
	Issues  []ValidationIssue
}

// ValidationIssue describes a specific problem found
type ValidationIssue struct {
	Type        IssueType
	Path        string
	Package     string
	Description string
}

// IssueType categorizes validation problems
type IssueType int

const (
	IssueBrokenLink IssueType = iota
	IssueMissingLink
	IssueExtraLink
	IssueWrongTarget
	IssueNotSymlink
)

// Validate checks manifest consistency with filesystem
func (v *Validator) Validate(ctx context.Context, targetDir domain.TargetPath, manifest Manifest) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Issues:  []ValidationIssue{},
	}

	// Check all links in manifest exist and are valid
	for _, pkg := range manifest.Packages {
		for _, linkPath := range pkg.Links {
			if ctx.Err() != nil {
				break
			}

			// Reject absolute paths - links should be relative to target directory
			if filepath.IsAbs(linkPath) {
				result.IsValid = false
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        IssueMissingLink,
					Path:        linkPath,
					Package:     pkg.Name,
					Description: "Link path must be relative, not absolute",
				})
				continue
			}

			fullPath := filepath.Join(targetDir.String(), linkPath)
			issue := v.validateLink(ctx, fullPath, linkPath, pkg.Name)
			if issue != nil {
				result.IsValid = false
				result.Issues = append(result.Issues, *issue)
			}
		}
	}

	return result
}

// validateLink checks if a specific link is valid
// fullPath is the absolute path to the link (targetDir + linkPath)
// linkPath is the relative path from manifest (for display purposes)
func (v *Validator) validateLink(ctx context.Context, fullPath, linkPath, pkgName string) *ValidationIssue {
	// Check if link exists
	exists := v.fs.Exists(ctx, fullPath)
	if !exists {
		return &ValidationIssue{
			Type:        IssueMissingLink,
			Path:        linkPath,
			Package:     pkgName,
			Description: "Link specified in manifest does not exist",
		}
	}

	// Check if it's a symlink
	isSymlink, err := v.fs.IsSymlink(ctx, fullPath)
	if err != nil {
		return &ValidationIssue{
			Type:        IssueBrokenLink,
			Path:        linkPath,
			Package:     pkgName,
			Description: fmt.Sprintf("Cannot check if path is symlink: %v", err),
		}
	}
	if !isSymlink {
		return &ValidationIssue{
			Type:        IssueNotSymlink,
			Path:        linkPath,
			Package:     pkgName,
			Description: "Path is not a symlink",
		}
	}

	// Check if link target exists
	target, err := v.fs.ReadLink(ctx, fullPath)
	if err != nil {
		return &ValidationIssue{
			Type:        IssueBrokenLink,
			Path:        linkPath,
			Package:     pkgName,
			Description: fmt.Sprintf("Cannot read link: %v", err),
		}
	}

	// Resolve relative targets against the symlink's directory
	resolvedTarget := target
	if !filepath.IsAbs(target) {
		// Relative target - resolve against symlink's directory
		linkDir := filepath.Dir(fullPath)
		resolvedTarget = filepath.Clean(filepath.Join(linkDir, target))
	}

	// Check if resolved target exists
	targetExists := v.fs.Exists(ctx, resolvedTarget)
	if !targetExists {
		return &ValidationIssue{
			Type:        IssueBrokenLink,
			Path:        linkPath,
			Package:     pkgName,
			Description: fmt.Sprintf("Link target does not exist: %s", target),
		}
	}

	return nil
}
