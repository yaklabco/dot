package config

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzLoadFromFile tests config loading with random file content.
// Run with: go test -fuzz=FuzzLoadFromFile -fuzztime=30s
func FuzzLoadFromFile(f *testing.F) {
	// Seed corpus with valid YAML configs
	f.Add([]byte("directories:\n  package: /home/user/.dotfiles\n  target: /home/user\n"))
	f.Add([]byte("logging:\n  level: INFO\n  format: text\n"))
	f.Add([]byte("symlinks:\n  mode: relative\n  folding: true\n"))
	f.Add([]byte("network:\n  http_proxy: http://proxy.example.com:8080\n  timeout: 10\n"))

	// Seed with potentially problematic input
	f.Add([]byte("invalid: !!binary |\n  R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs=\n"))
	f.Add([]byte("directories:\n  package: \"\x00null\"\n"))
	f.Add([]byte("directories:\n  package: /../../etc/passwd\n"))
	f.Add([]byte("---\ndirectories:\n  package: test\n...\n"))
	f.Add([]byte("directories:\n  package: \"very long string\"\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Write config to a temp file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return // Skip invalid filesystem operations
		}

		// Should not panic on any input
		_, _ = LoadExtendedFromFile(configPath)
	})
}

// FuzzValidateExtended tests extended config validation with random input.
func FuzzValidateExtended(f *testing.F) {
	// Seed with valid configs
	f.Add("/home/user/.dotfiles", "/home/user")
	f.Add("~/.dotfiles", "~")
	f.Add("/tmp/test", "/tmp/target")

	// Seed with potentially problematic input
	f.Add("", "")
	f.Add("/../../etc/passwd", "/tmp")
	f.Add("/tmp\x00null", "/tmp")
	f.Add(string(make([]byte, 10000)), "/tmp")

	f.Fuzz(func(t *testing.T, packageDir, targetDir string) {
		cfg := &ExtendedConfig{
			Directories: DirectoriesConfig{
				Package: packageDir,
				Target:  targetDir,
			},
		}

		// Should not panic on any input
		_ = cfg.Validate()
	})
}

// FuzzLoaderLoad tests loader with random config paths.
func FuzzLoaderLoad(f *testing.F) {
	// Seed corpus
	f.Add("~/.dotfiles/config.yaml")
	f.Add("/home/user/.config/dot/config.yaml")
	f.Add("./relative/path/config.yaml")
	f.Add("../parent/path/config.yaml")
	f.Add("~/../../etc/passwd")
	f.Add("/tmp/\x00null")
	f.Add(string(make([]byte, 1000)))

	f.Fuzz(func(t *testing.T, path string) {
		loader := NewLoader("dot", path)
		// Should not panic on any input
		_, _ = loader.Load()
	})
}
