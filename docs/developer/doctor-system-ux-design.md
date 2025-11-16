# Doctor System UX Design Specification

## Executive Summary

This document provides detailed UX specifications for the redesigned dot doctor system, based on the Customer Usage Journeys identified in `doctor-system-cujs.md`. Each CUJ includes concrete command examples, expected outputs, interactive flows, and user decision points.

The design prioritizes clarity, safety, and actionability while supporting multiple interaction modes (interactive, batch, dry-run, CI/CD).

## Design Principles

### Output Design
- **Progressive Disclosure**: Show summary first, details on request
- **Severity Signaling**: Clear visual distinction between errors/warnings/info
- **Actionable Guidance**: Every issue includes concrete next steps
- **Context Awareness**: Adapt output based on terminal capabilities and user intent

### Interaction Design
- **Safe Defaults**: Conservative defaults that prevent data loss
- **Escape Hatches**: Advanced users can override and customize
- **Consistent Patterns**: Similar problems use similar interaction patterns
- **Respectful Automation**: Batch modes available but require explicit opt-in

### Command Structure
- **Orthogonal Concerns**: Separate diagnosis from repair
- **Composable Operations**: Small focused commands that combine well
- **Predictable Behavior**: Same command always behaves same way

---

## Command Overview

### Core Commands

```bash
# Diagnostic Commands (Read-Only)
dot doctor                      # Smart health check (all packages, fast)
dot doctor --deep              # Comprehensive scan with orphan detection
dot doctor vim                 # Check specific package
dot doctor vim zsh git         # Check multiple packages
dot doctor manifest            # Validate manifest integrity
dot doctor structure           # Check repository structure compliance
dot doctor security            # Security audit (secrets, permissions)
dot doctor platform            # Platform compatibility analysis

# Repair Commands (Write Operations)
dot doctor fix                 # Interactive repair workflow
dot doctor fix --auto          # Auto-fix safe issues
dot doctor fix --dry-run       # Preview changes
dot doctor clean orphans       # Remove orphaned symlinks
dot doctor sync manifest       # Sync manifest with filesystem

# Management Commands
dot doctor ignore add <path>   # Add to ignore list
dot doctor ignore list         # Show ignored items
dot doctor ignore remove <path> # Remove from ignore list
dot doctor history             # Show doctor operation history
dot doctor undo                # Undo last doctor operation
```

### Global Flags

```bash
# Output Format
--format <text|json|table|list|compact>  # Output format
--compact                       # Compact table (fewer columns)
--issues-only                   # Show only packages with issues
--no-color                      # Disable color output

# Verbosity
--quiet, -q                     # Minimal output (summary only)
--verbose, -v                   # Detailed output (-vv, -vvv for more)

# Sorting and Filtering
--sort <status|name|links>      # Sort order
--limit <N>                     # Show only first N packages

# Pagination
--pager <auto|always|never>     # Control pagination (default: auto)
--no-pager                      # Disable pagination (alias for --pager never)

# CI/CD
--exit-zero                     # Always exit 0 (for CI)
--non-interactive               # No prompts (for scripts)
```

---

## Category A: Initial Setup and Migration

### CUJ-A1: First Time Installation

**Scenario**: User runs `dot manage vim` for the first time, but `~/.vimrc` already exists.

#### Command
```bash
$ dot manage vim
```

#### Output
```
Analyzing package vim...

Conflicts detected at 3 locations:
  ✗ ~/.vimrc
    Existing: regular file (last modified 2024-11-10)
    Wanted: symlink to ~/dotfiles/vim/dot-vimrc
    
  ✗ ~/.vim/
    Existing: directory (3 files)
    Wanted: symlink to ~/dotfiles/vim/dot-vim/
    
  ⚠ ~/.vim/colors/
    Existing: directory (5 files)
    Wanted: symlink to ~/dotfiles/vim/dot-vim/colors/
    Note: Target contains 3 files not in existing directory

Conflict resolution options:
  1. backup   - Backup existing files to *.bak and create symlinks
  2. skip     - Skip conflicting files, link others
  3. abort    - Cancel operation, make no changes
  4. adopt    - Move existing files into package, then create symlinks
  
Choose resolution strategy [abort]: 
```

#### User selects "backup"
```
Choose resolution strategy [abort]: backup

Creating backups and installing symlinks...
  ✓ Backed up ~/.vimrc → ~/.vimrc.bak
  ✓ Backed up ~/.vim/ → ~/.vim.bak/
  ✓ Created symlink ~/.vimrc → ~/dotfiles/vim/dot-vimrc
  ✓ Created symlink ~/.vim/ → ~/dotfiles/vim/dot-vim/

Package vim managed successfully.
Backups created: 2 files

Your existing files have been backed up. Would you like to adopt them into
the package to preserve your customizations? This will:
  • Replace package files with your backed-up versions
  • Keep your existing configuration
  • Bring your dotfiles under version control

Adopt backed-up files into package vim? [y/N]: y

Adopting backed-up files...
  ✓ Replaced ~/dotfiles/vim/dot-vimrc with ~/.vimrc.bak
  ✓ Merged ~/.vim.bak/ into ~/dotfiles/vim/dot-vim/
    • Preserved 3 custom files
    • Kept 2 package files (no conflicts)
  ✓ Updated symlinks

Adoption complete!

Your existing configuration is now managed by dot.

Next steps:
  • Review changes: cd ~/dotfiles/vim && git diff
  • Commit to version control: git add -A && git commit -m "feat(vim): adopt existing configuration"
  • Remove backups: rm -rf ~/.vimrc.bak ~/.vim.bak/

Verify installation:
  dot doctor vim
```

#### User declines adoption
```
Adopt backed-up files into package vim? [y/N]: n

Backups retained for your review.

Next steps:
  • Review backed up files: ls -la ~/*.bak
  • Compare configurations: diff ~/.vimrc.bak ~/.vimrc
  • Manually merge if desired: cp ~/.vimrc.bak ~/dotfiles/vim/dot-vimrc
  • Remove backups when satisfied: rm -rf ~/*.bak

Verify installation:
  dot doctor vim
```

#### Pre-flight Check Mode
```bash
$ dot manage vim --dry-run
```

```
Dry-run: Analyzing package vim...

Would create 3 symlinks:
  ~/.vimrc → ~/dotfiles/vim/dot-vimrc
  ~/.vim/ → ~/dotfiles/vim/dot-vim/
  ~/.vim/colors/ → ~/dotfiles/vim/dot-vim/colors/

Conflicts detected: 2
  ✗ ~/.vimrc (regular file exists)
  ✗ ~/.vim/ (directory exists)

No changes made (dry-run mode).

Would you like to proceed now? This will require conflict resolution.

[p] proceed - Continue with manage operation
[c] configure - Set conflict strategy and proceed
[a] abort - Cancel operation

Choice [a]: c

Select conflict resolution strategy:
  1. backup - Backup existing files and create symlinks
  2. skip - Skip conflicting files, link others
  3. adopt - Move existing files into package, then create symlinks
  4. abort - Cancel operation

Strategy [1]: 1

Proceeding with strategy: backup

Creating backups and installing symlinks...
  ✓ Backed up ~/.vimrc → ~/.vimrc.bak
  ✓ Backed up ~/.vim/ → ~/.vim.bak/
  ✓ Created symlink ~/.vimrc → ~/dotfiles/vim/dot-vimrc
  ✓ Created symlink ~/.vim/ → ~/dotfiles/vim/dot-vim/

Package vim managed successfully.
Backups created: 2 files

Your existing files have been backed up. Would you like to adopt them into
the package to preserve your customizations? This will:
  • Replace package files with your backed-up versions
  • Keep your existing configuration
  • Bring your dotfiles under version control

Adopt backed-up files into package vim? [y/N]: 
```

#### Pre-flight Check Without Interactive Prompt
```bash
# Non-interactive mode (for scripts/automation)
$ dot manage vim --dry-run --non-interactive
```

```
Dry-run: Analyzing package vim...

Would create 3 symlinks:
  ~/.vimrc → ~/dotfiles/vim/dot-vimrc
  ~/.vim/ → ~/dotfiles/vim/dot-vim/
  ~/.vim/colors/ → ~/dotfiles/vim/dot-vim/colors/

Conflicts detected: 2
  ✗ ~/.vimrc (regular file exists)
  ✗ ~/.vim/ (directory exists)

No changes made (dry-run mode).

To proceed:
  dot manage vim --on-conflict backup
  dot manage vim --on-conflict adopt
  dot manage vim --on-conflict skip
```

#### After Installation - Verification
```bash
$ dot doctor vim
```

```
Checking package vim...

Package Health: Healthy ✓

Links: 3 total
  ✓ ~/.vimrc → ~/dotfiles/vim/dot-vimrc
  ✓ ~/.vim/ → ~/dotfiles/vim/dot-vim/
  ✓ ~/.vim/colors/ → ~/dotfiles/vim/dot-vim/colors/

All links valid and pointing to correct targets.
No issues found.
```

---

