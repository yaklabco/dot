package install

import (
	"bufio"
	"bytes"
	"context"
	"path/filepath"
	"strings"
)

const (
	pacmanLocalDB  = "/var/lib/pacman/local"
	pacmanPackage  = "dot"
	pacmanDescFile = "desc"
)

// PacmanProbe detects Pacman installations on Arch-based systems.
type PacmanProbe struct {
	fs FileSystem
}

// NewPacmanProbe creates a new Pacman probe.
func NewPacmanProbe(fs FileSystem) *PacmanProbe {
	return &PacmanProbe{fs: fs}
}

// Name returns the probe identifier.
func (p *PacmanProbe) Name() string {
	return "pacman"
}

// Platforms returns the platforms this probe supports.
func (p *PacmanProbe) Platforms() []string {
	return []string{"linux"}
}

// Detect checks if dot was installed via Pacman.
func (p *PacmanProbe) Detect(ctx context.Context, execPath string) (*Info, error) {
	// List pacman local database
	entries, err := p.fs.ReadDir(pacmanLocalDB)
	if err != nil {
		return nil, nil //nolint:nilerr // pacman not available
	}

	// Look for dot package directory (format: dot-VERSION)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, pacmanPackage+"-") {
			continue
		}

		// Found potential match, read desc file
		descPath := filepath.Join(pacmanLocalDB, name, pacmanDescFile)
		data, err := p.fs.ReadFile(descPath)
		if err != nil {
			continue
		}

		info := p.parseDescFile(data)
		if info != nil && info.Metadata["name"] == pacmanPackage {
			info.ExecutablePath = execPath
			return info, nil
		}
	}

	return nil, nil
}

// parseDescFile parses a pacman desc file.
func (p *PacmanProbe) parseDescFile(data []byte) *Info {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	metadata := make(map[string]string)
	var currentSection string
	var version string

	for scanner.Scan() {
		line := scanner.Text()

		// Section headers are in %SECTION% format
		if strings.HasPrefix(line, "%") && strings.HasSuffix(line, "%") {
			currentSection = strings.Trim(line, "%")
			continue
		}

		// Empty line resets section
		if line == "" {
			currentSection = ""
			continue
		}

		// Read value based on current section
		switch currentSection {
		case "NAME":
			metadata["name"] = line
		case "VERSION":
			version = line
			metadata["version"] = line
		case "ARCH":
			metadata["arch"] = line
		case "URL":
			metadata["url"] = line
		case "DESC":
			metadata["description"] = line
		}
	}

	if metadata["name"] != pacmanPackage {
		return nil
	}

	return &Info{
		Source:              SourcePacman,
		Version:             version,
		Metadata:            metadata,
		CanAutoUpgrade:      true,
		UpgradeInstructions: "sudo pacman -Syu dot",
	}
}

// Ensure PacmanProbe implements Probe.
var _ Probe = (*PacmanProbe)(nil)
