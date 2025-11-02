# Introduction and Core Concepts

## What is dot?

dot is a symbolic link manager for configuration files and dotfiles. It automates the creation and management of symlinks between a package repository (package directory) and installation location (target directory), enabling centralized configuration management with version control.

### Primary Use Cases

1. **Dotfile Management**: Centralize shell configurations, editor settings, and application preferences in a version-controlled repository
2. **Multi-Machine Synchronization**: Maintain consistent configurations across multiple systems
3. **Configuration Isolation**: Organize configurations by application or purpose in separate packages
4. **Safe Updates**: Apply configuration changes with transactional safety and rollback capability

## Core Concepts

### Stow Directory

The **package directory** is the source directory containing one or more packages. Each subdirectory within the package directory represents a package.

Default location: Current working directory

Example structure:
```
~/dotfiles/              # Stow directory
├── vim/                 # Package
│   ├── dot-vimrc
│   └── dot-vim/
│       └── colors/
├── zsh/                 # Package
│   ├── dot-zshrc
│   └── dot-zshenv
└── tmux/                # Package
    └── dot-tmux.conf
```

### Target Directory

The **target directory** is the destination where symlinks are created, typically the user's home directory.

Default location: `$HOME`

After managing packages:
```
~/                       # Target directory
├── .vimrc -> ~/dotfiles/vim/dot-vimrc
├── .vim/ -> ~/dotfiles/vim/dot-vim/
├── .zshrc -> ~/dotfiles/zsh/dot-zshrc
├── .zshenv -> ~/dotfiles/zsh/dot-zshenv
└── .tmux.conf -> ~/dotfiles/tmux/dot-tmux.conf
```

### Package

A **package** is a directory within the package directory containing related configuration files. The package structure mirrors the desired structure in the target directory.

Package characteristics:
- Self-contained collection of related files
- Structure matches target directory layout
- Can contain files and directories
- May include package-specific metadata

Example package structure:
```
vim/                     # Package name
├── dot-vimrc            # → ~/.vimrc
├── dot-vim/             # → ~/.vim/
│   ├── colors/
│   │   └── theme.vim
│   └── autoload/
│       └── plugin.vim
└── .dotmeta            # Optional package metadata
```

### Symlink

A **symbolic link** (symlink) is a filesystem reference that points to another file or directory. dot creates symlinks from the target directory to files in the package directory.

Link types:
- **Relative**: Links use relative paths (e.g., `../../dotfiles/vim/dot-vimrc`)
- **Absolute**: Links use absolute paths (e.g., `/home/user/dotfiles/vim/dot-vimrc`)

Default: Relative links (portable across different mount points)

### Dotfile Translation

**Dotfile translation** automatically converts file names with the `dot-` prefix to dotfiles (names starting with `.`) in the target directory.

Translation rules:
- `dot-bashrc` → `.bashrc`
- `dot-config/` → `.config/`
- `dot-local/bin/` → `.local/bin/`
- `regular-file` → `regular-file` (no translation)

Rationale: Version control systems and tools handle non-hidden files better, making dotfiles visible in repositories improves discoverability.

Example:
```
# In package directory:
vim/dot-vimrc
vim/dot-vim/colors/theme.vim

# In target directory:
~/.vimrc -> ~/dotfiles/vim/dot-vimrc
~/.vim/ -> ~/dotfiles/vim/dot-vim/
```

### Directory Folding

**Directory folding** optimizes symlink count by creating directory-level links when all contents belong to a single package.

Without folding (per-file links):
```
~/.vim/colors/theme.vim -> ~/dotfiles/vim/dot-vim/colors/theme.vim
~/.vim/autoload/plugin.vim -> ~/dotfiles/vim/dot-vim/autoload/plugin.vim
~/.vim/ftplugin/go.vim -> ~/dotfiles/vim/dot-vim/ftplugin/go.vim
```

With folding (directory link):
```
~/.vim/ -> ~/dotfiles/vim/dot-vim/
```

Folding rules:
- Applied when directory exclusively contains files from one package
- Disabled if multiple packages contribute files to the same directory
- Automatically unfolds when package conflicts arise
- Can be disabled with `--no-folding` flag

Benefits:
- Reduced symlink count
- Improved filesystem performance
- Simplified link management

### Manifest

The **manifest** is a state file (`.dot-manifest.json`) created in the target directory that tracks installed packages, their files, and content hashes.

Manifest contents:
```json
{
  "version": "1.0",
  "updated_at": "2025-10-07T10:30:00Z",
  "packages": {
    "vim": {
      "name": "vim",
      "installed_at": "2025-10-07T10:30:00Z",
      "link_count": 3,
      "links": [
        "~/.vimrc",
        "~/.vim/colors/theme.vim",
        "~/.vim/autoload/plugin.vim"
      ]
    }
  },
  "hashes": {
    "vim": "a3f2c8b4d9e1f0..."
  }
}
```

Manifest purposes:
- Fast status queries without filesystem scanning
- Incremental change detection via content hashes
- State validation and drift detection
- Package inventory tracking

### Incremental Operations

**Incremental operations** use content-based change detection to process only modified packages during updates.

Process:
1. Compute content hash for each package
2. Compare with hash stored in manifest
3. Skip packages with unchanged hashes
4. Process only packages with changes

Benefits:
- Fast updates for large package collections
- Reduced I/O operations
- Efficient restow operations