### CUJ-A2: Migration from GNU Stow

**Scenario**: User has stow-managed dotfiles and wants to migrate to dot.

#### Initial State Detection
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Warnings ⚠

Managed packages: 0
Orphaned symlinks detected: 47 links
  ├─ 15 links → ~/.local/stow/vim/
  ├─ 12 links → ~/.local/stow/zsh/
  ├─ 8 links → ~/.local/stow/tmux/
  └─ 12 links → other locations

Pattern analysis suggests GNU Stow installation.
  • Stow directory detected: ~/.local/stow/
  • Packages detected: vim, zsh, tmux

Migration recommended:
  dot doctor migrate from-stow ~/.local/stow
```

#### Migration Command
```bash
$ dot doctor migrate from-stow ~/.local/stow
```

```
Analyzing Stow installation at ~/.local/stow/...

Detected 3 Stow packages:
  1. vim (15 symlinks)
  2. zsh (12 symlinks)  
  3. tmux (8 symlinks)

Migration plan:
  1. Create dot-compatible package structure
  2. Adopt existing symlinks into manifest
  3. Verify all symlinks transferred
  4. Optionally uninstall from Stow

This operation will:
  ✓ Preserve all existing symlinks
  ✓ Update manifest to track them
  ✗ Not modify Stow installation (manual cleanup required)

Proceed with migration? [y/N]: y

Migrating packages...
  ✓ Adopted vim (15 symlinks) → ~/dotfiles/vim
  ✓ Adopted zsh (12 symlinks) → ~/dotfiles/zsh
  ✓ Adopted tmux (8 symlinks) → ~/dotfiles/tmux

Migration complete: 3 packages, 35 symlinks

Verification:
  ✓ All symlinks functional
  ✓ Manifest updated
  ✓ No broken links

Next steps:
  • Verify migration: dot status
  • Test functionality: open new shell, vim, etc.
  • Uninstall from Stow: stow -D vim zsh tmux
  • Remove Stow directory: rm -rf ~/.local/stow (after testing)

Run health check:
  dot doctor
```

#### Post-Migration Verification
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Healthy ✓

Package Summary (3 packages, 35 links, 0.3s):

┌─────────┬───────┬────────┬─────────┬──────────┐
│ Package │ Links │ Status │ Broken  │ Warnings │
├─────────┼───────┼────────┼─────────┼──────────┤
│ vim     │    15 │ ✓      │       0 │        0 │
│ zsh     │    12 │ ✓      │       0 │        0 │
│ tmux    │     8 │ ✓      │       0 │        0 │
└─────────┴───────┴────────┴─────────┴──────────┘

All systems healthy after migration.

Migration complete: All Stow packages successfully adopted.
```

#### Post-Migration with Many Packages
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Healthy ✓

Package Summary (15 packages, 237 links, 1.2s):

┌──────────────┬───────┬────────┐
│ Package      │ Links │ Status │
├──────────────┼───────┼────────┤
│ vim          │    23 │ ✓      │
│ zsh          │    18 │ ✓      │
│ git          │     8 │ ✓      │
│ tmux         │    12 │ ✓      │
│ nvim         │    31 │ ✓      │
│ alacritty    │     4 │ ✓      │
│ kitty        │     6 │ ✓      │
│ bash         │     5 │ ✓      │
│ fish         │     8 │ ✓      │
│ starship     │     2 │ ✓      │
│ btop         │     3 │ ✓      │
│ scripts      │    27 │ ✓      │
│ i3           │    14 │ ✓      │
│ polybar      │     9 │ ✓      │
│ dunst        │     3 │ ✓      │
└──────────────┴───────┴────────┘

All 15 packages migrated successfully from Stow.

Next steps:
  • Commit migrated packages: cd ~/dotfiles && git add -A && git commit
  • Clean up Stow: stow -D vim zsh tmux ... (all packages)
  • Verify specific packages: dot doctor vim zsh
```

---

### CUJ-A3: Existing Dotfile Repository Adoption

**Scenario**: User has custom repository structure without dot- prefix.

#### Repository Structure Analysis
```bash
$ dot doctor check structure
```

```
Analyzing repository structure at ~/dotfiles/...

Repository: ~/dotfiles/
Packages found: 5

Structure compliance:
  ✓ vim/ - Compliant (uses dot- prefix)
  ✗ zsh/ - Non-compliant (files without dot- prefix)
    Missing: zshrc → dot-zshrc, zshenv → dot-zshenv
  ✗ tmux/ - Non-compliant (files without dot- prefix)
    Missing: tmux.conf → dot-tmux.conf
  ✓ git/ - Compliant (uses dot- prefix)
  ✗ nvim/ - Non-compliant (directory structure non-standard)
    Issue: Nested config directory without dot- prefix

Recommendations:
  1. Rename files to use dot- prefix
  2. Use dot convention helpers

Automated fix available:
  dot doctor fix structure --apply
```

#### Automated Structure Fix
```bash
$ dot doctor fix structure --dry-run
```

```
Dry-run: Structure fix for ~/dotfiles/

Would make the following changes:

Package: zsh
  Rename: zshrc → dot-zshrc
  Rename: zshenv → dot-zshenv
  Rename: zprofile → dot-zprofile

Package: tmux
  Rename: tmux.conf → dot-tmux.conf

Package: nvim
  Rename: config/ → dot-config/
  Note: Preserves nested structure

Total changes: 5 renames across 3 packages
No data loss will occur (renames only)

To apply:
  dot doctor fix structure --apply
```

#### Apply Fix
```bash
$ dot doctor fix structure --apply
```

```
Applying structure fixes...

Package: zsh
  ✓ Renamed zshrc → dot-zshrc
  ✓ Renamed zshenv → dot-zshenv
  ✓ Renamed zprofile → dot-zprofile

Package: tmux
  ✓ Renamed tmux.conf → dot-tmux.conf

Package: nvim
  ✓ Renamed config/ → dot-config/

Structure fixes complete: 5 renames

Verification:
  ✓ All renames successful
  ✓ Repository structure now compliant

Next steps:
  • Commit changes: git add -A && git commit -m "refactor: adopt dot naming convention"
  • Manage packages: dot manage zsh tmux nvim
  • Verify: dot doctor
```

---

## Category B: Routine Operations

### CUJ-B1: Daily Health Check

**Scenario**: Quick morning health check as part of daily routine.

#### Fast Mode (Default) - Healthy System
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Healthy ✓

Package Summary (12 packages, 147 links, 0.8s):

┌──────────────┬───────┬────────┬─────────┬──────────┐
│ Package      │ Links │ Status │ Broken  │ Warnings │
├──────────────┼───────┼────────┼─────────┼──────────┤
│ vim          │    23 │ ✓      │       0 │        0 │
│ zsh          │    18 │ ✓      │       0 │        0 │
│ git          │     8 │ ✓      │       0 │        0 │
│ tmux         │    12 │ ✓      │       0 │        0 │
│ nvim         │    31 │ ✓      │       0 │        0 │
│ alacritty    │     4 │ ✓      │       0 │        0 │
│ kitty        │     6 │ ✓      │       0 │        0 │
│ bash         │     5 │ ✓      │       0 │        0 │
│ fish         │     8 │ ✓      │       0 │        0 │
│ starship     │     2 │ ✓      │       0 │        0 │
│ btop         │     3 │ ✓      │       0 │        0 │
│ scripts      │    27 │ ✓      │       0 │        0 │
└──────────────┴───────┴────────┴─────────┴──────────┘

All systems healthy.

For comprehensive scan: dot doctor --deep
For package details: dot doctor <package>
```

#### Fast Mode - With Issues
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Errors ✗

Package Summary (12 packages, 147 links, 0.9s):

┌──────────────┬───────┬────────┬─────────┬──────────┐
│ Package      │ Links │ Status │ Broken  │ Warnings │
├──────────────┼───────┼────────┼─────────┼──────────┤
│ vim          │    23 │ ✗      │       2 │        0 │
│ zsh          │    18 │ ⚠      │       0 │        1 │
│ git          │     8 │ ✓      │       0 │        0 │
│ tmux         │    12 │ ✓      │       0 │        0 │
│ nvim         │    31 │ ✓      │       0 │        0 │
│ alacritty    │     4 │ ✓      │       0 │        0 │
│ kitty        │     6 │ ✓      │       0 │        0 │
│ bash         │     5 │ ✓      │       0 │        0 │
│ fish         │     8 │ ✓      │       0 │        0 │
│ starship     │     2 │ ✓      │       0 │        0 │
│ btop         │     3 │ ✓      │       0 │        0 │
│ scripts      │    27 │ ✓      │       0 │        0 │
└──────────────┴───────┴────────┴─────────┴──────────┘

Issues found in 2 packages:

vim (2 errors):
  ✗ ~/.vimrc → target does not exist
  ✗ ~/.vim/colors/ → target does not exist

zsh (1 warning):
  ⚠ ~/.bashrc → points outside package directory

