# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**dot** is a type-safe symbolic link manager for dotfiles written in Go. It provides transactional symlink management with conflict detection, incremental updates, and cross-platform support.

## Build and Development Commands

```bash
# Build and install
make build              # Build binary with version embedding
make install            # Install to GOBIN

# Testing
make test               # Run all tests with race detection + coverage
make test-tparse        # Tests with formatted output (recommended for development)
go test -v ./internal/planner/...  # Run tests for a specific package
go test -v -run TestFunctionName ./internal/planner/  # Run a single test

# Linting and quality
make lint               # golangci-lint v2
make vet                # go vet
make fmt-fix            # Fix formatting with goimports
make check-coverage     # Verify 60% threshold (UI files excluded)
make qa                 # Full QA: tests + lint + vet + coverage summary

# Full verification before commit
make check              # test + coverage + lint + vet + vuln
```

## Architecture

The codebase follows **Functional Core, Imperative Shell** architecture with six layers:

```
CLI (cmd/dot/)
    ↓
Public API (pkg/dot/) → Client facade with domain services
    ↓
Pipeline (internal/pipeline/) → Generic, composable operation stages
    ↓
Executor (internal/executor/) → Two-phase commit, transactional ops
    ↓
Core Logic (scanner/, planner/, ignore/) → Pure functional logic
    ↓
Domain (internal/domain/) → Phantom-typed paths, Result types
    ↓
Adapters (internal/adapters/) → Filesystem abstraction
```

### Key Architectural Concepts

**Phantom Types**: Compile-time path type safety prevents mixing package paths with target paths:
```go
type Path[K PathKind] struct { path string }
type PackageDirKind struct{}
type TargetDirKind struct{}
```

**Result Type**: Monadic error handling for functional composition.

**Viper Isolation**: Direct Viper usage is prohibited outside `internal/config/`. Use the config package for all configuration access.

### Critical Packages

| Package | Purpose |
|---------|---------|
| `internal/domain/` | Type-safe domain model, phantom types, port interfaces |
| `internal/planner/` | Desired state computation, conflict detection |
| `internal/scanner/` | Package scanning, dotfile translation |
| `internal/executor/` | Two-phase commit, transactional operations |
| `internal/config/` | All Viper usage isolated here |
| `pkg/dot/` | Public API facade (Client, services) |

## Code Standards

### Strictly Enforced

- **TDD mandatory**: Tests before implementation, 80%+ coverage target
- **No emojis**: In code, documentation, comments, or output
- **Standard errors only**: Use `errors` and `fmt.Errorf` with `%w`, never `github.com/pkg/errors`
- **testify only**: Use `github.com/stretchr/testify`, never `gotest.tools/v3`
- **Conventional Commits**: `<type>(scope): description`
- **Cyclomatic complexity**: Maximum 15

### golangci-lint v2 Configuration

Key linters enabled: `contextcheck`, `copyloopvar`, `depguard`, `dupl`, `gocritic`, `gocyclo` (15), `gosec`, `importas`, `misspell`, `nakedret`, `nolintlint`, `prealloc`, `revive`, `unconvert`, `whitespace`

Import restrictions enforced via depguard:
- `github.com/pkg/errors` → use standard `errors`
- `gotest.tools/v3` → use `testify`
- `github.com/spf13/viper` → only allowed in `internal/config/`

### Testing Patterns

- Table-driven tests with descriptive names: `TestFunctionName_Scenario_ExpectedBehavior`
- Both success and error paths tested
- Test files alongside implementation with `_test.go` suffix
- Test fixtures in `testdata/` directories

## Coverage Requirements

- **Local development**: 60% threshold (Bubble Tea UI files excluded)
- **CI/CD**: 75% threshold
- Excluded files: `selector.go`, `scanner.go`, `interactive.go`, `discovery.go` (interactive TUI)

## CI/CD Pipeline

Jobs run in order: lint → format → vet → vuln → test (75% coverage) → build (multi-platform matrix)

All checks must pass before merge. Race detection enabled. Codecov upload on success.
