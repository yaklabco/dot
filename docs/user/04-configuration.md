# Configuration Reference

Complete reference for configuring dot.

## Configuration Sources

dot reads configuration from multiple sources with defined precedence order.

### Precedence Order (Highest to Lowest)

1. **Command-line flags**: `--dir`, `--target`, etc.
2. **Environment variables**: `DOT_*` prefix
3. **Project-local config**: `./.dotrc` in current directory
4. **User global config**: `~/.config/dot/config.yaml` or `~/.dotrc`
5. **System config**: `/etc/dot/config.yaml`
6. **Built-in defaults**

Later sources override earlier sources for scalar values. Array merging behavior is configurable.

## Configuration File Locations

### XDG Base Directory Specification

dot follows XDG standards on Unix systems:

**Primary location**: `$XDG_CONFIG_HOME/dot/config.yaml`
- Default: `~/.config/dot/config.yaml`

**Fallback location**: `~/.dotrc`

**System-wide**: `/etc/dot/config.yaml`

**Project-local**: `./.dotrc` in working directory

### macOS Specific

Uses XDG paths by default:
- `~/.config/dot/config.yaml`
- Or `~/Library/Application Support/dot/config.yaml`

### Windows Specific

- `%APPDATA%\dot\config.yaml`
- Typically: `C:\Users\<username>\AppData\Roaming\dot\config.yaml`

## Supported Formats

dot accepts configuration in multiple formats (detected by extension):

- **YAML**: `.yaml` or `.yml` (recommended)
- **JSON**: `.json`
- **TOML**: `.toml`

All examples below use YAML format.

## Configuration Management

### Creating Configuration

Create a new configuration file with defaults:

```bash
dot config init
```

This creates `~/.config/dot/config.yaml` with default values and documentation comments.

Options:
- `--force`: Overwrite existing configuration
- `--format`: Specify format (yaml, json, toml)

### Upgrading Configuration

When upgrading dot to a new version, configuration files may need updating to include new fields or migrate deprecated options.

#### Automatic Upgrade

Upgrade your configuration safely:

```bash
dot config upgrade
```

The upgrade process:

1. **Creates backup**: Timestamped backup saved to `~/.config/dot/backups/`
2. **Merges configuration**: Preserves your customizations while adding new fields
3. **Migrates deprecated fields**: Converts old patterns to new syntax
4. **Validates result**: Ensures upgraded config is valid before saving
5. **Cleans old backups**: Keeps last 5 backups automatically

#### Force Upgrade

Skip the confirmation prompt:

```bash
dot config upgrade --force
```

#### What Gets Preserved

During upgrade, your customizations are preserved:

- Directory paths
- Logging settings
- Symlink modes and options
- Custom ignore patterns
- All user-modified values

#### What Gets Added

New configuration fields from the latest version are added with their defaults:

- New feature flags
- Additional configuration options
- Performance tuning parameters

#### Deprecated Field Migration

Some fields may be deprecated in favor of better alternatives. The upgrade command automatically migrates these:

**Example**: `ignore.overrides` → `ignore.patterns` with negation (`!`)

```yaml
# Before upgrade
ignore:
  patterns:
    - "*.log"
  overrides:
    - "important.conf"

# After upgrade
ignore:
  patterns:
    - "*.log"
    - "!important.conf"  # Migrated from overrides
  overrides:  # Preserved for safety but deprecated
    - "important.conf"
```

#### Backup Management

Backups are stored in `~/.config/dot/backups/` with format:

```
YYYYMMDD-HHMMSS-config.bak
```

Example: `20241110-153045-config.bak`

The last 5 backups are automatically retained. Older backups are cleaned up after each successful upgrade.

#### Manual Rollback

If needed, restore from a backup:

```bash
cp ~/.config/dot/backups/20241110-153045-config.bak ~/.config/dot/config.yaml
```

#### Upgrade Header

Upgraded configuration files include a header with upgrade information:

```yaml
# Dot Configuration
# Upgraded on 2024-11-10 15:30:45
# Backup saved to: /home/user/.config/dot/backups/20241110-153045-config.bak
# 
# Deprecated fields migrated:
#   - ignore.overrides → ignore.patterns (with ! prefix for negation)
#
# See https://github.com/yaklabco/dot for documentation

directories:
  package: "."
  # ... rest of configuration
```

