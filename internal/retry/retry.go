// Package retry provides utilities for retrying operations with exponential backoff.
package retry

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Config configures retry behavior.
type Config struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int
	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the factor by which delay increases after each attempt
	Multiplier float64
	// Jitter is the maximum random variation added to delays (0.0 to 1.0)
	Jitter float64
}

// DefaultConfig returns a default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.3,
	}
}

// Do executes the given function with retry logic.
// The function is retried up to MaxAttempts times with exponential backoff and jitter.
// Returns nil if the function succeeds, or the last error if all attempts fail.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// First attempt: no delay
		// Subsequent attempts: delay with jitter
		if attempt > 1 {
			// Calculate jitter: random value between -jitter and +jitter
			// #nosec G404 -- Using math/rand for jitter is acceptable (not security-sensitive)
			jitterAmount := time.Duration(float64(delay) * cfg.Jitter * (rand.Float64()*2 - 1))
			sleepDuration := delay + jitterAmount

			// Ensure sleep duration is non-negative
			if sleepDuration < 0 {
				sleepDuration = 0
			}

			// Wait for the delay or context cancellation
			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Execute the function
		if err := fn(); err != nil {
			lastErr = err

			// If this was the last attempt, return the error
			if attempt >= cfg.MaxAttempts {
				break
			}

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}

			continue
		}

		// Success!
		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// DoWithData executes a function that returns data, with retry logic.
// Similar to Do but for functions that return values.
func DoWithData[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var lastErr error
	var result T
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if attempt > 1 {
			// #nosec G404 -- Using math/rand for jitter is acceptable (not security-sensitive)
			jitterAmount := time.Duration(float64(delay) * cfg.Jitter * (rand.Float64()*2 - 1))
			sleepDuration := delay + jitterAmount

			if sleepDuration < 0 {
				sleepDuration = 0
			}

			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}

		if data, err := fn(); err != nil {
			lastErr = err

			if attempt >= cfg.MaxAttempts {
				break
			}

			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}

			continue
		} else {
			return data, nil
		}
	}

	return result, fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}
