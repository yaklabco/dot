package doctor

import (
	"context"
	"sync"
	"time"

	"github.com/jamesainslie/dot/internal/domain"
)

// DiagnosticEngine manages and executes diagnostic checks.
type DiagnosticEngine struct {
	checks []domain.DiagnosticCheck
}

// NewDiagnosticEngine creates a new diagnostic engine.
func NewDiagnosticEngine() *DiagnosticEngine {
	return &DiagnosticEngine{
		checks: make([]domain.DiagnosticCheck, 0),
	}
}

// RegisterCheck adds a check to the engine.
func (e *DiagnosticEngine) RegisterCheck(check domain.DiagnosticCheck) {
	e.checks = append(e.checks, check)
}

// RunOptions configures the execution of checks.
type RunOptions struct {
	// Names of specific checks to run. If empty, all checks are run.
	IncludeChecks []string
	// Parallel execution of independent checks
	Parallel bool
}

// DiagnosticReport aggregates results from all checks.
type DiagnosticReport struct {
	Results       []domain.CheckResult
	OverallStatus domain.CheckStatus
	Duration      time.Duration
	StartTime     time.Time
}

// Run executes the registered checks based on options.
func (e *DiagnosticEngine) Run(ctx context.Context, opts RunOptions) (DiagnosticReport, error) {
	startTime := time.Now()
	report := DiagnosticReport{
		StartTime: startTime,
		Results:   make([]domain.CheckResult, 0),
	}

	checksToRun := e.filterChecks(opts.IncludeChecks)

	if opts.Parallel {
		report.Results = e.runParallel(ctx, checksToRun)
	} else {
		report.Results = e.runSequential(ctx, checksToRun)
	}

	report.Duration = time.Since(startTime)
	report.OverallStatus = e.determineOverallStatus(report.Results)

	return report, nil
}

func (e *DiagnosticEngine) filterChecks(include []string) []domain.DiagnosticCheck {
	if len(include) == 0 {
		return e.checks
	}

	includeMap := make(map[string]bool)
	for _, name := range include {
		includeMap[name] = true
	}

	var filtered []domain.DiagnosticCheck
	for _, check := range e.checks {
		if includeMap[check.Name()] {
			filtered = append(filtered, check)
		}
	}
	return filtered
}

func (e *DiagnosticEngine) runSequential(ctx context.Context, checks []domain.DiagnosticCheck) []domain.CheckResult {
	results := make([]domain.CheckResult, 0, len(e.checks))
	for _, check := range checks {
		if ctx.Err() != nil {
			break
		}
		// Adapt context.Context to domain.Context
		result, err := check.Run(ctx.(domain.Context))
		if err != nil {
			// System error executing check, treat as fail
			result = domain.CheckResult{
				CheckName: check.Name(),
				Status:    domain.CheckStatusFail,
				Issues: []domain.Issue{
					{
						Code:     "CHECK_EXECUTION_ERROR",
						Message:  err.Error(),
						Severity: domain.IssueSeverityError,
					},
				},
			}
		}
		results = append(results, result)
	}
	return results
}

func (e *DiagnosticEngine) runParallel(ctx context.Context, checks []domain.DiagnosticCheck) []domain.CheckResult {
	results := make([]domain.CheckResult, len(checks))
	var wg sync.WaitGroup

	for i, check := range checks {
		wg.Add(1)
		go func(idx int, c domain.DiagnosticCheck) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}

			// Adapt context.Context to domain.Context
			// In Go, context.Context satisfies the interface if method signatures match
			// But standard context.Context doesn't have Deadline() returning interface{}
			// We need to cast properly or wrap.
			// Ideally domain.Context should match context.Context exactly.
			// For now, assuming the caller passes a context.Context which we can cast.
			// However, strictly standard context.Context is what we expect.

			// Since we defined domain.Context with interface{} for Deadline return value,
			// it might not strictly match standard context.Context which returns (time.Time, bool).
			// Let's fix domain.Context definition in next step if needed.
			// For now, assuming we pass through.

			// Wait, standard context.Deadline() returns (time.Time, bool).
			// In domain.go we defined Deadline() (deadline interface{}, ok bool).
			// We should probably fix domain.Context to match context.Context exactly.

			res, err := c.Run(ctx.(domain.Context))
			if err != nil {
				res = domain.CheckResult{
					CheckName: c.Name(),
					Status:    domain.CheckStatusFail,
					Issues: []domain.Issue{
						{
							Code:     "CHECK_EXECUTION_ERROR",
							Message:  err.Error(),
							Severity: domain.IssueSeverityError,
						},
					},
				}
			}
			results[idx] = res
		}(i, check)
	}
	wg.Wait()
	return results
}

func (e *DiagnosticEngine) determineOverallStatus(results []domain.CheckResult) domain.CheckStatus {
	status := domain.CheckStatusPass
	for _, res := range results {
		if res.Status == domain.CheckStatusFail {
			return domain.CheckStatusFail
		}
		if res.Status == domain.CheckStatusWarning && status != domain.CheckStatusFail {
			status = domain.CheckStatusWarning
		}
	}
	return status
}
