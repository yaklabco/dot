package output

import (
	"errors"

	"github.com/yaklabco/dot/internal/domain"
)

// Exit codes for different error types.
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitInvalidArguments = 2
	ExitConflict         = 3
	ExitPermissionDenied = 4
	ExitPackageNotFound  = 5
)

// GetExitCode returns the appropriate exit code for an error.
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Check for domain errors
	var invalidPath domain.ErrInvalidPath
	if errors.As(err, &invalidPath) {
		return ExitInvalidArguments
	}

	var pkgNotFound domain.ErrPackageNotFound
	if errors.As(err, &pkgNotFound) {
		return ExitPackageNotFound
	}

	var conflict domain.ErrConflict
	if errors.As(err, &conflict) {
		return ExitConflict
	}

	var permDenied domain.ErrPermissionDenied
	if errors.As(err, &permDenied) {
		return ExitPermissionDenied
	}

	// Default to general error
	return ExitGeneralError
}
