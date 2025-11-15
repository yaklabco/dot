package renderer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jamesainslie/dot/internal/cli/pretty"
	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/pkg/dot"
)

// TableRenderer renders output as tables.
type TableRenderer struct {
	colorize   bool
	scheme     ColorScheme
	width      int
	tableStyle string // "default" = modern borders, "simple" = legacy plain text
}

// RenderStatus renders installation status as a table.
func (r *TableRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	if len(status.Packages) == 0 {
		fmt.Fprintln(w, "No packages installed")
		return nil
	}

	// Sort packages by name for consistent output
	sort.Slice(status.Packages, func(i, j int) bool {
		return status.Packages[i].Name < status.Packages[j].Name
	})

	// Calculate statistics
	healthyCount := 0
	unhealthyCount := 0
	for _, pkg := range status.Packages {
		if pkg.IsHealthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	// Use legacy simple rendering if configured
	if r.tableStyle == "simple" {
		return r.renderStatusSimple(w, status)
	}

	// Create table with Light style for clean, professional look
	table := pretty.NewTableWriter(pretty.StyleLight, pretty.TableConfig{
		ColorEnabled: r.colorize,
		AutoWrap:     true,
		MaxWidth:     0, // Auto-detect terminal width
	})

	// Set header
	table.SetHeader("Health", "Package", "Links", "Installed")

	// Add rows
	for _, pkg := range status.Packages {
		healthSymbol := "✓"
		if !pkg.IsHealthy {
			healthSymbol = "✗ " + pkg.IssueType
		}
		table.AppendRow(
			healthSymbol,
			pkg.Name,
			fmt.Sprintf("%d", pkg.LinkCount),
			formatDuration(pkg.InstalledAt),
		)
	}

	// Render
	table.Render(w)

	// Print statistics summary
	fmt.Fprintln(w)
	if unhealthyCount == 0 {
		fmt.Fprintf(w, "%d healthy\n", healthyCount)
	} else {
		fmt.Fprintf(w, "%d healthy, %d unhealthy\n", healthyCount, unhealthyCount)
	}

	return nil
}

// renderStatusSimple renders status using legacy plain text format.
func (r *TableRenderer) renderStatusSimple(w io.Writer, status dot.Status) error {
	headers := []string{"Health", "Package", "Links", "Installed"}
	rows := make([][]string, 0, len(status.Packages))

	healthyCount := 0
	unhealthyCount := 0

	for _, pkg := range status.Packages {
		healthSymbol := "✓"
		if !pkg.IsHealthy {
			healthSymbol = "✗ " + pkg.IssueType
			unhealthyCount++
		} else {
			healthyCount++
		}

		row := []string{
			healthSymbol,
			pkg.Name,
			fmt.Sprintf("%d", pkg.LinkCount),
			formatDuration(pkg.InstalledAt),
		}
		rows = append(rows, row)
	}

	if err := r.renderTableSimple(w, headers, rows); err != nil {
		return err
	}

	// Print statistics summary
	fmt.Fprintln(w)
	if unhealthyCount == 0 {
		fmt.Fprintf(w, "  %d healthy\n", healthyCount)
	} else {
		fmt.Fprintf(w, "  %d healthy, %d unhealthy\n", healthyCount, unhealthyCount)
	}

	return nil
}

func (r *TableRenderer) resetColor() string {
	if r.colorize {
		return "\033[0m"
	}
	return ""
}

// renderTableSimple renders a simple table with plain text formatting (legacy style).
func (r *TableRenderer) renderTableSimple(w io.Writer, headers []string, rows [][]string) error {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Render header
	r.renderRowSimple(w, headers, widths, true)
	r.renderSeparatorSimple(w, widths)

	// Render rows
	for _, row := range rows {
		r.renderRowSimple(w, row, widths, false)
	}

	return nil
}

// renderRowSimple renders a single row with simple formatting.
func (r *TableRenderer) renderRowSimple(w io.Writer, cells []string, widths []int, header bool) {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		width := widths[i]
		if header && r.colorize {
			parts[i] = fmt.Sprintf("%s%-*s%s", r.scheme.Info, width, cell, r.resetColor())
		} else {
			parts[i] = fmt.Sprintf("%-*s", width, cell)
		}
	}
	fmt.Fprintf(w, "  %s  \n", strings.Join(parts, "  "))
}

