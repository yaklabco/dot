# Command Reference

Complete reference for all dot commands and options.

## Command Structure

```bash
dot [global-options] <command> [command-options] [arguments]
```

**Components**:
- `global-options`: Flags affecting all commands
- `command`: Operation to perform (manage, status, etc.)
- `command-options`: Flags specific to command
- `arguments`: Command-specific arguments (package names, files, etc.)

## Global Options

Available for all commands.

### Directory Options

#### `-d, --dir PATH`

Specify package directory (source directory containing packages).

**Default**: Current directory  
**Example**:
```bash
dot --dir ~/dotfiles manage vim
dot -d /opt/configs status
```

#### `-t, --target PATH`

Specify target directory (destination for symlinks).

**Default**: `$HOME`  
**Example**:
```bash
dot --target ~ manage vim
dot -t /home/user unmanage zsh
```

### Execution Mode Options

#### `-n, --dry-run`

Preview operations without applying changes.

**Example**:
```bash
dot --dry-run manage vim
dot -n unmanage zsh
```

Shows planned operations with no filesystem modifications.

#### `--quiet`

Suppress non-error output.

**Example**:
```bash
dot --quiet manage vim
```

Only errors printed. Useful for scripting.

### Verbosity Options

#### `-v, --verbose`

Increase verbosity (repeatable).

**Levels**:
- No flag: Errors and warnings
- `-v`: Info messages
- `-vv`: Debug messages
- `-vvv`: Trace messages

**Example**:
```bash
dot -v manage vim      # Info level
dot -vv status         # Debug level
dot -vvv remanage zsh  # Trace level
```

### Output Format Options

#### `--log-json`

Output logs in JSON format.

**Example**:
```bash
dot --log-json manage vim
```

JSON output for log aggregation and parsing.

#### `--color WHEN`

Control color output.

**Values**: `auto`, `always`, `never`  
**Default**: `auto`  
**Example**:
```bash
dot --color always status
dot --color never list
```

### Link Options

#### `--absolute`

Create absolute symlinks instead of relative.

**Example**:
```bash
dot --absolute manage vim
```

#### `--no-folding`

Disable directory folding optimization.

**Example**:
```bash
dot --no-folding manage vim
```

Creates per-file links instead of directory links.

### Ignore Options

#### `--ignore PATTERN`

Add ignore pattern (repeatable).

**Example**:
```bash
dot --ignore "*.log" manage vim
dot --ignore "*.log" --ignore "*.tmp" manage zsh
```

#### `--override PATTERN`

Force include pattern despite ignore rules (repeatable).

**Example**:
```bash
dot --override ".gitignore" manage git
```

### Conflict Resolution Options

#### `--on-conflict POLICY`

Set conflict resolution policy.

**Values**: `fail`, `backup`, `overwrite`, `skip`  
**Default**: `fail`  
**Example**:
```bash
dot --on-conflict backup manage vim
dot --on-conflict skip manage zsh
```

## Package Management Commands

### clone

Clone a dotfiles repository and install packages.

**Synopsis**:
```bash
dot clone [options] REPOSITORY_URL
```

**Arguments**:
- `REPOSITORY_URL`: Git repository URL (HTTPS or SSH format)

**Options**:
- `--profile NAME`: Installation profile from bootstrap config
- `--interactive`: Interactively select packages to install
- `--force`: Overwrite package directory if exists
- `--branch NAME`: Branch to clone (defaults to repository default)

All global options also apply.

**Description**:

The clone command provides single-command setup for new machines. It clones a dotfiles repository and installs packages based on optional bootstrap configuration.

Like `git clone`, the repository is cloned into a subdirectory named after the repository. For example, `dot clone https://github.com/user/my-dotfiles` creates a `my-dotfiles` directory in the current location. Use `--dir` to specify a different target directory.

**Workflow**:
1. Determines target directory (from repository name or `--dir` flag)
2. Validates target directory is empty (unless `--force`)
3. Clones repository to target directory
4. Loads optional `.dotbootstrap.yaml` configuration
5. Selects packages (via profile, interactively, or all)
6. Filters packages by current platform
7. Installs selected packages via `manage` command
8. Updates manifest with repository tracking information

**Authentication**:

Authentication is automatically resolved in priority order:
1. `GITHUB_TOKEN` environment variable (GitHub repositories)
2. `GIT_TOKEN` environment variable (general git repositories)
3. SSH keys in `~/.ssh/` directory (for SSH URLs like `git@github.com:user/repo.git`)
4. GitHub CLI (`gh`) authenticated session (for HTTPS GitHub repositories)
5. No authentication (public repositories only)

If you've authenticated with `gh auth login`, dot will automatically use your GitHub CLI credentials when cloning private GitHub repositories via HTTPS. For SSH URLs, SSH keys are preferred as expected.

**Bootstrap Configuration**:

If `.dotbootstrap.yaml` exists in repository root, it defines:
- Available packages with platform requirements
- Named installation profiles
- Default profile and conflict resolution policies

Without bootstrap configuration, all discovered packages are offered for installation.

See [Bootstrap Configuration Specification](bootstrap-config-spec.md) for complete documentation.

**Examples**:

```bash
# Clone and install all packages (creates ./dotfiles directory)
dot clone https://github.com/user/dotfiles

# Clone creates ./my-dotfiles directory based on repo name
dot clone https://github.com/user/my-dotfiles

# Clone specific branch
dot clone https://github.com/user/dotfiles --branch develop

# Use named profile from bootstrap config
dot clone https://github.com/user/dotfiles --profile minimal

# Force interactive selection
dot clone https://github.com/user/dotfiles --interactive

# Clone to specific directory (overrides default behavior)
dot clone --dir ~/my-packages https://github.com/user/dotfiles

# Overwrite existing package directory
dot clone --force https://github.com/user/dotfiles

# Clone via SSH
dot clone git@github.com:user/dotfiles.git

# Preview what would be installed
dot --dry-run clone https://github.com/user/dotfiles
```

**Error Handling**:

Common errors and solutions:

- **Package directory not empty**: Use `--force` to overwrite
- **Authentication failed**: Set `GITHUB_TOKEN` or configure SSH keys
- **Clone failed**: Verify URL, network connection, and repository access
- **Bootstrap invalid**: Check `.dotbootstrap.yaml` syntax
- **Profile not found**: Verify profile exists in bootstrap config

**Platform Filtering**:

Packages in bootstrap configuration can specify target platforms:

```yaml
packages:
  - name: dot-vim           # All platforms
  - name: dot-linux-config  # Linux only
    platform: [linux]
  - name: dot-macos-config  # macOS only
    platform: [darwin]
```

Platform filtering is automatic based on current system.

**Related Commands**:
- `manage`: Manually install additional packages after cloning
- `status`: Check installation status and repository information
- `unmanage`: Remove installed packages
- `clone bootstrap`: Generate bootstrap configuration from installation

### clone bootstrap

Generate bootstrap configuration from existing dotfiles installation.

**Synopsis**:
```bash
dot clone bootstrap [options]
```

**Options**:
- `-o, --output PATH`: Output file path (default: `.dotbootstrap.yaml` in package directory)
- `--dry-run`: Print configuration to stdout instead of writing file
- `--from-manifest`: Only include packages currently in manifest
- `--conflict-policy POLICY`: Default conflict policy (backup, fail, overwrite, skip)
- `--force`: Overwrite existing bootstrap file

All global options also apply.

**Description**:

The clone bootstrap subcommand generates a `.dotbootstrap.yaml` configuration file from your current dotfiles installation. This allows you to create a bootstrap configuration for an existing repository, enabling others to clone your dotfiles with predefined package selections and profiles.

The command discovers all packages in the package directory and creates a bootstrap configuration with sensible defaults. The generated file includes helpful comments and example structures that you should review and customize before committing.

**Generated Configuration**:

The output includes:
- All discovered packages marked as `required: false`
- Default conflict resolution policy
- Example profile structures with comments
- Helpful guidance for customization
- Timestamps and documentation links

**Examples**:

```bash
# Generate bootstrap config in package directory
dot clone bootstrap

# Specify custom output location
dot clone bootstrap --output ~/dotfiles/.dotbootstrap.yaml

# Preview without writing file
dot clone bootstrap --dry-run

# Only include packages from manifest
dot clone bootstrap --from-manifest

# Set default conflict policy
dot clone bootstrap --conflict-policy backup

# Overwrite existing file
dot clone bootstrap --force
```

**Workflow**:

1. Run command in dotfiles repository
2. Review generated `.dotbootstrap.yaml`
3. Customize package requirements and platform restrictions
4. Define installation profiles for different use cases
5. Commit configuration to repository
6. Others can clone with `--profile` flag