### Viewing Configuration

View current configuration:

```bash
dot config list
```

Get a specific value:

```bash
dot config get directories.package
```

Show configuration file path:

```bash
dot config path
```

### Modifying Configuration

Set a specific value:

```bash
dot config set directories.package ~/dotfiles
```

## Configuration Options

### Directory Options

#### packageDir

Source directory containing packages.

**Type**: string  
**Default**: `.` (current directory)  
**Example**:
```yaml
packageDir: ~/dotfiles
```

Absolute or relative paths accepted. Relative paths resolved from working directory.

#### targetDir

Destination directory where symlinks are created.

**Type**: string  
**Default**: `$HOME`  
**Example**:
```yaml
targetDir: ~
```

Typically set to home directory. Must be absolute path or tilde-expandable.

#### manifestDir

Directory where package manifest is stored.

**Type**: string  
**Default**: `$XDG_DATA_HOME/dot/manifest` or `~/.local/share/dot/manifest`  
**Example**:
```yaml
directories:
  package: ~/dotfiles
  target: ~
  manifest: ~/.local/share/dot/manifest
```

The manifest tracks installed packages, their links, and content hashes for incremental updates. 

**Note**: The manifest is a single JSON file stored as `.dot-manifest.json` within this directory.

### Link Options

#### linkMode

Symlink type to create.

**Type**: string  
**Default**: `relative`  
**Values**: `relative` or `absolute`  
**Example**:
```yaml
linkMode: relative
```

**Relative links**:
- Portable across different mount points
- Break if package directory moves relative to target
- Recommended for most use cases

**Absolute links**:
- Robust against package directory relocation
- Less portable across machines
- Use when target and stow on different filesystems

#### folding

Enable directory-level symlink optimization.

**Type**: boolean  
**Default**: `true`  
**Example**:
```yaml
folding: true
```

When enabled, creates directory symlinks when all contents belong to single package. Disable for per-file granularity.

### Package Translation

#### packageNameMapping

Enable package name to target directory mapping.

**Type**: boolean  
**Default**: `true`  
**Example**:
```yaml
dotfile:
  package_name_mapping: true
```

When enabled, package names determine target directories:
- Package `dot-vim` → files installed to `~/.vim/`
- Package `dot-gnupg` → files installed to `~/.gnupg/`
- Package `config` → files installed to `~/config/`

This eliminates redundant directory nesting. Package `dot-gnupg/gpg.conf` links to `~/.gnupg/gpg.conf` instead of `~/gpg.conf`.

**Breaking Change in v0.2.0**: This is now the default behavior. Set to `false` to restore legacy behavior where package names are ignored.

**Legacy behavior** (`false`):
- Package structure must mirror target structure
- Package name used only for identification
- Requires redundant nesting like `dot-vim/dot-vim/`

### Ignore Patterns

#### ignore

Patterns for files to exclude from management.

**Type**: array of strings  
**Default**: Built-in patterns (see below)  
**Example**:
```yaml
ignore:
  - "*.log"
  - "*.tmp"
  - ".git"
  - ".DS_Store"
  - "node_modules"
  - "*.swp"
  - "*.bak"
  - ".#*"
```

**Pattern Types**:
- **Glob patterns**: `*.log`, `test_*`
- **Regex patterns**: `/^test.*\.go$/`
- **Directory patterns**: `.git`, `node_modules`

**Built-in Default Patterns**:
```yaml
ignore:
  - ".git"
  - ".DS_Store"
  - "Thumbs.db"
  - "desktop.ini"
  - "*.swp"
  - "*.swo"
  - "*~"
  - ".#*"
  - "#*#"
```

#### override (DEPRECATED)

Patterns to force include despite ignore rules.

**Type**: array of strings  
**Default**: `[]`  
**Deprecated**: Use negation patterns (`!pattern`) instead  
**Example**:
```yaml
override:
  - ".gitignore"
  - ".gitconfig"
```

Override patterns have higher priority than ignore patterns.

**Note**: This field is deprecated. Use negation patterns in the `patterns` array instead:
```yaml
patterns:
  - ".*"           # Ignore all dotfiles
  - "!.gitignore"  # Except .gitignore
  - "!.gitconfig"  # And .gitconfig
```

#### per_package_ignore

Enable per-package `.dotignore` files.

