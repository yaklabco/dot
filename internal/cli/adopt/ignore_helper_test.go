package adopt

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendToGlobalDotignore_CreatesFile(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	configDir := "/home/user/.config/dot"

	// Ensure config directory exists
	require.NoError(t, fs.MkdirAll(ctx, configDir, 0755))

	ignorePath := filepath.Join(configDir, ".dotignore")
	pattern := ".cache"

	err := AppendToGlobalDotignore(ctx, fs, configDir, pattern)
	require.NoError(t, err)

	// Verify file was created
	_, err = fs.Stat(ctx, ignorePath)
	require.NoError(t, err, "ignore file should exist")

	// Verify content
	content, err := fs.ReadFile(ctx, ignorePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), pattern)
}

func TestAppendToGlobalDotignore_AppendsToExisting(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	configDir := "/home/user/.config/dot"

	// Create config directory and existing .dotignore
	require.NoError(t, fs.MkdirAll(ctx, configDir, 0755))
	ignorePath := filepath.Join(configDir, ".dotignore")
	require.NoError(t, fs.WriteFile(ctx, ignorePath, []byte("existing-pattern\n"), 0644))

	pattern := "new-pattern"

	err := AppendToGlobalDotignore(ctx, fs, configDir, pattern)
	require.NoError(t, err)

	// Verify content includes both patterns
	content, err := fs.ReadFile(ctx, ignorePath)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	assert.Contains(t, lines, "existing-pattern")
	assert.Contains(t, lines, "new-pattern")
}

func TestAppendToGlobalDotignore_AddsNewline(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	configDir := "/home/user/.config/dot"

	require.NoError(t, fs.MkdirAll(ctx, configDir, 0755))

	pattern := "test-pattern"

	err := AppendToGlobalDotignore(ctx, fs, configDir, pattern)
	require.NoError(t, err)

	ignorePath := filepath.Join(configDir, ".dotignore")
	content, err := fs.ReadFile(ctx, ignorePath)
	require.NoError(t, err)

	// Should end with newline
	assert.True(t, strings.HasSuffix(string(content), "\n"), "Content should end with newline")
}

func TestAppendToGlobalDotignore_HandlesNoTrailingNewline(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	configDir := "/home/user/.config/dot"

	// Create file without trailing newline
	require.NoError(t, fs.MkdirAll(ctx, configDir, 0755))
	ignorePath := filepath.Join(configDir, ".dotignore")
	require.NoError(t, fs.WriteFile(ctx, ignorePath, []byte("existing-pattern"), 0644))

	pattern := "new-pattern"

	err := AppendToGlobalDotignore(ctx, fs, configDir, pattern)
	require.NoError(t, err)

	content, err := fs.ReadFile(ctx, ignorePath)
	require.NoError(t, err)

	// Should have existing pattern, comment, and new pattern
	assert.Contains(t, string(content), "existing-pattern")
	assert.Contains(t, string(content), "# Added by dot adopt on")
	assert.Contains(t, string(content), "new-pattern")
}

func TestAppendToGlobalDotignore_EmptyInputs(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Test empty config dir
	err := AppendToGlobalDotignore(ctx, fs, "", "pattern")
	assert.Error(t, err)

	// Test empty pattern
	err = AppendToGlobalDotignore(ctx, fs, "/config", "")
	assert.Error(t, err)
}

func TestAppendToGlobalDotignore_MultiplePatterns(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	configDir := "/home/user/.config/dot"

	require.NoError(t, fs.MkdirAll(ctx, configDir, 0755))

	patterns := []string{
		".cache",
		"node_modules",
		"*.log",
	}

	for _, pattern := range patterns {
		err := AppendToGlobalDotignore(ctx, fs, configDir, pattern)
		require.NoError(t, err)
	}

	ignorePath := filepath.Join(configDir, ".dotignore")
	content, err := fs.ReadFile(ctx, ignorePath)
	require.NoError(t, err)

	for _, pattern := range patterns {
		assert.Contains(t, string(content), pattern)
	}
}

func TestAppendToGlobalDotignore_WithCommentedPattern(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	configDir := "/home/user/.config/dot"

	require.NoError(t, fs.MkdirAll(ctx, configDir, 0755))

	// Create file with some comments
	ignorePath := filepath.Join(configDir, ".dotignore")
	initialContent := `# Ignore cache directories
.cache
# Ignore logs
*.log
`
	require.NoError(t, fs.WriteFile(ctx, ignorePath, []byte(initialContent), 0644))

	pattern := "new-pattern"

	err := AppendToGlobalDotignore(ctx, fs, configDir, pattern)
	require.NoError(t, err)

	content, err := fs.ReadFile(ctx, ignorePath)
	require.NoError(t, err)

	// Should preserve comments and add new pattern
	assert.Contains(t, string(content), "# Ignore cache directories")
	assert.Contains(t, string(content), "# Ignore logs")
	assert.Contains(t, string(content), pattern)
}
