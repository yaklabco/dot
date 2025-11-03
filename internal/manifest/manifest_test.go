package manifest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifest_New(t *testing.T) {
	m := New()

	assert.Equal(t, "1.0", m.Version)
	assert.NotNil(t, m.Packages)
	assert.NotNil(t, m.Hashes)
	assert.False(t, m.UpdatedAt.IsZero())
}

func TestManifest_AddPackage(t *testing.T) {
	m := New()

	pkg := PackageInfo{
		Name:        "vim",
		InstalledAt: time.Now(),
		LinkCount:   5,
		Links:       []string{".vimrc", ".vim/colors"},
	}

	m.AddPackage(pkg)

	retrieved, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, "vim", retrieved.Name)
	assert.Equal(t, 5, retrieved.LinkCount)
	assert.Len(t, retrieved.Links, 2)
}

func TestManifest_RemovePackage(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})

	removed := m.RemovePackage("vim")

	assert.True(t, removed)
	_, exists := m.GetPackage("vim")
	assert.False(t, exists)
}

func TestManifest_RemovePackage_NotExists(t *testing.T) {
	m := New()

	removed := m.RemovePackage("nonexistent")

	assert.False(t, removed)
}

func TestManifest_SetHash(t *testing.T) {
	m := New()

	m.SetHash("vim", "abc123")

	hash, exists := m.GetHash("vim")
	assert.True(t, exists)
	assert.Equal(t, "abc123", hash)
}

func TestManifest_GetHash_NotExists(t *testing.T) {
	m := New()

	hash, exists := m.GetHash("nonexistent")

	assert.False(t, exists)
	assert.Empty(t, hash)
}

func TestManifest_PackageList(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})
	m.AddPackage(PackageInfo{Name: "zsh"})

	packages := m.PackageList()

	assert.Len(t, packages, 2)
	names := []string{packages[0].Name, packages[1].Name}
	assert.Contains(t, names, "vim")
	assert.Contains(t, names, "zsh")
}

func TestManifest_PackageList_Empty(t *testing.T) {
	m := New()

	packages := m.PackageList()

	assert.Empty(t, packages)
}

func TestManifest_RemovePackage_RemovesHash(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})
	m.SetHash("vim", "abc123")

	removed := m.RemovePackage("vim")

	assert.True(t, removed)
	_, exists := m.GetHash("vim")
	assert.False(t, exists)
}

func TestManifest_UpdatesTimestamp_OnAdd(t *testing.T) {
	m := New()
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.AddPackage(PackageInfo{Name: "vim"})

	assert.True(t, m.UpdatedAt.After(originalTime))
}

func TestManifest_UpdatesTimestamp_OnRemove(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})
	time.Sleep(10 * time.Millisecond)
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.RemovePackage("vim")

	assert.True(t, m.UpdatedAt.After(originalTime))
}

func TestManifest_UpdatesTimestamp_OnSetHash(t *testing.T) {
	m := New()
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.SetHash("vim", "abc123")

	assert.True(t, m.UpdatedAt.After(originalTime))
}

func TestManifest_SetRepository(t *testing.T) {
	m := New()

	repo := RepositoryInfo{
		URL:       "https://github.com/user/dotfiles",
		Branch:    "main",
		ClonedAt:  time.Now(),
		CommitSHA: "abc123def456",
	}

	m.SetRepository(repo)

	retrieved, exists := m.GetRepository()
	assert.True(t, exists)
	assert.Equal(t, "https://github.com/user/dotfiles", retrieved.URL)
	assert.Equal(t, "main", retrieved.Branch)
	assert.Equal(t, "abc123def456", retrieved.CommitSHA)
}

func TestManifest_GetRepository_NotSet(t *testing.T) {
	m := New()

	repo, exists := m.GetRepository()
	assert.False(t, exists)
	assert.Equal(t, RepositoryInfo{}, repo)
}

func TestManifest_ClearRepository(t *testing.T) {
	m := New()
	m.SetRepository(RepositoryInfo{
		URL:    "https://github.com/user/dotfiles",
		Branch: "main",
	})

	m.ClearRepository()

	repo, exists := m.GetRepository()
	assert.False(t, exists)
	assert.Equal(t, RepositoryInfo{}, repo)
}

func TestManifest_UpdatesTimestamp_OnSetRepository(t *testing.T) {
	m := New()
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.SetRepository(RepositoryInfo{
		URL:    "https://github.com/user/dotfiles",
		Branch: "main",
	})

	assert.True(t, m.UpdatedAt.After(originalTime))
}

func TestManifest_UpdatesTimestamp_OnClearRepository(t *testing.T) {
	m := New()
	m.SetRepository(RepositoryInfo{
		URL:    "https://github.com/user/dotfiles",
		Branch: "main",
	})
	time.Sleep(10 * time.Millisecond)
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.ClearRepository()

	assert.True(t, m.UpdatedAt.After(originalTime))
}

