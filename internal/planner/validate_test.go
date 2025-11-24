package planner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

func TestDotOperationalPaths(t *testing.T) {
	paths := DotOperationalPaths()

	// Should return at least config and data paths
	assert.GreaterOrEqual(t, len(paths), 2)

	// Should contain dot directories
	foundConfig := false
	foundData := false
	for _, p := range paths {
		if filepath.Base(p) == "dot" {
			if filepath.Base(filepath.Dir(p)) == ".config" || filepath.Base(filepath.Dir(filepath.Dir(p))) == ".config" {
				foundConfig = true
			}
			if strings.Contains(p, ".local/share") || strings.Contains(p, "share") {
				foundData = true
			}
		}
	}

	assert.True(t, foundConfig || foundData, "should contain at least one dot operational path")
}

func TestValidateNoSelfManagement_AllowsNormalPackages(t *testing.T) {
	srcVimrc := domain.NewFilePath("/packages/vim/dot-vimrc").Unwrap()
	tgtVimrc := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	srcZshrc := domain.NewFilePath("/packages/zsh/dot-zshrc").Unwrap()
	tgtZshrc := domain.NewTargetPath("/home/user/.zshrc").Unwrap()

	desired := DesiredState{
		Links: map[string]LinkSpec{
			"/home/user/.vimrc": {Source: srcVimrc, Target: tgtVimrc},
			"/home/user/.zshrc": {Source: srcZshrc, Target: tgtZshrc},
		},
		Dirs: map[string]DirSpec{
			"/home/user/.vim": {},
		},
	}

	err := ValidateNoSelfManagement("vim", desired)
	assert.NoError(t, err)
}

func TestValidateNoSelfManagement_DetectsConfigDirectory(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(homeDir, ".config")
	}

	dotConfig := filepath.Join(xdgConfig, "dot", "config.yaml")
	srcPath := domain.NewFilePath("/packages/dot/config.yaml").Unwrap()
	tgtPath := domain.NewTargetPath(dotConfig).Unwrap()

	desired := DesiredState{
		Links: map[string]LinkSpec{
			dotConfig: {Source: srcPath, Target: tgtPath},
		},
		Dirs: map[string]DirSpec{},
	}

	err := ValidateNoSelfManagement("dot", desired)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operational directory")
}

func TestValidateNoSelfManagement_DetectsDataDirectory(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData == "" {
		xdgData = filepath.Join(homeDir, ".local", "share")
	}

	dotState := filepath.Join(xdgData, "dot", "state.json")
	srcPath := domain.NewFilePath("/packages/dot/state.json").Unwrap()
	tgtPath := domain.NewTargetPath(dotState).Unwrap()

	desired := DesiredState{
		Links: map[string]LinkSpec{
			dotState: {Source: srcPath, Target: tgtPath},
		},
		Dirs: map[string]DirSpec{},
	}

	err := ValidateNoSelfManagement("dot", desired)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operational directory")
}

func TestValidateNoSelfManagement_DetectsDirectoryConflict(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(homeDir, ".config")
	}

	desired := DesiredState{
		Links: map[string]LinkSpec{},
		Dirs: map[string]DirSpec{
			filepath.Join(xdgConfig, "dot"): {},
		},
	}

	err := ValidateNoSelfManagement("dot", desired)
	assert.Error(t, err)
}

func TestValidateNoSelfManagement_AllowsNeighboringDirectories(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(homeDir, ".config")
	}

	// Neighboring directories like ~/.config/nvim should be fine
	nvimInit := filepath.Join(xdgConfig, "nvim", "init.vim")
	srcPath := domain.NewFilePath("/packages/nvim/init.vim").Unwrap()
	tgtPath := domain.NewTargetPath(nvimInit).Unwrap()

	desired := DesiredState{
		Links: map[string]LinkSpec{
			nvimInit: {Source: srcPath, Target: tgtPath},
		},
		Dirs: map[string]DirSpec{
			filepath.Join(xdgConfig, "nvim"): {},
		},
	}

	err := ValidateNoSelfManagement("nvim", desired)
	assert.NoError(t, err)
}

func TestValidateNoSelfManagement_EmptyDesiredState(t *testing.T) {
	desired := DesiredState{
		Links: map[string]LinkSpec{},
		Dirs:  map[string]DirSpec{},
	}

	err := ValidateNoSelfManagement("empty", desired)
	assert.NoError(t, err)
}
