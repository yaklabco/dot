package install

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPacmanProbe_Name(t *testing.T) {
	probe := NewPacmanProbe(OSFileSystem{})
	assert.Equal(t, "pacman", probe.Name())
}

func TestPacmanProbe_Platforms(t *testing.T) {
	probe := NewPacmanProbe(OSFileSystem{})
	platforms := probe.Platforms()
	assert.Contains(t, platforms, "linux")
	assert.Len(t, platforms, 1)
}

func TestPacmanProbe_Detect_NoPacmanDatabase(t *testing.T) {
	mockFS := &MockFileSystem{
		Files: map[string][]byte{},
		Dirs:  map[string][]os.DirEntry{},
	}

	probe := NewPacmanProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	assert.NoError(t, err)
	assert.Nil(t, info)
}

func TestPacmanProbe_Detect_PackageInstalled(t *testing.T) {
	descContent := `%NAME%
dot

%VERSION%
1.5.0

%ARCH%
x86_64

%URL%
https://github.com/yaklabco/dot

%DESC%
Dotfile manager for your shell
`

	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/var/lib/pacman/local/dot-1.5.0/desc": []byte(descContent),
		},
		Dirs: map[string][]os.DirEntry{
			"/var/lib/pacman/local": {
				mockDirEntry{name: "bash-5.1", isDir: true},
				mockDirEntry{name: "dot-1.5.0", isDir: true},
				mockDirEntry{name: "vim-8.2", isDir: true},
			},
		},
	}

	probe := NewPacmanProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, SourcePacman, info.Source)
	assert.Equal(t, "1.5.0", info.Version)
	assert.Equal(t, "dot", info.Metadata["name"])
	assert.Equal(t, "x86_64", info.Metadata["arch"])
	assert.Equal(t, "https://github.com/yaklabco/dot", info.Metadata["url"])
	assert.True(t, info.CanAutoUpgrade)
	assert.Equal(t, "sudo pacman -Syu dot", info.UpgradeInstructions)
}

func TestPacmanProbe_Detect_PackageNotInstalled(t *testing.T) {
	mockFS := &MockFileSystem{
		Files: map[string][]byte{},
		Dirs: map[string][]os.DirEntry{
			"/var/lib/pacman/local": {
				mockDirEntry{name: "bash-5.1", isDir: true},
				mockDirEntry{name: "vim-8.2", isDir: true},
			},
		},
	}

	probe := NewPacmanProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	assert.NoError(t, err)
	assert.Nil(t, info)
}

func TestPacmanProbe_ParseDescFile(t *testing.T) {
	descContent := `%NAME%
mypackage

%VERSION%
2.0.0

%DESC%
A test package
`

	probe := NewPacmanProbe(nil)
	info := probe.parseDescFile([]byte(descContent))

	// Should return nil because name is not "dot"
	assert.Nil(t, info)
}

func TestPacmanProbe_Detect_FileEntry(t *testing.T) {
	// Test that file entries (not directories) are skipped
	mockFS := &MockFileSystem{
		Files: map[string][]byte{},
		Dirs: map[string][]os.DirEntry{
			"/var/lib/pacman/local": {
				mockDirEntry{name: "dot-1.0.0", isDir: false}, // File, not dir
			},
		},
	}

	probe := NewPacmanProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	assert.NoError(t, err)
	assert.Nil(t, info)
}
