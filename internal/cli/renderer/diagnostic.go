package renderer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jamesainslie/dot/internal/domain"
)

// DiagnosticRenderer renders diagnostic check results.
type DiagnosticRenderer struct {
	writer   io.Writer
	verbose  bool
	colorize bool
}

// NewDiagnosticRenderer creates a new diagnostic renderer.
func NewDiagnosticRenderer(w io.Writer, verbose, colorize bool) *DiagnosticRenderer {
	return &DiagnosticRenderer{
		writer:   w,
		verbose:  verbose,
		colorize: colorize,
	}
}

// RenderReport renders a complete diagnostic report.
func (r *DiagnosticRenderer) RenderReport(results []domain.CheckResult, overallStatus domain.CheckStatus) error {
	if r.verbose {
		return r.renderDetailed(results, overallStatus)
	}
	return r.renderSuccinct(results, overallStatus)
}

// renderSuccinct provides a compact summary view.
func (r *DiagnosticRenderer) renderSuccinct(results []domain.CheckResult, overallStatus domain.CheckStatus) error {
	// Overall status
	statusSymbol := r.statusSymbol(overallStatus)
	statusText := r.colorizeStatus(string(overallStatus), overallStatus)
	fmt.Fprintf(r.writer, "\nOverall Status: %s %s\n\n", statusSymbol, statusText)

	// Check summary
	for _, result := range results {
		symbol := r.statusSymbol(result.Status)
		name := result.CheckName
		issueCount := len(result.Issues)

		if issueCount == 0 {
			fmt.Fprintf(r.writer, "%s %s\n", symbol, name)
		} else {
			plural := "issue"
			if issueCount > 1 {
				plural = "issues"
			}
			fmt.Fprintf(r.writer, "%s %s (%d %s)\n", symbol, name, issueCount, plural)
		}
	}

	// Count issues by severity
	errorCount, warningCount := r.countIssuesBySeverity(results)
	if errorCount > 0 || warningCount > 0 {
		fmt.Fprintf(r.writer, "\n")
		if errorCount > 0 {
			fmt.Fprintf(r.writer, "Errors: %d\n", errorCount)
		}
		if warningCount > 0 {
			fmt.Fprintf(r.writer, "Warnings: %d\n", warningCount)
		}
	}

	return nil
}

// renderDetailed provides comprehensive output with all issues and remediation.
func (r *DiagnosticRenderer) renderDetailed(results []domain.CheckResult, overallStatus domain.CheckStatus) error {
	// Overall status
	statusSymbol := r.statusSymbol(overallStatus)
	statusText := r.colorizeStatus(string(overallStatus), overallStatus)
	fmt.Fprintf(r.writer, "\n%s Overall Status: %s\n", statusSymbol, statusText)
	fmt.Fprintf(r.writer, "%s\n\n", strings.Repeat("=", 60))

	// Detailed check results
	for i, result := range results {
		if i > 0 {
			fmt.Fprintf(r.writer, "\n")
		}

		symbol := r.statusSymbol(result.Status)
		fmt.Fprintf(r.writer, "%s Check: %s\n", symbol, result.CheckName)

		// Show statistics if available
		if len(result.Stats) > 0 {
			r.renderStats(result.Stats)
		}

		// Show issues
		if len(result.Issues) > 0 {
			fmt.Fprintf(r.writer, "\n  Issues:\n")
			for _, issue := range result.Issues {
				r.renderIssue(issue)
			}
		}

		fmt.Fprintf(r.writer, "%s\n", strings.Repeat("-", 60))
	}

	// Summary
	errorCount, warningCount := r.countIssuesBySeverity(results)
	fmt.Fprintf(r.writer, "\nSummary:\n")
	fmt.Fprintf(r.writer, "  Checks run: %d\n", len(results))
	fmt.Fprintf(r.writer, "  Errors: %d\n", errorCount)
	fmt.Fprintf(r.writer, "  Warnings: %d\n", warningCount)

	return nil
}

// renderStats renders statistics for a check.
func (r *DiagnosticRenderer) renderStats(stats map[string]any) {
	if len(stats) == 0 {
		return
	}

	fmt.Fprintf(r.writer, "  Statistics:\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := stats[key]
		fmt.Fprintf(r.writer, "    %s: %v\n", r.formatStatKey(key), value)
	}
}