**Type**: boolean  
**Default**: `true`  
**Example**:
```yaml
per_package_ignore: true
```

When enabled, dot reads `.dotignore` files from package directories. These files use the same syntax as patterns but are scoped to the package.

#### max_file_size

Maximum file size to include when scanning packages (in bytes).

**Type**: integer  
**Default**: `0` (no limit)  
**Example**:
```yaml
max_file_size: 104857600  # 100 MB
```

Files exceeding this limit are either prompted for (interactive mode) or skipped automatically (batch mode). Use `0` to disable size filtering.

**Human-readable format**: Use command-line flags with human-readable sizes:
```bash
dot manage mypackage --max-file-size 100MB
```

#### interactive_large_files

Prompt for large files in interactive mode.

**Type**: boolean  
**Default**: `true`  
**Example**:
```yaml
interactive_large_files: true
```

When `true` and running in a TTY, prompts user whether to include files exceeding `max_file_size`. When `false` or in batch mode, large files are automatically skipped.

**See also**: [Ignore System Guide](ignore-system.md) for complete documentation on ignore patterns, negation, and size filtering.

### Conflict Resolution

#### onConflict

Default policy when conflicts detected.

**Type**: string  
**Default**: `fail`  
**Values**: `fail`, `backup`, `overwrite`, `skip`  
**Example**:
```yaml
onConflict: fail
```

**Policies**:
- `fail`: Stop and report conflict (safe default)
- `backup`: Move conflicting file to backup location
- `overwrite`: Replace conflicting file with symlink
- `skip`: Skip conflicting file and continue

#### backupDir

Directory for storing conflict backups.

**Type**: string  
**Default**: None (backups stored alongside originals with `.bak` suffix)  
**Example**:
```yaml
backupDir: ~/.dot-backups
```

When set, all backups stored in specified directory with timestamp.

### Logging and Output

#### verbosity

Logging detail level.

**Type**: integer  
**Default**: `0`  
**Range**: `0-3`  
**Example**:
```yaml
verbosity: 1
```

**Levels**:
- `0`: Errors and warnings only
- `1`: Info messages (operations summary)
- `2`: Debug messages (per-operation details)
- `3`: Trace messages (internal state)

Command-line `-v` flags increment this value.

#### logFormat

Log output format.

**Type**: string  
**Default**: `text`  
**Values**: `text`, `json`  
**Example**:
```yaml
logFormat: text
```

**Formats**:
- `text`: Human-readable console output with colors
- `json`: Structured JSON for log aggregation

#### quiet

Suppress non-error output.

**Type**: boolean  
**Default**: `false`  
**Example**:
```yaml
quiet: false
```

When `true`, only errors printed to stderr. Useful for scripting.

### Performance Options

#### concurrency

Maximum concurrent operations.

**Type**: integer  
**Default**: `0` (auto-detect CPU cores)  
**Example**:
```yaml
concurrency: 4
```

Set to number of parallel operations. Value of `0` uses number of CPU cores. Higher values may improve performance with many packages.

#### enableIncremental

Enable incremental change detection.

**Type**: boolean  
**Default**: `true`  
**Example**:
```yaml
enableIncremental: true
```

When enabled, `remanage` only processes changed packages using content hashing.

## Per-Package Configuration

Package-specific overrides via `.dotmeta` file in package directory.

### Package Metadata Format

`package/.dotmeta`:
```yaml
# Package-specific settings
ignore:
  - "*.local"
  - "cache/"

linkMode: absolute

folding: false

# Package metadata
description: "Vim configuration with plugins"
version: "1.0.0"
```

Package settings override global settings for that package only.

## Environment Variables

All configuration options available as environment variables with `DOT_` prefix.

### Variable Naming Convention

```bash
# Format: DOT_<OPTION_NAME>
export DOT_STOW_DIR=~/dotfiles
export DOT_TARGET_DIR=~
export DOT_LINK_MODE=relative
export DOT_FOLDING=true
export DOT_VERBOSITY=1
```

**Naming Rules**:
- Uppercase with underscores
- Nested options use double underscore: `DOT_ON_CONFLICT`
- Boolean values: `true`/`false`, `1`/`0`
- Arrays: comma-separated: `DOT_IGNORE=*.log,*.tmp`

### Common Environment Variables

