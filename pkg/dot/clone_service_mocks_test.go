package dot

import (
	"context"

	"github.com/yaklabco/dot/internal/adapters"
)

// mockGitCloner is a test double for GitCloner.
type mockGitCloner struct {
	cloneFn func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error
}

func (m *mockGitCloner) Clone(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
	if m.cloneFn != nil {
		return m.cloneFn(ctx, url, dest, opts)
	}
	return nil
}

// mockPackageSelector is a test double for PackageSelector.
type mockPackageSelector struct {
	selectFn func(ctx context.Context, packages []string) ([]string, error)
}

func (m *mockPackageSelector) Select(ctx context.Context, packages []string) ([]string, error) {
	if m.selectFn != nil {
		return m.selectFn(ctx, packages)
	}
	return packages, nil
}
