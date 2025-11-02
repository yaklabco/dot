package updater

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// PackageManager represents a system package manager.
type PackageManager interface {
	// Name returns the package manager name
	Name() string
	// IsAvailable checks if the package manager is available on the system
	IsAvailable() bool
	// UpgradeCommand returns the command to upgrade dot
	UpgradeCommand() []string
	// Validate validates the package manager and its upgrade command for security
	Validate() error
}

// validateCommand validates a command array for security concerns.
// It checks for shell metacharacters and other potentially dangerous patterns.
func validateCommand(cmd []string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}

	// Dangerous shell metacharacters that should not appear in command arguments
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r", "&&", "||"}

	for i, arg := range cmd {
		// Check for shell metacharacters in all arguments
		for _, char := range dangerousChars {
			if strings.Contains(arg, char) {
				return fmt.Errorf("command argument %d contains shell metacharacter %q: %s", i, char, arg)
			}
		}

		// Check for null bytes
		if strings.Contains(arg, "\x00") {
			return fmt.Errorf("command argument %d contains null byte", i)
		}
	}

	return nil
}

// allowedPackageManagers is a whitelist of supported package managers
var allowedPackageManagers = map[string]bool{
	"brew":   true,
	"apt":    true,
	"yum":    true,
	"pacman": true,
	"dnf":    true,
	"zypper": true,
	"manual": true,
}

// validatePackageManager validates that the package manager name is in the allowed list.
func validatePackageManager(name string) error {
	if !allowedPackageManagers[name] {
		return fmt.Errorf("unsupported package manager: %s (allowed: brew, apt, yum, pacman, dnf, zypper, manual)", name)
	}
	return nil
}

// BrewManager represents Homebrew package manager.
type BrewManager struct{}

func (b *BrewManager) Name() string {
	return "brew"
}

func (b *BrewManager) IsAvailable() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func (b *BrewManager) UpgradeCommand() []string {
	return []string{"brew", "upgrade", "dot"}
}

func (b *BrewManager) Validate() error {
	if err := validatePackageManager(b.Name()); err != nil {
		return err
	}
	return validateCommand(b.UpgradeCommand())
}

// AptManager represents APT package manager.
type AptManager struct{}

func (a *AptManager) Name() string {
	return "apt"
}

func (a *AptManager) IsAvailable() bool {
	_, err := exec.LookPath("apt")
	return err == nil
}

func (a *AptManager) UpgradeCommand() []string {
	// Use apt-get install --only-upgrade which works with exec.Command
	// Note: This assumes the package is already installed
	return []string{"sudo", "apt-get", "install", "--only-upgrade", "-y", "dot"}
}

func (a *AptManager) Validate() error {
	if err := validatePackageManager(a.Name()); err != nil {
		return err
	}
	return validateCommand(a.UpgradeCommand())
}

// YumManager represents YUM package manager.
type YumManager struct{}

func (y *YumManager) Name() string {
	return "yum"
}

func (y *YumManager) IsAvailable() bool {
	_, err := exec.LookPath("yum")
	return err == nil
}

func (y *YumManager) UpgradeCommand() []string {
	return []string{"sudo", "yum", "upgrade", "-y", "dot"}
}

func (y *YumManager) Validate() error {
	if err := validatePackageManager(y.Name()); err != nil {
		return err
	}
	return validateCommand(y.UpgradeCommand())
}

// PacmanManager represents Pacman package manager.
type PacmanManager struct{}

func (p *PacmanManager) Name() string {
	return "pacman"
}

func (p *PacmanManager) IsAvailable() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

func (p *PacmanManager) UpgradeCommand() []string {
	return []string{"sudo", "pacman", "-Syu", "--noconfirm", "dot"}
}

func (p *PacmanManager) Validate() error {
	if err := validatePackageManager(p.Name()); err != nil {
		return err
	}
	return validateCommand(p.UpgradeCommand())
}

// DnfManager represents DNF package manager.
type DnfManager struct{}

func (d *DnfManager) Name() string {
	return "dnf"
}

func (d *DnfManager) IsAvailable() bool {
	_, err := exec.LookPath("dnf")
	return err == nil
}

func (d *DnfManager) UpgradeCommand() []string {
	return []string{"sudo", "dnf", "upgrade", "-y", "dot"}
}

func (d *DnfManager) Validate() error {
	if err := validatePackageManager(d.Name()); err != nil {
		return err
	}
	return validateCommand(d.UpgradeCommand())
}

// ZypperManager represents Zypper package manager.
type ZypperManager struct{}

func (z *ZypperManager) Name() string {
	return "zypper"
}

func (z *ZypperManager) IsAvailable() bool {
	_, err := exec.LookPath("zypper")
	return err == nil
}

func (z *ZypperManager) UpgradeCommand() []string {
	return []string{"sudo", "zypper", "update", "-y", "dot"}
}

func (z *ZypperManager) Validate() error {
	if err := validatePackageManager(z.Name()); err != nil {
		return err
	}
	return validateCommand(z.UpgradeCommand())
}

// ManualManager represents manual installation (download from GitHub).
type ManualManager struct{}

func (m *ManualManager) Name() string {
	return "manual"
}

func (m *ManualManager) IsAvailable() bool {
	return true // Always available as fallback
}

func (m *ManualManager) UpgradeCommand() []string {
	// This will be handled specially by showing GitHub release URL
	return []string{}
}

func (m *ManualManager) Validate() error {
	// Manual manager doesn't execute commands, so validation always passes
	return nil
}

// GetPackageManager returns the appropriate package manager based on the name.
func GetPackageManager(name string) (PackageManager, error) {
	switch name {
	case "brew":
		return &BrewManager{}, nil
	case "apt":
		return &AptManager{}, nil
	case "yum":
		return &YumManager{}, nil
	case "pacman":
		return &PacmanManager{}, nil
	case "dnf":
		return &DnfManager{}, nil
	case "zypper":
		return &ZypperManager{}, nil
	case "manual":
		return &ManualManager{}, nil
	default:
		return nil, fmt.Errorf("unknown package manager: %s", name)
	}
}

// DetectPackageManager attempts to detect the system package manager.
func DetectPackageManager() PackageManager {
	managers := []PackageManager{
		&BrewManager{},
		&AptManager{},
		&DnfManager{},
		&YumManager{},
		&PacmanManager{},
		&ZypperManager{},
	}

	// On macOS, prefer Homebrew
	if runtime.GOOS == "darwin" {
		brew := &BrewManager{}
		if brew.IsAvailable() {
			return brew
		}
	}

	// Check each package manager
	for _, mgr := range managers {
		if mgr.IsAvailable() {
			return mgr
		}
	}

	// Fallback to manual
	return &ManualManager{}
}

// ResolvePackageManager resolves the package manager from config.
// If "auto" is specified, it detects the system package manager.
func ResolvePackageManager(configuredManager string) (PackageManager, error) {
	if configuredManager == "auto" {
		return DetectPackageManager(), nil
	}

	mgr, err := GetPackageManager(configuredManager)
	if err != nil {
		return nil, err
	}

	if !mgr.IsAvailable() {
		return nil, fmt.Errorf("package manager %s is not available on this system", configuredManager)
	}

	return mgr, nil
}
