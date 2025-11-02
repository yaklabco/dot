package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/ignore"
)

// setupBenchmarkPackage creates a temporary package directory with the specified number of files.
func setupBenchmarkPackage(b *testing.B, fileCount int) string {
	b.Helper()

	tmpDir, err := os.MkdirTemp("", "dot-benchmark-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}

	// Create nested directory structure
	for i := 0; i < fileCount; i++ {
		// Create some nested directories
		subdir := filepath.Join(tmpDir, fmt.Sprintf("dir%d", i%10))
		if err := os.MkdirAll(subdir, 0755); err != nil {
			b.Fatalf("failed to create subdir: %v", err)
		}

		// Create files
		filePath := filepath.Join(subdir, fmt.Sprintf("file%d.txt", i))
		content := []byte(fmt.Sprintf("content for file %d\n", i))
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}
	}

	return tmpDir
}

// BenchmarkScanPackage benchmarks package scanning.
func BenchmarkScanPackage(b *testing.B) {
	tmpDir := setupBenchmarkPackage(b, 100)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	fs := adapters.NewOSFilesystem()
	pkgPath := domain.NewPackagePath(tmpDir).Unwrap()
	ignoreSet := ignore.NewDefaultIgnoreSet()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ScanPackage(ctx, fs, pkgPath, "test-package", ignoreSet)
	}
}

// BenchmarkScanPackage_SmallPackage benchmarks small package (10 files).
func BenchmarkScanPackage_SmallPackage(b *testing.B) {
	tmpDir := setupBenchmarkPackage(b, 10)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	fs := adapters.NewOSFilesystem()
	pkgPath := domain.NewPackagePath(tmpDir).Unwrap()
	ignoreSet := ignore.NewDefaultIgnoreSet()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ScanPackage(ctx, fs, pkgPath, "test-package", ignoreSet)
	}
}

// BenchmarkScanPackage_LargePackage benchmarks large package (1000 files).
func BenchmarkScanPackage_LargePackage(b *testing.B) {
	tmpDir := setupBenchmarkPackage(b, 1000)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	fs := adapters.NewOSFilesystem()
	pkgPath := domain.NewPackagePath(tmpDir).Unwrap()
	ignoreSet := ignore.NewDefaultIgnoreSet()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ScanPackage(ctx, fs, pkgPath, "test-package", ignoreSet)
	}
}

// BenchmarkScanPackage_WithIgnorePatterns benchmarks scanning with ignore patterns.
func BenchmarkScanPackage_WithIgnorePatterns(b *testing.B) {
	tmpDir := setupBenchmarkPackage(b, 100)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	fs := adapters.NewOSFilesystem()
	pkgPath := domain.NewPackagePath(tmpDir).Unwrap()

	// Create ignore set with common patterns
	ignoreSet := ignore.NewIgnoreSet()
	patterns := []string{"*.log", "*.tmp", ".git", "node_modules", "__pycache__"}
	for _, pattern := range patterns {
		if p := ignore.NewPattern(pattern); p.IsOk() {
			ignoreSet.AddPattern(p.Unwrap())
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ScanPackage(ctx, fs, pkgPath, "test-package", ignoreSet)
	}
}

// BenchmarkScanTree benchmarks tree scanning.
func BenchmarkScanTree(b *testing.B) {
	tmpDir := setupBenchmarkPackage(b, 100)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	fs := adapters.NewOSFilesystem()
	treePath := domain.NewFilePath(tmpDir).Unwrap()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ScanTree(ctx, fs, treePath)
	}
}

// BenchmarkScanTree_DeepNesting benchmarks tree scanning with deep nesting.
func BenchmarkScanTree_DeepNesting(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "dot-benchmark-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create deeply nested structure (10 levels deep)
	currentPath := tmpDir
	for i := 0; i < 10; i++ {
		currentPath = filepath.Join(currentPath, fmt.Sprintf("level%d", i))
		if err := os.MkdirAll(currentPath, 0755); err != nil {
			b.Fatalf("failed to create nested dir: %v", err)
		}

		// Add a file at each level
		filePath := filepath.Join(currentPath, "file.txt")
		if err := os.WriteFile(filePath, []byte("content\n"), 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}
	}

	ctx := context.Background()
	fs := adapters.NewOSFilesystem()
	treePath := domain.NewFilePath(tmpDir).Unwrap()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ScanTree(ctx, fs, treePath)
	}
}

// BenchmarkTranslatePackageName benchmarks package name translation.
func BenchmarkTranslatePackageName(b *testing.B) {
	testCases := []string{
		"dot-bashrc",
		"dot-vim",
		"dot-config-fish-config-fish",
		"ssh-dot-config",
		"file",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, name := range testCases {
			_ = TranslatePackageName(name)
		}
	}
}
