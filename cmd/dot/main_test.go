package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up test environment to prevent writing to source tree.
func TestMain(m *testing.M) {
	// Save original environment
	oldXDGData := os.Getenv("XDG_DATA_HOME")
	oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	oldXDGState := os.Getenv("XDG_STATE_HOME")

	// Create temp directory for XDG paths
	tmpDir, err := os.MkdirTemp("", "dot-test-*")
	if err != nil {
		panic("failed to create temp dir for tests: " + err.Error())
	}

	// Set XDG variables to temp directories to prevent writing to source tree
	os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	os.Setenv("XDG_STATE_HOME", filepath.Join(tmpDir, "state"))

	// Run tests
	exitCode := m.Run()

	// Cleanup temp directory
	os.RemoveAll(tmpDir)

	// Restore original environment
	if oldXDGData != "" {
		os.Setenv("XDG_DATA_HOME", oldXDGData)
	} else {
		os.Unsetenv("XDG_DATA_HOME")
	}
	if oldXDGConfig != "" {
		os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)
	} else {
		os.Unsetenv("XDG_CONFIG_HOME")
	}
	if oldXDGState != "" {
		os.Setenv("XDG_STATE_HOME", oldXDGState)
	} else {
		os.Unsetenv("XDG_STATE_HOME")
	}

	os.Exit(exitCode)
}

func TestMain_Exists(t *testing.T) {
	// This test verifies that main function exists and can be referenced.
	// Actual CLI testing happens through command tests.
	require.NotNil(t, main)
}

func TestSetupSignalHandler(t *testing.T) {
	t.Run("context is created and not initially canceled", func(t *testing.T) {
		ctx := setupSignalHandler()
		require.NotNil(t, ctx)

		// Verify context is not yet canceled
		select {
		case <-ctx.Done():
			t.Fatal("context should not be canceled initially")
		default:
			// Expected: context is not canceled
		}
	})

	// Note: Actual signal handling behavior (SIGINT/SIGTERM) is tested in
	// tests/integration/signal_test.go using subprocess isolation.
	// Testing signals in unit tests by sending them to the test process itself
	// is unsafe because:
	// - Signal handlers are process-global, not goroutine-local
	// - Multiple concurrent tests could interfere with each other
	// - The test runner itself could be affected by these signals
	// - May cause flaky test behavior in CI or when running with -p flag
}

func TestSetupProfiling(t *testing.T) {
	t.Run("returns no-op cleanup when no profiling flags set", func(t *testing.T) {
		previous := cliFlags
		t.Cleanup(func() {
			cliFlags = previous
		})
		cliFlags = CLIFlags{} // All profiling flags empty

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)
		cleanup() // Should not panic
	})

	t.Run("handles invalid CPU profile path gracefully", func(t *testing.T) {
		previous := cliFlags
		t.Cleanup(func() {
			cliFlags = previous
		})
		cliFlags = CLIFlags{
			cpuProfile: "/invalid/path/that/does/not/exist/cpu.prof",
		}

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)
		cleanup() // Should not panic
	})

	t.Run("handles invalid memory profile path gracefully", func(t *testing.T) {
		previous := cliFlags
		t.Cleanup(func() {
			cliFlags = previous
		})
		cliFlags = CLIFlags{
			memProfile: "/invalid/path/that/does/not/exist/mem.prof",
		}

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)
		cleanup() // Should not panic even with invalid path
	})

	t.Run("creates CPU profile file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cpuFile := tmpDir + "/cpu.prof"

		previous := cliFlags
		t.Cleanup(func() {
			cliFlags = previous
		})
		cliFlags = CLIFlags{
			cpuProfile: cpuFile,
		}

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)

		// CPU profiling should be active
		time.Sleep(10 * time.Millisecond) // Let some profiling data collect

		cleanup() // Stop profiling and write file

		// Verify file was created
		_, err := os.Stat(cpuFile)
		assert.NoError(t, err, "CPU profile file should exist")
	})

	t.Run("creates memory profile file", func(t *testing.T) {
		tmpDir := t.TempDir()
		memFile := tmpDir + "/mem.prof"

		previous := cliFlags
		t.Cleanup(func() {
			cliFlags = previous
		})
		cliFlags = CLIFlags{
			memProfile: memFile,
		}

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)

		cleanup() // Write memory profile

		// Verify file was created
		_, err := os.Stat(memFile)
		assert.NoError(t, err, "memory profile file should exist")
	})
}
