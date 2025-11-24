package planner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/planner"
)

func TestComputeDesiredState_EmptyPackage(t *testing.T) {
	packages := []domain.Package{}
	target := domain.NewTargetPath("/home/user").Unwrap()

	result := planner.ComputeDesiredState(packages, target, false)
	require.True(t, result.IsOk())

	state := result.Unwrap()
	assert.Empty(t, state.Links)
	assert.Empty(t, state.Dirs)
}

func TestComputeDesiredState_SingleFile(t *testing.T) {
	// Package with single file: vim/vimrc -> ~/.vimrc
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := domain.NewTargetPath("/home/user").Unwrap()

	// Create a file node representing vim/vimrc
	fileNode := domain.Node{
		Path: domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap(),
		Type: domain.NodeFile,
	}

	pkg := domain.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &fileNode,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, target, false)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// File "vimrc" (no dot- prefix) should create /home/user/vimrc
	assert.Len(t, state.Links, 1)

	linkSpec, exists := state.Links["/home/user/vimrc"]
	require.True(t, exists, "Expected link at /home/user/vimrc")
	assert.Equal(t, "/home/user/.dotfiles/vim/vimrc", linkSpec.Source.String())
	assert.Equal(t, "/home/user/vimrc", linkSpec.Target.String())
}

func TestComputeDesiredState_DotfileTranslation(t *testing.T) {
	// Package with dot-vimrc -> should become .vimrc in target
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := domain.NewTargetPath("/home/user").Unwrap()

	fileNode := domain.Node{
		Path: domain.NewFilePath("/home/user/.dotfiles/vim/dot-vimrc").Unwrap(),
		Type: domain.NodeFile,
	}

	pkg := domain.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &fileNode,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, target, false)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// dot-vimrc should translate to .vimrc
	linkSpec, exists := state.Links["/home/user/.vimrc"]
	require.True(t, exists, "Expected link at /home/user/.vimrc (translated)")
	assert.Equal(t, "/home/user/.dotfiles/vim/dot-vimrc", linkSpec.Source.String())
}

func TestComputeDesiredState_NestedFiles(t *testing.T) {
	// Package with nested structure: vim/colors/desert.vim
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := domain.NewTargetPath("/home/user").Unwrap()

	// Build tree: vim/ -> colors/ -> desert.vim
	fileNode := domain.Node{
		Path: domain.NewFilePath("/home/user/.dotfiles/vim/colors/desert.vim").Unwrap(),
		Type: domain.NodeFile,
	}

	colorsDir := domain.Node{
		Path:     domain.NewFilePath("/home/user/.dotfiles/vim/colors").Unwrap(),
		Type:     domain.NodeDir,
		Children: []domain.Node{fileNode},
	}

	rootNode := domain.Node{
		Path:     domain.NewFilePath("/home/user/.dotfiles/vim").Unwrap(),
		Type:     domain.NodeDir,
		Children: []domain.Node{colorsDir},
	}

	pkg := domain.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &rootNode,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, target, false)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// Should create: /home/user/colors/.vim/colors/desert.vim -> source
	// Or more likely: /home/user/.vim/colors/desert.vim -> source
	assert.NotEmpty(t, state.Links)

	// Should create parent directory: /home/user/.vim/colors
	assert.NotEmpty(t, state.Dirs)
}

func TestLinkSpec(t *testing.T) {
	source := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()

	spec := planner.LinkSpec{
		Source: source,
		Target: target,
	}

	assert.Equal(t, source, spec.Source)
	assert.Equal(t, target, spec.Target)
}

func TestDirSpec(t *testing.T) {
	path := domain.NewFilePath("/home/user/.vim").Unwrap()

	spec := planner.DirSpec{
		Path: path,
	}

	assert.Equal(t, path, spec.Path)
}

func TestDesiredState(t *testing.T) {
	source := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	dirPath := domain.NewFilePath("/home/user/.vim").Unwrap()

	state := planner.DesiredState{
		Links: map[string]planner.LinkSpec{
			target.String(): {Source: source, Target: target},
		},
		Dirs: map[string]planner.DirSpec{
			dirPath.String(): {Path: dirPath},
		},
	}

	assert.Len(t, state.Links, 1)
	assert.Len(t, state.Dirs, 1)
	assert.Contains(t, state.Links, target.String())
	assert.Contains(t, state.Dirs, dirPath.String())
}

// Task 7.5: Test Integration with Planner
func TestPlanResult(t *testing.T) {
	t.Run("without resolution", func(t *testing.T) {
		desired := planner.DesiredState{
			Links: make(map[string]planner.LinkSpec),
			Dirs:  make(map[string]planner.DirSpec),
		}

		result := planner.PlanResult{
			Desired: desired,
		}

		assert.NotNil(t, result.Desired)
		assert.Nil(t, result.Resolved)
		assert.False(t, result.HasConflicts())
	})

	t.Run("with resolution", func(t *testing.T) {
		desired := planner.DesiredState{
			Links: make(map[string]planner.LinkSpec),
			Dirs:  make(map[string]planner.DirSpec),
		}

		targetPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict := planner.NewConflict(planner.ConflictFileExists, targetPath, "File exists")

		resolved := planner.NewResolveResult(nil).WithConflict(conflict)

		result := planner.PlanResult{
			Desired:  desired,
			Resolved: &resolved,
		}

		assert.NotNil(t, result.Resolved)
		assert.True(t, result.HasConflicts())
	})
}

