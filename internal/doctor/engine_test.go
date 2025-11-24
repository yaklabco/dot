package doctor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/internal/domain"
)

// mockCheck is a test check implementation
type mockCheck struct {
	name        string
	description string
	result      domain.CheckResult
	err         error
	delay       time.Duration
}

func (m *mockCheck) Name() string        { return m.name }
func (m *mockCheck) Description() string { return m.description }
func (m *mockCheck) Run(ctx context.Context) (domain.CheckResult, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
			// Delay completed
		case <-ctx.Done():
			// Context cancelled during delay
			return domain.CheckResult{}, ctx.Err()
		}
	}
	return m.result, m.err
}

func TestNewDiagnosticEngine(t *testing.T) {
	engine := NewDiagnosticEngine()
	assert.NotNil(t, engine)
	assert.Empty(t, engine.checks)
}

func TestDiagnosticEngine_RegisterCheck(t *testing.T) {
	engine := NewDiagnosticEngine()
	check := &mockCheck{name: "test-check", description: "Test check"}

	engine.RegisterCheck(check)

	assert.Len(t, engine.checks, 1)
	assert.Equal(t, check, engine.checks[0])
}

func TestDiagnosticEngine_Run_SingleCheck_Pass(t *testing.T) {
	engine := NewDiagnosticEngine()
	check := &mockCheck{
		name:        "passing-check",
		description: "A check that passes",
		result: domain.CheckResult{
			CheckName: "passing-check",
			Status:    domain.CheckStatusPass,
			Issues:    []domain.Issue{},
			Stats:     map[string]any{"count": 42},
		},
	}
	engine.RegisterCheck(check)

	report, err := engine.Run(context.Background(), RunOptions{})

	require.NoError(t, err)
	assert.Len(t, report.Results, 1)
	assert.Equal(t, domain.CheckStatusPass, report.Results[0].Status)
	assert.Equal(t, domain.CheckStatusPass, report.OverallStatus)
	assert.Equal(t, 42, report.Results[0].Stats["count"])
}

func TestDiagnosticEngine_Run_SingleCheck_Fail(t *testing.T) {
	engine := NewDiagnosticEngine()
	check := &mockCheck{
		name:        "failing-check",
		description: "A check that fails",
		result: domain.CheckResult{
			CheckName: "failing-check",
			Status:    domain.CheckStatusFail,
			Issues: []domain.Issue{
				{
					Code:     "TEST_FAIL",
					Message:  "Test failure",
					Severity: domain.IssueSeverityError,
				},
			},
			Stats: map[string]any{},
		},
	}
	engine.RegisterCheck(check)

	report, err := engine.Run(context.Background(), RunOptions{})

	require.NoError(t, err)
	assert.Len(t, report.Results, 1)
	assert.Equal(t, domain.CheckStatusFail, report.Results[0].Status)
	assert.Equal(t, domain.CheckStatusFail, report.OverallStatus)
	assert.Len(t, report.Results[0].Issues, 1)
}

func TestDiagnosticEngine_Run_CheckError(t *testing.T) {
	engine := NewDiagnosticEngine()
	expectedErr := errors.New("check execution error")
	check := &mockCheck{
		name:        "error-check",
		description: "A check that errors",
		err:         expectedErr,
	}
	engine.RegisterCheck(check)

	report, err := engine.Run(context.Background(), RunOptions{})

	require.NoError(t, err)
	require.Len(t, report.Results, 1)
	assert.Equal(t, domain.CheckStatusFail, report.Results[0].Status)
	assert.Contains(t, report.Results[0].Issues[0].Message, "check execution error")
}

