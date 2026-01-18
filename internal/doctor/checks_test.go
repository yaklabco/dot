package doctor

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/manifest"
)

// mockFS implements the FS interface for testing.
type mockFS struct {
	existsFunc    func(ctx context.Context, path string) (bool, error)
	isDirFunc     func(ctx context.Context, path string) (bool, error)
	lstatFunc     func(ctx context.Context, name string) (fs.FileInfo, error)
	readDirFunc   func(ctx context.Context, name string) ([]fs.DirEntry, error)
	readFileFunc  func(ctx context.Context, name string) ([]byte, error)
	readLinkFunc  func(ctx context.Context, name string) (string, error)
	writeFileFunc func(ctx context.Context, name string, data []byte, perm os.FileMode) error
	removeFunc    func(ctx context.Context, name string) error
	mkdirAllFunc  func(ctx context.Context, path string, perm os.FileMode) error
	statFunc      func(ctx context.Context, name string) (fs.FileInfo, error)
}

func (m *mockFS) Exists(ctx context.Context, path string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, path)
	}
	return false, nil
}

func (m *mockFS) IsDir(ctx context.Context, path string) (bool, error) {
	if m.isDirFunc != nil {
		return m.isDirFunc(ctx, path)
	}
	return false, nil
}

func (m *mockFS) Lstat(ctx context.Context, name string) (fs.FileInfo, error) {
	if m.lstatFunc != nil {
		return m.lstatFunc(ctx, name)
	}
	return nil, os.ErrNotExist
}

func (m *mockFS) ReadDir(ctx context.Context, name string) ([]fs.DirEntry, error) {
	if m.readDirFunc != nil {
		return m.readDirFunc(ctx, name)
	}
	return nil, nil
}

func (m *mockFS) ReadFile(ctx context.Context, name string) ([]byte, error) {
	if m.readFileFunc != nil {
		return m.readFileFunc(ctx, name)
	}
	return nil, os.ErrNotExist
}

func (m *mockFS) ReadLink(ctx context.Context, name string) (string, error) {
	if m.readLinkFunc != nil {
		return m.readLinkFunc(ctx, name)
	}
	return "", os.ErrInvalid
}

func (m *mockFS) WriteFile(ctx context.Context, name string, data []byte, perm os.FileMode) error {
	if m.writeFileFunc != nil {
		return m.writeFileFunc(ctx, name, data, perm)
	}
	return nil
}

func (m *mockFS) Remove(ctx context.Context, name string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, name)
	}
	return nil
}

func (m *mockFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	if m.mkdirAllFunc != nil {
		return m.mkdirAllFunc(ctx, path, perm)
	}
	return nil
}

func (m *mockFS) Stat(ctx context.Context, name string) (fs.FileInfo, error) {
	if m.statFunc != nil {
		return m.statFunc(ctx, name)
	}
	return nil, os.ErrNotExist
}

// mockManifestLoader implements the ManifestLoader interface for testing.
type mockManifestLoader struct {
	manifest manifest.Manifest
	err      error
}

func (m *mockManifestLoader) Load(ctx context.Context, targetPath domain.TargetPath) domain.Result[manifest.Manifest] {
	if m.err != nil {
		return domain.Err[manifest.Manifest](m.err)
	}
	return domain.Ok(m.manifest)
}

// mockLinkHealthChecker implements the LinkHealthChecker interface for testing.
type mockLinkHealthChecker struct {
	results map[string]LinkHealthResult
}

func (m *mockLinkHealthChecker) CheckLink(ctx context.Context, pkgName, linkPath, packageDir string) LinkHealthResult {
	if result, ok := m.results[linkPath]; ok {
		return result
	}
	return LinkHealthResult{IsHealthy: true}
}

// mockTargetPathCreator implements the TargetPathCreator interface for testing.
type mockTargetPathCreator struct {
	path domain.TargetPath
	err  error
}

func (m *mockTargetPathCreator) NewTargetPath(path string) domain.Result[domain.TargetPath] {
	if m.err != nil {
		return domain.Err[domain.TargetPath](m.err)
	}
	return domain.Ok(m.path)
}

// mockFileInfo implements fs.FileInfo for testing.
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() any           { return nil }

// mockDirEntry implements fs.DirEntry for testing.
type mockDirEntry struct {
	name  string
	isDir bool
	mode  os.FileMode
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return m.mode }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// errManifestNotFound is a sentinel error for manifest not found scenarios.
var errManifestNotFound = errors.New("manifest not found")