**Error Handling**:

Common errors and solutions:

- **No packages found**: Ensure package directory contains subdirectories
- **Bootstrap file exists**: Use `--force` to overwrite
- **Invalid conflict policy**: Use backup, fail, overwrite, or skip
- **Package directory not found**: Check global `--dir` flag

**Customization After Generation**:

After generating the configuration, customize it by:

1. **Mark required packages**: Set `required: true` for essential packages
2. **Add platform restrictions**: Specify `platform: [linux, darwin]` as needed
3. **Create profiles**: Define named sets like `minimal`, `work`, `full`
4. **Set conflict policies**: Override default per-package if needed
5. **Add descriptions**: Document profiles for other users

Example customization:

```yaml
version: "1.0"

defaults:
  on_conflict: backup
  profile: minimal

packages:
  - name: vim
    required: true  # Essential package
  - name: zsh
    required: true
  - name: macos-config
    required: false
    platform: [darwin]  # macOS only

profiles:
  minimal:
    description: Basic shell and editor
    packages:
      - vim
      - zsh
  full:
    description: Complete development environment
    packages:
      - vim
      - zsh
      - git
      - tmux
```

**Related Commands**:
- `clone`: Clone repository using bootstrap configuration
- `status`: View currently installed packages
- `list`: List all installed packages

See [Bootstrap Configuration Specification](bootstrap-config-spec.md) for complete configuration reference.

### manage

Install packages by creating symlinks.

**Synopsis**:
```bash
dot manage [options] PACKAGE [PACKAGE...]
```

**Arguments**:
- `PACKAGE`: One or more package names to install

**Options**: All global options

**Examples**:
```bash
# Single package
dot manage vim

# Multiple packages
dot manage vim zsh tmux git

# With options
dot --no-folding manage vim
dot --absolute manage configs
dot --dry-run manage test-package

# Different directories
dot --dir ~/dotfiles --target ~ manage vim
```

**Behavior**:
1. Scans package directories
2. Computes desired symlink state
3. Detects conflicts
4. Resolves conflicts per policy
5. Creates symlinks with dependency ordering
6. Updates manifest

**Exit Codes**:
- `0`: Success
- `1`: Error during operation
- `2`: Invalid arguments
- `3`: Conflicts detected (with fail policy)
- `4`: Permission denied

### unmanage

Remove packages by deleting symlinks, with optional restoration or cleanup.

**Synopsis**:
```bash
dot unmanage [options] PACKAGE [PACKAGE...]
dot unmanage --all [options]
```

**Arguments**:
- `PACKAGE`: One or more package names to remove

**Options**:
- All global options
- `--all`: Remove all managed packages
- `--yes, --force`: Skip confirmation prompt (for use with --all)
- `--purge`: Delete package directory after removing links
- `--no-restore`: Skip restoring adopted packages to target
- `--cleanup`: Remove orphaned packages from manifest only

**Examples**:
```bash
# Remove managed package (removes links only)
dot unmanage vim

# Remove adopted package (restores files to target, keeps in package)
dot unmanage dot-ssh

# Remove with purge (deletes package directory)
dot unmanage --purge vim

# Remove without restoring (for adopted packages)
dot unmanage --no-restore dot-ssh

# Clean up orphaned packages
dot unmanage --cleanup dot-old-package

# Remove all packages (with confirmation prompt)
dot unmanage --all

# Remove all packages without confirmation
dot unmanage --all --yes

# Remove all packages and delete directories
dot unmanage --all --purge --force

# Preview removing all packages
dot --dry-run unmanage --all
```

**Behavior**:

For **managed packages** (created with `dot manage`):
1. Removes symlinks
2. Cleans up empty directories
3. Removes from manifest
4. Package directory preserved (unless `--purge`)

For **adopted packages** (created with `dot adopt`):
1. Removes symlinks
2. **Copies files back to target** (unless `--no-restore`)
3. Removes from manifest  
4. Package directory preserved (unless `--purge`)

**Restoration for Adopted Packages**:

By default, `unmanage` **restores** adopted files to their original locations:

```bash
# Before unmanage:
~/.ssh -> ~/dotfiles/dot-ssh  # Symlink
~/dotfiles/dot-ssh/config     # Files in package

# After: dot unmanage dot-ssh
~/.ssh/config                 # Files restored (copied back)
~/dotfiles/dot-ssh/config     # Package preserved as backup
```

