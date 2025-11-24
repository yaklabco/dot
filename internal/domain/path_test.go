package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

func TestNewPackagePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "absolute path",
			path:    "/home/user/.dotfiles",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "dotfiles",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.NewPackagePath(tt.path)
			if tt.wantErr {
				assert.True(t, result.IsErr())
			} else {
				assert.True(t, result.IsOk())
				path := result.Unwrap()
				assert.NotEmpty(t, path.String())
			}
		})
	}
}

func TestNewTargetPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "absolute path",
			path:    "/home/user",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.NewTargetPath(tt.path)
			if tt.wantErr {
				assert.True(t, result.IsErr())
			} else {
				assert.True(t, result.IsOk())
			}
		})
	}
}

func TestPathJoin(t *testing.T) {
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles").Unwrap()

	joined := pkgPath.Join("vim")
	assert.Contains(t, joined.String(), "vim")
	assert.Contains(t, joined.String(), "/home/user/.dotfiles")
}

func TestPathParent(t *testing.T) {
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()

	parent := pkgPath.Parent()
	require.True(t, parent.IsOk())

	parentPath := parent.Unwrap()
	assert.Equal(t, "/home/user/.dotfiles", parentPath.String())
}

func TestPathString(t *testing.T) {
	path := "/home/user/.dotfiles"
	pkgPath := domain.NewPackagePath(path).Unwrap()

	assert.Equal(t, path, pkgPath.String())
}

func TestPathClean(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double slashes",
			input:    "/home//user",
			expected: "/home/user",
		},
		{
			name:     "trailing slash",
			input:    "/home/user/",
			expected: "/home/user",
		},
		{
			name:     "dot segments",
			input:    "/home/./user",
			expected: "/home/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.NewPackagePath(tt.input)
			require.True(t, result.IsOk())

			path := result.Unwrap()
			assert.Equal(t, tt.expected, path.String())
		})
	}
}

func TestPathEquality(t *testing.T) {
	path1 := domain.NewPackagePath("/home/user/.dotfiles").Unwrap()
	path2 := domain.NewPackagePath("/home/user/.dotfiles").Unwrap()
	path3 := domain.NewPackagePath("/home/user/other").Unwrap()

	assert.True(t, path1.Equals(path2))
	assert.False(t, path1.Equals(path3))
}

func TestFilePath(t *testing.T) {
	filePath := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()

	assert.Contains(t, filePath.String(), "vimrc")

	parent := filePath.Parent()
	require.True(t, parent.IsOk())
	assert.Contains(t, parent.Unwrap().String(), "vim")
}
