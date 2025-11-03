package manifest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPopulateMissingMetadata_EmptyFields(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{
		Name:        "vim",
		InstalledAt: time.Now(),
		LinkCount:   2,
		Links:       []string{".vimrc"},
		// TargetDir and PackageDir not set
	})

	PopulateMissingMetadata(&m, "/home/user", "/home/user/dotfiles")

	pkg, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, "/home/user", pkg.TargetDir)
	assert.Equal(t, "/home/user/dotfiles", pkg.PackageDir)
}

func TestPopulateMissingMetadata_PartialFields(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{
		Name:       "vim",
		LinkCount:  2,
		Links:      []string{".vimrc"},
		TargetDir:  "/custom/target",
		PackageDir: "", // Only PackageDir missing
	})

	PopulateMissingMetadata(&m, "/home/user", "/home/user/dotfiles")

	pkg, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, "/custom/target", pkg.TargetDir)       // Should preserve existing
	assert.Equal(t, "/home/user/dotfiles", pkg.PackageDir) // Should populate missing
}

func TestPopulateMissingMetadata_AllFieldsSet(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{
		Name:       "vim",
		LinkCount:  2,
		Links:      []string{".vimrc"},
		TargetDir:  "/custom/target",
		PackageDir: "/custom/package",
	})

	PopulateMissingMetadata(&m, "/home/user", "/home/user/dotfiles")

	pkg, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, "/custom/target", pkg.TargetDir)   // Should preserve existing
	assert.Equal(t, "/custom/package", pkg.PackageDir) // Should preserve existing
}

func TestPopulateMissingMetadata_MultiplePackages(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 2,
		Links:     []string{".vimrc"},
		// No directories
	})
	m.AddPackage(PackageInfo{
		Name:       "zsh",
		LinkCount:  1,
		Links:      []string{".zshrc"},
		TargetDir:  "/custom/target",
		PackageDir: "/custom/package",
	})

	PopulateMissingMetadata(&m, "/home/user", "/home/user/dotfiles")

	// Check vim got populated
	vim, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, "/home/user", vim.TargetDir)
	assert.Equal(t, "/home/user/dotfiles", vim.PackageDir)

	// Check zsh preserved its values
	zsh, exists := m.GetPackage("zsh")
	assert.True(t, exists)
	assert.Equal(t, "/custom/target", zsh.TargetDir)
	assert.Equal(t, "/custom/package", zsh.PackageDir)
}

func TestPopulateMissingMetadata_EmptyManifest(t *testing.T) {
	m := New()

	// Should not panic on empty manifest
	PopulateMissingMetadata(&m, "/home/user", "/home/user/dotfiles")

	assert.Equal(t, 0, len(m.Packages))
}
