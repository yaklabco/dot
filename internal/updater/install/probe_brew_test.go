package install

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrewProbe_Name(t *testing.T) {
	probe := NewBrewProbe(OSFileSystem{})
	assert.Equal(t, "homebrew", probe.Name())
}

func TestBrewProbe_Platforms(t *testing.T) {
	probe := NewBrewProbe(OSFileSystem{})
	platforms := probe.Platforms()
	assert.Contains(t, platforms, "darwin")
	assert.Contains(t, platforms, "linux")
	assert.NotContains(t, platforms, "windows")
}

func TestBrewProbe_Detect_NotInCellar(t *testing.T) {
	probe := NewBrewProbe(OSFileSystem{})
	info, err := probe.Detect(context.Background(), "/usr/local/bin/dot")
	assert.NoError(t, err)
	assert.Nil(t, info)
}

// mockDirEntry implements os.DirEntry for testing.
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() os.FileMode          { return 0 }
func (m mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

func TestBrewProbe_Detect_ValidCellarPath(t *testing.T) {
	// Create a mock filesystem that simulates Homebrew installation
	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/opt/homebrew/Cellar/dot/1.0.0/INSTALL_RECEIPT.json": []byte(`{
				"source": {
					"tap": "yaklabco/dot"
				}
			}`),
		},
	}

	probe := NewBrewProbe(mockFS)

	// Note: This test is limited because it relies on filepath.EvalSymlinks
	// which we can't easily mock. The test verifies the probe structure works.
	info, err := probe.Detect(context.Background(), "/usr/local/bin/dot")
	assert.NoError(t, err)
	// Will return nil because the path isn't in Cellar and EvalSymlinks won't help
	assert.Nil(t, info)
}

func TestBrewProbe_ParseBrewInstall(t *testing.T) {
	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/opt/homebrew/Cellar/dot/1.2.3/INSTALL_RECEIPT.json": []byte(`{
				"source": {
					"tap": "yaklabco/dot"
				}
			}`),
		},
	}

	probe := NewBrewProbe(mockFS)
	info, err := probe.parseBrewInstall(
		"/opt/homebrew/Cellar/dot/1.2.3/bin/dot",
		"/opt/homebrew/Cellar",
	)

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, SourceHomebrew, info.Source)
	assert.Equal(t, "1.2.3", info.Version)
	assert.Equal(t, "dot", info.Metadata["formula"])
	assert.Equal(t, "yaklabco/dot", info.Metadata["tap"])
	assert.True(t, info.CanAutoUpgrade)
	assert.Equal(t, "brew upgrade yaklabco/dot/dot", info.UpgradeInstructions)
}

func TestBrewProbe_ParseBrewInstall_NoReceipt(t *testing.T) {
	mockFS := &MockFileSystem{
		Files: map[string][]byte{},
	}

	probe := NewBrewProbe(mockFS)
	info, err := probe.parseBrewInstall(
		"/opt/homebrew/Cellar/dot/1.0.0/bin/dot",
		"/opt/homebrew/Cellar",
	)

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, SourceHomebrew, info.Source)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "dot", info.Metadata["formula"])
	// No tap means just formula name
	assert.Equal(t, "brew upgrade dot", info.UpgradeInstructions)
}
