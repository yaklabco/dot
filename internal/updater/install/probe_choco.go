//go:build windows

package install

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
)

const (
	chocoPackage = "dot"
)

// ChocoProbe detects Chocolatey installations on Windows.
type ChocoProbe struct {
	fs FileSystem
}

// NewChocoProbe creates a new Chocolatey probe.
func NewChocoProbe(fs FileSystem) *ChocoProbe {
	return &ChocoProbe{fs: fs}
}

// Name returns the probe identifier.
func (p *ChocoProbe) Name() string {
	return "chocolatey"
}

// Platforms returns the platforms this probe supports.
func (p *ChocoProbe) Platforms() []string {
	return []string{"windows"}
}

// Detect checks if dot was installed via Chocolatey.
func (p *ChocoProbe) Detect(ctx context.Context, execPath string) (*Info, error) {
	// Get Chocolatey lib path
	chocoPath := p.getChocoLibPath()
	if chocoPath == "" {
		return nil, nil
	}

	// Check if executable is within Chocolatey directory
	if !strings.HasPrefix(strings.ToLower(execPath), strings.ToLower(chocoPath)) {
		// Also check ProgramData\chocolatey\bin for shims
		binPath := filepath.Join(filepath.Dir(chocoPath), "bin")
		if !strings.HasPrefix(strings.ToLower(execPath), strings.ToLower(binPath)) {
			return nil, nil
		}
	}

	// Look for the package nuspec
	packageDir := filepath.Join(chocoPath, chocoPackage)
	entries, err := p.fs.ReadDir(packageDir)
	if err != nil {
		return nil, nil //nolint:nilerr // package not installed
	}

	// Find the nuspec file
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".nuspec") {
			nuspecPath := filepath.Join(packageDir, entry.Name())
			return p.parseNuspec(nuspecPath, execPath)
		}
	}

	return nil, nil
}

// getChocoLibPath returns the Chocolatey lib directory path.
func (p *ChocoProbe) getChocoLibPath() string {
	// Check ChocolateyInstall environment variable
	if chocoInstall := os.Getenv("ChocolateyInstall"); chocoInstall != "" {
		return filepath.Join(chocoInstall, "lib")
	}

	// Default path
	programData := os.Getenv("ProgramData")
	if programData == "" {
		programData = `C:\ProgramData`
	}
	return filepath.Join(programData, "chocolatey", "lib")
}

// nuspecPackage represents the structure of a Chocolatey .nuspec file.
type nuspecPackage struct {
	XMLName  xml.Name `xml:"package"`
	Metadata struct {
		ID          string `xml:"id"`
		Version     string `xml:"version"`
		Title       string `xml:"title"`
		Authors     string `xml:"authors"`
		Description string `xml:"description"`
	} `xml:"metadata"`
}

// parseNuspec parses a Chocolatey .nuspec file.
func (p *ChocoProbe) parseNuspec(path, execPath string) (*Info, error) {
	data, err := p.fs.ReadFile(path)
	if err != nil {
		return nil, nil //nolint:nilerr // cannot read nuspec
	}

	var pkg nuspecPackage
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return nil, nil //nolint:nilerr // invalid nuspec
	}

	if !strings.EqualFold(pkg.Metadata.ID, chocoPackage) {
		return nil, nil
	}

	metadata := map[string]string{
		"package":     pkg.Metadata.ID,
		"title":       pkg.Metadata.Title,
		"authors":     pkg.Metadata.Authors,
		"description": pkg.Metadata.Description,
	}

	return &Info{
		Source:              SourceChocolatey,
		Version:             pkg.Metadata.Version,
		ExecutablePath:      execPath,
		Metadata:            metadata,
		CanAutoUpgrade:      true,
		UpgradeInstructions: "choco upgrade dot",
	}, nil
}

// Ensure ChocoProbe implements Probe.
var _ Probe = (*ChocoProbe)(nil)
