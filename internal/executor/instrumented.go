package executor

import (
	"context"
	"time"

	"github.com/yaklabco/dot/internal/domain"
)

// InstrumentedExecutor wraps Executor with metrics collection.
type InstrumentedExecutor struct {
	inner   *Executor
	metrics domain.Metrics
}

// NewInstrumented creates an executor with metrics instrumentation.
func NewInstrumented(inner *Executor, metrics domain.Metrics) *InstrumentedExecutor {
	return &InstrumentedExecutor{
		inner:   inner,
		metrics: metrics,
	}
}

// Execute executes a plan with metrics collection.
func (e *InstrumentedExecutor) Execute(ctx context.Context, plan domain.Plan) domain.Result[ExecutionResult] {
	start := time.Now()

	e.metrics.Counter("executor.executions.total").Inc()
	e.metrics.Gauge("executor.operations.queued").Set(float64(len(plan.Operations)))

	result := e.inner.Execute(ctx, plan)

	duration := time.Since(start)
	e.metrics.Histogram("executor.duration.seconds").Observe(duration.Seconds())

	if result.IsOk() {
		execResult := result.Unwrap()
		e.metrics.Counter("executor.executions.success").Inc()
		e.metrics.Counter("executor.operations.executed").Add(float64(len(execResult.Executed)))

		if len(plan.Batches) > 0 {
			e.metrics.Histogram("executor.parallel.batches").Observe(float64(len(plan.Batches)))
		}
	} else {
		e.metrics.Counter("executor.executions.failed").Inc()

		execResult := result.UnwrapOr(ExecutionResult{})
		if len(execResult.Failed) > 0 {
			e.metrics.Counter("executor.operations.failed").Add(float64(len(execResult.Failed)))
		}
		if len(execResult.RolledBack) > 0 {
			e.metrics.Counter("executor.operations.rolled_back").Add(float64(len(execResult.RolledBack)))
		}
	}

	return result
}
