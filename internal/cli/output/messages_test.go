package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

func TestNewPrinter(t *testing.T) {
	printer := NewPrinter(true, 1, false)
	assert.NotNil(t, printer)
	assert.True(t, printer.colorEnabled)
	assert.Equal(t, 1, printer.verbose)
	assert.False(t, printer.quiet)
}

func TestPrinter_QuietMode(t *testing.T) {
	printer := NewPrinter(true, 2, true)

	// In quiet mode, these should not panic
	printer.PrintSuccess("success")
	printer.PrintWarning("warning")
	printer.PrintInfo("info")
	printer.PrintDebug("debug")
}

func TestPrinter_VerbosityLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   int
		quiet   bool
		methods []func(*Printer)
	}{
		{
			name:  "level 0",
			level: 0,
			quiet: false,
			methods: []func(*Printer){
				func(p *Printer) { p.PrintSuccess("test") },
			},
		},
		{
			name:  "level 1",
			level: 1,
			quiet: false,
			methods: []func(*Printer){
				func(p *Printer) { p.PrintInfo("test") },
			},
		},
		{
			name:  "level 2",
			level: 2,
			quiet: false,
			methods: []func(*Printer){
				func(p *Printer) { p.PrintDebug("test") },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := NewPrinter(false, tt.level, tt.quiet)
			// Should not panic
			for _, method := range tt.methods {
				method(printer)
			}
		})
	}
}

func TestExecutionSummary_AllFields(t *testing.T) {
	summary := ExecutionSummary{
		PackageNames: []string{"vim", "tmux"},
		LinksCreated: 42,
		DirsCreated:  7,
		Duration:     2 * time.Second,
		DryRun:       false,
	}

	assert.Equal(t, 2, len(summary.PackageNames))
	assert.Equal(t, 42, summary.LinksCreated)
	assert.Equal(t, 7, summary.DirsCreated)
	assert.Equal(t, 2*time.Second, summary.Duration)
	assert.False(t, summary.DryRun)
}

func TestPrinter_PrintSummary(t *testing.T) {
	printer := NewPrinter(false, 1, false)
	summary := ExecutionSummary{
		PackageNames: []string{"vim"},
		LinksCreated: 10,
		Duration:     time.Second,
	}

	// Should not panic
	printer.PrintSummary(summary)
}

func TestPrinter_PrintSummary_DryRun(t *testing.T) {
	printer := NewPrinter(false, 1, false)
	summary := ExecutionSummary{
		DryRun: true,
	}

	// Should not panic
	printer.PrintSummary(summary)
}

func TestPrinter_PrintSummary_Quiet(t *testing.T) {
	printer := NewPrinter(false, 1, true)
	summary := ExecutionSummary{
		PackageNames: []string{"vim"},
		LinksCreated: 10,
	}

	// In quiet mode, should not panic
	printer.PrintSummary(summary)
}

func TestPrinter_PrintDryRunSummary(t *testing.T) {
	printer := NewPrinter(false, 0, false)
	plan := domain.Plan{
		Operations: []domain.Operation{
			domain.LinkCreate{
				Source: domain.MustParsePath("/src"),
				Target: domain.MustParseTargetPath("/tgt"),
			},
		},
	}

	// Should not panic
	printer.PrintDryRunSummary(plan)
}

func TestPrinter_PrintDryRunSummary_Verbose(t *testing.T) {
	printer := NewPrinter(false, 1, false)
	plan := domain.Plan{
		Operations: []domain.Operation{
			domain.LinkCreate{
				Source: domain.MustParsePath("/src"),
				Target: domain.MustParseTargetPath("/tgt"),
			},
			domain.DirCreate{
				Path: domain.MustParsePath("/dir"),
			},
		},
	}

	// Should not panic
	printer.PrintDryRunSummary(plan)
}

func TestPrinter_PrintDryRunSummary_Quiet(t *testing.T) {
	printer := NewPrinter(false, 1, true)
	plan := domain.Plan{
		Operations: []domain.Operation{
			domain.LinkCreate{
				Source: domain.MustParsePath("/src"),
				Target: domain.MustParseTargetPath("/tgt"),
			},
		},
	}

	// In quiet mode, should not panic
	printer.PrintDryRunSummary(plan)
}
