package planner

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

// TestBackupPolicy_PreservesContent tests end-to-end backup policy with content verification
func TestBackupPolicy_PreservesContent(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup directories
	require.NoError(t, fs.MkdirAll(ctx, "/packages/vim", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/user", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/backup", 0755))

	testCases := []struct {
		name            string
		conflictContent []byte
		packageContent  []byte
		targetPath      string
		sourcePath      string
	}{
		{
			name:            "simple text conflict",
			conflictContent: []byte("existing vimrc configuration"),
			packageContent:  []byte("new vimrc from package"),
			targetPath:      "/home/user/.vimrc",
			sourcePath:      "/packages/vim/.vimrc",
		},
		{
			name:            "binary conflict",
			conflictContent: []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE},
			packageContent:  []byte{0x12, 0x34, 0x56, 0x78},
			targetPath:      "/home/user/binary",
			sourcePath:      "/packages/vim/binary",
		},
		{
			name:            "unicode conflict",
			conflictContent: []byte("Original 世界 configuration 🌍"),
			packageContent:  []byte("Package مرحبا content"),
			targetPath:      "/home/user/.configrc",
			sourcePath:      "/packages/vim/.configrc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create conflicting target file with known content
			require.NoError(t, fs.WriteFile(ctx, tc.targetPath, tc.conflictContent, 0644))

			// Create package source file
			require.NoError(t, fs.WriteFile(ctx, tc.sourcePath, tc.packageContent, 0644))

			// Calculate original conflict checksum
			originalHash := sha256.Sum256(tc.conflictContent)

			// Create LinkCreate operation that will conflict
			sourceFilePath := domain.MustParsePath(tc.sourcePath)
			targetFilePath := domain.MustParseTargetPath(tc.targetPath)
			linkOp := domain.NewLinkCreate("link1", sourceFilePath, targetFilePath)

			// Create conflict
			targetPathForConflict := domain.MustParsePath(tc.targetPath)
			conflict := NewConflict(ConflictFileExists, targetPathForConflict, "File exists at target")

			// Apply backup policy
			outcome := applyBackupPolicy(linkOp, conflict, "/backup")

			// Verify policy generated correct operations
			require.Equal(t, ResolveOK, outcome.Status, "backup policy should resolve successfully")
			require.Len(t, outcome.Operations, 3, "should generate 3 operations: backup, delete, link")

			// Verify operation types
			assert.IsType(t, domain.FileBackup{}, outcome.Operations[0])
			assert.IsType(t, domain.FileDelete{}, outcome.Operations[1])
			assert.IsType(t, domain.LinkCreate{}, outcome.Operations[2])

			// Execute operations in order
			for i, op := range outcome.Operations {
				err := op.Execute(ctx, fs)
				require.NoError(t, err, "operation %d (%s) should execute successfully", i, op.String())
			}

			// Verify backup exists and contains original conflict content
			backupOp := outcome.Operations[0].(domain.FileBackup)
			backupPath := backupOp.Backup.String()

			assert.True(t, fs.Exists(ctx, backupPath), "backup file must exist")

			backupData, err := fs.ReadFile(ctx, backupPath)
			require.NoError(t, err)

			// Byte-for-byte verification
			assert.Equal(t, tc.conflictContent, backupData, "backup must contain exact original file content")

			// Checksum verification
			backupHash := sha256.Sum256(backupData)
			assert.Equal(t, originalHash, backupHash, "backup checksum must match original")

			// Verify symlink was created pointing to package source
			// (The original file was deleted and replaced with a symlink)
			// ReadLink succeeding proves it's a symlink
			linkTarget, err := fs.ReadLink(ctx, tc.targetPath)
			require.NoError(t, err, "symlink should exist at target path")
			assert.Equal(t, tc.sourcePath, linkTarget, "symlink must point to package source")

			// Cleanup for next iteration
			require.NoError(t, fs.Remove(ctx, backupPath))
			require.NoError(t, fs.Remove(ctx, tc.targetPath))
		})
	}
}

