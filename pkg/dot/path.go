package dot

import "github.com/jamesainslie/dot/internal/domain"

// Concrete path types re-exported from internal/domain.
// These use proper type aliases (=) and include all methods from domain types.
//
// Note: The generic Path[K PathKind] type is NOT re-exported to avoid
// Go 1.25.4 generic type alias limitations. Users should use the concrete
// types (PackagePath, TargetPath, FilePath) which work perfectly as aliases.

// PackagePath is a path to a package directory.
// Includes methods: String(), Join(), Parent(), Equals()
type PackagePath = domain.PackagePath

// TargetPath is a path to a target directory.
// Includes methods: String(), Join(), Parent(), Equals()
type TargetPath = domain.TargetPath

// FilePath is a path to a file or directory within a package.
// Includes methods: String(), Join(), Parent(), Equals()
type FilePath = domain.FilePath

// NewPackagePath creates a new package path with validation.
func NewPackagePath(s string) Result[PackagePath] {
	r := domain.NewPackagePath(s)
	return Result[PackagePath](r)
}

// NewTargetPath creates a new target path with validation.
func NewTargetPath(s string) Result[TargetPath] {
	r := domain.NewTargetPath(s)
	return Result[TargetPath](r)
}

// NewFilePath creates a new file path with validation.
func NewFilePath(s string) Result[FilePath] {
	r := domain.NewFilePath(s)
	return Result[FilePath](r)
}

// MustParsePath creates a FilePath from a string, panicking on error.
// This function is intended for use in tests only where paths are known to be valid.
// Production code should use NewFilePath which returns a Result for proper error handling.
func MustParsePath(s string) FilePath {
	return domain.MustParsePath(s)
}

// MustParseTargetPath creates a TargetPath from a string, panicking on error.
// This function is intended for use in tests only where paths are known to be valid.
// Production code should use NewTargetPath which returns a Result for proper error handling.
//
// Panic is appropriate here because:
// - Function is only used in test code with hardcoded, known-valid paths
// - Panicking on test setup errors fails fast and clearly indicates test bugs
// - Test failures from panic are easier to debug than silent errors
func MustParseTargetPath(s string) TargetPath {
	r := domain.NewTargetPath(s)
	if !r.IsOk() {
		panic(r.UnwrapErr())
	}
	return r.Unwrap()
}
