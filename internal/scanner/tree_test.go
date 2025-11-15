package scanner_test

import (
	"context"
	"io/fs"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockFS implements the FS interface for testing scanner logic.
type MockFS struct {
	mock.Mock
}

func (m *MockFS) Stat(ctx context.Context, name string) (domain.FileInfo, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.FileInfo), args.Error(1)
}

func (m *MockFS) Lstat(ctx context.Context, name string) (domain.FileInfo, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.FileInfo), args.Error(1)
}

func (m *MockFS) ReadDir(ctx context.Context, name string) ([]domain.DirEntry, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DirEntry), args.Error(1)
}

func (m *MockFS) ReadLink(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func (m *MockFS) ReadFile(ctx context.Context, name string) ([]byte, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFS) WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error {
	args := m.Called(ctx, name, data, perm)
	return args.Error(0)
}

func (m *MockFS) Mkdir(ctx context.Context, name string, perm fs.FileMode) error {
	args := m.Called(ctx, name, perm)
	return args.Error(0)
}

func (m *MockFS) MkdirAll(ctx context.Context, name string, perm fs.FileMode) error {
	args := m.Called(ctx, name, perm)
	return args.Error(0)
}

func (m *MockFS) Remove(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockFS) RemoveAll(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockFS) Symlink(ctx context.Context, oldname, newname string) error {
	args := m.Called(ctx, oldname, newname)
	return args.Error(0)
}

func (m *MockFS) Rename(ctx context.Context, oldname, newname string) error {
	args := m.Called(ctx, oldname, newname)
	return args.Error(0)
}

func (m *MockFS) Exists(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)
	return args.Bool(0)
}

func (m *MockFS) IsDir(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockFS) IsSymlink(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func TestScanTree_SingleFile(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/file.txt").Unwrap()

	// Mock: path is not a symlink, and is a file (not a directory)
	mockFS.On("IsSymlink", ctx, "/test/file.txt").Return(false, nil)
	mockFS.On("IsDir", ctx, "/test/file.txt").Return(false, nil)

	result := scanner.ScanTree(ctx, mockFS, path)
	require.True(t, result.IsOk())

	node := result.Unwrap()
	assert.Equal(t, path, node.Path)
	assert.Equal(t, domain.NodeFile, node.Type)
	assert.Nil(t, node.Children)

	mockFS.AssertExpectations(t)
}

func TestScanTree_EmptyDirectory(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/dir").Unwrap()

	// Mock: path is not a symlink, is a directory with no children
	mockFS.On("IsSymlink", ctx, "/test/dir").Return(false, nil)
	mockFS.On("IsDir", ctx, "/test/dir").Return(true, nil)
	mockFS.On("ReadDir", ctx, "/test/dir").Return([]domain.DirEntry{}, nil)

	result := scanner.ScanTree(ctx, mockFS, path)
	require.True(t, result.IsOk())

	node := result.Unwrap()
	assert.Equal(t, path, node.Path)
	assert.Equal(t, domain.NodeDir, node.Type)
	assert.Empty(t, node.Children)

	mockFS.AssertExpectations(t)
}

func TestScanTree_Symlink(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/link").Unwrap()

	// Mock: path is a symlink
	mockFS.On("IsSymlink", ctx, "/test/link").Return(true, nil)

	result := scanner.ScanTree(ctx, mockFS, path)
	require.True(t, result.IsOk())

	node := result.Unwrap()
	assert.Equal(t, path, node.Path)
	assert.Equal(t, domain.NodeSymlink, node.Type)
	assert.Nil(t, node.Children)

	mockFS.AssertExpectations(t)
}

func TestScanTree_Error(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/error").Unwrap()

	// Mock: IsSymlink returns an error
	mockFS.On("IsSymlink", ctx, "/test/error").Return(false, assert.AnError)

	result := scanner.ScanTree(ctx, mockFS, path)
	assert.True(t, result.IsErr())

	mockFS.AssertExpectations(t)
}

func TestWalk(t *testing.T) {
	// Build a simple tree: dir -> file1, file2
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1").Unwrap(),
				Type: domain.NodeFile,
			},
			{
				Path: domain.NewFilePath("/test/file2").Unwrap(),
				Type: domain.NodeFile,
			},
		},
	}

	// Collect all visited paths
	var visited []string
	err := scanner.Walk(root, func(n domain.Node) error {
		visited = append(visited, n.Path.String())
		return nil
	})

	require.NoError(t, err)
	assert.Len(t, visited, 3) // root + 2 children
	assert.Contains(t, visited, "/test")
	assert.Contains(t, visited, "/test/file1")
	assert.Contains(t, visited, "/test/file2")
}

func TestWalk_ErrorStopsTraversal(t *testing.T) {
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1").Unwrap(),
				Type: domain.NodeFile,
			},
		},
	}

	// Return error on first visit
	err := scanner.Walk(root, func(n domain.Node) error {
		return assert.AnError
	})

	assert.Error(t, err)
}

func TestCollectFiles(t *testing.T) {
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1.txt").Unwrap(),
				Type: domain.NodeFile,
			},
			{
				Path: domain.NewFilePath("/test/subdir").Unwrap(),
				Type: domain.NodeDir,
				Children: []domain.Node{
					{
						Path: domain.NewFilePath("/test/subdir/file2.txt").Unwrap(),
						Type: domain.NodeFile,
					},
				},
			},
		},
	}

	files := scanner.CollectFiles(root)
	assert.Len(t, files, 2)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.String()
	}

	assert.Contains(t, paths, "/test/file1.txt")
	assert.Contains(t, paths, "/test/subdir/file2.txt")
}