// TestBackupPolicy_MultipleConcurrentBackups tests handling multiple backups in one operation
func TestBackupPolicy_MultipleConcurrentBackups(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/packages/vim", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/user", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/backup", 0755))

	// Create multiple conflicting files with unique random content
	numFiles := 5
	files := make([]struct {
		targetPath string
		sourcePath string
		content    []byte
		checksum   [32]byte
	}, numFiles)

	for i := 0; i < numFiles; i++ {
		// Generate random content
		content := make([]byte, 1024)
		_, err := rand.Read(content)
		require.NoError(t, err)

		targetPathResult := domain.NewFilePath("/home/user/file" + string(rune('A'+i)))
		require.True(t, targetPathResult.IsOk())
		targetPath := targetPathResult.Unwrap()

		sourcePathResult := domain.NewFilePath("/packages/vim/file" + string(rune('A'+i)))
		require.True(t, sourcePathResult.IsOk())
		sourcePath := sourcePathResult.Unwrap()

		files[i].targetPath = targetPath.String()
		files[i].sourcePath = sourcePath.String()
		files[i].content = content
		files[i].checksum = sha256.Sum256(content)

		// Create conflicting target file
		require.NoError(t, fs.WriteFile(ctx, files[i].targetPath, content, 0644))

		// Create package source file
		require.NoError(t, fs.WriteFile(ctx, files[i].sourcePath, []byte("package content"), 0644))
	}

	// Apply backup policy to all files and execute
	for i, file := range files {
		sourceFilePath := domain.MustParsePath(file.sourcePath)
		targetFilePath := domain.MustParseTargetPath(file.targetPath)
		linkOp := domain.NewLinkCreate(domain.OperationID("link"+string(rune('A'+i))), sourceFilePath, targetFilePath)

		targetPathForConflict := domain.MustParsePath(file.targetPath)
		conflict := NewConflict(ConflictFileExists, targetPathForConflict, "File exists")

		// Apply backup policy
		outcome := applyBackupPolicy(linkOp, conflict, "/backup")
		require.Equal(t, ResolveOK, outcome.Status)

		// Execute all operations
		for _, op := range outcome.Operations {
			require.NoError(t, op.Execute(ctx, fs))
		}

		// Verify backup integrity
		backupOp := outcome.Operations[0].(domain.FileBackup)
		backupData, err := fs.ReadFile(ctx, backupOp.Backup.String())
		require.NoError(t, err)

		// Verify no cross-contamination
		backupHash := sha256.Sum256(backupData)
		assert.Equal(t, file.checksum, backupHash,
			"backup %d must contain correct original content (no cross-contamination)", i)
	}
}

// TestBackupPolicy_PermissionPreservation tests that backup policy preserves file permissions
func TestBackupPolicy_PermissionPreservation(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/packages/vim", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/user", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/backup", 0755))

	permissions := []struct {
		name string
		mode os.FileMode
	}{
		{"read-only", 0400},
		{"owner-rw", 0600},
		{"standard", 0644},
		{"executable", 0755},
	}

	for _, perm := range permissions {
		t.Run(perm.name, func(t *testing.T) {
			targetPath := "/home/user/.testfile"
			sourcePath := "/packages/vim/.testfile"

			// Create file with specific permissions
			content := []byte("test content")
			require.NoError(t, fs.WriteFile(ctx, targetPath, content, perm.mode))
			require.NoError(t, fs.WriteFile(ctx, sourcePath, []byte("package"), 0644))

			// Get original permissions
			origInfo, err := fs.Stat(ctx, targetPath)
			require.NoError(t, err)

			// Apply backup policy
			sourceFilePath := domain.MustParsePath(sourcePath)
			targetFilePath := domain.MustParseTargetPath(targetPath)
			linkOp := domain.NewLinkCreate("link1", sourceFilePath, targetFilePath)

			targetPathForConflict := domain.MustParsePath(targetPath)
			conflict := NewConflict(ConflictFileExists, targetPathForConflict, "File exists")

			outcome := applyBackupPolicy(linkOp, conflict, "/backup")
			require.Equal(t, ResolveOK, outcome.Status)

			// Execute backup operation only
			backupOp := outcome.Operations[0]
			require.NoError(t, backupOp.Execute(ctx, fs))

			// Verify backup has same permissions
			backupFilePath := backupOp.(domain.FileBackup).Backup
			backupInfo, err := fs.Stat(ctx, backupFilePath.String())
			require.NoError(t, err)

			assert.Equal(t, origInfo.Mode(), backupInfo.Mode(),
				"backup must preserve original file permissions")

			// Cleanup
			require.NoError(t, fs.Remove(ctx, backupFilePath.String()))
			require.NoError(t, fs.Remove(ctx, targetPath))
			require.NoError(t, fs.Remove(ctx, sourcePath))
		})
	}
}
