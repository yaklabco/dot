# Error Handling Patterns

## Overview

This document explains the error handling patterns used in the dot codebase and provides guidance on when to use each approach.

## Error Handling Approaches

The codebase uses two primary error handling patterns:

1. **Monadic Result[T]** - For domain boundary validation
2. **Standard Go errors** - For all other error handling

## Result[T] - Monadic Error Handling

### When to Use

Use `Result[T]` **only** for phantom type constructors and domain boundary validation:

- Path constructors (`NewFilePath`, `NewPackagePath`, `NewTargetPath`)
- Domain entity creation with validation
- Pipeline stage outputs (already using Result[T])

### Why Result[T] for Paths

Phantom-typed paths (`FilePath`, `PackagePath`, `TargetPath`) provide compile-time guarantees that paths are valid. The `Result[T]` type allows us to:

1. **Validate at construction time**: Paths are validated once when created
2. **Type safety**: Once a `FilePath` exists, it's guaranteed valid
3. **Functional composition**: Pipeline stages compose cleanly with Result[T]

### Example Usage

```go
// Path construction - uses Result[T]
pathResult := domain.NewFilePath("/path/to/file")
if !pathResult.IsOk() {
    return pathResult.UnwrapErr()
}
path := pathResult.Unwrap()

// Use the validated path (guaranteed valid)
data, err := fs.ReadFile(ctx, path.String())
```

### Result[T] Pattern

```go
type Result[T any] struct {
    value T
    err   error
}

func (r Result[T]) IsOk() bool
func (r Result[T]) IsErr() bool  
func (r Result[T]) Unwrap() T
func (r Result[T]) UnwrapErr() error
```

## Standard Go Errors

### When to Use

Use standard Go error returns for **everything else**:

- Filesystem operations
- Service method calls
- Business logic functions
- Network operations
- Database operations
- All I/O operations

### Why Standard Errors

Go's error handling is:

1. **Idiomatic**: Follows Go community standards
2. **Simple**: Easy to understand and reason about
3. **Composable**: Works with standard library (`errors.Is`, `errors.As`)
4. **Debuggable**: Stack traces and error wrapping well-supported

### Example Usage

```go
// Standard error handling
func (s *Service) DoWork(ctx context.Context) error {
    data, err := s.fetch(ctx)
    if err != nil {
        return fmt.Errorf("fetch failed: %w", err)
    }
    
    if err := s.process(data); err != nil {
        return fmt.Errorf("process failed: %w", err)
    }
    
    return nil
}
```

## Error Wrapping

### Use fmt.Errorf with %w

Always wrap errors with context:

```go
// Good
if err := operation(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Better - more context
if err := operation(id); err != nil {
    return fmt.Errorf("operation failed for id=%s: %w", id, err)
}
```

### Use Domain Error Helpers

For consistent error wrapping:

```go
import "github.com/yaklabco/dot/internal/domain"

// Wrap with context
err := domain.WrapError(err, "operation failed")

// Wrap with formatted context
err := domain.WrapErrorf(err, "operation failed for %s", name)
```

## Pattern Summary

| Situation | Pattern | Example |
|-----------|---------|---------|
| Path construction | Result[T] | `NewFilePath(path)` |
| Domain entity creation | Result[T] | `NewPackagePath(path)` |
| Pipeline stage output | Result[T] | `pipeline.Execute(ctx, input)` |
| Filesystem operations | Standard error | `fs.ReadFile(ctx, path)` |
| Service methods | Standard error | `svc.Manage(ctx, pkg)` |
| Business logic | Standard error | `compute(data)` |
| I/O operations | Standard error | `http.Get(url)` |

## Anti-Patterns to Avoid

### Don't Use Result[T] for Everything

```go
// Bad - unnecessary Result for simple operation
func compute(x int) domain.Result[int] {
    if x < 0 {
        return domain.Err[int](errors.New("negative"))
    }
    return domain.Ok(x * 2)
}

// Good - use standard error
func compute(x int) (int, error) {
    if x < 0 {
        return 0, errors.New("negative")
    }
    return x * 2, nil
}
```

### Don't Mix Patterns Unnecessarily

```go
// Bad - mixing Result and error in same function
func process(path string) (int, error) {
    pathResult := domain.NewFilePath(path)
    if !pathResult.IsOk() {
        return 0, pathResult.UnwrapErr()
    }
    // ... more logic returning error
}

// Good - convert to error at boundary
func process(path string) (int, error) {
    filePath, err := validatePath(path)
    if err != nil {
        return 0, err
    }
    // ... more logic
}

func validatePath(path string) (domain.FilePath, error) {
    result := domain.NewFilePath(path)
    if !result.IsOk() {
        return domain.FilePath{}, result.UnwrapErr()
    }
    return result.Unwrap(), nil
}
```

## Conversion Between Patterns

### Result[T] to Error

```go
// Method 1: Manual
pathResult := domain.NewFilePath(path)
if !pathResult.IsOk() {
    return pathResult.UnwrapErr()
}
path := pathResult.Unwrap()

// Method 2: Helper
path, err := domain.UnwrapResult(domain.NewFilePath(path), "invalid path")
if err != nil {
    return err
}
```

### Error to Result[T]

Only needed in pipeline stages:

```go
func stageFn(ctx context.Context, input Input) domain.Result[Output] {
    output, err := operation(ctx, input)
    if err != nil {
        return domain.Err[Output](err)
    }
    return domain.Ok(output)
}
```

## Historical Context

### Why This Decision

The codebase originally used Result[T] more broadly, inspired by functional programming patterns from Rust and Haskell. However, this created:

1. Cognitive overhead for Go developers
2. Inconsistency with Go idioms
3. Difficulty integrating with standard library
4. Mixed patterns in the same codebase

The current approach balances:
- **Type safety** at domain boundaries (phantom types)
- **Go idioms** for business logic (standard errors)
- **Functional composition** where it adds value (pipelines)

### Evolution

- **Phase 1**: Result[T] everywhere (too functional)
- **Phase 2**: Mixed usage (confusing)  
- **Phase 3**: Current pattern (balanced) ✓

## Guidelines for New Code

1. **Start with standard errors** - Use Go idioms by default
2. **Use Result[T] for paths** - Path constructors need validation
3. **Pipeline stages return Result[T]** - For functional composition
4. **Everything else returns error** - Simple and idiomatic

## Testing

### Testing Result[T]

```go
func TestPathConstruction(t *testing.T) {
    result := domain.NewFilePath("/valid/path")
    assert.True(t, result.IsOk())
    
    path := result.Unwrap()
    assert.Equal(t, "/valid/path", path.String())
}
```

### Testing Standard Errors

```go
func TestOperation(t *testing.T) {
    err := service.Operation(ctx)
    assert.NoError(t, err)
    
    // Or for error cases
    err = service.Operation(ctx)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected message")
}
```

## See Also

- [Architecture Documentation](architecture.md) - Overall system design
- [Testing Standards](testing.md) - Test requirements
- `internal/domain/result.go` - Result[T] implementation
- `internal/domain/errors_helpers.go` - Error wrapping helpers

