package install

import (
	"context"
	"runtime/debug"
	"strings"
)

const (
	// goModulePath is the expected module path for go install detection.
	goModulePath = "github.com/yaklabco/dot"
)

// GoInstallProbe detects go install installations.
type GoInstallProbe struct {
	// buildInfo allows injecting a mock for testing.
	buildInfo func() (*debug.BuildInfo, bool)
}

// NewGoInstallProbe creates a new go install probe.
func NewGoInstallProbe() *GoInstallProbe {
	return &GoInstallProbe{
		buildInfo: debug.ReadBuildInfo,
	}
}

// Name returns the probe identifier.
func (p *GoInstallProbe) Name() string {
	return "go-install"
}

// Platforms returns the platforms this probe supports.
func (p *GoInstallProbe) Platforms() []string {
	return nil // All platforms
}

// Detect checks if dot was installed via go install.
func (p *GoInstallProbe) Detect(ctx context.Context, execPath string) (*Info, error) {
	bi, ok := p.buildInfo()
	if !ok {
		return nil, nil
	}

	// Check if this is our module
	if bi.Path != goModulePath && !strings.HasPrefix(bi.Path, goModulePath+"/") {
		return nil, nil
	}

	metadata := map[string]string{
		"module":    bi.Path,
		"goVersion": bi.GoVersion,
	}

	// Extract build settings
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

	// Extract version from Main module
	version := bi.Main.Version
	if version == "(devel)" {
		// This is likely a source build, not go install
		return nil, nil
	}

	return &Info{
		Source:              SourceGoInstall,
		Version:             version,
		ExecutablePath:      execPath,
		Metadata:            metadata,
		CanAutoUpgrade:      true,
		UpgradeInstructions: "go install " + goModulePath + "/cmd/dot@latest",
	}, nil
}

// Ensure GoInstallProbe implements Probe.
var _ Probe = (*GoInstallProbe)(nil)
