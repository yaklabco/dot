# Contributing to dot

Thank you for your interest in contributing to dot. This guide outlines the contribution process and project standards.

## Code of Conduct

Be respectful, professional, and collaborative. Harassment and unprofessional behavior are not tolerated.

## Getting Started

### Prerequisites

- Go 1.25 or later
- Git
- Make
- golangci-lint v2

### Development Setup

1. **Fork and clone**:
```bash
git fork https://github.com/jamesainslie/dot.git
cd dot
```

2. **Verify environment**:
```bash
go version  # Should be 1.25+
make check  # Runs tests, linting, build
```

3. **Create feature branch**:
```bash
git checkout -b feature-description
```

## Development Workflow

### Test-Driven Development

dot follows strict TDD. Tests must be written before implementation.

**Process**:
1. Write failing test describing desired behavior
2. Run test to verify it fails for right reason (red)
3. Implement minimum code to pass test
4. Run test to verify it passes (green)
5. Refactor while keeping tests green
6. Commit with atomic, conventional commit message

**Example**:

```bash
# 1. Write test
vim internal/scanner/scanner_test.go
# Add TestScanPackage function

# 2. Run test (should fail)
make test

# 3. Implement
vim internal/scanner/scanner.go
# Add ScanPackage function

# 4. Run test (should pass)
make test

# 5. Refactor if needed
# ... improve code ...

# 6. Commit
git add internal/scanner/
git commit -m "feat(scanner): add package scanning functionality"
```

### Testing Requirements

- **Minimum 75% coverage** for new code
- **Unit tests** for all functions
- **Integration tests** for complete workflows
- **Property-based tests** for core algorithms
- **Table-driven tests** for multiple scenarios

Run tests:
```bash
make test              # All tests
make test-unit         # Unit tests only
make test-integration  # Integration tests only
make test-coverage     # With coverage report
```

### Linting and Formatting

All code must pass linting without warnings.

**Run linters**:
```bash
make lint       # All linters
make lint-go    # Go linters only
make fmt        # Format code
```

**Configured linters**:
- contextcheck: Context usage
- copyloopvar: Loop variable copies
- depguard: Dependency restrictions
- dupl: Code duplication
- gocritic: Code quality
- gocyclo: Cyclomatic complexity (≤15)
- gosec: Security issues
- importas: Import aliases
- misspell: Spelling
- nakedret: Naked returns
- nolintlint: Linter directive usage
- prealloc: Slice preallocation
- revive: Code style
- unconvert: Unnecessary conversions
- whitespace: Whitespace issues

### Code Quality

**Quality gates** (all must pass):
```bash
make check  # Runs tests, linting, build
```

This command:
1. Runs all tests with coverage check
2. Runs all linters
3. Builds the binary

Do not submit pull requests until `make check` passes.

## Commit Standards

### Conventional Commits