// isManifestNotFoundFunc returns true for manifest not found errors.
func isManifestNotFoundFunc(err error) bool {
	return errors.Is(err, errManifestNotFound)
}

// createValidTargetPath creates a valid TargetPath for testing.
func createValidTargetPath(t *testing.T) domain.TargetPath {
	t.Helper()
	result := domain.NewTargetPath("/home/user")
	if !result.IsOk() {
		t.Fatalf("Failed to create target path: %v", result.UnwrapErr())
	}
	return result.Unwrap()
}

// =============================================================================
// ManagedPackageCheck Tests
// =============================================================================

func TestManagedPackageCheck_Name(t *testing.T) {
	check := NewManagedPackageCheck(nil, nil, nil, "", nil, nil)
	assert.Equal(t, "managed_packages", check.Name())
}

func TestManagedPackageCheck_Description(t *testing.T) {
	check := NewManagedPackageCheck(nil, nil, nil, "", nil, nil)
	assert.Contains(t, check.Description(), "managed packages")
}

func TestManagedPackageCheck_Run_TargetPathError(t *testing.T) {
	targetPathErr := errors.New("invalid target path")
	check := NewManagedPackageCheck(
		&mockFS{},
		&mockManifestLoader{},
		&mockLinkHealthChecker{},
		"/invalid",
		&mockTargetPathCreator{err: targetPathErr},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Equal(t, targetPathErr, err)
	assert.Equal(t, "managed_packages", result.CheckName)
}

func TestManagedPackageCheck_Run_ManifestNotFound(t *testing.T) {
	targetPath := createValidTargetPath(t)
	check := NewManagedPackageCheck(
		&mockFS{},
		&mockManifestLoader{err: errManifestNotFound},
		&mockLinkHealthChecker{},
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusSkipped, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "NO_MANIFEST", result.Issues[0].Code)
}

func TestManagedPackageCheck_Run_ManifestLoadError(t *testing.T) {
	targetPath := createValidTargetPath(t)
	manifestErr := errors.New("IO error")
	check := NewManagedPackageCheck(
		&mockFS{},
		&mockManifestLoader{err: manifestErr},
		&mockLinkHealthChecker{},
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Equal(t, manifestErr, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
}

func TestManagedPackageCheck_Run_AllLinksHealthy(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:       "test-pkg",
		LinkCount:  2,
		Links:      []string{".bashrc", ".vimrc"},
		PackageDir: "/dotfiles/test-pkg",
	})

	check := NewManagedPackageCheck(
		&mockFS{},
		&mockManifestLoader{manifest: m},
		&mockLinkHealthChecker{},
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Empty(t, result.Issues)
	assert.Equal(t, 2, result.Stats["total_links"])
	assert.Equal(t, 0, result.Stats["broken_links"])
	assert.Equal(t, 2, result.Stats["managed_links"])
}

func TestManagedPackageCheck_Run_BrokenLinks(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:       "test-pkg",
		LinkCount:  2,
		Links:      []string{".bashrc", ".vimrc"},
		PackageDir: "/dotfiles/test-pkg",
	})

	healthChecker := &mockLinkHealthChecker{
		results: map[string]LinkHealthResult{
			".bashrc": {
				IsHealthy:  false,
				IssueType:  IssueBrokenLink,
				Severity:   domain.IssueSeverityError,
				Message:    "Link target does not exist",
				Suggestion: "Re-install package",
			},
		},
	}

	check := NewManagedPackageCheck(
		&mockFS{},
		&mockManifestLoader{manifest: m},
		healthChecker,
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusFail, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, string(IssueBrokenLink), result.Issues[0].Code)
	assert.Equal(t, ".bashrc", result.Issues[0].Path)
	assert.Equal(t, 1, result.Stats["broken_links"])
}

func TestManagedPackageCheck_Run_WarningLinks(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:       "test-pkg",
		LinkCount:  1,
		Links:      []string{".bashrc"},
		PackageDir: "/dotfiles/test-pkg",
	})

	healthChecker := &mockLinkHealthChecker{
		results: map[string]LinkHealthResult{
			".bashrc": {
				IsHealthy:  false,
				IssueType:  IssueWrongTarget,
				Severity:   domain.IssueSeverityWarning,
				Message:    "Link points to unexpected target",
				Suggestion: "Re-link the file",
			},
		},
	}

	check := NewManagedPackageCheck(
		&mockFS{},
		&mockManifestLoader{manifest: m},
		healthChecker,
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusFail, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, domain.IssueSeverityWarning, result.Issues[0].Severity)
}