Recommendations:
  • Fix broken links: dot doctor fix
  • Review warnings: dot doctor zsh -v
  • Check individual package: dot doctor vim

Exit code: 2 (errors detected)
```

#### Compact Mode (Many Packages)
```bash
$ dot doctor --compact
```

```
Health: Errors ✗ (2 packages affected, 0.9s)

┌──────────────┬───────┬────────┐
│ Package      │ Links │ Status │
├──────────────┼───────┼────────┤
│ vim          │    23 │ ✗ (2)  │
│ zsh          │    18 │ ⚠ (1)  │
│ git          │     8 │ ✓      │
│ tmux         │    12 │ ✓      │
│ nvim         │    31 │ ✓      │
│ alacritty    │     4 │ ✓      │
│ kitty        │     6 │ ✓      │
│ bash         │     5 │ ✓      │
│ fish         │     8 │ ✓      │
│ starship     │     2 │ ✓      │
│ btop         │     3 │ ✓      │
│ scripts      │    27 │ ✓      │
└──────────────┴───────┴────────┘

2 errors, 1 warning. Run 'dot doctor -v' for details.
```

#### List Format (Alternative)
```bash
$ dot doctor --format list
```

```
Analyzing dotfile health...

Overall Health: Errors ✗

Packages (12 total, 147 links):
  ✗ vim          23 links  (2 broken)
  ⚠ zsh          18 links  (1 warning)
  ✓ git           8 links
  ✓ tmux         12 links
  ✓ nvim         31 links
  ✓ alacritty     4 links
  ✓ kitty         6 links
  ✓ bash          5 links
  ✓ fish          8 links
  ✓ starship      2 links
  ✓ btop          3 links
  ✓ scripts      27 links

2 errors in vim, 1 warning in zsh

Details: dot doctor vim zsh
```

#### Filtering Options
```bash
# Show only packages with issues
$ dot doctor --issues-only
```

```
Health: Errors ✗

Packages with issues (2 of 12):

┌─────────┬───────┬────────┬─────────┬──────────┐
│ Package │ Links │ Status │ Broken  │ Warnings │
├─────────┼───────┼────────┼─────────┼──────────┤
│ vim     │    23 │ ✗      │       2 │        0 │
│ zsh     │    18 │ ⚠      │       0 │        1 │
└─────────┴───────┴────────┴─────────┴──────────┘

vim (2 errors):
  ✗ ~/.vimrc → target does not exist
  ✗ ~/.vim/colors/ → target does not exist

zsh (1 warning):
  ⚠ ~/.bashrc → points outside package directory

10 packages healthy (not shown)
Show all: dot doctor
```

#### Sorting Options
```bash
# Sort by number of links (largest first)
$ dot doctor --sort links

# Sort by package name (alphabetical)
$ dot doctor --sort name

# Sort by status (errors first, then warnings, then healthy)
$ dot doctor --sort status
```

#### Configuration for Default Display
```yaml
# ~/.config/dot/config.yaml
doctor:
  # Default output mode for multiple packages
  output_mode: table  # table | list | compact
  
  # Auto-switch to compact mode when package count exceeds threshold
  compact_threshold: 15
  
  # Show only packages with issues by default
  issues_only: false
  
  # Default sort order
  sort_by: status  # status | name | links
  
  # Maximum packages to show before pagination
  max_display: 20
```

#### Adaptive Display (Recommended Default)
The system automatically chooses the best format based on context:

- **1-5 packages**: Detailed inline view with full issue descriptions
- **6-15 packages**: Table view with summary, issues shown below
- **16+ packages**: Compact table, use `--issues-only` or `-v` for details
- **Terminal < 80 cols**: Automatically switches to list format
- **Non-TTY output**: Plain list format (no tables)
- **Output > terminal height**: Automatically paginated (see below)

#### Automatic Pagination
When output exceeds terminal height, `dot doctor` automatically enables pagination:

```bash
$ dot doctor  # 30 packages on 24-line terminal
```

```
Analyzing dotfile health...

Overall Health: Healthy ✓

Package Summary (30 packages, 412 links, 1.4s):

┌──────────────┬───────┬────────┐
│ Package      │ Links │ Status │
├──────────────┼───────┼────────┤
│ vim          │    23 │ ✓      │
│ zsh          │    18 │ ✓      │
│ git          │     8 │ ✓      │
│ tmux         │    12 │ ✓      │
│ nvim         │    31 │ ✓      │
│ alacritty    │     4 │ ✓      │
│ kitty        │     6 │ ✓      │
│ bash         │     5 │ ✓      │
│ fish         │     8 │ ✓      │
│ starship     │     2 │ ✓      │
│ btop         │     3 │ ✓      │
│ scripts      │    27 │ ✓      │
│ i3           │    14 │ ✓      │
│ polybar      │     9 │ ✓      │
│ dunst        │     3 │ ✓      │
│ rofi         │     5 │ ✓      │
│ picom        │     2 │ ✓      │
:                              (Press SPACE for more, q to quit)
```

**Pagination behavior:**
- Automatically detects terminal height using `LINES` env or ioctl
- Reserves 2 lines for status/navigation
- Uses built-in pager (similar to git, systemctl)
- Standard navigation: SPACE/f (forward), b (back), q (quit), / (search)
- Respects `PAGER` environment variable (less, more, bat, etc.)
- Disabled for non-TTY output (pipes, redirects)

**Manual control:**
```bash
dot doctor --no-pager        # Disable pagination
dot doctor --pager always    # Force pagination even for short output
dot doctor | less            # Manual piping works as expected
```

#### Terminal Size Detection

The system dynamically calculates terminal dimensions:

```go
// Pseudocode for implementation
terminalWidth := detectWidth()   // Default 80, max 120 for tables
terminalHeight := detectHeight() // Default 24

// Detection methods (in order of preference):
1. ioctl TIOCGWINSZ syscall (most accurate)
2. COLUMNS/LINES environment variables
3. tput cols/lines commands
4. Fallback to 80x24

// Responsive behavior:
if terminalWidth < 80 {
    useListFormat()  // Narrow terminal
} else if terminalWidth >= 120 {
    useWideTableFormat()  // Show more columns
}

if outputLines > (terminalHeight - 2) {
    enablePagination()  // Auto-page long output
}
```

**Configuration:**
```yaml
doctor:
  # Pagination settings
  auto_page: true              # Auto-paginate when output > terminal
  page_threshold: 0            # 0 = auto-detect terminal height
                               # N = paginate if output > N lines
  pager: auto                  # auto | always | never
  pager_command: ""            # Empty = use $PAGER or built-in
                               # Set to "less -R" or "bat" if preferred
  
  # Table width adaptation
  adaptive_width: true         # Adjust columns to terminal width
  max_table_width: 120         # Maximum table width
  min_table_width: 60          # Switch to list if narrower
```

#### Wide Terminal Support
On wide terminals (≥120 columns), show additional columns:

```
┌──────────────┬───────┬────────┬─────────┬──────────┬──────────────┬─────────────┐
│ Package      │ Links │ Status │ Broken  │ Warnings │ Last Changed │ Package Dir │
├──────────────┼───────┼────────┼─────────┼──────────┼──────────────┼─────────────┤
│ vim          │    23 │ ✓      │       0 │        0 │ 2 days ago   │ ~/dot/vim   │
│ zsh          │    18 │ ✓      │       0 │        0 │ 1 week ago   │ ~/dot/zsh   │
...
```

#### Narrow Terminal Handling
On narrow terminals (<80 columns), automatically switch to list format:

```
Packages (12 total):
  ✓ vim (23 links)
  ✓ zsh (18 links)
  ✓ git (8 links)
  ...
```

#### Scripting Mode (CI/CD)
```bash
$ dot doctor --format json --quiet
```

```json
{
  "overall_health": "errors",
  "timestamp": "2025-11-16T10:30:00Z",
  "duration_ms": 847,
  "statistics": {
    "total_packages": 12,
    "total_links": 147,
    "valid_links": 145,
    "broken_links": 2,
    "warnings": 1
  },
  "issues": [
    {
      "severity": "error",
      "type": "broken_link",
      "path": "~/.config/nvim/init.vim",
      "package": "nvim",
      "message": "Broken symlink (target does not exist)",
      "target": "~/dotfiles/nvim/dot-config/nvim/init.vim",
      "suggestions": [
        "Run 'dot remanage nvim'",
        "Check if file was renamed or deleted"
      ]
    },
    {
      "severity": "error",
      "type": "broken_link",
      "path": "~/.tmux.conf",
      "package": "tmux",
      "message": "Broken symlink (target does not exist)",
      "target": "~/dotfiles/tmux/dot-tmux.conf",
      "suggestions": [
        "Run 'dot remanage tmux'",
        "Check if file was renamed or deleted"
      ]
    },
    {
      "severity": "warning",
      "type": "wrong_target",
      "path": "~/.bashrc",
      "package": "bash",
      "message": "Points outside package directory",
      "target": "/usr/local/share/bash/bashrc"
    }
  ]
}
```

#### Exit Codes
```bash
$ dot doctor
$ echo $?
0   # Healthy

