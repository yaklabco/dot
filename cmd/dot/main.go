package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- pprof is intentionally exposed for diagnostics
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// Version information (set via ldflags at build time)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// cleanupGracePeriod is the time allowed for operations to complete
// gracefully after receiving the first shutdown signal (SIGINT/SIGTERM).
// This gives atomic operations time to finish their current transaction
// before a forced exit on the second signal.
const cleanupGracePeriod = 100 * time.Millisecond

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

func run() int {
	// Setup signal handling for graceful shutdown
	ctx := setupSignalHandler()

	// Setup profiling (CPU, memory, pprof server)
	cleanup := setupProfiling()
	defer cleanup()

	rootCmd := NewRootCommand(version, commit, date)

	// Execute command with signal-aware context
	executedCmd, err := executeCommand(ctx, rootCmd)
	if err != nil {
		// Show usage for argument validation errors
		// (Flag errors are handled by SetFlagErrorFunc in root.go)
		if executedCmd != nil && isArgValidationError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			_ = executedCmd.Usage()
		} else {
			// Print all other errors to stderr
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		// Handle doctor-specific exit codes
		return getDoctorExitCode(err)
	}

	return 0
}

// setupProfiling initializes CPU profiling, memory profiling, and pprof HTTP server based on flags.
// Returns a cleanup function that should be deferred.
func setupProfiling() func() {
	cleanupFuncs := []func(){}

	// CPU profiling
	if globalCfg.cpuProfile != "" {
		f, err := os.Create(globalCfg.cpuProfile)
		if err != nil {
			slog.Error("failed to create CPU profile", "error", err, "file", globalCfg.cpuProfile)
		} else {
			if err := pprof.StartCPUProfile(f); err != nil {
				slog.Error("failed to start CPU profile", "error", err)
				f.Close()
			} else {
				slog.Info("CPU profiling enabled", "file", globalCfg.cpuProfile)
				cleanupFuncs = append(cleanupFuncs, func() {
					pprof.StopCPUProfile()
					f.Close()
					slog.Info("CPU profile written", "file", globalCfg.cpuProfile)
				})
			}
		}
	}

	// pprof HTTP server
	if globalCfg.pprofAddr != "" {
		go func() {
			slog.Info("starting pprof server", "addr", globalCfg.pprofAddr)
			// The pprof handlers are automatically registered via the import
			// #nosec G114 -- pprof server is for diagnostics only, no timeout needed
			if err := http.ListenAndServe(globalCfg.pprofAddr, nil); err != nil {
				slog.Error("pprof server failed", "error", err)
			}
		}()
	}

	// Memory profile (written on exit)
	if globalCfg.memProfile != "" {
		cleanupFuncs = append(cleanupFuncs, func() {
			f, err := os.Create(globalCfg.memProfile)
			if err != nil {
				slog.Error("failed to create memory profile", "error", err, "file", globalCfg.memProfile)
				return
			}
			defer f.Close()

			runtime.GC() // Run GC before taking heap profile
			if err := pprof.WriteHeapProfile(f); err != nil {
				slog.Error("failed to write memory profile", "error", err)
			} else {
				slog.Info("memory profile written", "file", globalCfg.memProfile)
			}
		})
	}

	// Return cleanup function that runs all cleanup funcs
	return func() {
		for _, cleanup := range cleanupFuncs {
			cleanup()
		}
	}
}

// setupSignalHandler creates a context that is canceled on interrupt signals.
// It supports graceful shutdown on first signal and forced exit on second signal.
func setupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("received signal, initiating graceful shutdown", "signal", sig)
		cancel()

		// Give atomic operations time to complete their current transaction
		// before allowing forced exit on second signal
		time.Sleep(cleanupGracePeriod)

		// Force exit if second signal received
		sig = <-sigChan
		slog.Error("received second signal, forcing exit", "signal", sig)
		// Only force exit if not in test mode
		if !isTestMode() {
			os.Exit(130) // 128 + SIGINT(2)
		}
	}()

	return ctx
}

// isTestMode detects if we're running in test mode
func isTestMode() bool {
	// Check if any test flags are present or if we're being run by go test
	for _, arg := range os.Args {
		if arg == "-test.v" || arg == "-test.run" || strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

// executeCommand executes the root command with the given context and returns the executed command and any error.
func executeCommand(ctx context.Context, rootCmd *cobra.Command) (*cobra.Command, error) {
	var executedCmd *cobra.Command

	// Set context on root command
	rootCmd.SetContext(ctx)

	// Use PreRun hook to capture the executed command
	originalPreRun := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		executedCmd = cmd
		if originalPreRun != nil {
			return originalPreRun(cmd, args)
		}
		return nil
	}

	err := rootCmd.Execute()
	return executedCmd, err
}

// isArgValidationError determines if an error is from argument validation.
func isArgValidationError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Common argument validation error patterns from Cobra
	argPatterns := []string{
		"accepts",
		"requires",
		"requires at least",
		"requires at most",
		"accepts at most",
		"too many arguments",
		"unknown command",
	}

	for _, pattern := range argPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// getDoctorExitCode returns the appropriate exit code for doctor command errors.
func getDoctorExitCode(err error) int {
	if err == nil {
		return 0
	}

	errMsg := err.Error()

	// Doctor command uses specific error messages for different health states
	if strings.Contains(errMsg, "health check detected errors") {
		return 2
	}
	if strings.Contains(errMsg, "health check detected warnings") {
		return 1
	}

	// Default error exit code
	return 1
}