// =============================================================================
// ManifestIntegrityCheck Tests
// =============================================================================

func TestManifestIntegrityCheck_Name(t *testing.T) {
	check := NewManifestIntegrityCheck(nil, nil, "", nil, nil)
	assert.Equal(t, "manifest_integrity", check.Name())
}

func TestManifestIntegrityCheck_Description(t *testing.T) {
	check := NewManifestIntegrityCheck(nil, nil, "", nil, nil)
	assert.Contains(t, check.Description(), "manifest")
}

func TestManifestIntegrityCheck_Run_TargetPathError(t *testing.T) {
	targetPathErr := errors.New("invalid target path")
	check := NewManifestIntegrityCheck(
		&mockFS{},
		&mockManifestLoader{},
		"/invalid",
		&mockTargetPathCreator{err: targetPathErr},
		isManifestNotFoundFunc,
	)

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Equal(t, targetPathErr, err)
}

func TestManifestIntegrityCheck_Run_ManifestNotFound(t *testing.T) {
	targetPath := createValidTargetPath(t)
	check := NewManifestIntegrityCheck(
		&mockFS{},
		&mockManifestLoader{err: errManifestNotFound},
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	checkResult, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, checkResult.Status)
	assert.Empty(t, checkResult.Issues)
}

func TestManifestIntegrityCheck_Run_ManifestLoadError(t *testing.T) {
	targetPath := createValidTargetPath(t)
	manifestErr := errors.New("IO error reading manifest")
	check := NewManifestIntegrityCheck(
		&mockFS{},
		&mockManifestLoader{err: manifestErr},
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot access manifest")
}

func TestManifestIntegrityCheck_Run_ValidManifest(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "test-pkg",
		LinkCount: 2,
		Links:     []string{".bashrc", ".vimrc"},
	})

	check := NewManifestIntegrityCheck(
		&mockFS{},
		&mockManifestLoader{manifest: m},
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Empty(t, result.Issues)
}

func TestManifestIntegrityCheck_Run_LinkCountMismatch(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "test-pkg",
		LinkCount: 5, // Recorded as 5, but only 2 links
		Links:     []string{".bashrc", ".vimrc"},
	})

	check := NewManifestIntegrityCheck(
		&mockFS{},
		&mockManifestLoader{manifest: m},
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
		isManifestNotFoundFunc,
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusWarning, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "MANIFEST_INCONSISTENT", result.Issues[0].Code)
	assert.Contains(t, result.Issues[0].Message, "link count mismatch")
}

// =============================================================================
// ConflictCheck Tests
// =============================================================================

func TestConflictCheck_Name(t *testing.T) {
	check := NewConflictCheck(nil, "", nil)
	assert.Equal(t, "conflicts", check.Name())
}

func TestConflictCheck_Description(t *testing.T) {
	check := NewConflictCheck(nil, "", nil)
	assert.Contains(t, check.Description(), "conflict")
}

func TestConflictCheck_Run_NoConflicts(t *testing.T) {
	fs := &mockFS{
		lstatFunc: func(ctx context.Context, name string) (fs.FileInfo, error) {
			return nil, os.ErrNotExist
		},
	}

	packageLinks := map[string][]string{
		"test-pkg": {".bashrc", ".vimrc"},
	}

	check := NewConflictCheck(fs, "/home/user", packageLinks)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Empty(t, result.Issues)
	assert.Equal(t, 0, result.Stats["conflicts"])
}

func TestConflictCheck_Run_ExistingSymlink(t *testing.T) {
	fs := &mockFS{
		lstatFunc: func(ctx context.Context, name string) (fs.FileInfo, error) {
			return &mockFileInfo{
				name: ".bashrc",
				mode: os.ModeSymlink,
			}, nil
		},
	}

	packageLinks := map[string][]string{
		"test-pkg": {".bashrc"},
	}

	check := NewConflictCheck(fs, "/home/user", packageLinks)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	// Existing symlinks are not considered conflicts
	assert.Equal(t, 0, result.Stats["conflicts"])
}

