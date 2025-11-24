package scanner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/scanner"
)

func TestScanPackage(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	packagePath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	ignoreSet := ignore.NewIgnoreSet()

	// Mock: package directory exists and is empty
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/vim").Return(true)
	mockFS.On("IsSymlink", ctx, "/home/user/.dotfiles/vim").Return(false, nil)
	mockFS.On("IsDir", ctx, "/home/user/.dotfiles/vim").Return(true, nil)
	mockFS.On("ReadDir", ctx, "/home/user/.dotfiles/vim").Return([]domain.DirEntry{}, nil)

	result := scanner.ScanPackage(ctx, mockFS, packagePath, "vim", ignoreSet)
	require.True(t, result.IsOk())

	pkg := result.Unwrap()
	assert.Equal(t, "vim", pkg.Name)
	assert.Equal(t, packagePath, pkg.Path)
	require.NotNil(t, pkg.Tree, "Tree should be populated")
	assert.Equal(t, domain.NodeDir, pkg.Tree.Type)

	mockFS.AssertExpectations(t)
}

func TestScanPackage_PackageNotFound(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	packagePath := domain.NewPackagePath("/home/user/.dotfiles/missing").Unwrap()
	ignoreSet := ignore.NewIgnoreSet()

	// Mock: package directory does not exist
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/missing").Return(false)

	result := scanner.ScanPackage(ctx, mockFS, packagePath, "missing", ignoreSet)
	assert.True(t, result.IsErr())

	// Should return ErrPackageNotFound
	err := result.UnwrapErr()
	_, ok := err.(domain.ErrPackageNotFound)
	assert.True(t, ok, "Expected ErrPackageNotFound")

	mockFS.AssertExpectations(t)
}

func TestScanPackage_WithIgnorePatterns(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	packagePath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	ignoreSet := ignore.NewIgnoreSet()
	ignoreSet.Add(".git")

	// Mock: package exists and is a directory
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/vim").Return(true)
	mockFS.On("IsSymlink", ctx, "/home/user/.dotfiles/vim").Return(false, nil)
	mockFS.On("IsDir", ctx, "/home/user/.dotfiles/vim").Return(true, nil)
	mockFS.On("ReadDir", ctx, "/home/user/.dotfiles/vim").Return([]domain.DirEntry{}, nil)

	result := scanner.ScanPackage(ctx, mockFS, packagePath, "vim", ignoreSet)
	require.True(t, result.IsOk())

	pkg := result.Unwrap()
	assert.Equal(t, "vim", pkg.Name)
	require.NotNil(t, pkg.Tree, "Tree should be scanned")

	// Tree filtering is applied during scan
	// With empty directory, tree has no children (nothing to filter)
	assert.Empty(t, pkg.Tree.Children)

	mockFS.AssertExpectations(t)
}

func TestScanPackageWithConfig_WithDotignore(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory with .dotignore
	packagePath := "/test/package"
	require.NoError(t, fs.Mkdir(ctx, packagePath, 0755))

	// Create .dotignore file
	dotignorePath := packagePath + "/.dotignore"
	dotignoreContent := []byte("*.log\n*.tmp\n")
	require.NoError(t, fs.WriteFile(ctx, dotignorePath, dotignoreContent, 0644))

	// Create some files
	require.NoError(t, fs.WriteFile(ctx, packagePath+"/config.txt", []byte("config"), 0644))
	require.NoError(t, fs.WriteFile(ctx, packagePath+"/debug.log", []byte("logs"), 0644))
	require.NoError(t, fs.WriteFile(ctx, packagePath+"/temp.tmp", []byte("temp"), 0644))

	globalIgnoreSet := ignore.NewIgnoreSet()
	cfg := scanner.ScanConfig{
		PerPackageIgnore: true,
		Interactive:      false,
	}

	pkgPath := domain.NewPackagePath(packagePath).Unwrap()
	result := scanner.ScanPackageWithConfig(ctx, fs, pkgPath, "testpkg", globalIgnoreSet, cfg)

	require.True(t, result.IsOk(), "scan should succeed")
	pkg := result.Unwrap()

	assert.Equal(t, "testpkg", pkg.Name)
	assert.NotNil(t, pkg.Tree)

	// Verify .log and .tmp files are filtered out
	hasLogFile := false
	hasTmpFile := false
	hasConfigFile := false
	for _, child := range pkg.Tree.Children {
		name := child.Path.String()
		if name == packagePath+"/debug.log" {
			hasLogFile = true
		}
		if name == packagePath+"/temp.tmp" {
			hasTmpFile = true
		}
		if name == packagePath+"/config.txt" {
			hasConfigFile = true
		}
	}

	assert.False(t, hasLogFile, ".log files should be ignored")
	assert.False(t, hasTmpFile, ".tmp files should be ignored")
	assert.True(t, hasConfigFile, ".txt files should not be ignored")
}

func TestScanPackageWithConfig_WithMaxFileSize(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory
	packagePath := "/test/package"
	require.NoError(t, fs.Mkdir(ctx, packagePath, 0755))

	// Create small and large files
	smallFile := packagePath + "/small.txt"
	largeFile := packagePath + "/large.txt"
	require.NoError(t, fs.WriteFile(ctx, smallFile, []byte("small"), 0644))
	require.NoError(t, fs.WriteFile(ctx, largeFile, make([]byte, 2048), 0644)) // 2KB file

	globalIgnoreSet := ignore.NewIgnoreSet()
	cfg := scanner.ScanConfig{
		MaxFileSize:      1024, // 1KB limit
		Interactive:      false,
		PerPackageIgnore: false,
	}

	pkgPath := domain.NewPackagePath(packagePath).Unwrap()
	result := scanner.ScanPackageWithConfig(ctx, fs, pkgPath, "testpkg", globalIgnoreSet, cfg)

	require.True(t, result.IsOk(), "scan should succeed")
	pkg := result.Unwrap()

	// In batch mode, large files should be excluded
	hasSmallFile := false
	hasLargeFile := false
	for _, child := range pkg.Tree.Children {
		if child.Path.String() == smallFile {
			hasSmallFile = true
		}
		if child.Path.String() == largeFile {
			hasLargeFile = true
		}
	}

	assert.True(t, hasSmallFile, "small file should be included")
	assert.False(t, hasLargeFile, "large file should be excluded in batch mode")
}

