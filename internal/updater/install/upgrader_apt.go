package install

// AptUpgrader handles APT/dpkg upgrades.
type AptUpgrader struct {
	*baseUpgrader
}

// NewAptUpgrader creates a new APT upgrader.
func NewAptUpgrader(executor *Executor, detector Detector) *AptUpgrader {
	return &AptUpgrader{
		baseUpgrader: &baseUpgrader{
			source:         SourceApt,
			executor:       executor,
			detector:       detector,
			metadataKey:    "package",
			defaultPkgName: "dot",
		},
	}
}

// Ensure AptUpgrader implements Upgrader.
var _ Upgrader = (*AptUpgrader)(nil)
