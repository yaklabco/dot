package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManageCommand_Integration_Execute(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create test package
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("set nocompatible"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vim-colors"), []byte("color theme"), 0644))

	// Set global config
	cliFlags = CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
		dryRun:     false,
		verbose:    0,
		quiet:      false,
	}

	// Create command (cliFlags is already set)
	cmd := newManageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"vim"})

	// Execute
	err := cmd.Execute()
	require.NoError(t, err)

	// Verify links created (with package name mapping enabled)
	// Package "vim" → files go to target/vim/
	vimrcLink := filepath.Join(targetDir, "vim", ".vimrc")
	vimColorsLink := filepath.Join(targetDir, "vim", ".vim-colors")

	assert.FileExists(t, vimrcLink)
	assert.FileExists(t, vimColorsLink)

	// Verify links point to correct source
	vimrcTarget, err := os.Readlink(vimrcLink)
	require.NoError(t, err)
	assert.Contains(t, vimrcTarget, "dot-vimrc")
}

func TestManageCommand_Integration_DryRun(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("test"), 0644))

	cliFlags = CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
		dryRun:     true,
		verbose:    0,
		quiet:      false,
	}

	cmd := newManageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"vim"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify no links created (with package name mapping)
	vimrcLink := filepath.Join(targetDir, "vim", ".vimrc")
	assert.NoFileExists(t, vimrcLink)
}

func TestManageCommand_Integration_MultiplePackages(t *testing.T) {
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create vim package
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("vim"), 0644))

	// Create zsh package
	zshPackage := filepath.Join(packageDir, "zsh")
	require.NoError(t, os.MkdirAll(zshPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(zshPackage, "dot-zshrc"), []byte("zsh"), 0644))

	cliFlags = CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
		dryRun:     false,
		verbose:    0,
		quiet:      false,
	}

	cmd := newManageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"vim", "zsh"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify both packages installed (with package name mapping)
	assert.FileExists(t, filepath.Join(targetDir, "vim", ".vimrc"))
	assert.FileExists(t, filepath.Join(targetDir, "zsh", ".zshrc"))
}

func TestManageCommand_Integration_PackageNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	cliFlags = CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
		dryRun:     false,
		verbose:    0,
		quiet:      false,
	}

	cmd := newManageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