func TestManifest_JSON_WithRepository(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim", LinkCount: 2})
	m.SetRepository(RepositoryInfo{
		URL:       "https://github.com/user/dotfiles",
		Branch:    "main",
		ClonedAt:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		CommitSHA: "abc123",
	})

	// Marshal to JSON
	data, err := json.Marshal(m)
	require.NoError(t, err)

	// Unmarshal back
	var loaded Manifest
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	// Verify repository info is preserved
	repo, exists := loaded.GetRepository()
	assert.True(t, exists)
	assert.Equal(t, "https://github.com/user/dotfiles", repo.URL)
	assert.Equal(t, "main", repo.Branch)
	assert.Equal(t, "abc123", repo.CommitSHA)
}

func TestManifest_JSON_WithoutRepository(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim", LinkCount: 2})

	// Marshal to JSON
	data, err := json.Marshal(m)
	require.NoError(t, err)

	// Verify repository field is omitted when nil
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)
	_, hasRepo := raw["repository"]
	assert.False(t, hasRepo, "repository field should be omitted when nil")

	// Unmarshal back
	var loaded Manifest
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	// Verify no repository info
	repo, exists := loaded.GetRepository()
	assert.False(t, exists)
	assert.Equal(t, RepositoryInfo{}, repo)
}

func TestManifest_JSON_BackwardCompatibility(t *testing.T) {
	// Old manifest format without repository field
	oldJSON := `{
		"version": "1.0",
		"updated_at": "2025-01-01T12:00:00Z",
		"packages": {
			"vim": {
				"name": "vim",
				"installed_at": "2025-01-01T12:00:00Z",
				"link_count": 2,
				"links": [".vimrc", ".vim"]
			}
		},
		"hashes": {
			"vim": "hash123"
		}
	}`

	var m Manifest
	err := json.Unmarshal([]byte(oldJSON), &m)
	require.NoError(t, err)

	// Verify old fields still work
	assert.Equal(t, "1.0", m.Version)
	pkg, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, "vim", pkg.Name)

	// Verify repository is nil (not set in old format)
	repo, exists := m.GetRepository()
	assert.False(t, exists)
	assert.Equal(t, RepositoryInfo{}, repo)
}

func TestPackageInfo_WithDirectories(t *testing.T) {
	m := New()

	pkg := PackageInfo{
		Name:        "vim",
		InstalledAt: time.Now(),
		LinkCount:   5,
		Links:       []string{".vimrc", ".vim/colors"},
		TargetDir:   "/home/user",
		PackageDir:  "/home/user/dotfiles",
	}

	m.AddPackage(pkg)

	retrieved, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, "vim", retrieved.Name)
	assert.Equal(t, "/home/user", retrieved.TargetDir)
	assert.Equal(t, "/home/user/dotfiles", retrieved.PackageDir)
}

func TestPackageInfo_JSON_WithDirectories(t *testing.T) {
	pkg := PackageInfo{
		Name:        "vim",
		InstalledAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		LinkCount:   2,
		Links:       []string{".vimrc"},
		TargetDir:   "/home/user",
		PackageDir:  "/home/user/dotfiles",
		Source:      SourceManaged,
	}

	// Marshal to JSON
	data, err := json.Marshal(pkg)
	require.NoError(t, err)

	// Unmarshal back
	var loaded PackageInfo
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	// Verify all fields preserved
	assert.Equal(t, "vim", loaded.Name)
	assert.Equal(t, "/home/user", loaded.TargetDir)
	assert.Equal(t, "/home/user/dotfiles", loaded.PackageDir)
	assert.Equal(t, SourceManaged, loaded.Source)
}

func TestPackageInfo_JSON_WithoutDirectories(t *testing.T) {
	pkg := PackageInfo{
		Name:        "vim",
		InstalledAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		LinkCount:   2,
		Links:       []string{".vimrc"},
	}

	// Marshal to JSON
	data, err := json.Marshal(pkg)
	require.NoError(t, err)

	// Verify directory fields are omitted when empty
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)
	_, hasTargetDir := raw["target_dir"]
	_, hasPackageDir := raw["package_dir"]
	assert.False(t, hasTargetDir, "target_dir should be omitted when empty")
	assert.False(t, hasPackageDir, "package_dir should be omitted when empty")

	// Unmarshal back
	var loaded PackageInfo
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	// Verify fields are empty strings
	assert.Equal(t, "", loaded.TargetDir)
	assert.Equal(t, "", loaded.PackageDir)
}

func TestPackageInfo_BackwardCompatibility_NoDirectories(t *testing.T) {
	// Old package format without target_dir and package_dir
	oldJSON := `{
		"name": "vim",
		"installed_at": "2025-01-01T12:00:00Z",
		"link_count": 2,
		"links": [".vimrc", ".vim"],
		"source": "managed"
	}`

	var pkg PackageInfo
	err := json.Unmarshal([]byte(oldJSON), &pkg)
	require.NoError(t, err)

	// Verify old fields work
	assert.Equal(t, "vim", pkg.Name)
	assert.Equal(t, 2, pkg.LinkCount)
	assert.Equal(t, SourceManaged, pkg.Source)

	// Verify new fields default to empty
	assert.Equal(t, "", pkg.TargetDir)
	assert.Equal(t, "", pkg.PackageDir)
}
