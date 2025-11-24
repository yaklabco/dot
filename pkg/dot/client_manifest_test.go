package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestClient_ManifestTracking(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/tool", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/tool/dot-config", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage should create manifest
	err = client.Manage(ctx, "tool")
	require.NoError(t, err)

	// Status should read from manifest
	status, err := client.Status(ctx)
	require.NoError(t, err)
	require.Len(t, status.Packages, 1)
	assert.Equal(t, "tool", status.Packages[0].Name)
	assert.True(t, status.Packages[0].LinkCount > 0)
	assert.NotEmpty(t, status.Packages[0].Links)

	// Unmanage should update manifest
	err = client.Unmanage(ctx, "tool")
	require.NoError(t, err)

	// Status should show no packages
	status, err = client.Status(ctx)
	require.NoError(t, err)
	assert.Empty(t, status.Packages)
}

func TestClient_LinkCounting(t *testing.T) {
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

	// Manage
	err = client.Manage(ctx, "multi")
	require.NoError(t, err)

	// Check link count
	status, err := client.Status(ctx, "multi")
	require.NoError(t, err)
	require.Len(t, status.Packages, 1)
	assert.Equal(t, 3, status.Packages[0].LinkCount, "Expected 3 links")
}