Files are **copied** (not moved), so they remain in the package as a backup.

**Remove All Packages**:

Use `--all` to remove all managed packages at once:

```bash
# With confirmation prompt
dot unmanage --all

# Skip confirmation
dot unmanage --all --yes
```

When using `--all`:
1. Shows summary of all packages to be removed
2. Displays operation type for each (remove, restore, purge)
3. Requires confirmation unless `--yes`, `--force`, or `--dry-run` specified
4. Applies same behavior as individual unmanage (restore adopted by default)

This is useful for completely resetting your system to pre-dot state.

**Cleanup Mode**:

Use `--cleanup` to remove orphaned packages (missing links or directories):

```bash
dot unmanage --cleanup old-package
```

Only updates manifest, no filesystem operations.

**Safety Guarantees**:
- Only removes links pointing to package directory
- Preserves non-managed files
- Validates link targets before deletion
- Adopted packages restored by default (preserves your data)
- Confirmation required for `--all` operations

**Exit Codes**:
- `0`: Success
- `1`: Error during operation
- `5`: Package not found/not installed

### remanage

Update packages efficiently using incremental detection and restore missing symlinks.

**Synopsis**:
```bash
dot remanage [options] PACKAGE [PACKAGE...]
```

**Arguments**:
- `PACKAGE`: One or more package names to update

**Options**: All global options

**Examples**:
```bash
# Single package
dot remanage vim

# Multiple packages
dot remanage vim zsh tmux

# Preview changes
dot --dry-run remanage vim

# Verbose output to see detection details
dot -vv remanage zsh
```

**Behavior**:
1. Loads manifest with previous state
2. Computes content hashes for package directories
3. Verifies all symlinks still exist
4. Compares with stored hashes and link states
5. Processes changed or broken packages
6. Updates manifest while preserving package source type

**Incremental Detection**:
- **Unchanged packages with valid links**: Skipped entirely (no-op)
- **Changed packages**: Unmanaged then managed (full update)
- **Packages with missing links**: Recreates missing symlinks
- **New packages**: Managed
- **Adopted packages**: Preserves adoption structure (single directory symlink)

**Missing Link Detection**:

If symlinks were accidentally deleted, `remanage` automatically recreates them:

```bash
# Symlink accidentally deleted
rm ~/.vimrc

# Check status
dot doctor
# ✗ error: .vimrc link does not exist

# Recreate missing link
dot remanage vim
# Successfully remanaged 1 package(s)

# Link restored
ls -la ~/.vimrc
# ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
```

**Package Source Preservation**:

`remanage` preserves the original package type:
- **Adopted packages**: Maintains single directory symlink structure
- **Managed packages**: Maintains individual file symlinks

This ensures adopted directories aren't converted to managed packages.

**Exit Codes**:
- `0`: Success, changes applied or no changes needed
- `1`: Error during operation

### adopt

Move existing files or directories into a package and create symlinks.

**Synopsis**:
```bash
# Auto-naming mode (single file/directory)
dot adopt [options] FILE|DIRECTORY

# Glob expansion mode (multiple files with common prefix)
dot adopt [options] PATTERN...

# Explicit package mode
dot adopt [options] PACKAGE FILE|DIRECTORY [FILE|DIRECTORY...]
```

**Arguments**:
- `FILE|DIRECTORY`: Path to file or directory to adopt
- `PACKAGE`: Explicit package name (optional)
- `PATTERN`: Shell glob pattern (e.g., `.git*`)

**Options**: All global options

**Modes**:

#### Auto-Naming Mode
Single file or directory - package name derived automatically:
```bash
dot adopt .vimrc      # Creates package: dot-vimrc
dot adopt .ssh        # Creates package: dot-ssh
dot adopt .config     # Creates package: dot-config
```

#### Glob Expansion Mode
Multiple files with common prefix - package name derived from prefix:
```bash
dot adopt .git*       # Expands to .gitconfig, .gitignore, etc.
                      # Creates package: dot-git
                      # All files adopted into single package

dot adopt .vim*       # Expands to .vimrc, .viminfo, etc.
                      # Creates package: dot-vim
```

#### Explicit Package Mode
Specify package name explicitly:
```bash
dot adopt vim .vimrc .vim/          # Package: vim
dot adopt configs .config/ .local/  # Package: configs
```

