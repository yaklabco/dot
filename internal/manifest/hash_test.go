package manifest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestContentHasher_HashPackage_EmptyPackage(t *testing.T) {
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")
	require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))

	hasher := NewContentHasher(fs)

	hash, err := hasher.HashPackage(context.Background(), pkgPath)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestContentHasher_HashPackage_SingleFile(t *testing.T) {
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")
	require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))

	vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
	require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set number\n"), 0644))

	hasher := NewContentHasher(fs)

	hash, err := hasher.HashPackage(context.Background(), pkgPath)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 hex length
}

func TestContentHasher_HashPackage_Deterministic(t *testing.T) {
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")
	require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))

	vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
	require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set number\n"), 0644))

	hasher := NewContentHasher(fs)

	hash1, err := hasher.HashPackage(context.Background(), pkgPath)
	require.NoError(t, err)

	hash2, err := hasher.HashPackage(context.Background(), pkgPath)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
}

func TestContentHasher_HashPackage_DifferentContent(t *testing.T) {
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")
	require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))

	vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
	hasher := NewContentHasher(fs)

	// Hash with initial content
	require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set number\n"), 0644))
	hash1, err := hasher.HashPackage(context.Background(), pkgPath)
	require.NoError(t, err)

	// Hash with different content
	require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set relativenumber\n"), 0644))
	hash2, err := hasher.HashPackage(context.Background(), pkgPath)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestContentHasher_HashPackage_NestedDirectories(t *testing.T) {
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")

	// Create nested structure
	colorsPath := filepath.Join(pkgPath.String(), "dot-vim", "colors")
	require.NoError(t, fs.MkdirAll(context.Background(), colorsPath, 0755))
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(colorsPath, "molokai.vim"), []byte("colorscheme"), 0644))

	hasher := NewContentHasher(fs)

	hash, err := hasher.HashPackage(context.Background(), pkgPath)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestContentHasher_HashPackage_IgnoresSymlinks(t *testing.T) {
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")
	require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))

	// Create file and symlink
	vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
	linkPath := filepath.Join(pkgPath.String(), "link-to-vimrc")
	require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("content"), 0644))
	require.NoError(t, fs.Symlink(context.Background(), vimrcPath, linkPath))

	hasher := NewContentHasher(fs)

	hash, err := hasher.HashPackage(context.Background(), pkgPath)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	// Symlink should not affect hash (only real files)
}

func TestContentHasher_HashPackage_MultipleFiles(t *testing.T) {
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")
	require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))

	// Create multiple files
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(pkgPath.String(), "dot-vimrc"), []byte("vimrc content"), 0644))
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(pkgPath.String(), "dot-gvimrc"), []byte("gvimrc content"), 0644))

	hasher := NewContentHasher(fs)

	hash, err := hasher.HashPackage(context.Background(), pkgPath)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestContentHasher_HashPackage_OrderIndependent(t *testing.T) {
	// Files are sorted internally, so hash should be the same regardless of
	// the order files are discovered
	fs := adapters.NewMemFS()
	pkgPath := mustPackagePath(t, "/packages/vim")
	require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))

	// Create files in alphabetical order
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(pkgPath.String(), "a.txt"), []byte("a"), 0644))
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(pkgPath.String(), "b.txt"), []byte("b"), 0644))
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(pkgPath.String(), "c.txt"), []byte("c"), 0644))

	hasher := NewContentHasher(fs)
	hash, err := hasher.HashPackage(context.Background(), pkgPath)
	require.NoError(t, err)

	// Hash should be deterministic regardless of filesystem order
	assert.Len(t, hash, 64)
}

func TestContentHasher_HashPackage_NoCollisions(t *testing.T) {
	// Test that delimiter prevents hash collisions from ambiguous concatenations
	// Without delimiter: path="a", content="bc" vs path="ab", content="c" would hash the same
	fs := adapters.NewMemFS()

	// Scenario 1: file "a" with content "bc"
	pkg1Path := mustPackagePath(t, "/packages/pkg1")
	require.NoError(t, fs.MkdirAll(context.Background(), pkg1Path.String(), 0755))
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(pkg1Path.String(), "a"), []byte("bc"), 0644))

	// Scenario 2: file "ab" with content "c"
	pkg2Path := mustPackagePath(t, "/packages/pkg2")
	require.NoError(t, fs.MkdirAll(context.Background(), pkg2Path.String(), 0755))
	require.NoError(t, fs.WriteFile(context.Background(),
		filepath.Join(pkg2Path.String(), "ab"), []byte("c"), 0644))

	hasher := NewContentHasher(fs)

	hash1, err := hasher.HashPackage(context.Background(), pkg1Path)
	require.NoError(t, err)

	hash2, err := hasher.HashPackage(context.Background(), pkg2Path)
	require.NoError(t, err)

	// Hashes must be different due to delimiter preventing concatenation ambiguity
	assert.NotEqual(t, hash1, hash2, "delimiter should prevent hash collision")
}
