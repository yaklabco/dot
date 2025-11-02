package domain_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkCreate_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))

	source := domain.MustParsePath("/source/file")
	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkCreate("link1", source, target)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify link was created
	isLink, _ := fs.IsSymlink(ctx, "/target/link")
	assert.True(t, isLink)
}

func TestLinkCreate_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/source/file", "/target/link"))

	source := domain.MustParsePath("/source/file")
	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkCreate("link1", source, target)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify link was removed
	assert.False(t, fs.Exists(ctx, "/target/link"))
}

func TestLinkDelete_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/source/file", "/target/link"))

	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkDelete("del1", target)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify link was deleted
	assert.False(t, fs.Exists(ctx, "/target/link"))
}

func TestLinkDelete_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))

	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	// LinkDelete rollback needs the original source to recreate the link
	// Since we don't store that, rollback returns ErrNotImplemented
	op := domain.NewLinkDelete("del1", target)

	err := op.Rollback(ctx, fs)
	// LinkDelete rollback returns nil (cannot restore without knowing source)
	assert.NoError(t, err)
}

func TestDirCreate_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent", 0755))

	path := domain.MustParsePath("/parent/newdir")
	op := domain.NewDirCreate("dir1", path)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify directory was created
	isDir, _ := fs.IsDir(ctx, "/parent/newdir")
	assert.True(t, isDir)
}

func TestDirCreate_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent/dir", 0755))

	path := domain.MustParsePath("/parent/dir")
	op := domain.NewDirCreate("dir1", path)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify directory was removed
	assert.False(t, fs.Exists(ctx, "/parent/dir"))
}

func TestDirDelete_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent/dir", 0755))

	path := domain.MustParsePath("/parent/dir")
	op := domain.NewDirDelete("del1", path)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify directory was deleted
	assert.False(t, fs.Exists(ctx, "/parent/dir"))
}

func TestDirDelete_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent", 0755))

	path := domain.MustParsePath("/parent/restoreddir")
	op := domain.NewDirDelete("del1", path)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify directory was recreated
	exists := fs.Exists(ctx, "/parent/restoreddir")
	assert.True(t, exists)
}

func TestDirRemoveAll_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Create directory with nested content
	require.NoError(t, fs.MkdirAll(ctx, "/parent/dir/subdir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/parent/dir/file1.txt", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/parent/dir/subdir/file2.txt", []byte("content2"), 0644))

	path := domain.MustParsePath("/parent/dir")
	op := domain.NewDirRemoveAll("del1", path)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify directory and all contents were deleted
	assert.False(t, fs.Exists(ctx, "/parent/dir"))
	assert.False(t, fs.Exists(ctx, "/parent/dir/file1.txt"))
	assert.False(t, fs.Exists(ctx, "/parent/dir/subdir"))
	assert.False(t, fs.Exists(ctx, "/parent/dir/subdir/file2.txt"))
}

func TestDirRemoveAll_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent", 0755))

	path := domain.MustParsePath("/parent/deleteddir")
	op := domain.NewDirRemoveAll("del1", path)

	err := op.Rollback(ctx, fs)
	// DirRemoveAll rollback returns nil (cannot restore without backup)
	assert.NoError(t, err)
}

func TestFileBackup_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/file.txt", []byte("original"), 0644))

	source := domain.MustParsePath("/test/file.txt")
	backup := domain.MustParsePath("/test/file.txt.bak")

	op := domain.NewFileBackup("bak1", source, backup)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify backup was created
	assert.True(t, fs.Exists(ctx, "/test/file.txt.bak"))

	// Verify content
	data, _ := fs.ReadFile(ctx, "/test/file.txt.bak")
	assert.Equal(t, []byte("original"), data)
}

func TestFileBackup_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/file.bak", []byte("backup"), 0644))

	source := domain.MustParsePath("/test/file")
	backup := domain.MustParsePath("/test/file.bak")

	op := domain.NewFileBackup("bak1", source, backup)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify backup was deleted
	assert.False(t, fs.Exists(ctx, "/test/file.bak"))
}

