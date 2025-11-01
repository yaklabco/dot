# dot User Guide

Comprehensive guide to using dot for dotfile and configuration management.

## Table of Contents

### Getting Started

1. [Introduction and Core Concepts](01-introduction.md)
   - What is dot
   - Core terminology
   - When to use dot
   - Comparison with GNU Stow

2. [Installation Guide](02-installation.md)
   - Binary releases
   - Package managers
   - Building from source
   - Platform-specific notes
   - Verification

3. [Quick Start Tutorial](03-quickstart.md)
   - First package setup
   - Basic operations
   - Common workflows
   - Next steps

### Configuration and Usage

4. [Configuration Reference](04-configuration.md)
   - Configuration sources and precedence
   - File locations and formats
   - All configuration options
   - Per-package configuration
   - Environment variables

5. [Command Reference](05-commands.md)
   - Command structure
   - Global options
   - Clone command (repository setup)
   - Clone bootstrap (generate config)
   - Manage command
   - Unmanage command
   - Remanage command
   - Adopt command
   - Status command
   - Doctor command
   - List command
   - Exit codes

6. [Common Workflows](06-workflows.md)
   - Initial setup
   - Multi-machine synchronization
   - Package organization
   - Conflict resolution
   - Adoption workflow
   - CI/CD integration
   - Backup and recovery
   - Migration from GNU Stow

### Advanced Topics

7. [Advanced Features](07-advanced.md)
   - Ignore pattern system
   - Directory folding
   - Dry-run mode
   - Resolution policies
   - State management
   - Incremental operations
   - Parallel execution
   - Performance tuning

8. [Troubleshooting Guide](08-troubleshooting.md)
   - Common issues
   - Error messages
   - Diagnostic procedures
   - Recovery procedures
   - Platform-specific issues
   - FAQ

9. [Glossary](09-glossary.md)
   - Technical terms
   - Command terminology
   - Concept definitions

10. [Updates and Version Management](10-updates.md)
   - Upgrade command
   - Package manager configuration
   - Startup version checking
   - Pre-release versions
   - Troubleshooting updates
   - Security considerations

## Additional Resources

### For Developers

- [Release Workflow](../developer/release-workflow.md)
- [Contributing Guide](../../CONTRIBUTING.md)

### Examples

- [Basic Examples](../../examples/basic/)
- [Examples README](../../examples/README.md)

### Reference

- [Bootstrap Configuration Specification](bootstrap-config-spec.md)
- [Migration from GNU Stow](migration-from-stow.md)
- [Homebrew Installation](installation-homebrew.md)

## Getting Help

- **GitHub Issues**: Report bugs or request features
- **GitHub Discussions**: Ask questions and share experiences
- **Documentation**: Search this guide for answers
- **Man Pages**: `man dot` for quick reference

## Documentation Conventions

This guide uses the following conventions:

- **Code blocks**: Commands and code examples
- `inline code`: Commands, options, and file paths
- **Bold**: Important concepts and warnings
- *Italic*: Emphasis and variable names

### Command Examples

```bash
# Comments explain what commands do
dot manage vim

# Output examples shown where helpful
$ dot status
Package: vim
Status: installed
Links: 3
```

### Notes and Warnings

Important information is highlighted:

**Note**: Informational messages providing context.

**Warning**: Critical information about potential issues or data loss.

**Tip**: Helpful suggestions for efficient usage.

## Version Information

This documentation corresponds to dot v0.4.1.

For older versions, see the documentation in the appropriate release tag.

## Navigation

**[↑ Back to Main README](../../README.md)** | [Documentation Index](../README.md)