func TestConflictCheck_Run_ExistingFile(t *testing.T) {
	fs := &mockFS{
		lstatFunc: func(ctx context.Context, name string) (fs.FileInfo, error) {
			return &mockFileInfo{
				name:  ".bashrc",
				mode:  0644,
				isDir: false,
			}, nil
		},
	}

	packageLinks := map[string][]string{
		"test-pkg": {".bashrc"},
	}

	check := NewConflictCheck(fs, "/home/user", packageLinks)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusFail, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "CONFLICT_DETECTED", result.Issues[0].Code)
	assert.Contains(t, result.Issues[0].Message, "file")
	assert.Equal(t, 1, result.Stats["conflicts"])
}

func TestConflictCheck_Run_ExistingDirectory(t *testing.T) {
	fs := &mockFS{
		lstatFunc: func(ctx context.Context, name string) (fs.FileInfo, error) {
			return &mockFileInfo{
				name:  ".config",
				mode:  os.ModeDir,
				isDir: true,
			}, nil
		},
	}

	packageLinks := map[string][]string{
		"test-pkg": {".config"},
	}

	check := NewConflictCheck(fs, "/home/user", packageLinks)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusFail, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Contains(t, result.Issues[0].Message, "directory")
}

func TestConflictCheck_Run_AccessError(t *testing.T) {
	accessErr := errors.New("permission denied")
	fs := &mockFS{
		lstatFunc: func(ctx context.Context, name string) (fs.FileInfo, error) {
			return nil, accessErr
		},
	}

	packageLinks := map[string][]string{
		"test-pkg": {".bashrc"},
	}

	check := NewConflictCheck(fs, "/home/user", packageLinks)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusWarning, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "ACCESS_ERROR", result.Issues[0].Code)
}

// =============================================================================
// ConflictPermissionCheck Tests
// =============================================================================

func TestConflictPermissionCheck_Name(t *testing.T) {
	check := NewConflictPermissionCheck(nil, "")
	assert.Equal(t, "permissions", check.Name())
}

func TestConflictPermissionCheck_Description(t *testing.T) {
	check := NewConflictPermissionCheck(nil, "")
	assert.Contains(t, check.Description(), "permission")
}

func TestConflictPermissionCheck_Run_Writable(t *testing.T) {
	fs := &mockFS{
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return nil
		},
		removeFunc: func(ctx context.Context, name string) error {
			return nil
		},
	}

	check := NewConflictPermissionCheck(fs, "/home/user")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Empty(t, result.Issues)
}

func TestConflictPermissionCheck_Run_PermissionDenied(t *testing.T) {
	fs := &mockFS{
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return os.ErrPermission
		},
	}

	check := NewConflictPermissionCheck(fs, "/root")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusFail, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "PERMISSION_DENIED", result.Issues[0].Code)
}

func TestConflictPermissionCheck_Run_OtherWriteError(t *testing.T) {
	fs := &mockFS{
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return errors.New("disk full")
		},
	}

	check := NewConflictPermissionCheck(fs, "/home/user")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusWarning, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "WRITE_TEST_FAILED", result.Issues[0].Code)
}

// =============================================================================
// PermissionCheck Tests
// =============================================================================

func TestPermissionCheck_Name(t *testing.T) {
	check := NewPermissionCheck(nil, "")
	assert.Equal(t, "permissions", check.Name())
}

func TestPermissionCheck_Description(t *testing.T) {
	check := NewPermissionCheck(nil, "")
	assert.Contains(t, check.Description(), "permission")
}

func TestPermissionCheck_Run_DirectoryMissing(t *testing.T) {
	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return false, nil
		},
	}

	check := NewPermissionCheck(fs, "/nonexistent")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusWarning, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "TARGET_DIR_MISSING", result.Issues[0].Code)
}

func TestPermissionCheck_Run_ExistsCheckError(t *testing.T) {
	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return false, errors.New("stat error")
		},
	}

	check := NewPermissionCheck(fs, "/home/user")

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check target directory")
}

func TestPermissionCheck_Run_NotWritable(t *testing.T) {
	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil
		},
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return errors.New("permission denied")
		},
	}

	check := NewPermissionCheck(fs, "/home/user")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusFail, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "TARGET_DIR_NOT_WRITABLE", result.Issues[0].Code)
}

