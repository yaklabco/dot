package domain

import (
	"path/filepath"
	"strings"
)

// PathKind is a marker interface for phantom type parameters.
// Implementations exist only at compile time to enforce type safety.
type PathKind interface {
	pathKind()
}

// PackageDirKind marks paths pointing to package directories.
type PackageDirKind struct{}

func (PackageDirKind) pathKind() {}

// TargetDirKind marks paths pointing to target directories.
type TargetDirKind struct{}

func (TargetDirKind) pathKind() {}

// FileDirKind marks paths pointing to file directories.
type FileDirKind struct{}

func (FileDirKind) pathKind() {}

// Path represents a filesystem path with phantom typing for compile-time safety.
// The type parameter K ensures paths of different kinds cannot be mixed accidentally.
type Path[K PathKind] struct {
	path string
}

// PackagePath is a path to a package directory.
type PackagePath = Path[PackageDirKind]

// TargetPath is a path to a target directory.
type TargetPath = Path[TargetDirKind]

// FilePath is a path to a file or directory within a package.
type FilePath = Path[FileDirKind]

// NewPackagePath creates a new package path with validation.
// Returns error if path is not absolute or contains traversal sequences.
func NewPackagePath(s string) Result[PackagePath] {
	// Clean path first to normalize format
	cleaned := clean(s)

	validators := []PathValidator{
		&NonEmptyPathValidator{},
		&AbsolutePathValidator{},
		&TraversalFreeValidator{},
	}

	if err := ValidateWithValidators(cleaned, validators); err != nil {
		return Err[PackagePath](err)
	}

	return Ok(Path[PackageDirKind]{path: cleaned})
}

// NewTargetPath creates a new target path with validation.
// Returns error if path is not absolute or contains traversal sequences.
func NewTargetPath(s string) Result[TargetPath] {
	// Clean path first to normalize format
	cleaned := clean(s)

	validators := []PathValidator{
		&NonEmptyPathValidator{},
		&AbsolutePathValidator{},
		&TraversalFreeValidator{},
	}

	if err := ValidateWithValidators(cleaned, validators); err != nil {
		return Err[TargetPath](err)
	}

	return Ok(Path[TargetDirKind]{path: cleaned})
}

// NewFilePath creates a new file path with validation.
// Returns error if path is not absolute or contains traversal sequences.
func NewFilePath(s string) Result[FilePath] {
	// Clean path first to normalize format
	cleaned := clean(s)

	validators := []PathValidator{
		&NonEmptyPathValidator{},
		&AbsolutePathValidator{},
		&TraversalFreeValidator{},
	}

	if err := ValidateWithValidators(cleaned, validators); err != nil {
		return Err[FilePath](err)
	}

	return Ok(Path[FileDirKind]{path: cleaned})
}

// String returns the string representation of the path.
func (p Path[K]) String() string {
	return p.path
}

// JoinSafe appends a path element with validation to prevent traversal attacks.
// Returns an error if the resulting path would escape the base path.
func (p Path[K]) JoinSafe(elem string) Result[Path[K]] {
	cleanedElem := filepath.Clean(elem)

	// Check for traversal sequences in the cleaned element
	if strings.HasPrefix(cleanedElem, "..") || strings.Contains(cleanedElem, string(filepath.Separator)+"..") {
		return Err[Path[K]](ErrInvalidPath{
			Path:   elem,
			Reason: "path contains traversal sequence",
		})
	}

	joined := filepath.Join(p.path, elem)
	cleanedJoined := filepath.Clean(joined)

	// Verify the joined path stays within the base directory
	basePath := filepath.Clean(p.path)
	if !strings.HasPrefix(cleanedJoined, basePath) {
		return Err[Path[K]](ErrInvalidPath{
			Path:   elem,
			Reason: "path escapes base directory",
		})
	}

	return Ok(Path[K]{path: cleanedJoined})
}

// Join appends a path component, returning a FilePath.
//
// Deprecated: Use JoinSafe for user-provided paths to prevent path traversal attacks.
func (p Path[K]) Join(elem string) Path[K] {
	joined := filepath.Join(p.path, elem)
	return Path[K]{path: joined}
}

// Parent returns the parent directory of this path.
func (p Path[K]) Parent() Result[Path[K]] {
	parent := filepath.Dir(p.path)
	if parent == p.path {
		return Err[Path[K]](ErrInvalidPath{Path: p.path, Reason: "path has no parent"})
	}
	return Ok(Path[K]{path: parent})
}

// Equals checks if two paths are equal.
func (p Path[K]) Equals(other Path[K]) bool {
	return p.path == other.path
}

// clean normalizes a path by removing redundant separators and resolving dots.
func clean(path string) string {
	cleaned := filepath.Clean(path)
	// Remove trailing slash except for root
	if len(cleaned) > 1 && strings.HasSuffix(cleaned, string(filepath.Separator)) {
		cleaned = strings.TrimSuffix(cleaned, string(filepath.Separator))
	}
	return cleaned
}
