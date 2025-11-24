package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestClient_ScanConfigurations(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Test different scan modes
	t.Run("DefaultScan", func(t *testing.T) {
		report, err := client.DoctorWithScan(ctx, dot.DefaultScanConfig())
		require.NoError(t, err)
		assert.Equal(t, 0, report.Statistics.OrphanedLinks)
	})

	t.Run("ScopedScan", func(t *testing.T) {
		report, err := client.DoctorWithScan(ctx, dot.ScopedScanConfig())
		require.NoError(t, err)
		assert.NotNil(t, report)
	})

	t.Run("DeepScan", func(t *testing.T) {
		scanCfg := dot.DeepScanConfig(2)
		report, err := client.DoctorWithScan(ctx, scanCfg)
		require.NoError(t, err)
		assert.NotNil(t, report)
	})
}

func TestClient_StatusFiltering(t *testing.T) {
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

	// Test filtering
	t.Run("AllPackages", func(t *testing.T) {
		status, err := client.Status(ctx)
		require.NoError(t, err)
		assert.Len(t, status.Packages, 3)
	})

	t.Run("SinglePackage", func(t *testing.T) {
		status, err := client.Status(ctx, "vim")
		require.NoError(t, err)
		assert.Len(t, status.Packages, 1)
		assert.Equal(t, "vim", status.Packages[0].Name)
	})

	t.Run("MultiplePackages", func(t *testing.T) {
		status, err := client.Status(ctx, "vim", "zsh")
		require.NoError(t, err)
		assert.Len(t, status.Packages, 2)
	})

	t.Run("NonExistentPackage", func(t *testing.T) {
		status, err := client.Status(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, status.Packages)
	})
}