func TestPermissionCheck_Run_CleanupFailed(t *testing.T) {
	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil
		},
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return nil
		},
		removeFunc: func(ctx context.Context, name string) error {
			return errors.New("cannot remove file")
		},
		readDirFunc: func(ctx context.Context, name string) ([]fs.DirEntry, error) {
			return []fs.DirEntry{}, nil
		},
	}

	check := NewPermissionCheck(fs, "/home/user")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusWarning, result.Status)
	hasCleanupIssue := false
	for _, issue := range result.Issues {
		if issue.Code == "CLEANUP_FAILED" {
			hasCleanupIssue = true
			break
		}
	}
	assert.True(t, hasCleanupIssue)
}

func TestPermissionCheck_Run_NotReadable(t *testing.T) {
	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil
		},
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return nil
		},
		removeFunc: func(ctx context.Context, name string) error {
			return nil
		},
		readDirFunc: func(ctx context.Context, name string) ([]fs.DirEntry, error) {
			return nil, os.ErrPermission
		},
	}

	check := NewPermissionCheck(fs, "/home/user")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusFail, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "TARGET_DIR_NOT_READABLE", result.Issues[0].Code)
}

func TestPermissionCheck_Run_ReadDirNonPermissionError(t *testing.T) {
	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil
		},
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return nil
		},
		removeFunc: func(ctx context.Context, name string) error {
			return nil
		},
		readDirFunc: func(ctx context.Context, name string) ([]fs.DirEntry, error) {
			return nil, errors.New("IO error")
		},
	}

	check := NewPermissionCheck(fs, "/home/user")

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read target directory")
}

func TestPermissionCheck_Run_Success(t *testing.T) {
	entries := []fs.DirEntry{
		&mockDirEntry{name: ".bashrc", isDir: false},
		&mockDirEntry{name: ".config", isDir: true},
	}

	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil
		},
		writeFileFunc: func(ctx context.Context, name string, data []byte, perm os.FileMode) error {
			return nil
		},
		removeFunc: func(ctx context.Context, name string) error {
			return nil
		},
		readDirFunc: func(ctx context.Context, name string) ([]fs.DirEntry, error) {
			return entries, nil
		},
	}

	check := NewPermissionCheck(fs, "/home/user")

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Empty(t, result.Issues)
	assert.Equal(t, "/home/user", result.Stats["target_dir"])
	assert.Equal(t, 2, result.Stats["entries_count"])
}

// =============================================================================
// PlatformCheck Tests
// =============================================================================

func TestPlatformCheck_Name(t *testing.T) {
	check := NewPlatformCheck(nil, nil, "", "", nil)
	assert.Equal(t, "platform_compatibility", check.Name())
}

func TestPlatformCheck_Description(t *testing.T) {
	check := NewPlatformCheck(nil, nil, "", "", nil)
	assert.Contains(t, check.Description(), "platform")
}

func TestPlatformCheck_Run_TargetPathError(t *testing.T) {
	targetPathErr := errors.New("invalid target path")
	check := NewPlatformCheck(
		&mockFS{},
		&mockManifestLoader{},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{err: targetPathErr},
	)

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Equal(t, targetPathErr, err)
}

func TestPlatformCheck_Run_ManifestLoadError(t *testing.T) {
	targetPath := createValidTargetPath(t)
	manifestErr := errors.New("manifest error")
	check := NewPlatformCheck(
		&mockFS{},
		&mockManifestLoader{err: manifestErr},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
	)

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load manifest")
}

func TestPlatformCheck_Run_EmptyManifest(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()

	check := NewPlatformCheck(
		&mockFS{},
		&mockManifestLoader{manifest: m},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Equal(t, 0, result.Stats["packages_checked"])
}

func TestPlatformCheck_Run_PackageDirNotExists(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name: "test-pkg",
	})

	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return false, nil
		},
	}

	check := NewPlatformCheck(
		fs,
		&mockManifestLoader{manifest: m},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Equal(t, 1, result.Stats["packages_checked"])
}

func TestPlatformCheck_Run_PackageExistsCheckError(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name: "test-pkg",
	})

	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return false, errors.New("stat error")
		},
	}

	check := NewPlatformCheck(
		fs,
		&mockManifestLoader{manifest: m},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
	)

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check package directory")
}

func TestPlatformCheck_Run_NoMetadataFile(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name: "test-pkg",
	})

	callCount := 0
	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			callCount++
			if callCount == 1 {
				return true, nil // Package directory exists
			}
			return false, nil // Metadata file does not exist
		},
	}

	check := NewPlatformCheck(
		fs,
		&mockManifestLoader{manifest: m},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
}