```bash
# Directories
export DOT_STOW_DIR=/path/to/dotfiles
export DOT_TARGET_DIR=$HOME

# Link mode
export DOT_LINK_MODE=absolute

# Conflict handling
export DOT_ON_CONFLICT=backup
export DOT_BACKUP_DIR=~/.dot-backups

# Ignore patterns
export DOT_IGNORE="*.log,*.tmp,.git"

# Output control
export DOT_VERBOSITY=2
export DOT_LOG_FORMAT=json
export DOT_QUIET=false

# Performance
export DOT_CONCURRENCY=4
```

## Complete Configuration Example

### Comprehensive YAML Example

`~/.config/dot/config.yaml`:
```yaml
# Dot configuration file

# Directories
packageDir: ~/dotfiles
targetDir: ~

# Link configuration
linkMode: relative
folding: true

# Ignore patterns
ignore:
  - "*.log"
  - "*.tmp"
  - ".git"
  - ".svn"
  - ".DS_Store"
  - "Thumbs.db"
  - "node_modules"
  - "*.swp"
  - "*.swo"
  - "*.bak"
  - "*~"
  - ".#*"
  - "#*#"

# Override patterns (force include despite ignore)
override:
  - ".gitignore"
  - ".gitattributes"

# Conflict resolution
onConflict: fail
backupDir: ~/.dot-backups

# Logging
verbosity: 1
logFormat: text
quiet: false

# Performance
concurrency: 0  # Auto-detect
enableIncremental: true

# Package-specific overrides
packages:
  vim:
    linkMode: absolute
    folding: false
  
  nvim:
    ignore:
      - "*.local"
      - "shada/"
```

### JSON Example

`~/.config/dot/config.json`:
```json
{
  "packageDir": "~/dotfiles",
  "targetDir": "~",
  "linkMode": "relative",
  "folding": true,
  "ignore": [
    "*.log",
    ".git",
    ".DS_Store"
  ],
  "onConflict": "fail",
  "verbosity": 1
}
```

### TOML Example

`~/.config/dot/config.toml`:
```toml
packageDir = "~/dotfiles"
targetDir = "~"
linkMode = "relative"
folding = true

ignore = [
    "*.log",
    ".git",
    ".DS_Store"
]

onConflict = "fail"
verbosity = 1

[packages.vim]
linkMode = "absolute"
folding = false
```

## Configuration Management Commands

### Initialize Configuration

Create default configuration file:

```bash
# Interactive initialization
dot config init

# Non-interactive with defaults
dot config init --defaults

# Specify format
dot config init --format yaml
```

### View Configuration

Display current configuration:

```bash
# Show all configuration
dot config show

# Show specific key
dot config get packageDir
dot config get linkMode

# Show as JSON
dot config show --format json
```

### Modify Configuration

Update configuration values:

```bash
# Set value
dot config set packageDir ~/dotfiles
dot config set verbosity 2

# Set array value
dot config set ignore "*.log,*.tmp"

# Unset value (use default)
dot config unset backupDir
```

### Validate Configuration

Check configuration validity:

```bash
# Validate current configuration
dot config validate

# Validate specific file
dot config validate ~/custom-config.yaml
```

Expected output:
```
Configuration valid
- Package directory exists
- Target directory exists
- All paths are absolute
- No conflicting options
```

### Configuration File Location

Find active configuration file:

```bash
dot config path
```

Output:
```
/home/user/.config/dot/config.yaml
```

## Configuration Scenarios

### Scenario 1: Multiple Machine Setup

Use different configurations per machine:

**Laptop** (`~/.config/dot/config.yaml`):
```yaml
packageDir: ~/dotfiles
targetDir: ~
linkMode: relative
verbosity: 1
```

**Server** (`~/.config/dot/config.yaml`):
```yaml
packageDir: /opt/dotfiles
targetDir: /home/admin
linkMode: absolute  # Different filesystem
verbosity: 0        # Less output
quiet: true
```

### Scenario 2: Per-Project Configuration

Project-specific settings:

`./.dotrc`:
```yaml
# Project-specific dot configuration
packageDir: ./config
targetDir: ./deployment
linkMode: absolute
onConflict: overwrite  # Aggressive for development
verbosity: 2
```

### Scenario 3: CI/CD Environment

Non-interactive, scripted usage:

```yaml
# CI configuration
packageDir: /build/configs
targetDir: /app
linkMode: absolute
onConflict: overwrite
verbosity: 0
logFormat: json
quiet: true
enableIncremental: false  # Always full deployment
```

