package pipeline

import (
	"context"
	"errors"
	"sync"

	"github.com/yaklabco/dot/internal/domain"
)

// Pipeline represents a functional pipeline stage that transforms input A to output B.
// Pipelines are pure functions that can be composed to build complex workflows.
// All operations are context-aware for cancellation and timeout support.
type Pipeline[A, B any] func(context.Context, A) domain.Result[B]

// Compose combines two pipeline stages sequentially.
// The output of the first pipeline becomes the input to the second.
// If the first pipeline fails, the second is not executed.
// This implements the monadic bind operation for pipelines.
func Compose[A, B, C any](p1 Pipeline[A, B], p2 Pipeline[B, C]) Pipeline[A, C] {
	return func(ctx context.Context, a A) domain.Result[C] {
		// Check context before starting
		select {
		case <-ctx.Done():
			return domain.Err[C](ctx.Err())
		default:
		}

		// Execute first pipeline
		r1 := p1(ctx, a)
		if !r1.IsOk() {
			return domain.Err[C](r1.UnwrapErr())
		}

		// Execute second pipeline with result from first
		return p2(ctx, r1.Unwrap())
	}
}

// Parallel executes multiple pipelines concurrently with the same input.
// All pipelines receive the same input value and execute in parallel.
// Returns a slice of results in the same order as the input pipelines.
// If any pipeline fails, the entire operation fails with the first error encountered.
func Parallel[A, B any](pipelines []Pipeline[A, B]) Pipeline[A, []B] {
	return func(ctx context.Context, a A) domain.Result[[]B] {
		if len(pipelines) == 0 {
			return domain.Ok([]B{})
		}

		// Check context before starting
		select {
		case <-ctx.Done():
			return domain.Err[[]B](ctx.Err())
		default:
		}

		type result struct {
			index int
			value domain.Result[B]
		}

		resultChan := make(chan result, len(pipelines))
		var wg sync.WaitGroup

		// Launch goroutines for each pipeline
		for i, pipeline := range pipelines {
			wg.Add(1)
			go func(idx int, p Pipeline[A, B]) {
				defer wg.Done()
				r := p(ctx, a)
				resultChan <- result{index: idx, value: r}
			}(i, pipeline)
		}

		// Wait for all goroutines to complete
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect results maintaining order
		results := make([]domain.Result[B], len(pipelines))
		for r := range resultChan {
			results[r.index] = r.value
		}

		// Check for errors and collect values
		values := make([]B, 0, len(results))
		for _, r := range results {
			if !r.IsOk() {
				return domain.Err[[]B](r.UnwrapErr())
			}
			values = append(values, r.Unwrap())
		}

		return domain.Ok(values)
	}
}

// Map lifts a pure function into a pipeline stage.
// The function is applied to the input value, transforming it to the output type.
// Since the function cannot fail, the pipeline always succeeds.
func Map[A, B any](fn func(A) B) Pipeline[A, B] {
	return func(ctx context.Context, a A) domain.Result[B] {
		select {
		case <-ctx.Done():
			return domain.Err[B](ctx.Err())
		default:
			return domain.Ok(fn(a))
		}
	}
}

// FlatMap lifts a Result-returning function into a pipeline stage.
// This is useful for functions that can fail but don't need context.
func FlatMap[A, B any](fn func(A) domain.Result[B]) Pipeline[A, B] {
	return func(ctx context.Context, a A) domain.Result[B] {
		select {
		case <-ctx.Done():
			return domain.Err[B](ctx.Err())
		default:
			return fn(a)
		}
	}
}

// Filter creates a pipeline that only passes values matching a predicate.
// Values that don't match the predicate result in an error.
func Filter[A any](pred func(A) bool) Pipeline[A, A] {
	return func(ctx context.Context, a A) domain.Result[A] {
		select {
		case <-ctx.Done():
			return domain.Err[A](ctx.Err())
		default:
			if pred(a) {
				return domain.Ok(a)
			}
			return domain.Err[A](errors.New("value filtered out by predicate"))
		}
	}
}
