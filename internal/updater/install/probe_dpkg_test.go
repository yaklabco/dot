package install

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDpkgProbe_Name(t *testing.T) {
	probe := NewDpkgProbe(OSFileSystem{})
	assert.Equal(t, "dpkg", probe.Name())
}

func TestDpkgProbe_Platforms(t *testing.T) {
	probe := NewDpkgProbe(OSFileSystem{})
	platforms := probe.Platforms()
	assert.Contains(t, platforms, "linux")
	assert.Len(t, platforms, 1)
}

func TestDpkgProbe_Detect_NoDpkgDatabase(t *testing.T) {
	mockFS := &MockFileSystem{
		Files: map[string][]byte{},
	}

	probe := NewDpkgProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	assert.NoError(t, err)
	assert.Nil(t, info)
}

func TestDpkgProbe_Detect_PackageInstalled(t *testing.T) {
	dpkgStatus := `Package: bash
Status: install ok installed
Version: 5.1-6ubuntu1

Package: dot
Status: install ok installed
Priority: optional
Section: utils
Installed-Size: 1234
Architecture: amd64
Version: 1.0.0
Description: Dotfile manager

Package: vim
Status: install ok installed
Version: 8.2.0
`

	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/var/lib/dpkg/status": []byte(dpkgStatus),
		},
	}

	probe := NewDpkgProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, SourceApt, info.Source)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "dot", info.Metadata["package"])
	assert.Equal(t, "amd64", info.Metadata["architecture"])
	assert.True(t, info.CanAutoUpgrade)
	assert.Equal(t, "sudo apt-get update && sudo apt-get install --only-upgrade dot", info.UpgradeInstructions)
}

func TestDpkgProbe_Detect_PackageNotInstalled(t *testing.T) {
	dpkgStatus := `Package: bash
Status: install ok installed
Version: 5.1-6ubuntu1

Package: vim
Status: install ok installed
Version: 8.2.0
`

	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/var/lib/dpkg/status": []byte(dpkgStatus),
		},
	}

	probe := NewDpkgProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	assert.NoError(t, err)
	assert.Nil(t, info)
}

func TestDpkgProbe_Detect_PackageNotFullyInstalled(t *testing.T) {
	// Package exists but status is not "install ok installed"
	dpkgStatus := `Package: dot
Status: deinstall ok config-files
Version: 1.0.0
`

	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/var/lib/dpkg/status": []byte(dpkgStatus),
		},
	}

	probe := NewDpkgProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	assert.NoError(t, err)
	assert.Nil(t, info) // Not fully installed
}

func TestDpkgProbe_ParseDpkgStatus_LastPackage(t *testing.T) {
	// Test when dot is the last package in the file (no trailing newline)
	dpkgStatus := `Package: bash
Status: install ok installed
Version: 5.1

Package: dot
Status: install ok installed
Version: 2.0.0
Architecture: arm64`

	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/var/lib/dpkg/status": []byte(dpkgStatus),
		},
	}

	probe := NewDpkgProbe(mockFS)
	info, err := probe.Detect(context.Background(), "/usr/bin/dot")

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "2.0.0", info.Version)
	assert.Equal(t, "arm64", info.Metadata["architecture"])
}
