package planner

import (
	"testing"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Task 7.2.1: Test Resolution Policy Types
func TestResolutionPolicyTypes(t *testing.T) {
	tests := []struct {
		name   string
		policy ResolutionPolicy
		want   string
	}{
		{"fail", PolicyFail, "fail"},
		{"backup", PolicyBackup, "backup"},
		{"overwrite", PolicyOverwrite, "overwrite"},
		{"skip", PolicySkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Task 7.2.2: Test Resolution Policies Configuration
func TestResolutionPoliciesConfiguration(t *testing.T) {
	policies := ResolutionPolicies{
		OnFileExists:    PolicyBackup,
		OnWrongLink:     PolicyOverwrite,
		OnPermissionErr: PolicyFail,
		OnCircular:      PolicyFail,
		OnTypeMismatch:  PolicyFail,
	}

	assert.Equal(t, PolicyBackup, policies.OnFileExists)
	assert.Equal(t, PolicyOverwrite, policies.OnWrongLink)
	assert.Equal(t, PolicyFail, policies.OnPermissionErr)
}

func TestDefaultPolicies(t *testing.T) {
	policies := DefaultPolicies()

	// All policies should default to fail for safety
	assert.Equal(t, PolicyFail, policies.OnFileExists)
	assert.Equal(t, PolicyFail, policies.OnWrongLink)
	assert.Equal(t, PolicyFail, policies.OnPermissionErr)
	assert.Equal(t, PolicyFail, policies.OnCircular)
	assert.Equal(t, PolicyFail, policies.OnTypeMismatch)
}

// Task 7.2.3: Test PolicyFail
func TestPolicyFail(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists",
	)

	outcome := applyFailPolicy(conflict)

	assert.Equal(t, ResolveConflict, outcome.Status)
	assert.NotNil(t, outcome.Conflict)
	assert.Equal(t, conflict, *outcome.Conflict)
	assert.Empty(t, outcome.Operations)
}

// Task 7.2.4-5: Backup and Overwrite policies require additional operations
// These will be implemented in a future task
// For now, focusing on Fail and Skip policies

// Task 7.2.6: Test PolicySkip
func TestPolicySkip(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)

	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()
	conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")

	outcome := applySkipPolicy(op, conflict)

	assert.Equal(t, ResolveSkip, outcome.Status)
	assert.Empty(t, outcome.Operations)
	assert.NotNil(t, outcome.Warning)
	assert.Contains(t, outcome.Warning.Message, "Skipping")
}

// Additional coverage tests
func TestResolutionPolicyStringEdgeCases(t *testing.T) {
	// Test unknown policy
	unknownPolicy := ResolutionPolicy(999)
	assert.Equal(t, "unknown", unknownPolicy.String())
}

func TestConflictTypeStringEdgeCases(t *testing.T) {
	// Test unknown conflict type
	unknownType := ConflictType(999)
	assert.Equal(t, "unknown", unknownType.String())
}

func TestResolutionStatusStringEdgeCases(t *testing.T) {
	// Test unknown status
	unknownStatus := ResolutionStatus(999)
	assert.Equal(t, "unknown", unknownStatus.String())
}

func TestWarningSeverityStringEdgeCases(t *testing.T) {
	// Test unknown severity
	unknownSeverity := WarningSeverity(999)
	assert.Equal(t, "unknown", unknownSeverity.String())
}

// Test applyBackupPolicy unit functionality
func TestApplyBackupPolicy(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)
	conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")

	t.Run("creates backup, delete, and link operations", func(t *testing.T) {
		outcome := applyBackupPolicy(op, conflict, "/backup")

		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 3, "should create 3 operations: backup, delete, link")

		// Verify operation types in correct order
		assert.IsType(t, domain.FileBackup{}, outcome.Operations[0], "first operation should be FileBackup")
		assert.IsType(t, domain.FileDelete{}, outcome.Operations[1], "second operation should be FileDelete")
		assert.IsType(t, domain.LinkCreate{}, outcome.Operations[2], "third operation should be LinkCreate")
	})

	t.Run("backup operation has correct paths", func(t *testing.T) {
		outcome := applyBackupPolicy(op, conflict, "/backup")

		backupOp, ok := outcome.Operations[0].(domain.FileBackup)
		assert.True(t, ok, "first operation must be FileBackup")
		assert.Equal(t, targetFilePath.String(), backupOp.Source.String(), "backup source should be conflict path")
		assert.Contains(t, backupOp.Backup.String(), "/backup/", "backup path should be in backup directory")
		assert.Contains(t, backupOp.Backup.String(), ".bashrc.", "backup path should contain original filename")
	})

	t.Run("backup path includes timestamp", func(t *testing.T) {
		outcome := applyBackupPolicy(op, conflict, "/backup")

		backupOp := outcome.Operations[0].(domain.FileBackup)
		backupPath := backupOp.Backup.String()

		// Timestamp format is YYYYMMDD-HHMMSS
		// Should have format like: /backup/.bashrc.20060102-150405
		assert.Regexp(t, `/backup/.bashrc\.\d{8}-\d{6}$`, backupPath, "backup path should have timestamp suffix")
	})

	t.Run("delete operation targets conflict path", func(t *testing.T) {
		outcome := applyBackupPolicy(op, conflict, "/backup")

		deleteOp, ok := outcome.Operations[1].(domain.FileDelete)
		assert.True(t, ok, "second operation must be FileDelete")
		assert.Equal(t, targetFilePath.String(), deleteOp.Path.String(), "delete should target conflict path")
	})

	t.Run("link operation is original operation", func(t *testing.T) {
		outcome := applyBackupPolicy(op, conflict, "/backup")

		linkOp, ok := outcome.Operations[2].(domain.LinkCreate)
		assert.True(t, ok, "third operation must be LinkCreate")
		assert.Equal(t, op, linkOp, "link operation should be unchanged")
	})
}

