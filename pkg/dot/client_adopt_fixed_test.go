package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestClient_Adopt_FileNotFound(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/misc", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Try to adopt non-existent file
	err = client.Adopt(ctx, []string{".nonexistent"}, "misc")
	require.Error(t, err)
}

func TestClient_Adopt_PackageNotFound(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/target/.file", []byte("content"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt to non-existent package - should create the package directory
	err = client.Adopt(ctx, []string{".file"}, "newpackage")
	require.NoError(t, err)

	// Verify package directory was created
	exists := fs.Exists(ctx, "/test/packages/newpackage")
	assert.True(t, exists, "Package directory should be created")

	// Verify file was moved
	exists = fs.Exists(ctx, "/test/packages/newpackage/dot-file")
	assert.True(t, exists, "File should be moved to package")

	// Verify symlink was created
	isLink, err := fs.IsSymlink(ctx, "/test/target/.file")
	require.NoError(t, err)
	assert.True(t, isLink, "Should create symlink in target")
}

func TestClient_PlanAdopt_Success(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/misc", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/target/.config", []byte("cfg"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan adopt
	plan, err := client.PlanAdopt(ctx, []string{".config"}, "misc")
	require.NoError(t, err)

	assert.NotEmpty(t, plan.Operations)
	assert.True(t, len(plan.Operations) >= 2, "Expected move and link operations")
}