// renderSeparatorSimple renders a separator line with simple formatting.
func (r *TableRenderer) renderSeparatorSimple(w io.Writer, widths []int) {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}
	fmt.Fprintf(w, "  %s  \n", strings.Join(parts, "  "))
}

// RenderDiagnostics renders diagnostic report as a table.
func (r *TableRenderer) RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error {
	// Show overall health
	healthColor := r.scheme.Success
	if report.OverallHealth == dot.HealthWarnings {
		healthColor = r.scheme.Warning
	} else if report.OverallHealth == dot.HealthErrors {
		healthColor = r.scheme.Error
	}

	fmt.Fprintf(w, "%sHealth Status: %s%s\n\n", r.colorText(healthColor), report.OverallHealth.String(), r.resetColor())

	// Show statistics
	fmt.Fprintln(w, "Statistics:")
	fmt.Fprintf(w, "  Total Links: %d\n", report.Statistics.TotalLinks)
	fmt.Fprintf(w, "  Managed Links: %d\n", report.Statistics.ManagedLinks)
	fmt.Fprintf(w, "  Broken Links: %d\n", report.Statistics.BrokenLinks)
	fmt.Fprintf(w, "  Orphaned Links: %d\n\n", report.Statistics.OrphanedLinks)

	// Show issues in a table
	if len(report.Issues) == 0 {
		fmt.Fprintln(w, "No issues found")
		return nil
	}

	// Use legacy simple rendering if configured
	if r.tableStyle == "simple" {
		return r.renderDiagnosticsSimple(w, report.Issues)
	}

	// Create table with Light style
	table := pretty.NewTableWriter(pretty.StyleLight, pretty.TableConfig{
		ColorEnabled: r.colorize,
		AutoWrap:     true,
		MaxWidth:     0, // Auto-detect terminal width
	})

	// Set header
	table.SetHeader("#", "Severity", "Type", "Path", "Message")

	// Add rows
	for i, issue := range report.Issues {
		table.AppendRow(
			fmt.Sprintf("%d", i+1),
			issue.Severity.String(),
			issue.Type.String(),
			issue.Path, // Let TableWriter handle truncation/wrapping
			issue.Message,
		)
	}

	// Render
	table.Render(w)
	return nil
}

// renderDiagnosticsSimple renders diagnostics issues using legacy plain text format.
func (r *TableRenderer) renderDiagnosticsSimple(w io.Writer, issues []dot.Issue) error {
	headers := []string{"#", "Severity", "Type", "Path", "Message"}
	rows := make([][]string, 0, len(issues))

	for i, issue := range issues {
		pathDisplay := issue.Path
		if len(pathDisplay) > 30 {
			pathDisplay = pathDisplay[:27] + "..."
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			issue.Severity.String(),
			issue.Type.String(),
			pathDisplay,
			issue.Message,
		})
	}

	return r.renderTableSimple(w, headers, rows)
}

func (r *TableRenderer) colorText(color string) string {
	if r.colorize && color != "" {
		return color
	}
	return ""
}

// operationDisplay holds display information for an operation.
type operationDisplay struct {
	Action  string
	Type    string
	Details string
}

