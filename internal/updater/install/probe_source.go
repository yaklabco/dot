package install

import (
	"context"
	"runtime/debug"
	"strings"
)

// SourceProbe detects development/source builds.
type SourceProbe struct {
	version string
	// buildInfo allows injecting a mock for testing.
	buildInfo func() (*debug.BuildInfo, bool)
}

// NewSourceProbe creates a new source build probe.
func NewSourceProbe(version string) *SourceProbe {
	return &SourceProbe{
		version:   version,
		buildInfo: debug.ReadBuildInfo,
	}
}

// Name returns the probe identifier.
func (p *SourceProbe) Name() string {
	return "source"
}

// Platforms returns the platforms this probe supports.
func (p *SourceProbe) Platforms() []string {
	return nil // All platforms
}

// Detect checks if dot was built from source.
func (p *SourceProbe) Detect(ctx context.Context, execPath string) (*Info, error) {
	// Check version string for development indicators
	if p.isDevVersion() {
		return p.buildInfo_(), nil
	}

	// Check runtime build info
	bi, ok := p.buildInfo()
	if !ok {
		return nil, nil
	}

	// "(devel)" version indicates local build
	if bi.Main.Version == "(devel)" {
		return p.buildInfo_(), nil
	}

	// Check for dirty VCS state
	for _, setting := range bi.Settings {
		if setting.Key == "vcs.modified" && setting.Value == "true" {
			return p.buildInfo_(), nil
		}
	}

	return nil, nil
}

// isDevVersion checks if the version string indicates a development build.
func (p *SourceProbe) isDevVersion() bool {
	// Common development version patterns
	devIndicators := []string{
		"dev",
		"devel",
		"local",
		"snapshot",
		"-dirty",
		"+dirty",
		"(devel)",
	}

	lower := strings.ToLower(p.version)
	for _, indicator := range devIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	return false
}

// buildInfo_ creates Info for a source build.
func (p *SourceProbe) buildInfo_() *Info {
	metadata := make(map[string]string)

	// Try to get VCS info
	if bi, ok := p.buildInfo(); ok {
		metadata["goVersion"] = bi.GoVersion
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs.revision":
				metadata["vcsRevision"] = setting.Value
			case "vcs.time":
				metadata["vcsTime"] = setting.Value
			case "vcs.modified":
				metadata["vcsModified"] = setting.Value
			}
		}
	}

	return &Info{
		Source:              SourceBuild,
		Version:             p.version,
		Metadata:            metadata,
		CanAutoUpgrade:      false,
		UpgradeInstructions: "git pull && make build",
	}
}

// Ensure SourceProbe implements Probe.
var _ Probe = (*SourceProbe)(nil)
