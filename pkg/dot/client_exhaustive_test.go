package dot_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// Exhaustive tests to push coverage well above 80% threshold

func TestClient_Manage_WithMultipleFiles(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup package with multiple files
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/multi", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/multi/dot-file1", []byte("1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/multi/dot-file2", []byte("2"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/multi/dot-file3", []byte("3"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package with multiple files
	err = client.Manage(ctx, "multi")
	require.NoError(t, err)

	// Verify all links created
	for i := 1; i <= 3; i++ {
		linkPath := filepath.Join("/test/target", ".file"+string(rune('0'+i)))
		isLink, _ := fs.IsSymlink(ctx, linkPath)
		assert.True(t, isLink, "Expected link for file%d", i)
	}
}

func TestClient_Unmanage_WithEmptyPlan(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage non-existent package (empty plan path)
	err = client.Unmanage(ctx, "notinstalled")
	require.NoError(t, err) // Should succeed with empty plan
}

func TestClient_Remanage_NoChangesDetected(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/stable", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/stable/dot-file", []byte("unchanged"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Initial manage
	err = client.Manage(ctx, "stable")
	require.NoError(t, err)

	// Remanage without changes
	err = client.Remanage(ctx, "stable")
	require.NoError(t, err) // Should detect no changes and succeed
}

func TestClient_StatusWithNonExistentFilter(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-file", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage one package
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Request status for non-existent package
	status, err := client.Status(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, status.Packages, "Non-existent package should return empty")

	// Request status for mix of existent and non-existent
	status, err = client.Status(ctx, "app", "nonexistent", "alsononexistent")
	require.NoError(t, err)
	assert.Len(t, status.Packages, 1, "Should only return existing package")
}

func TestClient_PlanManage_MultiplePackages(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	pkgs := []string{"a", "b", "c"}
	for _, pkg := range pkgs {
		pkgDir := filepath.Join("/test/packages", pkg)
		require.NoError(t, fs.MkdirAll(ctx, pkgDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, pkgDir+"/dot-file", []byte("x"), 0644))
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

	// Plan for multiple packages
	plan, err := client.PlanManage(ctx, pkgs...)
	require.NoError(t, err)

	assert.NotEmpty(t, plan.Operations)
	assert.Equal(t, 3, plan.Metadata.PackageCount)

	// Verify PackageOperations is populated
	assert.NotNil(t, plan.PackageOperations, "Expected PackageOperations to be initialized")
	assert.True(t, len(plan.PackageOperations) > 0, "Expected at least one package in PackageOperations")
}

func TestClient_PlanUnmanage_MultiplePackages(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	pkgs := []string{"x", "y"}
	for _, pkg := range pkgs {
		pkgDir := filepath.Join("/test/packages", pkg)
		require.NoError(t, fs.MkdirAll(ctx, pkgDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, pkgDir+"/dot-file", []byte("x"), 0644))
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

	// Manage both
	err = client.Manage(ctx, pkgs...)
	require.NoError(t, err)

	// Plan unmanage both
	plan, err := client.PlanUnmanage(ctx, pkgs...)
	require.NoError(t, err)

	assert.NotEmpty(t, plan.Operations)
	assert.Equal(t, 2, plan.Metadata.PackageCount)
}

func TestClient_ManageWithExecutionSuccess(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/tool", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/tool/dot-rc", []byte("config"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage and verify execution success
	err = client.Manage(ctx, "tool")
	require.NoError(t, err)

	// Verify manifest was updated (tests updateManifest path)
	status, err := client.Status(ctx)
	require.NoError(t, err)
	require.Len(t, status.Packages, 1)
	assert.True(t, status.Packages[0].LinkCount > 0)
	assert.NotEmpty(t, status.Packages[0].Links)
}
