package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_Exists(t *testing.T) {
	// This test verifies that main function exists and can be referenced.
	// Actual CLI testing happens through command tests.
	require.NotNil(t, main)
}

func TestSetupSignalHandler(t *testing.T) {
	t.Run("context is created", func(t *testing.T) {
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

	t.Run("context cancels on SIGINT", func(t *testing.T) {
		ctx := setupSignalHandler()

		// Send SIGINT to self
		err := syscall.Kill(os.Getpid(), syscall.SIGINT)
		require.NoError(t, err)

		// Wait for context to be canceled
		select {
		case <-ctx.Done():
			// Expected: context was canceled
			assert.Equal(t, context.Canceled, ctx.Err())
		case <-time.After(500 * time.Millisecond):
			t.Fatal("context should have been canceled")
		}
	})

	t.Run("context cancels on SIGTERM", func(t *testing.T) {
		ctx := setupSignalHandler()

		// Send SIGTERM to self
		err := syscall.Kill(os.Getpid(), syscall.SIGTERM)
		require.NoError(t, err)

		// Wait for context to be canceled
		select {
		case <-ctx.Done():
			// Expected: context was canceled
			assert.Equal(t, context.Canceled, ctx.Err())
		case <-time.After(500 * time.Millisecond):
			t.Fatal("context should have been canceled")
		}
	})
}

func TestSetupProfiling(t *testing.T) {
	t.Run("returns no-op cleanup when no profiling flags set", func(t *testing.T) {
		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg = globalConfig{} // All profiling flags empty

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)
		cleanup() // Should not panic
	})

	t.Run("handles invalid CPU profile path gracefully", func(t *testing.T) {
		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg = globalConfig{
			cpuProfile: "/invalid/path/that/does/not/exist/cpu.prof",
		}

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)
		cleanup() // Should not panic
	})

	t.Run("handles invalid memory profile path gracefully", func(t *testing.T) {
		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg = globalConfig{
			memProfile: "/invalid/path/that/does/not/exist/mem.prof",
		}

		cleanup := setupProfiling()
		require.NotNil(t, cleanup)
		cleanup() // Should not panic even with invalid path
	})

	t.Run("creates CPU profile file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cpuFile := tmpDir + "/cpu.prof"

		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg = globalConfig{
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

		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg = globalConfig{
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
