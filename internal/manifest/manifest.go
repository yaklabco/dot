package manifest

import "time"

// Manifest tracks installed package state
type Manifest struct {
	Version    string                 `json:"version"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Packages   map[string]PackageInfo `json:"packages"`
	Hashes     map[string]string      `json:"hashes"`
	Repository *RepositoryInfo        `json:"repository,omitempty"`
}

// PackageSource indicates how a package was installed
type PackageSource string

const (
	// SourceManaged indicates package was installed via manage command
	SourceManaged PackageSource = "managed"
	// SourceAdopted indicates package was created via adopt command
	SourceAdopted PackageSource = "adopted"
)

// PackageInfo contains installation metadata for a package
type PackageInfo struct {
	Name        string            `json:"name"`
	InstalledAt time.Time         `json:"installed_at"`
	LinkCount   int               `json:"link_count"`
	Links       []string          `json:"links"`
	Backups     map[string]string `json:"backups,omitempty"` // target path -> backup path
	Source      PackageSource     `json:"source,omitempty"`  // How package was installed (adopted vs managed)
}

// RepositoryInfo contains metadata about the cloned repository.
type RepositoryInfo struct {
	// URL is the git repository URL.
	URL string `json:"url"`

	// Branch is the cloned branch name.
	Branch string `json:"branch"`

	// ClonedAt is the timestamp when the repository was cloned.
	ClonedAt time.Time `json:"cloned_at"`

	// CommitSHA is the commit hash at clone time (optional).
	CommitSHA string `json:"commit_sha,omitempty"`
}

// New creates a new empty manifest
func New() Manifest {
	return Manifest{
		Version:   "1.0",
		UpdatedAt: time.Now(),
		Packages:  make(map[string]PackageInfo),
		Hashes:    make(map[string]string),
	}
}

// AddPackage adds or updates package information
func (m *Manifest) AddPackage(pkg PackageInfo) {
	m.Packages[pkg.Name] = pkg
	m.UpdatedAt = time.Now()
}

// RemovePackage removes package from manifest
func (m *Manifest) RemovePackage(name string) bool {
	if _, exists := m.Packages[name]; !exists {
		return false
	}
	delete(m.Packages, name)
	delete(m.Hashes, name)
	m.UpdatedAt = time.Now()
	return true
}

// GetPackage retrieves package information
func (m *Manifest) GetPackage(name string) (PackageInfo, bool) {
	pkg, exists := m.Packages[name]
	return pkg, exists
}

// SetHash updates content hash for package
func (m *Manifest) SetHash(name, hash string) {
	m.Hashes[name] = hash
	m.UpdatedAt = time.Now()
}

// GetHash retrieves content hash for package
func (m *Manifest) GetHash(name string) (string, bool) {
	hash, exists := m.Hashes[name]
	return hash, exists
}

// PackageList returns all packages as slice
func (m *Manifest) PackageList() []PackageInfo {
	packages := make([]PackageInfo, 0, len(m.Packages))
	for _, pkg := range m.Packages {
		packages = append(packages, pkg)
	}
	return packages
}

// SetRepository sets the repository information for the manifest.
func (m *Manifest) SetRepository(info RepositoryInfo) {
	m.Repository = &info
	m.UpdatedAt = time.Now()
}

// GetRepository retrieves the repository information.
// Returns the repository info and true if set, or empty info and false if not set.
func (m *Manifest) GetRepository() (RepositoryInfo, bool) {
	if m.Repository == nil {
		return RepositoryInfo{}, false
	}
	return *m.Repository, true
}

// ClearRepository removes the repository information from the manifest.
func (m *Manifest) ClearRepository() {
	m.Repository = nil
	m.UpdatedAt = time.Now()
}
