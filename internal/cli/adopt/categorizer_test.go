package adopt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeFile_Shell(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"bashrc", ".bashrc"},
		{"bash_profile", ".bash_profile"},
		{"zshrc", ".zshrc"},
		{"zshenv", ".zshenv"},
		{"fish", "fish"},
		{"profile", ".profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeFile(tt.filename)
			assert.Equal(t, "shell", result)
		})
	}
}

func TestCategorizeFile_VCS(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"gitconfig", ".gitconfig"},
		{"gitignore", ".gitignore"},
		{"git-credentials", ".git-credentials"},
		{"hgrc", ".hgrc"},
		{"svn", ".svn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeFile(tt.filename)
			assert.Equal(t, "vcs", result)
		})
	}
}

func TestCategorizeFile_Editor(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"vimrc", ".vimrc"},
		{"vim", ".vim"},
		{"nvim", "nvim"},
		{"emacs", ".emacs"},
		{"spacemacs", ".spacemacs"},
		{"vscode", "vscode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeFile(tt.filename)
			assert.Equal(t, "editor", result)
		})
	}
}

func TestCategorizeFile_Tool(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"ssh", ".ssh"},
		{"docker", ".docker"},
		{"aws", ".aws"},
		{"kube", ".kube"},
		{"tmux", ".tmux"},
		{"tmux.conf", ".tmux.conf"},
		{"gnupg", ".gnupg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeFile(tt.filename)
			assert.Equal(t, "tool", result)
		})
	}
}

func TestCategorizeFile_Misc(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"unknown", ".unknown"},
		{"random", ".randomfile"},
		{"custom", ".customconfig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeFile(tt.filename)
			assert.Equal(t, "misc", result)
		})
	}
}

func TestSuggestPackageName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		category string
		expected string
	}{
		{"bashrc", ".bashrc", "shell", "bash"},
		{"bash_profile", ".bash_profile", "shell", "bash"},
		{"zshrc", ".zshrc", "shell", "zsh"},
		{"gitconfig", ".gitconfig", "vcs", "git"},
		{"gitignore", ".gitignore", "vcs", "git"},
		{"vimrc", ".vimrc", "editor", "vim"},
		{"nvim", "nvim", "editor", "nvim"},
		{"vim dir", ".vim", "editor", "vim"},
		{"tmux.conf", ".tmux.conf", "tool", "tmux"},
		{"ssh", ".ssh", "tool", "ssh"},
		{"aws", ".aws", "tool", "aws"},
		{"docker", ".docker", "tool", "docker"},
		{"unknown", ".unknown", "misc", "unknown"},
		{"custom", ".customfile", "misc", "customfile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suggestPackageName(tt.filename, tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSuggestPackageName_StripsDot(t *testing.T) {
	result := suggestPackageName(".randomfile", "misc")
	assert.Equal(t, "randomfile", result)
	assert.NotContains(t, result, ".")
}

func TestGroupByCategory(t *testing.T) {
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			Category:     "shell",
			SuggestedPkg: "bash",
		},
		{
			Path:         "/home/user/.bash_profile",
			RelPath:      ".bash_profile",
			Category:     "shell",
			SuggestedPkg: "bash",
		},
		{
			Path:         "/home/user/.zshrc",
			RelPath:      ".zshrc",
			Category:     "shell",
			SuggestedPkg: "zsh",
		},
		{
			Path:         "/home/user/.gitconfig",
			RelPath:      ".gitconfig",
			Category:     "vcs",
			SuggestedPkg: "git",
		},
	}

	// Select indices 0, 1, 3 (two bash files and gitconfig)
	selections := []int{0, 1, 3}

	groups := GroupByCategory(candidates, selections)

	assert.Len(t, groups, 2)
	assert.Len(t, groups["bash"], 2)
	assert.Len(t, groups["git"], 1)

	// Verify bash group contains correct files
	bashPaths := make(map[string]bool)
	for _, c := range groups["bash"] {
		bashPaths[c.RelPath] = true
	}
	assert.True(t, bashPaths[".bashrc"])
	assert.True(t, bashPaths[".bash_profile"])

	// Verify git group contains gitconfig
	assert.Equal(t, ".gitconfig", groups["git"][0].RelPath)
}

func TestGroupByCategory_InvalidIndices(t *testing.T) {
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			SuggestedPkg: "bash",
		},
	}

	// Include invalid indices
	selections := []int{-1, 0, 999}

	groups := GroupByCategory(candidates, selections)

	// Should only group the valid index
	assert.Len(t, groups, 1)
	assert.Len(t, groups["bash"], 1)
}

func TestGroupByCategory_EmptySelections(t *testing.T) {
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			SuggestedPkg: "bash",
		},
	}

	selections := []int{}

	groups := GroupByCategory(candidates, selections)

	assert.Empty(t, groups)
}

func TestGroupByCategory_MultiplePackages(t *testing.T) {
	now := time.Now()
	candidates := []DotfileCandidate{
		{Path: "/home/user/.bashrc", RelPath: ".bashrc", SuggestedPkg: "bash", ModTime: now},
		{Path: "/home/user/.zshrc", RelPath: ".zshrc", SuggestedPkg: "zsh", ModTime: now},
		{Path: "/home/user/.vimrc", RelPath: ".vimrc", SuggestedPkg: "vim", ModTime: now},
		{Path: "/home/user/.gitconfig", RelPath: ".gitconfig", SuggestedPkg: "git", ModTime: now},
	}

	// Select all
	selections := []int{0, 1, 2, 3}

	groups := GroupByCategory(candidates, selections)

	assert.Len(t, groups, 4)
	assert.Len(t, groups["bash"], 1)
	assert.Len(t, groups["zsh"], 1)
	assert.Len(t, groups["vim"], 1)
	assert.Len(t, groups["git"], 1)
}