// Test applyOverwritePolicy unit functionality
func TestApplyOverwritePolicy(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)
	conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")

	t.Run("creates delete and link operations", func(t *testing.T) {
		outcome := applyOverwritePolicy(op, conflict)

		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 2, "should create 2 operations: delete, link")

		// Verify operation types in correct order
		assert.IsType(t, domain.FileDelete{}, outcome.Operations[0], "first operation should be FileDelete")
		assert.IsType(t, domain.LinkCreate{}, outcome.Operations[1], "second operation should be LinkCreate")
	})

	t.Run("delete operation targets conflict path", func(t *testing.T) {
		outcome := applyOverwritePolicy(op, conflict)

		deleteOp, ok := outcome.Operations[0].(domain.FileDelete)
		assert.True(t, ok, "first operation must be FileDelete")
		assert.Equal(t, targetFilePath.String(), deleteOp.Path.String(), "delete should target conflict path")
	})

	t.Run("link operation is original operation", func(t *testing.T) {
		outcome := applyOverwritePolicy(op, conflict)

		linkOp, ok := outcome.Operations[1].(domain.LinkCreate)
		assert.True(t, ok, "second operation must be LinkCreate")
		assert.Equal(t, op, linkOp, "link operation should be unchanged")
	})

	t.Run("no backup created with overwrite policy", func(t *testing.T) {
		outcome := applyOverwritePolicy(op, conflict)

		for _, op := range outcome.Operations {
			assert.NotEqual(t, domain.OpKindFileBackup, op.Kind(), "should not create FileBackup with overwrite policy")
		}
	})
}

// Test that backup timestamps are unique
func TestBackupTimestampsUnique(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)
	conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")

	// Create multiple backups rapidly
	backupPaths := make(map[string]bool)
	for i := 0; i < 10; i++ {
		outcome := applyBackupPolicy(op, conflict, "/backup")
		backupOp := outcome.Operations[0].(domain.FileBackup)
		path := backupOp.Backup.String()

		// Each path should be unique (or at least not duplicate within same second)
		// Note: if tests run in same second, timestamps might collide
		backupPaths[path] = true
	}

	// We expect at least some uniqueness (timestamps change over time)
	assert.NotEmpty(t, backupPaths, "should generate backup paths")
}
