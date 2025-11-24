package output

import (
	"fmt"
	"os"
	"time"

	"github.com/yaklabco/dot/internal/cli/render"
	"github.com/yaklabco/dot/internal/domain"
)

// Printer handles output formatting and display.
type Printer struct {
	colorEnabled bool
	verbose      int
	quiet        bool
}

// NewPrinter creates a new output printer.
func NewPrinter(colorEnabled bool, verbose int, quiet bool) *Printer {
	return &Printer{
		colorEnabled: colorEnabled,
		verbose:      verbose,
		quiet:        quiet,
	}
}

// PrintSuccess outputs a success message.
func (p *Printer) PrintSuccess(message string) {
	if p.quiet {
		return
	}
	if p.colorEnabled {
		fmt.Println(render.SuccessStyle(message))
	} else {
		fmt.Println(message)
	}
}

// PrintError outputs an error message to stderr.
func (p *Printer) PrintError(message string) {
	if p.colorEnabled {
		fmt.Fprintln(os.Stderr, render.ErrorStyle(message))
	} else {
		fmt.Fprintln(os.Stderr, message)
	}
}

// PrintWarning outputs a warning message.
func (p *Printer) PrintWarning(message string) {
	if p.quiet {
		return
	}
	if p.colorEnabled {
		fmt.Println(render.WarningStyle(message))
	} else {
		fmt.Println(message)
	}
}

// PrintInfo outputs an informational message.
func (p *Printer) PrintInfo(message string) {
	if p.quiet {
		return
	}
	if p.verbose >= 1 {
		if p.colorEnabled {
			fmt.Println(render.InfoStyle(message))
		} else {
			fmt.Println(message)
		}
	}
}

// PrintDebug outputs a debug message.
func (p *Printer) PrintDebug(message string) {
	if p.quiet {
		return
	}
	if p.verbose >= 2 {
		if p.colorEnabled {
			fmt.Println(render.DimStyle(message))
		} else {
			fmt.Println(message)
		}
	}
}

// ExecutionSummary contains execution result statistics.
type ExecutionSummary struct {
	PackageNames []string
	LinksCreated int
	DirsCreated  int
	Duration     time.Duration
	DryRun       bool
}

// PrintSummary outputs operation summary.
func (p *Printer) PrintSummary(summary ExecutionSummary) {
	if p.quiet {
		return
	}

	if summary.DryRun {
		if !p.quiet {
			if p.colorEnabled {
				fmt.Println(render.InfoStyle("Dry-run: No changes applied"))
			} else {
				fmt.Println("Dry-run: No changes applied")
			}
		}
		return
	}

	layout := render.NewLayoutAuto()

	fmt.Println()
	if p.colorEnabled {
		fmt.Println(render.SuccessStyle("Summary:"))
	} else {
		fmt.Println("Summary:")
	}

	if len(summary.PackageNames) > 0 {
		pkgList := layout.List(summary.PackageNames, "")
		fmt.Printf("  Packages:\n    %s\n", pkgList)
	}

	if summary.LinksCreated > 0 {
		fmt.Printf("  Links created: %d\n", summary.LinksCreated)
	}

	if summary.DirsCreated > 0 {
		fmt.Printf("  Directories created: %d\n", summary.DirsCreated)
	}

	if summary.Duration > 0 {
		fmt.Printf("  Duration: %s\n", summary.Duration.Round(time.Millisecond))
	}
}

// PrintDryRunSummary outputs dry-run summary.
func (p *Printer) PrintDryRunSummary(plan domain.Plan) {
	if p.quiet {
		return
	}

	fmt.Println()
	p.PrintWarning("Dry-run: No changes will be applied")
	fmt.Println()

	fmt.Println("Planned Operations:")
	opCount := len(plan.Operations)
	fmt.Printf("  Total operations: %d\n", opCount)

	if p.verbose >= 1 {
		// Count by operation type
		linkCount := 0
		dirCount := 0
		for _, op := range plan.Operations {
			switch op.Kind() {
			case domain.OpKindLinkCreate, domain.OpKindLinkDelete:
				linkCount++
			case domain.OpKindDirCreate, domain.OpKindDirDelete:
				dirCount++
			}
		}
		if linkCount > 0 {
			fmt.Printf("  Link operations: %d\n", linkCount)
		}
		if dirCount > 0 {
			fmt.Printf("  Directory operations: %d\n", dirCount)
		}
	}

	fmt.Println()
	fmt.Println("Run without --dry-run to apply changes.")
}
