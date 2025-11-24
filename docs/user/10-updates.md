# Updates and Version Management

dot includes built-in functionality to check for and install updates to the application itself.

## Upgrade Command

The `dot upgrade` command checks for new versions from GitHub and can automatically upgrade dot using your system's package manager.

### Basic Usage

```bash
# Check for and install updates interactively
dot upgrade

# Check for updates without installing
dot upgrade --check-only

# Upgrade without confirmation prompt
dot upgrade --yes
```

### How It Works

The upgrade command:

1. Queries the GitHub releases API for the latest version
2. Compares the latest version with your current version
3. Displays release information including version numbers and release notes
4. Offers to upgrade using your configured package manager
5. Executes the appropriate package manager command to perform the upgrade

### Package Managers

dot supports the following package managers for automated upgrades:

- **brew** (Homebrew) - macOS and Linux
- **apt** - Debian/Ubuntu
- **yum** - RHEL/CentOS (legacy)
- **dnf** - Fedora/RHEL 8+
- **pacman** - Arch Linux
- **zypper** - openSUSE
- **manual** - Download from GitHub releases

The package manager is auto-detected by default, but can be explicitly configured.

### Configuration

Upgrade behavior is configured in `~/.config/dot/config.yaml`:

```yaml
update:
  # Enable automatic version checking at startup
  check_on_startup: false

  # Frequency of version checks in hours (0 = always check)
  check_frequency: 24

  # Package manager to use: auto, brew, apt, yum, pacman, dnf, zypper, manual
  package_manager: auto

  # GitHub repository for releases
  repository: yaklabco/dot

  # Include pre-release versions
  include_prerelease: false
```

### Package Manager Configuration

#### Automatic Detection

By default, dot automatically detects your package manager:

```yaml
update:
  package_manager: auto
```

On macOS, Homebrew is preferred if available. On Linux, dot checks for package managers in this order: dnf, yum, apt, pacman, zypper.

#### Explicit Configuration

To use a specific package manager:

```yaml
update:
  package_manager: brew
```

#### Manual Upgrades

If your package manager is not supported or you installed dot manually:

```yaml
update:
  package_manager: manual
```

When set to manual, `dot upgrade` will display the GitHub releases URL and instructions for manual download.

## Startup Version Checking

dot can automatically check for new versions at startup and display a notification if an update is available.

### Enabling Startup Checks

Enable automatic version checking in your configuration:

```yaml
update:
  check_on_startup: true
  check_frequency: 24
```

### Check Frequency

The `check_frequency` setting controls how often dot checks for updates:

- **24** (default) - Check once per day
- **168** - Check once per week
- **0** - Check on every invocation (not recommended)
- **-1** or disabled - Never check automatically

### Notification Display

When an update is available, dot displays a notification box:

```
╭─────────────────────────────────────────────────────────────╮
│  A new version of dot is available!                        │
│                                                             │
│  Current: v0.3.0                                            │
│  Latest:  v0.4.0                                            │
│                                                             │
│  Run 'dot upgrade' to update                                │
╰─────────────────────────────────────────────────────────────╯
```

The notification:
- Appears before command execution
- Never blocks or delays your command
- Only shows once per check frequency period
- Silently fails if network is unavailable

### Disabling Startup Checks

To disable automatic update checks:

```yaml
update:
  check_on_startup: false
```

Or use check frequency -1:

```yaml
update:
  check_on_startup: true
  check_frequency: -1
```

## Pre-Release Versions

To include pre-release versions (beta, alpha, rc) in update checks:

```yaml
update:
  include_prerelease: true
```

When enabled:
- `dot upgrade --check-only` will show pre-release versions
- `dot upgrade` will offer to install pre-release versions
- Startup notifications will include pre-release versions

**Note:** Pre-release versions may contain bugs and are not recommended for production use.

## Custom Repository

If you maintain a fork or custom build of dot, configure the repository:

```yaml
update:
  repository: your-username/dot
```

The repository must follow the same release structure as the official dot repository.

## Troubleshooting

### Upgrade Command Not Available

If `dot upgrade` is not available, your version of dot may be too old. Update manually:

**Homebrew:**
```bash
brew upgrade dot
```

**Manual:**
Download the latest release from https://github.com/yaklabco/dot/releases

### Network Errors

Version checking requires internet access to GitHub's API. If you're behind a proxy or firewall:

- Set appropriate HTTP_PROXY/HTTPS_PROXY environment variables
- Or disable automatic checks with `check_on_startup: false`

### Package Manager Not Found

If dot cannot find your package manager:

1. Verify the package manager is installed
2. Ensure it's in your PATH
3. Or set `package_manager: manual` and upgrade manually

### Permission Errors

Package manager upgrades typically require sudo access. If you get permission errors:

- Ensure your user has sudo privileges
- Or use `package_manager: manual` for user-space installations

## Security Considerations

### GitHub API Rate Limits

GitHub's API has rate limits for unauthenticated requests:
- 60 requests per hour per IP address
- Startup checks count toward this limit
- Recommendation: Use `check_frequency: 24` or higher

### Verification

The upgrade command:
- Uses HTTPS for all GitHub API requests
- Delegates actual download and installation to your package manager
- Never downloads or executes arbitrary code directly

Your package manager handles:
- Download verification
- Signature checking
- Secure installation

### Privacy

Version checking:
- Makes HTTPS requests to https://api.github.com
- Sends standard HTTP headers (User-Agent: dot-updater)
- Does not transmit any personal information
- Does not track usage or analytics

## Best Practices

1. **Enable startup checks for workstations**
   ```yaml
   check_on_startup: true
   check_frequency: 168  # Weekly
   ```

2. **Disable for CI/CD environments**
   ```yaml
   check_on_startup: false
   ```

3. **Stay on stable releases**
   ```yaml
   include_prerelease: false
   ```

4. **Use appropriate check frequency**
- Workstations: 24-168 hours (daily to weekly)
- Servers: 168+ hours (weekly or longer)
- CI/CD: disabled

5. **Test upgrades in development first**
   ```bash
   dot upgrade --check-only  # See what's available
   dot upgrade               # Upgrade in dev environment
   # Test thoroughly before upgrading production
   ```

## Examples

### Check for Updates

```bash
# See if update is available
$ dot upgrade --check-only

A new version is available!

  Current version:  v0.3.0
  Latest version:   v0.4.0
  Release URL:      https://github.com/yaklabco/dot/releases/tag/v0.4.0

Release Notes:
  - Add upgrade command
  - Improve performance
  ...

Run dot upgrade to upgrade.
```

### Interactive Upgrade

```bash
$ dot upgrade

Checking for updates...

ⓘ A new version is available!

  Current version:  v0.3.0
  Latest version:   v0.4.0
  Release URL:      https://github.com/yaklabco/dot/releases/tag/v0.4.0

Release Notes:
  - Add upgrade command
  - Improve performance

Package manager: brew
Upgrade command: brew upgrade dot

Do you want to upgrade now? [y/N]: y

→ Upgrading...

==> Downloading https://homebrew.bintray.com/bottles/dot-0.4.0.bottle.tar.gz
...
✓ Upgrade completed successfully!
Run dot --version to verify the new version.
```

### Automated Upgrade

```bash
# Non-interactive upgrade (useful for scripts)
dot upgrade --yes
```

### Configuration Example

Complete update configuration:

```yaml
# ~/.config/dot/config.yaml
update:
  # Check for updates at startup
  check_on_startup: true
  
  # Check weekly
  check_frequency: 168
  
  # Use Homebrew on macOS
  package_manager: auto
  
  # Official repository
  repository: yaklabco/dot
  
  # Stable releases only
  include_prerelease: false
```