Environment variables:
```bash
export DOT_STOW_DIR=/build/configs
export DOT_TARGET_DIR=/app
export DOT_ON_CONFLICT=overwrite
export DOT_QUIET=true
```

### Scenario 4: Shared Team Configuration

Team-wide defaults with personal overrides:

**System** (`/etc/dot/config.yaml`):
```yaml
# Team defaults
packageDir: ~/dotfiles
linkMode: relative
folding: true

ignore:
  - "*.log"
  - "*.local"  # Personal overrides
  - ".git"

onConflict: fail  # Safe default
```

**Personal** (`~/.config/dot/config.yaml`):
```yaml
# Personal overrides
verbosity: 2     # I want more detail
onConflict: backup  # I prefer backups

# Additional ignore patterns
ignore:
  - "*.debug"
```

## Merge Strategies

### Array Merge Behavior

Configure how arrays merge across configuration sources:

```yaml
# In global config
ignore:
  - "*.log"
  - ".git"

# In local config with replace strategy
ignoreStrategy: replace
ignore:
  - "*.tmp"
# Result: ["*.tmp"] - local replaces global

# With append strategy (default)
ignoreStrategy: append
ignore:
  - "*.tmp"
# Result: ["*.log", ".git", "*.tmp"] - local appends to global

# With merge strategy (union)
ignoreStrategy: merge
ignore:
  - "*.tmp"
# Result: ["*.log", ".git", "*.tmp"] - deduplicated union
```

**Strategies**:
- `append`: Add local to global (default)
- `replace`: Local completely replaces global
- `merge`: Union of global and local (deduplicated)

## Configuration Best Practices

### 1. Use Version Control

Store configuration in repository:

```bash
cd ~/dotfiles
cp ~/.config/dot/config.yaml dot-config.yaml
git add dot-config.yaml
git commit -m "docs(config): add dot configuration"
```

### 2. Environment-Specific Files

Separate configurations per environment:

```
dotfiles/
├── dot-config-laptop.yaml
├── dot-config-desktop.yaml
└── dot-config-server.yaml
```

Symlink appropriate file:
```bash
ln -s ~/dotfiles/dot-config-laptop.yaml ~/.config/dot/config.yaml
```

### 3. Minimal Configuration

Only set values that differ from defaults:

```yaml
# Minimal - only what changes
packageDir: ~/dotfiles
verbosity: 1
```

Better than:
```yaml
# Verbose - includes all defaults
packageDir: ~/dotfiles
targetDir: ~
linkMode: relative
folding: true
verbosity: 1
# ... etc
```

### 4. Document Custom Settings

Add comments explaining non-obvious choices:

```yaml
# Use absolute links because stow and target on different filesystems
linkMode: absolute

# Aggressive conflict resolution for development environment
onConflict: overwrite

# High verbosity for debugging
verbosity: 3
```

### 5. Validate Before Committing

Always validate configuration changes:

```bash
# Edit configuration
vim ~/.config/dot/config.yaml

# Validate
dot config validate

# Test with dry-run
dot --dry-run manage test-package
```

## Troubleshooting Configuration

### Configuration Not Loading

Check precedence and file location:

```bash
# Show active configuration
dot config show

# Show configuration file path
dot config path

# Verify file exists
ls -la $(dot config path)
```

### Configuration Errors

Validate syntax:

```bash
# Check for syntax errors
dot config validate

# Show detailed validation errors
dot config validate -v
```

Common errors:
- Invalid YAML/JSON/TOML syntax
- Unknown configuration keys
- Invalid values for options
- Path not existing
- Permission issues

### Unexpected Values

Debug configuration resolution:

```bash
# Show all sources and precedence
dot config show --all-sources

# Show where each value comes from
dot config show --with-sources
```

Example output:
```
packageDir: ~/dotfiles
  Source: ~/.config/dot/config.yaml

targetDir: /tmp/test
  Source: command-line flag

verbosity: 2
  Source: environment variable (DOT_VERBOSITY)
```

## Next Steps

- [Command Reference](05-commands.md): Learn all commands with configuration options
- [Common Workflows](06-workflows.md): See configuration in real-world scenarios
- [Advanced Features](07-advanced.md): Deep dive into ignore patterns and policies

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

