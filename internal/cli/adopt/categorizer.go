package adopt

import (
	"path/filepath"
	"strings"
)

// categorizeFile attempts to categorize a file based on its name.
// Returns category hint: "shell", "vcs", "editor", "tool", "misc"
func categorizeFile(name string) string {
	base := strings.ToLower(filepath.Base(name))

	// Shell configurations
	shellPatterns := []string{
		"bashrc", "bash_profile", "bash_aliases", "bash_logout",
		"zshrc", "zshenv", "zprofile", "zlogin", "zlogout",
		"profile", "shrc",
		"fish", // .config/fish
	}
	for _, pattern := range shellPatterns {
		if strings.Contains(base, pattern) {
			return "shell"
		}
	}

	// Version control
	vcsPatterns := []string{".git", ".hg", ".svn"}
	for _, pattern := range vcsPatterns {
		if strings.Contains(base, pattern) {
			return "vcs"
		}
	}

	// Editors
	editorPatterns := []string{
		".vim", ".nvim", ".emacs", ".spacemacs",
		"nvim", "vim", "emacs", "vscode", "code",
	}
	for _, pattern := range editorPatterns {
		if strings.Contains(base, pattern) {
			return "editor"
		}
	}

	// Development tools
	toolPatterns := []string{
		".docker", ".aws", ".kube", "tmux", ".tmux",
		".ssh", ".gnupg", ".gpg",
	}
	for _, pattern := range toolPatterns {
		if strings.Contains(base, pattern) {
			return "tool"
		}
	}

	return "misc"
}

// suggestPackageName suggests a package name based on the file/directory.
func suggestPackageName(name string, category string) string {
	base := filepath.Base(name)

	// Remove leading dot for suggestion
	suggested := strings.TrimPrefix(base, ".")

	// For common patterns, use specific names
	patterns := map[string]string{
		"bashrc":           "bash",
		"bash_profile":     "bash",
		"bash_aliases":     "bash",
		"bash_logout":      "bash",
		"zshrc":            "zsh",
		"zshenv":           "zsh",
		"zprofile":         "zsh",
		"zlogin":           "zsh",
		"gitconfig":        "git",
		"gitignore":        "git",
		"gitignore_global": "git",
		"git-credentials":  "git",
		"vimrc":            "vim",
		"nvim":             "nvim",
		"vim":              "vim",
		"emacs":            "emacs",
		"spacemacs":        "emacs",
		"tmux.conf":        "tmux",
		"tmux":             "tmux",
		"ssh":              "ssh",
		"gnupg":            "gnupg",
		"gpg":              "gnupg",
		"docker":           "docker",
		"aws":              "aws",
		"kube":             "kubernetes",
		"fish":             "fish",
	}

	// Check for exact pattern match first
	suggestedLower := strings.ToLower(suggested)
	if pkg, ok := patterns[suggestedLower]; ok {
		return pkg
	}

	// Then check for substring matches (for cases like .bash_profile containing "bash")
	for pattern, pkg := range patterns {
		if strings.Contains(suggestedLower, pattern) {
			return pkg
		}
	}

	// Default: use the name itself (cleaned)
	return suggested
}

// GroupByCategory groups candidates by their suggested package names.
// Provides intelligent defaults but user can override.
func GroupByCategory(candidates []DotfileCandidate, selections []int) map[string][]DotfileCandidate {
	groups := make(map[string][]DotfileCandidate)

	for _, idx := range selections {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		candidate := candidates[idx]
		pkg := candidate.SuggestedPkg
		groups[pkg] = append(groups[pkg], candidate)
	}

	return groups
}
