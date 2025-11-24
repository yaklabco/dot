# Migration Guide: v0.2 → v0.3

## Breaking Changes in v0.3.0

### Consistent Dotfile Prefix Handling

Version 0.3 introduces consistent "dot-" prefix handling across all commands.

## What Changed

### Package Naming

**Before (v0.2):**
- Auto-naming stripped dots: `.ssh` → package `ssh`
- Nested structure: `ssh/dot-ssh/config`

**After (v0.3):**
- Auto-naming preserves dots: `.ssh` → package `dot-ssh`  
- Flat structure: `dot-ssh/config`

### Directory Structure

**Before:**
```text
~/dotfiles/
  └── ssh/              ← Package directory
      └── dot-ssh/      ← Nested adopted directory
          ├── config
          ├── known_hosts
          └── id_ed25519

~/.ssh → ~/dotfiles/ssh/dot-ssh/
```

**After:**
```text
~/dotfiles/
  └── dot-ssh/          ← Package root (with prefix)
      ├── config        ← Contents at root (flat)
      ├── known_hosts
      └── id_ed25519

~/.ssh → ~/dotfiles/dot-ssh/
```

### Dotfile Translation Rules

**Consistent across all items:**
- Package names: `.ssh` → `dot-ssh`
- File names: `.vimrc` → `dot-vimrc`
- Directory names: `.cache` → `dot-cache`
- Regular items: `config` → `config` (no change)

## Migration Steps

### Option 1: Clean Slate (Recommended)

```bash
# 1. Save your current package list
dot list > ~/dot-packages-backup.txt

# 2. Unmanage all packages (restores files to target)
for pkg in $(dot list --format=json | jq -r '.[].name'); do
  dot unmanage "$pkg"
done

# 3. Update dot binary
cd /path/to/dot
git pull origin refactor-dotprefix
make install

# 4. Verify new version
dot --version

# 5. Re-adopt with new structure
dot adopt .ssh .config .vim
```

### Option 2: Manual Restructuring

For each package with old structure:

```bash
# Example: Migrating ssh package
# Old: ~/dotfiles/ssh/dot-ssh/
# New: ~/dotfiles/dot-ssh/

# 1. Unmanage the package (removes symlinks)
dot unmanage ssh --no-restore

# 2. Restructure manually
mv ~/dotfiles/ssh/dot-ssh ~/dotfiles/dot-ssh-temp
mv ~/dotfiles/dot-ssh-temp/* ~/dotfiles/dot-ssh/
rmdir ~/dotfiles/dot-ssh-temp ~/dotfiles/ssh

# 3. Update manifest manually or re-manage
rm ~/.local/share/dot/manifest/.dot-manifest.json
dot manage dot-ssh
```

## Verification

After migration:

```bash
# Check structure
dot list
# Should show: dot-ssh, dot-vim, dot-config (with dot- prefix)

# Verify health
dot doctor
# Should show no issues

# Check package structure
ls -la ~/dotfiles/
# Should see: dot-ssh/, dot-vim/, dot-config/ (flat, no nesting)

# Verify symlinks work
ls -la ~/.ssh
# Should point to: ~/dotfiles/dot-ssh/
```

## New Features in v0.3

### Enhanced Adopt
- Auto-package naming: `dot adopt .ssh`
- Flat structure (no nesting)
- Cross-device support

### Enhanced Unmanage  
- `--cleanup`: Remove orphaned manifest entries
- `--purge`: Delete package directories
- `--no-restore`: Skip restoration
- Copy-based restoration (preserves files in package)

### Tab Completion
- Package name completion for all commands
- Config key completion
- Context-aware (installed vs available)

### Other Improvements
- Manifest directory configuration
- Enhanced config display (all 10 sections)
- Config precedence fixes
- List command shows context header

## Compatibility Notes

- **Manifest format:** Compatible (adds `source` field)
- **Package structure:** Incompatible (requires migration)
- **Commands:** Fully compatible (same syntax)

## Troubleshooting

### Orphaned Packages

If migration leaves orphaned entries:

```bash
# Clean up orphaned packages
for pkg in $(dot list --format=json | jq -r '.[].name'); do
  dot unmanage "$pkg" --cleanup
done
```

### Missing Files

If files are missing after migration:

```bash
# Check package directory manually
ls -la ~/dotfiles/

# Files should be at package root
ls -la ~/dotfiles/dot-ssh/
# Should show: config, known_hosts, etc. (not nested)
```

## Rollback

If you need to rollback to v0.2:

```bash
# 1. Unmanage all v0.3 packages
dot unmanage --all

# 2. Checkout v0.2 tag
git checkout v0.2.0
make install

# 3. Restore your old manifest backup if needed
```

## Support

For migration issues, see:
- [Troubleshooting Guide](08-troubleshooting.md)
- [GitHub Issues](https://github.com/yaklabco/dot/issues)
- [Developer Documentation](../developer/architecture.md)

