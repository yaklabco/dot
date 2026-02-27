# Code Review Fixes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Address all issues identified in the comprehensive Go code review, improving security, API stability, code quality, and test coverage.

**Architecture:** Fixes are organized into four phases by priority. Phase 1 addresses critical security and API issues. Phase 2 fixes warning-level error handling problems. Phase 3 addresses code quality warnings. Phase 4 implements suggestions and improvements.

**Tech Stack:** Go 1.25.4, golangci-lint v2, testify, Cobra, Viper

**Beads Issues:** 50 issues created (7 P0, 33 P1, 10 P2)

---

## Phase 1: Critical Issues (P0)

### Task 1.1: Fix Path Join Validation Bypass (dot-4u5)

**Files:**
- Modify: `internal/domain/path.go:107-109`
- Modify: `internal/domain/path_test.go`

**Step 1: Write the failing test**

```go
// In internal/domain/path_test.go
func TestPath_Join_RejectsTraversal(t *testing.T) {
    tests := []struct {
        name     string
        base     string
        elem     string
        wantErr  bool
    }{
        {"simple join", "/home/user", "file.txt", false},
        {"nested join", "/home/user", "dir/file.txt", false},
        {"traversal attack", "/home/user", "../../../etc/passwd", true},
        {"hidden traversal", "/home/user", "foo/../../../etc/passwd", true},
        {"double dot only", "/home/user", "..", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            base := domain.NewPackagePath(tt.base)
            require.True(t, base.IsOk())

            result := base.Unwrap().JoinSafe(tt.elem)
            if tt.wantErr {
                assert.True(t, result.IsErr(), "expected error for %s", tt.elem)
            } else {
                assert.True(t, result.IsOk(), "expected success for %s", tt.elem)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestPath_Join_RejectsTraversal ./internal/domain/`
Expected: FAIL - JoinSafe method does not exist

**Step 3: Write minimal implementation**

```go
// In internal/domain/path.go

// JoinSafe appends a path element with validation to prevent traversal attacks.
// Returns an error if the resulting path would escape the base path.
func (p Path[K]) JoinSafe(elem string) Result[Path[K]] {
    // Clean the element to normalize any traversal sequences
    cleanedElem := filepath.Clean(elem)

    // Check for path traversal attempts
    if strings.HasPrefix(cleanedElem, "..") || strings.Contains(cleanedElem, string(filepath.Separator)+"..") {
        return Err[Path[K]](ErrInvalidPath{
            Path:   elem,
            Reason: "path contains traversal sequence",
        })
    }

    joined := filepath.Join(p.path, elem)
    cleanedJoined := filepath.Clean(joined)

    // Verify the result is still under the base path
    if !strings.HasPrefix(cleanedJoined, filepath.Clean(p.path)) {
        return Err[Path[K]](ErrInvalidPath{
            Path:   elem,
            Reason: "path escapes base directory",
        })
    }

    return Ok(Path[K]{path: cleanedJoined})
}

// Join appends a path element without validation.
// Deprecated: Use JoinSafe for user-provided paths. Join is retained for
// internal use where the element is known to be safe.
func (p Path[K]) Join(elem string) Path[K] {
    joined := filepath.Join(p.path, elem)
    return Path[K]{path: joined}
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestPath_Join_RejectsTraversal ./internal/domain/`
Expected: PASS

**Step 5: Run full test suite**

Run: `make test`
Expected: All tests pass

**Step 6: Commit**