#### Path Resolution

File paths are resolved based on the following rules:

1. **Absolute paths** (`/etc/config`, `~/file`): Used as-is
2. **Explicit relative paths** (`./file`, `../dir`): Resolved from current working directory
3. **Bare paths** (`file`, `.config/nvim`): Resolved from target directory (default: `$HOME`)

**Examples:**

```bash
# From ~/.config directory:
cd ~/.config
dot adopt ado-cli ./ado-cli        # Adopts ~/.config/ado-cli (from pwd)
dot adopt fish .config/fish        # Adopts $HOME/.config/fish (from target)

# Using explicit pwd paths:
cd ~/.config
dot adopt nvim ./nvim              # Adopts ~/.config/nvim
dot adopt configs ./fish ./nvim    # Adopts multiple from pwd

# Backward compatible - bare paths from target:
cd /tmp
dot adopt .vimrc                   # Adopts $HOME/.vimrc (not /tmp/.vimrc)
```

**Note:** The `./` prefix explicitly means "from current directory", while bare paths maintain backward compatibility by resolving from the target directory.

**Directory Adoption**:

When adopting a directory, `dot` creates a **flat structure** in the package with the directory contents at the package root:

```bash
# Before: ~/.ssh/ with files
~/.ssh/
├── config
├── id_rsa
└── known_hosts

# After: dot adopt .ssh
~/dotfiles/dot-ssh/       # Package root contains directory contents
├── config
├── id_rsa
└── known_hosts

~/.ssh -> ~/dotfiles/dot-ssh  # Single symlink to package root
```

**File Adoption**:

Single files are placed in a package directory with dotfile translation:

```bash
# Before: ~/.vimrc

# After: dot adopt .vimrc
~/dotfiles/dot-vimrc/
└── dot-vimrc

~/.vimrc -> ~/dotfiles/dot-vimrc/dot-vimrc
```

**Dotfile Translation**:

Dotfiles (starting with `.`) have the dot replaced with `dot-` prefix:
- `.vimrc` → `dot-vimrc`
- `.ssh` → `dot-ssh`
- `.config` → `dot-config`
- Nested: `.config/nvim/init.vim` → `dot-config/nvim/init.vim`

**Behavior**:
1. Determines adoption mode (auto-naming, glob, or explicit)
2. Derives or uses provided package name
3. Creates package directory structure
4. Moves files/directories to package (applying dotfile translation)
5. Creates symlinks in original locations
6. Records package as "adopted" in manifest

**Exit Codes**:
- `0`: Success
- `1`: Error during operation
- `2`: Invalid arguments
- `4`: Permission denied

## Query Commands

### status

Display installation status for packages.

**Synopsis**:
```bash
dot status [options] [PACKAGE...]
```

**Arguments**:
- `PACKAGE` (optional): Specific packages to query (default: all)

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`)
- All global options

**Examples**:
```bash
# All packages
dot status

# Specific packages
dot status vim zsh

# JSON output
dot status --format json

# YAML output
dot status --format yaml

# Table format
dot status --format table

# Combine with verbosity
dot -v status vim
```

**Output Fields**:
- Package name
- Installation status
- Link count
- Installation date
- List of symlinks
- Conflicts or issues

**Example Output (text)**:
```
Package: vim
  Status: installed
  Links: 3
  Installed: 2025-10-07 10:30:00
  
  Links:
    ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
    ~/.vim/colors/ -> ~/dotfiles/vim/dot-vim/colors/
    ~/.vim/autoload/ -> ~/dotfiles/vim/dot-vim/autoload/
```

**Example Output (JSON)**:
```json
{
  "packages": [
    {
      "name": "vim",
      "status": "installed",
      "link_count": 3,
      "installed_at": "2025-10-07T10:30:00Z",
      "links": [
        {
          "target": "~/.vimrc",
          "source": "~/dotfiles/vim/dot-vimrc",
          "type": "file"
        }
      ]
    }
  ]
}
```

**Exit Codes**:
- `0`: Success
- `1`: Error querying status

### doctor

Validate installation health and detect issues.

**Synopsis**:
```bash
dot doctor [options]
```

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`)
- `--scan-mode MODE`: Orphaned link detection mode (`off`, `scoped`, `deep`) (default: `scoped`)
- `--color MODE`: Color output mode (`auto`, `always`, `never`) (default: `auto`)
- All global options

