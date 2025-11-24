package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestUnmanage_AdoptedPackage_Restore(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup directories
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	// Create a file to adopt
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.config", []byte("my config"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt the file (package gets dot- prefix)
	err = client.Adopt(ctx, []string{".config"}, "dot-config")
	require.NoError(t, err)

	// Verify file was moved to package
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-config/dot-config"))
	// Verify symlink was created
	isLink, _ := fs.IsSymlink(ctx, targetDir+"/.config")
	assert.True(t, isLink)

	// Unmanage (should restore the file)
	err = client.Unmanage(ctx, "dot-config")
	require.NoError(t, err)

	// Verify symlink was removed
	exists := fs.Exists(ctx, targetDir+"/.config")
	assert.True(t, exists, "File should be restored to target")

	// Verify it's NOT a symlink anymore
	isLink, _ = fs.IsSymlink(ctx, targetDir+"/.config")
	assert.False(t, isLink, "Should be a regular file, not symlink")

	// Verify content is preserved
	data, err := fs.ReadFile(ctx, targetDir+"/.config")
	require.NoError(t, err)
	assert.Equal(t, []byte("my config"), data)

	// Verify file ALSO remains in package directory (copied, not moved)
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-config/dot-config"), "File should remain in package directory")
	data, err = fs.ReadFile(ctx, packageDir+"/dot-config/dot-config")
	require.NoError(t, err)
	assert.Equal(t, []byte("my config"), data, "Package file should have same content")
}

func TestUnmanage_AdoptedPackage_WithPurge(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup directories
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	// Create a file to adopt
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.vimrc", []byte("vim config"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt the file
	err = client.Adopt(ctx, []string{".vimrc"}, "vim")
	require.NoError(t, err)

	// Unmanage with purge option
	opts := dot.UnmanageOptions{
		Purge:   true,
		Restore: false,
	}
	err = client.UnmanageWithOptions(ctx, opts, "vim")
	require.NoError(t, err)

	// Verify symlink was removed
	exists := fs.Exists(ctx, targetDir+"/.vimrc")
	assert.False(t, exists, "Symlink should be removed")

	// Verify package directory was deleted
	exists = fs.Exists(ctx, packageDir+"/vim")
	assert.False(t, exists, "Package directory should be purged")
}

func TestUnmanage_AdoptedPackage_NoRestore(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup directories
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	// Create a file to adopt
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.bashrc", []byte("bash config"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt the file
	err = client.Adopt(ctx, []string{".bashrc"}, "bash")
	require.NoError(t, err)

	// Unmanage without restore
	opts := dot.UnmanageOptions{
		Purge:   false,
		Restore: false, // Don't restore
	}
	err = client.UnmanageWithOptions(ctx, opts, "bash")
	require.NoError(t, err)

	// Verify symlink was removed
	exists := fs.Exists(ctx, targetDir+"/.bashrc")
	assert.False(t, exists, "Symlink should be removed")

	// Verify file stays in package directory
	exists = fs.Exists(ctx, packageDir+"/bash/dot-bashrc")
	assert.True(t, exists, "File should remain in package directory")
}

func TestUnmanage_ManagedPackage_NoRestore(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup directories and package
	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/git", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/git/dot-gitconfig", []byte("[user]"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage the package
	err = client.Manage(ctx, "git")
	require.NoError(t, err)

	// Verify symlink was created
	isLink, _ := fs.IsSymlink(ctx, targetDir+"/.gitconfig")
	assert.True(t, isLink)

	// Unmanage managed package (should just remove symlink)
	err = client.Unmanage(ctx, "git")
	require.NoError(t, err)

	// Verify symlink was removed
	exists := fs.Exists(ctx, targetDir+"/.gitconfig")
	assert.False(t, exists)

	// Verify package directory still exists
	exists = fs.Exists(ctx, packageDir+"/git/dot-gitconfig")
	assert.True(t, exists, "Managed package files should remain")
}

func TestUnmanageOptions_Defaults(t *testing.T) {
	opts := dot.DefaultUnmanageOptions()

	assert.False(t, opts.Purge, "Purge should be false by default")
	assert.True(t, opts.Restore, "Restore should be true by default")
}

func TestUnmanage_AdoptedDirectory_Restore(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup directories
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.ssh", 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
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

	// Adopt the entire .ssh directory (package gets dot- prefix)
	err = client.Adopt(ctx, []string{".ssh"}, "dot-ssh")
	require.NoError(t, err)

	// Verify directory was moved to package (flat structure)
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh"))

	// Debug: Check what's in the package BEFORE unmanage
	t.Log("Before unmanage, checking package directory...")
	if fs.Exists(ctx, packageDir+"/dot-ssh") {
		entries, _ := fs.ReadDir(ctx, packageDir+"/dot-ssh")
		t.Logf("Package /dot-ssh contains %d entries:", len(entries))
		for _, e := range entries {
			t.Logf("  - %s", e.Name())
		}
		// Check if config file exists at root
		configExists := fs.Exists(ctx, packageDir+"/dot-ssh/config")
		t.Logf("config file exists: %v", configExists)
	}

	// Verify symlink was created
	isLink, _ := fs.IsSymlink(ctx, targetDir+"/.ssh")
	assert.True(t, isLink)

	// Unmanage (should restore the entire directory)
	err = client.Unmanage(ctx, "dot-ssh")
	require.NoError(t, err)

	// Debug: Check package directory after unmanage
	t.Log("After unmanage, checking package directory...")
	if fs.Exists(ctx, packageDir+"/dot-ssh") {
		entries, _ := fs.ReadDir(ctx, packageDir+"/dot-ssh")
		t.Logf("Package /dot-ssh contains %d entries:", len(entries))
		for _, e := range entries {
			t.Logf("  - %s (isDir: %v)", e.Name(), e.IsDir())
		}
	} else {
		t.Log("Package /dot-ssh directory does NOT exist")
	}

	// Verify .ssh directory was restored
	assert.True(t, fs.Exists(ctx, targetDir+"/.ssh"))

	// Verify it's a real directory, not a symlink
	isLink, _ = fs.IsSymlink(ctx, targetDir+"/.ssh")
	assert.False(t, isLink)

	// Verify all files are back
	assert.True(t, fs.Exists(ctx, targetDir+"/.ssh/config"))
	assert.True(t, fs.Exists(ctx, targetDir+"/.ssh/known_hosts"))

	// Verify contents
	data, _ := fs.ReadFile(ctx, targetDir+"/.ssh/config")
	assert.Equal(t, []byte("ssh config"), data)

	// Verify directory ALSO remains in package directory (copied, not moved)
	t.Log("Verifying package directory preserved...")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh"), "Directory should remain in package directory")
	if fs.Exists(ctx, packageDir+"/dot-ssh") {
		assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/config"), "Files should remain at package root")

		// Verify package directory has same content
		data, _ = fs.ReadFile(ctx, packageDir+"/dot-ssh/config")
		assert.Equal(t, []byte("ssh config"), data, "Package files should have same content")
	}
}
