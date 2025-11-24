package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// Additional tests for edge cases and helper functions to increase coverage buffer

func TestClient_Adopt_DryRun(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/misc", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/target/.adoptme", []byte("content"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
		DryRun:     true,
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt in dry-run mode
	err = client.Adopt(ctx, []string{".adoptme"}, "misc")
	require.NoError(t, err)

	// File should NOT be moved in dry-run
	exists := fs.Exists(ctx, "/test/target/.adoptme")
	assert.True(t, exists, "File should remain in dry-run")

	// Should NOT be a symlink
	isLink, _ := fs.IsSymlink(ctx, "/test/target/.adoptme")
	assert.False(t, isLink, "Should not be symlink in dry-run")
}

func TestClient_Adopt_EmptyFilesList(t *testing.T) {
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

	// Adopt with empty files list
	err = client.Adopt(ctx, []string{}, "misc")
	require.NoError(t, err) // Should succeed with no operations
}

func TestClient_UnmanageMultiple(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup multiple packages
	for _, pkg := range []string{"vim", "zsh", "git"} {
		pkgDir := "/test/packages/" + pkg
		require.NoError(t, fs.MkdirAll(ctx, pkgDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, pkgDir+"/dot-"+pkg+"rc", []byte("cfg"), 0644))
	}
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage all
	err = client.Manage(ctx, "vim", "zsh", "git")
	require.NoError(t, err)

	// Unmanage multiple at once
	err = client.Unmanage(ctx, "vim", "git")
	require.NoError(t, err)

	// Verify only zsh remains
	status, err := client.Status(ctx)
	require.NoError(t, err)
	require.Len(t, status.Packages, 1)
	assert.Equal(t, "zsh", status.Packages[0].Name)
}

func TestClient_RemanageMultipleWithMixedChanges(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app1", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app2", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app1/dot-file1", []byte("v1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app2/dot-file2", []byte("v1"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Initial manage
	err = client.Manage(ctx, "app1", "app2")
	require.NoError(t, err)

	// Modify only app1
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app1/dot-file1", []byte("v2"), 0644))

	// Remanage both
	err = client.Remanage(ctx, "app1", "app2")
	require.NoError(t, err)

	// Both should still be installed
	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.Len(t, status.Packages, 2)
}

func TestClient_PlanRemanageEmptySlice(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage first
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Plan remanage with no package list (empty slice)
	plan, err := client.PlanRemanage(ctx)
	require.NoError(t, err)
	assert.NotNil(t, plan)
}
