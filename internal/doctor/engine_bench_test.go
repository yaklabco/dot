package doctor

import (
	"context"
	"testing"

	"github.com/yaklabco/dot/internal/domain"
)

// BenchmarkDiagnosticEngine_Sequential benchmarks sequential execution.
func BenchmarkDiagnosticEngine_Sequential(b *testing.B) {
	engine := NewDiagnosticEngine()

	// Register 5 checks
	for i := 0; i < 5; i++ {
		check := &mockCheck{
			name:        string(rune('A' + i)),
			description: "Benchmark check",
			result: domain.CheckResult{
				CheckName: string(rune('A' + i)),
				Status:    domain.CheckStatusPass,
				Issues:    []domain.Issue{},
				Stats:     map[string]any{"iterations": 100},
			},
		}
		engine.RegisterCheck(check)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Run(context.Background(), RunOptions{Parallel: false})
	}
}

// BenchmarkDiagnosticEngine_Parallel benchmarks parallel execution.
func BenchmarkDiagnosticEngine_Parallel(b *testing.B) {
	engine := NewDiagnosticEngine()

	// Register 5 checks
	for i := 0; i < 5; i++ {
		check := &mockCheck{
			name:        string(rune('A' + i)),
			description: "Benchmark check",
			result: domain.CheckResult{
				CheckName: string(rune('A' + i)),
				Status:    domain.CheckStatusPass,
				Issues:    []domain.Issue{},
				Stats:     map[string]any{"iterations": 100},
			},
		}
		engine.RegisterCheck(check)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Run(context.Background(), RunOptions{Parallel: true})
	}
}

// BenchmarkDiagnosticEngine_FastMode benchmarks fast mode with essential checks only.
func BenchmarkDiagnosticEngine_FastMode(b *testing.B) {
	engine := NewDiagnosticEngine()

	// Fast mode: only 2 essential checks
	for i := 0; i < 2; i++ {
		check := &mockCheck{
			name:        string(rune('A' + i)),
			description: "Fast check",
			result: domain.CheckResult{
				CheckName: string(rune('A' + i)),
				Status:    domain.CheckStatusPass,
				Issues:    []domain.Issue{},
				Stats:     map[string]any{},
			},
		}
		engine.RegisterCheck(check)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Run(context.Background(), RunOptions{Parallel: true})
	}
}

// BenchmarkDiagnosticEngine_DeepMode benchmarks deep mode with all checks.
func BenchmarkDiagnosticEngine_DeepMode(b *testing.B) {
	engine := NewDiagnosticEngine()

	// Deep mode: 10 comprehensive checks
	for i := 0; i < 10; i++ {
		check := &mockCheck{
			name:        string(rune('A' + i)),
			description: "Deep check",
			result: domain.CheckResult{
				CheckName: string(rune('A' + i)),
				Status:    domain.CheckStatusPass,
				Issues:    []domain.Issue{},
				Stats:     map[string]any{"scanned_paths": 1000},
			},
		}
		engine.RegisterCheck(check)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Run(context.Background(), RunOptions{Parallel: true})
	}
}

// BenchmarkDiagnosticEngine_WithIssues benchmarks with issues in results.
func BenchmarkDiagnosticEngine_WithIssues(b *testing.B) {
	engine := NewDiagnosticEngine()

	// Checks that report issues
	for i := 0; i < 5; i++ {
		issues := make([]domain.Issue, 10)
		for j := 0; j < 10; j++ {
			issues[j] = domain.Issue{
				Code:     "TEST_ISSUE",
				Message:  "Test issue message",
				Severity: domain.IssueSeverityWarning,
				Path:     "/test/path",
				Context:  map[string]any{"index": j},
			}
		}

		check := &mockCheck{
			name:        string(rune('A' + i)),
			description: "Check with issues",
			result: domain.CheckResult{
				CheckName: string(rune('A' + i)),
				Status:    domain.CheckStatusWarning,
				Issues:    issues,
				Stats:     map[string]any{},
			},
		}
		engine.RegisterCheck(check)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Run(context.Background(), RunOptions{Parallel: true})
	}
}