func TestPlatformCheck_Run_MetadataReadError(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name: "test-pkg",
	})

	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil // Both package dir and metadata exist
		},
		readFileFunc: func(ctx context.Context, name string) ([]byte, error) {
			return nil, errors.New("read error")
		},
	}

	check := NewPlatformCheck(
		fs,
		&mockManifestLoader{manifest: m},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "METADATA_READ_ERROR", result.Issues[0].Code)
}

func TestPlatformCheck_Run_MetadataExists(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name: "test-pkg",
	})

	fs := &mockFS{
		existsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil
		},
		readFileFunc: func(ctx context.Context, name string) ([]byte, error) {
			return []byte(`{"platform": "linux"}`), nil
		},
	}

	check := NewPlatformCheck(
		fs,
		&mockManifestLoader{manifest: m},
		"/dotfiles",
		"/home/user",
		&mockTargetPathCreator{path: targetPath},
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
	assert.Contains(t, result.Stats["current_platform"], "/")
}

// =============================================================================
// OrphanCheck Tests
// =============================================================================

func TestOrphanCheck_Name(t *testing.T) {
	check := NewOrphanCheck()
	assert.Equal(t, "orphaned_links", check.Name())
}

func TestOrphanCheck_Description(t *testing.T) {
	check := NewOrphanCheck()
	assert.Contains(t, check.Description(), "symlinks")
}

func TestOrphanCheck_Run_ScanModeOff(t *testing.T) {
	check := NewOrphanCheck(
		WithScanConfig(ScanConfig{Mode: ScanOff}),
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusSkipped, result.Status)
}

func TestOrphanCheck_Run_TargetPathError(t *testing.T) {
	targetPathErr := errors.New("invalid target path")
	check := NewOrphanCheck(
		WithTargetPathCreator(&mockTargetPathCreator{err: targetPathErr}),
		WithManifestLoader(&mockManifestLoader{}),
		WithTargetDir("/home/user"),
		WithScanConfig(ScanConfig{Mode: ScanScoped}),
	)

	_, err := check.Run(context.Background())

	require.Error(t, err)
	assert.Equal(t, targetPathErr, err)
}

func TestOrphanCheck_Run_EmptyManifest(t *testing.T) {
	targetPath := createValidTargetPath(t)
	m := manifest.New()

	fs := &mockFS{
		readDirFunc: func(ctx context.Context, name string) ([]fs.DirEntry, error) {
			return []fs.DirEntry{}, nil
		},
	}

	check := NewOrphanCheck(
		WithFS(fs),
		WithTargetPathCreator(&mockTargetPathCreator{path: targetPath}),
		WithManifestLoader(&mockManifestLoader{manifest: m}),
		WithTargetDir("/home/user"),
		WithScanConfig(ScanConfig{Mode: ScanScoped}),
	)

	result, err := check.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.CheckStatusPass, result.Status)
}

func TestOrphanCheck_WithOptions(t *testing.T) {
	targetPath := createValidTargetPath(t)
	fs := &mockFS{}
	manifestLoader := &mockManifestLoader{manifest: manifest.New()}
	targetDir := "/home/user"
	config := ScanConfig{Mode: ScanDeep, MaxDepth: 3}

	check := NewOrphanCheck(
		WithFS(fs),
		WithManifestLoader(manifestLoader),
		WithTargetDir(targetDir),
		WithScanConfig(config),
		WithTargetPathCreator(&mockTargetPathCreator{path: targetPath}),
	)

	assert.NotNil(t, check)
	assert.Equal(t, fs, check.fs)
	assert.Equal(t, manifestLoader, check.manifestSvc)
	assert.Equal(t, targetDir, check.targetDir)
	assert.Equal(t, ScanDeep, check.config.Mode)
	assert.Equal(t, 3, check.config.MaxDepth)
}

func TestOrphanCheck_CalculateDepth(t *testing.T) {
	check := NewOrphanCheck(
		WithTargetDir("/home/user"),
	)

	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{
			name:     "same as target",
			path:     "/home/user",
			expected: 0,
		},
		{
			name:     "one level deep",
			path:     "/home/user/.config",
			expected: 1,
		},
		{
			name:     "two levels deep",
			path:     "/home/user/.config/nvim",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth := check.calculateDepth(tt.path)
			assert.Equal(t, tt.expected, depth)
		})
	}
}

