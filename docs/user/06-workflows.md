# Common Workflows

Real-world usage patterns and workflows for dot.

## Initial Dotfiles Setup

### Scenario: First-Time Setup

Starting from scratch with no existing dotfiles repository.

**Steps**:

1. Create dotfiles repository
2. Create first package
3. Install package
4. Commit to version control

```bash
# Create repository
mkdir ~/dotfiles
cd ~/dotfiles
git init

# Create vim package
mkdir vim
cat > vim/dot-vimrc << 'EOF'
set number
syntax on
EOF

# Install
dot manage vim

# Commit
git add vim/
git commit -m "feat(vim): add initial configuration"
git remote add origin https://github.com/username/dotfiles.git
git push -u origin main
```

## Multi-Machine Synchronization

### New Machine Setup

```bash
# Clone dotfiles
git clone https://github.com/username/dotfiles.git ~/dotfiles
cd ~/dotfiles

# Install packages
dot manage vim zsh tmux

# Verify
dot status
```

### Keeping Machines in Sync

On machine with changes:
```bash
cd ~/dotfiles
vim vim/dot-vimrc
dot remanage vim
git add vim/
git commit -m "feat(vim): update configuration"
git push
```

On other machines:
```bash
cd ~/dotfiles
git pull
dot remanage vim zsh tmux
```

## Conflict Resolution

### Backup and Compare

```bash
# Conflicts detected
dot manage vim
# Error: conflict at ~/.vimrc

# Backup existing
dot --on-conflict backup manage vim

# Compare versions
diff ~/.vimrc.bak ~/dotfiles/vim/dot-vimrc

# Merge if needed
vim ~/dotfiles/vim/dot-vimrc
dot remanage vim
```

### Adopt Existing Files

**Single File Adoption**:

```bash
# Conflict exists
dot manage vim
# Error: conflict at ~/.vimrc

# Adopt instead (explicit package name)
dot adopt vim ~/.vimrc

# Edit in package
vim ~/dotfiles/vim/dot-vimrc
git add vim/
git commit -m "feat(vim): adopt existing configuration"
```

**Auto-Naming (Single File)**:

```bash
# Let dot derive package name from filename
dot adopt ~/.vimrc
# Creates package: dot-vimrc

# Or for a directory
dot adopt ~/.ssh
# Creates package: dot-ssh
```

**Glob Expansion (Multiple Related Files)**:

```bash
# Adopt all git-related config files into single package
dot adopt .git*
# Shell expands to: .gitconfig .gitignore .git-credentials
# Creates package: dot-git
# All files adopted into one package

# Or with zsh configs
dot adopt .zsh*
# Adopts: .zshrc .zshenv .zprofile
# Creates package: dot-zsh

# Commit
git add dot-git/ dot-zsh/
git commit -m "feat(git,zsh): adopt existing configurations"
```

The glob workflow is useful for adopting multiple related configuration files that share a common prefix, keeping them organized in a single package.

## Testing New Packages

### Dry-Run Testing

```bash
# Create package
mkdir ~/dotfiles/test-package
echo "test" > ~/dotfiles/test-package/dot-testrc

# Preview
dot --dry-run manage test-package

# Install if satisfied
dot manage test-package

# Remove if not needed
dot unmanage test-package
rm -rf ~/dotfiles/test-package
```

## Package Updates

### Update Single Package

```bash
# Edit configuration
vim ~/dotfiles/vim/dot-vimrc

# Preview changes
dot --dry-run remanage vim

# Apply
dot remanage vim

# Verify
dot status vim

# Commit
git add vim/
git commit -m "feat(vim): add plugins"
git push
```

### Bulk Updates

```bash
# Pull changes
cd ~/dotfiles
git pull

# Update all packages (incremental)
dot remanage vim zsh tmux git

# Verify
dot doctor
```

## Backup and Recovery

### Create Backup

```bash
# Backup before changes
tar -czf ~/dotfiles-backup-$(date +%Y%m%d).tar.gz ~/dotfiles ~/.dot-manifest.json

# Make changes
# ...

# Restore if needed
tar -xzf ~/dotfiles-backup-20251007.tar.gz
dot remanage vim zsh
```

### Disaster Recovery

```bash
# Reinstall dot
curl -L https://github.com/yaklabco/dot/releases/latest/download/dot-Linux-x86_64.tar.gz | tar xz
sudo mv dot /usr/local/bin/

# Clone dotfiles
git clone https://github.com/username/dotfiles.git ~/dotfiles

# Reinstall all
cd ~/dotfiles
dot manage $(ls -d */ | tr -d '/')

# Verify
dot status
dot doctor
```

## Migration from GNU Stow

### Gradual Migration

```bash
# Phase 1: Test with one package
stow -D test-package
dot manage test-package

# Phase 2: Migrate all
cd ~/dotfiles
for pkg in */; do stow -D "$pkg"; done
dot manage $(ls -d */ | tr -d '/')

# Phase 3: Remove Stow
brew uninstall stow
```

## CI/CD Integration

### GitHub Actions Example

`.github/workflows/deploy.yml`:
```yaml
name: Deploy Dotfiles
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install dot
        run: |
          curl -L https://github.com/yaklabco/dot/releases/latest/download/dot-Linux-x86_64.tar.gz | tar xz
          sudo mv dot /usr/local/bin/
      - name: Deploy
        run: dot --quiet manage vim zsh tmux
      - name: Verify
        run: dot doctor
```

## Next Steps

- [Advanced Features](07-advanced.md): Deep dive into features
- [Troubleshooting Guide](08-troubleshooting.md): Solve common issues
- [Configuration Reference](04-configuration.md): Customize behavior

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