$ dot doctor  # (with warnings)
$ echo $?
1   # Warnings detected

$ dot doctor  # (with errors)
$ echo $?
2   # Errors detected
```

---

### CUJ-B2: Post-Update Verification

**Scenario**: After pulling dotfiles repo changes and running remanage.

#### Diff Mode
```bash
$ dot doctor diff --package vim --since last-check
```

```
Comparing vim package state...

Changes since last check (2 hours ago):

Added (2 files):
  + ~/.vim/after/plugin/telescope.lua
    Target: ~/dotfiles/vim/dot-vim/after/plugin/telescope.lua
    Status: ✓ Valid
    
  + ~/.config/nvim/lua/plugins/lsp.lua
    Target: ~/dotfiles/vim/dot-config/nvim/lua/plugins/lsp.lua
    Status: ✓ Valid

Removed (1 file):
  - ~/.vim/vimrc.old
    Was: symlink to ~/dotfiles/vim/dot-vimrc.old
    Status: Removed (file deleted from package)

Modified (1 file):
  ~ ~/.vimrc
    Target unchanged: ~/dotfiles/vim/dot-vimrc
    File modified: 2025-11-16 10:15:00
    Status: ✓ Valid

Summary:
  Total changes: 4
  ✓ 3 valid
  ✗ 0 broken
  ⚠ 0 warnings

Package vim is healthy after update.
```

#### Completeness Check
```bash
$ dot doctor check completeness vim
```

```
Checking completeness of package vim...

Package directory: ~/dotfiles/vim/
Managed links: 23

Completeness analysis:

Files in package: 25
  ✓ 23 files linked correctly
  ⚠ 2 files not linked

Unlinked files:
  ⚠ ~/dotfiles/vim/dot-vimrc.backup
    Not linked (matches ignore pattern: *.backup)
    
  ✗ ~/dotfiles/vim/dot-vim/after/plugin/new-plugin.lua
    Not linked (no symlink exists)
    Location: ~/.vim/after/plugin/new-plugin.lua (expected)
    Status: Missing symlink

Recommendations:
  • Review ignore patterns: dot config show ignore
  • Complete package: dot remanage vim
  • Create missing link manually: ln -s ~/dotfiles/vim/dot-vim/after/plugin/new-plugin.lua ~/.vim/after/plugin/new-plugin.lua

To fix automatically:
  dot remanage vim
```

---

### CUJ-B3: Package Addition

**Scenario**: Adding new neovim package with conflicts.

#### Pre-Manage Conflict Analysis
```bash
$ dot doctor check conflicts neovim --before-manage
```

```
Analyzing conflicts for package neovim...

Package: neovim
Source: ~/dotfiles/neovim/
Target directory: ~/

Conflict analysis:

Would create 5 symlinks:
  ~/.config/nvim/ → ~/dotfiles/neovim/dot-config/nvim/
  ~/.local/share/nvim/ → ~/dotfiles/neovim/dot-local/share/nvim/
  ~/.cache/nvim/ → ~/dotfiles/neovim/dot-cache/nvim/
  ~/.config/nvim/init.lua → ~/dotfiles/neovim/dot-config/nvim/init.lua
  ~/.config/nvim/lua/ → ~/dotfiles/neovim/dot-config/nvim/lua/

Conflicts detected: 3

Critical conflicts (would cause data loss):
  ✗ ~/.config/nvim/
    Existing: directory with 15 files (last modified: 2 days ago)
    Contains: init.lua, lua/, after/, plugin/
    Risk: HIGH - Contains user configuration
    Recommendation: adopt or backup

Minor conflicts (low risk):
  ⚠ ~/.cache/nvim/
    Existing: directory with cache files
    Contains: .keep, shada/
    Risk: LOW - Cache directory
    Recommendation: skip or overwrite

Safe operations: 2
  ✓ ~/.local/share/nvim/ - No conflict
  ✓ ~/.config/nvim/lua/ - No conflict (parent handled)

Recommended strategy:
  1. Backup existing config: --on-conflict backup
  2. Or adopt into package: dot adopt neovim ~/.config/nvim

To proceed:
  dot manage neovim --on-conflict backup
  dot adopt neovim ~/.config/nvim ~/.cache/nvim
```

#### Safe Manage with Adoption
```bash
$ dot adopt neovim ~/.config/nvim ~/.cache/nvim
```

```
Adopting files into package neovim...

Analyzing adoption candidates:
  ~/.config/nvim/ (directory, 15 files)
  ~/.cache/nvim/ (directory, 3 files)

Adoption plan:
  1. Move ~/.config/nvim/ → ~/dotfiles/neovim/dot-config/nvim/
  2. Move ~/.cache/nvim/ → ~/dotfiles/neovim/dot-cache/nvim/
  3. Create symlinks back to original locations
  4. Update package structure
  5. Update manifest

⚠ Warning: This will modify your package directory.
   Files will be moved from home directory into dotfiles repo.

Proceed with adoption? [y/N]: y

Adopting files...
  ✓ Moved ~/.config/nvim/ → ~/dotfiles/neovim/dot-config/nvim/
  ✓ Moved ~/.cache/nvim/ → ~/dotfiles/neovim/dot-cache/nvim/
  ✓ Created symlink ~/.config/nvim/ → ~/dotfiles/neovim/dot-config/nvim/
  ✓ Created symlink ~/.cache/nvim/ → ~/dotfiles/neovim/dot-cache/nvim/

Adoption complete: 2 directories, 18 files total

Next steps:
  • Review package structure: ls -la ~/dotfiles/neovim/
  • Commit to version control: cd ~/dotfiles && git add neovim/ && git commit
  • Verify links: dot doctor neovim

Verify adoption:
  dot doctor neovim
```

---

## Category C: System Changes and Drift

### CUJ-C1: Operating System Upgrade

**Scenario**: After OS upgrade, many absolute links are broken.

#### Detecting Systematic Issues
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Errors ✗

Managed packages: 12
Total links: 147
  ✓ 98 valid
  ✗ 49 broken links
  ⚠ 0 warnings

Pattern detected: Systematic path changes
  45 links reference old path: /home/blake/
  45 links should reference: /home/blake.smith/

Root cause analysis:
  Username changed: blake → blake.smith
  Home directory moved: /home/blake → /home/blake.smith
  Package directory affected: /home/blake/dotfiles → /home/blake.smith/dotfiles

This appears to be the result of:
  • Username change
  • Home directory relocation
  • System upgrade or migration

Recommendations:
  • Update configuration: dot config set directories.package ~/dotfiles
  • Remanage all packages: dot remanage --all
  • Verify fix: dot doctor

Automated fix available:
  dot doctor fix path-migration --from /home/blake --to /home/blake.smith
```

#### Automated Path Migration
```bash
$ dot doctor fix path-migration --from /home/blake --to /home/blake.smith --dry-run
```

```
Dry-run: Path migration analysis

Migration plan:
  Old base: /home/blake
  New base: /home/blake.smith

Affected components:

Configuration:
  • directories.package: /home/blake/dotfiles → /home/blake.smith/dotfiles

Manifest:
  • 12 packages with absolute paths
  • 49 absolute symlinks to update

Links requiring updates: 49
  vim: 8 links
  zsh: 12 links
  tmux: 3 links
  nvim: 15 links
  git: 4 links
  (and 7 more packages)

Migration strategy:
  1. Update configuration file
  2. Update manifest package paths
  3. Recreate all affected symlinks
  4. Verify all links functional

Estimated time: < 5 seconds
Risk: LOW (operation is reversible)

No changes made (dry-run mode).

To apply migration:
  dot doctor fix path-migration --from /home/blake --to /home/blake.smith --apply
```

#### Apply Migration
```bash
$ dot doctor fix path-migration --from /home/blake --to /home/blake.smith --apply
```

```
Applying path migration...

Step 1/4: Updating configuration
  ✓ Updated directories.package
  ✓ Configuration saved

Step 2/4: Updating manifest
  ✓ Updated 12 package paths
  ✓ Manifest saved

Step 3/4: Recreating symlinks
  Remanaging packages...
  ✓ vim (8 links)
  ✓ zsh (12 links)
  ✓ tmux (3 links)
  ✓ nvim (15 links)
  ✓ git (4 links)
  ✓ bash (3 links)
  ✓ alacritty (2 links)
  ✓ kitty (2 links)

Step 4/4: Verification
  ✓ All 49 links recreated successfully
  ✓ No broken links
  ✓ Configuration consistent

Migration complete!
  49 links updated across 12 packages

Verification:
  dot doctor
```

#### Post-Migration Health Check
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Healthy ✓

