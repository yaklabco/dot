package adopt

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/jamesainslie/dot/internal/domain"
)

// AppendToGlobalDotignore appends a pattern to the global .dotignore file.
// Creates the file if it doesn't exist. Returns error if operation fails.
func AppendToGlobalDotignore(ctx context.Context, fs domain.FS, configDir, pattern string) error {
	// Validate inputs
	if configDir == "" {
		return fmt.Errorf("config directory cannot be empty")
	}
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	// Construct path to global .dotignore
	dotignorePath := filepath.Join(configDir, ".dotignore")

	// Ensure config directory exists
	if err := fs.MkdirAll(ctx, configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Check if file exists
	exists := fs.Exists(ctx, dotignorePath)

	// Read existing content
	var existingContent []byte
	if exists {
		content, err := fs.ReadFile(ctx, dotignorePath)
		if err != nil {
			return fmt.Errorf("read .dotignore: %w", err)
		}
		existingContent = content
	}

	// Build new content with timestamp comment and pattern
	timestamp := time.Now().Format("2006-01-02")
	entry := fmt.Sprintf("# Added by dot adopt on %s\n%s\n", timestamp, pattern)

	// Append to existing content
	newContent := existingContent
	if len(existingContent) > 0 && existingContent[len(existingContent)-1] != '\n' {
		newContent = append(newContent, '\n')
	}
	newContent = append(newContent, []byte(entry)...)

	// Write back to file with secure permissions (0644)
	if err := fs.WriteFile(ctx, dotignorePath, newContent, 0644); err != nil {
		return fmt.Errorf("write .dotignore: %w", err)
	}

	return nil
}
