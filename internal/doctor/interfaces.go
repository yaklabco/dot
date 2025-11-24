package doctor

import (
	"context"
	"io/fs"
	"os"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/manifest"
)

// FS defines the filesystem abstraction interface needed by doctor checks.
type FS interface {
	Exists(ctx context.Context, path string) (bool, error)
	IsDir(ctx context.Context, path string) (bool, error)
	Lstat(ctx context.Context, name string) (fs.FileInfo, error)
	ReadDir(ctx context.Context, name string) ([]fs.DirEntry, error)
	ReadFile(ctx context.Context, name string) ([]byte, error)
	ReadLink(ctx context.Context, name string) (string, error)
	WriteFile(ctx context.Context, name string, data []byte, perm os.FileMode) error
	Remove(ctx context.Context, name string) error
	MkdirAll(ctx context.Context, path string, perm os.FileMode) error
	Stat(ctx context.Context, name string) (fs.FileInfo, error)
}

// ManifestLoader defines the interface for loading manifests.
type ManifestLoader interface {
	Load(ctx context.Context, targetPath domain.TargetPath) domain.Result[manifest.Manifest]
	LoadManifest(ctx context.Context) (*manifest.Manifest, error)
}

// LinkHealthChecker defines the interface for checking link health.
type LinkHealthChecker interface {
	CheckLink(ctx context.Context, pkgName, linkPath, packageDir string) LinkHealthResult
}

// LinkHealthResult contains detailed health information for a single link.
type LinkHealthResult struct {
	IsHealthy  bool
	IssueType  IssueType
	Severity   domain.IssueSeverity
	Message    string
	Suggestion string
}

// TargetPathCreator defines the interface for creating TargetPath.
type TargetPathCreator interface {
	NewTargetPath(path string) domain.Result[domain.TargetPath]
}

// ManifestNotFoundChecker checks if an error is a manifest not found error.
type ManifestNotFoundChecker func(err error) bool

// ScanMode defines the depth of scanning for orphaned links.
type ScanMode int

const (
	// ScanOff disables orphan scanning.
	ScanOff ScanMode = iota
	// ScanScoped scans directories containing managed links (default).
	ScanScoped
	// ScanDeep performs full recursive scan with depth limits.
	ScanDeep
)

// ScanConfig configures the behavior of diagnostic scans.
type ScanConfig struct {
	Mode         ScanMode
	MaxDepth     int
	MaxWorkers   int
	MaxIssues    int
	ScopeToDirs  []string
	SkipPatterns []string
}

// IssueType defines the type of issue.
type IssueType string

const (
	// IssueBrokenLink indicates a symlink pointing to a non-existent target.
	IssueBrokenLink IssueType = "broken_link"
	// IssueOrphanedLink indicates a symlink not managed by any package.
	IssueOrphanedLink IssueType = "orphaned_link"
	// IssueWrongTarget indicates a symlink pointing to an unexpected target.
	IssueWrongTarget IssueType = "wrong_target"
)

// DiagnosticStats contains summary statistics.
type DiagnosticStats struct {
	TotalLinks    int
	BrokenLinks   int
	OrphanedLinks int
	ManagedLinks  int
}
