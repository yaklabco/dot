package domain

import (
	"fmt"
	"strings"
)

// Domain Errors

// ErrInvalidPath indicates a path failed validation.
type ErrInvalidPath struct {
	Path   string
	Reason string
}

func (e ErrInvalidPath) Error() string {
	return fmt.Sprintf("invalid path %q: %s", e.Path, e.Reason)
}

// ErrPackageNotFound indicates a requested package does not exist.
type ErrPackageNotFound struct {
	Package string
}

func (e ErrPackageNotFound) Error() string {
	return fmt.Sprintf("package %q not found", e.Package)
}

// ErrConflict indicates a conflict that prevents an operation.
type ErrConflict struct {
	Path   string
	Reason string
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("conflict at %q: %s", e.Path, e.Reason)
}

// ErrCyclicDependency indicates a circular dependency in operations.
type ErrCyclicDependency struct {
	Cycle []string
}

func (e ErrCyclicDependency) Error() string {
	return fmt.Sprintf("cyclic dependency detected: %s", strings.Join(e.Cycle, " -> "))
}

// Infrastructure Errors

// ErrFilesystemOperation indicates a filesystem operation failed.
type ErrFilesystemOperation struct {
	Operation string
	Path      string
	Err       error
}

func (e ErrFilesystemOperation) Error() string {
	return fmt.Sprintf("filesystem operation %q failed at %q: %v", e.Operation, e.Path, e.Err)
}

func (e ErrFilesystemOperation) Unwrap() error {
	return e.Err
}

// ErrPermissionDenied indicates insufficient permissions for an operation.
type ErrPermissionDenied struct {
	Path      string
	Operation string
}

func (e ErrPermissionDenied) Error() string {
	return fmt.Sprintf("permission denied: cannot %s %q", e.Operation, e.Path)
}

// Executor Errors

// ErrEmptyPlan indicates an attempt to execute a plan with no operations.
type ErrEmptyPlan struct{}

func (e ErrEmptyPlan) Error() string {
	return "cannot execute empty plan"
}

// ErrExecutionCancelled indicates execution was cancelled via context.
type ErrExecutionCancelled struct {
	Executed int
	Skipped  int
}

func (e ErrExecutionCancelled) Error() string {
	return fmt.Sprintf("execution cancelled: %d operations completed, %d skipped", e.Executed, e.Skipped)
}

// ErrExecutionFailed indicates one or more operations failed during execution.
type ErrExecutionFailed struct {
	Executed   int
	Failed     int
	RolledBack int
	Errors     []error
}

func (e ErrExecutionFailed) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "execution failed: %d succeeded, %d failed", e.Executed, e.Failed)
	if e.RolledBack > 0 {
		fmt.Fprintf(&b, ", %d rolled back", e.RolledBack)
	}
	if len(e.Errors) > 0 {
		fmt.Fprintf(&b, "\nerrors:\n")
		for i, err := range e.Errors {
			fmt.Fprintf(&b, "  %d. %v\n", i+1, err)
		}
	}
	return b.String()
}

// Unwrap returns the underlying errors.
func (e ErrExecutionFailed) Unwrap() []error {
	return e.Errors
}

// ErrSourceNotFound indicates an operation source file does not exist.
type ErrSourceNotFound struct {
	Path string
}

func (e ErrSourceNotFound) Error() string {
	// Detect if path looks like it was resolved from target vs pwd
	if strings.Contains(e.Path, "/") {
		return fmt.Sprintf("source does not exist: %q (paths without ./ or ../ are resolved from target directory)", e.Path)
	}
	return fmt.Sprintf("source does not exist: %q", e.Path)
}

// ErrParentNotFound indicates a parent directory does not exist.
type ErrParentNotFound struct {
	Path string
}

func (e ErrParentNotFound) Error() string {
	return fmt.Sprintf("parent directory does not exist: %q", e.Path)
}

// ErrCheckpointNotFound indicates a checkpoint ID was not found.
type ErrCheckpointNotFound struct {
	ID string
}

func (e ErrCheckpointNotFound) Error() string {
	return fmt.Sprintf("checkpoint not found: %q", e.ID)
}

// ErrNotImplemented indicates functionality is not yet implemented.
type ErrNotImplemented struct {
	Feature string
}

func (e ErrNotImplemented) Error() string {
	return fmt.Sprintf("not implemented: %s", e.Feature)
}

// Error Aggregation

// ErrMultiple aggregates multiple errors into one.
type ErrMultiple struct {
	Errors []error
}

func (e ErrMultiple) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d errors occurred:\n", len(e.Errors))
	for i, err := range e.Errors {
		fmt.Fprintf(&b, "  %d. %v\n", i+1, err)
	}
	return b.String()
}

// Unwrap returns the underlying errors for errors.Is and errors.As support.
func (e ErrMultiple) Unwrap() []error {
	return e.Errors
}

// User-Facing Error Messages

// UserFacingError converts an error into a user-friendly message.
// Removes technical jargon and provides actionable information.
func UserFacingError(err error) string {
	switch e := err.(type) {
	case ErrPackageNotFound:
		return fmt.Sprintf("Package %q not found. Check that the package exists in your package directory.", e.Package)

	case ErrInvalidPath:
		return fmt.Sprintf("Invalid path %q: %s", e.Path, e.Reason)

	case ErrConflict:
		return fmt.Sprintf("Cannot proceed: conflict at %q\n%s", e.Path, e.Reason)

	case ErrCyclicDependency:
		return fmt.Sprintf("Circular dependency detected in operations: %s", strings.Join(e.Cycle, " → "))

	case ErrFilesystemOperation:
		return fmt.Sprintf("Failed to %s: %v", e.Operation, e.Err)

	case ErrPermissionDenied:
		return fmt.Sprintf("Permission denied: cannot %s %q\nCheck file permissions and try again.", e.Operation, e.Path)

	case ErrEmptyPlan:
		return "Cannot execute empty plan. Ensure the plan contains operations."

	case ErrExecutionCancelled:
		return fmt.Sprintf("Execution was cancelled: %d operations completed, %d skipped.", e.Executed, e.Skipped)

	case ErrExecutionFailed:
		return fmt.Sprintf("Execution failed: %d operations succeeded, %d failed.\nRolled back %d operations.", e.Executed, e.Failed, e.RolledBack)

	case ErrSourceNotFound:
		return fmt.Sprintf("Source file not found: %q\nEnsure the file exists before creating a link.", e.Path)

	case ErrParentNotFound:
		return fmt.Sprintf("Parent directory not found: %q\nCreate the parent directory first.", e.Path)

	case ErrMultiple:
		if len(e.Errors) == 1 {
			return UserFacingError(e.Errors[0])
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Multiple errors occurred:\n")
		for i, subErr := range e.Errors {
			fmt.Fprintf(&b, "%d. %s\n", i+1, UserFacingError(subErr))
		}
		return b.String()

	default:
		return err.Error()
	}
}
