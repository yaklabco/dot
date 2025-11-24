package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

func TestCompose(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		p1 := func(ctx context.Context, n int) domain.Result[int] {
			return domain.Ok(n + 1)
		}
		p2 := func(ctx context.Context, n int) domain.Result[int] {
			return domain.Ok(n * 2)
		}

		composed := Compose(p1, p2)
		result := composed(context.Background(), 5)

		require.True(t, result.IsOk())
		val := result.Unwrap()
		assert.Equal(t, 12, val) // (5 + 1) * 2
	})

	t.Run("first stage error propagates", func(t *testing.T) {
		expectedErr := errors.New("first error")
		p1 := func(ctx context.Context, n int) domain.Result[int] {
			return domain.Err[int](expectedErr)
		}
		p2 := func(ctx context.Context, n int) domain.Result[int] {
			t.Fatal("should not be called")
			return domain.Ok(n)
		}

		composed := Compose(p1, p2)
		result := composed(context.Background(), 5)

		require.False(t, result.IsOk())
		assert.Equal(t, expectedErr, result.UnwrapErr())
	})

	t.Run("second stage error propagates", func(t *testing.T) {
		expectedErr := errors.New("second error")
		p1 := func(ctx context.Context, n int) domain.Result[int] {
			return domain.Ok(n + 1)
		}
		p2 := func(ctx context.Context, n int) domain.Result[int] {
			return domain.Err[int](expectedErr)
		}

		composed := Compose(p1, p2)
		result := composed(context.Background(), 5)

		require.False(t, result.IsOk())
		assert.Equal(t, expectedErr, result.UnwrapErr())
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		p1 := func(ctx context.Context, n int) domain.Result[int] {
			select {
			case <-ctx.Done():
				return domain.Err[int](ctx.Err())
			default:
				return domain.Ok(n + 1)
			}
		}
		p2 := func(ctx context.Context, n int) domain.Result[int] {
			return domain.Ok(n * 2)
		}

		composed := Compose(p1, p2)
		result := composed(ctx, 5)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestParallel(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		pipelines := []Pipeline[int, int]{
			func(ctx context.Context, n int) domain.Result[int] {
				return domain.Ok(n + 1)
			},
			func(ctx context.Context, n int) domain.Result[int] {
				return domain.Ok(n * 2)
			},
			func(ctx context.Context, n int) domain.Result[int] {
				return domain.Ok(n * 3)
			},
		}

		parallel := Parallel(pipelines)
		result := parallel(context.Background(), 5)

		require.True(t, result.IsOk())
		vals := result.Unwrap()
		assert.ElementsMatch(t, []int{6, 10, 15}, vals)
	})

	t.Run("one error fails all", func(t *testing.T) {
		expectedErr := errors.New("parallel error")
		pipelines := []Pipeline[int, int]{
			func(ctx context.Context, n int) domain.Result[int] {
				return domain.Ok(n + 1)
			},
			func(ctx context.Context, n int) domain.Result[int] {
				return domain.Err[int](expectedErr)
			},
			func(ctx context.Context, n int) domain.Result[int] {
				return domain.Ok(n * 3)
			},
		}

		parallel := Parallel(pipelines)
		result := parallel(context.Background(), 5)

		require.False(t, result.IsOk())
		// Error should be in the results
		err := result.UnwrapErr()
		assert.NotNil(t, err)
	})

	t.Run("empty pipeline list", func(t *testing.T) {
		pipelines := []Pipeline[int, int]{}

		parallel := Parallel(pipelines)
		result := parallel(context.Background(), 5)

		require.True(t, result.IsOk())
		vals := result.Unwrap()
		assert.Empty(t, vals)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		pipelines := []Pipeline[int, int]{
			func(ctx context.Context, n int) domain.Result[int] {
				select {
				case <-ctx.Done():
					return domain.Err[int](ctx.Err())
				default:
					return domain.Ok(n + 1)
				}
			},
		}

		parallel := Parallel(pipelines)
		result := parallel(ctx, 5)

		require.False(t, result.IsOk())
	})
}

func TestMap(t *testing.T) {
	t.Run("transforms value", func(t *testing.T) {
		fn := func(n int) string {
			return "value"
		}

		pipeline := Map(fn)
		result := pipeline(context.Background(), 42)

		require.True(t, result.IsOk())
		assert.Equal(t, "value", result.Unwrap())
	})

	t.Run("never fails", func(t *testing.T) {
		fn := func(n int) int {
			return n * 2
		}

		pipeline := Map(fn)
		result := pipeline(context.Background(), 5)

		require.True(t, result.IsOk())
		assert.Equal(t, 10, result.Unwrap())
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		fn := func(n int) domain.Result[int] {
			return domain.Ok(n * 2)
		}

		pipeline := FlatMap(fn)
		result := pipeline(context.Background(), 5)

		require.True(t, result.IsOk())
		assert.Equal(t, 10, result.Unwrap())
	})

	t.Run("propagates error", func(t *testing.T) {
		expectedErr := errors.New("flatmap error")
		fn := func(n int) domain.Result[int] {
			return domain.Err[int](expectedErr)
		}

		pipeline := FlatMap(fn)
		result := pipeline(context.Background(), 5)

		require.False(t, result.IsOk())
		assert.Equal(t, expectedErr, result.UnwrapErr())
	})
}

func TestFilter(t *testing.T) {
	t.Run("passes matching value", func(t *testing.T) {
		pred := func(n int) bool {
			return n > 5
		}

		pipeline := Filter(pred)
		result := pipeline(context.Background(), 10)

		require.True(t, result.IsOk())
		assert.Equal(t, 10, result.Unwrap())
	})

	t.Run("rejects non-matching value", func(t *testing.T) {
		pred := func(n int) bool {
			return n > 5
		}

		pipeline := Filter(pred)
		result := pipeline(context.Background(), 3)

		require.False(t, result.IsOk())
		err := result.UnwrapErr()
		assert.Contains(t, err.Error(), "filtered out")
	})
}