All systems healthy after migration.
```

---

### CUJ-C2: Dotfiles Directory Moved

**Scenario**: User moves dotfiles from `~/dotfiles` to `~/.local/share/dotfiles`.

#### Detection
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Critical ✗

Error: Package directory not accessible
  Configured: ~/dotfiles
  Status: Directory not found

Possible causes:
  • Directory was moved or deleted
  • Filesystem not mounted
  • Permission issues

Attempting to locate dotfiles...

Search results:
  ✓ Found: ~/.local/share/dotfiles/ (5 packages, 147 files)
    Last modified: 2 minutes ago
    Confidence: HIGH (matches expected structure)

  ⚠ Found: ~/old-dotfiles/ (3 packages, 89 files)
    Last modified: 30 days ago
    Confidence: MEDIUM (outdated)

Recommendations:
  • Update package directory: dot config set directories.package ~/.local/share/dotfiles
  • Then remanage all: dot remanage --all
  • Verify: dot doctor

Automated recovery available:
  dot doctor recover relocated-directory ~/.local/share/dotfiles
```

#### Automated Recovery
```bash
$ dot doctor recover relocated-directory ~/.local/share/dotfiles
```

```
Recovering from relocated package directory...

Current state:
  Configured package directory: ~/dotfiles (missing)
  Detected package directory: ~/.local/share/dotfiles (found)

Recovery plan:
  1. Update configuration to point to new location
  2. Verify package structure is intact
  3. Check manifest consistency
  4. Recreate all symlinks (relative links broken by move)

Verification checks:
  ✓ Directory exists: ~/.local/share/dotfiles
  ✓ Contains expected packages: 12 found
  ✓ Manifest file present
  ✓ Structure appears valid

⚠ Note: All relative symlinks will need recreation
  Affected: ~145 links (relative mode)

Proceed with recovery? [y/N]: y

Step 1/4: Updating configuration
  ✓ Set directories.package = ~/.local/share/dotfiles
  ✓ Configuration saved

Step 2/4: Verifying packages
  ✓ All 12 packages present and valid

Step 3/4: Recreating symlinks
  Remanaging all packages...
  ✓ vim (23 links)
  ✓ zsh (18 links)
  ✓ tmux (8 links)
  ... (9 more packages)
  
Step 4/4: Verification
  ✓ All 147 links recreated
  ✓ No broken links
  ✓ Manifest consistent

Recovery complete!

Summary:
  • Configuration updated
  • 147 links recreated
  • All packages healthy

Verify recovery:
  dot doctor
```

---

### CUJ-C3: External Tool Interference

**Scenario**: System package manager overwrote dot-managed symlink.

#### Detection
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Warnings ⚠

Managed packages: 12
Total links: 147
  ✓ 145 valid
  ✗ 0 broken
  ⚠ 2 warnings

Issues found:

Warnings (2):
  ⚠ ~/.bashrc
    Link target hijacked (points to wrong location)
    Expected: ~/dotfiles/bash/dot-bashrc
    Actual: /etc/skel/.bashrc
    Package: bash
    
    Analysis:
      • File type: symlink (expected)
      • Target changed externally
      • Likely cause: System package installation
      • Detected pattern: system skeleton file
    
    Resolution options:
      1. restore - Restore dot management (recreate symlink)
      2. ignore - Accept system management (add to ignore list)
      3. analyze - Show more information
    
  ⚠ ~/.config/git/config
    Link replaced with regular file
    Expected: symlink to ~/dotfiles/git/dot-config/git/config
    Actual: regular file (modified today)
    Package: git
    
    Analysis:
      • File type: regular file (expected: symlink)
      • Content differs from package source
      • Likely cause: Git command modified config directly
      • Size: 1.2 KB (package source: 0.8 KB)
    
    Resolution options:
      1. restore - Replace with symlink (lose file changes)
      2. merge - Merge changes into package source
      3. adopt - Replace package source with this file
      4. ignore - Keep as regular file, remove from management

Recommendations:
  • Review external changes before restoring
  • Consider if external management is appropriate

Interactive resolution:
  dot doctor fix --interactive
```

#### Interactive Resolution
```bash
$ dot doctor fix --interactive
```

```
Interactive issue resolution

Issue 1 of 2:
  ⚠ ~/.bashrc - Link target hijacked

Expected: ~/dotfiles/bash/dot-bashrc
Actual: /etc/skel/.bashrc

This symlink points to a system file instead of your dotfiles.
This likely happened when a system package was installed or updated.

Options:
  [r] restore   - Restore dot management (point back to ~/dotfiles/bash/dot-bashrc)
  [i] ignore    - Accept system management (add ~/.bashrc to ignore list)
  [d] diff      - Show differences between files
  [s] skip      - Skip this issue for now
  [q] quit      - Exit without making changes

Choice [r]: d

Comparing files...

--- /etc/skel/.bashrc (system)
+++ ~/dotfiles/bash/dot-bashrc (your dotfiles)
@@ -1,5 +1,8 @@
 # ~/.bashrc
 
+# Custom prompt
+PS1='\[\e[1;32m\]\u@\h\[\e[0m\]:\[\e[1;34m\]\w\[\e[0m\]\$ '
+
 # If not running interactively, don't do anything
 [[ $- != *i* ]] && return

System file is generic, your dotfiles have customizations.

Choice [r]: r

✓ Restored ~/.bashrc → ~/dotfiles/bash/dot-bashrc

Issue 2 of 2:
  ⚠ ~/.config/git/config - Link replaced with regular file

This was a symlink but is now a regular file.
The file was modified more recently than your package source.

Options:
  [r] restore   - Replace with symlink (lose changes in regular file)
  [m] merge     - Merge changes into package source
  [a] adopt     - Replace package source with this file
  [d] diff      - Show differences
  [i] ignore    - Keep as regular file, remove from management
  [s] skip      - Skip this issue
  [q] quit      - Exit

Choice [d]: d

Comparing files...

--- ~/dotfiles/git/dot-config/git/config (package source)
+++ ~/.config/git/config (current file)
@@ -10,6 +10,12 @@
 [user]
     name = Blake Smith
     email = blake@example.com
+    signingkey = ABC123DEF456
+
+[commit]
+    gpgsign = true
+
+[tag]
+    gpgsign = true

The current file has additional GPG signing configuration.

Choice [a]: a

Adopting changes into package...
  ✓ Backed up package source: ~/dotfiles/git/dot-config/git/config.bak
  ✓ Copied current file to package: ~/dotfiles/git/dot-config/git/config
  ✓ Recreated symlink: ~/.config/git/config → ~/dotfiles/git/dot-config/git/config

⚠ Package directory was modified. Remember to commit changes:
  cd ~/dotfiles && git add git/dot-config/git/config && git commit

Resolution complete: 2 issues fixed

Summary:
  • 1 link restored
  • 1 file adopted

Verify fixes:
  dot doctor
```

---

### CUJ-C4: Filesystem Corruption

**Scenario**: System crash corrupted manifest and created circular symlinks.

#### Detection
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Error: Manifest file corrupted
  Location: ~/.dot-manifest.json
  Issue: JSON parse error at line 47, column 12
  
Attempting recovery...

Recovery options:

1. Repair manifest (recommended)
   • Validate and fix JSON syntax
   • Verify structure integrity
   • Preserve as much data as possible

2. Rebuild from filesystem
   • Scan for symlinks
   • Infer package structure
   • Create new manifest
   • Risk: May lose metadata (timestamps, notes)

3. Restore from backup
   • Requires backup file
   • Full restoration
   • No data loss

Choose recovery method:
  dot doctor recover manifest --repair
  dot doctor recover manifest --rebuild
  dot doctor recover manifest --from-backup <file>
```

#### Manifest Repair
```bash
$ dot doctor recover manifest --repair --dry-run
```

```
Dry-run: Manifest repair analysis

Analyzing manifest: ~/.dot-manifest.json

Issues detected:

1. JSON syntax error (line 47, col 12)
   Error: unexpected end of input
   Context: ...links": [
   Fix: Add missing closing bracket

2. Truncated package entry
   Package: tmux
   Issue: Incomplete links array
   Fix: Recover from filesystem

3. Count mismatch
   Package: vim
   Manifest: 23 links
   Filesystem: 25 links
   Fix: Update count, add missing links

Repair plan:
  1. Fix JSON syntax errors
  2. Complete truncated entries by scanning filesystem
  3. Update counts to match reality
  4. Validate all package references
  5. Create backup of original

Estimated recovery:
  ✓ Full recovery possible
  ⚠ 2 links may lose metadata (timestamps)
  
No changes made (dry-run mode).

To apply repair:
  dot doctor recover manifest --repair --apply
```

#### Apply Repair
```bash
$ dot doctor recover manifest --repair --apply
```

```
Repairing manifest...

Step 1/5: Creating backup
  ✓ Backed up corrupt manifest: ~/.dot-manifest.json.corrupt

Step 2/5: Fixing JSON syntax
  ✓ Fixed 1 syntax error
  ✓ Added missing bracket
  ✓ Manifest now valid JSON

Step 3/5: Completing truncated entries
  Scanning filesystem for tmux package...
  ✓ Found 8 symlinks
  ✓ Reconstructed package entry

Step 4/5: Updating counts
  Package vim:
    ✓ Updated count: 23 → 25
    ✓ Added missing links: 2

Step 5/5: Validation
  ✓ All packages valid
  ✓ All link paths exist
  ✓ Structure consistent

Manifest repair complete!

Summary:
  • Fixed 1 syntax error
  • Recovered 1 truncated package
  • Updated 1 count mismatch
  • Added 2 missing links

Corrupt manifest saved as: ~/.dot-manifest.json.corrupt

Verify repair:
  dot doctor
```

