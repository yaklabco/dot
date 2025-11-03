package dot

import "time"

// Status represents the installation state of packages.
type Status struct {
	Packages []PackageInfo `json:"packages" yaml:"packages"`
}

// PackageInfo contains metadata about an installed package.
type PackageInfo struct {
	Name        string    `json:"name" yaml:"name"`
	Source      string    `json:"source" yaml:"source"`
	InstalledAt time.Time `json:"installed_at" yaml:"installed_at"`
	LinkCount   int       `json:"link_count" yaml:"link_count"`
	Links       []string  `json:"links" yaml:"links"`
	TargetDir   string    `json:"target_dir,omitempty" yaml:"target_dir,omitempty"`
	PackageDir  string    `json:"package_dir,omitempty" yaml:"package_dir,omitempty"`
}
