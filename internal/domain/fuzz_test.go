package domain

import (
	"testing"
)

// FuzzNewPackagePath tests package path validation with random input.
// Run with: go test -fuzz=FuzzNewPackagePath -fuzztime=30s
func FuzzNewPackagePath(f *testing.F) {
	// Seed corpus with valid paths
	f.Add("/home/user/.dotfiles")
	f.Add("/tmp/test")
	f.Add("/var/lib/packages")

	// Seed with potentially problematic input
	f.Add("")
	f.Add("relative/path")
	f.Add("../../../etc/passwd")
	f.Add("/home/user/\x00null")
	f.Add("//double//slash")
	f.Add("/path/with/./dot")
	f.Add("/path/with/../parent")
	f.Add(string(make([]byte, 10000)))
	f.Add("\x00\x01\x02\x03")

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic on any input
		_ = NewPackagePath(path)
	})
}

// FuzzNewTargetPath tests target path validation with random input.
func FuzzNewTargetPath(f *testing.F) {
	// Seed corpus with valid paths
	f.Add("/home/user")
	f.Add("/tmp")
	f.Add("/var/lib")

	// Seed with potentially problematic input
	f.Add("")
	f.Add("relative/path")
	f.Add("../../../etc/passwd")
	f.Add("/home/user/\x00null")
	f.Add("//double//slash")
	f.Add("/path/with/./dot")
	f.Add("/path/with/../parent")
	f.Add(string(make([]byte, 10000)))
	f.Add("\x00\x01\x02\x03")

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic on any input
		_ = NewTargetPath(path)
	})
}

// FuzzNewFilePath tests file path validation with random input.
func FuzzNewFilePath(f *testing.F) {
	// Seed corpus with valid paths
	f.Add("/home/user/.vimrc")
	f.Add("/tmp/test.txt")
	f.Add("/var/lib/config.yaml")

	// Seed with potentially problematic input
	f.Add("")
	f.Add("relative/file.txt")
	f.Add("../../../etc/passwd")
	f.Add("/home/user/\x00null.txt")
	f.Add("//double//slash//file")
	f.Add("/path/with/./dot/file")
	f.Add("/path/with/../parent/file")
	f.Add(string(make([]byte, 10000)))
	f.Add("\x00\x01\x02\x03")

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic on any input
		_ = NewFilePath(path)
	})
}

// FuzzPathJoin tests path joining with random input.
func FuzzPathJoin(f *testing.F) {
	// Seed corpus with valid combinations
	f.Add("/home/user", "file.txt")
	f.Add("/tmp", "test")
	f.Add("/var/lib", "config/settings.yaml")

	// Seed with potentially problematic input
	f.Add("", "")
	f.Add("/path", "")
	f.Add("", "file")
	f.Add("/path", "\x00null")
	f.Add("/path", "../escape")
	f.Add("/path", "../../etc/passwd")
	f.Add(string(make([]byte, 1000)), string(make([]byte, 1000)))

	f.Fuzz(func(t *testing.T, base, elem string) {
		// Should not panic on any input
		result := NewFilePath(base)
		if result.IsOk() {
			path := result.Unwrap()
			_ = path.Join(elem)
		}
	})
}

// FuzzAbsolutePathValidator tests absolute path validator with random input.
func FuzzAbsolutePathValidator(f *testing.F) {
	// Seed corpus
	f.Add("/home/user")
	f.Add("/tmp")
	f.Add("relative/path")
	f.Add("")
	f.Add("\x00")
	f.Add(string(make([]byte, 10000)))

	f.Fuzz(func(t *testing.T, path string) {
		validator := &AbsolutePathValidator{}
		// Should not panic on any input
		_ = validator.Validate(path)
	})
}

// FuzzRelativePathValidator tests relative path validator with random input.
func FuzzRelativePathValidator(f *testing.F) {
	// Seed corpus
	f.Add("relative/path")
	f.Add("../parent")
	f.Add("/absolute/path")
	f.Add("")
	f.Add("\x00")
	f.Add(string(make([]byte, 10000)))

	f.Fuzz(func(t *testing.T, path string) {
		validator := &RelativePathValidator{}
		// Should not panic on any input
		_ = validator.Validate(path)
	})
}

// FuzzNonEmptyPathValidator tests non-empty path validator with random input.
func FuzzNonEmptyPathValidator(f *testing.F) {
	// Seed corpus
	f.Add("/path")
	f.Add("")
	f.Add(" ")
	f.Add("\x00")
	f.Add(string(make([]byte, 10000)))

	f.Fuzz(func(t *testing.T, path string) {
		validator := &NonEmptyPathValidator{}
		// Should not panic on any input
		_ = validator.Validate(path)
	})
}
