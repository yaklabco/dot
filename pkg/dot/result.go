package dot

import "github.com/yaklabco/dot/internal/domain"

// Result represents a value or an error, implementing a Result monad for error handling.
// This provides a functional approach to error handling with composition support.
//
// This is a re-export of domain.Result. See internal/domain for implementation.
type Result[T any] domain.Result[T]

// Ok creates a successful Result containing a value.
func Ok[T any](value T) Result[T] {
	return Result[T](domain.Ok(value))
}

// Err creates a failed Result containing an error.
func Err[T any](err error) Result[T] {
	return Result[T](domain.Err[T](err))
}

// IsOk returns true if the Result contains a value.
func (r Result[T]) IsOk() bool {
	return domain.Result[T](r).IsOk()
}

// IsErr returns true if the Result contains an error.
func (r Result[T]) IsErr() bool {
	return domain.Result[T](r).IsErr()
}

// Unwrap returns the contained value.
func (r Result[T]) Unwrap() T {
	return domain.Result[T](r).Unwrap()
}

// UnwrapErr returns the contained error.
func (r Result[T]) UnwrapErr() error {
	return domain.Result[T](r).UnwrapErr()
}

// UnwrapOr returns the contained value or a default if Result is Err.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	return domain.Result[T](r).UnwrapOr(defaultValue)
}

// Map applies a function to the contained value if Ok, otherwise propagates the error.
// This is the functorial map operation.
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	return Result[U](domain.Map(domain.Result[T](r), fn))
}

// FlatMap applies a function that returns a Result to the contained value if Ok.
// This is the monadic bind operation, enabling composition of Result-returning functions.
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	wrapped := func(t T) domain.Result[U] {
		return domain.Result[U](fn(t))
	}
	return Result[U](domain.FlatMap(domain.Result[T](r), wrapped))
}

// Collect aggregates a slice of Results into a Result containing a slice.
// Returns Err if any Result is Err, otherwise returns Ok with all values.
func Collect[T any](results []Result[T]) Result[[]T] {
	domainResults := make([]domain.Result[T], len(results))
	for i, r := range results {
		domainResults[i] = domain.Result[T](r)
	}
	return Result[[]T](domain.Collect(domainResults))
}
