// Package dot provides a modern, type-safe symlink manager for dotfiles.
//
// dot is a modern dotfile manager written in Go 1.25.4, following strict
// constitutional principles: test-driven development, atomic operations,
// functional programming, and comprehensive error handling.
//
// # Architecture
//
// The library uses an interface-based Client pattern to provide a clean
// public API while keeping internal implementation details hidden:
//
//   - Client interface in pkg/dot (stable public API)
//   - Implementation in internal/api (can evolve freely)
//   - Domain types in pkg/dot (shared between public and internal)
//
// This pattern avoids import cycles between pkg/dot (which contains domain
// types like Operation, Plan, Result) and internal packages (which depend on
// those domain types).
//
// # Basic Usage
//
// Create a client and manage packages:
//
//	import (
//		"context"
//		"log"
//
//		"github.com/yaklabco/dot/internal/adapters"
//		"github.com/yaklabco/dot/pkg/dot"
//	)
//
//	cfg := dot.Config{
//		PackageDir:   "/home/user/dotfiles",
//		TargetDir: "/home/user",
//		FS:        adapters.NewOSFilesystem(),
//		Logger:    adapters.NewNoopLogger(),
//	}
//
//	client, err := dot.NewClient(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	if err := client.Manage(ctx, "vim", "zsh", "git"); err != nil {
//		log.Fatal(err)
//	}
//
// # Dry Run Mode
//
// Preview operations without applying changes:
//
//	cfg.DryRun = true
//	client, _ := dot.NewClient(cfg)
//
//	// Shows what would be done without executing
//	if err := client.Manage(ctx, "vim"); err != nil {
//		log.Fatal(err)
//	}
//
// # Planning Operations
//
// Get execution plan without applying:
//
//	plan, err := client.PlanManage(ctx, "vim")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Would execute %d operations\n", len(plan.Operations))
//	for _, op := range plan.Operations {
//		fmt.Printf("  %s\n", op.Kind())
//	}
//
// # Query Operations
//
// Check installation status:
//
//	status, err := client.Status(ctx, "vim")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, pkg := range status.Packages {
//		fmt.Printf("%s: %d links\n", pkg.Name, pkg.LinkCount)
//	}
//
// List all installed packages:
//
//	packages, err := client.List(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, pkg := range packages {
//		fmt.Printf("%s (installed %s)\n", pkg.Name, pkg.InstalledAt)
//	}
//
// # Configuration
//
// The Config struct controls all dot behavior:
//
//   - PackageDir: Source directory containing packages (required, absolute path)
//   - TargetDir: Destination directory for symlinks (required, absolute path)
//   - FS: Filesystem implementation (required)
//   - Logger: Logger implementation (required)
//   - Tracer: Distributed tracing (optional, defaults to noop)
//   - Metrics: Metrics collection (optional, defaults to noop)
//   - LinkMode: Relative or absolute symlinks (default: relative)
//   - Folding: Enable directory folding (default: true)
//   - DryRun: Preview mode (default: false)
//   - Verbosity: Logging level (default: 0)
//   - BackupDir: Location for backup files (default: <TargetDir>/.dot-backup)
//   - Concurrency: Parallel operation limit (default: NumCPU)
//
// Configuration must be validated before use:
//
//	cfg := dot.Config{
//		PackageDir:   "/home/user/dotfiles",
//		TargetDir: "/home/user",
//		FS:        adapters.NewOSFilesystem(),
//		Logger:    adapters.NewNoopLogger(),
//	}
//
//	if err := cfg.Validate(); err != nil {
//		log.Fatal(err)
//	}
//
// # Observability
//
// The library provides first-class observability through injected ports:
//
//   - Structured logging via Logger interface (slog compatible)
//   - Distributed tracing via Tracer interface (OpenTelemetry compatible)
//   - Metrics collection via Metrics interface (Prometheus compatible)
//
// Example with full observability:
//
//	cfg := dot.Config{
//		PackageDir: "/home/user/dotfiles",
//		TargetDir: "/home/user",
//		FS:      adapters.NewOSFilesystem(),
//		Logger:  adapters.NewSlogLogger(slog.Default()),
//		Tracer:  otelTracer, // Your OpenTelemetry tracer
//		Metrics: promMetrics, // Your Prometheus metrics
//	}
//
// # Testing
//
// The library is designed for testability:
//
//   - All operations accept context.Context for cancellation
//   - Filesystem abstraction enables testing without disk I/O
//   - Pure functional core enables property-based testing
//   - Interface-based Client enables mocking
//
// Example test using in-memory filesystem:
//
//	func TestMyTool(t *testing.T) {
//		fs := adapters.NewMemFS()
//
//		cfg := dot.Config{
//			PackageDir:   "/test/packages",
//			TargetDir: "/test/target",
//			FS:        fs,
//			Logger:    adapters.NewNoopLogger(),
//		}
//
//		client, _ := dot.NewClient(cfg)
//		err := client.Manage(ctx, "vim")
//		require.NoError(t, err)
//	}
//
// # Error Handling
//
// All operations return explicit errors. Common error types:
//
//   - ErrInvalidPath: Path validation failed
//   - ErrPackageNotFound: Package doesn't exist in PackageDir
//   - ErrConflict: Conflict detected during operation
//   - ErrCyclicDependency: Circular dependency in operations
//   - ErrMultiple: Multiple errors occurred
//
// Errors include user-facing messages via UserFacingErrorMessage().
//
// # Safety Guarantees
//
// The library provides strong safety guarantees:
//
//   - Type safety: Phantom types prevent path mixing at compile time
//   - Transaction safety: Two-phase commit with automatic rollback on failure
//   - Conflict detection: All conflicts reported before modification
//   - Atomic operations: All-or-nothing semantics
//   - Thread safety: All Client operations safe for concurrent use
//
// # Implementation Status
//
// All core operations are fully implemented:
//   - Client interface with registration pattern
//   - Manage/PlanManage operations with dependency resolution
//   - Unmanage operations with restore, purge, and cleanup options
//   - Adopt operations with file and directory support
//   - Status/List query operations
//
// Future enhancements:
//   - Streaming API for large operations
//   - ConfigBuilder for fluent configuration
//   - Performance optimizations for large package sets
//
// For detailed examples, see examples_test.go.
// For architecture details, see docs/Architecture.md.
// For implementation roadmap, see docs/Phase-12-Plan.md.
package dot