dot follows [Conventional Commits](https://www.conventionalcommits.org/) specification strictly.

**Format**:
```
<type>(scope): <description>

[optional body]

[optional footer]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style (formatting, no logic change)
- `refactor`: Code restructuring (no behavior change)
- `test`: Test additions/improvements
- `chore`: Maintenance (no production code change)
- `perf`: Performance improvements
- `ci`: CI/CD changes
- `build`: Build system changes
- `revert`: Revert previous commit

**Scope**: Package or component affected (scanner, planner, cli, etc.)

**Examples**:
```
feat(scanner): add parallel package scanning
fix(planner): correct conflict detection logic
docs(readme): update installation instructions
test(scanner): add property tests for tree operations
refactor(executor): extract rollback logic
```

### Atomic Commits

Each commit must be:
- **Complete**: Represents one discrete logical change
- **Working**: Code compiles and tests pass
- **Independent**: Can be reverted without affecting others
- **Reviewable**: Small enough to understand easily

**Bad** (multiple changes):
```
feat: add scanner and planner functionality
```

**Good** (atomic):
```
feat(scanner): add package scanning
feat(planner): add state computation
```

### Commit Message Quality

**Subject line**:
- Under 50 characters
- Imperative mood: "add feature" not "added feature"
- No trailing period
- Capitalized

**Body** (for non-trivial changes):
- Wrap at 72 characters
- Explain what and why, not how
- Include motivation and context

**Example**:
```
feat(planner): add incremental change detection

Package updates are slow when processing large collections because
all packages are scanned even when unchanged. This adds content-based
change detection using fast hashing.

The implementation computes a hash for each package and compares with
the stored hash from the previous run. Unchanged packages are skipped
entirely, significantly improving performance for large repositories.

- Add content hashing using xxhash
- Store hashes in manifest
- Skip packages with unchanged hash
- Add tests for change detection

Closes #42
```

## Code Style

### Go Style

Follow standard Go conventions plus project-specific rules:

1. **Functional programming**: Prefer pure functions where possible
2. **No global state**: Avoid package-level variables
3. **Explicit errors**: Return errors explicitly, never panic for recoverable errors
4. **Immutable by default**: Minimize mutation
5. **Type safety**: Leverage type system for compile-time safety

### Naming Conventions

- **Files**: lowercase with underscores: `scanner_test.go`
- **Types**: PascalCase: `PackageScanner`
- **Functions**: camelCase: `scanPackage`
- **Constants**: PascalCase: `DefaultTimeout`
- **Interfaces**: PascalCase with -er suffix: `Scanner`, `Executor`

### Documentation

**Package documentation**:
```go
// Package scanner provides file tree scanning functionality.
//
// The scanner traverses package directories and builds tree
// representations for planning operations.
package scanner
```

**Function documentation**:
```go
// ScanPackage scans a package directory and returns its file tree.
//
// The scan respects ignore patterns and follows symbolic links.
// Returns an error if the package directory is inaccessible.
func ScanPackage(ctx context.Context, path Path[PackageDirKind]) (FileTree, error)
```

### Prohibited Practices

Never:
- Use emojis (in code, comments, docs, or output)
- Use `github.com/pkg/errors` (use standard library only)
- Use `gotest.tools/v3` (use testify)
- Ignore errors without documentation
- Use naked returns in functions >10 lines
- Add global state
- Panic for recoverable errors
- Use `--no-verify` flag when committing

## Pull Request Process

### Before Submitting

1. **Ensure quality**:
```bash
make check  # All tests and linters pass
```

2. **Update documentation**:
- Update relevant docs in `docs/`
- Add docstrings to new public functions
- Update CHANGELOG.md if applicable

3. **Write tests**:
- New functionality has tests
- Tests achieve ≥75% coverage
- All tests pass

### Submitting

1. **Push branch**:
```bash
git push origin feature-description
```

2. **Create pull request**:
- Clear title describing change
- Description explaining what, why, and how
- Link to related issues
- Include test results

3. **PR template**:
```markdown
## Description
Brief description of changes

## Motivation
Why this change is needed

## Changes
- List of specific changes

## Testing
- How was this tested?
- New tests added?
- Coverage maintained?

## Checklist
- [ ] Tests pass
- [ ] Linters pass
- [ ] Documentation updated
- [ ] Atomic commits
- [ ] Conventional commit messages
```

### Review Process

1. **Automated checks**: CI must pass
2. **Code review**: Maintainer reviews code
3. **Feedback**: Address review comments
4. **Approval**: Maintainer approves
5. **Merge**: Maintainer merges to main

### After Merge

1. Branch deleted automatically
2. Changes deployed to next release
3. CHANGELOG.md updated

## Architecture Constraints

### Port/Adapter Pattern

- **Domain layer**: Pure, no I/O dependencies
- **Port layer**: Interfaces for infrastructure
- **Adapter layer**: Concrete implementations
- **Core layer**: Pure functional logic
- **Shell layer**: Side-effecting execution

**New features must respect this architecture.**

### Dependency Rules

- Domain depends on nothing
- Core depends only on domain and ports
- Adapters depend on ports
- CLI depends on API layer only

**Dependencies flow inward, never outward.**

### Type Safety

Use phantom types for compile-time safety:

```go
// Good: Type-safe paths
func scanPackage(path Path[PackageDirKind]) error

// Bad: Untyped paths
func scanPackage(path string) error
```

## Documentation Standards

### Style

- **Academic tone**: Factual, precise, technical
- **No hyperbole**: Avoid subjective qualifiers
- **No emojis**: Never use emojis
- **Clear structure**: Logical organization
- **Complete**: All features documented

### When to Update

Update documentation when:
- Adding new features
- Changing behavior
- Fixing bugs that affect usage
- Updating configuration options

## Release Process

The release process is automated via **Release Please** and **GoReleaser**. Only maintainers perform releases.

### Overview

1. **Release Please** monitors the `main` branch for Conventional Commits.
2. It maintains a "Release PR" that updates the changelog and version.
3. When this PR is merged, it creates a git tag.
4. This tag triggers the **GoReleaser** workflow to build and publish artifacts.

### Pre-Release Checklist

Before merging the Release PR:

- [ ] All tests passing (`make test`)
- [ ] Linters passing (`make lint`)
- [ ] Code builds successfully (`make build`)
- [ ] Review the auto-generated CHANGELOG in the Release PR
- [ ] Verify version bump is correct (based on commit types)

### Release Steps

1. **Review Release PR**:
   - Go to Pull Requests tab on GitHub.
   - Look for the PR created by `release-please-bot`.
   - Verify the changes.

2. **Merge Release PR**:
   - Merge the PR into `main`.
   - Release Please will automatically tag the commit.

3. **Monitor Workflow**:
   - The tag push triggers `.github/workflows/release.yml`.
   - Verify it builds and publishes successfully.

### Troubleshooting

- If the Release PR is not created, ensure commits on `main` follow Conventional Commits.
- If the release fails to publish, check the Actions tab for `Release` workflow logs.

### Version Numbering

Follow [Semantic Versioning 2.0.0](https://semver.org/). Version bumps are determined automatically:

- **MAJOR** (`feat!`): Breaking changes
- **MINOR** (`feat`): New features
- **PATCH** (`fix`): Bug fixes


## Getting Help

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas
- **Documentation**: Search docs first
- **Architecture docs**: `docs/architecture/`

## Recognition

Contributors are recognized in:
- Git commit history
- CHANGELOG.md release notes
- Project README (significant contributions)

## Legal

By contributing, you agree that your contributions will be licensed under the MIT License.

## Summary

Key points:
1. **TDD required**: Write tests first
2. **75% coverage**: Maintain threshold
3. **Atomic commits**: One logical change per commit
4. **Conventional Commits**: Follow specification
5. **Quality gates**: `make check` must pass
6. **No emojis**: Ever
7. **Academic docs**: Factual, precise style

Thank you for contributing to dot!


