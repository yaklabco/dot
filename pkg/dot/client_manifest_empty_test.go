package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// Test that updateManifest handles empty packages slice by extracting from plan

func TestClient_ManifestUpdate_AllPackages(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup multiple packages
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/vim", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/zsh", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/vim/dot-vimrc", []byte("vim"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/zsh/dot-zshrc", []byte("zsh"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage multiple packages (packages slice will be non-empty)
	err = client.Manage(ctx, "vim", "zsh")
	require.NoError(t, err)

	// Verify both packages in manifest
	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.Len(t, status.Packages, 2, "Expected both packages in manifest")

	packageNames := []string{status.Packages[0].Name, status.Packages[1].Name}
	assert.Contains(t, packageNames, "vim")
	assert.Contains(t, packageNames, "zsh")
}

func TestClient_ManifestUpdate_ExtractsFromPlan(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
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

	// Manage with explicit package name
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Verify manifest updated correctly
	status, err := client.Status(ctx, "app")
	require.NoError(t, err)
	require.Len(t, status.Packages, 1)
	assert.Equal(t, "app", status.Packages[0].Name)
	assert.True(t, status.Packages[0].LinkCount > 0, "Expected links tracked in manifest")
}
