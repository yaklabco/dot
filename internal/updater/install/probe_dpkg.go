package install

import (
	"bufio"
	"bytes"
	"context"
	"strings"
)

const (
	dpkgStatusPath = "/var/lib/dpkg/status"
	dpkgPackage    = "dot"
)

// DpkgProbe detects APT/dpkg installations on Debian-based systems.
type DpkgProbe struct {
	fs FileSystem
}

// NewDpkgProbe creates a new dpkg probe.
func NewDpkgProbe(fs FileSystem) *DpkgProbe {
	return &DpkgProbe{fs: fs}
}

// Name returns the probe identifier.
func (p *DpkgProbe) Name() string {
	return "dpkg"
}

// Platforms returns the platforms this probe supports.
func (p *DpkgProbe) Platforms() []string {
	return []string{"linux"}
}

// Detect checks if dot was installed via APT/dpkg.
func (p *DpkgProbe) Detect(ctx context.Context, execPath string) (*Info, error) {
	// Check if dpkg database exists
	data, err := p.fs.ReadFile(dpkgStatusPath)
	if err != nil {
		return nil, nil //nolint:nilerr // dpkg not available
	}

	// Parse dpkg status file for the dot package
	info := p.parseDpkgStatus(data)
	if info == nil {
		return nil, nil
	}

	info.ExecutablePath = execPath
	return info, nil
}

// parseDpkgStatus parses /var/lib/dpkg/status to find the dot package.
func (p *DpkgProbe) parseDpkgStatus(data []byte) *Info {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var currentPackage string
	var version, arch, status string
	var inDotPackage bool

	for scanner.Scan() {
		line := scanner.Text()

		// Empty line marks end of package entry
		if line == "" {
			if inDotPackage && status == "install ok installed" {
				return p.buildInfo(version, arch)
			}
			currentPackage = ""
			version = ""
			arch = ""
			status = ""
			inDotPackage = false
			continue
		}

		// Parse key: value pairs
		if idx := strings.Index(line, ": "); idx > 0 {
			key := line[:idx]
			value := line[idx+2:]

			switch key {
			case "Package":
				currentPackage = value
				inDotPackage = (currentPackage == dpkgPackage)
			case "Version":
				if inDotPackage {
					version = value
				}
			case "Architecture":
				if inDotPackage {
					arch = value
				}
			case "Status":
				if inDotPackage {
					status = value
				}
			}
		}
	}

	// Check last package
	if inDotPackage && status == "install ok installed" {
		return p.buildInfo(version, arch)
	}

	return nil
}

// buildInfo constructs Info from parsed dpkg data.
func (p *DpkgProbe) buildInfo(version, arch string) *Info {
	metadata := map[string]string{
		"package":      dpkgPackage,
		"architecture": arch,
	}

	return &Info{
		Source:              SourceApt,
		Version:             version,
		Metadata:            metadata,
		CanAutoUpgrade:      true,
		UpgradeInstructions: "sudo apt-get update && sudo apt-get install --only-upgrade dot",
	}
}

// Ensure DpkgProbe implements Probe.
var _ Probe = (*DpkgProbe)(nil)
