# Automated Release Workflow

This document describes how to use the automated CI release workflow to cut releases and auto-update the changelog.

## Quick Start

### Option 1: GitHub Actions UI (Recommended)

1. Navigate to **Actions** tab in GitHub
2. Select **"Version Bump"** workflow from the left sidebar
3. Click **"Run workflow"** button (top right)
4. Select version bump type:
   - **patch** - Bug fixes and minor improvements (v0.4.3 → v0.4.4)
   - **minor** - New features, backward compatible (v0.4.3 → v0.5.0)
   - **major** - Breaking changes (v0.4.3 → v1.0.0)
5. Click **"Run workflow"**

The workflow will automatically:
- ✓ Run all quality checks (tests, linting, coverage)
- ✓ Generate updated CHANGELOG.md using git-chglog
- ✓ Create and push the release tag
- ✓ Trigger the release workflow
- ✓ Build binaries for all platforms
- ✓ Create GitHub Release with changelog notes
- ✓ Update Homebrew tap

### Option 2: GitHub CLI

```bash
# Patch release
gh workflow run version-bump.yml -f version_type=patch

# Minor release
gh workflow run version-bump.yml -f version_type=minor

# Major release
gh workflow run version-bump.yml -f version_type=major

# Watch the workflow
gh run watch
```

### Option 3: Local Release (Manual)

If you prefer to cut releases locally (not recommended for normal releases):

```bash
# Ensure you're on main and up to date
git checkout main
git pull origin main

# Run the appropriate version command
make version-patch  # or version-minor, version-major

# Push the changes and tag
git push origin main
git push --force origin vX.Y.Z  # Replace with your version

# The release workflow will trigger automatically
```

## How It Works

### Version Bump Workflow

When you trigger the version bump workflow:

1. **Quality Checks**: Runs full test suite, coverage checks, and linting
2. **Version Calculation**: Determines the next version based on current tags
3. **Changelog Generation**: Uses git-chglog to generate CHANGELOG.md from conventional commits
4. **Commit**: Commits the updated CHANGELOG.md to main
5. **Tag Creation**: Creates an annotated git tag
6. **Changelog Finalization**: Regenerates changelog with tag info and amends commit
7. **Push**: Pushes changes and tag to GitHub

### Release Workflow (Automatic)

When a tag matching `v*` is pushed:

1. **Quality Checks**: Verifies tests and linting pass
2. **Changelog Extraction**: Extracts release notes for this version from CHANGELOG.md
3. **GoReleaser**: Builds binaries for all platforms
4. **GitHub Release**: Creates release with extracted changelog
5. **Homebrew Tap**: Updates the Homebrew formula automatically
6. **Artifacts**: Uploads binaries as release assets

## Prerequisites

### For Automated Workflow

- GitHub Actions enabled
- `GITHUB_TOKEN` or `GORELEASER_TOKEN` secret configured (for Homebrew tap updates)
- Conventional commits for proper changelog generation

### For Local Releases

- `git-chglog` installed: `go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest`
- `golangci-lint` installed
- Clean working tree
- Up-to-date with `origin/main`

## Conventional Commits

The changelog is automatically generated from commit messages. Use these commit types:

**Included in changelog:**
- `feat:` - New features
- `fix:` - Bug fixes
- `perf:` - Performance improvements
- `refactor:` - Code refactoring
- `build:` - Build system changes
- `revert:` - Reverted changes

**Excluded from changelog:**
- `docs:` - Documentation updates
- `test:` - Test changes
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes
- `style:` - Code formatting

**Breaking changes:**
Add `BREAKING CHANGE:` in the commit footer for major version bumps.

Example:
```
feat(api): restructure configuration interface

The configuration API now uses a builder pattern.

BREAKING CHANGE: Configuration loading API changed.
Replace config.Load() with config.NewLoader().Load().
```

## Troubleshooting

### Workflow Failed During Quality Checks

If the version bump workflow fails during tests or linting:

1. Fix the issues locally
2. Commit and push to main
3. Re-run the workflow

### Tag Already Exists

If you need to recreate a tag:

```bash
# Delete local tag
git tag -d v0.4.4

# Delete remote tag
git push origin :refs/tags/v0.4.4

# Re-run the version bump workflow
```

### Changelog Missing Commits

Check your commit message format:
```bash
git log --oneline --format="%s" v0.4.3..HEAD
```

Ensure commits match pattern: `type(scope): description`

### Release Workflow Failed

1. Check the Actions tab for detailed logs
2. If it's a transient error, re-run the workflow
3. If it's a configuration issue, fix and delete/recreate the tag

## Comparison: Automated vs Local

| Aspect | Automated (CI) | Local (Make) |
|--------|----------------|--------------|
| Quality Checks | ✓ Always enforced | ⚠️ Manual |
| Consistency | ✓ Reproducible | ⚠️ Environment dependent |
| Audit Trail | ✓ GitHub Actions logs | ❌ Local only |
| Convenience | ✓ Click and go | ⚠️ Multiple steps |
| Speed | ~3-5 minutes | ~2-3 minutes |
| Control | ⚠️ Standardized | ✓ Full control |

**Recommendation**: Use automated workflow for standard releases, local workflow only for hotfixes or special cases.

## Related Documentation

- [Release Workflow (Detailed)](../../docs/developer/release-workflow.md)
- [Contributing Guidelines](../../CONTRIBUTING.md)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [git-chglog Documentation](https://github.com/git-chglog/git-chglog)

## Quick Reference

```bash
# Start a release from GitHub Actions UI
Go to Actions → Version Bump → Run workflow

# Or use GitHub CLI
gh workflow run version-bump.yml -f version_type=patch

# Check workflow status
gh run watch

# View releases
gh release list

# View latest release
gh release view
```