```bash
git add internal/domain/path.go internal/domain/path_test.go
git commit -m "fix(domain): add JoinSafe to prevent path traversal attacks

Adds JoinSafe method that validates path elements before joining.
Prevents directory traversal attacks like '../../../etc/passwd'.
Original Join method retained but deprecated for internal use.

Fixes: dot-4u5

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 1.2: Fix FileInfo.ModTime Return Type (dot-3eo)

**Files:**
- Modify: `internal/domain/ports.go:33-40`
- Modify: `internal/adapters/osfs.go` (if wrapper exists)
- Modify: `internal/adapters/memfs.go` (if wrapper exists)

**Step 1: Write the failing test**

```go
// In internal/domain/ports_test.go
func TestFileInfo_ModTime_ReturnsTime(t *testing.T) {
    // This test verifies the interface matches stdlib
    var _ interface {
        ModTime() time.Time
    } = (domain.FileInfo)(nil)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestFileInfo_ModTime_ReturnsTime ./internal/domain/`
Expected: FAIL - interface mismatch

**Step 3: Update the interface**

```go
// In internal/domain/ports.go

// FileInfo describes a file and is compatible with fs.FileInfo.
type FileInfo interface {
    Name() string
    Size() int64
    Mode() os.FileMode
    ModTime() time.Time  // Changed from any to time.Time
    IsDir() bool
    Sys() any
}
```

**Step 4: Update any adapters that wrap os.FileInfo**

Check if adapters need changes. If osfs.go returns os.FileInfo directly, no changes needed.
If memfs.go has a custom implementation, update it:

```go
// In internal/adapters/memfs.go (if needed)
func (f *memFileInfo) ModTime() time.Time {
    return f.modTime
}
```

**Step 5: Run test to verify it passes**

Run: `go test -v -run TestFileInfo_ModTime_ReturnsTime ./internal/domain/`
Expected: PASS

**Step 6: Run full test suite**

Run: `make test`
Expected: All tests pass

**Step 7: Commit**

```bash
git add internal/domain/ports.go internal/adapters/
git commit -m "fix(domain): change FileInfo.ModTime to return time.Time

Updates FileInfo interface to match stdlib fs.FileInfo signature.
This improves compatibility with standard library code.

Fixes: dot-3eo

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 1.3: Add Is/As Implementations for Clone Errors (dot-xjt)

**Files:**
- Modify: `pkg/dot/errors.go:69-129`
- Create: `pkg/dot/errors_test.go` (if not exists)

**Step 1: Write the failing test**

```go
// In pkg/dot/errors_test.go
func TestCloneErrors_SupportsErrorsIs(t *testing.T) {
    tests := []struct {
        name   string
        err    error
        target error
        want   bool
    }{
        {
            name:   "ErrProfileNotFound matches sentinel",
            err:    ErrProfileNotFound{Profile: "test"},
            target: ErrProfileNotFound{},
            want:   true,
        },
        {
            name:   "wrapped ErrProfileNotFound matches",
            err:    fmt.Errorf("context: %w", ErrProfileNotFound{Profile: "test"}),
            target: ErrProfileNotFound{},
            want:   true,
        },
        {
            name:   "ErrBootstrapNotFound matches sentinel",
            err:    ErrBootstrapNotFound{Path: "/path"},
            target: ErrBootstrapNotFound{},
            want:   true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := errors.Is(tt.err, tt.target)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestCloneErrors_SupportsErrorsIs ./pkg/dot/`
Expected: FAIL - errors.Is returns false

**Step 3: Add Is methods to error types**

```go
// In pkg/dot/errors.go

// Is implements errors.Is for ErrProfileNotFound.
func (e ErrProfileNotFound) Is(target error) bool {
    _, ok := target.(ErrProfileNotFound)
    return ok
}

// Is implements errors.Is for ErrBootstrapNotFound.
func (e ErrBootstrapNotFound) Is(target error) bool {
    _, ok := target.(ErrBootstrapNotFound)
    return ok
}

// Is implements errors.Is for ErrCloneFailed.
func (e ErrCloneFailed) Is(target error) bool {
    _, ok := target.(ErrCloneFailed)
    return ok
}

// Is implements errors.Is for ErrBootstrapFailed.
func (e ErrBootstrapFailed) Is(target error) bool {
    _, ok := target.(ErrBootstrapFailed)
    return ok
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestCloneErrors_SupportsErrorsIs ./pkg/dot/`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/dot/errors.go pkg/dot/errors_test.go
git commit -m "feat(api): add Is methods to clone error types

Enables errors.Is() checking for ErrProfileNotFound, ErrBootstrapNotFound,
ErrCloneFailed, and ErrBootstrapFailed. Consumers can now programmatically
detect these error types.

Fixes: dot-xjt

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 1.4: Return Typed ErrConflict (dot-4si)

**Files:**
- Modify: `pkg/dot/errors.go`
- Modify: `pkg/dot/manage_service.go:80`
- Modify: `pkg/dot/manage_service_test.go`

**Step 1: Write the failing test**

```go
// In pkg/dot/manage_service_test.go
func TestManageService_ReturnsTypedConflictError(t *testing.T) {
    // Setup a scenario that causes a conflict
    // ... setup code ...

    err := service.Manage(ctx, packages, options)

    var conflictErr ErrConflict
    assert.True(t, errors.As(err, &conflictErr), "expected ErrConflict type")
}
```

**Step 2: Define ErrConflict type**

```go
// In pkg/dot/errors.go

// ErrConflict indicates a conflict was detected during operation.
type ErrConflict struct {
    Conflicts []string
    Message   string
}

func (e ErrConflict) Error() string {
    return e.Message
}

// Is implements errors.Is for ErrConflict.
func (e ErrConflict) Is(target error) bool {
    _, ok := target.(ErrConflict)
    return ok
}
```

**Step 3: Update manage_service.go to use typed error**

```go
// In pkg/dot/manage_service.go:80 (approximate)

// Replace:
// return errors.New(conflictMsg)

// With:
return ErrConflict{
    Conflicts: conflictPaths,
    Message:   conflictMsg,
}
```

**Step 4: Run tests**

Run: `go test -v ./pkg/dot/...`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/dot/errors.go pkg/dot/manage_service.go pkg/dot/manage_service_test.go
git commit -m "feat(api): return typed ErrConflict for conflict errors

Consumers can now detect conflicts programmatically using errors.As().
The ErrConflict type includes the list of conflicting paths.

Fixes: dot-4si

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 1.5: Eliminate Global CLI Flags State (dot-h70)

**Files:**
- Modify: `cmd/dot/root.go:43`
- Modify: `cmd/dot/command_helpers.go`
- Modify: All command files that access `cliFlags`

**Step 1: Define context key and helpers**

```go
// In cmd/dot/context.go (new file)
package main

import "context"

type cliFlagsKey struct{}

// WithCLIFlags adds CLIFlags to the context.
func WithCLIFlags(ctx context.Context, flags *CLIFlags) context.Context {
    return context.WithValue(ctx, cliFlagsKey{}, flags)
}

// CLIFlagsFromContext retrieves CLIFlags from context.
// Returns nil if not set.
func CLIFlagsFromContext(ctx context.Context) *CLIFlags {
    if flags, ok := ctx.Value(cliFlagsKey{}).(*CLIFlags); ok {
        return flags
    }
    return nil
}

// MustCLIFlagsFromContext retrieves CLIFlags from context or panics.
func MustCLIFlagsFromContext(ctx context.Context) *CLIFlags {
    flags := CLIFlagsFromContext(ctx)
    if flags == nil {
        panic("CLIFlags not set in context")
    }
    return flags
}
```

**Step 2: Update root.go to use context**

```go
// In cmd/dot/root.go

func NewRootCommand(version, commit, date string) *cobra.Command {
    flags := &CLIFlags{}  // Local, not global

    rootCmd := &cobra.Command{
        // ...
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Store flags in context
            ctx := WithCLIFlags(cmd.Context(), flags)
            cmd.SetContext(ctx)
            // ... rest of setup
            return nil
        },
    }

    // Bind flags to local struct
    rootCmd.PersistentFlags().StringVarP(&flags.packageDir, "dir", "d", "", "Package directory")
    // ... other flags

    return rootCmd
}
```

**Step 3: Update all commands to use context**

Replace all occurrences of `cliFlags.X` with `CLIFlagsFromContext(cmd.Context()).X` or use the helper.

**Step 4: Remove global variable**

```go
// In cmd/dot/root.go
// Delete: var cliFlags CLIFlags
```

**Step 5: Update tests**

```go
// Tests now set context properly
ctx := WithCLIFlags(context.Background(), &CLIFlags{...})
cmd.SetContext(ctx)
```

**Step 6: Run tests**

Run: `make test`
Expected: PASS

**Step 7: Commit**

```bash
git add cmd/dot/
git commit -m "refactor(cli): eliminate global cliFlags state

Moves CLI flags from global variable to context. This fixes test isolation
issues and potential race conditions. All commands now retrieve flags
from context using CLIFlagsFromContext().

Fixes: dot-h70

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 1.6: Make stdin/stdout Configurable (dot-l8t)

**Files:**
- Modify: `pkg/dot/config.go`
- Modify: `pkg/dot/client.go:129`

**Step 1: Add IO fields to Config**

```go
// In pkg/dot/config.go

type Config struct {
    // ... existing fields ...

    // Stdin is the input reader for interactive prompts.
    // Defaults to os.Stdin if nil.
    Stdin io.Reader

    // Stdout is the output writer for interactive prompts.
    // Defaults to os.Stdout if nil.
    Stdout io.Writer
}

// getStdin returns the configured stdin or os.Stdin.
func (c *Config) getStdin() io.Reader {
    if c.Stdin != nil {
        return c.Stdin
    }
    return os.Stdin
}

// getStdout returns the configured stdout or os.Stdout.
func (c *Config) getStdout() io.Writer {
    if c.Stdout != nil {
        return c.Stdout
    }
    return os.Stdout
}
```

**Step 2: Update client.go to use config**

```go
// In pkg/dot/client.go:129 (approximate)

// Replace:
// packageSelector := selector.NewInteractiveSelector(os.Stdin, os.Stdout)

// With:
packageSelector := selector.NewInteractiveSelector(config.getStdin(), config.getStdout())
```

**Step 3: Write test**

```go
// In pkg/dot/client_test.go
func TestClient_UsesConfiguredIO(t *testing.T) {
    stdin := bytes.NewBufferString("input\n")
    stdout := &bytes.Buffer{}

    cfg := &Config{
        Stdin:  stdin,
        Stdout: stdout,
        // ... other required fields
    }

    client, err := NewClient(cfg)
    require.NoError(t, err)
    assert.NotNil(t, client)
    // Further assertions based on how IO is used
}
```

**Step 4: Run tests**

Run: `go test -v ./pkg/dot/...`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/dot/config.go pkg/dot/client.go pkg/dot/client_test.go
git commit -m "feat(api): make stdin/stdout configurable via Config

Adds Stdin and Stdout fields to Config for dependency injection.
Defaults to os.Stdin/os.Stdout when nil. This enables testing
interactive features without patching global state.

Fixes: dot-l8t

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 1.7: Add API Contract Tests (dot-adh)

**Files:**
- Modify: `pkg/dot/client_test.go`

**Step 1: Add method contract tests**

```go
// In pkg/dot/client_test.go

func TestClient_MethodSignatures(t *testing.T) {
    // Verify all public methods exist with expected signatures
    var client *Client

    // These will fail to compile if signatures change
    var _ func(context.Context, []string, ManageOptions) error = client.Manage
    var _ func(context.Context, []string, UnmanageOptions) error = client.Unmanage
    var _ func(context.Context) ([]PackageInfo, error) = client.Status
    var _ func(context.Context, DoctorOptions) ([]Issue, error) = client.Doctor
    // ... add all public methods
}

func TestClient_ErrorTypes_SupportErrorsIs(t *testing.T) {
    // Test that all error types can be detected with errors.Is
    errorTypes := []error{
        ErrInvalidPath{},
        ErrConflict{},
        ErrProfileNotFound{},
        ErrBootstrapNotFound{},
        ErrCloneFailed{},
    }

    for _, errType := range errorTypes {
        t.Run(fmt.Sprintf("%T", errType), func(t *testing.T) {
            wrapped := fmt.Errorf("wrapped: %w", errType)
            assert.True(t, errors.Is(wrapped, errType),
                "errors.Is should match wrapped %T", errType)
        })
    }
}

func TestClient_DryRun_NoSideEffects(t *testing.T) {
    // Verify dry-run produces no filesystem changes
    tmpDir := t.TempDir()
    targetDir := t.TempDir()

    // Create a package
    pkgDir := filepath.Join(tmpDir, "dot-test")
    require.NoError(t, os.MkdirAll(pkgDir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "file"), []byte("content"), 0644))

    cfg := &Config{
        PackageDir: tmpDir,
        TargetDir:  targetDir,
        DryRun:     true,
    }

    client, err := NewClient(cfg)
    require.NoError(t, err)

    // Run manage in dry-run mode
    err = client.Manage(context.Background(), []string{"dot-test"}, ManageOptions{})
    require.NoError(t, err)

    // Verify no symlinks were created
    entries, err := os.ReadDir(targetDir)
    require.NoError(t, err)
    assert.Empty(t, entries, "dry-run should not create files")
}
```

**Step 2: Run tests**

Run: `go test -v ./pkg/dot/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/dot/client_test.go
git commit -m "test(api): add comprehensive API contract tests

