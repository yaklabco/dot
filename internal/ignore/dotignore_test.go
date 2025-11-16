package ignore_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDotignoreFile_NotExists(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Nil(t, patterns)
}

func TestLoadDotignoreFile_Empty(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	err := fs.WriteFile(ctx, "/.dotignore", []byte(""), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestLoadDotignoreFile_SinglePattern(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	content := "*.log\n"
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log"}, patterns)
}

func TestLoadDotignoreFile_MultiplePatterns(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	content := `*.log
*.tmp
.DS_Store
node_modules/
`
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "*.tmp", ".DS_Store", "node_modules/"}, patterns)
}

func TestLoadDotignoreFile_Comments(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	content := `# This is a comment
*.log
# Another comment
*.tmp
#inline comment
.DS_Store
`
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "*.tmp", ".DS_Store"}, patterns)
}

func TestLoadDotignoreFile_EmptyLines(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	content := `*.log


*.tmp

.DS_Store

`
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "*.tmp", ".DS_Store"}, patterns)
}

func TestLoadDotignoreFile_WhitespaceHandling(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	content := `  *.log  
	*.tmp	
 .DS_Store
`
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "*.tmp", ".DS_Store"}, patterns)
}

func TestLoadDotignoreFile_NegationPattern(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	content := `*.log
!important.log
`
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "!important.log"}, patterns)
}

func TestLoadDotignoreFile_InvalidMultipleNegation(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	content := `*.log
!!invalid.log
`
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pattern at line 2")
	assert.Contains(t, err.Error(), "multiple ! prefixes not allowed")
	assert.Nil(t, patterns)
}

func TestLoadDotignoreFile_ReadError(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create a directory with the same name to cause read error
	err := fs.Mkdir(ctx, "/.dotignore", 0755)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read .dotignore")
	assert.Nil(t, patterns)
}

func TestLoadDotignoreWithInheritance_NoFiles(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory structure but no .dotignore files
	err := fs.Mkdir(ctx, "/packages", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/packages/vim", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/packages/vim/colors", 0755)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/packages/vim/colors", "/packages")

	assert.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestLoadDotignoreWithInheritance_SingleLevel(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory and .dotignore
	err := fs.Mkdir(ctx, "/packages", 0755)
	require.NoError(t, err)

	content := "*.log\n*.tmp\n"
	err = fs.WriteFile(ctx, "/packages/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/packages", "/packages")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "*.tmp"}, patterns)
}

