package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/term"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/config"
	"github.com/jamesainslie/dot/internal/updater"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/spf13/cobra"
)

// Global configuration shared across commands
type globalConfig struct {
	packageDir string
	targetDir  string
	backupDir  string
	dryRun     bool
	verbose    int
	quiet      bool
	logJSON    bool
	noColor    bool
	cpuProfile string
	memProfile string
	pprofAddr  string
}

var globalCfg globalConfig

// NewRootCommand creates the root cobra command.
func NewRootCommand(version, commit, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dot",
		Short: "Modern symlink manager for dotfiles",
		Long: `dot is a type-safe dotfile manager written in Go.

dot manages dotfiles by creating symlinks from a source directory 
(package directory) to a target directory. It provides atomic operations,
comprehensive conflict detection, and incremental updates.`,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Perform startup version check (async, non-blocking)
			go performStartupVersionCheckAsync(version)
			return nil
		},
	}

	// Set up flag error function to show usage on flag parsing errors
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
		_ = cmd.Usage()
		return err
	})

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&globalCfg.packageDir, "dir", "d", ".",
		"Source directory containing packages")

	// Compute cross-platform home directory default
	defaultTarget, err := os.UserHomeDir()
	if err != nil || defaultTarget == "" {
		// Fall back to current working directory
		defaultTarget, err = os.Getwd()
		if err != nil || defaultTarget == "" {
			defaultTarget = "."
		}
	}

	rootCmd.PersistentFlags().StringVarP(&globalCfg.targetDir, "target", "t", defaultTarget,
		"Target directory for symlinks")
	rootCmd.PersistentFlags().StringVar(&globalCfg.backupDir, "backup-dir", "",
		"Directory for backup files (default: <target>/.dot-backup)")
	rootCmd.PersistentFlags().BoolVarP(&globalCfg.dryRun, "dry-run", "n", false,
		"Show what would be done without applying changes")
	rootCmd.PersistentFlags().CountVarP(&globalCfg.verbose, "verbose", "v",
		"Increase verbosity: -v (info), -vv (debug), -vvv (trace)")
	rootCmd.PersistentFlags().BoolVarP(&globalCfg.quiet, "quiet", "q", false,
		"Suppress all non-error output")
	rootCmd.PersistentFlags().BoolVar(&globalCfg.logJSON, "log-json", false,
		"Output logs in JSON format")
	rootCmd.PersistentFlags().BoolVar(&globalCfg.noColor, "no-color", false,
		"Disable color output")
	rootCmd.PersistentFlags().StringVar(&globalCfg.cpuProfile, "cpu-profile", "",
		"Write CPU profile to file (for diagnostics)")
	rootCmd.PersistentFlags().StringVar(&globalCfg.memProfile, "mem-profile", "",
		"Write memory profile to file (for diagnostics)")
	rootCmd.PersistentFlags().StringVar(&globalCfg.pprofAddr, "pprof", "",
		"Enable pprof HTTP server on address (e.g. :6060)")
	rootCmd.PersistentFlags().Bool("batch", false,
		"Batch mode for scripting (implies --quiet)")

	// Add subcommands
	rootCmd.AddCommand(
		newManageCommand(),
		newUnmanageCommand(),
		newRemanageCommand(),
		newAdoptCommand(),
		newStatusCommand(),
		newListCommand(),
		newDoctorCommand(),
		newConfigCommand(),
		newCloneCommand(),
		newUpgradeCommand(version),
	)

	return rootCmd
}

// buildConfig creates a dot.Config from global flags and adapters.
// Precedence: flags (if set) > config file > defaults
func buildConfig() (dot.Config, error) {
	return buildConfigWithCmd(nil)
}

// buildConfigWithCmd creates config with flag precedence awareness.
func buildConfigWithCmd(cmd *cobra.Command) (dot.Config, error) {
	// Check batch mode first (if cmd is provided)
	if cmd != nil {
		if batch, _ := cmd.Flags().GetBool("batch"); batch {
			globalCfg.quiet = true
		}
	}

	// Create adapters
	fs := adapters.NewOSFilesystem()
	logger := createLogger()

	// Load extended config - check repo location first, then XDG location
	configPath := getConfigFilePath()
	extCfg, err := loadConfigWithRepoPriority(configPath)
	if err != nil {
		return dot.Config{}, fmt.Errorf("load configuration: %w", err)
	}

	// Start with config file values
	var packageDir, targetDir, backupDir, manifestDir string
	var backup, overwrite bool

	if extCfg != nil {
		packageDir = extCfg.Directories.Package
		targetDir = extCfg.Directories.Target
		backupDir = extCfg.Symlinks.BackupDir
		manifestDir = extCfg.Directories.Manifest
		backup = extCfg.Symlinks.Backup
		overwrite = extCfg.Symlinks.Overwrite
	}

	// Override with globalCfg if set (covers both flag and test scenarios)
	// For flags to override config, they must be non-default values
	if globalCfg.packageDir != "" && globalCfg.packageDir != "." {
		packageDir = globalCfg.packageDir
	}

	homeDir, _ := os.UserHomeDir()
	if globalCfg.targetDir != "" && globalCfg.targetDir != homeDir {
		targetDir = globalCfg.targetDir
	}

	if globalCfg.backupDir != "" {
		backupDir = globalCfg.backupDir
	}

	// Apply final defaults if still empty
	if packageDir == "" {
		packageDir = "."
	}
	if targetDir == "" {
		targetDir, _ = os.UserHomeDir()
		if targetDir == "" {
			targetDir = "."
		}
	}

	// Make paths absolute
	packageDir, err = filepath.Abs(packageDir)
	if err != nil {
		return dot.Config{}, fmt.Errorf("invalid package directory: %w", err)
	}

	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		return dot.Config{}, fmt.Errorf("invalid target directory: %w", err)
	}

	cfg := dot.Config{
		PackageDir:         packageDir,
		TargetDir:          targetDir,
		BackupDir:          backupDir,
		Backup:             backup,
		Overwrite:          overwrite,
		ManifestDir:        manifestDir,
		DryRun:             globalCfg.dryRun,
		Verbosity:          globalCfg.verbose,
		PackageNameMapping: true, // Default: true (pre-1.0 breaking change)
		FS:                 fs,
		Logger:             logger,
	}

	return cfg.WithDefaults(), nil
}

