package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestAdopt_Directory_MovesContents(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Create package dir and target dir
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.testdir", 0755))

	// Create files INSIDE the directory
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.testdir/file1.txt", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.testdir/file2.txt", []byte("content2"), 0644))

	// Verify files exist before adopt
	assert.True(t, fs.Exists(ctx, targetDir+"/.testdir/file1.txt"))
	assert.True(t, fs.Exists(ctx, targetDir+"/.testdir/file2.txt"))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt the directory (package name gets dot- prefix)
	err = client.Adopt(ctx, []string{".testdir"}, "dot-testdir")
	require.NoError(t, err)

	// Check the package directory structure (NEW FLAT STRUCTURE)
	t.Log("After adopt, checking package directory...")
	pkgDir := packageDir + "/dot-testdir" // Changed: flat structure, no nested dir
	assert.True(t, fs.Exists(ctx, pkgDir), "Package directory should exist")

	if fs.Exists(ctx, pkgDir) {
		entries, _ := fs.ReadDir(ctx, pkgDir)
		t.Logf("Package directory contains %d entries:", len(entries))
		for _, e := range entries {
			t.Logf("  - %s", e.Name())
		}
	}

	// FILES SHOULD BE AT PACKAGE ROOT (FLAT STRUCTURE)
	assert.True(t, fs.Exists(ctx, pkgDir+"/file1.txt"), "file1.txt should be at package root")
	assert.True(t, fs.Exists(ctx, pkgDir+"/file2.txt"), "file2.txt should be at package root")

	// Verify contents
	data, err := fs.ReadFile(ctx, pkgDir+"/file1.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("content1"), data)
}