**Scan Modes**:

- **off**: Skip orphaned link detection (fastest, ~50ms)
  - Only checks managed links from manifest
  - Use for quick health checks or in automated scripts

- **scoped** (default): Scan directories containing managed links (fast, ~600ms)
  - Limited to depth 3 to avoid deep recursion
  - Skips common large directories (Library, node_modules, .docker, etc.)
  - Parallel scanning using multiple CPU cores
  - Recommended for regular health checks

- **deep**: Full recursive scan of target directory (thorough, ~3-5s)
  - Scans entire home directory up to depth 10
  - Still skips large cache/build directories
  - Use when investigating orphaned links from other tools
  - Significantly slower but more comprehensive

**Performance Notes**:

The doctor command has been optimized for speed:
- Parallel directory scanning using worker pools
- DirEntry type checking (no extra syscalls for regular files)
- Intelligent skip patterns for common large directories
- Depth limits to prevent excessive recursion

For systems with many symlinks (10,000+), use `scoped` mode for regular checks
and `deep` mode only when investigating specific issues.

**Examples**:
```bash
# Basic health check (scoped scan - default, fast)
dot doctor

# Quick check without orphan detection (fastest)
dot doctor --scan-mode=off

# Deep scan for comprehensive orphan detection
dot doctor --scan-mode=deep

# Detailed output with verbose logging
dot -v doctor

# JSON output for scripting
dot doctor --format json

# Table format
dot doctor --format table

# Force color output even when piped
dot doctor --color=always | less -R
```

**Checks Performed**:
1. **Broken symlinks**: Links pointing to non-existent targets
2. **Orphaned links**: Links not in manifest but pointing to package directory
3. **Wrong links**: Links in manifest but pointing elsewhere
4. **Manifest consistency**: Manifest matches filesystem state
5. **Permission issues**: Files with incorrect permissions
6. **Circular dependencies**: Circular symlink chains

**Example Output (healthy)**:
```
Running health checks...

✓ All symlinks valid
✓ No broken links
✓ No orphaned links
✓ Manifest consistent
✓ No permission issues

Health check passed: 0 issues found
```

**Example Output (issues)**:
```
Running health checks...

✗ Broken links found: 2
  ~/.vimrc -> ~/dotfiles/vim/dot-vimrc (target missing)
  ~/.zshrc -> ~/dotfiles/zsh/dot-zshrc (target missing)

✗ Orphaned links found: 1
  ~/.bashrc -> ~/old-dotfiles/bash/bashrc

Suggestions:
  - Remove broken links: dot doctor --fix-broken
  - Adopt orphaned links: dot adopt bash ~/.bashrc
  - Reinstall packages: dot remanage vim zsh

Health check failed: 3 issues found
```

**Exit Codes**:
- `0`: No issues found
- `1`: Issues detected
- `2`: Invalid arguments

### list

Show installed package inventory with health status indicators.

**Synopsis**:
```bash
dot list [options]
```

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`)
- `-s, --sort FIELD`: Sort by field (`name`, `links`, `date`)
- All global options

**Health Status**:

Each package is automatically checked for health when listing. A package is considered healthy if all its managed symlinks exist and point to their correct targets. Health indicators:

- `✓` (green checkmark): All symlinks are valid
- `✗` (red X): Package has issues with specific type indicated (e.g., "broken links", "wrong target", "missing links")

The health check is fast and only validates symlink existence and targets, without the full diagnostic scan that `doctor` performs.

**Examples**:
```bash
# List all packages with health status
dot list

# Sort by link count
dot list --sort links

# Sort by installation date
dot list --sort date

# JSON output (includes health status)
dot list --format json

# Table format with health column
dot list --format table

# Combine sorting and format
dot list --sort links --format table
```

**Example Output (text)**:
```
Packages: 3 packages in /home/user/dotfiles

✓  vim    (3 links)  installed 2 hours ago
✗  zsh    (2 links)  broken links  installed 1 day ago
✓  tmux   (1 link)   installed 3 days ago

2 healthy, 1 unhealthy
```

**Example Output (table)**:
```
Health          Package  Links  Installed
✓               vim      3      2 hours ago
✗ broken links  zsh      2      1 day ago
✓               tmux     1      3 days ago