func TestDiagnosticEngine_Run_MultipleChecks_OverallStatus(t *testing.T) {
	tests := []struct {
		name           string
		checkStatuses  []domain.CheckStatus
		expectedStatus domain.CheckStatus
	}{
		{
			name:           "all pass",
			checkStatuses:  []domain.CheckStatus{domain.CheckStatusPass, domain.CheckStatusPass},
			expectedStatus: domain.CheckStatusPass,
		},
		{
			name:           "one warning",
			checkStatuses:  []domain.CheckStatus{domain.CheckStatusPass, domain.CheckStatusWarning},
			expectedStatus: domain.CheckStatusWarning,
		},
		{
			name:           "one fail",
			checkStatuses:  []domain.CheckStatus{domain.CheckStatusPass, domain.CheckStatusFail},
			expectedStatus: domain.CheckStatusFail,
		},
		{
			name: "fail takes precedence over warning",
			checkStatuses: []domain.CheckStatus{
				domain.CheckStatusWarning,
				domain.CheckStatusFail,
				domain.CheckStatusPass,
			},
			expectedStatus: domain.CheckStatusFail,
		},
		{
			name:           "skipped with pass",
			checkStatuses:  []domain.CheckStatus{domain.CheckStatusPass, domain.CheckStatusSkipped},
			expectedStatus: domain.CheckStatusPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewDiagnosticEngine()
			for i, status := range tt.checkStatuses {
				check := &mockCheck{
					name:        string(status),
					description: "Test check",
					result: domain.CheckResult{
						CheckName: string(status),
						Status:    status,
						Issues:    []domain.Issue{},
						Stats:     map[string]any{"index": i},
					},
				}
				engine.RegisterCheck(check)
			}

			report, err := engine.Run(context.Background(), RunOptions{})

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, report.OverallStatus)
			assert.Len(t, report.Results, len(tt.checkStatuses))
		})
	}
}

func TestDiagnosticEngine_Run_Parallel(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Add multiple checks with delays to verify parallel execution
	for i := 0; i < 3; i++ {
		check := &mockCheck{
			name:        string(rune('A' + i)),
			description: "Parallel check",
			delay:       50 * time.Millisecond,
			result: domain.CheckResult{
				CheckName: string(rune('A' + i)),
				Status:    domain.CheckStatusPass,
				Issues:    []domain.Issue{},
				Stats:     map[string]any{},
			},
		}
		engine.RegisterCheck(check)
	}

	start := time.Now()
	report, err := engine.Run(context.Background(), RunOptions{Parallel: true})
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, report.Results, 3)
	// Parallel execution should take roughly the time of the longest check,
	// not the sum of all checks. Allow some overhead.
	assert.Less(t, duration, 150*time.Millisecond, "parallel execution took too long")
}

func TestDiagnosticEngine_Run_IncludeChecks(t *testing.T) {
	engine := NewDiagnosticEngine()

	check1 := &mockCheck{
		name:        "check-a",
		description: "Check A",
		result: domain.CheckResult{
			CheckName: "check-a",
			Status:    domain.CheckStatusPass,
			Issues:    []domain.Issue{},
			Stats:     map[string]any{},
		},
	}
	check2 := &mockCheck{
		name:        "check-b",
		description: "Check B",
		result: domain.CheckResult{
			CheckName: "check-b",
			Status:    domain.CheckStatusPass,
			Issues:    []domain.Issue{},
			Stats:     map[string]any{},
		},
	}

	engine.RegisterCheck(check1)
	engine.RegisterCheck(check2)

	report, err := engine.Run(context.Background(), RunOptions{
		IncludeChecks: []string{"check-a"},
	})

	require.NoError(t, err)
	assert.Len(t, report.Results, 1)
	assert.Equal(t, "check-a", report.Results[0].CheckName)
}

func TestDiagnosticEngine_Run_NoChecks(t *testing.T) {
	engine := NewDiagnosticEngine()

	report, err := engine.Run(context.Background(), RunOptions{})

	require.NoError(t, err)
	assert.Empty(t, report.Results)
	assert.Equal(t, domain.CheckStatusPass, report.OverallStatus)
}

func TestDiagnosticEngine_Run_ContextCancellation(t *testing.T) {
	engine := NewDiagnosticEngine()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	check := &mockCheck{
		name:        "slow-check",
		description: "A slow check",
		delay:       1 * time.Second,
		result: domain.CheckResult{
			CheckName: "slow-check",
			Status:    domain.CheckStatusPass,
			Issues:    []domain.Issue{},
			Stats:     map[string]any{},
		},
	}
	engine.RegisterCheck(check)

	report, err := engine.Run(ctx, RunOptions{})

	// Engine handles cancellation gracefully by stopping execution
	require.NoError(t, err)
	// With timeout, check returns error which becomes a failed check result
	require.Len(t, report.Results, 1)
	assert.Equal(t, domain.CheckStatusFail, report.Results[0].Status)
	assert.Contains(t, report.Results[0].Issues[0].Message, "context deadline exceeded")
}
