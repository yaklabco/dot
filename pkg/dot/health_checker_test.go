package dot

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthChecker_CheckLink tests the unified health checking logic.
func TestHealthChecker_CheckLink(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := "/home"
	packageDir := "/packages/config"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	checker := newHealthChecker(fs, targetDir)

	t.Run("healthy link", func(t *testing.T) {
		// Create source file
		sourceFile := filepath.Join(packageDir, "dot-bashrc")
		require.NoError(t, fs.WriteFile(ctx, sourceFile, []byte("content"), 0644))

		// Create symlink
		linkPath := ".bashrc"
		fullLink := filepath.Join(targetDir, linkPath)
		require.NoError(t, fs.Symlink(ctx, sourceFile, fullLink))

		result := checker.CheckLink(ctx, "config", linkPath, packageDir)
		assert.True(t, result.IsHealthy)
		assert.Empty(t, result.IssueType)
	})

	t.Run("missing link", func(t *testing.T) {
		result := checker.CheckLink(ctx, "config", ".missing", packageDir)
		assert.False(t, result.IsHealthy)
		assert.Equal(t, IssueBrokenLink, result.IssueType)
		assert.Contains(t, result.Message, "does not exist")
	})

	t.Run("broken link target", func(t *testing.T) {
		linkPath := ".broken"
		fullLink := filepath.Join(targetDir, linkPath)
		require.NoError(t, fs.Symlink(ctx, "/nonexistent", fullLink))

		result := checker.CheckLink(ctx, "config", linkPath, packageDir)
		assert.False(t, result.IsHealthy)
		assert.Equal(t, IssueBrokenLink, result.IssueType)
		assert.Contains(t, result.Message, "target does not exist")
	})

	t.Run("regular file instead of symlink", func(t *testing.T) {
		linkPath := ".regularfile"
		fullLink := filepath.Join(targetDir, linkPath)
		require.NoError(t, fs.WriteFile(ctx, fullLink, []byte("content"), 0644))

		result := checker.CheckLink(ctx, "config", linkPath, packageDir)
		assert.False(t, result.IsHealthy)
		assert.Equal(t, IssueWrongTarget, result.IssueType)
		assert.Contains(t, result.Message, "Expected symlink")
	})

	t.Run("target outside package directory", func(t *testing.T) {
		outsideFile := "/other/location/file"
		require.NoError(t, fs.MkdirAll(ctx, filepath.Dir(outsideFile), 0755))
		require.NoError(t, fs.WriteFile(ctx, outsideFile, []byte("content"), 0644))

		linkPath := ".outside"
		fullLink := filepath.Join(targetDir, linkPath)
		require.NoError(t, fs.Symlink(ctx, outsideFile, fullLink))

		result := checker.CheckLink(ctx, "config", linkPath, packageDir)
		assert.False(t, result.IsHealthy)
		assert.Equal(t, IssueWrongTarget, result.IssueType)
		assert.Contains(t, result.Message, "outside package directory")
	})
}

// TestHealthChecker_CheckPackage tests package-level health checking.
func TestHealthChecker_CheckPackage(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := "/home"
	packageDir := "/packages/config"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	checker := newHealthChecker(fs, targetDir)

	t.Run("all links healthy", func(t *testing.T) {
		// Create source files
		for _, name := range []string{"dot-bashrc", "dot-vimrc"} {
			sourceFile := filepath.Join(packageDir, name)
			require.NoError(t, fs.WriteFile(ctx, sourceFile, []byte("content"), 0644))
		}

		// Create symlinks
		for _, link := range []string{".bashrc", ".vimrc"} {
			fullLink := filepath.Join(targetDir, link)
			sourceFile := filepath.Join(packageDir, "dot-"+link[1:])
			require.NoError(t, fs.Symlink(ctx, sourceFile, fullLink))
		}

		healthy, issueType := checker.CheckPackage(ctx, "config", []string{".bashrc", ".vimrc"}, packageDir)
		assert.True(t, healthy)
		assert.Empty(t, issueType)
	})

	t.Run("some missing links", func(t *testing.T) {
		links := []string{".bashrc", ".missing"}
		healthy, issueType := checker.CheckPackage(ctx, "config", links, packageDir)
		assert.False(t, healthy)
		assert.Equal(t, "missing links", issueType)
	})

	t.Run("some broken link targets", func(t *testing.T) {
		// Create a broken symlink (target doesn't exist)
		brokenLink := filepath.Join(targetDir, ".broken2")
		require.NoError(t, fs.Symlink(ctx, "/nonexistent", brokenLink))

		links := []string{".bashrc", ".broken2"}
		healthy, issueType := checker.CheckPackage(ctx, "config", links, packageDir)
		assert.False(t, healthy)
		assert.Equal(t, "broken links", issueType)
	})

	t.Run("wrong target", func(t *testing.T) {
		// Create a regular file
		regularFile := filepath.Join(targetDir, ".regular")
		require.NoError(t, fs.WriteFile(ctx, regularFile, []byte("content"), 0644))

		links := []string{".regular"}
		healthy, issueType := checker.CheckPackage(ctx, "config", links, packageDir)
		assert.False(t, healthy)
		assert.Equal(t, "wrong target", issueType)
	})
}

