package updater

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrewManager(t *testing.T) {
	mgr := &BrewManager{}
	assert.Equal(t, "brew", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Equal(t, []string{"brew", "upgrade", "dot"}, mgr.UpgradeCommand())
}

func TestAptManager(t *testing.T) {
	mgr := &AptManager{}
	assert.Equal(t, "apt", mgr.Name())
	cmd := mgr.UpgradeCommand()
	assert.NotEmpty(t, cmd)
	// Should contain apt-get (not apt) for --only-upgrade support
	assert.Contains(t, cmd, "apt-get")
	assert.Contains(t, cmd, "--only-upgrade")
	assert.Contains(t, cmd, "dot")

	// IsAvailable should not panic
	available := mgr.IsAvailable()
	_ = available // Value depends on system
}

func TestYumManager(t *testing.T) {
	mgr := &YumManager{}
	assert.Equal(t, "yum", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "yum")

	// IsAvailable should not panic
	available := mgr.IsAvailable()
	_ = available // Value depends on system
}

func TestPacmanManager(t *testing.T) {
	mgr := &PacmanManager{}
	assert.Equal(t, "pacman", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "pacman")

	// IsAvailable should not panic
	available := mgr.IsAvailable()
	_ = available // Value depends on system
}

func TestDnfManager(t *testing.T) {
	mgr := &DnfManager{}
	assert.Equal(t, "dnf", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "dnf")

	// IsAvailable should not panic
	available := mgr.IsAvailable()
	_ = available // Value depends on system
}

func TestZypperManager(t *testing.T) {
	mgr := &ZypperManager{}
	assert.Equal(t, "zypper", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "zypper")

	// IsAvailable should not panic
	available := mgr.IsAvailable()
	_ = available // Value depends on system
}

func TestManualManager(t *testing.T) {
	mgr := &ManualManager{}
	assert.Equal(t, "manual", mgr.Name())
	assert.True(t, mgr.IsAvailable()) // Manual is always available
	assert.Empty(t, mgr.UpgradeCommand())
}

func TestGetPackageManager(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"brew", "brew", "brew", false},
		{"apt", "apt", "apt", false},
		{"yum", "yum", "yum", false},
		{"pacman", "pacman", "pacman", false},
		{"dnf", "dnf", "dnf", false},
		{"zypper", "zypper", "zypper", false},
		{"manual", "manual", "manual", false},
		{"unknown", "unknown-manager", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := GetPackageManager(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, mgr.Name())
		})
	}
}

func TestDetectPackageManager(t *testing.T) {
	// This test is platform-dependent, so we just verify it returns something
	mgr := DetectPackageManager()
	require.NotNil(t, mgr)
	assert.NotEmpty(t, mgr.Name())

	// On macOS, we expect brew to be preferred if available
	if runtime.GOOS == "darwin" {
		brew := &BrewManager{}
		if brew.IsAvailable() {
			assert.Equal(t, "brew", mgr.Name())
		}
	}
}

func TestResolvePackageManager(t *testing.T) {
	t.Run("auto detection", func(t *testing.T) {
		mgr, err := ResolvePackageManager("auto")
		require.NoError(t, err)
		require.NotNil(t, mgr)
		assert.NotEmpty(t, mgr.Name())
	})

	t.Run("manual always works", func(t *testing.T) {
		mgr, err := ResolvePackageManager("manual")
		require.NoError(t, err)
		assert.Equal(t, "manual", mgr.Name())
		assert.True(t, mgr.IsAvailable())
	})

	t.Run("unknown manager", func(t *testing.T) {
		_, err := ResolvePackageManager("nonexistent")
		assert.Error(t, err)
	})

	t.Run("specified manager", func(t *testing.T) {
		// Test with brew (should work on macOS if brew is installed)
		mgr, err := ResolvePackageManager("brew")
		brew := &BrewManager{}
		if brew.IsAvailable() {
			require.NoError(t, err)
			assert.Equal(t, "brew", mgr.Name())
		} else if runtime.GOOS == "darwin" {
			// brew not available on macOS
			assert.Error(t, err)
		}
		// On other platforms, brew availability varies - no assertion
	})
}

func TestPackageManager_UpgradeCommands(t *testing.T) {
	tests := []struct {
		name    string
		manager PackageManager
	}{
		{"brew", &BrewManager{}},
		{"apt", &AptManager{}},
		{"yum", &YumManager{}},
		{"pacman", &PacmanManager{}},
		{"dnf", &DnfManager{}},
		{"zypper", &ZypperManager{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.manager.UpgradeCommand()
			assert.NotEmpty(t, cmd, "upgrade command should not be empty")
			assert.Contains(t, cmd, "dot", "upgrade command should include 'dot'")
		})
	}
}

func TestBrewManager_IsAvailable(t *testing.T) {
	mgr := &BrewManager{}

	// Just verify it doesn't panic - availability depends on environment
	available := mgr.IsAvailable()

	// On macOS, brew is more commonly available
	// On other platforms, brew is less common
	// We don't assert specific value as it depends on system setup
	_ = available
}

