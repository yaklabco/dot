package main

import (
	"testing"
)

// TestCommonPrefix removed - function removed as part of adopt simplification
// TestDeriveCommonPackageName removed - function removed as part of adopt simplification
// TestFileExists removed - function removed as part of adopt simplification

// Placeholder test to keep package valid
func TestAdoptHelperFunctionsRemoved(t *testing.T) {
	// These helper functions were removed as part of the adopt command simplification.
	// Glob mode auto-detection has been removed in favor of explicit package names.
	// Users must now provide package names when adopting multiple files:
	//   dot adopt .ssh              # Auto-naming for single file
	//   dot adopt vim .vimrc .vim   # Explicit package for multiple files
	t.Skip("Helper function tests removed - functions no longer exist")
}