#### Circular Symlink Detection
```bash
$ dot doctor check --deep
```

```
Deep scan analysis...

Overall Health: Errors ✗

Critical issue detected: Circular symlinks

Circular reference detected:
  ~/.config/app/config
    ↓ symlink to ~/.config/app/config.link
    ↓ symlink to ~/.config/app/config
    ↻ circular reference (depth: 2)

Impact:
  • Cannot traverse link
  • Applications may fail
  • Filesystem operations affected

Resolution:
  Break circular reference by removing intermediate link

Automated fix:
  dot doctor fix circular-links
```

---

## Category D: Error Recovery

### CUJ-D1: Accidental Unmanage

**Scenario**: User accidentally unmanaged package and wants to restore.

#### Detection of Missing Package
```bash
$ dot doctor check orphaned-packages
```

```
Checking for orphaned packages...

Analysis of ~/dotfiles/:

Managed packages: 11
Repository packages: 12

Orphaned package detected:
  ✗ vim
    Location: ~/dotfiles/vim/
    Status: Present in repository, not in manifest
    Files: 25
    Last managed: 2 hours ago (detected from history)
    
    This package exists in your repository but is not managed.
    
    Possible causes:
      • Accidentally unmanaged
      • Manifest corruption
      • Manual manifest edit

Recommendations:
  • Re-manage package: dot manage vim
  • Check operation history: dot doctor history

Restore automatically:
  dot doctor restore orphaned-packages
```

#### Operation History
```bash
$ dot doctor history
```

```
Doctor operation history (last 10 operations):

2025-11-16 14:30:00 | Unmanage | Package: vim
  Action: Removed 25 symlinks
  Status: Completed
  Reversible: Yes
  
2025-11-16 10:15:00 | Fix | Path migration
  Action: Updated 49 symlinks
  Status: Completed
  Reversible: Yes (with caution)
  
2025-11-15 09:00:00 | Manage | Package: neovim
  Action: Created 15 symlinks
  Status: Completed
  Reversible: Yes
  
... (7 more operations)

To undo last operation:
  dot doctor undo

To undo specific operation:
  dot doctor undo --operation <id>
```

#### Undo Operation
```bash
$ dot doctor undo
```

```
Undo last operation: Unmanage vim

Operation details:
  Timestamp: 2025-11-16 14:30:00 (2 hours ago)
  Action: Removed 25 symlinks from package vim
  Package directory: ~/dotfiles/vim/
  Affected files: 25 symlinks

Undo plan:
  1. Restore package to manifest
  2. Recreate 25 symlinks
  3. Verify all links functional
  4. Update operation history

This will restore the package as if unmanage was never run.

Proceed with undo? [y/N]: y

Undoing operation...
  ✓ Restored package vim to manifest
  ✓ Recreated 25 symlinks
  ✓ Verified all links functional
  ✓ Updated history

Undo complete: Package vim restored

Verify restoration:
  dot doctor vim
```

---

### CUJ-D2: Broken Package After Edit

**Scenario**: User renamed files in package, manifest out of sync.

#### Detection
```bash
$ dot doctor vim
```

```
Checking package vim...

Package Health: Errors ✗

Structure inconsistency detected:

Manifest references: 25 files
Filesystem links: 23 valid, 2 broken

Broken links:
  ✗ ~/.vimrc
    Target: ~/dotfiles/vim/dot-vimrc (does not exist)
    
  ✗ ~/.vim/plugins/old-plugin.vim
    Target: ~/dotfiles/vim/dot-vim/plugins/old-plugin.vim (does not exist)

Unlinked files in package:
  + ~/dotfiles/vim/dot-vimrc.new (not linked)
  + ~/dotfiles/vim/dot-vim/plugins/new-plugin.vim (not linked)

Analysis:
  Pattern suggests file renames:
    • dot-vimrc → dot-vimrc.new
    • old-plugin.vim → new-plugin.vim

Root cause: Package structure changed without remanage

Resolution:
  Run remanage to sync manifest with package content
  
Fix automatically:
  dot remanage vim
```

#### Remanage to Fix
```bash
$ dot remanage vim
```

```
Remanaging package vim...

Analyzing changes:
  - Removed 2 broken links
  + Added 2 new files
  ~ Updated manifest

Changes:
  Removed links (broken):
    - ~/.vimrc (target missing)
    - ~/.vim/plugins/old-plugin.vim (target missing)
    
  Added links (new files):
    + ~/.vimrc.new → ~/dotfiles/vim/dot-vimrc.new
    + ~/.vim/plugins/new-plugin.vim → ~/dotfiles/vim/dot-vim/plugins/new-plugin.vim

⚠ Note: File renames detected
  If these were renames (not new files), you may want:
    • Rename in package: mv dot-vimrc.new dot-vimrc
    • Then remanage again: dot remanage vim

Apply changes? [y/N]: y

Applying changes...
  ✓ Removed 2 broken links
  ✓ Created 2 new links
  ✓ Updated manifest

Remanage complete: Package vim now consistent

Verify:
  dot doctor vim
```

---

### CUJ-D3: Manifest Manually Edited

**Scenario**: User edited manifest, introducing inconsistencies.

#### Validation
```bash
$ dot doctor check manifest
```

```
Validating manifest integrity...

Manifest: ~/.dot-manifest.json

Issues found:

1. Count mismatch (package: zsh)
   Manifest field: links_count = 18
   Actual links: 16
   Fix: Update count field

2. Phantom link (package: tmux)
   Manifest references: ~/.tmux/plugins/tpm
   Filesystem: Link does not exist
   Fix: Remove from manifest or create link

3. Untracked link (package: vim)
   Filesystem: ~/.vim/after/ftplugin/rust.vim
   Manifest: Not referenced
   Fix: Add to manifest

4. Invalid path (package: git)
   Manifest references: ~/.config/git/config with invalid characters
   Issue: Contains null byte
   Fix: Sanitize path

Consistency check:
  ✗ Manifest inconsistent with filesystem
  ⚠ 4 issues found across 4 packages

Recommendations:
  • Sync manifest with filesystem: dot doctor sync manifest
  • Rebuild manifest: dot doctor recover manifest --rebuild

Automated fix:
  dot doctor sync manifest
```

#### Sync Manifest
```bash
$ dot doctor sync manifest --dry-run
```

```
Dry-run: Manifest synchronization

Synchronization plan:

Package: zsh
  Action: Update count
  Change: links_count 18 → 16
  
Package: tmux
  Action: Remove phantom link
  Remove: ~/.tmux/plugins/tpm (link does not exist)
  
Package: vim
  Action: Add missing link
  Add: ~/.vim/after/ftplugin/rust.vim
  
Package: git
  Action: Sanitize path
  Change: Remove null byte from path

Summary:
  • 1 count updated
  • 1 phantom link removed
  • 1 missing link added
  • 1 path sanitized

Result: Manifest will be consistent with filesystem

No changes made (dry-run mode).

To apply:
  dot doctor sync manifest --apply
```

---

### CUJ-D4: Orphaned Links After Package Removal

**Scenario**: User deleted package directory without unmanaging.

#### Detection
```bash
$ dot doctor
```

```
Analyzing dotfile health...

Overall Health: Errors ✗

Phantom package detected:
  ✗ tmux
    Manifest: Package exists
    Filesystem: Package directory missing
    Location: ~/dotfiles/tmux/ (not found)
    Managed links: 8 (all broken)

All links for package tmux are broken:
  ✗ ~/.tmux.conf
  ✗ ~/.tmux/plugins/
  ✗ ~/.tmux/themes/
  ... (5 more)

This package appears to have been deleted without unmanaging.

Recommendations:
  • Remove package and clean links: dot doctor clean phantom-packages
  • Restore package: If you have backup, restore directory and remanage

Clean up automatically:
  dot doctor clean phantom-packages
```

#### Clean Phantom Package
```bash
$ dot doctor clean phantom-packages --dry-run
```

```
Dry-run: Phantom package cleanup

Detected phantom packages: 1

Package: tmux
  Manifest entry: Exists
  Package directory: ~/dotfiles/tmux/ (missing)
  Broken links: 8

Cleanup plan:
  1. Remove all broken symlinks (8 files)
  2. Remove package from manifest
  3. Create cleanup report

Links to be removed:
  - ~/.tmux.conf
  - ~/.tmux/plugins/
  - ~/.tmux/themes/
  - ~/.tmux/resurrect/
  - ~/.tmux.conf.local
  - ~/.config/tmux/
  - ~/.local/share/tmux/
  - ~/.cache/tmux/

⚠ This is a destructive operation
  Symlinks will be deleted permanently
  
No changes made (dry-run mode).

To apply:
  dot doctor clean phantom-packages --apply
```

