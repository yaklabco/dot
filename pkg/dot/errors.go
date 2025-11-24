package dot

import (
	"fmt"

	"github.com/yaklabco/dot/internal/domain"
)

// Error types re-exported from internal/domain

// ErrInvalidPath represents a path validation error.
type ErrInvalidPath = domain.ErrInvalidPath

// ErrPackageNotFound represents a missing package error.
type ErrPackageNotFound = domain.ErrPackageNotFound

// ErrConflict represents a conflict during installation.
type ErrConflict = domain.ErrConflict

// ErrCyclicDependency represents a dependency cycle error.
type ErrCyclicDependency = domain.ErrCyclicDependency

// ErrFilesystemOperation represents a filesystem operation error.
type ErrFilesystemOperation = domain.ErrFilesystemOperation

// ErrPermissionDenied represents a permission denied error.
type ErrPermissionDenied = domain.ErrPermissionDenied

// ErrMultiple represents multiple aggregated errors.
type ErrMultiple = domain.ErrMultiple

// ErrEmptyPlan represents an empty plan error.
type ErrEmptyPlan = domain.ErrEmptyPlan

// ErrSourceNotFound represents a missing source file error.
type ErrSourceNotFound = domain.ErrSourceNotFound

// ErrExecutionFailed represents an execution failure error.
type ErrExecutionFailed = domain.ErrExecutionFailed

// ErrParentNotFound represents a missing parent directory error.
type ErrParentNotFound = domain.ErrParentNotFound

// ErrCheckpointNotFound represents a missing checkpoint error.
type ErrCheckpointNotFound = domain.ErrCheckpointNotFound

// ErrNotImplemented represents a not implemented error.
type ErrNotImplemented = domain.ErrNotImplemented

// Clone-specific error types

// ErrPackageDirNotEmpty indicates the package directory is not empty.
type ErrPackageDirNotEmpty struct {
	Path  string
	Cause error
}

func (e ErrPackageDirNotEmpty) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("package directory not empty: %s: %v", e.Path, e.Cause)
	}
	return fmt.Sprintf("package directory not empty: %s", e.Path)
}

func (e ErrPackageDirNotEmpty) Unwrap() error {
	return e.Cause
}

// ErrBootstrapNotFound indicates the bootstrap configuration file was not found.
type ErrBootstrapNotFound struct {
	Path string
}

func (e ErrBootstrapNotFound) Error() string {
	return fmt.Sprintf("bootstrap configuration not found: %s", e.Path)
}

// ErrInvalidBootstrap indicates the bootstrap configuration is invalid.
type ErrInvalidBootstrap struct {
	Reason string
	Cause  error
}

func (e ErrInvalidBootstrap) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalid bootstrap configuration: %s: %v", e.Reason, e.Cause)
	}
	return fmt.Sprintf("invalid bootstrap configuration: %s", e.Reason)
}

func (e ErrInvalidBootstrap) Unwrap() error {
	return e.Cause
}

// ErrAuthFailed indicates authentication failure during git clone.
type ErrAuthFailed struct {
	Cause error
}

func (e ErrAuthFailed) Error() string {
	return fmt.Sprintf("authentication failed: %v", e.Cause)
}

func (e ErrAuthFailed) Unwrap() error {
	return e.Cause
}

// ErrCloneFailed indicates repository cloning failed.
type ErrCloneFailed struct {
	URL   string
	Cause error
}

func (e ErrCloneFailed) Error() string {
	return fmt.Sprintf("clone failed for %s: %v", e.URL, e.Cause)
}

func (e ErrCloneFailed) Unwrap() error {
	return e.Cause
}

// ErrProfileNotFound indicates the requested profile does not exist.
type ErrProfileNotFound struct {
	Profile string
}

func (e ErrProfileNotFound) Error() string {
	return fmt.Sprintf("profile not found: %s", e.Profile)
}

// ErrBootstrapExists indicates the bootstrap file already exists.
type ErrBootstrapExists struct {
	Path string
}

func (e ErrBootstrapExists) Error() string {
	return fmt.Sprintf("bootstrap file already exists: %s", e.Path)
}

// UserFacingError converts an error into a user-friendly message.
func UserFacingError(err error) string {
	return domain.UserFacingError(err)
}