Adds tests verifying:
- All public method signatures (compile-time check)
- All error types support errors.Is()
- DryRun mode produces no side effects

Fixes: dot-adh

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Phase 2: Warning Issues - Error Handling (P1)

### Task 2.1: Replace deprecated strings.Title (dot-6o2)

**Files:**
- Modify: `cmd/dot/unmanage.go:285`
- Modify: `internal/cli/output/format.go:29`
- Modify: `go.mod` (add golang.org/x/text)

**Step 1: Add dependency**

Run: `go get golang.org/x/text/cases golang.org/x/text/language`

**Step 2: Update unmanage.go**

```go
// In cmd/dot/unmanage.go

import (
    "golang.org/x/text/cases"
    "golang.org/x/text/language"
)

// Replace:
// strings.Title(operation)

// With:
cases.Title(language.English).String(operation)
```

**Step 3: Update format.go**

```go
// In internal/cli/output/format.go

import (
    "golang.org/x/text/cases"
    "golang.org/x/text/language"
)

// Replace:
// verb = strings.Title(verb)

// With:
verb = cases.Title(language.English).String(verb)
```

**Step 4: Run tests**

Run: `make test`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/dot/unmanage.go internal/cli/output/format.go go.mod go.sum
git commit -m "fix(cli): replace deprecated strings.Title with cases.Title

