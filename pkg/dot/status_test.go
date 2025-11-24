package dot_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestStatus(t *testing.T) {
	now := time.Now()
	status := dot.Status{
		Packages: []dot.PackageInfo{
			{
				Name:        "vim",
				InstalledAt: now,
				LinkCount:   3,
				Links:       []string{".vimrc", ".vim/colors/"},
			},
		},
	}

	require.Len(t, status.Packages, 1)
	require.Equal(t, "vim", status.Packages[0].Name)
	require.Equal(t, 3, status.Packages[0].LinkCount)
	require.Equal(t, now, status.Packages[0].InstalledAt)
}

func TestPackageInfo(t *testing.T) {
	now := time.Now()
	info := dot.PackageInfo{
		Name:        "zsh",
		Source:      "managed",
		InstalledAt: now,
		LinkCount:   5,
		Links:       []string{".zshrc", ".zshenv", ".zsh/"},
	}

	require.Equal(t, "zsh", info.Name)
	require.Equal(t, "managed", info.Source)
	require.Equal(t, 5, info.LinkCount)
	require.Len(t, info.Links, 3)
}

func TestPackageInfo_AdoptedSource(t *testing.T) {
	now := time.Now()
	info := dot.PackageInfo{
		Name:        "ssh",
		Source:      "adopted",
		InstalledAt: now,
		LinkCount:   2,
		Links:       []string{".ssh/config", ".ssh/known_hosts"},
	}

	require.Equal(t, "ssh", info.Name)
	require.Equal(t, "adopted", info.Source)
	require.Equal(t, 2, info.LinkCount)
}

func TestStatusEmpty(t *testing.T) {
	status := dot.Status{
		Packages: []dot.PackageInfo{},
	}

	require.Empty(t, status.Packages)
}
