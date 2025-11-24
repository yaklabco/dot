# Homebrew Installation Guide

This guide covers installation and management of dot using Homebrew on macOS and Linux.

## Prerequisites

### macOS

Homebrew is available on macOS 10.15 (Catalina) or later. Install Homebrew if not already present:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Linux

Homebrew is supported on most Linux distributions. Install via:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Follow the post-installation instructions to add Homebrew to your PATH.

## Installation

### Add Tap

The dot formula is hosted in a custom tap. Add the tap to your Homebrew:

```bash
brew tap yaklabco/dot
```

This registers the tap and makes the dot formula available for installation.

### Install dot

Install the dot CLI tool:

```bash
brew install dot
```

Homebrew automatically:
- Downloads the appropriate binary for your platform (Intel or Apple Silicon on macOS, x86_64 or ARM64 on Linux)
- Installs the binary to `/usr/local/bin/dot` (Intel) or `/opt/homebrew/bin/dot` (Apple Silicon)
- Verifies the installation by running basic tests

### Verify Installation

Confirm installation success:

```bash
dot --version
```

Expected output format:

```
dot version v0.x.x
```

Test basic functionality:

```bash
dot --help
```

## Platform-Specific Notes

### macOS Intel

Binary location: `/usr/local/bin/dot`

The Homebrew prefix is `/usr/local`. Ensure this directory is in your PATH:

```bash
export PATH="/usr/local/bin:$PATH"
```

### macOS Apple Silicon

Binary location: `/opt/homebrew/bin/dot`

The Homebrew prefix is `/opt/homebrew`. Ensure this directory is in your PATH:

```bash
export PATH="/opt/homebrew/bin:$PATH"
```

### Linux x86_64

Binary location typically: `/home/linuxbrew/.linuxbrew/bin/dot`

Add Homebrew to your PATH in your shell configuration file:

```bash
# ~/.bashrc or ~/.zshrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
```

### Linux ARM64

Binary location typically: `/home/linuxbrew/.linuxbrew/bin/dot`

Configuration is identical to x86_64 Linux installations.

## Updates

### Check for Updates

View available updates:

```bash
brew outdated dot
```

### Update dot

Update to the latest version:

```bash
brew update
brew upgrade dot
```

The `brew update` command refreshes the tap repository, ensuring the latest formula is available. The `brew upgrade` command installs the new version.

### Update All Packages

Update all Homebrew packages including dot:

```bash
brew update
brew upgrade
```

## Uninstallation

### Remove dot

Uninstall the dot binary:

```bash
brew uninstall dot
```

This removes the binary but preserves your configuration files and package repositories.

### Remove Tap

Remove the tap entirely:

```bash
brew untap yaklabco/dot
```

This removes the tap registration but does not affect installed binaries unless they are also uninstalled.

### Clean Up Configuration

Manually remove configuration files if desired:

```bash
rm -rf ~/.config/dot
rm -f ~/.dotrc
```

## Troubleshooting

### Command Not Found

If `dot` is not found after installation:

1. Verify installation:
   ```bash
   brew list dot
   ```

2. Check binary location:
   ```bash
   which dot
   ```

3. Ensure Homebrew bin directory is in PATH:
   ```bash
   echo $PATH | grep -o "[^:]*brew[^:]*"
   ```

4. Reload shell configuration:
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

### Permission Denied

If installation fails with permission errors:

1. Check Homebrew directory ownership:
   ```bash
   ls -ld $(brew --prefix)
   ```

2. Fix ownership if necessary (macOS):
   ```bash
   sudo chown -R $(whoami) /usr/local/Cellar /usr/local/Homebrew
   ```

3. Never use `sudo` with `brew install`

### Checksum Mismatch

If installation fails with checksum errors:

1. Update Homebrew and retry:
   ```bash
   brew update
   brew install dot
   ```

2. Clear Homebrew cache:
   ```bash
   brew cleanup
   rm -rf $(brew --cache)
   brew install dot
   ```

### Wrong Architecture

If Homebrew installs the wrong binary architecture:

1. Verify system architecture:
   ```bash
   uname -m
   ```

2. Check installed package info:
   ```bash
   brew info dot
   ```

3. Reinstall with architecture specification:
   ```bash
   brew uninstall dot
   arch -arm64 brew install dot  # Apple Silicon
   arch -x86_64 brew install dot  # Intel
   ```

### Formula Not Found

If `brew install dot` reports formula not found:

1. Verify tap is added:
   ```bash
   brew tap
   ```

2. Re-add tap if missing:
   ```bash
   brew tap yaklabco/dot
   ```

3. Update tap:
   ```bash
   brew update
   ```

### Version Mismatch

If `dot --version` shows unexpected version:

1. Check which binary is being executed:
   ```bash
   which dot
   ```

2. Check for multiple installations:
   ```bash
   find /usr/local /opt/homebrew -name dot 2>/dev/null
   ```

3. Uninstall and reinstall:
   ```bash
   brew uninstall dot
   brew install dot
   ```

## Configuration

After installation, configure dot for your environment:

### Configuration File

Create user configuration file:

```bash
mkdir -p ~/.config/dot
cat > ~/.config/dot/config.yaml << 'EOF'
packageDir: ~/dotfiles
targetDir: ~
linkMode: relative
folding: true
verbosity: 0

ignore:
  - "*.log"
  - ".git"
  - ".DS_Store"
  - "*.swp"
EOF
```

### Environment Variables

Set environment variables for per-session configuration:

```bash
export DOT_PACKAGE_DIR="$HOME/dotfiles"
export DOT_TARGET_DIR="$HOME"
export DOT_VERBOSITY=1
```

Add to shell configuration file (`.bashrc`, `.zshrc`) for persistence.

### Verification

Test configuration:

```bash
dot status
```

If configured correctly, this displays installation status without errors.

## Integration with Dotfiles Repository

### Initial Setup

Create dotfiles repository structure:

```bash
mkdir -p ~/dotfiles/{vim,tmux,zsh}
cd ~/dotfiles
git init
```

### Package Installation

Install packages from dotfiles repository:

```bash
cd ~/dotfiles
dot manage vim tmux zsh
```

### Status Verification

Verify package status:

```bash
dot status
```

Expected output shows installed packages with symlink counts and status.

## Next Steps

- Read [Quick Start Guide](03-quickstart.md) for usage examples
- Configure dot via [Configuration Reference](04-configuration.md)
- Explore commands in [Command Reference](05-commands.md)
- Review workflows in [Common Workflows](06-workflows.md)

## Support

For issues specific to Homebrew installation:

1. Check [Troubleshooting Guide](08-troubleshooting.md)
2. Search [GitHub Issues](https://github.com/yaklabco/dot/issues)
3. Report bugs with `brew gist-logs dot` output

For general dot usage questions, see the main [documentation](../../README.md).

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

