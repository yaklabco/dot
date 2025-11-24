package domain

import (
	"context"
)

// DiagnosticCheck represents a pluggable health check.
type DiagnosticCheck interface {
	// Name returns the unique identifier for this check.
	Name() string
	// Description returns a human-readable description of what this check does.
	Description() string
	// Run executes the check and returns the result.
	Run(ctx context.Context) (CheckResult, error)
}

// CheckResult contains the outcome of a diagnostic check.
type CheckResult struct {
	CheckName string
	Status    CheckStatus
	Issues    []Issue
	Stats     map[string]any
}

// CheckStatus represents the high-level outcome of a check.
type CheckStatus string

const (
	CheckStatusPass    CheckStatus = "pass"
	CheckStatusWarning CheckStatus = "warning"
	CheckStatusFail    CheckStatus = "fail"
	CheckStatusSkipped CheckStatus = "skipped"
)

// Issue represents a specific problem found during a check.
type Issue struct {
	Code        string
	Message     string
	Severity    IssueSeverity
	Path        string // Optional: associated file path
	Context     map[string]any
	Remediation *Remediation
}

// IssueSeverity indicates the impact of an issue.
type IssueSeverity string

const (
	IssueSeverityInfo    IssueSeverity = "info"
	IssueSeverityWarning IssueSeverity = "warning"
	IssueSeverityError   IssueSeverity = "error"
	IssueSeverityFatal   IssueSeverity = "fatal"
)

// Remediation describes how to fix an issue.
type Remediation struct {
	Description string
	// Action is a function that attempts to fix the issue.
	// It returns an error if the fix failed.
	Action func(context.Context) error
}

// Context is an alias for context.Context to maintain interface compatibility.
//
// Deprecated: Use context.Context directly instead.
type Context = context.Context