// formatOperationForTable extracts display information from an operation.
func formatOperationForTable(op domain.Operation) operationDisplay {
	// Normalize: dereference pointers to get value type for switching
	normalized := normalizeOperation(op)

	display := operationDisplay{Action: "Create"}

	switch typed := normalized.(type) {
	case domain.DirCreate:
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case domain.LinkCreate:
		display.Type = "Symlink"
		display.Details = fmt.Sprintf("%s -> %s", typed.Target.String(), typed.Source.String())

	case domain.FileMove:
		display.Action = "Move"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Dest.String())

	case domain.FileBackup:
		display.Action = "Backup"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Backup.String())

	case domain.DirDelete:
		display.Action = "Delete"
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case domain.LinkDelete:
		display.Action = "Delete"
		display.Type = "Symlink"
		display.Details = typed.Target.String()

	default:
		// Handle unknown operation types with clear, informative display
		display.Action = "Unknown"
		display.Type = fmt.Sprintf("%T", op)
		display.Details = op.String()
	}

	return display
}

// RenderPlan renders an execution plan as a table.
func (r *TableRenderer) RenderPlan(w io.Writer, plan domain.Plan) error {
	fmt.Fprintf(w, "%sDry run mode - no changes will be applied%s\n\n", r.colorText(r.scheme.Warning), r.resetColor())

	if len(plan.Operations) == 0 {
		fmt.Fprintln(w, "No operations required")
		return nil
	}

	// Use legacy simple rendering if configured
	if r.tableStyle == "simple" {
		if err := r.renderPlanSimple(w, plan.Operations); err != nil {
			return err
		}
	} else {
		// Create table with Light style
		table := pretty.NewTableWriter(pretty.StyleLight, pretty.TableConfig{
			ColorEnabled: r.colorize,
			AutoWrap:     true,
			MaxWidth:     0, // Auto-detect terminal width
		})

		// Set header
		table.SetHeader("#", "Action", "Type", "Details")

		// Add rows
		for i, op := range plan.Operations {
			display := formatOperationForTable(op)

			table.AppendRow(
				fmt.Sprintf("%d", i+1),
				display.Action,
				display.Type,
				display.Details, // Let TableWriter handle truncation/wrapping
			)
		}

		// Render
		table.Render(w)
	}

	// Summary
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary:")

	// Count all operation kinds in a single pass
	counts := make(map[domain.OperationKind]int)
	for _, op := range plan.Operations {
		counts[op.Kind()]++
	}

	// Display counts with semantic labels for each operation kind
	if count := counts[domain.OpKindDirCreate]; count > 0 {
		fmt.Fprintf(w, "  Directories created: %d\n", count)
	}
	if count := counts[domain.OpKindLinkCreate]; count > 0 {
		fmt.Fprintf(w, "  Symlinks created: %d\n", count)
	}
	if count := counts[domain.OpKindFileMove]; count > 0 {
		fmt.Fprintf(w, "  Files moved: %d\n", count)
	}
	if count := counts[domain.OpKindFileBackup]; count > 0 {
		fmt.Fprintf(w, "  Backups created: %d\n", count)
	}
	if count := counts[domain.OpKindDirDelete]; count > 0 {
		fmt.Fprintf(w, "  Directories deleted: %d\n", count)
	}
	if count := counts[domain.OpKindLinkDelete]; count > 0 {
		fmt.Fprintf(w, "  Symlinks deleted: %d\n", count)
	}

	// Always show conflicts count
	fmt.Fprintf(w, "  Conflicts: %d\n", len(plan.Metadata.Conflicts))

	return nil
}

// renderPlanSimple renders execution plan using legacy plain text format.
func (r *TableRenderer) renderPlanSimple(w io.Writer, operations []domain.Operation) error {
	headers := []string{"#", "Action", "Type", "Details"}
	rows := make([][]string, 0, len(operations))

	for i, op := range operations {
		display := formatOperationForTable(op)

		// Truncate details if too long
		details := display.Details
		if len(details) > 60 {
			details = details[:57] + "..."
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			display.Action,
			display.Type,
			details,
		})
	}

	return r.renderTableSimple(w, headers, rows)
}