strings.Title was deprecated in Go 1.18. Uses golang.org/x/text/cases
for proper Unicode title casing.

Fixes: dot-6o2

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 2.2: Handle os.UserHomeDir Error (dot-2j2)

**Files:**
- Modify: `internal/planner/validate.go:12-14`

**Step 1: Update function signature and implementation**

```go
// In internal/planner/validate.go

// DotOperationalPaths returns paths that dot uses for its own operation.
// Returns an error if the home directory cannot be determined.
func DotOperationalPaths() ([]string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("determine home directory: %w", err)
    }

    return []string{
        filepath.Join(homeDir, ".config", "dot"),
        filepath.Join(homeDir, ".local", "share", "dot"),
        filepath.Join(homeDir, ".local", "state", "dot"),
    }, nil
}
```

**Step 2: Update all callers to handle error**

Find and update all callers of `DotOperationalPaths()`.

**Step 3: Run tests**

Run: `make test`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/planner/
git commit -m "fix(planner): handle os.UserHomeDir error in DotOperationalPaths

Previously ignored error from os.UserHomeDir which could cause
validation to silently pass with empty paths.

Fixes: dot-2j2

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 2.3: Fix Aliasing in Fluent API Methods (dot-rn5)

**Files:**
- Modify: `internal/planner/resolver.go:75-84, 177-186`
- Add tests to verify immutability

