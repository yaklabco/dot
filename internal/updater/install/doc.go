// Package install provides installation detection and upgrade functionality for dot.
//
// The package detects how dot was installed (Homebrew, APT, Pacman, Chocolatey,
// go install, or source build) using pure Go filesystem operations without
// shelling out to external commands.
//
// For package manager installations, it provides secure upgrade execution using
// type-safe command construction with validation to prevent shell injection.
//
// # Detection
//
// Detection uses a probe chain pattern where each probe checks for a specific
// installation method. Probes are ordered by platform and priority, with the
// first matching probe determining the installation source.
//
// # Upgrade Execution
//
// Upgrades are executed via validated commands constructed from whitelisted
// specifications. All dynamic arguments are validated against regex patterns,
// and shell metacharacters are explicitly rejected.
//
// # Security
//
// The package follows these security principles:
//   - No shell interpolation (no sh -c)
//   - Whitelisted command structures only
//   - Regex validation for dynamic arguments
//   - Explicit shell metacharacter rejection
//   - Direct exec.Command usage without shell
package install
