//go:build !windows

package install

import (
	"context"
	"fmt"
)

// ChocoUpgrader is a stub for non-Windows platforms.
type ChocoUpgrader struct{}

// NewChocoUpgrader creates a stub Chocolatey upgrader for non-Windows platforms.
func NewChocoUpgrader(_ *Executor, _ Detector) *ChocoUpgrader {
	return &ChocoUpgrader{}
}

// CanUpgrade returns false on non-Windows platforms.
func (u *ChocoUpgrader) CanUpgrade(info *Info) bool {
	return false
}

// Upgrade returns an error on non-Windows platforms.
func (u *ChocoUpgrader) Upgrade(_ context.Context, _ *Info) (*UpgradeResult, error) {
	return nil, fmt.Errorf("chocolatey upgrades not supported on this platform")
}

// VerifyUpgrade returns an error on non-Windows platforms.
func (u *ChocoUpgrader) VerifyUpgrade(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("chocolatey verification not supported on this platform")
}

// Ensure ChocoUpgrader implements Upgrader.
var _ Upgrader = (*ChocoUpgrader)(nil)
