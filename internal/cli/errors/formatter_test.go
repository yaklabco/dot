package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

func TestNewFormatter(t *testing.T) {
	f := NewFormatter(true, 1)
	require.NotNil(t, f)
	assert.True(t, f.colorEnabled)
	assert.Equal(t, 1, f.verbosity)
	assert.Greater(t, f.width, 0)
}

func TestFormatter_Format_InvalidPath(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrInvalidPath{
		Path:   "/invalid/path",
		Reason: "contains invalid characters",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Invalid Path")
	assert.Contains(t, result, "/invalid/path")
	assert.Contains(t, result, "contains invalid characters")
	assert.Contains(t, result, "Suggestions:")
}

func TestFormatter_Format_PackageNotFound(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrPackageNotFound{
		Package: "vim",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Package Not Found")
	assert.Contains(t, result, "vim")
	assert.Contains(t, result, "Suggestions:")
	assert.Contains(t, result, "dot list")
}

func TestFormatter_Format_Conflict(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrConflict{
		Path:   "/home/user/.vimrc",
		Reason: "file already exists",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Conflict Detected")
	assert.Contains(t, result, "/home/user/.vimrc")
	assert.Contains(t, result, "file already exists")
	assert.Contains(t, result, "Suggestions:")
}

func TestFormatter_Format_CyclicDependency(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrCyclicDependency{
		Cycle: []string{"A", "B", "C", "A"},
	}

	result := f.Format(err)
	assert.Contains(t, result, "Circular Dependency Detected")
	assert.Contains(t, result, "A -> B -> C -> A")
	assert.Contains(t, result, "Suggestions:")
}

func TestFormatter_Format_PermissionDenied(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrPermissionDenied{
		Path:      "/root/file",
		Operation: "write",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Permission Denied")
	assert.Contains(t, result, "write")
	assert.Contains(t, result, "/root/file")
	assert.Contains(t, result, "Suggestions:")
}

func TestFormatter_Format_FilesystemOperation(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrFilesystemOperation{
		Operation: "symlink",
		Path:      "/home/user/.vimrc",
		Err:       errors.New("no such file or directory"),
	}

	result := f.Format(err)
	assert.Contains(t, result, "Filesystem Operation Failed")
	assert.Contains(t, result, "symlink")
	assert.Contains(t, result, "/home/user/.vimrc")
	assert.Contains(t, result, "no such file or directory")
}

func TestFormatter_Format_EmptyPlan(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrEmptyPlan{}

	result := f.Format(err)
	assert.Contains(t, result, "Empty Plan")
	assert.Contains(t, result, "No operations to execute")
}

func TestFormatter_Format_ExecutionFailed(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrExecutionFailed{
		Executed:   5,
		Failed:     2,
		RolledBack: 2,
		Errors: []error{
			errors.New("error 1"),
			errors.New("error 2"),
		},
	}

	result := f.Format(err)
	assert.Contains(t, result, "Execution Failed")
	assert.Contains(t, result, "5 operations succeeded")
	assert.Contains(t, result, "2 operations failed")
	assert.Contains(t, result, "2 operations rolled back")
}

func TestFormatter_Format_SourceNotFound(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrSourceNotFound{
		Path: "/packages/vim/vimrc",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Source Not Found")
	assert.Contains(t, result, "/packages/vim/vimrc")
}

func TestFormatter_Format_ParentNotFound(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrParentNotFound{
		Path: "/home/user/.config/nvim/init.vim",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Parent Directory Not Found")
	assert.Contains(t, result, "/home/user/.config/nvim/init.vim")
}

func TestFormatter_Format_CheckpointNotFound(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrCheckpointNotFound{
		ID: "checkpoint-123",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Checkpoint Not Found")
	assert.Contains(t, result, "checkpoint-123")
	assert.Contains(t, result, "Verify checkpoint ID or list available checkpoints")
}

func TestFormatter_Format_NotImplemented(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrNotImplemented{
		Feature: "parallel execution",
	}

	result := f.Format(err)
	assert.Contains(t, result, "Not Implemented")
	assert.Contains(t, result, "parallel execution")
	assert.Contains(t, result, "Use an alternative supported operation or file an issue")
}

func TestFormatter_Format_MultipleErrors(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrMultiple{
		Errors: []error{
			domain.ErrPackageNotFound{Package: "vim"},
			domain.ErrPackageNotFound{Package: "tmux"},
			domain.ErrPackageNotFound{Package: "zsh"},
		},
	}

	result := f.Format(err)
	assert.Contains(t, result, "Multiple errors occurred (3 total)")
	assert.Contains(t, result, "1.")
	assert.Contains(t, result, "2.")
	assert.Contains(t, result, "3.")
	assert.Contains(t, result, "vim")
	assert.Contains(t, result, "tmux")
	assert.Contains(t, result, "zsh")
}

func TestFormatter_Format_MultipleErrors_SingleError(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrMultiple{
		Errors: []error{
			domain.ErrPackageNotFound{Package: "vim"},
		},
	}

	result := f.Format(err)
	// Should format as single error
	assert.Contains(t, result, "Package Not Found")
	assert.Contains(t, result, "vim")
	assert.NotContains(t, result, "Multiple errors")
}

func TestFormatter_Format_MultipleErrors_Empty(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrMultiple{
		Errors: []error{},
	}

	result := f.Format(err)
	assert.Equal(t, "no errors", result)
}

func TestFormatter_Format_UnknownError(t *testing.T) {
	f := NewFormatter(false, 0)
	err := errors.New("unknown error type")

	result := f.Format(err)
	assert.Equal(t, "unknown error type", result)
}

func TestFormatter_Format_NilError(t *testing.T) {
	f := NewFormatter(false, 0)
	result := f.Format(nil)
	assert.Equal(t, "", result)
}

func TestFormatter_FormatWithContext(t *testing.T) {
	f := NewFormatter(false, 0)
	ctx := ErrorContext{
		Command:   "dot manage",
		Arguments: []string{"vim"},
		Config: ConfigSummary{
			PackageDir: "/home/user/dotfiles",
			TargetDir:  "/home/user",
			DryRun:     false,
			Verbose:    1,
		},
	}

	err := domain.ErrPackageNotFound{Package: "vim"}
	result := f.FormatWithContext(err, ctx)

	assert.Contains(t, result, "Package Not Found")
	assert.Contains(t, result, "vim")
	// Context should be used by suggestion engine
	assert.Contains(t, result, "/home/user/dotfiles")
}

func TestFormatter_Format_WithColor(t *testing.T) {
	f := NewFormatter(true, 0)
	err := domain.ErrInvalidPath{
		Path:   "/invalid",
		Reason: "test",
	}

	result := f.Format(err)
	// Should contain ANSI color codes
	assert.Contains(t, result, "\033[")
	assert.Contains(t, result, "Invalid Path")
}

func TestFormatter_Format_WithoutColor(t *testing.T) {
	f := NewFormatter(false, 0)
	err := domain.ErrInvalidPath{
		Path:   "/invalid",
		Reason: "test",
	}

	result := f.Format(err)
	// Should not contain ANSI color codes
	assert.NotContains(t, result, "\033[")
	assert.Contains(t, result, "Invalid Path")
}

func TestFormatter_Format_VerbosityLevels(t *testing.T) {
	tests := []struct {
		name      string
		verbosity int
	}{
		{"level 0", 0},
		{"level 1", 1},
		{"level 2", 2},
		{"level 3", 3},
	}

	err := domain.ErrPackageNotFound{Package: "vim"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(false, tt.verbosity)
			result := f.Format(err)
			assert.Contains(t, result, "Package Not Found")
			assert.Contains(t, result, "vim")
		})
	}
}

func TestFormatter_Format_WideTerminal(t *testing.T) {
	f := &Formatter{
		colorEnabled: false,
		verbosity:    0,
		width:        200,
	}

	// Create an error with a very long description to test wrapping
	longReason := "contains many invalid characters and other problems that would cause wrapping on narrow terminals but should remain on fewer lines with a wide terminal width"
	err := domain.ErrInvalidPath{
		Path:   "/a/very/long/path/that/would/normally/wrap/on/narrow/terminals",
		Reason: longReason,
	}

	result := f.Format(err)
	assert.Contains(t, result, "Invalid Path")
	assert.Contains(t, result, longReason)

	// With wide terminal, output should be properly formatted
	assert.NotEmpty(t, result)
}

func TestFormatter_Format_NarrowTerminal(t *testing.T) {
	f := &Formatter{
		colorEnabled: false,
		verbosity:    0,
		width:        40,
	}

	err := domain.ErrInvalidPath{
		Path:   "/a/very/long/path",
		Reason: "contains many invalid characters",
	}

	result := f.Format(err)
	lines := strings.Split(result, "\n")

	// With narrow terminal, lines should be shorter
	for _, line := range lines {
		// Strip color codes if any
		cleanLine := strings.ReplaceAll(line, "\033[", "")
		// Most lines should fit in narrow width (with some tolerance)
		if !strings.HasPrefix(strings.TrimSpace(line), "•") {
			assert.LessOrEqual(t, len(cleanLine), 80)
		}
	}
}
