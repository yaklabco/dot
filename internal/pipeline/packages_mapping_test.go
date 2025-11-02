package pipeline

import (
	"testing"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUnderPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		basePath string
		want     bool
	}{
		{
			name:     "direct child",
			path:     "/packages/vim/dot-vimrc",
			basePath: "/packages/vim",
			want:     true,
		},
		{
			name:     "nested child",
			path:     "/packages/vim/colors/theme.vim",
			basePath: "/packages/vim",
			want:     true,
		},
		{
			name:     "sibling",
			path:     "/packages/zsh/dot-zshrc",
			basePath: "/packages/vim",
			want:     false,
		},
		{
			name:     "parent",
			path:     "/packages",
			basePath: "/packages/vim",
			want:     false,
		},
		{
			name:     "same path",
			path:     "/packages/vim",
			basePath: "/packages/vim",
			want:     false,
		},
		{
			name:     "prefix match but different",
			path:     "/packages/vimrc",
			basePath: "/packages/vim",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnderPath(tt.path, tt.basePath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOperationBelongsToPackage(t *testing.T) {
	vimPkgPath := "/packages/vim"

	tests := []struct {
		name    string
		op      domain.Operation
		pkgPath string
		want    bool
	}{
		{
			name: "LinkCreate from package",
			op: domain.NewLinkCreate(
				domain.OperationID("test-1"),
				mustFilePath("/packages/vim/dot-vimrc"),
				mustTargetPath("/home/user/.vimrc"),
			),
			pkgPath: vimPkgPath,
			want:    true,
		},
		{
			name: "LinkCreate from different package",
			op: domain.NewLinkCreate(
				domain.OperationID("test-2"),
				mustFilePath("/packages/zsh/dot-zshrc"),
				mustTargetPath("/home/user/.zshrc"),
			),
			pkgPath: vimPkgPath,
			want:    false,
		},
		{
			name: "DirCreate",
			op: domain.NewDirCreate(
				domain.OperationID("test-3"),
				mustFilePath("/home/user/.vim"),
			),
			pkgPath: vimPkgPath,
			want:    false,
		},
		{
			name: "LinkDelete",
			op: domain.NewLinkDelete(
				domain.OperationID("test-4"),
				mustTargetPath("/home/user/.vimrc"),
			),
			pkgPath: vimPkgPath,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For basic LinkCreate and FileMove tests, targetToPackage is not needed
			got := operationBelongsToPackage(tt.op, "", tt.pkgPath, make(map[string]string))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildPackageOperationMapping(t *testing.T) {
	// Create test packages
	packages := []domain.Package{
		{
			Name: "vim",
			Path: mustPackagePath("/packages/vim"),
		},
		{
			Name: "zsh",
			Path: mustPackagePath("/packages/zsh"),
		},
	}

	// Create test operations
	ops := []domain.Operation{
		domain.NewLinkCreate(
			domain.OperationID("vim-link-1"),
			mustFilePath("/packages/vim/dot-vimrc"),
			mustTargetPath("/home/user/.vimrc"),
		),
		domain.NewLinkCreate(
			domain.OperationID("vim-link-2"),
			mustFilePath("/packages/vim/dot-vim-colors"),
			mustTargetPath("/home/user/.vim-colors"),
		),
		domain.NewLinkCreate(
			domain.OperationID("zsh-link-1"),
			mustFilePath("/packages/zsh/dot-zshrc"),
			mustTargetPath("/home/user/.zshrc"),
		),
		domain.NewDirCreate(
			domain.OperationID("dir-1"),
			mustFilePath("/home/user/.config"),
		),
	}

	// Build mapping
	mapping := buildPackageOperationMapping(packages, ops)

	// Verify vim operations
	require.Contains(t, mapping, "vim")
	assert.Len(t, mapping["vim"], 2)
	assert.Contains(t, mapping["vim"], domain.OperationID("vim-link-1"))
	assert.Contains(t, mapping["vim"], domain.OperationID("vim-link-2"))

	// Verify zsh operations
	require.Contains(t, mapping, "zsh")
	assert.Len(t, mapping["zsh"], 1)
	assert.Contains(t, mapping["zsh"], domain.OperationID("zsh-link-1"))

	// Verify dir operation not assigned to any package
	for _, opIDs := range mapping {
		assert.NotContains(t, opIDs, domain.OperationID("dir-1"))
	}
}

func TestBuildPackageOperationMapping_EmptyPackages(t *testing.T) {
	ops := []domain.Operation{
		domain.NewLinkCreate(
			domain.OperationID("link-1"),
			mustFilePath("/packages/vim/dot-vimrc"),
			mustTargetPath("/home/user/.vimrc"),
		),
	}

	mapping := buildPackageOperationMapping([]domain.Package{}, ops)
	assert.Len(t, mapping, 0)
}

func TestBuildPackageOperationMapping_EmptyOperations(t *testing.T) {
	packages := []domain.Package{
		{
			Name: "vim",
			Path: mustPackagePath("/packages/vim"),
		},
	}

	mapping := buildPackageOperationMapping(packages, []domain.Operation{})
	assert.Len(t, mapping, 0)
}

func TestBuildPackageOperationMapping_NoMatchingOperations(t *testing.T) {
	packages := []domain.Package{
		{
			Name: "vim",
			Path: mustPackagePath("/packages/vim"),
		},
	}

	ops := []domain.Operation{
		domain.NewDirCreate(
			domain.OperationID("dir-1"),
			mustFilePath("/home/user/.config"),
		),
	}

	mapping := buildPackageOperationMapping(packages, ops)
	assert.Len(t, mapping, 0, "should not create mapping for packages with no matching operations")
}

// Test helpers

func mustFilePath(path string) domain.FilePath {
	result := domain.NewFilePath(path)
	if !result.IsOk() {
		panic("invalid file path: " + path)
	}
	return result.Unwrap()
}

func mustTargetPath(path string) domain.TargetPath {
	result := domain.NewTargetPath(path)
	if !result.IsOk() {
		panic("invalid target path: " + path)
	}
	return result.Unwrap()
}

func mustPackagePath(path string) domain.PackagePath {
	result := domain.NewPackagePath(path)
	if !result.IsOk() {
		panic("invalid package path: " + path)
	}
	return result.Unwrap()
}
