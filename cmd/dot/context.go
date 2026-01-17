package main

import "context"

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
