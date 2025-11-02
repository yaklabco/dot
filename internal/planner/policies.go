package planner

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jamesainslie/dot/internal/domain"
)

// ResolutionPolicy defines how to handle conflicts
type ResolutionPolicy int

const (
	// PolicyFail stops and reports conflict (default, safest)
	PolicyFail ResolutionPolicy = iota
	// PolicyBackup backs up conflicting file before linking
	PolicyBackup
	// PolicyOverwrite replaces conflicting file with link
	PolicyOverwrite
	// PolicySkip skips conflicting operation
	PolicySkip
)

// String returns the string representation of ResolutionPolicy
func (rp ResolutionPolicy) String() string {
	switch rp {
	case PolicyFail:
		return "fail"
	case PolicyBackup:
		return "backup"
	case PolicyOverwrite:
		return "overwrite"
	case PolicySkip:
		return "skip"
	default:
		return "unknown"
	}
}

// ResolutionPolicies configures conflict resolution behavior per conflict type
type ResolutionPolicies struct {
	OnFileExists    ResolutionPolicy
	OnWrongLink     ResolutionPolicy
	OnPermissionErr ResolutionPolicy
	OnCircular      ResolutionPolicy
	OnTypeMismatch  ResolutionPolicy
}

// DefaultPolicies returns safe default policies (all fail)
func DefaultPolicies() ResolutionPolicies {
	return ResolutionPolicies{
		OnFileExists:    PolicyFail,
		OnWrongLink:     PolicyFail,
		OnPermissionErr: PolicyFail,
		OnCircular:      PolicyFail,
		OnTypeMismatch:  PolicyFail,
	}
}

// applyFailPolicy returns unresolved conflict
func applyFailPolicy(c Conflict) ResolutionOutcome {
	return ResolutionOutcome{
		Status:   ResolveConflict,
		Conflict: &c,
	}
}

// applySkipPolicy skips operation with warning
func applySkipPolicy(op domain.LinkCreate, c Conflict) ResolutionOutcome {
	warning := Warning{
		Message:  "Skipping due to conflict: " + op.Target.String(),
		Severity: WarnInfo,
	}

	return ResolutionOutcome{
		Status:  ResolveSkip,
		Warning: &warning,
	}
}

// applyBackupPolicy creates backup of existing file then creates symlink
func applyBackupPolicy(
	op domain.LinkCreate,
	conflict Conflict,
	backupDir string,
) ResolutionOutcome {
	// Generate timestamp for backup file
	timestamp := time.Now().Format("20060102-150405")

	// Extract filename from conflict path
	filename := filepath.Base(conflict.Path.String())

	// Generate backup path: <backupDir>/<filename>.<timestamp>
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s", filename, timestamp))
	backupFilePath := domain.NewFilePath(backupPath).Unwrap()

	// Create operations:
	// 1. FileBackup: backs up the conflicting file
	backupOpID := domain.OperationID(fmt.Sprintf("backup-%s-%s", conflict.Path.String(), timestamp))
	backupOp := domain.NewFileBackup(backupOpID, conflict.Path, backupFilePath)

	// 2. FileDelete: removes the original file
	deleteOpID := domain.OperationID(fmt.Sprintf("delete-%s", conflict.Path.String()))
	deleteOp := domain.NewFileDelete(deleteOpID, conflict.Path)

	// 3. LinkCreate: creates the symlink (original operation)

	return ResolutionOutcome{
		Status:     ResolveOK,
		Operations: []domain.Operation{backupOp, deleteOp, op},
	}
}

// applyOverwritePolicy deletes existing file then creates symlink
func applyOverwritePolicy(
	op domain.LinkCreate,
	conflict Conflict,
) ResolutionOutcome {
	// Create operations:
	// 1. FileDelete: removes the conflicting file
	deleteOpID := domain.OperationID(fmt.Sprintf("delete-%s", conflict.Path.String()))
	deleteOp := domain.NewFileDelete(deleteOpID, conflict.Path)

	// 2. LinkCreate: creates the symlink (original operation)

	return ResolutionOutcome{
		Status:     ResolveOK,
		Operations: []domain.Operation{deleteOp, op},
	}
}