func TestFileDelete_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/file.txt", []byte("content"), 0644))

	path := domain.MustParsePath("/test/file.txt")
	op := domain.NewFileDelete("del1", path)

	// Verify file exists before deletion
	assert.True(t, fs.Exists(ctx, "/test/file.txt"))

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify file was deleted
	assert.False(t, fs.Exists(ctx, "/test/file.txt"))
}

func TestFileDelete_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	path := domain.MustParsePath("/test/file.txt")
	op := domain.NewFileDelete("del1", path)

	// Rollback cannot restore deleted file without backup
	// It should succeed but not restore the file
	err := op.Rollback(ctx, fs)
	require.NoError(t, err, "rollback should succeed even though it cannot restore")

	// File should still not exist (rollback is a no-op)
	assert.False(t, fs.Exists(ctx, "/test/file.txt"))
}

// Helper functions for backup integrity testing

func generateRandomBytes(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data)
	return data, err
}

func computeSHA256(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func verifyFileIntegrity(t *testing.T, fs domain.FS, ctx context.Context, path string, expectedContent []byte) {
	t.Helper()

	actualContent, err := fs.ReadFile(ctx, path)
	require.NoError(t, err, "failed to read file %s", path)

	// Byte-for-byte equality
	assert.Equal(t, expectedContent, actualContent, "file content must match exactly")

	// Checksum verification
	expectedHash := computeSHA256(expectedContent)
	actualHash := computeSHA256(actualContent)
	assert.Equal(t, expectedHash, actualHash, "SHA256 checksums must match")

	// Size verification
	assert.Equal(t, len(expectedContent), len(actualContent), "file size must match")
}

// TestFileBackup_ContentIntegrity tests backup operation with various content types
func TestFileBackup_ContentIntegrity(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Generate random data for moderate size test
	moderateData, err := generateRandomBytes(2 * 1024 * 1024) // 2MB
	require.NoError(t, err)

	testCases := []struct {
		name    string
		content []byte
	}{
		{"empty file", []byte{}},
		{"simple text", []byte("hello world")},
		{"multiline", []byte("line1\nline2\nline3\n")},
		{"special chars", []byte("!@#$%^&*()_+-={}[]|\\:\";<>?,./")},
		{"unicode", []byte("Hello 世界 🌍 Привет مرحبا")},
		{"binary data", []byte{0x00, 0xFF, 0xDE, 0xAD, 0xBE, 0xEF, 0x12, 0x34}},
		{"moderate size", moderateData},
		{"null bytes", []byte("before\x00middle\x00after\x00end")},
		{"mixed newlines", []byte("unix\nwindows\r\nmac\rend")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create source file
			sourcePath := "/test/source_" + tc.name + ".txt"
			backupPath := "/test/backup_" + tc.name + ".bak"
			require.NoError(t, fs.WriteFile(ctx, sourcePath, tc.content, 0644))

			// Calculate original checksum
			originalHash := computeSHA256(tc.content)

			// Execute backup operation
			source := domain.MustParsePath(sourcePath)
			backup := domain.MustParsePath(backupPath)
			op := domain.NewFileBackup("bak1", source, backup)
			require.NoError(t, op.Execute(ctx, fs))

			// Verify backup exists
			assert.True(t, fs.Exists(ctx, backupPath), "backup file must exist")

			// Verify backup content integrity
			verifyFileIntegrity(t, fs, ctx, backupPath, tc.content)

			// Verify backup checksum matches original
			backupData, err := fs.ReadFile(ctx, backupPath)
			require.NoError(t, err)
			backupHash := computeSHA256(backupData)
			assert.Equal(t, originalHash, backupHash, "backup checksum must match original")
		})
	}
}

