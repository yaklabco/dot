# Troubleshooting Guide

Solutions for common issues and problems.

## Installation Issues

### Command Not Found

**Problem**: `dot: command not found`

**Diagnosis**:
```bash
which dot
echo $PATH
```

**Solutions**:

1. **Binary not in PATH**:
```bash
# Add to PATH
export PATH="$PATH:/usr/local/bin"

# Make permanent
echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc
source ~/.bashrc
```

2. **Binary not installed**:
```bash
# Verify installation location
ls -la /usr/local/bin/dot

# Reinstall if missing
curl -L https://github.com/yaklabco/dot/releases/latest/download/dot-$(uname -s)-$(uname -m).tar.gz | tar xz
sudo mv dot /usr/local/bin/
```

### Permission Denied

**Problem**: `permission denied: dot`

**Solution**:
```bash
# Add execute permission
chmod +x /usr/local/bin/dot
```

### Version Mismatch

**Problem**: Multiple dot installations with different versions

**Diagnosis**:
```bash
which -a dot
```

**Solution**:
```bash
# Remove old versions
sudo rm /path/to/old/dot

# Verify
dot --version
```

## Operation Failures

### Symlink Creation Failed

**Problem**: Cannot create symlinks

**Common Causes**:

1. **No symlink support** (FAT32, exFAT):
```bash
# Check filesystem
df -T .
```

**Solution**: Use filesystem with symlink support (ext4, APFS, etc.)

2. **Permission denied**:
```bash
# Check permissions
ls -la ~/

# Fix permissions
chmod u+w ~/
```

3. **Windows without privileges**:

**Solution**: Enable Developer Mode or run as administrator

### File Conflicts

**Problem**: `Error: conflict at ~/.vimrc`

**Diagnosis**:
```bash
# Check what exists
ls -la ~/.vimrc

# If symlink, check target
readlink ~/.vimrc
```

**Solutions**:

1. **Backup existing file**:
```bash
dot --on-conflict backup manage vim
diff ~/.vimrc.bak ~/dotfiles/vim/dot-vimrc
```

2. **Adopt existing file**:
```bash
dot adopt vim ~/.vimrc
```

3. **Remove conflicting file**:
```bash
rm ~/.vimrc
dot manage vim
```

4. **Skip conflicts**:
```bash
dot --on-conflict skip manage vim
```

### Package Not Found

**Problem**: `Error: package not found: vim`

**Diagnosis**:
```bash
# Check package directory exists
ls -la ~/dotfiles/vim

# Check current directory
pwd
```

**Solutions**:

1. **Wrong directory**:
```bash
cd ~/dotfiles
dot manage vim
```

2. **Specify directory**:
```bash
dot --dir ~/dotfiles manage vim
```

3. **Package doesn't exist**:
```bash
# Create package
mkdir ~/dotfiles/vim
```

### Broken Symlinks

**Problem**: Symlinks point to non-existent targets or were accidentally deleted

**Diagnosis**:
```bash
# Check for broken links with doctor
dot doctor

# Find broken links manually
find ~ -xtype l

# Check specific link
ls -la ~/.vimrc
readlink ~/.vimrc
```

**Solutions**:

1. **Recreate missing links** (Recommended):
```bash
# Check what's broken
dot doctor

# Remanage automatically detects and recreates missing links
dot remanage vim

# For multiple packages
dot remanage vim zsh tmux
```

The `remanage` command now automatically detects missing symlinks and recreates them, even if the package content hasn't changed.

2. **Fix target path if package moved**:
```bash
# Update package location in config
dot config set directories.package /new/path

# Recreate links
dot remanage vim
```

