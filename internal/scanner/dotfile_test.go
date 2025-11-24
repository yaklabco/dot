package scanner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/scanner"
)

func TestTranslateDotfile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "dot- prefix becomes dot",
			input:    "dot-vimrc",
			expected: ".vimrc",
		},
		{
			name:     "dot-bashrc becomes .bashrc",
			input:    "dot-bashrc",
			expected: ".bashrc",
		},
		{
			name:     "no translation for regular files",
			input:    "README.md",
			expected: "README.md",
		},
		{
			name:     "already a dotfile",
			input:    ".vimrc",
			expected: ".vimrc",
		},
		{
			name:     "dot- in middle is not translated",
			input:    "some-dot-file",
			expected: "some-dot-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.TranslateDotfile(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUntranslateDotfile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     ".vimrc becomes dot-vimrc",
			input:    ".vimrc",
			expected: "dot-vimrc",
		},
		{
			name:     ".bashrc becomes dot-bashrc",
			input:    ".bashrc",
			expected: "dot-bashrc",
		},
		{
			name:     "no translation for regular files",
			input:    "README.md",
			expected: "README.md",
		},
		{
			name:     "dot-vimrc stays as is",
			input:    "dot-vimrc",
			expected: "dot-vimrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.UntranslateDotfile(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTripTranslation(t *testing.T) {
	// Test that translation is bidirectional
	tests := []string{
		"dot-vimrc",
		".vimrc",
		"README.md",
	}

	for _, original := range tests {
		t.Run(original, func(t *testing.T) {
			// For dot- prefixed files, round trip should work
			if len(original) > 4 && original[:4] == "dot-" {
				dotted := scanner.TranslateDotfile(original)
				backToDot := scanner.UntranslateDotfile(dotted)
				assert.Equal(t, original, backToDot)
			}
		})
	}
}

func TestTranslatePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple dot- file",
			input:    "dot-vimrc",
			expected: ".vimrc",
		},
		{
			name:     "nested path with dot-",
			input:    "vim/dot-vimrc",
			expected: "vim/.vimrc",
		},
		{
			name:     "multiple levels",
			input:    "a/b/dot-config",
			expected: "a/b/.config",
		},
		{
			name:     "no translation needed",
			input:    "vim/README.md",
			expected: "vim/README.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.TranslatePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUntranslatePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple dotfile",
			input:    ".vimrc",
			expected: "dot-vimrc",
		},
		{
			name:     "nested path with dotfile",
			input:    "vim/.vimrc",
			expected: "vim/dot-vimrc",
		},
		{
			name:     "multiple levels",
			input:    "a/b/.config",
			expected: "a/b/dot-config",
		},
		{
			name:     "no translation needed",
			input:    "vim/README.md",
			expected: "vim/README.md",
		},
		{
			name:     "root level dotfile",
			input:    ".bashrc",
			expected: "dot-bashrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.UntranslatePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTranslatePackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "dot- prefix becomes dot",
			input:    "dot-gnupg",
			expected: ".gnupg",
		},
		{
			name:     "dot-config becomes .config",
			input:    "dot-config",
			expected: ".config",
		},
		{
			name:     "dot-vim becomes .vim",
			input:    "dot-vim",
			expected: ".vim",
		},
		{
			name:     "no prefix stays as is",
			input:    "vim",
			expected: "vim",
		},
		{
			name:     "no prefix with path",
			input:    "config",
			expected: "config",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "just dot-",
			input:    "dot-",
			expected: ".",
		},
		{
			name:     "already dotfile",
			input:    ".gnupg",
			expected: ".gnupg",
		},
		{
			name:     "multiple dashes",
			input:    "dot-my-package",
			expected: ".my-package",
		},
		{
			name:     "dot- in middle not translated",
			input:    "my-dot-package",
			expected: "my-dot-package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.TranslatePackageName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