// loadConfigWithRepoPriority loads config checking repository location first.
//
// Priority order:
//  1. If packageDir flag is set, check <packageDir>/.config/dot/config.yaml
//  2. Otherwise, check standard repo location ~/.dotfiles/.config/dot/config.yaml
//  3. Fall back to XDG location (provided configPath)
//  4. Use defaults
//
// This allows repositories to define their own configuration without circular dependency.
func loadConfigWithRepoPriority(xdgConfigPath string) (*config.ExtendedConfig, error) {
	var packageDir string

	// Check if packageDir was explicitly set via flag
	if globalCfg.packageDir != "" && globalCfg.packageDir != "." {
		packageDir = globalCfg.packageDir
	} else {
		// Use default packageDir location
		homeDir, err := os.UserHomeDir()
		if err == nil {
			packageDir = filepath.Join(homeDir, ".dotfiles")
		}
	}

	// Try to load from repository first
	if packageDir != "" {
		repoConfigPath := filepath.Join(packageDir, ".config", "dot", "config.yaml")
		if _, err := os.Stat(repoConfigPath); err == nil {
			// Repository config exists - use it
			loader := config.NewLoader("dot", repoConfigPath)
			cfg, err := loader.LoadWithEnv()
			if err == nil {
				return cfg, nil
			}
			// If repo config exists but fails to load, that's an error
			return nil, fmt.Errorf("load repository config: %w", err)
		}
	}

	// Fall back to XDG location
	loader := config.NewLoader("dot", xdgConfigPath)
	return loader.LoadWithEnv()
}

// createLogger creates appropriate logger based on flags.
func createLogger() dot.Logger {
	if globalCfg.quiet {
		return adapters.NewNoopLogger()
	}

	level := verbosityToLevel(globalCfg.verbose)

	if globalCfg.logJSON {
		return adapters.NewSlogLogger(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})))
	}

	return adapters.NewSlogLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}

// verbosityToLevel converts verbosity count to log level.
// Level mapping:
//   - 0 (no flag): ERROR only - suppress all logs, show only user messages
//   - 1 (-v): INFO - show high-level progress
//   - 2 (-vv): DEBUG - show detailed operation info
//   - 3+ (-vvv): More verbose DEBUG levels
func verbosityToLevel(v int) slog.Level {
	switch {
	case v == 0:
		return slog.LevelError // Suppress INFO/DEBUG/WARN, only show errors
	case v == 1:
		return slog.LevelInfo // Show high-level progress
	case v == 2:
		return slog.LevelDebug // Show detailed operations
	default:
		// Even more verbose
		return slog.LevelDebug - slog.Level(v-2)
	}
}

// formatError converts domain errors to user-friendly messages.
func formatError(err error) error {
	// For now, just return the error
	// In the future, this can be enhanced to provide better error messages
	return err
}

// argsWithUsage wraps a Cobra Args validator to show usage on validation errors.
func argsWithUsage(validator cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		err := validator(cmd, args)
		if err != nil {
			// Print error and usage
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
			_ = cmd.Usage()
		}
		return err
	}
}

// shouldColorize determines if output should be colorized based on the color flag.
// Precedence: --no-color flag > NO_COLOR env > --color flag > auto
func shouldColorize(color string) bool {
	// Check --no-color flag first (highest precedence)
	if globalCfg.noColor {
		return false
	}

	// Respect NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check --color flag value
	switch color {
	case "always":
		return true
	case "never":
		return false
	case "auto":
		// Check if stdout is a terminal using portable detection
		return term.IsTerminal(int(os.Stdout.Fd()))
	default:
		// Default to auto behavior
		return term.IsTerminal(int(os.Stdout.Fd()))
	}
}

// performStartupVersionCheck performs a non-blocking version check at startup.
func performStartupVersionCheck(currentVersion string) {
	// Don't check if this is a dev build
	if currentVersion == "dev" {
		return
	}

	// Load configuration
	configPath := getConfigFilePath()
	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	if err != nil {
		// If config fails to load, use defaults (which has checking disabled by default)
		cfg = config.DefaultExtended()
	}

	// Don't check if disabled
	if !cfg.Update.CheckOnStartup {
		return
	}

	// Perform check
	configDir := filepath.Dir(configPath)
	checker := updater.NewStartupChecker(currentVersion, cfg, configDir, os.Stdout)
	result, err := checker.Check()
	if err != nil {
		return // Silent failure
	}
	checker.ShowNotification(result)
}

// performStartupVersionCheckAsync performs an async version check with timeout.
func performStartupVersionCheckAsync(currentVersion string) {
	// Create a context with timeout to prevent hanging
	// Use 3 seconds to allow for DNS resolution and network latency
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Run check in a channel
	done := make(chan struct{})
	go func() {
		performStartupVersionCheck(currentVersion)
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// Check completed
		return
	case <-ctx.Done():
		// Timeout - silently abort
		return
	}
}
