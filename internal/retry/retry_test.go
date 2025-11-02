package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo_Success(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
	}

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := Do(ctx, cfg, fn)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "should succeed on first attempt")
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := Do(ctx, cfg, fn)
	require.NoError(t, err)
	assert.Equal(t, 3, callCount, "should succeed on third attempt")
}

func TestDo_AllAttemptsFail(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
	}

	callCount := 0
	expectedErr := errors.New("persistent error")
	fn := func() error {
		callCount++
		return expectedErr
	}

	err := Do(ctx, cfg, fn)
	require.Error(t, err)
	assert.Equal(t, 3, callCount, "should attempt all retries")
	assert.Contains(t, err.Error(), "failed after 3 attempts")
	assert.ErrorIs(t, err, expectedErr)
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := Config{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 2 {
			// Cancel context after second attempt
			cancel()
		}
		return errors.New("temporary error")
	}

	err := Do(ctx, cfg, fn)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.LessOrEqual(t, callCount, 2, "should stop retrying after context cancellation")
}

func TestDo_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  4,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.0, // No jitter for predictable timing
	}

	start := time.Now()
	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("temporary error")
	}

	_ = Do(ctx, cfg, fn)

	elapsed := time.Since(start)
	// Expected delays: 0ms + 10ms + 20ms + 40ms = 70ms
	// With some tolerance for test execution time
	assert.GreaterOrEqual(t, elapsed, 70*time.Millisecond, "should respect exponential backoff")
	assert.Equal(t, 4, callCount)
}

func TestDo_MaxDelay(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     20 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.0,
	}

	start := time.Now()
	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("temporary error")
	}

	_ = Do(ctx, cfg, fn)

	elapsed := time.Since(start)
	// Expected delays capped at MaxDelay: 0ms + 10ms + 20ms + 20ms + 20ms = 70ms
	assert.GreaterOrEqual(t, elapsed, 70*time.Millisecond, "should cap at MaxDelay")
	assert.Less(t, elapsed, 100*time.Millisecond, "should not exceed expected delay")
}

func TestDoWithData_Success(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.InitialDelay = 10 * time.Millisecond

	result, err := DoWithData(ctx, cfg, func() (string, error) {
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestDoWithData_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		if callCount < 3 {
			return 0, errors.New("temporary error")
		}
		return 42, nil
	}

	result, err := DoWithData(ctx, cfg, fn)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, callCount)
}

func TestDoWithData_AllAttemptsFail(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
	}

	result, err := DoWithData(ctx, cfg, func() (string, error) {
		return "", errors.New("persistent error")
	})

	require.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "failed after 2 attempts")
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.Equal(t, 1*time.Second, cfg.InitialDelay)
	assert.Equal(t, 10*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.Multiplier)
	assert.Equal(t, 0.3, cfg.Jitter)
}

func TestDo_NegativeJitter(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxAttempts:  3,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.5, // Large jitter that could produce negative delays
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	// Should not panic even with jitter that could produce negative delays
	err := Do(ctx, cfg, fn)
	require.NoError(t, err)
	assert.Equal(t, 3, callCount)
}