func TestCountNodes(t *testing.T) {
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1").Unwrap(),
				Type: domain.NodeFile,
			},
			{
				Path: domain.NewFilePath("/test/dir").Unwrap(),
				Type: domain.NodeDir,
				Children: []domain.Node{
					{
						Path: domain.NewFilePath("/test/dir/file2").Unwrap(),
						Type: domain.NodeFile,
					},
				},
			},
		},
	}

	count := scanner.CountNodes(root)
	assert.Equal(t, 4, count) // root + file1 + dir + file2
}

func TestRelativePath(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		target   string
		expected string
		wantErr  bool
	}{
		{
			name:     "same directory",
			base:     "/home/user/.dotfiles",
			target:   "/home/user/.dotfiles/file.txt",
			expected: "file.txt",
			wantErr:  false,
		},
		{
			name:     "nested directory",
			base:     "/home/user/.dotfiles",
			target:   "/home/user/.dotfiles/vim/vimrc",
			expected: "vim/vimrc",
			wantErr:  false,
		},
		{
			name:     "same path",
			base:     "/home/user/.dotfiles",
			target:   "/home/user/.dotfiles",
			expected: ".",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := domain.NewFilePath(tt.base).Unwrap()
			target := domain.NewFilePath(tt.target).Unwrap()

			result := scanner.RelativePath(base, target)

			if tt.wantErr {
				assert.True(t, result.IsErr())
			} else {
				require.True(t, result.IsOk())
				assert.Equal(t, tt.expected, result.Unwrap())
			}
		})
	}
}

func TestScanTreeWithConfig_WithPrompter(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create test directory with files
	testDir := "/test/tree"
	require.NoError(t, fs.Mkdir(ctx, testDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, testDir+"/small.txt", []byte("small"), 0644))
	require.NoError(t, fs.WriteFile(ctx, testDir+"/large.bin", make([]byte, 2048), 0644))

	path := domain.NewFilePath(testDir).Unwrap()
	maxSize := int64(1024)
	prompter := scanner.NewBatchPrompter()

	result := scanner.ScanTreeWithConfig(ctx, fs, path, maxSize, prompter)

	require.True(t, result.IsOk(), "scan should succeed")
	tree := result.Unwrap()

	// Verify small file is included, large file excluded in batch mode
	hasSmallFile := false
	hasLargeFile := false
	for _, child := range tree.Children {
		if child.Path.String() == testDir+"/small.txt" {
			hasSmallFile = true
		}
		if child.Path.String() == testDir+"/large.bin" {
			hasLargeFile = true
		}
	}

	assert.True(t, hasSmallFile, "small file should be included")
	assert.False(t, hasLargeFile, "large file should be excluded")
}

func TestScanTreeWithConfig_LargeFiles(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory with only large files
	testDir := "/test/large"
	require.NoError(t, fs.Mkdir(ctx, testDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, testDir+"/huge1.bin", make([]byte, 5000), 0644))
	require.NoError(t, fs.WriteFile(ctx, testDir+"/huge2.bin", make([]byte, 6000), 0644))

	path := domain.NewFilePath(testDir).Unwrap()
	maxSize := int64(1024)
	prompter := scanner.NewBatchPrompter()

	result := scanner.ScanTreeWithConfig(ctx, fs, path, maxSize, prompter)

	require.True(t, result.IsOk(), "scan should succeed")
	tree := result.Unwrap()

	// All files should be excluded
	assert.Empty(t, tree.Children, "all large files should be excluded")
}

func TestScanTreeWithConfig_PrompterAccepts(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create test file
	testDir := "/test/accept"
	require.NoError(t, fs.Mkdir(ctx, testDir, 0755))
	testFile := testDir + "/file.txt"
	require.NoError(t, fs.WriteFile(ctx, testFile, []byte("test"), 0644))

	path := domain.NewFilePath(testDir).Unwrap()

	// Use nil prompter (no size limit)
	result := scanner.ScanTreeWithConfig(ctx, fs, path, 0, nil)

	require.True(t, result.IsOk(), "scan should succeed")
	tree := result.Unwrap()

	// File should be included when no size limit
	hasFile := false
	for _, child := range tree.Children {
		if child.Path.String() == testFile {
			hasFile = true
		}
	}
	assert.True(t, hasFile, "file should be included without size limit")
}

func TestScanTreeWithConfig_PrompterRejects(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory with large file
	testDir := "/test/reject"
	require.NoError(t, fs.Mkdir(ctx, testDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, testDir+"/large.dat", make([]byte, 3000), 0644))

	path := domain.NewFilePath(testDir).Unwrap()
	maxSize := int64(512)
	prompter := scanner.NewBatchPrompter() // Always rejects

	result := scanner.ScanTreeWithConfig(ctx, fs, path, maxSize, prompter)

	require.True(t, result.IsOk(), "scan should succeed")
	tree := result.Unwrap()

	// Large file should be rejected by batch prompter
	assert.Empty(t, tree.Children, "large file should be rejected")
}

func TestErrFileTooLarge_Error(t *testing.T) {
	// Test ErrFileTooLarge.Error() method
	err := scanner.ErrFileTooLarge{
		Path:  "/test/large.bin",
		Size:  2048,
		Limit: 1024,
	}

	errorMsg := err.Error()

	assert.Contains(t, errorMsg, "file too large")
	assert.Contains(t, errorMsg, "/test/large.bin")
	assert.Contains(t, errorMsg, "2.0 KB")
	assert.Contains(t, errorMsg, "1.0 KB")
}
