package dot_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// testConfig returns a standard test configuration
func testConfig(t *testing.T) dot.Config {
	t.Helper()
	return dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
	}
}

// setupTestFixtures creates test packages with sample files
func setupTestFixtures(t *testing.T, fs dot.FS, packages ...string) {
	t.Helper()
	ctx := context.Background()

	// Create package directory structure
	for _, pkg := range packages {
		pkgDir := filepath.Join("/test/packages", pkg)
		require.NoError(t, fs.MkdirAll(ctx, pkgDir, 0755))

		// Create sample dotfile
		dotfile := filepath.Join(pkgDir, "dot-config")
		require.NoError(t, fs.WriteFile(ctx, dotfile, []byte("test content"), 0644))
	}

	// Create target directory
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
}
