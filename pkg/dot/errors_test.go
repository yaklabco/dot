package dot_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestErrInvalidPath(t *testing.T) {
	err := dot.ErrInvalidPath{
		Path:   "/some/path",
		Reason: "must be absolute",
	}

	assert.Contains(t, err.Error(), "/some/path")
	assert.Contains(t, err.Error(), "must be absolute")
}

func TestErrPackageNotFound(t *testing.T) {
	err := dot.ErrPackageNotFound{
		Package: "vim",
	}

	assert.Contains(t, err.Error(), "vim")
	assert.Contains(t, err.Error(), "not found")
}

func TestErrConflict(t *testing.T) {
	err := dot.ErrConflict{
		Path:   "/home/user/.vimrc",
		Reason: "file already exists",
	}

	assert.Contains(t, err.Error(), "/home/user/.vimrc")
	assert.Contains(t, err.Error(), "file already exists")
}

func TestErrCyclicDependency(t *testing.T) {
	err := dot.ErrCyclicDependency{
		Cycle: []string{"a", "b", "c", "a"},
	}

	msg := err.Error()
	assert.Contains(t, msg, "a")
	assert.Contains(t, msg, "b")
	assert.Contains(t, msg, "c")
	assert.Contains(t, msg, "cyclic")
}

func TestErrFilesystemOperation(t *testing.T) {
	inner := errors.New("permission denied")
	err := dot.ErrFilesystemOperation{
		Operation: "create symlink",
		Path:      "/home/user/.vimrc",
		Err:       inner,
	}

	assert.Contains(t, err.Error(), "create symlink")
	assert.Contains(t, err.Error(), "/home/user/.vimrc")
	assert.ErrorIs(t, err, inner)
}

func TestErrPermissionDenied(t *testing.T) {
	err := dot.ErrPermissionDenied{
		Path:      "/root/.vimrc",
		Operation: "write",
	}

	assert.Contains(t, err.Error(), "/root/.vimrc")
	assert.Contains(t, err.Error(), "write")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestErrMultiple(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	multi := dot.ErrMultiple{
		Errors: []error{err1, err2, err3},
	}

	msg := multi.Error()
	assert.Contains(t, msg, "3 errors")
	assert.Contains(t, msg, "error 1")
	assert.Contains(t, msg, "error 2")
	assert.Contains(t, msg, "error 3")
}

func TestErrMultipleUnwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	multi := dot.ErrMultiple{
		Errors: []error{err1, err2},
	}

	unwrapped := multi.Unwrap()
	assert.Len(t, unwrapped, 2)
	assert.Equal(t, err1, unwrapped[0])
	assert.Equal(t, err2, unwrapped[1])
}

func TestUserFacingErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name: "ErrPackageNotFound",
			err: dot.ErrPackageNotFound{
				Package: "vim",
			},
			contains: []string{"vim", "not found"},
		},
		{
			name: "ErrInvalidPath",
			err: dot.ErrInvalidPath{
				Path:   "relative/path",
				Reason: "must be absolute",
			},
			contains: []string{"relative/path", "absolute"},
		},
		{
			name: "ErrConflict",
			err: dot.ErrConflict{
				Path:   "/home/user/.vimrc",
				Reason: "file exists",
			},
			contains: []string{".vimrc", "file exists"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := dot.UserFacingError(tt.err)
			for _, expected := range tt.contains {
				assert.Contains(t, msg, expected)
			}
		})
	}
}

func TestErrEmptyPlan(t *testing.T) {
	err := dot.ErrEmptyPlan{}
	assert.Equal(t, "cannot execute empty plan", err.Error())
}

