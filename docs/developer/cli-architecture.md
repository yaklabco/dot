# CLI Layer Architecture

## Overview

The CLI layer (`cmd/dot/`) is the top-most layer in the dot architecture. It depends exclusively on the public API (`pkg/dot/`) and does not import any `internal/` packages directly, except for CLI-specific rendering and UI helpers in `internal/cli/`.

## Layering Rules

### Allowed Dependencies

- `pkg/dot/*` - Public API (required)
- `internal/cli/*` - CLI-specific UI/rendering helpers (allowed for presentation logic)
- Standard library
- Third-party CLI frameworks (cobra, pflag, etc.)

### Prohibited Dependencies  

- `internal/adapters` - Use `pkg/dot.NewOSFilesystem()`, `pkg/dot.NewSlogLogger()`, etc.
- `internal/config` - Use `pkg/dot.ConfigLoader`, `pkg/dot.ExtendedConfig`
- `internal/domain` - Use types re-exported from `pkg/dot`
- `internal/executor` - Use `pkg/dot.Client` methods
- `internal/manifest` - Use `pkg/dot.Client` methods
- `internal/pipeline` - Use `pkg/dot.Client` methods
- `internal/planner` - Use `pkg/dot.Client` methods
- `internal/scanner` - Use `pkg/dot` helper functions
- `internal/updater` - Use `pkg/dot.VersionChecker`, `pkg/dot.StartupChecker`
- Any other `internal/*` package

## Configuration Responsibility

### Configuration Loading

Configuration loading follows this precedence (highest to lowest):

1. Command-line flags (parsed by cobra)
2. Environment variables (`DOT_*` prefix)
3. Configuration files (repository, XDG, system)
4. Built-in defaults

The CLI layer:
- Parses command-line flags using cobra
- Calls `pkg/dot.NewConfigLoader()` to load extended configuration
- Merges flag values over loaded configuration
- Builds `dot.Config` from the merged result
- Creates `dot.Client` with the final configuration

The configuration layer (`internal/config`, exposed via `pkg/dot`):
- Loads configuration from files
- Applies environment variable overrides  
- Validates configuration values
- Provides defaults

### Startup Version Checking

Startup version checking is a **CLI concern** implemented in `cmd/dot/root.go`:

- Triggered in `PersistentPreRunE` hook (runs before all commands)
- Uses `pkg/dot.StartupChecker` (facade for `internal/updater`)
- Runs asynchronously with 3-second timeout
- Fails silently (does not block or error on failure)
- Respects `update.check_on_startup` and `update.check_frequency` config

**Rationale**: Version checking is presentation-layer functionality tied to the CLI lifecycle. Library users of `pkg/dot` should not have update notifications injected into their applications. The API layer remains pure domain logic without CLI concerns.

## State Management

### Global State (Anti-Pattern)

The current `globalCfg` variable in `root.go` is **global mutable state** and represents a deviation from functional principles. This is a known anti-pattern that should be refactored.

**Current State** (Technical Debt):
```go
var globalCfg globalConfig  // Mutable global state
```

**Planned Refactoring**:
- Pass configuration through cobra command context
- Use a `CLIContext` struct attached to cobra commands
- Make configuration construction pure functions with explicit inputs

### Command Execution Flow

1. User invokes command (e.g., `dot manage vim`)
2. Cobra parses flags and routes to command handler
3. Command handler calls `buildConfig()` or `buildConfigWithCmd(cmd)`
4. Configuration builder:
   - Loads extended config from file/env (`pkg/dot.ConfigLoader`)
   - Applies flag overrides from cobra
   - Creates adapters (`pkg/dot.NewOSFilesystem()`, `pkg/dot.NewSlogLogger()`)
   - Returns `dot.Config`
5. Command handler creates `dot.Client` with config
6. Command handler invokes client methods (e.g., `client.Manage()`)
7. Results are rendered using `internal/cli/render` helpers

## CLI-Specific Helpers

The `internal/cli/` packages provide rendering and UI helpers:

- `internal/cli/adopt` - Interactive adoption workflow
- `internal/cli/errors` - Error formatting and suggestions
- `internal/cli/golden` - Golden file testing for CLI output
- `internal/cli/help` - Help text formatting
- `internal/cli/output` - Output formatting (JSON, YAML, table)
- `internal/cli/pretty` - Pretty-printing for diagnostics
- `internal/cli/progress` - Progress indicators
- `internal/cli/prompt` - User prompts and confirmations
- `internal/cli/render` - Color and styling
- `internal/cli/renderer` - Table rendering
- `internal/cli/selector` - Interactive selection menus
- `internal/cli/terminal` - Terminal detection and control

These packages are CLI presentation concerns and do not contain domain logic. They may remain as direct imports in the CLI layer.

## Testing

CLI tests (`cmd/dot/*_test.go`) should:
- Use `pkg/dot` types and interfaces
- Mock adapters (filesystem, logger) using `pkg/dot` constructors
- Test flag parsing and configuration precedence
- Test error formatting and user-facing messages
- Use golden files for output snapshot testing

Integration tests (`tests/integration/`) should:
- Execute the CLI binary directly
- Verify end-to-end behavior
- Test against real (or memory-backed) filesystems
- Validate exit codes and output formats

## Migration Notes

The following facade functions were added to `pkg/dot` to support CLI layer isolation:

- `GetConfigPath(appName)` - XDG config path resolution
- `NewConfigLoader(appName, path)` - Configuration loader
- `ConfigLoader.LoadWithEnv()` - Load with environment overrides
- `NewConfigWriter(path)` - Configuration writer
- `ConfigWriter.WriteDefault(opts)` - Write default configuration
- `ConfigWriter.Update(key, value)` - Update configuration value
- `UpgradeConfig(path, force)` - Upgrade configuration format
- `NewVersionChecker(repository)` - Version checker
- `VersionChecker.CheckForUpdate(version, prerelease)` - Check for updates
- `NewStartupChecker(version, cfg, configDir, output)` - Startup version checker
- `StartupChecker.Check()` - Perform startup check
- `StartupChecker.ShowNotification(result)` - Display update notification
- `ResolvePackageManager(configured)` - Resolve package manager for upgrades
- `DefaultSensitivePatterns()` - Secret detection patterns
- `DetectSecrets(files, patterns)` - Detect secrets in files
- `UntranslateDotfile(name)` - Convert dotfile name to package name

These facades maintain a stable CLI API while allowing internal refactoring.

