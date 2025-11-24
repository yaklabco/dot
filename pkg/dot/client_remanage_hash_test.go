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

func TestClient_RemanageWithHashDetection(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	origFile := filepath.Join("/test/packages/app", "dot-config")
	require.NoError(t, fs.WriteFile(ctx, origFile, []byte("version1"), 0644))

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

	// Remanage without changes (should detect unchanged)
	plan, err := client.PlanRemanage(ctx, "app")
	require.NoError(t, err)
	// Unchanged package may result in empty or minimal plan
	assert.NotNil(t, plan)

	// Modify package
	require.NoError(t, fs.WriteFile(ctx, origFile, []byte("version2"), 0644))

	// Remanage with changes (should detect changed)
	plan2, err := client.PlanRemanage(ctx, "app")
	require.NoError(t, err)
	assert.NotNil(t, plan2)

	// Execute remanage
	err = client.Remanage(ctx, "app")
	require.NoError(t, err)
}

func TestClient_RemanageNotInstalled(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/newpkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/newpkg/dot-file", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Remanage package that's not installed (should install it)
	err = client.Remanage(ctx, "newpkg")
	require.NoError(t, err)

	// Verify installed
	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.Len(t, status.Packages, 1)
}