func TestErrExecutionFailed(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := dot.ErrExecutionFailed{
			Executed: 5,
			Failed:   2,
		}
		msg := err.Error()
		assert.Contains(t, msg, "5 succeeded")
		assert.Contains(t, msg, "2 failed")
	})

	t.Run("with rollback", func(t *testing.T) {
		err := dot.ErrExecutionFailed{
			Executed:   3,
			Failed:     1,
			RolledBack: 2,
		}
		msg := err.Error()
		assert.Contains(t, msg, "2 rolled back")
	})

	t.Run("with errors", func(t *testing.T) {
		err := dot.ErrExecutionFailed{
			Executed: 1,
			Failed:   2,
			Errors: []error{
				errors.New("first error"),
				errors.New("second error"),
			},
		}
		msg := err.Error()
		assert.Contains(t, msg, "first error")
		assert.Contains(t, msg, "second error")

		unwrapped := err.Unwrap()
		assert.Len(t, unwrapped, 2)
	})
}

func TestErrSourceNotFound(t *testing.T) {
	err := dot.ErrSourceNotFound{Path: "/missing/file"}
	msg := err.Error()
	assert.Contains(t, msg, "/missing/file")
	assert.Contains(t, msg, "source does not exist")
}

func TestErrParentNotFound(t *testing.T) {
	err := dot.ErrParentNotFound{Path: "/missing/parent"}
	msg := err.Error()
	assert.Contains(t, msg, "/missing/parent")
	assert.Contains(t, msg, "parent directory")
}

func TestErrCheckpointNotFound(t *testing.T) {
	err := dot.ErrCheckpointNotFound{ID: "checkpoint-123"}
	msg := err.Error()
	assert.Contains(t, msg, "checkpoint-123")
	assert.Contains(t, msg, "not found")
}

func TestErrNotImplemented(t *testing.T) {
	err := dot.ErrNotImplemented{Feature: "advanced feature"}
	msg := err.Error()
	assert.Contains(t, msg, "advanced feature")
	assert.Contains(t, msg, "not implemented")
}

func TestUserFacingErrorComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "ErrCyclicDependency",
			err:      dot.ErrCyclicDependency{Cycle: []string{"a", "b", "a"}},
			contains: []string{"Circular dependency", "a", "b"},
		},
		{
			name:     "ErrFilesystemOperation",
			err:      dot.ErrFilesystemOperation{Operation: "write", Path: "/file", Err: errors.New("permission denied")},
			contains: []string{"Failed to write", "permission denied"},
		},
		{
			name:     "ErrPermissionDenied",
			err:      dot.ErrPermissionDenied{Path: "/restricted", Operation: "read"},
			contains: []string{"Permission denied", "/restricted"},
		},
		{
			name:     "ErrMultiple",
			err:      dot.ErrMultiple{Errors: []error{errors.New("err1"), errors.New("err2")}},
			contains: []string{"Multiple errors", "err1", "err2"},
		},
		{
			name:     "ErrEmptyPlan",
			err:      dot.ErrEmptyPlan{},
			contains: []string{"empty plan"},
		},
		{
			name:     "ErrExecutionFailed",
			err:      dot.ErrExecutionFailed{Executed: 3, Failed: 2},
			contains: []string{"Execution failed", "3 operations succeeded", "2 failed"},
		},
		{
			name:     "ErrSourceNotFound",
			err:      dot.ErrSourceNotFound{Path: "/src"},
			contains: []string{"Source file not found", "/src"},
		},
		{
			name:     "ErrParentNotFound",
			err:      dot.ErrParentNotFound{Path: "/parent"},
			contains: []string{"parent", "/parent"},
		},
		{
			name:     "ErrCheckpointNotFound",
			err:      dot.ErrCheckpointNotFound{ID: "chk1"},
			contains: []string{"checkpoint", "chk1"},
		},
		{
			name:     "ErrNotImplemented",
			err:      dot.ErrNotImplemented{Feature: "feat"},
			contains: []string{"not implemented", "feat"},
		},
		{
			name:     "generic error",
			err:      errors.New("generic error message"),
			contains: []string{"generic error message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := dot.UserFacingError(tt.err)
			for _, contain := range tt.contains {
				assert.Contains(t, msg, contain, "expected message to contain %q", contain)
			}
		})
	}
}
