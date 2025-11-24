package renderer

import (
	"fmt"
	"io"
	"sort"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/pkg/dot"
)

// TextRenderer renders output as human-readable plain text.
type TextRenderer struct {
	colorize     bool
	scheme       ColorScheme
	width        int
	displayLimit int // Maximum number of items to display before truncation
}

// RenderStatus renders installation status as plain text.
func (r *TextRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	if len(status.Packages) == 0 {
		fmt.Fprintln(w, "No packages installed")
		return nil
	}

	// Sort packages by name for consistent output
	sort.Slice(status.Packages, func(i, j int) bool {
		return status.Packages[i].Name < status.Packages[j].Name
	})

	for _, pkg := range status.Packages {
		fmt.Fprintf(w, "%s%s%s\n", r.colorText(r.scheme.Info), pkg.Name, r.resetColor())
		fmt.Fprintf(w, "  Links: %d\n", pkg.LinkCount)
		fmt.Fprintf(w, "  Installed: %s\n", formatDuration(pkg.InstalledAt))

		if len(pkg.Links) > 0 {
			// Sort links for consistent output
			links := make([]string, len(pkg.Links))
			copy(links, pkg.Links)
			sort.Strings(links)

			fmt.Fprintf(w, "  Files:\n")
			for i, link := range links {
				// If displayLimit is 0, show all; otherwise respect the limit
				if r.displayLimit > 0 && i >= r.displayLimit {
					remaining := len(pkg.Links) - r.displayLimit
					fmt.Fprintf(w, "    ... and %d more\n", remaining)
					break
				}
				fmt.Fprintf(w, "    %s\n", link)
			}
		}
		fmt.Fprintln(w)
	}

	return nil
}

func (r *TextRenderer) colorText(color string) string {
	if r.colorize && color != "" {
		return color
	}
	return ""
}

func (r *TextRenderer) resetColor() string {
	if r.colorize {
		return "\033[0m"
	}
	return ""
}

// RenderDiagnostics renders diagnostic report as plain text.
func (r *TextRenderer) RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error {
	// Show overall health
	healthColor := r.scheme.Success
	healthSymbol := "✓"
	if report.OverallHealth == dot.HealthWarnings {
		healthColor = r.scheme.Warning
		healthSymbol = "⚠"
	} else if report.OverallHealth == dot.HealthErrors {
		healthColor = r.scheme.Error
		healthSymbol = "✗"
	}

	fmt.Fprintf(w, "%s%s Health Status: %s%s\n\n", r.colorText(healthColor), healthSymbol, report.OverallHealth.String(), r.resetColor())

	// Show statistics
	fmt.Fprintf(w, "Statistics:\n")
	fmt.Fprintf(w, "  Total Links: %d\n", report.Statistics.TotalLinks)
	fmt.Fprintf(w, "  Managed Links: %d\n", report.Statistics.ManagedLinks)
	if report.Statistics.BrokenLinks > 0 {
		fmt.Fprintf(w, "  %sBroken Links: %d%s\n", r.colorText(r.scheme.Error), report.Statistics.BrokenLinks, r.resetColor())
	}
	if report.Statistics.OrphanedLinks > 0 {
		fmt.Fprintf(w, "  %sOrphaned Links: %d%s\n", r.colorText(r.scheme.Warning), report.Statistics.OrphanedLinks, r.resetColor())
	}
	fmt.Fprintln(w)

	// Show issues
	if len(report.Issues) == 0 {
		fmt.Fprintf(w, "%sNo issues found%s\n", r.colorText(r.scheme.Success), r.resetColor())
		return nil
	}

	fmt.Fprintf(w, "Issues Found: %d\n\n", len(report.Issues))

	for i, issue := range report.Issues {
		severityColor := r.scheme.Info
		severitySymbol := "ℹ"
		if issue.Severity == dot.SeverityWarning {
			severityColor = r.scheme.Warning
			severitySymbol = "⚠"
		} else if issue.Severity == dot.SeverityError {
			severityColor = r.scheme.Error
			severitySymbol = "✗"
		}

		fmt.Fprintf(w, "%d. %s%s %s%s\n", i+1, r.colorText(severityColor), severitySymbol, issue.Severity.String(), r.resetColor())
		fmt.Fprintf(w, "   Type: %s\n", issue.Type.String())
		if issue.Path != "" {
			fmt.Fprintf(w, "   Path: %s\n", issue.Path)
		}
		fmt.Fprintf(w, "   %s\n", issue.Message)
		if issue.Suggestion != "" {
			fmt.Fprintf(w, "   %sSuggestion:%s %s\n", r.colorText(r.scheme.Info), r.resetColor(), issue.Suggestion)
		}
		fmt.Fprintln(w)
	}

	return nil
}