// TestFileBackup_SourcePreservation verifies that backup operation doesn't modify source file
func TestFileBackup_SourcePreservation(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	testCases := []struct {
		name    string
		content []byte
	}{
		{"simple", []byte("original content")},
		{"binary", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"large", func() []byte {
			data, _ := generateRandomBytes(1024 * 1024) // 1MB
			return data
		}()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourcePath := "/test/source.txt"
			backupPath := "/test/backup.bak"

			// Write source file
			require.NoError(t, fs.WriteFile(ctx, sourcePath, tc.content, 0644))

			// Calculate original checksum before backup
			beforeHash := computeSHA256(tc.content)

			// Execute backup
			source := domain.MustParsePath(sourcePath)
			backup := domain.MustParsePath(backupPath)
			op := domain.NewFileBackup("bak1", source, backup)
			require.NoError(t, op.Execute(ctx, fs))

			// Read source file after backup
			sourceAfter, err := fs.ReadFile(ctx, sourcePath)
			require.NoError(t, err)

			// Verify source unchanged
			assert.Equal(t, tc.content, sourceAfter, "source file content must be unchanged")

			// Verify source checksum unchanged
			afterHash := computeSHA256(sourceAfter)
			assert.Equal(t, beforeHash, afterHash, "source file checksum must be unchanged")

			// Verify source and backup are identical
			backupData, err := fs.ReadFile(ctx, backupPath)
			require.NoError(t, err)
			assert.Equal(t, sourceAfter, backupData, "source and backup must be identical")
		})
	}
}

// TestFileBackup_LargeFileIntegrity tests backup with large files (stress test)
func TestFileBackup_LargeFileIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large file test in short mode")
	}

	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	testSizes := []struct {
		name string
		size int
	}{
		{"10MB", 10 * 1024 * 1024},
		{"50MB", 50 * 1024 * 1024},
		{"100MB", 100 * 1024 * 1024},
	}

	for _, tc := range testSizes {
		t.Run(tc.name, func(t *testing.T) {
			// Generate large random content
			largeContent, err := generateRandomBytes(tc.size)
			require.NoError(t, err, "failed to generate random data")

			// Calculate original checksum
			originalChecksum := computeSHA256(largeContent)

			// Write and backup
			sourcePath := "/test/large.bin"
			backupPath := "/test/large.bak"
			require.NoError(t, fs.WriteFile(ctx, sourcePath, largeContent, 0644))

			source := domain.MustParsePath(sourcePath)
			backup := domain.MustParsePath(backupPath)
			op := domain.NewFileBackup("bak1", source, backup)
			require.NoError(t, op.Execute(ctx, fs))

			// Verify backup integrity with checksum
			backupData, err := fs.ReadFile(ctx, backupPath)
			require.NoError(t, err)
			backupChecksum := computeSHA256(backupData)

			assert.Equal(t, originalChecksum, backupChecksum, "large file backup checksum must match")
			assert.Equal(t, tc.size, len(backupData), "backup size must match original")

			// Verify byte-for-byte equality (this is the ultimate test)
			assert.Equal(t, largeContent, backupData, "large file backup must be byte-perfect")
		})
	}
}

// TestFileBackup_PermissionsPreserved verifies that file permissions are preserved during backup
func TestFileBackup_PermissionsPreserved(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	testCases := []struct {
		name        string
		permissions os.FileMode
	}{
		{"read-only", 0400},
		{"owner rw", 0600},
		{"standard", 0644},
		{"executable", 0755},
		{"group rw", 0664},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := []byte("test content")
			sourcePath := "/test/source.txt"
			backupPath := "/test/backup.bak"

			// Create file with specific permissions
			require.NoError(t, fs.WriteFile(ctx, sourcePath, content, tc.permissions))

			// Get source file info
			sourceInfo, err := fs.Stat(ctx, sourcePath)
			require.NoError(t, err)

			// Execute backup
			source := domain.MustParsePath(sourcePath)
			backup := domain.MustParsePath(backupPath)
			op := domain.NewFileBackup("bak1", source, backup)
			require.NoError(t, op.Execute(ctx, fs))

			// Get backup file info
			backupInfo, err := fs.Stat(ctx, backupPath)
			require.NoError(t, err)

			// Verify permissions match
			assert.Equal(t, sourceInfo.Mode(), backupInfo.Mode(),
				"backup must preserve file permissions")
		})
	}
}
