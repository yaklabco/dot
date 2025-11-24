# Release Workflow

## Overview

The dot project uses [Release Please](https://github.com/googleapis/release-please) for automated versioning and changelog generation. This ensures that:

1. Releases are created automatically via Pull Requests
2. Changelogs are accurate and generated from [Conventional Commits](https://www.conventionalcommits.org/)
3. Versions strictly follow [Semantic Versioning](https://semver.org/)

## Prerequisites

- Git commit messages **must** follow the Conventional Commits specification.
- All tests (`make test`) and linters (`make lint`) must pass before merging.

## Release Process

### 1. Develop and Merge Features

Work on features or fixes in feature branches. When committing, use conventional commit messages:

- `feat: add new capability` (triggers minor version bump)
- `fix: resolve issue` (triggers patch version bump)
- `feat!: breaking change` (triggers major version bump)
- `chore: maintenance` (no release trigger usually)

Merge these PRs into the `main` branch.

### 2. Automated Release PR

When changes are pushed to `main`, the Release Please workflow runs.

- It analyzes the commits since the last release.
- It creates or updates a **Release PR**.
- This PR contains:
- The calculated next version.
- The generated `CHANGELOG.md` updates.
- Updates to `.release-please-manifest.json`.

### 3. Review and Release

1. **Review the Release PR**: Check the changelog and version bump.
2. **Merge the Release PR**: When you are ready to release, simply merge this PR into `main`.

### 4. Automated Tagging and Publishing

Once the Release PR is merged:

1. Release Please automatically creates a git tag (e.g., `v1.2.3`).
2. This tag push triggers the **Release** workflow (`.github/workflows/release.yml`).
3. GoReleaser builds the binaries and publishes the GitHub Release.
4. Homebrew taps are updated automatically.

## Local Verification

You can verify the state of the repository locally using:

```bash
make check
```

To preview the changelog for the next version locally (optional):

```bash
make changelog-next
```

## Troubleshooting

### Release PR not created

- Check if the workflow `.github/workflows/release-please.yml` failed.
- Ensure your commits on `main` contain releasable types (feat, fix). `chore` or `docs` might not trigger a release depending on configuration.

### Release not published

- Check the `Release` workflow in GitHub Actions.
- Ensure the tag was created successfully by Release Please.

## Manual Hotfixes

Ideally, hotfixes should follow the normal PR process:
1. Create a branch `fix/critical-issue`.
2. Commit `fix: resolve critical issue`.
3. Merge to `main`.
4. Merge the resulting Release PR.

If a manual override is absolutely required, you can manually tag `main`, but this is discouraged as it desyncs the Release Please manifest.