func TestScanPackageWithConfig_InteractivePrompter(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory with a large file
	packagePath := "/test/package"
	require.NoError(t, fs.Mkdir(ctx, packagePath, 0755))
	require.NoError(t, fs.WriteFile(ctx, packagePath+"/large.bin", make([]byte, 2048), 0644))

	globalIgnoreSet := ignore.NewIgnoreSet()
	cfg := scanner.ScanConfig{
		MaxFileSize:      1024,
		Interactive:      true,
		PerPackageIgnore: false,
	}

	pkgPath := domain.NewPackagePath(packagePath).Unwrap()
	result := scanner.ScanPackageWithConfig(ctx, fs, pkgPath, "testpkg", globalIgnoreSet, cfg)

	// Result should succeed (prompter will be used if running in TTY)
	require.True(t, result.IsOk(), "scan should succeed")
}

func TestScanPackageWithConfig_BatchPrompter(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory
	packagePath := "/test/package"
	require.NoError(t, fs.Mkdir(ctx, packagePath, 0755))
	require.NoError(t, fs.WriteFile(ctx, packagePath+"/file.txt", []byte("data"), 0644))

	globalIgnoreSet := ignore.NewIgnoreSet()
	cfg := scanner.ScanConfig{
		MaxFileSize:      1024,
		Interactive:      false, // Batch mode
		PerPackageIgnore: false,
	}

	pkgPath := domain.NewPackagePath(packagePath).Unwrap()
	result := scanner.ScanPackageWithConfig(ctx, fs, pkgPath, "testpkg", globalIgnoreSet, cfg)

	require.True(t, result.IsOk(), "scan should succeed in batch mode")
	pkg := result.Unwrap()
	assert.Equal(t, "testpkg", pkg.Name)
}

func TestScanPackageWithConfig_InvalidDotignore(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory with invalid .dotignore
	packagePath := "/test/package"
	require.NoError(t, fs.Mkdir(ctx, packagePath, 0755))

	// Create .dotignore with invalid pattern
	dotignorePath := packagePath + "/.dotignore"
	dotignoreContent := []byte("*.log\n!!invalid\n") // Double !! is invalid
	require.NoError(t, fs.WriteFile(ctx, dotignorePath, dotignoreContent, 0644))

	globalIgnoreSet := ignore.NewIgnoreSet()
	cfg := scanner.ScanConfig{
		PerPackageIgnore: true,
		Interactive:      false,
	}

	pkgPath := domain.NewPackagePath(packagePath).Unwrap()
	result := scanner.ScanPackageWithConfig(ctx, fs, pkgPath, "testpkg", globalIgnoreSet, cfg)

	// Should return error due to invalid pattern
	assert.True(t, result.IsErr(), "should fail with invalid .dotignore")
	err := result.UnwrapErr()
	assert.Contains(t, err.Error(), "load .dotignore")
}

func TestFilterTree_EmptyTree(t *testing.T) {
	ignoreSet := ignore.NewIgnoreSet()

	// Empty node (no path)
	emptyNode := domain.Node{}

	result := scanner.FilterTreeForTest(emptyNode, ignoreSet)

	// Should return empty node
	assert.Equal(t, "", result.Path.String())
}

func TestFilterTree_IgnoreRoot(t *testing.T) {
	ignoreSet := ignore.NewIgnoreSet()
	ignoreSet.Add("/root/*")

	rootPath := domain.NewFilePath("/root/file.txt").Unwrap()
	node := domain.Node{
		Path: rootPath,
		Type: domain.NodeFile,
	}

	result := scanner.FilterTreeForTest(node, ignoreSet)

	// File should be ignored (returned as empty node)
	assert.Equal(t, "", result.Path.String())
}

func TestFilterTree_MixedIgnore(t *testing.T) {
	ignoreSet := ignore.NewIgnoreSet()
	ignoreSet.Add("*.log")

	dirPath := domain.NewFilePath("/test").Unwrap()
	file1Path := domain.NewFilePath("/test/keep.txt").Unwrap()
	file2Path := domain.NewFilePath("/test/ignore.log").Unwrap()
	file3Path := domain.NewFilePath("/test/data.json").Unwrap()

	node := domain.Node{
		Path: dirPath,
		Type: domain.NodeDir,
		Children: []domain.Node{
			{Path: file1Path, Type: domain.NodeFile},
			{Path: file2Path, Type: domain.NodeFile},
			{Path: file3Path, Type: domain.NodeFile},
		},
	}

	result := scanner.FilterTreeForTest(node, ignoreSet)

	// Directory should remain, but .log file should be filtered out
	assert.Equal(t, dirPath.String(), result.Path.String())
	assert.Equal(t, 2, len(result.Children), "should have 2 children (log file filtered)")

	// Verify kept files
	keptPaths := make(map[string]bool)
	for _, child := range result.Children {
		keptPaths[child.Path.String()] = true
	}

	assert.True(t, keptPaths["/test/keep.txt"])
	assert.True(t, keptPaths["/test/data.json"])
	assert.False(t, keptPaths["/test/ignore.log"])
}
