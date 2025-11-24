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

// Targeted tests to boost coverage of low-coverage functions

func TestClient_Adopt_MultipleFiles(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/adopt", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))

	// Create multiple files in target to adopt
	files := []string{".file1", ".file2", ".file3"}
	for _, f := range files {
		require.NoError(t, fs.WriteFile(ctx, filepath.Join("/test/target", f), []byte("content"), 0644))
	}

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan adopt multiple files
	plan, err := client.PlanAdopt(ctx, files, "adopt")
	require.NoError(t, err)

	// Should have 6 operations (3 moves + 3 links)
	assert.True(t, len(plan.Operations) >= 6, "Expected move and link for each file")
}

func TestClient_Unmanage_MultipleWithManifestUpdate(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app1", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app2", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app1/dot-file1", []byte("1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app2/dot-file2", []byte("2"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage both
	err = client.Manage(ctx, "app1", "app2")
	require.NoError(t, err)

	// Verify both links exist
	link1 := fs.Exists(ctx, "/test/target/.file1")
	link2 := fs.Exists(ctx, "/test/target/.file2")
	require.True(t, link1 && link2)

	// Unmanage both
	err = client.Unmanage(ctx, "app1", "app2")
	require.NoError(t, err)

	// Verify both removed
	exists1 := fs.Exists(ctx, "/test/target/.file1")
	exists2 := fs.Exists(ctx, "/test/target/.file2")
	assert.False(t, exists1 || exists2)

	// Verify manifest empty
	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.Empty(t, status.Packages)
}

func TestClient_List_AfterOperations(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup multiple packages
	pkgs := []string{"vim", "zsh", "git"}
	for _, pkg := range pkgs {
		pkgDir := filepath.Join("/test/packages", pkg)
		require.NoError(t, fs.MkdirAll(ctx, pkgDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, pkgDir+"/dot-"+pkg+"rc", []byte("x"), 0644))
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
	err = client.Manage(ctx, pkgs...)
	require.NoError(t, err)

	// List should show all
	packages, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, packages, 3)

	// Unmanage one
	err = client.Unmanage(ctx, "git")
	require.NoError(t, err)

	// List should show 2
	packages, err = client.List(ctx)
	require.NoError(t, err)
	require.Len(t, packages, 2)
}

func TestClient_PlanUnmanage_AllEdgeCases(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/test", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/test/dot-file", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan unmanage before install
	plan1, err := client.PlanUnmanage(ctx, "test")
	require.NoError(t, err)
	assert.Empty(t, plan1.Operations, "Should be empty before install")

	// Manage
	err = client.Manage(ctx, "test")
	require.NoError(t, err)

	// Plan unmanage after install
	plan2, err := client.PlanUnmanage(ctx, "test")
	require.NoError(t, err)
	assert.NotEmpty(t, plan2.Operations, "Should have delete operations")

	// Verify plan has correct metadata
	assert.Equal(t, 1, plan2.Metadata.PackageCount)
	assert.True(t, plan2.Metadata.OperationCount > 0)
}

func TestClient_Remanage_SucceedsOnContentChange(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("v1"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Initial manage
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Modify package content
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("v2"), 0644))

	// Remanage should detect changes and succeed
	err = client.Remanage(ctx, "app")
	require.NoError(t, err)
}

func TestClient_ManageWithManifestUpdate(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/full", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/full/dot-a", []byte("a"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/full/dot-b", []byte("b"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "full")
	require.NoError(t, err)

	// Verify manifest has correct link count
	status, err := client.Status(ctx, "full")
	require.NoError(t, err)
	require.Len(t, status.Packages, 1)
	assert.Equal(t, 2, status.Packages[0].LinkCount, "Expected 2 links tracked")
}
