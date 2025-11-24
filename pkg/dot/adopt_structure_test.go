package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestAdopt_DirectoryStructure(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Create directories
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.ssh", 0755))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/config", []byte("ssh config"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/known_hosts", []byte("hosts"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt .ssh directory (package name should be dot-ssh with new behavior)
	err = client.Adopt(ctx, []string{".ssh"}, "dot-ssh")
	require.NoError(t, err)

	// Check where the directory was stored (NEW FLAT STRUCTURE)
	t.Log("Checking package structure after adopt...")

	// NEW: Package is named "dot-ssh" and files are at root
	pkgDir := packageDir + "/dot-ssh"
	if fs.Exists(ctx, pkgDir) {
		entries, _ := fs.ReadDir(ctx, pkgDir)
		t.Logf("Package /dot-ssh contains %d entries:", len(entries))
		for _, e := range entries {
			t.Logf("  - %s (isDir: %v)", e.Name(), e.IsDir())
		}
	}

	// NEW: Files should be at package root
	assert.True(t, fs.Exists(ctx, pkgDir), "Package directory dot-ssh should exist")
	assert.True(t, fs.Exists(ctx, pkgDir+"/config"), "config should be at package root")
	assert.True(t, fs.Exists(ctx, pkgDir+"/known_hosts"), "known_hosts should be at package root")

	// Verify original was replaced with symlink
	isLink, _ := fs.IsSymlink(ctx, targetDir+"/.ssh")
	assert.True(t, isLink, "Target should be a symlink")

	// Symlink should point to package root
	linkTarget, _ := fs.ReadLink(ctx, targetDir+"/.ssh")
	assert.Contains(t, linkTarget, "/dot-ssh")
	assert.NotContains(t, linkTarget, "/dot-ssh/dot-ssh", "Should not have nested structure")
}