func TestComputeOperationsFromDesiredState(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

	desired := planner.DesiredState{
		Links: map[string]planner.LinkSpec{
			targetPath.String(): {
				Source: sourcePath,
				Target: targetPath,
			},
		},
		Dirs: make(map[string]planner.DirSpec),
	}

	ops := planner.ComputeOperationsFromDesiredState(desired)

	assert.Len(t, ops, 1)
	linkOp, ok := ops[0].(domain.LinkCreate)
	assert.True(t, ok)
	assert.Equal(t, sourcePath, linkOp.Source)
	assert.Equal(t, targetPath, linkOp.Target)
}

func TestComputeOperationsFromDesiredStateWithDirs(t *testing.T) {
	dirPath := domain.NewFilePath("/home/user/.config").Unwrap()
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.config/bash").Unwrap()

	desired := planner.DesiredState{
		Links: map[string]planner.LinkSpec{
			targetPath.String(): {
				Source: sourcePath,
				Target: targetPath,
			},
		},
		Dirs: map[string]planner.DirSpec{
			dirPath.String(): {Path: dirPath},
		},
	}

	ops := planner.ComputeOperationsFromDesiredState(desired)

	assert.Len(t, ops, 2) // One dir + one link

	// Should have both operation types
	hasDirCreate := false
	hasLinkCreate := false
	for _, op := range ops {
		switch op.Kind() {
		case domain.OpKindDirCreate:
			hasDirCreate = true
		case domain.OpKindLinkCreate:
			hasLinkCreate = true
		}
	}
	assert.True(t, hasDirCreate)
	assert.True(t, hasLinkCreate)
}

func TestComputeDesiredStateWithMultipleFiles(t *testing.T) {
	targetDir := domain.NewTargetPath("/home/user").Unwrap()

	// Create package with multiple files
	pkgPath := domain.NewPackagePath("/packages/bash").Unwrap()
	pkgRoot := domain.NewFilePath("/packages/bash").Unwrap()
	file1Path := pkgPath.Join("dot-bashrc")
	file2Path := pkgPath.Join("dot-profile")
	file1 := domain.NewFilePath(file1Path.String()).Unwrap()
	file2 := domain.NewFilePath(file2Path.String()).Unwrap()

	tree := &domain.Node{
		Path: pkgRoot,
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: file1,
				Type: domain.NodeFile,
			},
			{
				Path: file2,
				Type: domain.NodeFile,
			},
		},
	}

	pkg := domain.Package{
		Name: "bash",
		Path: pkgPath,
		Tree: tree,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, targetDir, false)

	assert.True(t, result.IsOk())
	state := result.Unwrap()

	// Should have 2 links
	assert.Len(t, state.Links, 2)
}