**Step 1: Write failing test**

```go
// In internal/planner/resolver_test.go
func TestConflict_WithContext_IsImmutable(t *testing.T) {
    original := Conflict{
        Context: map[string]string{"key1": "val1"},
    }

    modified := original.WithContext("key2", "val2")

    // Original should be unchanged
    assert.NotContains(t, original.Context, "key2")
    assert.Contains(t, modified.Context, "key2")
}

func TestConflict_WithSuggestion_IsImmutable(t *testing.T) {
    original := Conflict{
        Suggestions: []Suggestion{{Text: "first"}},
    }

    modified := original.WithSuggestion(Suggestion{Text: "second"})

    // Original should be unchanged
    assert.Len(t, original.Suggestions, 1)
    assert.Len(t, modified.Suggestions, 2)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestConflict_With ./internal/planner/`
Expected: FAIL

**Step 3: Fix implementation**

```go
// In internal/planner/resolver.go

func (c Conflict) WithContext(key, value string) Conflict {
    newContext := make(map[string]string, len(c.Context)+1)
    for k, v := range c.Context {
        newContext[k] = v
    }
    newContext[key] = value
    c.Context = newContext
    return c
}

func (c Conflict) WithSuggestion(s Suggestion) Conflict {
    newSuggestions := make([]Suggestion, len(c.Suggestions)+1)
    copy(newSuggestions, c.Suggestions)
    newSuggestions[len(c.Suggestions)] = s
    c.Suggestions = newSuggestions
    return c
}

func (r ResolveResult) WithConflict(c Conflict) ResolveResult {
    newConflicts := make([]Conflict, len(r.Conflicts)+1)
    copy(newConflicts, r.Conflicts)
    newConflicts[len(r.Conflicts)] = c
    r.Conflicts = newConflicts
    return r
}

func (r ResolveResult) WithWarning(w Warning) ResolveResult {
    newWarnings := make([]Warning, len(r.Warnings)+1)
    copy(newWarnings, r.Warnings)
    newWarnings[len(r.Warnings)] = w
    r.Warnings = newWarnings
    return r
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestConflict_With ./internal/planner/`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/planner/resolver.go internal/planner/resolver_test.go
git commit -m "fix(planner): ensure fluent API methods are immutable

