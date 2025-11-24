package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

func TestSuggestionEngine_Generate_InvalidPath(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrInvalidPath{
		Path:   "/some/path",
		Reason: "invalid characters",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	assert.Contains(t, suggestions[0], "absolute paths")
}

func TestSuggestionEngine_Generate_InvalidPath_WithTilde(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrInvalidPath{
		Path:   "~/dotfiles",
		Reason: "contains tilde",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "home directory") {
			found = true
			break
		}
	}
	assert.True(t, found, "should suggest expanding tilde")
}

func TestSuggestionEngine_Generate_InvalidPath_WithDotDot(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrInvalidPath{
		Path:   "../relative/path",
		Reason: "contains relative components",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "relative path") {
			found = true
			break
		}
	}
	assert.True(t, found, "should warn about .. in path")
}

func TestSuggestionEngine_Generate_PackageNotFound(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrPackageNotFound{
		Package: "vim",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	assert.Contains(t, suggestions[0], "dot list")
}

func TestSuggestionEngine_Generate_PackageNotFound_WithContext(t *testing.T) {
	engine := SuggestionEngine{
		context: ErrorContext{
			Config: ConfigSummary{
				PackageDir: "/home/user/dotfiles",
			},
		},
	}
	err := domain.ErrPackageNotFound{
		Package: "vim",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "/home/user/dotfiles") {
			found = true
			break
		}
	}
	assert.True(t, found, "should include package directory from context")
}

func TestSuggestionEngine_Generate_Conflict(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrConflict{
		Path:   "/home/user/.vimrc",
		Reason: "file exists",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if assert.Contains(t, s, "adopt") {
			found = true
			break
		}
	}
	assert.True(t, found, "should suggest adopt command")
}

func TestSuggestionEngine_Generate_CyclicDependency(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrCyclicDependency{
		Cycle: []string{"A", "B", "C", "A"},
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if assert.Contains(t, s, "bug") {
			found = true
			break
		}
	}
	assert.True(t, found, "should mention this is a bug")
}

func TestSuggestionEngine_Generate_PermissionDenied(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrPermissionDenied{
		Path:      "/root/.vimrc",
		Operation: "write",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if assert.Contains(t, s, "permissions") {
			found = true
			break
		}
	}
	assert.True(t, found, "should suggest checking permissions")
}

func TestSuggestionEngine_Generate_FilesystemOperation(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrFilesystemOperation{
		Operation: "symlink",
		Path:      "/home/user/.vimrc",
		Err:       errors.New("no space left"),
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "disk space") {
			found = true
			break
		}
	}
	assert.True(t, found, "should suggest checking disk space")
}

func TestSuggestionEngine_Generate_ExecutionFailed(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrExecutionFailed{
		Executed:   5,
		Failed:     2,
		RolledBack: 0,
		Errors:     []error{errors.New("error 1"), errors.New("error 2")},
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	assert.Contains(t, suggestions[0], "individual error messages")
}

func TestSuggestionEngine_Generate_ExecutionFailed_WithRollback(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrExecutionFailed{
		Executed:   5,
		Failed:     2,
		RolledBack: 2,
		Errors:     []error{errors.New("error 1"), errors.New("error 2")},
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "rolled back") {
			found = true
			break
		}
	}
	assert.True(t, found, "should mention rollback")
}

func TestSuggestionEngine_Generate_ExecutionFailed_WithDryRun(t *testing.T) {
	engine := SuggestionEngine{
		context: ErrorContext{
			Config: ConfigSummary{
				DryRun: true,
			},
		},
	}
	err := domain.ErrExecutionFailed{
		Executed: 0,
		Failed:   5,
		Errors:   []error{errors.New("error 1")},
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "dry-run") {
			found = true
			break
		}
	}
	assert.True(t, found, "should suggest removing --dry-run")
}

func TestSuggestionEngine_Generate_SourceNotFound(t *testing.T) {
	engine := SuggestionEngine{}
	err := domain.ErrSourceNotFound{
		Path: "/packages/vim/vimrc",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if assert.Contains(t, s, "source file") {
			found = true
			break
		}
	}
	assert.True(t, found, "should suggest verifying source file")
}

func TestSuggestionEngine_Generate_SourceNotFound_WithContext(t *testing.T) {
	engine := SuggestionEngine{
		context: ErrorContext{
			Config: ConfigSummary{
				PackageDir: "/home/user/dotfiles",
			},
		},
	}
	err := domain.ErrSourceNotFound{
		Path: "/packages/vim/vimrc",
	}

	suggestions := engine.Generate(err)

	assert.NotEmpty(t, suggestions)
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "/home/user/dotfiles") {
			found = true
			break
		}
	}
	assert.True(t, found, "should include package directory from context")
}

func TestSuggestionEngine_Generate_UnknownError(t *testing.T) {
	engine := SuggestionEngine{}
	err := errors.New("unknown error type")

	suggestions := engine.Generate(err)

	assert.Empty(t, suggestions)
}

func TestSuggestionEngine_Generate_NilError(t *testing.T) {
	engine := SuggestionEngine{}
	suggestions := engine.Generate(nil)

	assert.Nil(t, suggestions)
}

func TestSuggestionEngine_Prioritize(t *testing.T) {
	engine := SuggestionEngine{}
	suggestions := []string{
		"suggestion one",
		"suggestion two",
		"suggestion three",
	}

	result := engine.Prioritize(suggestions)

	// Currently returns as-is
	assert.Equal(t, suggestions, result)
	assert.Equal(t, 3, len(result))
}

func TestSuggestionEngine_Prioritize_Empty(t *testing.T) {
	engine := SuggestionEngine{}
	suggestions := []string{}

	result := engine.Prioritize(suggestions)

	assert.Empty(t, result)
}

func TestSuggestionEngine_Prioritize_Nil(t *testing.T) {
	engine := SuggestionEngine{}
	result := engine.Prioritize(nil)

	assert.Nil(t, result)
}

func TestSuggestionEngine_ContextInfluence(t *testing.T) {
	tests := []struct {
		name    string
		context ErrorContext
		err     error
		check   func(*testing.T, []string)
	}{
		{
			name: "package not found with package dir",
			context: ErrorContext{
				Config: ConfigSummary{
					PackageDir: "/custom/packages",
				},
			},
			err: domain.ErrPackageNotFound{Package: "vim"},
			check: func(t *testing.T, suggestions []string) {
				found := false
				for _, s := range suggestions {
					if strings.Contains(s, "/custom/packages") {
						found = true
						break
					}
				}
				assert.True(t, found)
			},
		},
		{
			name: "execution failed in dry run",
			context: ErrorContext{
				Config: ConfigSummary{
					DryRun: true,
				},
			},
			err: domain.ErrExecutionFailed{
				Executed: 0,
				Failed:   1,
				Errors:   []error{errors.New("test")},
			},
			check: func(t *testing.T, suggestions []string) {
				found := false
				for _, s := range suggestions {
					if strings.Contains(s, "dry-run") {
						found = true
						break
					}
				}
				assert.True(t, found)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := SuggestionEngine{context: tt.context}
			suggestions := engine.Generate(tt.err)
			tt.check(t, suggestions)
		})
	}
}