// renderIssue renders a single issue with details.
func (r *DiagnosticRenderer) renderIssue(issue domain.Issue) {
	severitySymbol := r.severitySymbol(issue.Severity)
	fmt.Fprintf(r.writer, "\n  %s [%s] %s\n", severitySymbol, issue.Code, issue.Message)

	if issue.Path != "" {
		fmt.Fprintf(r.writer, "     Path: %s\n", issue.Path)
	}

	if len(issue.Context) > 0 {
		r.renderContext(issue.Context)
	}

	if issue.Remediation != nil {
		fmt.Fprintf(r.writer, "     Remediation: %s\n", issue.Remediation.Description)
	}
}

// renderContext renders issue context information.
func (r *DiagnosticRenderer) renderContext(context map[string]any) {
	if len(context) == 0 {
		return
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(context))
	for k := range context {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := context[key]
		fmt.Fprintf(r.writer, "     %s: %v\n", key, value)
	}
}

// statusSymbol returns a symbol for a check status.
func (r *DiagnosticRenderer) statusSymbol(status domain.CheckStatus) string {
	if !r.colorize {
		switch status {
		case domain.CheckStatusPass:
			return "[OK]"
		case domain.CheckStatusWarning:
			return "[WARN]"
		case domain.CheckStatusFail:
			return "[FAIL]"
		case domain.CheckStatusSkipped:
			return "[SKIP]"
		default:
			return "[????]"
		}
	}

	// With colorization
	switch status {
	case domain.CheckStatusPass:
		return "\033[32m✓\033[0m" // Green checkmark
	case domain.CheckStatusWarning:
		return "\033[33m⚠\033[0m" // Yellow warning
	case domain.CheckStatusFail:
		return "\033[31m✗\033[0m" // Red X
	case domain.CheckStatusSkipped:
		return "\033[90m-\033[0m" // Gray dash
	default:
		return "?"
	}
}

// severitySymbol returns a symbol for an issue severity.
func (r *DiagnosticRenderer) severitySymbol(severity domain.IssueSeverity) string {
	if !r.colorize {
		switch severity {
		case domain.IssueSeverityError:
			return "[ERROR]"
		case domain.IssueSeverityWarning:
			return "[WARN]"
		case domain.IssueSeverityInfo:
			return "[INFO]"
		default:
			return "[?]"
		}
	}

	switch severity {
	case domain.IssueSeverityError:
		return "\033[31m✗\033[0m" // Red X
	case domain.IssueSeverityWarning:
		return "\033[33m!\033[0m" // Yellow exclamation
	case domain.IssueSeverityInfo:
		return "\033[34mℹ\033[0m" // Blue info
	default:
		return "?"
	}
}

// colorizeStatus colorizes status text.
func (r *DiagnosticRenderer) colorizeStatus(text string, status domain.CheckStatus) string {
	if !r.colorize {
		return text
	}

	switch status {
	case domain.CheckStatusPass:
		return fmt.Sprintf("\033[32m%s\033[0m", text)
	case domain.CheckStatusWarning:
		return fmt.Sprintf("\033[33m%s\033[0m", text)
	case domain.CheckStatusFail:
		return fmt.Sprintf("\033[31m%s\033[0m", text)
	case domain.CheckStatusSkipped:
		return fmt.Sprintf("\033[90m%s\033[0m", text)
	default:
		return text
	}
}

// formatStatKey formats a statistic key for display.
func (r *DiagnosticRenderer) formatStatKey(key string) string {
	// Convert snake_case to Title Case
	words := strings.Split(key, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// countIssuesBySeverity counts issues by severity across all results.
func (r *DiagnosticRenderer) countIssuesBySeverity(results []domain.CheckResult) (errors, warnings int) {
	for _, result := range results {
		for _, issue := range result.Issues {
			switch issue.Severity {
			case domain.IssueSeverityError:
				errors++
			case domain.IssueSeverityWarning:
				warnings++
			}
		}
	}
	return errors, warnings
}
