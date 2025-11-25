package install

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
)

// detector implements the Detector interface.
type detector struct {
	fs           FileSystem
	probes       []Probe
	execResolver func() (string, error)
	version      string
}

// Option configures the detector.
type Option func(*detector)

// WithFileSystem sets the filesystem implementation.
func WithFileSystem(fs FileSystem) Option {
	return func(d *detector) {
		d.fs = fs
	}
}

// WithProbes sets custom probes.
func WithProbes(probes ...Probe) Option {
	return func(d *detector) {
		d.probes = probes
	}
}

// WithExecResolver sets a custom executable resolver.
func WithExecResolver(resolver func() (string, error)) Option {
	return func(d *detector) {
		d.execResolver = resolver
	}
}

// WithVersion sets the current version for source detection.
func WithVersion(version string) Option {
	return func(d *detector) {
		d.version = version
	}
}

// NewDetector creates a detector with platform-appropriate probes.
func NewDetector(opts ...Option) Detector {
	d := &detector{
		fs:           OSFileSystem{},
		execResolver: resolveExecutable,
	}

	for _, opt := range opts {
		opt(d)
	}

	// Set default probes if none provided
	if len(d.probes) == 0 {
		d.probes = defaultProbes(d.fs, d.version)
	}

	return d
}

// defaultProbes returns platform-appropriate probes.
func defaultProbes(fs FileSystem, version string) []Probe {
	probes := []Probe{
		// Package managers (platform-specific, higher priority)
		NewBrewProbe(fs),
		NewDpkgProbe(fs),
		NewPacmanProbe(fs),
		NewChocoProbe(fs),
		// Go install (all platforms)
		NewGoInstallProbe(),
		// Source build detection (all platforms, lower priority)
		NewSourceProbe(version),
	}

	// Filter probes by platform
	goos := runtime.GOOS
	var filtered []Probe
	for _, probe := range probes {
		platforms := probe.Platforms()
		if len(platforms) == 0 {
			// Probe supports all platforms
			filtered = append(filtered, probe)
			continue
		}
		for _, p := range platforms {
			if p == goos {
				filtered = append(filtered, probe)
				break
			}
		}
	}

	return filtered
}

// Detect discovers how dot was installed.
func (d *detector) Detect(ctx context.Context) (*Info, error) {
	// Get the executable path
	execPath, err := d.execResolver()
	if err != nil {
		return d.manualFallback(""), err
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Try each probe in order
	for _, probe := range d.probes {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		info, err := probe.Detect(ctx, realPath)
		if err != nil {
			// Log but continue to next probe
			continue
		}
		if info != nil {
			return info, nil
		}
	}

	// No probe matched - return manual installation
	return d.manualFallback(realPath), nil
}

// manualFallback creates Info for an unknown installation source.
func (d *detector) manualFallback(execPath string) *Info {
	return &Info{
		Source:              SourceManual,
		Version:             d.version,
		ExecutablePath:      execPath,
		Metadata:            map[string]string{},
		CanAutoUpgrade:      false,
		UpgradeInstructions: "Download the latest release from GitHub releases page",
	}
}

// resolveExecutable returns the path to the currently running executable.
func resolveExecutable() (string, error) {
	return os.Executable()
}

// Ensure detector implements Detector.
var _ Detector = (*detector)(nil)