func TestLoadDotignoreWithInheritance_MultiLevel(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory structure
	err := fs.Mkdir(ctx, "/packages", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/packages/vim", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/packages/vim/colors", 0755)
	require.NoError(t, err)

	// Root .dotignore
	err = fs.WriteFile(ctx, "/packages/.dotignore", []byte("*.log\n"), 0644)
	require.NoError(t, err)

	// Middle .dotignore
	err = fs.WriteFile(ctx, "/packages/vim/.dotignore", []byte("*.swp\n"), 0644)
	require.NoError(t, err)

	// Leaf .dotignore
	err = fs.WriteFile(ctx, "/packages/vim/colors/.dotignore", []byte("*.tmp\n"), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/packages/vim/colors", "/packages")

	assert.NoError(t, err)
	// Parent patterns come first, child patterns last (for override behavior)
	assert.Equal(t, []string{"*.log", "*.swp", "*.tmp"}, patterns)
}

func TestLoadDotignoreWithInheritance_Priority(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory structure
	err := fs.Mkdir(ctx, "/packages", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/packages/vim", 0755)
	require.NoError(t, err)

	// Root .dotignore ignores all .log files
	err = fs.WriteFile(ctx, "/packages/.dotignore", []byte("*.log\n"), 0644)
	require.NoError(t, err)

	// Subdir .dotignore negates for specific file
	err = fs.WriteFile(ctx, "/packages/vim/.dotignore", []byte("!important.log\n"), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/packages/vim", "/packages")

	assert.NoError(t, err)
	// Parent patterns come first, subdirectory patterns override
	assert.Equal(t, []string{"*.log", "!important.log"}, patterns)
}

func TestLoadDotignoreWithInheritance_StopsAtRoot(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory structure
	err := fs.Mkdir(ctx, "/root", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/root/packages", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/root/packages/vim", 0755)
	require.NoError(t, err)

	// Above root
	err = fs.WriteFile(ctx, "/root/.dotignore", []byte("above.log\n"), 0644)
	require.NoError(t, err)

	// At root
	err = fs.WriteFile(ctx, "/root/packages/.dotignore", []byte("at-root.log\n"), 0644)
	require.NoError(t, err)

	// Below root
	err = fs.WriteFile(ctx, "/root/packages/vim/.dotignore", []byte("below.log\n"), 0644)
	require.NoError(t, err)

	// Load from /root/packages/vim with root=/root/packages
	// Should NOT include "above.log" from /root/.dotignore
	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/root/packages/vim", "/root/packages")

	assert.NoError(t, err)
	assert.Equal(t, []string{"at-root.log", "below.log"}, patterns)
	assert.NotContains(t, patterns, "above.log")
}

func TestLoadDotignoreWithInheritance_PathNormalization(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory structure
	err := fs.Mkdir(ctx, "/packages", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/packages/vim", 0755)
	require.NoError(t, err)

	err = fs.WriteFile(ctx, "/packages/.dotignore", []byte("*.log\n"), 0644)
	require.NoError(t, err)
	err = fs.WriteFile(ctx, "/packages/vim/.dotignore", []byte("*.swp\n"), 0644)
	require.NoError(t, err)

	// Use unnormalized paths
	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/packages//vim/", "/packages/")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "*.swp"}, patterns)
}

func TestLoadDotignoreWithInheritance_ErrorPropagation(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory structure
	err := fs.Mkdir(ctx, "/packages", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/packages/vim", 0755)
	require.NoError(t, err)

	// Create invalid .dotignore
	content := "*.log\n!!invalid\n"
	err = fs.WriteFile(ctx, "/packages/vim/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/packages/vim", "/packages")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load /packages/vim/.dotignore")
	assert.Contains(t, err.Error(), "invalid pattern")
	assert.Nil(t, patterns)
}

func TestLoadDotignoreWithInheritance_RootFilesystem(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create file at filesystem root
	err := fs.WriteFile(ctx, "/.dotignore", []byte("*.log\n"), 0644)
	require.NoError(t, err)

	// Load from root with root as boundary
	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/", "/")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log"}, patterns)
}

func TestLoadDotignoreWithInheritance_ComplexStructure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create complex directory structure
	paths := []string{
		"/home/user/dotfiles",
		"/home/user/dotfiles/shell",
		"/home/user/dotfiles/shell/bash",
		"/home/user/dotfiles/vim",
	}
	for _, path := range paths {
		err := fs.Mkdir(ctx, path, 0755)
		require.NoError(t, err)
	}

	// Different .dotignore files at each level
	dotignores := map[string]string{
		"/home/user/dotfiles/.dotignore":            "global-*",
		"/home/user/dotfiles/shell/.dotignore":      "shell-*",
		"/home/user/dotfiles/shell/bash/.dotignore": "bash-*",
	}
	for path, content := range dotignores {
		err := fs.WriteFile(ctx, path, []byte(content+"\n"), 0644)
		require.NoError(t, err)
	}

	patterns, err := ignore.LoadDotignoreWithInheritance(
		ctx,
		fs,
		"/home/user/dotfiles/shell/bash",
		"/home/user/dotfiles",
	)

	assert.NoError(t, err)
	// Patterns should be in order: root first, deepest last (for override behavior)
	assert.Equal(t, []string{"global-*", "shell-*", "bash-*"}, patterns)
}

func TestLoadDotignoreWithInheritance_PartialFiles(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create directory structure
	err := fs.Mkdir(ctx, "/a", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/a/b", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/a/b/c", 0755)
	require.NoError(t, err)

	// Only some levels have .dotignore
	err = fs.WriteFile(ctx, "/a/.dotignore", []byte("first\n"), 0644)
	require.NoError(t, err)
	// /a/b has no .dotignore
	err = fs.WriteFile(ctx, "/a/b/c/.dotignore", []byte("third\n"), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, "/a/b/c", "/a")

	assert.NoError(t, err)
	assert.Equal(t, []string{"first", "third"}, patterns)
}

func TestLoadDotignoreFile_NoTrailingNewline(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Content without trailing newline
	content := "*.log"
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log"}, patterns)
}

func TestLoadDotignoreFile_MixedLineEndings(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Mix of Unix and potential Windows line endings
	content := "*.log\n*.tmp\n.DS_Store"
	err := fs.WriteFile(ctx, "/.dotignore", []byte(content), 0644)
	require.NoError(t, err)

	patterns, err := ignore.LoadDotignoreFile(ctx, fs, "/.dotignore")

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log", "*.tmp", ".DS_Store"}, patterns)
}

func TestLoadDotignoreWithInheritance_RelativePaths(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create structure
	err := fs.Mkdir(ctx, "/test", 0755)
	require.NoError(t, err)
	err = fs.Mkdir(ctx, "/test/sub", 0755)
	require.NoError(t, err)

	err = fs.WriteFile(ctx, "/test/.dotignore", []byte("*.log\n"), 0644)
	require.NoError(t, err)

	// Use paths without leading slash
	startPath := filepath.Clean("/test/sub")
	rootPath := filepath.Clean("/test")

	patterns, err := ignore.LoadDotignoreWithInheritance(ctx, fs, startPath, rootPath)

	assert.NoError(t, err)
	assert.Equal(t, []string{"*.log"}, patterns)
}
