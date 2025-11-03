package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsReservedPackageName(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		expected bool
	}{
		{"dot lowercase", "dot", true},
		{"dot uppercase", "DOT", true},
		{"dot mixed case", "Dot", true},
		{"hidden dot", ".dot", true},
		{"dot-config", "dot-config", true},
		{"DOT-CONFIG uppercase", "DOT-CONFIG", true},
		{"regular package", "vim", false},
		{"package with dot prefix", "dotfiles", false},
		{"package with dot suffix", "mydot", false},
		{"empty string", "", false},
		{"zsh package", "zsh", false},
		{"tmux package", "tmux", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReservedPackageName(tt.pkgName)
			assert.Equal(t, tt.expected, result, "IsReservedPackageName(%q) = %v, want %v", tt.pkgName, result, tt.expected)
		})
	}
}

func TestGetReservedPackageReason(t *testing.T) {
	reason := GetReservedPackageReason("dot")
	assert.NotEmpty(t, reason)
	assert.Contains(t, reason, "reserved")
	assert.Contains(t, reason, "internal")
}
