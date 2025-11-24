# Ignore System

The dot ignore system provides flexible control over which files are included when managing packages. This system supports global configuration, per-package rules, and size-based filtering.

## Overview

Ignore patterns work similarly to `.gitignore`, allowing you to exclude files from being symlinked. The system supports:

- Default ignore patterns for common system files
- Custom global patterns via configuration or flags
- Per-package `.dotignore` files
- Negation patterns to un-ignore files
- Size-based filtering for large files

## Pattern Syntax

### Glob Patterns

Patterns use glob syntax:

- `*` - Matches any sequence of characters
- `?` - Matches any single character
- `*.ext` - Matches all files with extension
- `dirname/` - Matches directory and contents

### Negation Patterns

Patterns starting with `!` un-ignore previously ignored files:

```
*.log          # Ignore all .log files
!important.log # But include important.log
```

**Important**: Order matters. Patterns are processed sequentially, and the last matching pattern wins.

### Examples

```
# Ignore all temporary files
*.tmp
*.swp
*~

# But keep backup files
!*.bak

# Ignore cache directories
.cache/
node_modules/

# Ignore large files
*.qcow2
*.vmdk
*.iso
```

## Configuration

### Global Configuration

Configure ignore settings in your `config.yaml`:

```yaml
ignore:
  # Use default patterns (.git, .DS_Store, etc.)
  use_defaults: true
  
  # Additional patterns to ignore
  patterns:
    - "*.log"
    - "*.tmp"
    - "!important.log"  # Negation
  
  # Enable per-package .dotignore files
  per_package_ignore: true
  
  # Maximum file size in bytes (0 = no limit)
  max_file_size: 104857600  # 100MB
  
  # Prompt for large files in interactive mode
  interactive_large_files: true
```

### Command-Line Flags

Override configuration with flags:

```bash
# Add ignore patterns
dot manage mypackage --ignore "*.qcow2" --ignore "*.vmdk"

# Set size limit
dot manage mypackage --max-file-size 100MB

# Disable default patterns
dot manage mypackage --no-defaults

# Disable .dotignore files
dot manage mypackage --no-dotignore

# Batch mode (non-interactive, auto-skip large files)
dot manage mypackage --batch --max-file-size 50MB
```

### Size Format

File sizes support human-readable formats:

- `B` or `b` - Bytes
- `KB`, `K`, `k` - Kilobytes (1024 bytes)
- `MB`, `M`, `m` - Megabytes (1024 KB)
- `GB`, `G`, `g` - Gigabytes (1024 MB)
- `TB`, `T`, `t` - Terabytes (1024 GB)

Examples: `100MB`, `1.5GB`, `500M`, `1024KB`

## Per-Package .dotignore Files

Create a `.dotignore` file in any package directory:

```
# .dotignore in colima package
# Ignore VM disk images
*.qcow2
*.vmdk

# But keep configuration backups
!*.qcow2.backup

# Ignore logs
logs/*.log

# Comments are supported
# Empty lines are ignored
```

### Inheritance

`.dotignore` files support inheritance from parent directories, similar to `.gitignore`:

```
packages/
├── .dotignore           # Applies to all packages
├── vim/
│   └── .dotignore       # Applies to vim and subdirs
└── colima/
    └── .dotignore       # Applies to colima
```

Files closer to the root have lower priority. Child `.dotignore` files can override parent patterns using negation.

## Default Ignore Patterns

When `use_defaults: true`, these patterns are automatically applied:

- `.git` - Git repository metadata
- `.svn` - Subversion metadata
- `.hg` - Mercurial metadata
- `.DS_Store` - macOS metadata
- `Thumbs.db` - Windows thumbnails
- `desktop.ini` - Windows folder settings
- `.Trash` - Trash directory
- `.Spotlight-V100` - macOS Spotlight index
- `.TemporaryItems` - macOS temporary items

## Size-Based Filtering

### Interactive Mode

When a file exceeds the size limit in TTY mode with `interactive_large_files: true`:

```
Large file detected:
  Path: .colima/default/diffdisk.qcow2
  Size: 2.5 GB (limit: 100.0 MB)

Options:
  i) Include this file
  s) Skip this file
  a) Skip all large files
Choice [s]:
```

### Batch Mode

In batch mode (`--batch` flag or non-TTY), large files are automatically skipped silently.

## Common Use Cases

### Ignore Virtual Machine Images

For virtualization tools like Colima:

```yaml
# In config.yaml
ignore:
  patterns:
    - "*.qcow2"
    - "*.vmdk"
    - "*.vdi"
  max_file_size: 104857600  # 100MB
```

Or in package `.dotignore`:

```
# .dotfiles/colima/.dotignore
*.qcow2
*.vmdk
diffdisk.*
```

### Ignore Cache and Build Artifacts

```yaml
ignore:
  patterns:
    - ".cache/"
    - "node_modules/"
    - "*.pyc"
    - "__pycache__/"
    - "*.o"
    - "*.so"
```

### Development with Selective Inclusion

```
# Ignore all logs
*.log

# But keep important logs
!error.log
!access.log

# Ignore compiled files
*.o
*.so
*.a

# But keep specific libraries
!libimportant.so
```

## Precedence Order

Configuration sources are applied in this order (later overrides earlier):

1. Default ignore patterns (if enabled)
2. Global config file patterns
3. Per-package `.dotignore` files (parent to child)
4. Command-line `--ignore` flags

Within each source, patterns are processed sequentially, with later patterns overriding earlier ones.

## Best Practices

1. **Use Defaults**: Keep `use_defaults: true` to automatically ignore system files
2. **Size Limits**: Set reasonable limits for package types (e.g., 100MB for configs, 1GB for application data)
3. **Per-Package Rules**: Use `.dotignore` for package-specific exclusions
4. **Negation Sparingly**: Use negation patterns when you need exceptions, but keep rules simple
5. **Document Patterns**: Add comments to explain non-obvious patterns
6. **Test Patterns**: Use `dot manage --dry-run` to preview what will be managed

## Troubleshooting

### Files Not Being Ignored

1. Check pattern syntax - remember `*` matches within directories
2. Verify pattern order - later patterns override earlier ones
3. Check if file matches default patterns
4. Use absolute paths if relative matching fails

### Files Unexpectedly Ignored

1. Check for overly broad patterns
2. Look for default patterns that might match
3. Check parent `.dotignore` files
4. Use negation to explicitly include files

### Size Filtering Issues

1. Verify size format is correct (e.g., `100MB` not `100 MB`)
2. Check if `max_file_size` is set to 0 (disabled)
3. Ensure `interactive_large_files` matches your use case
4. Use `--batch` for non-interactive environments

## Environment Variables

All ignore configuration can be set via environment variables:

```bash
export DOT_IGNORE_USE_DEFAULTS=true
export DOT_IGNORE_PATTERNS="*.log,*.tmp"
export DOT_IGNORE_MAX_FILE_SIZE=104857600
export DOT_IGNORE_PER_PACKAGE_IGNORE=true
export DOT_IGNORE_INTERACTIVE_LARGE_FILES=true
```

## See Also

- [Configuration Guide](04-configuration.md)
- [Commands Reference](05-commands.md)
- [Advanced Usage](07-advanced.md)



