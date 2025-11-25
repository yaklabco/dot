package install

// PacmanUpgrader handles Pacman upgrades.
type PacmanUpgrader struct {
	*baseUpgrader
}

// NewPacmanUpgrader creates a new Pacman upgrader.
func NewPacmanUpgrader(executor *Executor, detector Detector) *PacmanUpgrader {
	return &PacmanUpgrader{
		baseUpgrader: &baseUpgrader{
			source:         SourcePacman,
			executor:       executor,
			detector:       detector,
			metadataKey:    "name",
			defaultPkgName: "dot",
		},
	}
}

// Ensure PacmanUpgrader implements Upgrader.
var _ Upgrader = (*PacmanUpgrader)(nil)
