package manifest

import (
	"context"

	"github.com/yaklabco/dot/internal/domain"
)

// ManifestStore provides persistence for manifests
type ManifestStore interface {
	// Load retrieves manifest from target directory
	// Returns empty manifest if file doesn't exist
	Load(ctx context.Context, targetDir domain.TargetPath) domain.Result[Manifest]

	// Save persists manifest to target directory
	// Write is atomic via temp file and rename
	Save(ctx context.Context, targetDir domain.TargetPath, manifest Manifest) error
}