// RenderPlan renders an execution plan as plain text.
func (r *TextRenderer) RenderPlan(w io.Writer, plan domain.Plan) error {
	// Header
	fmt.Fprintf(w, "%sDry run mode - no changes will be applied%s\n\n", r.colorText(r.scheme.Warning), r.resetColor())

	// Plan operations
	fmt.Fprintln(w, "Plan:")
	if len(plan.Operations) == 0 {
		fmt.Fprintln(w, "  No operations required")
	} else {
		for _, op := range plan.Operations {
			r.renderOperation(w, op)
		}
	}
	fmt.Fprintln(w)

	// Summary counts
	fmt.Fprintln(w, "Summary:")
	counts := r.countOperations(plan)
	if counts.DirCreate > 0 {
		fmt.Fprintf(w, "  Directories: %d\n", counts.DirCreate)
	}
	if counts.LinkCreate > 0 {
		fmt.Fprintf(w, "  Symlinks: %d\n", counts.LinkCreate)
	}
	if counts.FileMove > 0 {
		fmt.Fprintf(w, "  File moves: %d\n", counts.FileMove)
	}
	if counts.FileBackup > 0 {
		fmt.Fprintf(w, "  Backups: %d\n", counts.FileBackup)
	}
	if counts.DirDelete > 0 {
		fmt.Fprintf(w, "  Directory deletions: %d\n", counts.DirDelete)
	}
	if counts.LinkDelete > 0 {
		fmt.Fprintf(w, "  Symlink deletions: %d\n", counts.LinkDelete)
	}

	if len(plan.Metadata.Conflicts) > 0 {
		fmt.Fprintf(w, "  %sConflicts: %d%s\n", r.colorText(r.scheme.Error), len(plan.Metadata.Conflicts), r.resetColor())
	} else {
		fmt.Fprintf(w, "  Conflicts: 0\n")
	}

	return nil
}

// renderOperation renders a single operation.
func (r *TextRenderer) renderOperation(w io.Writer, op domain.Operation) {
	symbol := r.colorText(r.scheme.Success) + "+" + r.resetColor()

	// Normalize: dereference pointers to get value type for switching
	normalized := normalizeOperation(op)

	switch typed := normalized.(type) {
	case domain.DirCreate:
		fmt.Fprintf(w, "  %s Create directory: %s\n", symbol, typed.Path.String())

	case domain.LinkCreate:
		fmt.Fprintf(w, "  %s Create symlink: %s -> %s\n", symbol, typed.Target.String(), typed.Source.String())

	case domain.FileMove:
		fmt.Fprintf(w, "  %s Move file: %s -> %s\n", symbol, typed.Source.String(), typed.Dest.String())

	case domain.FileBackup:
		fmt.Fprintf(w, "  %s Backup file: %s -> %s\n", symbol, typed.Source.String(), typed.Backup.String())

	case domain.DirDelete:
		deleteSymbol := r.colorText(r.scheme.Error) + "-" + r.resetColor()
		fmt.Fprintf(w, "  %s Delete directory: %s\n", deleteSymbol, typed.Path.String())

	case domain.LinkDelete:
		deleteSymbol := r.colorText(r.scheme.Error) + "-" + r.resetColor()
		fmt.Fprintf(w, "  %s Delete symlink: %s\n", deleteSymbol, typed.Target.String())

	default:
		// Handle unknown operation types with clear, informative output
		fmt.Fprintf(w, "  %s Unknown operation: %T - %s\n", symbol, op, op.String())
	}
}

// operationCounts holds counts of different operation types.
type operationCounts struct {
	DirCreate  int
	DirDelete  int
	LinkCreate int
	LinkDelete int
	FileMove   int
	FileBackup int
}

// countOperations counts operations by type.
func (r *TextRenderer) countOperations(plan domain.Plan) operationCounts {
	var counts operationCounts

	for _, op := range plan.Operations {
		switch op.Kind() {
		case domain.OpKindDirCreate:
			counts.DirCreate++
		case domain.OpKindDirDelete:
			counts.DirDelete++
		case domain.OpKindLinkCreate:
			counts.LinkCreate++
		case domain.OpKindLinkDelete:
			counts.LinkDelete++
		case domain.OpKindFileMove:
			counts.FileMove++
		case domain.OpKindFileBackup:
			counts.FileBackup++
		}
	}

	return counts
}
