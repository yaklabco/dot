package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// Final tests to push coverage above 80%

func TestClient_ConfigAccess(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
		DryRun:     true,
		BackupDir:  "/custom/backup",
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Test Config() returns correct values
	resultCfg := client.Config()
	require.Equal(t, "/test/packages", resultCfg.PackageDir)
	require.Equal(t, "/test/target", resultCfg.TargetDir)
	require.True(t, resultCfg.DryRun)
	require.Equal(t, "/custom/backup", resultCfg.BackupDir)
}

func TestClient_DryRunBehavior(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/dry", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/dry/dot-file", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
		DryRun:     true,
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// All operations should succeed but not modify filesystem
	err = client.Manage(ctx, "dry")
	require.NoError(t, err)

	err = client.Remanage(ctx, "dry")
	require.NoError(t, err)

	err = client.Unmanage(ctx, "dry")
	require.NoError(t, err)

	// No files should exist
	exists := fs.Exists(ctx, "/test/target/.file")
	require.False(t, exists, "Dry-run should not create files")
}

func TestClient_PlanOperationsEmpty(t *testing.T) {
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

	// Plan operations on non-existent package
	plan, err := client.PlanUnmanage(ctx, "notinstalled")
	require.NoError(t, err)
	require.Empty(t, plan.Operations)

	// Remanage on non-existent package should error
	// (fallback to manage still errors if package doesn't exist in filesystem)
	plan, err = client.PlanRemanage(ctx, "notinstalled")
	require.Error(t, err)
}