WithContext, WithSuggestion, WithConflict, and WithWarning now copy
underlying maps/slices before modification to prevent aliasing bugs.

Fixes: dot-rn5

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 2.4: Check Result Before Unwrap (dot-yhg)

**Files:**
- Modify: `internal/planner/desired.go:147`
- Modify: `internal/planner/resolver.go:234, 248`

**Step 1: Update desired.go**

```go
// In internal/planner/desired.go:147

// Replace:
// dirPath := domain.NewFilePath(parentStr).Unwrap()

// With:
dirPathResult := domain.NewFilePath(parentStr)
if dirPathResult.IsErr() {
    return domain.Err[DesiredState](fmt.Errorf("invalid path %s: %w", parentStr, dirPathResult.UnwrapErr()))
}
dirPath := dirPathResult.Unwrap()
```

**Step 2: Update resolver.go**

Apply same pattern at lines 234 and 248.

**Step 3: Run tests**

Run: `make test`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/planner/desired.go internal/planner/resolver.go
git commit -m "fix(planner): check Result.IsOk before Unwrap calls

Prevents potential panics from unchecked Unwrap calls.
Now properly propagates errors when path creation fails.

Fixes: dot-yhg

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 2.5: Use errors.As in Executor (dot-289)

**Files:**
- Modify: `internal/executor/executor.go:92-95`

**Step 1: Update error checking**

```go
// In internal/executor/executor.go

// Replace:
// for _, err := range result.Errors {
//     if _, ok := err.(domain.ErrExecutionCancelled); ok {
//         isCancelled = true
//         break
//     }
// }

// With:
for _, err := range result.Errors {
    var cancelErr domain.ErrExecutionCancelled
    if errors.As(err, &cancelErr) {
        isCancelled = true
        break
    }
}
```

**Step 2: Run tests**

Run: `go test -v ./internal/executor/...`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/executor/executor.go
git commit -m "refactor(executor): use errors.As for error type checking

Replaces direct type assertion with errors.As for future-proof
error handling that works with wrapped errors.

Fixes: dot-289

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Phase 3: Warning Issues - Code Quality (P1)

### Task 3.1: Add nolint Comments for BindEnv (dot-4ez)

**Files:**
- Modify: `internal/config/loader.go:274-320`

**Step 1: Add nolint directive with explanation**

```go
// In internal/config/loader.go

// bindEnvironmentVariables binds environment variables to viper keys.
// BindEnv only fails when called with no arguments, which cannot happen
// with string literals. Errors are intentionally ignored.
//
//nolint:errcheck // BindEnv only fails with empty keys; string literals always valid
func bindEnvironmentVariables(v *viper.Viper) {
    v.BindEnv("directories.package")
    v.BindEnv("directories.target")
    // ... rest of bindings
}
```

**Step 2: Run linter**

Run: `make lint`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/config/loader.go
git commit -m "chore(config): add nolint directive for BindEnv calls

Documents why BindEnv errors are intentionally ignored.
BindEnv only fails with empty keys, which cannot happen
with string literal arguments.

Fixes: dot-4ez

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3.2: Remove Unused Function (dot-9v7)

