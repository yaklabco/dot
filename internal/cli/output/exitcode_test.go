package output

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

func TestGetExitCode_Success(t *testing.T) {
	code := GetExitCode(nil)
	assert.Equal(t, ExitSuccess, code)
}

func TestGetExitCode_InvalidPath(t *testing.T) {
	err := domain.ErrInvalidPath{
		Path:   "/invalid",
		Reason: "test",
	}
	code := GetExitCode(err)
	assert.Equal(t, ExitInvalidArguments, code)
}

func TestGetExitCode_PackageNotFound(t *testing.T) {
	err := domain.ErrPackageNotFound{
		Package: "vim",
	}
	code := GetExitCode(err)
	assert.Equal(t, ExitPackageNotFound, code)
}

func TestGetExitCode_Conflict(t *testing.T) {
	err := domain.ErrConflict{
		Path:   "/conflict",
		Reason: "test",
	}
	code := GetExitCode(err)
	assert.Equal(t, ExitConflict, code)
}

func TestGetExitCode_PermissionDenied(t *testing.T) {
	err := domain.ErrPermissionDenied{
		Path:      "/root",
		Operation: "write",
	}
	code := GetExitCode(err)
	assert.Equal(t, ExitPermissionDenied, code)
}

func TestGetExitCode_GeneralError(t *testing.T) {
	err := errors.New("generic error")
	code := GetExitCode(err)
	assert.Equal(t, ExitGeneralError, code)
}

func TestGetExitCode_WrappedError(t *testing.T) {
	innerErr := domain.ErrPackageNotFound{Package: "vim"}
	wrappedErr := errors.Join(innerErr, errors.New("additional context"))

	code := GetExitCode(wrappedErr)
	assert.Equal(t, ExitPackageNotFound, code)
}

func TestExitCodeConstants(t *testing.T) {
	// Verify exit codes are distinct
	codes := []int{
		ExitSuccess,
		ExitGeneralError,
		ExitInvalidArguments,
		ExitConflict,
		ExitPermissionDenied,
		ExitPackageNotFound,
	}

	seen := make(map[int]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "exit code %d should be unique", code)
		seen[code] = true
	}
}

func TestExitCodeValues(t *testing.T) {
	assert.Equal(t, 0, ExitSuccess)
	assert.Equal(t, 1, ExitGeneralError)
	assert.Equal(t, 2, ExitInvalidArguments)
	assert.Equal(t, 3, ExitConflict)
	assert.Equal(t, 4, ExitPermissionDenied)
	assert.Equal(t, 5, ExitPackageNotFound)
}