Example:
```bash
# Initial install
dot manage vim zsh tmux git  # Processes all packages

# After modifying only vim
dot remanage vim zsh tmux git  # Only processes vim, skips others
```

## dot Operations

### Manage (Install)

Create symlinks for one or more packages:

```bash
dot manage vim
dot manage vim zsh tmux
```

Operation steps:
1. Scan package contents
2. Compute desired symlink state
3. Detect conflicts with existing files
4. Resolve conflicts per policy
5. Create symlinks with dependency ordering
6. Update manifest with installed state

### Unmanage (Remove)

Remove symlinks for packages:

```bash
dot unmanage vim
```

Operation steps:
1. Load manifest to identify managed links
2. Validate link ownership
3. Remove symlinks
4. Clean up empty directories
5. Update manifest

Safety guarantees:
- Only removes links pointing to package directory
- Preserves non-managed files
- Validates link targets before deletion

### Remanage (Update)

Update packages efficiently using incremental detection:

```bash
dot remanage vim
```

Operation steps:
1. Load manifest with previous state
2. Compute content hashes for packages
3. Compare with stored hashes
4. Process only changed packages
5. Update manifest

Equivalent to unstow + stow but more efficient.

### Adopt (Import)

Move existing files into a package and replace with symlinks:

```bash
dot adopt vim ~/.vimrc ~/.vim/
```

Operation steps:
1. Validate source files exist
2. Determine target paths in package
3. Move files to package
4. Create symlinks in original locations
5. Update manifest

Use case: Bring existing unmanaged configurations under dot management.

### Status

Query installation state:

```bash
dot status
dot status vim
```

Reports:
- Installed packages
- Link counts
- Installation dates
- Conflicts or issues

### Doctor

Validate installation health:

```bash
dot doctor
```

Checks:
- Broken symlinks
- Orphaned links
- Manifest consistency
- Permission issues
- Circular dependencies

### List

Show installed package inventory:

```bash
dot list
```

Output:
- Package names
- Link counts
- Installation dates
- Sorting options

## When to Use dot

### Ideal Use Cases

1. **Version-Controlled Dotfiles**: Store configurations in Git with centralized management
2. **Multi-Machine Consistency**: Synchronize configurations across laptop, desktop, and servers
3. **Configuration Backup**: Maintain configuration history with version control
4. **Modular Setup**: Organize configurations by application or purpose
5. **Team Standardization**: Share configurations across team members
6. **Safe Experimentation**: Test configuration changes with easy rollback

### Considerations

1. **Symlink Requirements**: Target filesystem must support symbolic links
2. **Application Compatibility**: Some applications may not follow symlinks correctly
3. **Platform Differences**: Configurations may need platform-specific variations
4. **Security**: Configurations stored in package directory should have appropriate permissions

### Alternatives

When dot may not be appropriate:

1. **No Symlink Support**: Use direct file copying or templating tools
2. **Dynamic Configuration**: Applications generating configuration at runtime
3. **Per-User Variations**: Need significant per-machine customization (consider templating)
4. **Security Requirements**: Cannot use symlinks for security-sensitive configurations

## Comparison with GNU Stow

dot provides feature parity with GNU Stow plus modern enhancements:

| Feature | dot | GNU Stow |
|---------|-----|----------|
| Basic install/remove | Yes | Yes |
| Conflict detection | Yes | Yes |
| Directory folding | Yes | Yes |
| Dotfile translation | Yes | No |
| Transactional operations | Yes | No |
| Rollback on failure | Yes | No |
| Incremental updates | Yes | No |
| Content-based detection | Yes | No |
| Adopt existing files | Yes | No |
| Status queries | Yes | No |
| Health validation | Yes | No |
| Parallel execution | Yes | No |
| Type safety | Yes | No |
| Multiple output formats | Yes | No |
| Windows support | Limited | No |

Key differences:

1. **Transaction Safety**: dot uses two-phase commit with automatic rollback
2. **Performance**: Incremental detection and parallel execution
3. **State Tracking**: Manifest-based fast queries
4. **Modern Tooling**: JSON/YAML output, structured logging, metrics
5. **Type Safety**: Compile-time path safety prevents entire bug classes

Migration path: See [Migration from GNU Stow](migration-from-stow.md) for transition guide.

## Design Philosophy

### Principles

1. **Safety First**: Operations are transactional with rollback on failure
2. **Explicit Over Implicit**: Clear error messages, no silent failures
3. **Type Safety**: Leverage type system to prevent bugs at compile time
4. **Functional Core**: Pure planning logic separated from side effects
5. **Observable**: Comprehensive logging, tracing, and metrics
6. **Performance**: Efficient algorithms, parallel execution, incremental operations

### Goals

1. **Reliability**: Guarantee consistent state, never leave partial installations
2. **Clarity**: Clear error messages with actionable suggestions
3. **Efficiency**: Fast operations even with large package collections
4. **Portability**: Cross-platform support with platform-specific handling
5. **Embeddability**: Clean library API for integration into other tools

## Next Steps

- [Installation Guide](02-installation.md): Install dot on your system
- [Quick Start Tutorial](03-quickstart.md): Hands-on introduction to basic operations
- [Configuration Reference](04-configuration.md): Configure dot for your workflow
- [Command Reference](05-commands.md): Complete command documentation

For developers:
- [Architecture Overview](../developer/architecture.md): System design and structure

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

