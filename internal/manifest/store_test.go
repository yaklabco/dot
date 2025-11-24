package manifest

import (
	"context"
	"testing"

	"github.com/yaklabco/dot/internal/domain"
)

func TestManifestStore_Interface(t *testing.T) {
	// Verify interface is implemented by mock
	var _ ManifestStore = (*mockManifestStore)(nil)
}

type mockManifestStore struct {
	loadFn func(context.Context, domain.TargetPath) domain.Result[Manifest]
	saveFn func(context.Context, domain.TargetPath, Manifest) error
}

func (m *mockManifestStore) Load(ctx context.Context, target domain.TargetPath) domain.Result[Manifest] {
	return m.loadFn(ctx, target)
}

func (m *mockManifestStore) Save(ctx context.Context, target domain.TargetPath, manifest Manifest) error {
	return m.saveFn(ctx, target, manifest)
}