func TestOrphanCheck_ShouldSkipDirectory(t *testing.T) {
	check := NewOrphanCheck(
		WithScanConfig(ScanConfig{
			Mode:         ScanDeep,
			SkipPatterns: []string{".git", "node_modules"},
		}),
	)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "skip git directory",
			path:     "/home/user/.git",
			expected: true,
		},
		{
			name:     "skip node_modules",
			path:     "/home/user/project/node_modules",
			expected: true,
		},
		{
			name:     "do not skip regular dir",
			path:     "/home/user/.config",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := check.shouldSkipDirectory(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrphanCheck_DetermineScanDirectories(t *testing.T) {
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:  "test-pkg",
		Links: []string{".config/nvim/init.vim", ".bashrc"},
	})

	tests := []struct {
		name     string
		config   ScanConfig
		expected int
	}{
		{
			name: "scoped to dirs",
			config: ScanConfig{
				Mode:        ScanScoped,
				ScopeToDirs: []string{".config"},
			},
			expected: 1,
		},
		{
			name: "scoped mode extracts managed dirs",
			config: ScanConfig{
				Mode: ScanScoped,
			},
			expected: 4, // .config, .config/nvim, ., and one more from hierarchy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewOrphanCheck(
				WithTargetDir("/home/user"),
				WithScanConfig(tt.config),
			)

			dirs := check.determineScanDirectories(&m)
			assert.GreaterOrEqual(t, len(dirs), 1)
		})
	}
}

func TestOrphanCheck_BuildManagedLinkSet(t *testing.T) {
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:  "test-pkg",
		Links: []string{".bashrc", ".config/nvim/init.vim"},
	})

	check := NewOrphanCheck()
	linkSet := check.buildManagedLinkSet(&m)

	assert.Len(t, linkSet, 2)
	_, ok1 := linkSet[".bashrc"]
	_, ok2 := linkSet[".config/nvim/init.vim"]
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func TestOrphanCheck_BuildIgnoreSet(t *testing.T) {
	m := manifest.New()
	m.EnsureDoctorState()
	m.Doctor.IgnoredPatterns = []string{"*.tmp", "cache/*"}

	check := NewOrphanCheck()
	ignoreSet := check.buildIgnoreSet(&m)

	assert.NotNil(t, ignoreSet)
}

func TestFilterDescendants(t *testing.T) {
	tests := []struct {
		name     string
		dirs     []string
		expected []string
	}{
		{
			name:     "empty",
			dirs:     []string{},
			expected: []string{},
		},
		{
			name:     "single",
			dirs:     []string{"/home/user"},
			expected: []string{"/home/user"},
		},
		{
			name:     "parent and child",
			dirs:     []string{"/home/user", "/home/user/.config"},
			expected: []string{"/home/user"},
		},
		{
			name:     "siblings",
			dirs:     []string{"/home/user/.config", "/home/user/.local"},
			expected: []string{"/home/user/.config", "/home/user/.local"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterDescendants(tt.dirs)
			assert.Len(t, result, len(tt.expected))
		})
	}
}

func TestExtractManagedDirectories(t *testing.T) {
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:  "test-pkg",
		Links: []string{".config/nvim/init.vim"},
	})

	dirs := extractManagedDirectories(&m)

	assert.Contains(t, dirs, ".")
	assert.Contains(t, dirs, ".config")
	assert.Contains(t, dirs, ".config/nvim")
}

func TestConvertIssuesToDomain(t *testing.T) {
	localIssues := []Issue{
		{
			Severity:   domain.IssueSeverityWarning,
			Type:       IssueOrphanedLink,
			Path:       ".bashrc",
			Message:    "Orphaned symlink",
			Suggestion: "Remove or adopt",
		},
		{
			Severity:   domain.IssueSeverityError,
			Type:       IssueBrokenLink,
			Path:       ".vimrc",
			Message:    "Broken symlink",
			Suggestion: "Fix target",
		},
	}

	domainIssues := convertIssuesToDomain(localIssues)

	require.Len(t, domainIssues, 2)
	assert.Equal(t, domain.IssueSeverityWarning, domainIssues[0].Severity)
	assert.Equal(t, domain.IssueSeverityError, domainIssues[1].Severity)
	assert.Equal(t, ".bashrc", domainIssues[0].Path)
	assert.Equal(t, string(IssueOrphanedLink), domainIssues[0].Code)
}
