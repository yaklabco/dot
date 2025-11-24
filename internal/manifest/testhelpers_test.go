package manifest

import (
	"testing"

	"github.com/yaklabco/dot/internal/domain"
)

func mustTargetPath(t *testing.T, path string) domain.TargetPath {
	t.Helper()
	result := domain.NewTargetPath(path)
	if result.IsErr() {
		t.Fatalf("failed to create target path: %v", result.UnwrapErr())
	}
	return result.Unwrap()
}

func mustPackagePath(t *testing.T, path string) domain.PackagePath {
	t.Helper()
	result := domain.NewPackagePath(path)
	if result.IsErr() {
		t.Fatalf("failed to create package path: %v", result.UnwrapErr())
	}
	return result.Unwrap()
}
