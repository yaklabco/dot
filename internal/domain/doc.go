// Package domain provides core domain types and error handling primitives.
//
// # Result[T] Usage Guidelines
//
// The Result[T] type provides monadic error handling for composing fallible operations.
// Follow these guidelines to use Result safely and idiomatically:
//
// ## When to Use Result[T]
//
// Use Result[T] for:
//   - Pure functional transformations with Map/FlatMap
//   - Composing multiple fallible operations
//   - Internal pipeline stages
//   - Functions that benefit from monadic composition
//
// ## When to Use (T, error)
//
// Use (T, error) for:
//   - Public API boundaries (pkg/dot interfaces)
//   - Leaf functions interfacing with Go stdlib/external libraries
//   - Functions where error details are more important than composition
//   - CLI command handlers
//
// ## Safe Unwrap() Usage
//
// Unwrap() and UnwrapErr() panic if called on the wrong variant. Only use them when:
//
//  1. After an explicit IsOk() or IsErr() check:
//     if result.IsOk() {
//     value := result.Unwrap()  // SAFE
//     }
//
//  2. When you have a compile-time or logical guarantee about the variant:
//     // Example: op.Target is already a validated FilePath, so String() -> NewFilePath cannot fail
//     path := domain.NewFilePath(validatedPath).Unwrap()  // SAFE (but redundant)
//
//  3. In test code where panics are acceptable:
//     got := ComputeResult().Unwrap()  // OK in tests
//     assert.Equal(t, expected, got)
//
// ## Prefer Safe Alternatives
//
// Instead of Unwrap/UnwrapErr, prefer:
//
//   - UnwrapOr(default) for values with acceptable fallbacks:
//     count := CountItems().UnwrapOr(0)  // Returns 0 on error
//
//   - OrElse(fn) for lazy fallback computation:
//     config := LoadConfig().OrElse(func() Config { return DefaultConfig() })
//
//   - Explicit error checking and propagation:
//     result := Compute()
//     if !result.IsOk() {
//     return fmt.Errorf("computation failed: %w", result.UnwrapErr())
//     }
//     value := result.Unwrap()  // Safe after check
//
// ## Converting Between Result[T] and (T, error)
//
// From Result[T] to (T, error):
//
//	func ToError(result domain.Result[T]) (T, error) {
//	    if result.IsOk() {
//	        return result.Unwrap(), nil
//	    }
//	    return zero, result.UnwrapErr()
//	}
//
// From (T, error) to Result[T]:
//
//	func FromError(value T, err error) domain.Result[T] {
//	    if err != nil {
//	        return domain.Err[T](err)
//	    }
//	    return domain.Ok(value)
//	}
//
// ## Example: Pipeline Stage
//
// Good use of Result for composition:
//
//	func ProcessPipeline(input string) domain.Result[Output] {
//	    return domain.FlatMap(
//	        Parse(input),
//	        func(parsed Parsed) domain.Result[Output] {
//	            return domain.FlatMap(
//	                Validate(parsed),
//	                func(validated Validated) domain.Result[Output] {
//	                    return Transform(validated)
//	                },
//	            )
//	        },
//	    )
//	}
//
// ## Example: Error Path (Safe Unwrap)
//
// Common pattern in pipeline/planner code:
//
//	scanResult := ScanStage()(ctx, input)
//	if scanResult.IsErr() {
//	    return domain.Err[Plan](scanResult.UnwrapErr())  // SAFE: checked IsErr()
//	}
//	packages := scanResult.Unwrap()  // SAFE: checked above via negation
//
// ## Known Patterns
//
// These patterns are safe but may appear risky:
//
//  1. Reconstructing already-validated paths:
//     // op.Target is already a FilePath, so String() -> NewFilePath is redundant but safe
//     path := domain.NewFilePath(op.Target.String()).Unwrap()
//
//     Better: Just use op.Target directly
//
//  2. Checked in previous conditional:
//     if result.IsErr() {
//     return domain.Err[T](result.UnwrapErr())
//     }
//     value := result.Unwrap()  // Safe: IsErr() checked above, so this must be Ok()
//
// ## Testing Helpers
//
// In test code, MustUnwrap variants are acceptable:
//
//	got := domain.MustOk(ComputeValue())        // Panics on error (fine in tests)
//	err := domain.MustErr(ExpectFailure())      // Panics on Ok (fine in tests)
//
// These are defined in internal/domain/testing.go and should NEVER be used in production code.
package domain
