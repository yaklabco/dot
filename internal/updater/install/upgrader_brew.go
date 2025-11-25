package install

// BrewUpgrader handles Homebrew upgrades.
type BrewUpgrader struct {
	*baseUpgrader
}

// NewBrewUpgrader creates a new Homebrew upgrader.
func NewBrewUpgrader(executor *Executor, detector Detector) *BrewUpgrader {
	return &BrewUpgrader{
		baseUpgrader: &baseUpgrader{
			source:         SourceHomebrew,
			executor:       executor,
			detector:       detector,
			metadataKey:    "formula",
			defaultPkgName: "dot",
			buildModuleSpec: func(info *Info) string {
				// Build the tap-qualified formula reference
				formula := info.Metadata["formula"]
				if formula == "" {
					formula = "dot"
				}
				tap := info.Metadata["tap"]
				if tap != "" {
					return tap + "/" + formula
				}
				return formula
			},
		},
	}
}

// Ensure BrewUpgrader implements Upgrader.
var _ Upgrader = (*BrewUpgrader)(nil)