2 healthy, 1 unhealthy
```

**Example Output (JSON)**:
```json
[
  {
    "name": "vim",
    "link_count": 3,
    "installed_at": "2025-10-07T10:30:00Z",
    "is_healthy": true,
    "issue_type": ""
  },
  {
    "name": "zsh",
    "link_count": 2,
    "installed_at": "2025-10-07T10:31:00Z",
    "is_healthy": false,
    "issue_type": "broken links"
  }
]
```

**Issue Types**:
- `broken links`: Symlinks point to non-existent targets
- `wrong target`: Symlinks point to unexpected locations outside the package directory
- `missing links`: Expected symlinks do not exist

**Exit Codes**:
- `0`: Success
- `1`: Error listing packages

## Utility Commands

### version

Display version information.

**Synopsis**:
```bash
dot version [options]
```

**Options**:
- `--short`: Show version number only
- All global options

**Examples**:
```bash
# Full version info
dot version

# Short version
dot version --short

# Alternative using flag
dot --version
```

**Example Output**:
```
dot version v0.1.0
Built with Go 1.25.4
Commit: abc1234
Build date: 2025-10-07
Platform: linux/amd64
```

### help

Display help information.

**Synopsis**:
```bash
dot help [COMMAND]
```

**Arguments**:
- `COMMAND` (optional): Show help for specific command

**Examples**:
```bash
# General help
dot help

# Command-specific help
dot help manage
dot help status

# Alternative using flag
dot --help
dot manage --help
```

### completion

Generate shell completion script.

**Synopsis**:
```bash
dot completion SHELL
```

**Arguments**:
- `SHELL`: Shell type (`bash`, `zsh`, `fish`, `powershell`)

**Examples**:
```bash
# Bash
dot completion bash > /etc/bash_completion.d/dot

# Zsh
dot completion zsh > "${fpath[1]}/_dot"

# Fish
dot completion fish > ~/.config/fish/completions/dot.fish

# PowerShell
dot completion powershell > dot.ps1
```

## Exit Codes

Standard exit codes across all commands:

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Operation completed successfully |
| 1 | General error | Unspecified error occurred |
| 2 | Invalid arguments | Command-line arguments invalid |
| 3 | Conflicts detected | Conflicts found (with fail policy) |
| 4 | Permission denied | Insufficient permissions |
| 5 | Package not found | Specified package does not exist |

**Usage in Scripts**:
```bash
#!/bin/bash

# Check if operation succeeded
if dot manage vim; then
    echo "vim installed successfully"
else
    exit_code=$?
    case $exit_code in
        3) echo "Conflicts detected" ;;
        4) echo "Permission denied" ;;
        5) echo "Package not found" ;;
        *) echo "Error: $exit_code" ;;
    esac
    exit 1
fi
```

## Command Patterns

### Dry Run Pattern

Preview before applying:

```bash
# Always preview first
dot --dry-run manage vim

# Review output, then apply
dot manage vim
```

### Verbose Debugging Pattern

Debug issues with verbose output:

```bash
# Increase verbosity to see details
dot -vvv manage vim

# Or use with doctor
dot -vv doctor
```

### Scripting Pattern

Quiet mode with JSON output:

```bash
#!/bin/bash

# Run command quietly
output=$(dot --quiet --log-json manage vim 2>&1)

# Parse JSON output
if [ $? -eq 0 ]; then
    echo "Success"
else
    echo "$output" | jq '.error'
fi
```

### Batch Operations Pattern

Manage multiple packages from list:

```bash
# From file
cat packages.txt | xargs dot manage

# From array
packages=(vim zsh tmux git)
dot manage "${packages[@]}"

# With error checking
for pkg in vim zsh tmux; do
    if ! dot manage "$pkg"; then
        echo "Failed to manage: $pkg"
    fi
done
```

## Command Aliases

No built-in aliases, but shell aliases recommended:

```bash
# Common aliases
alias dm='dot manage'
alias du='dot unmanage'
alias dr='dot remanage'
alias ds='dot status'
alias dl='dot list'
alias dd='dot doctor'

# With default options
alias dot-dry='dot --dry-run'
alias dot-verbose='dot -vv'
```

Add to `~/.bashrc`, `~/.zshrc`, or equivalent.

## Next Steps

- [Common Workflows](06-workflows.md): See commands in real-world scenarios
- [Advanced Features](07-advanced.md): Deep dive into options and features
- [Troubleshooting Guide](08-troubleshooting.md): Solve common issues

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