// TestHealthChecker_Consistency verifies list and doctor use same logic.
func TestHealthChecker_Consistency(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := "/home"
	packageDir := "/packages/config"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	// Create a symlink pointing outside package directory
	outsideFile := "/other/file"
	require.NoError(t, fs.MkdirAll(ctx, filepath.Dir(outsideFile), 0755))
	require.NoError(t, fs.WriteFile(ctx, outsideFile, []byte("content"), 0644))

	linkPath := ".outside"
	fullLink := filepath.Join(targetDir, linkPath)
	require.NoError(t, fs.Symlink(ctx, outsideFile, fullLink))

	checker := newHealthChecker(fs, targetDir)

	// Both CheckLink and CheckPackage should report unhealthy
	linkResult := checker.CheckLink(ctx, "config", linkPath, packageDir)
	assert.False(t, linkResult.IsHealthy, "CheckLink should report unhealthy")

	packageHealthy, _ := checker.CheckPackage(ctx, "config", []string{linkPath}, packageDir)
	assert.False(t, packageHealthy, "CheckPackage should report unhealthy")
}

// TestHealthChecker_CheckPackage_PriorityOrder tests issue type prioritization.
func TestHealthChecker_CheckPackage_PriorityOrder(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := "/home"
	packageDir := "/packages/config"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	// Create a broken symlink (highest priority)
	brokenLink := filepath.Join(targetDir, ".broken")
	require.NoError(t, fs.Symlink(ctx, "/nonexistent", brokenLink))

	// Create a regular file (wrong target)
	regularFile := filepath.Join(targetDir, ".regular")
	require.NoError(t, fs.WriteFile(ctx, regularFile, []byte("content"), 0644))

	t.Run("broken links take priority", func(t *testing.T) {
		// When both broken links and wrong targets exist, broken links should be reported
		checker := newHealthChecker(fs, targetDir)
		healthy, issueType := checker.CheckPackage(ctx, "config", []string{".broken", ".regular"}, packageDir)
		assert.False(t, healthy)
		assert.Equal(t, "broken links", issueType)
	})

	t.Run("wrong target reported when no broken links", func(t *testing.T) {
		checker := newHealthChecker(fs, targetDir)
		healthy, issueType := checker.CheckPackage(ctx, "config", []string{".regular"}, packageDir)
		assert.False(t, healthy)
		assert.Equal(t, "wrong target", issueType)
	})

	t.Run("missing links reported when only missing", func(t *testing.T) {
		checker := newHealthChecker(fs, targetDir)
		healthy, issueType := checker.CheckPackage(ctx, "config", []string{".missing1", ".missing2"}, packageDir)
		assert.False(t, healthy)
		assert.Equal(t, "missing links", issueType)
	})
}

// TestHealthChecker_CheckLink_EmptyPackageDir tests behavior when packageDir is empty.
func TestHealthChecker_CheckLink_EmptyPackageDir(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := "/home"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create a file outside any package directory
	outsideFile := "/other/file"
	require.NoError(t, fs.MkdirAll(ctx, filepath.Dir(outsideFile), 0755))
	require.NoError(t, fs.WriteFile(ctx, outsideFile, []byte("content"), 0644))

	// Create symlink
	linkPath := ".config"
	fullLink := filepath.Join(targetDir, linkPath)
	require.NoError(t, fs.Symlink(ctx, outsideFile, fullLink))

	checker := newHealthChecker(fs, targetDir)

	// With empty packageDir, should skip target location validation
	result := checker.CheckLink(ctx, "legacy", linkPath, "")
	assert.True(t, result.IsHealthy, "Should be healthy when packageDir is empty")
}

// TestHealthChecker_CheckLink_RelativeTarget tests relative symlink resolution.
func TestHealthChecker_CheckLink_RelativeTarget(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := "/home"
	packageDir := "/packages/config"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	// Create source file
	sourceFile := filepath.Join(packageDir, "dot-bashrc")
	require.NoError(t, fs.WriteFile(ctx, sourceFile, []byte("content"), 0644))

	// Create relative symlink
	linkPath := ".bashrc"
	fullLink := filepath.Join(targetDir, linkPath)
	relativeTarget := "../../packages/config/dot-bashrc"
	require.NoError(t, fs.Symlink(ctx, relativeTarget, fullLink))

	checker := newHealthChecker(fs, targetDir)
	result := checker.CheckLink(ctx, "config", linkPath, packageDir)

	assert.True(t, result.IsHealthy)
	assert.Empty(t, result.IssueType)
}
