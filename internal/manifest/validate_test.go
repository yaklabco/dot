package manifest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestValidator_Validate_EmptyManifest(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	m := New()
	validator := NewValidator(fs)

	result := validator.Validate(context.Background(), targetDir, m)

	assert.True(t, result.IsValid)
	assert.Empty(t, result.Issues)
}

func TestValidator_Validate_ValidManifest(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")

	// Create source file
	vimrcSrc := "/packages/vim/dot-vimrc"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(vimrcSrc), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), vimrcSrc, []byte("content"), 0644))

	// Create target link
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
	require.NoError(t, fs.Symlink(context.Background(), vimrcSrc, vimrcTarget))

	// Create manifest
	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.True(t, result.IsValid)
	assert.Empty(t, result.Issues)
}

func TestValidator_Validate_BrokenLink(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")

	// Create broken symlink
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
	require.NoError(t, fs.Symlink(context.Background(), "/nonexistent", vimrcTarget))

	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, IssueBrokenLink, result.Issues[0].Type)
	assert.Equal(t, "vim", result.Issues[0].Package)
}

func TestValidator_Validate_MissingLink(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, IssueMissingLink, result.Issues[0].Type)
	assert.Contains(t, result.Issues[0].Path, ".vimrc")
}

func TestValidator_Validate_NotSymlink(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")

	// Create regular file where symlink should be
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), vimrcTarget, []byte("content"), 0644))

	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, IssueNotSymlink, result.Issues[0].Type)
}

func TestValidator_Validate_MultiplePackages(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create vim package and link
	vimrcSrc := "/packages/vim/dot-vimrc"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(vimrcSrc), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), vimrcSrc, []byte("vim"), 0644))
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.Symlink(context.Background(), vimrcSrc, vimrcTarget))

	// Create zsh package and link
	zshrcSrc := "/packages/zsh/dot-zshrc"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(zshrcSrc), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), zshrcSrc, []byte("zsh"), 0644))
	zshrcTarget := filepath.Join(targetDir.String(), ".zshrc")
	require.NoError(t, fs.Symlink(context.Background(), zshrcSrc, zshrcTarget))

	m := New()
	m.AddPackage(PackageInfo{Name: "vim", Links: []string{".vimrc"}})
	m.AddPackage(PackageInfo{Name: "zsh", Links: []string{".zshrc"}})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.True(t, result.IsValid)
	assert.Empty(t, result.Issues)
}

func TestValidator_Validate_MultipleIssues(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create source file and valid link
	sourcePath := "/some/path"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(sourcePath), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), sourcePath, []byte("content"), 0644))
	validTarget := filepath.Join(targetDir.String(), ".valid")
	require.NoError(t, fs.Symlink(context.Background(), sourcePath, validTarget))

	m := New()
	m.AddPackage(PackageInfo{
		Name:  "pkg",
		Links: []string{".valid", ".missing", ".broken"},
	})

	// Create broken link
	brokenTarget := filepath.Join(targetDir.String(), ".broken")
	require.NoError(t, fs.Symlink(context.Background(), "/nonexistent", brokenTarget))

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	assert.GreaterOrEqual(t, len(result.Issues), 2) // At least missing and broken
}

func TestValidator_Validate_RejectsAbsoluteLinkPath(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create manifest with absolute path in links (security issue)
	m := New()
	m.AddPackage(PackageInfo{
		Name:  "malicious",
		Links: []string{"/etc/passwd"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Contains(t, result.Issues[0].Description, "absolute")
	assert.Equal(t, "/etc/passwd", result.Issues[0].Path)
}

func TestValidator_Validate_RelativeSymlinkTarget(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create symlink with relative target
	srcDir := filepath.Join(targetDir.String(), "src")
	srcFile := filepath.Join(srcDir, "file.txt")
	require.NoError(t, fs.MkdirAll(context.Background(), srcDir, 0755))
	require.NoError(t, fs.WriteFile(context.Background(), srcFile, []byte("content"), 0644))

	linkPath := filepath.Join(targetDir.String(), ".config", "link")
	linkDir := filepath.Dir(linkPath)
	require.NoError(t, fs.MkdirAll(context.Background(), linkDir, 0755))

	// Create relative symlink from .config/link -> ../src/file.txt
	require.NoError(t, fs.Symlink(context.Background(), "../src/file.txt", linkPath))

	m := New()
	m.AddPackage(PackageInfo{
		Name:  "pkg",
		Links: []string{".config/link"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.True(t, result.IsValid, "relative symlink target should be resolved correctly")
	assert.Empty(t, result.Issues)
}

func TestValidator_Validate_BrokenRelativeSymlink(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create broken symlink with relative target
	linkPath := filepath.Join(targetDir.String(), ".config", "link")
	linkDir := filepath.Dir(linkPath)
	require.NoError(t, fs.MkdirAll(context.Background(), linkDir, 0755))

	// Create relative symlink to nonexistent file
	require.NoError(t, fs.Symlink(context.Background(), "../nonexistent", linkPath))

	m := New()
	m.AddPackage(PackageInfo{
		Name:  "pkg",
		Links: []string{".config/link"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, IssueBrokenLink, result.Issues[0].Type)
}