**Files:**
- Modify: `internal/config/loader.go:267-270`

**Step 1: Remove the function**

```go
// In internal/config/loader.go

// Delete:
// func getEnvWithPrefix(prefix, key string) string {
//     return os.Getenv(prefix + key)
// }
```

**Step 2: Run tests**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/config/loader.go
git commit -m "chore(config): remove unused getEnvWithPrefix function

Dead code removal.

Fixes: dot-9v7

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3.3: Fix MemFS Thread-Safety Comment (dot-ltv)

**Files:**
- Modify: `internal/adapters/memfs.go:15`

**Step 1: Update comment**

```go
// In internal/adapters/memfs.go

// MemFS implements an in-memory filesystem for testing.
// It is thread-safe and can be used in concurrent tests.
type MemFS struct {
    files map[string]*memFile
    mu    sync.RWMutex
}
```

**Step 2: Commit**

```bash
git add internal/adapters/memfs.go
git commit -m "docs(adapters): fix MemFS thread-safety comment

The comment incorrectly stated MemFS is not thread-safe,
but it uses sync.RWMutex for proper synchronization.

Fixes: dot-ltv

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3.4: Use Map for Reserved Names (dot-7gn)

**Files:**
- Modify: `internal/scanner/reserved.go:8-12`

**Step 1: Update implementation**

```go
// In internal/scanner/reserved.go

var reservedNames = map[string]struct{}{
    "dot":        {},
    ".dot":       {},
    "dot-config": {},
}

// IsReservedPackageName checks if a package name is reserved for dot's use.
func IsReservedPackageName(name string) bool {
    _, exists := reservedNames[strings.ToLower(name)]
    return exists
}
```

**Step 2: Run tests**

Run: `go test -v ./internal/scanner/...`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/scanner/reserved.go
git commit -m "perf(scanner): use map for O(1) reserved name lookup

Replaces linear search with map lookup for better performance.

Fixes: dot-7gn

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Phase 4: Additional Tasks

Due to the large number of remaining issues, they should be addressed in subsequent implementation sessions. The remaining issues include:

**P1 Tasks (remaining):**
- Handle errors in NewDefaultIgnoreSet (dot-5pf)
- Document Parallel error handling (dot-c01)
- Improve Filter error message (dot-ucb)
- Remove redundant context checks (dot-jth)
- Add comments for ignored errors (dot-bcm, dot-tn1, dot-p1g)
- Split large FS interfaces (dot-xnw, dot-8g2)
- Fix domain Validate error types (dot-d5p)
- Remove deprecated Context alias (dot-6m2)
- Remove unused packages field (dot-4hu)
- Add graceful pprof shutdown (dot-obd)
- Handle Scanln error (dot-jxg)
- Remove global doctor state (dot-2a1)
- Fix prealloc warnings (dot-k0q, dot-hj9)
- Use Logger instead of stderr (dot-lr8)
- Document thread safety (dot-v2f)
- Document flag aliases (dot-4zz)
- Add config error comments (dot-mtw)
- Increase doctor test coverage (dot-xd1)

**P2 Tasks:**
- Consider stdlib interfaces (dot-ate)
- Replace custom As helper (dot-5fl)
- Preallocate slices (dot-47t, dot-a5t)
- Use strings.Contains (dot-afg)
- Standardize Result pattern (dot-d31)
- Consolidate pluralize (dot-ak7)
- Extract context helper (dot-5l4)
- Consider builder pattern (dot-bxg)
- Add documentation (dot-uxm, dot-3ua)
- Add rollback integration test (dot-at1)
- Extract magic numbers (dot-g41)
- Consider typed SeverityLevel (dot-ies)
- Consider functional options (dot-3bt)
- Fix error assignment (dot-chu)

---

## Execution Checklist

Before each task:
- [ ] Create/switch to feature branch
- [ ] Read the relevant files

After each task:
- [ ] Run `make test`
- [ ] Run `make lint`
- [ ] Run `go-code-reviewer` agent
- [ ] Commit with conventional commit message
- [ ] Update beads issue status

After each phase:
- [ ] Run `make check` (full verification)
- [ ] Update CHANGELOG.md
- [ ] Consider PR creation
