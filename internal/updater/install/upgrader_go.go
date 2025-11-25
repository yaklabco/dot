package install

// GoInstallUpgrader handles go install upgrades.
type GoInstallUpgrader struct {
	*baseUpgrader
}

// NewGoInstallUpgrader creates a new go install upgrader.
func NewGoInstallUpgrader(executor *Executor, detector Detector) *GoInstallUpgrader {
	return &GoInstallUpgrader{
		baseUpgrader: &baseUpgrader{
			source:   SourceGoInstall,
			executor: executor,
			detector: detector,
			buildModuleSpec: func(info *Info) string {
				// Build the module path with @latest
				modulePath := info.Metadata["module"]
				if modulePath == "" {
					modulePath = goModulePath
				}
				return modulePath + "/cmd/dot@latest"
			},
		},
	}
}

// Ensure GoInstallUpgrader implements Upgrader.
var _ Upgrader = (*GoInstallUpgrader)(nil)