3. **Remove and recreate** (if remanage doesn't fix it):
```bash
dot unmanage vim
dot manage vim
```

## Manifest Issues

### Manifest Corrupted

**Problem**: `Error: cannot parse manifest`

**Diagnosis**:
```bash
# Check manifest
cat ~/.dot-manifest.json

# Validate JSON
jq . ~/.dot-manifest.json
```

**Solutions**:

1. **Repair from filesystem**:
```bash
# Rebuild manifest
dot doctor --repair
```

2. **Delete and recreate**:
```bash
# Backup first
cp ~/.dot-manifest.json ~/.dot-manifest.json.bak

# Remove
rm ~/.dot-manifest.json

# Reinstall packages
dot manage vim zsh tmux
```

### Manifest Out of Sync

**Problem**: Manifest doesn't match filesystem

**Diagnosis**:
```bash
dot doctor
```

**Solution**:
```bash
# Sync manifest with filesystem
dot doctor --repair

# Or remanage all packages
dot remanage $(dot list --format json | jq -r '.[].name')
```

## Configuration Issues

### Configuration Not Loading

**Problem**: Configuration settings not applied

**Diagnosis**:
```bash
# Show active configuration
dot config show

# Show configuration file path
dot config path

# Verify file exists
ls -la $(dot config path)
```

**Solutions**:

1. **Check syntax**:
```bash
dot config validate
```

2. **Check precedence**:
```bash
# Environment variables override config file
unset DOT_VERBOSITY

# Command-line flags override everything
dot manage vim  # Uses config file
dot -v manage vim  # -v overrides config
```

### Invalid Configuration

**Problem**: `Error: invalid configuration`

**Diagnosis**:
```bash
# Validate syntax
dot config validate
```

**Common Errors**:

1. **Invalid YAML**:
```yaml
# Wrong: missing colon
packageDir ~/dotfiles

# Correct
packageDir: ~/dotfiles
```

2. **Invalid paths**:
```yaml
# Wrong: relative path
packageDir: dotfiles

# Correct: absolute or tilde-expanded
packageDir: ~/dotfiles
```

3. **Invalid values**:
```yaml
# Wrong: invalid option
linkMode: mixed

# Correct: valid values only
linkMode: relative  # or absolute
```

## Performance Issues

### Slow Operations

**Problem**: Operations take too long

**Diagnosis**:
```bash
# Profile operation
time dot -vv manage vim
```

**Solutions**:

1. **Enable incremental**:
```yaml
enableIncremental: true
```

2. **Increase concurrency**:
```yaml
concurrency: 8
```

3. **Optimize ignore patterns**:
```yaml
# Remove overly complex regex patterns
ignore:
  - ".git"
  - "node_modules"
  # Avoid: "/^test.*complex.*regex$/"
```

4. **Enable folding**:
```yaml
folding: true
```

### High Memory Usage

**Problem**: dot consumes excessive memory

**Diagnosis**:
```bash
# Monitor memory during operation
/usr/bin/time -v dot manage large-package
```

**Solutions**:

1. **Reduce concurrency**:
```yaml
concurrency: 2
```

2. **Process packages in batches**:
```bash
# Instead of all at once
dot manage pkg1 pkg2 pkg3
dot manage pkg4 pkg5 pkg6
```

## Platform-Specific Issues

### macOS Issues

**Problem**: Gatekeeper blocks execution

**Solution**:
```bash
# Remove quarantine
xattr -d com.apple.quarantine /usr/local/bin/dot

# Or allow via System Preferences
```

**Problem**: Symlinks broken after upgrade

**Solution**:
```bash
# Remanage all packages
cd ~/dotfiles
dot remanage $(dot list --format json | jq -r '.[].name')
```

### Windows Issues

**Problem**: Symlink creation fails

**Solutions**:

1. **Enable Developer Mode**:
   - Settings → Update & Security → For Developers
   - Enable Developer Mode
   - Restart terminal

2. **Run as administrator**:
   - Right-click terminal
   - "Run as administrator"

3. **Check symlink support**:
```powershell
# Test symlink creation
New-Item -ItemType SymbolicLink -Path test -Target C:\Windows\System32
```

### Linux Issues

**Problem**: SELinux blocking operations

**Solution**:
```bash
# Allow symlinks
sudo setsebool -P allow_user_symlink_target 1
```

**Problem**: Permission denied on NFS

**Solution**:
```bash
# Check NFS mount options
mount | grep nfs

# Ensure no_root_squash option if needed
```

## Error Messages

### Common Errors and Solutions

#### "conflict at PATH: file exists"

**Cause**: File exists at target location

**Solutions**:
- Use `--on-conflict backup` to preserve existing file
- Use `dot adopt` to move file into package
- Remove conflicting file manually

#### "package not found: NAME"

**Cause**: Package directory doesn't exist

**Solutions**:
- Check package directory exists: `ls ~/dotfiles/NAME`
- Specify correct package directory: `--dir ~/dotfiles`
- Create package: `mkdir ~/dotfiles/NAME`

#### "permission denied"

**Cause**: Insufficient permissions

**Solutions**:
- Check file/directory permissions
- Ensure write access to target directory
- Run with appropriate privileges (not sudo unless necessary)

#### "manifest corrupted"

**Cause**: Invalid JSON in manifest file

**Solutions**:
- Repair: `dot doctor --repair`
- Delete and recreate: `rm ~/.dot-manifest.json && dot manage ...`

#### "broken symlink"

**Cause**: Symlink target doesn't exist

**Solutions**:
- Remanage package: `dot remanage PACKAGE`
- Fix target location
- Remove and reinstall: `dot unmanage PACKAGE && dot manage PACKAGE`

## Diagnostic Procedures

### Health Check Procedure

```bash
# 1. Run doctor
dot doctor

# 2. Check configuration
dot config validate

# 3. Verify installation
dot --version

# 4. Check packages
dot status

# 5. List installed
dot list

# 6. Test with dry-run
dot --dry-run manage test-package
```

### Debug Procedure

```bash
# 1. Enable maximum verbosity
dot -vvv manage vim 2>&1 | tee debug.log

# 2. Check system state
ls -la ~/dotfiles/
ls -la ~/

# 3. Check manifest
cat ~/.dot-manifest.json | jq .

# 4. Verify symlinks
find ~ -type l -ls

# 5. Check configuration
dot config show --with-sources
```

### Recovery Procedure

```bash
# 1. Backup current state
tar -czf ~/dot-backup-$(date +%Y%m%d).tar.gz ~/dotfiles ~/.dot-manifest.json

# 2. Unmanage all packages
dot unmanage $(dot list --format json | jq -r '.[].name')

# 3. Clean manifest
rm ~/.dot-manifest.json

# 4. Reinstall
cd ~/dotfiles
dot manage $(ls -d */ | tr -d '/')

# 5. Verify
dot doctor
```

## Getting Help

### Information to Provide

When reporting issues, include:

1. **Version information**:
```bash
dot --version
```

2. **Configuration**:
```bash
dot config show
```

3. **Error output**:
```bash
dot -vvv <command> 2>&1
```

4. **System information**:
```bash
uname -a
echo $SHELL
```

5. **Directory structure** (sanitized):
```bash
tree -L 2 ~/dotfiles
```

### Where to Get Help

- **Documentation**: Search this guide
- **GitHub Issues**: Report bugs
- **GitHub Discussions**: Ask questions
- **Doctor command**: `dot doctor` for automated diagnostics

## FAQ

**Q: Can I use dot with GNU Stow simultaneously?**

A: Yes, but not recommended. They may conflict. Migrate completely or keep separate package sets.

**Q: Does dot work on Windows?**

A: Limited support. Requires Developer Mode or administrator privileges. Symlink support varies by filesystem.

**Q: Can I use absolute and relative links together?**

A: Yes, configure per-package in `.dotmeta` files.

**Q: What happens if I move my dotfiles directory?**

A: Relative links break. Remanage all packages: `dot remanage $(dot list --format json | jq -r '.[].name')`

**Q: Can I nest packages?**

A: No, packages must be top-level directories in package directory.

**Q: Does dot follow symlinks in packages?**

A: No, symlinks in packages are copied as symlinks to target.

## Next Steps

- [Glossary](09-glossary.md): Reference for technical terms
- [Command Reference](05-commands.md): Complete command documentation
- [Configuration Reference](04-configuration.md): Configuration options

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

