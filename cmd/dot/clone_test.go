package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestCloneCommand_Flags(t *testing.T) {
	cmd := newCloneCommand()

	t.Run("has profile flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("profile")
		assert.NotNil(t, flag)
		assert.Equal(t, "string", flag.Value.Type())
	})

	t.Run("has interactive flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("interactive")
		assert.NotNil(t, flag)
		assert.Equal(t, "bool", flag.Value.Type())
	})

	t.Run("has force flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("force")
		assert.NotNil(t, flag)
		assert.Equal(t, "bool", flag.Value.Type())
	})

	t.Run("has branch flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("branch")
		assert.NotNil(t, flag)
		assert.Equal(t, "string", flag.Value.Type())
	})
}

func TestCloneCommand_Args(t *testing.T) {
	cmd := newCloneCommand()

	t.Run("requires exactly one argument", func(t *testing.T) {
		// No arguments
		err := cmd.Args(cmd, []string{})
		assert.Error(t, err)

		// Two arguments
		err = cmd.Args(cmd, []string{"url1", "url2"})
		assert.Error(t, err)

		// One argument (valid)
		err = cmd.Args(cmd, []string{"https://github.com/user/repo"})
		assert.NoError(t, err)
	})
}

func TestCloneCommand_Help(t *testing.T) {
	cmd := newCloneCommand()

	t.Run("has short description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, strings.ToLower(cmd.Short), "clone")
	})

	t.Run("has long description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Long)
		assert.Contains(t, cmd.Long, "authentication")
		assert.Contains(t, cmd.Long, "bootstrap")
	})

	t.Run("has usage examples", func(t *testing.T) {
		assert.Contains(t, cmd.Long, "Examples:")
		assert.Contains(t, cmd.Long, "dot clone")
	})
}

func TestFormatCloneError_PackageDirNotEmpty(t *testing.T) {
	err := dot.ErrPackageDirNotEmpty{Path: "/path/to/packages"}
	formatted := formatCloneError(err)

	errMsg := formatted.Error()
	assert.Contains(t, errMsg, "not empty")
	assert.Contains(t, errMsg, "--force")
}

func TestFormatCloneError_BootstrapNotFound(t *testing.T) {
	err := dot.ErrBootstrapNotFound{Path: "/path/.dotbootstrap.yaml"}
	formatted := formatCloneError(err)

	errMsg := formatted.Error()
	assert.Contains(t, errMsg, "not found")
	assert.Contains(t, errMsg, "properly cloned")
}

func TestFormatCloneError_InvalidBootstrap(t *testing.T) {
	err := dot.ErrInvalidBootstrap{Reason: "invalid YAML"}
	formatted := formatCloneError(err)

	errMsg := formatted.Error()
	assert.Contains(t, errMsg, "invalid")
	assert.Contains(t, errMsg, ".dotbootstrap.yaml")
}

func TestFormatCloneError_AuthFailed(t *testing.T) {
	err := dot.ErrAuthFailed{Cause: assert.AnError}
	formatted := formatCloneError(err)

	errMsg := formatted.Error()
	assert.Contains(t, errMsg, "authentication")
	assert.Contains(t, errMsg, "GITHUB_TOKEN")
	assert.Contains(t, errMsg, "SSH")
}

func TestFormatCloneError_CloneFailed(t *testing.T) {
	err := dot.ErrCloneFailed{URL: "https://github.com/user/repo", Cause: assert.AnError}
	formatted := formatCloneError(err)

	errMsg := formatted.Error()
	assert.Contains(t, errMsg, "clone")
	assert.Contains(t, errMsg, "URL")
	assert.Contains(t, errMsg, "accessible")
}

func TestFormatCloneError_ProfileNotFound(t *testing.T) {
	err := dot.ErrProfileNotFound{Profile: "minimal"}
	formatted := formatCloneError(err)

	errMsg := formatted.Error()
	assert.Contains(t, errMsg, "profile")
	assert.Contains(t, errMsg, ".dotbootstrap.yaml")
}

func TestFormatCloneError_GenericError(t *testing.T) {
	err := assert.AnError
	formatted := formatCloneError(err)

	// Generic errors are returned as-is
	assert.Equal(t, err, formatted)
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL with .git suffix",
			url:      "https://github.com/user/my-dotfiles.git",
			expected: "my-dotfiles",
		},
		{
			name:     "HTTPS URL without .git suffix",
			url:      "https://github.com/user/my-dotfiles",
			expected: "my-dotfiles",
		},
		{
			name:     "SSH URL with .git suffix",
			url:      "git@github.com:user/my-dotfiles.git",
			expected: "my-dotfiles",
		},
		{
			name:     "SSH URL without .git suffix",
			url:      "git@github.com:user/my-dotfiles",
			expected: "my-dotfiles",
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/user/dotfiles.git",
			expected: "dotfiles",
		},
		{
			name:     "GitLab SSH URL",
			url:      "git@gitlab.com:user/dotfiles.git",
			expected: "dotfiles",
		},
		{
			name:     "URL with nested path",
			url:      "https://github.com/org/team/repo.git",
			expected: "repo",
		},
		{
			name:     "Simple repo name",
			url:      "my-dotfiles",
			expected: "my-dotfiles",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "dotfiles",
		},
		{
			name:     "URL with query parameters",
			url:      "https://github.com/user/repo.git?ref=main",
			expected: "repo",
		},
		{
			name:     "Bitbucket SSH URL",
			url:      "git@bitbucket.org:user/my-config.git",
			expected: "my-config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoName(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCloneCommand_Integration(t *testing.T) {
	// Integration tests requiring actual repository would go here
	// Skipped in unit tests
	t.Skip("requires integration test setup with test repository")
}
