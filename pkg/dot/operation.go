package dot

import "github.com/yaklabco/dot/internal/domain"

// Operation type re-exports from internal/domain

// OperationKind identifies the type of operation.
type OperationKind = domain.OperationKind

// Operation kind constants
const (
	OpKindLinkCreate   = domain.OpKindLinkCreate
	OpKindLinkDelete   = domain.OpKindLinkDelete
	OpKindDirCreate    = domain.OpKindDirCreate
	OpKindDirDelete    = domain.OpKindDirDelete
	OpKindDirRemoveAll = domain.OpKindDirRemoveAll
	OpKindFileMove     = domain.OpKindFileMove
	OpKindFileBackup   = domain.OpKindFileBackup
	OpKindFileDelete   = domain.OpKindFileDelete
	OpKindDirCopy      = domain.OpKindDirCopy
)

// OperationID uniquely identifies an operation.
type OperationID = domain.OperationID

// Operation represents a filesystem operation to be executed.
type Operation = domain.Operation

// LinkCreate creates a symbolic link.
type LinkCreate = domain.LinkCreate

// LinkDelete removes a symbolic link.
type LinkDelete = domain.LinkDelete

// DirCreate creates a directory.
type DirCreate = domain.DirCreate

// DirDelete removes a directory.
type DirDelete = domain.DirDelete

// DirRemoveAll recursively removes a directory and all its contents.
type DirRemoveAll = domain.DirRemoveAll

// FileMove moves a file from one location to another.
type FileMove = domain.FileMove

// FileBackup backs up a file before modification.
type FileBackup = domain.FileBackup

// FileDelete deletes a file.
type FileDelete = domain.FileDelete

// DirCopy recursively copies a directory.
type DirCopy = domain.DirCopy

// NewLinkCreate creates a new LinkCreate operation.
func NewLinkCreate(id OperationID, source FilePath, target TargetPath) LinkCreate {
	return domain.NewLinkCreate(id, source, target)
}

// NewFileMove creates a new FileMove operation.
func NewFileMove(id OperationID, source TargetPath, dest FilePath) FileMove {
	return domain.NewFileMove(id, source, dest)
}

// NewLinkDelete creates a new LinkDelete operation.
func NewLinkDelete(id OperationID, target TargetPath) LinkDelete {
	return domain.NewLinkDelete(id, target)
}

// NewDirCreate creates a new DirCreate operation.
func NewDirCreate(id OperationID, path FilePath) DirCreate {
	return domain.NewDirCreate(id, path)
}

// NewDirDelete creates a new DirDelete operation.
func NewDirDelete(id OperationID, path FilePath) DirDelete {
	return domain.NewDirDelete(id, path)
}

// NewDirRemoveAll creates a new DirRemoveAll operation.
func NewDirRemoveAll(id OperationID, path FilePath) DirRemoveAll {
	return domain.NewDirRemoveAll(id, path)
}

// NewFileBackup creates a new FileBackup operation.
func NewFileBackup(id OperationID, source, backup FilePath) FileBackup {
	return domain.NewFileBackup(id, source, backup)
}

// NewFileDelete creates a new FileDelete operation.
func NewFileDelete(id OperationID, path FilePath) FileDelete {
	return domain.NewFileDelete(id, path)
}

// NewDirCopy creates a new DirCopy operation.
func NewDirCopy(id OperationID, source, dest FilePath) DirCopy {
	return domain.NewDirCopy(id, source, dest)
}