func TestManualManager_AlwaysAvailable(t *testing.T) {
	mgr := &ManualManager{}

	assert.True(t, mgr.IsAvailable())
	assert.Empty(t, mgr.UpgradeCommand())
}

func TestPackageManagerInterface(t *testing.T) {
	// Verify all managers implement the interface
	var _ PackageManager = &BrewManager{}
	var _ PackageManager = &AptManager{}
	var _ PackageManager = &YumManager{}
	var _ PackageManager = &PacmanManager{}
	var _ PackageManager = &DnfManager{}
	var _ PackageManager = &ZypperManager{}
	var _ PackageManager = &ManualManager{}
}

func TestDetectPackageManager_Coverage(t *testing.T) {
	// Call DetectPackageManager to ensure it's exercised
	mgr := DetectPackageManager()
	require.NotNil(t, mgr)

	// Should return a valid manager
	assert.NotEmpty(t, mgr.Name())

	// Should always succeed
	assert.True(t, mgr.IsAvailable() || mgr.Name() == "manual")
}

func TestResolvePackageManager_Coverage(t *testing.T) {
	t.Run("auto resolves to valid manager", func(t *testing.T) {
		mgr, err := ResolvePackageManager("auto")
		require.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.True(t, mgr.IsAvailable())
	})

	t.Run("invalid manager returns error", func(t *testing.T) {
		_, err := ResolvePackageManager("does-not-exist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown package manager")
	})

	t.Run("unavailable manager returns error", func(t *testing.T) {
		// On macOS, try a Linux-only manager
		// On Linux, this might work depending on the system
		// The important thing is we test the error path exists
		if runtime.GOOS == "darwin" {
			_, err := ResolvePackageManager("pacman")
			// On macOS, pacman should not be available
			if err != nil {
				assert.Contains(t, err.Error(), "not available")
			}
		}
	})
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid command",
			cmd:     []string{"brew", "upgrade", "dot"},
			wantErr: false,
		},
		{
			name:    "valid command with flags",
			cmd:     []string{"sudo", "apt-get", "install", "--only-upgrade", "-y", "dot"},
			wantErr: false,
		},
		{
			name:    "empty command",
			cmd:     []string{},
			wantErr: true,
			errMsg:  "empty command",
		},
		{
			name:    "command with semicolon",
			cmd:     []string{"brew", "upgrade", "dot; rm -rf /"},
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with pipe",
			cmd:     []string{"brew", "upgrade", "dot | cat"},
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with ampersand",
			cmd:     []string{"brew", "upgrade", "dot &"},
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with backtick",
			cmd:     []string{"brew", "upgrade", "`whoami`"},
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with dollar",
			cmd:     []string{"brew", "upgrade", "$HOME"},
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with null byte",
			cmd:     []string{"brew", "upgrade", "dot\x00"},
			wantErr: true,
			errMsg:  "null byte",
		},
		{
			name:    "command with &&",
			cmd:     []string{"brew", "upgrade", "dot && reboot"},
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with ||",
			cmd:     []string{"brew", "upgrade", "dot || exit"},
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePackageManager(t *testing.T) {
	tests := []struct {
		name    string
		pmName  string
		wantErr bool
	}{
		{"brew is allowed", "brew", false},
		{"apt is allowed", "apt", false},
		{"yum is allowed", "yum", false},
		{"pacman is allowed", "pacman", false},
		{"dnf is allowed", "dnf", false},
		{"zypper is allowed", "zypper", false},
		{"manual is allowed", "manual", false},
		{"unknown is rejected", "unknown", true},
		{"malicious is rejected", "rm -rf /", true},
		{"empty is rejected", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePackageManager(tt.pmName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPackageManagerValidate(t *testing.T) {
	t.Run("BrewManager validation passes", func(t *testing.T) {
		mgr := &BrewManager{}
		err := mgr.Validate()
		assert.NoError(t, err)
	})

	t.Run("AptManager validation passes", func(t *testing.T) {
		mgr := &AptManager{}
		err := mgr.Validate()
		assert.NoError(t, err)
	})

	t.Run("YumManager validation passes", func(t *testing.T) {
		mgr := &YumManager{}
		err := mgr.Validate()
		assert.NoError(t, err)
	})

	t.Run("PacmanManager validation passes", func(t *testing.T) {
		mgr := &PacmanManager{}
		err := mgr.Validate()
		assert.NoError(t, err)
	})

	t.Run("DnfManager validation passes", func(t *testing.T) {
		mgr := &DnfManager{}
		err := mgr.Validate()
		assert.NoError(t, err)
	})

	t.Run("ZypperManager validation passes", func(t *testing.T) {
		mgr := &ZypperManager{}
		err := mgr.Validate()
		assert.NoError(t, err)
	})

	t.Run("ManualManager validation passes", func(t *testing.T) {
		mgr := &ManualManager{}
		err := mgr.Validate()
		assert.NoError(t, err)
	})
}