func TestComputeDesiredState_PackageNameMapping(t *testing.T) {
	t.Run("with package name mapping enabled", func(t *testing.T) {
		// Package "dot-gnupg" with file "common.conf"
		// Should produce target "~/.gnupg/common.conf"
		pkgPath := domain.NewPackagePath("/home/user/dotfiles/dot-gnupg").Unwrap()
		target := domain.NewTargetPath("/home/user").Unwrap()

		fileNode := domain.Node{
			Path: domain.NewFilePath("/home/user/dotfiles/dot-gnupg/common.conf").Unwrap(),
			Type: domain.NodeFile,
		}

		pkg := domain.Package{
			Name: "dot-gnupg",
			Path: pkgPath,
			Tree: &fileNode,
		}

		result := planner.ComputeDesiredState([]domain.Package{pkg}, target, true)
		require.True(t, result.IsOk())

		state := result.Unwrap()

		// Should create link at ~/.gnupg/common.conf (not ~/common.conf)
		linkSpec, exists := state.Links["/home/user/.gnupg/common.conf"]
		require.True(t, exists, "Expected link at /home/user/.gnupg/common.conf")
		assert.Equal(t, "/home/user/dotfiles/dot-gnupg/common.conf", linkSpec.Source.String())
		assert.Equal(t, "/home/user/.gnupg/common.conf", linkSpec.Target.String())

		// Should create parent directory .gnupg
		_, dirExists := state.Dirs["/home/user/.gnupg"]
		assert.True(t, dirExists, "Expected parent directory /home/user/.gnupg")
	})

	t.Run("with package name mapping disabled", func(t *testing.T) {
		// Package "dot-gnupg" with file "common.conf"
		// Should produce target "~/common.conf" (legacy behavior)
		pkgPath := domain.NewPackagePath("/home/user/dotfiles/dot-gnupg").Unwrap()
		target := domain.NewTargetPath("/home/user").Unwrap()

		fileNode := domain.Node{
			Path: domain.NewFilePath("/home/user/dotfiles/dot-gnupg/common.conf").Unwrap(),
			Type: domain.NodeFile,
		}

		pkg := domain.Package{
			Name: "dot-gnupg",
			Path: pkgPath,
			Tree: &fileNode,
		}

		result := planner.ComputeDesiredState([]domain.Package{pkg}, target, false)
		require.True(t, result.IsOk())

		state := result.Unwrap()

		// Should create link at ~/common.conf (not ~/.gnupg/common.conf)
		linkSpec, exists := state.Links["/home/user/common.conf"]
		require.True(t, exists, "Expected link at /home/user/common.conf")
		assert.Equal(t, "/home/user/dotfiles/dot-gnupg/common.conf", linkSpec.Source.String())
	})

	t.Run("nested directories with package name mapping", func(t *testing.T) {
		// Package "dot-gnupg" with "public-keys.d/pubring.db"
		// Should produce target "~/.gnupg/public-keys.d/pubring.db"
		pkgPath := domain.NewPackagePath("/home/user/dotfiles/dot-gnupg").Unwrap()
		target := domain.NewTargetPath("/home/user").Unwrap()

		fileNode := domain.Node{
			Path: domain.NewFilePath("/home/user/dotfiles/dot-gnupg/public-keys.d/pubring.db").Unwrap(),
			Type: domain.NodeFile,
		}

		keysDir := domain.Node{
			Path:     domain.NewFilePath("/home/user/dotfiles/dot-gnupg/public-keys.d").Unwrap(),
			Type:     domain.NodeDir,
			Children: []domain.Node{fileNode},
		}

		rootNode := domain.Node{
			Path:     domain.NewFilePath("/home/user/dotfiles/dot-gnupg").Unwrap(),
			Type:     domain.NodeDir,
			Children: []domain.Node{keysDir},
		}

		pkg := domain.Package{
			Name: "dot-gnupg",
			Path: pkgPath,
			Tree: &rootNode,
		}

		result := planner.ComputeDesiredState([]domain.Package{pkg}, target, true)
		require.True(t, result.IsOk())

		state := result.Unwrap()

		// Should create link at ~/.gnupg/public-keys.d/pubring.db
		linkSpec, exists := state.Links["/home/user/.gnupg/public-keys.d/pubring.db"]
		require.True(t, exists, "Expected link at /home/user/.gnupg/public-keys.d/pubring.db")
		assert.Equal(t, "/home/user/dotfiles/dot-gnupg/public-keys.d/pubring.db", linkSpec.Source.String())

		// Should create parent directories
		_, gnupgExists := state.Dirs["/home/user/.gnupg"]
		assert.True(t, gnupgExists, "Expected directory /home/user/.gnupg")

		_, keysExists := state.Dirs["/home/user/.gnupg/public-keys.d"]
		assert.True(t, keysExists, "Expected directory /home/user/.gnupg/public-keys.d")
	})

	t.Run("non-prefixed package with mapping enabled", func(t *testing.T) {
		// Package "vim" with file "init.lua"
		// Should produce target "~/vim/init.lua"
		pkgPath := domain.NewPackagePath("/home/user/dotfiles/vim").Unwrap()
		target := domain.NewTargetPath("/home/user").Unwrap()

		fileNode := domain.Node{
			Path: domain.NewFilePath("/home/user/dotfiles/vim/init.lua").Unwrap(),
			Type: domain.NodeFile,
		}

		pkg := domain.Package{
			Name: "vim",
			Path: pkgPath,
			Tree: &fileNode,
		}

		result := planner.ComputeDesiredState([]domain.Package{pkg}, target, true)
		require.True(t, result.IsOk())

		state := result.Unwrap()

		// Should create link at ~/vim/init.lua (no dot translation for package name)
		linkSpec, exists := state.Links["/home/user/vim/init.lua"]
		require.True(t, exists, "Expected link at /home/user/vim/init.lua")
		assert.Equal(t, "/home/user/dotfiles/vim/init.lua", linkSpec.Source.String())
	})

	t.Run("file-level dot- translation with package mapping", func(t *testing.T) {
		// Package "vim" with file "dot-vimrc"
		// Should produce target "~/vim/.vimrc" (both package and file translation)
		pkgPath := domain.NewPackagePath("/home/user/dotfiles/vim").Unwrap()
		target := domain.NewTargetPath("/home/user").Unwrap()

		fileNode := domain.Node{
			Path: domain.NewFilePath("/home/user/dotfiles/vim/dot-vimrc").Unwrap(),
			Type: domain.NodeFile,
		}

		pkg := domain.Package{
			Name: "vim",
			Path: pkgPath,
			Tree: &fileNode,
		}

		result := planner.ComputeDesiredState([]domain.Package{pkg}, target, true)
		require.True(t, result.IsOk())

		state := result.Unwrap()

		// Should create link at ~/vim/.vimrc (file-level translation applied)
		linkSpec, exists := state.Links["/home/user/vim/.vimrc"]
		require.True(t, exists, "Expected link at /home/user/vim/.vimrc")
		assert.Equal(t, "/home/user/dotfiles/vim/dot-vimrc", linkSpec.Source.String())
	})
}
