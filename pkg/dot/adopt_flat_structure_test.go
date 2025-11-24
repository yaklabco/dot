package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// TestAdopt_Directory_FlatStructure tests the new flat package structure.
// When adopting .ssh, files should be at dot-ssh/ root, not dot-ssh/dot-ssh/
func TestAdopt_Directory_FlatStructure(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Create .ssh directory with files
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.ssh", 0755))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/config", []byte("ssh config"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/known_hosts", []byte("hosts"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/.hidden", []byte("hidden"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt .ssh directory
	err = client.Adopt(ctx, []string{".ssh"}, "dot-ssh")
	require.NoError(t, err)

	// NEW EXPECTATION: Flat structure
	// Files should be at package root: dot-ssh/config (not dot-ssh/dot-ssh/config)
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/config"), "config should be at package root")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/known_hosts"), "known_hosts should be at package root")

	// Dotfiles inside should get dot- prefix
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/dot-hidden"), ".hidden should become dot-hidden")

	// OLD STRUCTURE SHOULD NOT EXIST
	assert.False(t, fs.Exists(ctx, packageDir+"/dot-ssh/dot-ssh"), "Should NOT have nested dot-ssh/dot-ssh/")

	// Symlink should point to package root
	linkTarget, err := fs.ReadLink(ctx, targetDir+"/.ssh")
	require.NoError(t, err)
	assert.Contains(t, linkTarget, "/dot-ssh")
	assert.NotContains(t, linkTarget, "/dot-ssh/dot-ssh", "Symlink should point to package root, not nested dir")

	// Verify contents
	data, _ := fs.ReadFile(ctx, packageDir+"/dot-ssh/config")
	assert.Equal(t, []byte("ssh config"), data)
}

// TestAdopt_Directory_WithNestedDotfiles tests dotfile translation for nested items
func TestAdopt_Directory_WithNestedDotfiles(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Create directory with nested dotfiles
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.config", 0755))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.config/settings.json", []byte("{}"), 0644))
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.config/.cache", 0755))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.config/.cache/data", []byte("cache"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt .config directory
	err = client.Adopt(ctx, []string{".config"}, "dot-config")
	require.NoError(t, err)

	// Flat structure at package root
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-config/settings.json"), "Regular file at root")

	// Nested dotfile/directory gets dot- prefix
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-config/dot-cache"), ".cache → dot-cache")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-config/dot-cache/data"), "Files inside nested dir")
}

// TestAdopt_File_KeepsCurrentBehavior tests that single file adoption is unchanged
func TestAdopt_File_KeepsCurrentBehavior(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.vimrc", []byte("vim config"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt single file
	err = client.Adopt(ctx, []string{".vimrc"}, "dot-vimrc")
	require.NoError(t, err)

	// Single files are stored in package directory with translation
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-vimrc/dot-vimrc"))

	// Verify symlink
	linkTarget, _ := fs.ReadLink(ctx, targetDir+"/.vimrc")
	assert.Contains(t, linkTarget, "/dot-vimrc/dot-vimrc")
}
