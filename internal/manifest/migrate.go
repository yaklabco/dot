package manifest

// PopulateMissingMetadata fills in TargetDir and PackageDir for packages
// that were created before these fields existed. This ensures backward
// compatibility with older manifest files.
func PopulateMissingMetadata(m *Manifest, defaultTargetDir, defaultPackageDir string) {
	for name, pkg := range m.Packages {
		updated := false

		if pkg.TargetDir == "" {
			pkg.TargetDir = defaultTargetDir
			updated = true
		}

		if pkg.PackageDir == "" {
			pkg.PackageDir = defaultPackageDir
			updated = true
		}

		if updated {
			m.Packages[name] = pkg
		}
	}
}
