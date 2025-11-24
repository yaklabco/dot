package dot_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestOk(t *testing.T) {
	result := dot.Ok(42)

	assert.True(t, result.IsOk())
	assert.False(t, result.IsErr())
	assert.Equal(t, 42, result.Unwrap())
}

func TestErr(t *testing.T) {
	err := errors.New("test error")
	result := dot.Err[int](err)

	assert.False(t, result.IsOk())
	assert.True(t, result.IsErr())
	assert.Equal(t, err, result.UnwrapErr())
}

func TestUnwrapPanicsOnErr(t *testing.T) {
	result := dot.Err[int](errors.New("test error"))

	assert.Panics(t, func() {
		result.Unwrap()
	})
}

func TestUnwrapErrPanicsOnOk(t *testing.T) {
	result := dot.Ok(42)

	assert.Panics(t, func() {
		result.UnwrapErr()
	})
}

func TestMap(t *testing.T) {
	t.Run("map on Ok", func(t *testing.T) {
		result := dot.Ok(42)
		mapped := dot.Map(result, func(x int) int {
			return x * 2
		})

		assert.True(t, mapped.IsOk())
		assert.Equal(t, 84, mapped.Unwrap())
	})

	t.Run("map on Err", func(t *testing.T) {
		err := errors.New("test error")
		result := dot.Err[int](err)
		mapped := dot.Map(result, func(x int) int {
			return x * 2
		})

		assert.True(t, mapped.IsErr())
		assert.Equal(t, err, mapped.UnwrapErr())
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("flatMap on Ok returning Ok", func(t *testing.T) {
		result := dot.Ok(42)
		mapped := dot.FlatMap(result, func(x int) dot.Result[int] {
			return dot.Ok(x * 2)
		})

		assert.True(t, mapped.IsOk())
		assert.Equal(t, 84, mapped.Unwrap())
	})

	t.Run("flatMap on Ok returning Err", func(t *testing.T) {
		result := dot.Ok(42)
		err := errors.New("computation failed")
		mapped := dot.FlatMap(result, func(x int) dot.Result[int] {
			return dot.Err[int](err)
		})

		assert.True(t, mapped.IsErr())
		assert.Equal(t, err, mapped.UnwrapErr())
	})

	t.Run("flatMap on Err", func(t *testing.T) {
		err := errors.New("test error")
		result := dot.Err[int](err)
		mapped := dot.FlatMap(result, func(x int) dot.Result[int] {
			return dot.Ok(x * 2)
		})

		assert.True(t, mapped.IsErr())
		assert.Equal(t, err, mapped.UnwrapErr())
	})
}

func TestCollect(t *testing.T) {
	t.Run("all Ok", func(t *testing.T) {
		results := []dot.Result[int]{
			dot.Ok(1),
			dot.Ok(2),
			dot.Ok(3),
		}

		collected := dot.Collect(results)
		assert.True(t, collected.IsOk())
		assert.Equal(t, []int{1, 2, 3}, collected.Unwrap())
	})

	t.Run("contains Err", func(t *testing.T) {
		err := errors.New("test error")
		results := []dot.Result[int]{
			dot.Ok(1),
			dot.Err[int](err),
			dot.Ok(3),
		}

		collected := dot.Collect(results)
		assert.True(t, collected.IsErr())
		assert.Equal(t, err, collected.UnwrapErr())
	})

	t.Run("empty slice", func(t *testing.T) {
		results := []dot.Result[int]{}

		collected := dot.Collect(results)
		assert.True(t, collected.IsOk())
		assert.Empty(t, collected.Unwrap())
	})
}

func TestUnwrapOr(t *testing.T) {
	t.Run("Ok returns value", func(t *testing.T) {
		result := dot.Ok(42)
		value := result.UnwrapOr(10)
		assert.Equal(t, 42, value)
	})

	t.Run("Err returns default", func(t *testing.T) {
		result := dot.Err[int](errors.New("error"))
		value := result.UnwrapOr(10)
		assert.Equal(t, 10, value)
	})
}

// Test monad laws
func TestMonadLaws(t *testing.T) {
	t.Run("left identity", func(t *testing.T) {
		// return a >>= f ≡ f a
		a := 42
		f := func(x int) dot.Result[int] {
			return dot.Ok(x * 2)
		}

		left := dot.FlatMap(dot.Ok(a), f)
		right := f(a)

		assert.Equal(t, right.Unwrap(), left.Unwrap())
	})

	t.Run("right identity", func(t *testing.T) {
		// m >>= return ≡ m
		m := dot.Ok(42)
		result := dot.FlatMap(m, dot.Ok[int])

		assert.Equal(t, m.Unwrap(), result.Unwrap())
	})

	t.Run("associativity", func(t *testing.T) {
		// (m >>= f) >>= g ≡ m >>= (\x -> f x >>= g)
		m := dot.Ok(42)
		f := func(x int) dot.Result[int] {
			return dot.Ok(x * 2)
		}
		g := func(x int) dot.Result[int] {
			return dot.Ok(x + 10)
		}

		left := dot.FlatMap(dot.FlatMap(m, f), g)
		right := dot.FlatMap(m, func(x int) dot.Result[int] {
			return dot.FlatMap(f(x), g)
		})

		assert.Equal(t, left.Unwrap(), right.Unwrap())
	})
}
