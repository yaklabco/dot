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
// Note: Run always returns nil error. Individual check failures are encoded
// as CheckResult objects with Status=CheckStatusFail or issues with code
// CHECK_EXECUTION_ERROR. Callers should inspect DiagnosticReport.Results
// for per-check failures rather than relying on the returned error.
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
		result, err := check.Run(ctx)
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
			// If context cancelled before check starts, set a skipped result
			if ctx.Err() != nil {
				results[idx] = domain.CheckResult{
					CheckName: c.Name(),
					Status:    domain.CheckStatusSkipped,
					Issues: []domain.Issue{
						{
							Code:     "CHECK_CANCELLED",
							Message:  "Check cancelled due to context cancellation",
							Severity: domain.IssueSeverityWarning,
						},
					},
				}
				return
			}

			res, err := c.Run(ctx)
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

// determineOverallStatus computes aggregate status from check results.
// Status priority: Fail > Warning > Pass. Skipped checks don't affect status.
// Note: Unknown or future status values are treated as Pass.
func (e *DiagnosticEngine) determineOverallStatus(results []domain.CheckResult) domain.CheckStatus {
	status := domain.CheckStatusPass
	for _, res := range results {
		switch res.Status {
		case domain.CheckStatusFail:
			return domain.CheckStatusFail
		case domain.CheckStatusWarning:
			if status != domain.CheckStatusFail {
				status = domain.CheckStatusWarning
			}
		case domain.CheckStatusSkipped:
			// Skipped checks don't affect overall status
			continue
		case domain.CheckStatusPass:
			// Pass doesn't change status
			continue
		default:
			// Unknown status values treated as pass (no-op)
			continue
		}
	}
	return status
}
