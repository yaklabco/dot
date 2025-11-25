//go:build !windows

package install

import (
	"context"
)

// ChocoProbe is a stub for non-Windows platforms.
type ChocoProbe struct{}

// NewChocoProbe creates a stub Chocolatey probe for non-Windows platforms.
func NewChocoProbe(_ FileSystem) *ChocoProbe {
	return &ChocoProbe{}
}

// Name returns the probe identifier.
func (p *ChocoProbe) Name() string {
	return "chocolatey"
}

// Platforms returns the platforms this probe supports.
func (p *ChocoProbe) Platforms() []string {
	return []string{"windows"}
}

// Detect always returns nil on non-Windows platforms.
func (p *ChocoProbe) Detect(_ context.Context, _ string) (*Info, error) {
	return nil, nil
}

// Ensure ChocoProbe implements Probe.
var _ Probe = (*ChocoProbe)(nil)
