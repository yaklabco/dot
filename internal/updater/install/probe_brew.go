package install

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
)

// brewCellarPaths contains known Homebrew Cellar locations by platform.
var brewCellarPaths = map[string][]string{
	"darwin": {
		"/opt/homebrew/Cellar", // Apple Silicon
		"/usr/local/Cellar",    // Intel
	},
	"linux": {
		"/home/linuxbrew/.linuxbrew/Cellar",
		"/usr/local/Cellar",
	},
}

// BrewProbe detects Homebrew installations.
type BrewProbe struct {
	fs FileSystem
}

// NewBrewProbe creates a new Homebrew probe.
func NewBrewProbe(fs FileSystem) *BrewProbe {
	return &BrewProbe{fs: fs}
}

// Name returns the probe identifier.
func (p *BrewProbe) Name() string {
	return "homebrew"
}

// Platforms returns the platforms this probe supports.
func (p *BrewProbe) Platforms() []string {
	return []string{"darwin", "linux"}
}

// Detect checks if dot was installed via Homebrew.
func (p *BrewProbe) Detect(ctx context.Context, execPath string) (*Info, error) {
	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		// Not a symlink or error resolving - not Homebrew
		return nil, nil //nolint:nilerr // nil means no match, not error
	}

	// Check if path is within a known Cellar location
	cellars, ok := brewCellarPaths[runtime.GOOS]
	if !ok {
		return nil, nil
	}

	for _, cellar := range cellars {
		if strings.HasPrefix(realPath, cellar) {
			return p.parseBrewInstall(realPath, cellar)
		}
	}

	return nil, nil
}

// parseBrewInstall extracts installation info from the Cellar path.
func (p *BrewProbe) parseBrewInstall(realPath, cellar string) (*Info, error) {
	// Path format: /opt/homebrew/Cellar/dot/0.6.3/bin/dot
	// Extract version from path
	relPath, err := filepath.Rel(cellar, realPath)
	if err != nil {
		return nil, nil //nolint:nilerr // cannot determine relative path
	}

	// relPath: "dot/0.6.3/bin/dot"
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		return nil, nil
	}

	formulaName := parts[0]
	version := parts[1]

	// Build metadata
	metadata := map[string]string{
		"formula": formulaName,
		"cellar":  cellar,
	}

	// Try to read INSTALL_RECEIPT.json for tap information
	receiptPath := filepath.Join(cellar, formulaName, version, "INSTALL_RECEIPT.json")
	if data, err := p.fs.ReadFile(receiptPath); err == nil {
		var receipt brewReceipt
		if json.Unmarshal(data, &receipt) == nil && receipt.Source.Tap != "" {
			metadata["tap"] = receipt.Source.Tap
		}
	}

	// Build formula reference for upgrade command
	formulaRef := formulaName
	if tap, ok := metadata["tap"]; ok && tap != "" {
		formulaRef = tap + "/" + formulaName
	}

	return &Info{
		Source:              SourceHomebrew,
		Version:             version,
		ExecutablePath:      realPath,
		Metadata:            metadata,
		CanAutoUpgrade:      true,
		UpgradeInstructions: "brew upgrade " + formulaRef,
	}, nil
}

// brewReceipt represents the structure of Homebrew's INSTALL_RECEIPT.json.
type brewReceipt struct {
	Source struct {
		Tap string `json:"tap"`
	} `json:"source"`
}

// Ensure BrewProbe implements Probe.
var _ Probe = (*BrewProbe)(nil)