#### Apply Cleanup
```bash
$ dot doctor clean phantom-packages --apply
```

```
Cleaning phantom packages...

⚠ Warning: This will delete broken symlinks
  Package: tmux (8 links)

Confirm deletion? [y/N]: y

Removing broken symlinks...
  ✓ Removed ~/.tmux.conf
  ✓ Removed ~/.tmux/plugins/
  ✓ Removed ~/.tmux/themes/
  ✓ Removed ~/.tmux/resurrect/
  ✓ Removed ~/.tmux.conf.local
  ✓ Removed ~/.config/tmux/
  ✓ Removed ~/.local/share/tmux/
  ✓ Removed ~/.cache/tmux/

Updating manifest...
  ✓ Removed package tmux from manifest

Cleanup complete!

Summary:
  • Removed 8 broken symlinks
  • Removed 1 package from manifest
  • Cleanup report: ~/. dot-doctor-cleanup-2025-11-16.log

Verify cleanup:
  dot doctor
```

---

## Category E: Multi-Machine Synchronization

### CUJ-E1: Dotfiles Across Different Machines

**Scenario**: Platform-specific issues on different machines.

#### Platform Analysis
```bash
$ dot doctor check platform
```

```
Platform compatibility analysis

Current system:
  OS: macOS 15.1
  Platform: darwin-arm64
  Shell: /bin/zsh
  Home: /Users/blake

Analyzing dotfile assumptions...

Platform-specific issues found:

1. Hard-coded Linux paths (3 references)
   Package: bash
   Files:
     • ~/.bashrc references /home/blake (should be ~)
     • ~/.bash_profile sources /etc/profile.d/* (Linux-specific)
   
   Impact: HIGH
   Fix: Use portable path variables

2. Linux-specific tools referenced (2 tools)
   Package: scripts
   Files:
     • ~/.local/bin/update references apt-get (Debian/Ubuntu)
     • ~/.local/bin/clipboard uses xclip (X11, not macOS)
   
   Impact: MEDIUM
   Recommendation: Add platform detection to scripts

3. Missing macOS integrations (3 opportunities)
   Files:
     • ~/.config/git/config could use osxkeychain credential helper
     • ~/.zshrc could source /opt/homebrew/ paths
     • ~/.tmux.conf could use pbcopy/pbpaste
   
   Impact: LOW
   Recommendation: Add platform-specific config sections

Portability score: 6/10
  ✓ 80% of dotfiles are portable
  ⚠ 20% have platform-specific assumptions

Recommendations:
  • Use conditional logic: if [[ "$OSTYPE" == "darwin"* ]]; then
  • Abstract tool names: Use environment variables for tools
  • Test on each platform: dot doctor check platform

Generate platform report:
  dot doctor check platform --report platform-compatibility.md
```

---

### CUJ-E2: Fresh Clone on New Machine

**Scenario**: Bootstrap on new machine with pre-flight validation.

#### Pre-Bootstrap Validation
```bash
$ dot doctor check bootstrap --config bootstrap.yaml
```

```
Validating bootstrap configuration...

Configuration: bootstrap.yaml

Environment checks:

✓ Prerequisites met:
  ✓ Git installed (version 2.42.0)
  ✓ Shell: zsh (expected)
  ✓ Home directory writable
  ✓ Network connectivity

✗ Missing dependencies (3):
  ✗ neovim not found
    Required by: packages.neovim
    Install: brew install neovim
    
  ✗ tmux version too old
    Found: tmux 2.9
    Required: >= 3.0
    Install: brew upgrade tmux
    
  ✗ ~/.config/git/ not writable
    Required by: packages.git
    Fix: chmod u+w ~/.config/

⚠ Potential conflicts (2):
  ⚠ ~/.vimrc exists (regular file)
    Will conflict with: packages.vim
    Resolution available: --on-conflict backup
    
  ⚠ ~/.zshrc exists (symlink to different location)
    Points to: /usr/local/share/zsh/zshrc
    Will conflict with: packages.zsh
    Resolution required: Manual review

Bootstrap feasibility: Conditional ⚠
  Can proceed with: --skip-missing-deps
  Must resolve: Conflict issues

Recommendations:
  1. Install missing dependencies
  2. Review and resolve conflicts
  3. Run with: dot bootstrap --on-conflict backup

Dependency install script:
  brew install neovim
  brew upgrade tmux
  chmod u+w ~/.config/

To proceed with warnings:
  dot bootstrap --skip-missing-deps --on-conflict backup
```

---

### CUJ-E3: Selective Package Sync

**Scenario**: Machine-specific package expectations.

#### Machine Profile Definition
```bash
$ cat ~/.config/dot/profiles/work-laptop.yaml
```

```yaml
profile:
  name: work-laptop
  description: Work MacBook Pro configuration
  
packages:
  required:
    - vim
    - zsh
    - git
    - tmux
  
  optional:
    - alacritty
    - kitty
  
  excluded:
    - personal-scripts
    - gaming-tools

expectations:
  min_packages: 4
  allow_extra: false
```

#### Profile Validation
```bash
$ dot doctor check profile work-laptop
```

```
Validating against profile: work-laptop

Profile: ~/.config/dot/profiles/work-laptop.yaml

Compliance check:

Required packages (4):
  ✓ vim - Present and healthy
  ✓ zsh - Present and healthy
  ✓ git - Present and healthy
  ✓ tmux - Present and healthy

Optional packages (2):
  ✓ alacritty - Present
  ✗ kitty - Not installed (acceptable)

Excluded packages (2):
  ✗ personal-scripts - Present (should not be)
    Action required: Unmanage this package
  ✓ gaming-tools - Not present (correct)

Profile compliance: Violations found ⚠

Violations:
  1. Excluded package present: personal-scripts

Recommendations:
  • Unmanage excluded packages: dot unmanage personal-scripts
  • Update profile if package should be allowed
  • Sync with profile: dot doctor sync profile work-laptop

Auto-sync:
  dot doctor sync profile work-laptop --apply
```

---

## Category F: Advanced Scenarios

### CUJ-F1: Security Audit

**Scenario**: Audit dotfiles for security issues before team distribution.

#### Security Scan
```bash
$ dot doctor check security
```

```
Security audit of dotfiles...

Scanning for security issues...

Critical issues (2):
  ✗ Potential secret detected
    File: ~/dotfiles/git/dot-gitconfig
    Line 23: token = ghp_xxxxxxxxxxxxxxxxxxxx
    Pattern: GitHub Personal Access Token
    Risk: HIGH - Active credential exposure
    
  ✗ Private key in repository
    File: ~/dotfiles/ssh/dot-ssh/id_rsa
    Type: RSA private key (2048-bit)
    Risk: CRITICAL - Private key should never be in repository

High issues (3):
  ⚠ Insecure permissions
    File: ~/dotfiles/ssh/dot-ssh/config
    Permissions: 0644 (world-readable)
    Recommended: 0600
    
  ⚠ API key pattern detected
    File: ~/dotfiles/scripts/dot-local/bin/deploy
    Line 15: API_KEY="sk-xxxxxxxxxxxxxxxx"
    Pattern: OpenAI API key format
    Risk: HIGH - Potential active credential
    
  ⚠ Password in plain text
    File: ~/dotfiles/config/dot-config/app/config.json
    Line 8: "password": "mypassword123"
    Risk: HIGH - Plain text password storage

Medium issues (5):
  ⚠ Hardcoded IP addresses (3 instances)
  ⚠ Email addresses in configs (2 instances)

Permission issues (2):
  ⚠ World-writable file: ~/dotfiles/scripts/dot-local/bin/sync
  ⚠ Executable without shebang: ~/dotfiles/scripts/dot-local/bin/helper

Security score: 3/10 (Critical issues found)

⚠ DO NOT distribute these dotfiles without fixing critical issues

Recommendations:
  1. Remove all secrets: Use environment variables or secret manager
  2. Remove private keys: Never commit private keys
  3. Fix permissions: chmod 600 on sensitive files
  4. Review all detected patterns

Detailed report:
  dot doctor check security --report security-audit.md

Fix permissions automatically:
  dot doctor fix security-permissions
```

#### Security Report Generation
```bash
$ dot doctor check security --report security-audit.md --format markdown
```

```
Generating security audit report...

✓ Report generated: security-audit.md

Report includes:
  • Executive summary
  • Risk assessment
  • Detailed findings with line numbers
  • Remediation recommendations
  • Compliance checklist

View report:
  cat security-audit.md
  open security-audit.md
```

---

### CUJ-F2: Large-Scale Deployment

**Scenario**: Verify deployment across 50 machines.

#### Batch Health Check
```bash
# On each machine, run with JSON output
$ dot doctor --format json --quiet > /tmp/health-$(hostname).json
```

#### Aggregate Analysis (on management machine)
```bash
$ dot doctor analyze-fleet --reports /tmp/health-*.json
```

