package main

import (
	"context"

	"github.com/yaklabco/dot/pkg/dot"
)

// cliFlagsKey is the context key for CLIFlags.
type cliFlagsKey struct{}

// WithCLIFlags adds CLIFlags to the context.
func WithCLIFlags(ctx context.Context, flags *CLIFlags) context.Context {
	return context.WithValue(ctx, cliFlagsKey{}, flags)
}

// CLIFlagsFromContext retrieves CLIFlags from context.
// Returns nil if flags are not set in the context.
func CLIFlagsFromContext(ctx context.Context) *CLIFlags {
	if ctx == nil {
		return nil
	}
	if flags, ok := ctx.Value(cliFlagsKey{}).(*CLIFlags); ok {
		return flags
	}
	return nil
}

// MustCLIFlagsFromContext retrieves CLIFlags from context or panics.
// Use this when flags are required and must be present.
func MustCLIFlagsFromContext(ctx context.Context) *CLIFlags {
	flags := CLIFlagsFromContext(ctx)
	if flags == nil {
		panic("CLIFlags not set in context")
	}
	return flags
}

// doctorResultKey is the context key for DoctorResultHolder.
type doctorResultKey struct{}

// DoctorResultHolder holds the doctor command result.
// This is stored as a pointer in the context so it can be modified during execution.
type DoctorResultHolder struct {
	Executed bool
	Status   dot.HealthStatus
}

// WithDoctorResultHolder adds a DoctorResultHolder to the context.
func WithDoctorResultHolder(ctx context.Context, holder *DoctorResultHolder) context.Context {
	return context.WithValue(ctx, doctorResultKey{}, holder)
}

// DoctorResultHolderFromContext retrieves DoctorResultHolder from context.
// Returns nil if holder is not set in the context.
func DoctorResultHolderFromContext(ctx context.Context) *DoctorResultHolder {
	if ctx == nil {
		return nil
	}
	if holder, ok := ctx.Value(doctorResultKey{}).(*DoctorResultHolder); ok {
		return holder
	}
	return nil
}
