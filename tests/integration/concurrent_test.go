package integration

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestConcurrent_ParallelPackageScanning tests concurrent package scanning.
func TestConcurrent_ParallelPackageScanning(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create multiple packages
	for i := 0; i < 10; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	// Manage all packages (should use parallel scanning internally)
	packages := make([]string, 10)
	for i := 0; i < 10; i++ {
		packages[i] = filepath.Join("pkg", string(rune('a'+i)))
	}

	err := client.Manage(env.Context(), packages...)
	require.NoError(t, err)

	// Verify all packages installed
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 10)
}

// TestConcurrent_MultipleManageOperations tests concurrent manage operations.
func TestConcurrent_MultipleManageOperations(t *testing.T) {
	env := testutil.NewTestEnvironment(t)

	// Create multiple packages
	for i := 0; i < 5; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file"+string(rune('a'+i)), "content").
			Create()
	}

	// Launch concurrent manage operations
	var wg sync.WaitGroup
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		go func(pkg string) {
			defer wg.Done()
			// Each goroutine gets its own client to test concurrent safety
			client := testutil.NewTestClient(t, env)
			if err := client.Manage(context.Background(), pkg); err != nil {
				errors <- err
			}
		}(pkgName)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent manage failed: %v", err)
	}
}

// TestConcurrent_StatusQueriesDuringManage tests concurrent status queries during operations.
func TestConcurrent_StatusQueriesDuringManage(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create packages
	for i := 0; i < 5; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	var wg sync.WaitGroup
	done := make(chan bool)

	// Start background status queries
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				_, _ = client.Status(context.Background())
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Perform manage operations
	packages := make([]string, 5)
	for i := 0; i < 5; i++ {
		packages[i] = filepath.Join("pkg", string(rune('a'+i)))
	}
	err := client.Manage(env.Context(), packages...)
	require.NoError(t, err)

	close(done)
	wg.Wait()
}

// TestConcurrent_ManifestAccess tests concurrent manifest access.
func TestConcurrent_ManifestAccess(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage a package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Launch concurrent status queries (read manifest)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.Status(context.Background())
			if err != nil {
				t.Errorf("concurrent status failed: %v", err)
			}
		}()
	}

	wg.Wait()
}

// TestConcurrent_ParallelExecutionBatches tests parallel operation execution.
func TestConcurrent_ParallelExecutionBatches(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create multiple packages with multiple files
	for i := 0; i < 5; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		pkg := env.FixtureBuilder().Package(pkgName)
		for j := 0; j < 3; j++ {
			pkg.WithFile("dot-file"+string(rune('a'+j)), "content")
		}
		pkg.Create()
	}

	// Manage all packages (operations should be parallelized)
	packages := make([]string, 5)
	for i := 0; i < 5; i++ {
		packages[i] = filepath.Join("pkg", string(rune('a'+i)))
	}

	start := time.Now()
	err := client.Manage(env.Context(), packages...)
	duration := time.Since(start)

	require.NoError(t, err)
	t.Logf("Parallel execution completed in %v", duration)
}

// TestConcurrent_CancellationDuringExecution tests context cancellation during operations.
func TestConcurrent_CancellationDuringExecution(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create packages
	for i := 0; i < 10; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	packages := make([]string, 10)
	for i := 0; i < 10; i++ {
		packages[i] = filepath.Join("pkg", string(rune('a'+i)))
	}

	// This may or may not error depending on timing, but should not panic
	_ = client.Manage(ctx, packages...)
}

// TestConcurrent_StressTest performs stress testing with many operations.
func TestConcurrent_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create 20 packages
	for i := 0; i < 20; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+(i%26))))
		if i >= 26 {
			pkgName += string(rune('0' + (i / 26)))
		}
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	packages := make([]string, 20)
	for i := 0; i < 20; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+(i%26))))
		if i >= 26 {
			pkgName += string(rune('0' + (i / 26)))
		}
		packages[i] = pkgName
	}

	// Manage all packages
	err := client.Manage(env.Context(), packages...)
	require.NoError(t, err)

	// Verify all installed
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 20)
}

// TestConcurrent_NoRaceConditions runs with race detector to catch races.
// Run with: go test -race
func TestConcurrent_NoRaceConditions(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create packages
	for i := 0; i < 5; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	var wg sync.WaitGroup

	// Concurrent manages
	for i := 0; i < 3; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			pkgName := filepath.Join("pkg", string(rune('a'+idx)))
			_ = client.Manage(context.Background(), pkgName)
		}()
	}

	// Concurrent status queries
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.Status(context.Background())
		}()
	}

	wg.Wait()
}

// TestConcurrent_OperationIsolation tests that concurrent operations are properly isolated.
func TestConcurrent_OperationIsolation(t *testing.T) {
	env := testutil.NewTestEnvironment(t)

	// Create packages with unique files
	for i := 0; i < 5; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file"+string(rune('a'+i)), "content"+string(rune('a'+i))).
			Create()
	}

	var wg sync.WaitGroup
	results := make(chan string, 5)

	// Launch concurrent operations on different packages
	for i := 0; i < 5; i++ {
		wg.Add(1)
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		go func(pkg string) {
			defer wg.Done()
			client := testutil.NewTestClient(t, env)
			err := client.Manage(context.Background(), pkg)
			if err == nil {
				results <- pkg
			}
		}(pkgName)
	}

	wg.Wait()
	close(results)

	// Verify all packages were managed
	count := 0
	for range results {
		count++
	}
	assert.Equal(t, 5, count)
}
