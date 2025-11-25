//go:build windows

package install

// ChocoUpgrader handles Chocolatey upgrades.
type ChocoUpgrader struct {
	*baseUpgrader
}

// NewChocoUpgrader creates a new Chocolatey upgrader.
func NewChocoUpgrader(executor *Executor, detector Detector) *ChocoUpgrader {
	return &ChocoUpgrader{
		baseUpgrader: &baseUpgrader{
			source:         SourceChocolatey,
			executor:       executor,
			detector:       detector,
			metadataKey:    "package",
			defaultPkgName: "dot",
		},
	}
}

// Ensure ChocoUpgrader implements Upgrader.
var _ Upgrader = (*ChocoUpgrader)(nil)