```
Fleet health analysis

Total machines: 50
Report files: 50

Overall fleet health:

Healthy: 45 machines (90%)
  ✓ All systems operational
  ✓ No issues detected

Warnings: 3 machines (6%)
  ⚠ host-23: 2 orphaned links
  ⚠ host-31: 1 wrong target
  ⚠ host-47: 1 permission issue

Errors: 2 machines (4%)
  ✗ host-12: 5 broken links (package: vim)
  ✗ host-38: Manifest corrupted

Common issues across fleet:

1. Package vim (2 machines affected)
   Broken links: 5 instances on host-12
   Recommendation: Investigate package integrity

2. Orphaned links pattern (3 machines)
   Pattern: ~/.local/bin/* symlinks
   Common cause: System package installation
   Recommendation: Add to ignore patterns fleet-wide

3. Manifest corruption (1 machine)
   Machine: host-38
   Recommendation: Manual investigation required

Fleet-wide recommendations:
  • Update ignore patterns for ~/.local/bin/*
  • Investigate vim package on host-12
  • Manual fix required for host-38

Machine-specific reports:
  host-12: dot doctor --format json | jq .issues
  host-38: dot doctor recover manifest

Detailed report:
  dot doctor analyze-fleet --report fleet-health-2025-11-16.html
```

---

### CUJ-F3: Complex Package Interdependencies

**Scenario**: Analyzing cross-package dependencies.

#### Dependency Analysis
```bash
$ dot doctor analyze dependencies
```

```
Analyzing package dependencies...

Dependency graph (12 packages):

zsh
  ├─ requires: git (for version control aliases)
  ├─ requires: vim (for EDITOR variable)
  └─ sources: ~/.config/shell-common (from package: shell-common)

vim
  ├─ requires: git (for vim-fugitive plugin)
  └─ no external file references

nvim
  ├─ requires: git (for lazy.nvim)
  ├─ requires: node (for LSP servers) [external]
  └─ references: ~/.config/nvim/lua/... (self-contained)

git
  ├─ references: ~/.config/shell-common/aliases (from package: shell-common)
  └─ no package dependencies

tmux
  ├─ requires: zsh (for default shell)
  ├─ sources: ~/.tmux/plugins/... (self-contained)
  └─ no package dependencies

shell-common
  ├─ provides: ~/.config/shell-common/* (shared by zsh, bash)
  └─ no dependencies

Dependency issues:

⚠ Circular reference potential:
  zsh → shell-common → (provides files used by zsh)
  Risk: LOW (different file levels)

⚠ External dependencies (not managed):
  nvim → node (required for LSP)
  nvim → ripgrep (required for telescope)
  tmux → tpm (tmux plugin manager)

Recommendations:
  • Document external dependencies in README
  • Consider bootstrap script for external deps
  • Test: Can packages work independently?

Management order recommendation:
  1. shell-common (no dependencies)
  2. git (uses shell-common)
  3. vim, nvim (independent, use git)
  4. zsh (uses shell-common, git, vim)
  5. tmux (uses zsh)

Verify dependency health:
  dot doctor check dependencies --verify
```

---

### CUJ-F4: Performance Issues

**Scenario**: Doctor runs slowly with large dotfile collection.

#### Performance Analysis
```bash
$ dot doctor --profile
```

```
Analyzing dotfile health...

Overall Health: Healthy ✓

Performance profile:

Total time: 45.3 seconds

Time breakdown:
  Loading manifest: 0.1s (0.2%)
  Checking managed links: 2.3s (5.1%)
  Orphan scanning: 42.8s (94.5%)
    ├─ Directory traversal: 38.2s (84.3%)
    ├─ Symlink checking: 3.9s (8.6%)
    └─ Pattern matching: 0.7s (1.5%)
  Reporting: 0.1s (0.2%)

Performance issues detected:

✗ Orphan scanning too slow
  Directories scanned: 45,382
  Files checked: 234,567
  Bottleneck: Deep directory traversal
  
  Recommendations:
    • Use fast mode: dot doctor (skips orphan scan)
    • Use scoped scan: dot doctor --scan-mode scoped (default)
    • Exclude large dirs: Add to skip patterns in config
    • Increase workers: Set --max-workers 8

Suggested skip patterns:
  • ~/Library/ (macOS, 158,234 files)
  • ~/node_modules/ (if present)
  • ~/.cargo/ (Rust toolchain)
  • ~/.local/share/Steam/ (if present)

Performance optimization:
  dot config set doctor.skip_patterns '["Library", ".cargo", "node_modules"]'

Fast mode (recommended for daily use):
  dot doctor              # < 3 seconds
  dot doctor --deep       # Full scan when needed
```

#### Fast Mode Configuration
```bash
$ cat ~/.config/dot/config.yaml
```

```yaml
doctor:
  # Default scan mode for orphan detection
  default_scan_mode: scoped  # off | scoped | deep
  
  # Maximum directory depth for deep scans
  max_depth: 10
  
  # Directories to skip during orphan scanning
  skip_patterns:
    - Library              # macOS system library
    - .cargo               # Rust toolchain
    - node_modules         # Node.js dependencies
    - .local/share/Steam   # Games
    - .cache               # Various caches
    - .npm                 # NPM cache
    - go/pkg               # Go package cache
  
  # Performance tuning
  max_workers: 4           # Parallel workers for scanning
  max_issues: 100          # Stop after finding N issues
  
  # Caching (future feature)
  # cache_scan_results: true
  # cache_duration: 1h
```

---

## Command Reference Summary

### Quick Reference

```bash
# Daily use (fast)
dot doctor                              # Quick health check
dot doctor <package> [package...]       # Check specific package(s)

# Comprehensive analysis
dot doctor --deep                       # Deep scan
dot doctor manifest                     # Validate manifest
dot doctor structure                    # Check repo structure
dot doctor security                     # Security audit
dot doctor platform                     # Platform compatibility

# Repair operations
dot doctor fix                          # Interactive repair
dot doctor fix --auto                   # Auto-fix safe issues
dot doctor fix --dry-run                # Preview changes
dot doctor clean orphans                # Remove orphaned links
dot doctor sync manifest                # Sync manifest with filesystem

# Recovery operations
dot doctor recover manifest             # Repair/rebuild manifest
dot doctor recover relocated-directory  # Handle moved dotfiles
dot doctor undo                         # Undo last operation

# Analysis operations
dot doctor analyze dependencies         # Dependency graph
dot doctor analyze-fleet                # Multi-machine analysis
dot doctor check bootstrap              # Validate bootstrap config
dot doctor check profile                # Validate machine profile

# Migration/adoption
dot doctor migrate from-stow            # Migrate from GNU Stow
dot doctor check conflicts <package>    # Pre-manage conflict check

# Management
dot doctor ignore add <path>            # Add to ignore list
dot doctor ignore list                  # Show ignored items
dot doctor history                      # Show operation history
```

---

## Output Format Examples

### Text Format (Human-Readable)
Default for interactive use, with color and formatting.

### JSON Format (Machine-Readable)
```json
{
  "overall_health": "errors",
  "timestamp": "2025-11-16T10:30:00Z",
  "duration_ms": 847,
  "statistics": {...},
  "issues": [...]
}
```

### Table Format (Structured)
```
| Package | Links | Valid | Broken | Status  |
|---------|-------|-------|--------|---------|
| vim     | 23    | 23    | 0      | Healthy |
| zsh     | 18    | 16    | 2      | Errors  |
| tmux    | 8     | 8     | 0      | Healthy |
```

### Markdown Format (Reports)
For generating documentation and reports.

---

## Exit Codes

```
0   - Healthy (no issues)
1   - Warnings detected (non-critical issues)
2   - Errors detected (critical issues requiring attention)
3   - Manifest corruption or critical system error
4   - Invalid command or configuration error
```

---

## Progressive Disclosure Pattern

All commands follow a pattern of progressive disclosure:

1. **Summary** - High-level status (1-3 lines)
2. **Statistics** - Key metrics
3. **Issues** - Grouped by severity
4. **Details** - Shown on request (-v, --verbose)
5. **Recommendations** - Actionable next steps
6. **Commands** - Exact commands to run

Example verbosity levels:

```bash
# Default - Summary only
$ dot doctor
Overall Health: Errors ✗
2 broken links in package vim

# Verbose - Show details
$ dot doctor -v
Overall Health: Errors ✗
Package vim: 2 broken links
  ✗ ~/.vimrc → ~/dotfiles/vim/dot-vimrc (not found)
  ✗ ~/.vim/colors/ → ~/dotfiles/vim/dot-vim/colors/ (not found)

# Very verbose - Show analysis
$ dot doctor -vv
[Shows full diagnostic process, file checks, pattern analysis]

# Debug - Show everything
$ dot doctor -vvv
[Shows all internal operations, timing, decisions]
```

---

## Document Metadata

- **Version**: 1.0
- **Date**: 2025-11-16
- **Based On**: doctor-system-cujs.md v1.0
- **Status**: Draft for Review
- **Next Steps**: Implementation planning

