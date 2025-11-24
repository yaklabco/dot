package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/yaklabco/dot/internal/cli/pretty"
	"github.com/yaklabco/dot/internal/cli/render"
	"github.com/yaklabco/dot/internal/cli/renderer"
	"github.com/yaklabco/dot/pkg/dot"
)

// doctorFlags holds parsed flags.
type doctorFlags struct {
	format, color, scanMode, mode string
	maxDepth                      int
	triage, autoIgnore, verbose   bool
}

// parseDoctorFlags extracts flags from command.
func parseDoctorFlags(cmd *cobra.Command) doctorFlags {
	format, _ := cmd.Flags().GetString("format")
	color, _ := cmd.Flags().GetString("color")
	scanMode, _ := cmd.Flags().GetString("scan-mode")
	maxDepth, _ := cmd.Flags().GetInt("max-depth")
	triage, _ := cmd.Flags().GetBool("triage")
	autoIgnore, _ := cmd.Flags().GetBool("auto-ignore")
	mode, _ := cmd.Flags().GetString("mode")
	verbose, _ := cmd.Flags().GetBool("verbose")
	return doctorFlags{format, color, scanMode, mode, maxDepth, triage, autoIgnore, verbose}
}

// buildScanConfig creates scan configuration from flags.
func buildScanConfig(scanMode string, maxDepth int) (dot.ScanConfig, error) {
	switch scanMode {
	case "off":
		return dot.ScanConfig{
			Mode:         dot.ScanOff,
			MaxDepth:     10,
			ScopeToDirs:  nil,
			SkipPatterns: []string{".git", "node_modules", ".cache", ".npm", ".cargo", ".rustup"},
		}, nil
	case "scoped", "":
		return dot.ScopedScanConfig(), nil
	case "deep":
		return dot.DeepScanConfig(maxDepth), nil
	default:
		return dot.ScanConfig{}, fmt.Errorf("invalid scan-mode: %s (must be off, scoped, or deep)", scanMode)
	}
}

// parseDoctorMode converts mode string to DiagnosticMode.
func parseDoctorMode(mode string) (dot.DiagnosticMode, error) {
	switch mode {
	case "fast", "":
		return dot.DiagnosticFast, nil
	case "deep":
		return dot.DiagnosticDeep, nil
	default:
		return dot.DiagnosticFast, fmt.Errorf("invalid mode: %s (must be fast or deep)", mode)
	}
}

// renderDoctorOutput renders the report.
func renderDoctorOutput(cmd *cobra.Command, report dot.DiagnosticReport, flags doctorFlags, extCfg *dot.ExtendedConfig) error {
	colorize := shouldColorize(flags.color)
	tableStyle := ""
	if extCfg != nil {
		tableStyle = extCfg.Output.TableStyle
	}

	switch flags.format {
	case "json":
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	case "yaml":
		return yaml.NewEncoder(cmd.OutOrStdout()).Encode(report)
	case "text", "table":
		// For verbose text output, render more details
		// Since we don't have access to internal CheckResults here,
		// we'll enhance the succinct rendering when verbose is enabled
		var buf bytes.Buffer
		if flags.verbose {
			renderVerboseDiagnostics(&buf, report, colorize)
		} else {
			renderSuccinctDiagnostics(&buf, report, colorize, tableStyle)
		}
		pager := pretty.NewPager(pretty.PagerConfig{PageSize: 0, Output: cmd.OutOrStdout()})
		return pager.PageLines(strings.Split(buf.String(), "\n"))
	default:
		r, err := renderer.NewRenderer(flags.format, colorize, tableStyle)
		if err != nil {
			return fmt.Errorf("invalid format: %w", err)
		}
		return r.RenderDiagnostics(cmd.OutOrStdout(), report)
	}
}

// checkHealthStatus returns error if health check failed.
func checkHealthStatus(report dot.DiagnosticReport) error {
	if report.OverallHealth == dot.HealthErrors {
		return fmt.Errorf("health check detected errors")
	} else if report.OverallHealth == dot.HealthWarnings {
		return fmt.Errorf("health check detected warnings")
	}
	return nil
}

// newDoctorCommand creates the doctor command with configuration from global flags.
func newDoctorCommand() *cobra.Command {
	cmd := NewDoctorCommand(&dot.Config{})

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfigWithCmd(cmd)
		if err != nil {
			return err
		}

		flags := parseDoctorFlags(cmd)
		client, err := dot.NewClient(cfg)
		if err != nil {
			return formatError(err)
		}

		scanCfg, err := buildScanConfig(flags.scanMode, flags.maxDepth)
		if err != nil {
			return err
		}

		if flags.triage {
			return runTriage(cmd, client, scanCfg, flags.autoIgnore)
		}

		doctorMode, err := parseDoctorMode(flags.mode)
		if err != nil {
			return err
		}

		report, err := client.DoctorWithMode(cmd.Context(), doctorMode, scanCfg)
		if err != nil {
			return formatError(err)
		}

		configPath := getConfigFilePath()
		extCfg, _ := loadConfigWithRepoPriority(cliFlags.packageDir, configPath)

		if err := renderDoctorOutput(cmd, report, flags, extCfg); err != nil {
			return err
		}

		return checkHealthStatus(report)
	}

	return cmd
}

// renderVerboseDiagnostics outputs detailed diagnostics with all issue information.
func renderVerboseDiagnostics(w io.Writer, report dot.DiagnosticReport, colorize bool) {
	c := render.NewColorizer(colorize)

	// Health status header
	healthIcon, healthText, healthColor := getHealthDisplay(report.OverallHealth, c)
	fmt.Fprintf(w, "%s %s\n", healthIcon, healthColor(healthText))
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 60))

	// Detailed statistics
	if report.Statistics.TotalLinks > 0 {
		fmt.Fprintf(w, "Statistics:\n")
		fmt.Fprintf(w, "  Total links: %d\n", report.Statistics.TotalLinks)
		fmt.Fprintf(w, "  Managed links: %d\n", report.Statistics.ManagedLinks)
		fmt.Fprintf(w, "  Broken links: %d\n", report.Statistics.BrokenLinks)
		fmt.Fprintf(w, "  Orphaned links: %d\n", report.Statistics.OrphanedLinks)
		fmt.Fprintf(w, "\n")
	}

	// Issues grouped by severity
	errors := filterIssuesBySeverity(report.Issues, dot.SeverityError)
	warnings := filterIssuesBySeverity(report.Issues, dot.SeverityWarning)
	infos := filterIssuesBySeverity(report.Issues, dot.SeverityInfo)

	if len(errors) > 0 {
		fmt.Fprintf(w, "%s Errors (%d):\n", c.Error("✗"), len(errors))
		renderDetailedIssueList(w, errors, c)
		fmt.Fprintf(w, "\n")
	}

	if len(warnings) > 0 {
		fmt.Fprintf(w, "%s Warnings (%d):\n", c.Warning("⚠"), len(warnings))
		renderDetailedIssueList(w, warnings, c)
		fmt.Fprintf(w, "\n")
	}

	if len(infos) > 0 {
		fmt.Fprintf(w, "%s Info (%d):\n", c.Info("ℹ"), len(infos))
		renderDetailedIssueList(w, infos, c)
		fmt.Fprintf(w, "\n")
	}

	// Clean summary if no issues
	if len(report.Issues) == 0 {
		fmt.Fprintf(w, "%s\n", c.Success("No issues found"))
	}

	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 60))
	fmt.Fprintf(w, "Summary: %d total issues\n", len(report.Issues))
}

// renderDetailedIssueList renders issues with full details including suggestions.
func renderDetailedIssueList(w io.Writer, issues []dot.Issue, c *render.Colorizer) {
	for _, issue := range issues {
		severityIcon := getSeverityIcon(issue.Severity, c)
		fmt.Fprintf(w, "\n  %s [%s] %s\n", severityIcon, issue.Type, issue.Message)
		if issue.Path != "" {
			fmt.Fprintf(w, "     Path: %s\n", c.Dim(issue.Path))
		}
		if issue.Suggestion != "" {
			fmt.Fprintf(w, "     Suggestion: %s\n", c.Info(issue.Suggestion))
		}
	}
}

// getSeverityIcon returns the icon for an issue severity.
func getSeverityIcon(severity dot.IssueSeverity, c *render.Colorizer) string {
	switch severity {
	case dot.SeverityError:
		return c.Error("✗")
	case dot.SeverityWarning:
		return c.Warning("!")
	case dot.SeverityInfo:
		return c.Info("ℹ")
	default:
		return "?"
	}
}

// renderSuccinctDiagnostics outputs diagnostics in a succinct, colorized format.
func renderSuccinctDiagnostics(w io.Writer, report dot.DiagnosticReport, colorize bool, tableStyle string) {
	c := render.NewColorizer(colorize)

	// Health status header
	healthIcon, healthText, healthColor := getHealthDisplay(report.OverallHealth, c)
	fmt.Fprintf(w, "%s %s\n",
		healthIcon,
		healthColor(healthText),
	)

	// Statistics summary if available
	if report.Statistics.TotalLinks > 0 {
		fmt.Fprintf(w, "  %s %s\n",
			c.Dim("•"),
			c.Dim(fmt.Sprintf("%d total links (%d managed, %d broken, %d orphaned)",
				report.Statistics.TotalLinks,
				report.Statistics.ManagedLinks,
				report.Statistics.BrokenLinks,
				report.Statistics.OrphanedLinks)),
		)
	}

	// Issues grouped by severity
	errors := filterIssuesBySeverity(report.Issues, dot.SeverityError)
	warnings := filterIssuesBySeverity(report.Issues, dot.SeverityWarning)
	infos := filterIssuesBySeverity(report.Issues, dot.SeverityInfo)

	if len(errors) > 0 {
		fmt.Fprintf(w, "\n%s %s\n", c.Error("✗"), c.Error(fmt.Sprintf("%d errors:", len(errors))))
		renderIssueList(w, errors, c)
	}

	if len(warnings) > 0 {
		fmt.Fprintf(w, "\n%s %s\n", c.Warning("⚠"), c.Warning(fmt.Sprintf("%d warnings:", len(warnings))))
		renderIssueList(w, warnings, c)
	}

	if len(infos) > 0 {
		fmt.Fprintf(w, "\n%s %s\n", c.Info("ℹ"), c.Info(fmt.Sprintf("%d info:", len(infos))))
		renderIssueList(w, infos, c)
	}

	// Clean summary if no issues
	if len(report.Issues) == 0 {
		fmt.Fprintf(w, "  %s\n", c.Success("No issues found"))
	}
}

// getHealthDisplay returns icon, text, and color for health status
func getHealthDisplay(health dot.HealthStatus, c *render.Colorizer) (string, string, func(string) string) {
	switch health {
	case dot.HealthOK:
		return c.Success("✓"), "Healthy", c.Success
	case dot.HealthWarnings:
		return c.Warning("⚠"), "Warnings detected", c.Warning
	case dot.HealthErrors:
		return c.Error("✗"), "Errors detected", c.Error
	default:
		return c.Dim("?"), "Unknown status", c.Dim
	}
}

// filterIssuesBySeverity returns issues matching the given severity
func filterIssuesBySeverity(issues []dot.Issue, severity dot.IssueSeverity) []dot.Issue {
	var filtered []dot.Issue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// renderIssueList renders a list of issues succinctly
func renderIssueList(w io.Writer, issues []dot.Issue, c *render.Colorizer) {
	for _, issue := range issues {
		fmt.Fprintf(w, "  %s %s",
			c.Dim("•"),
			c.Bold(issue.Path),
		)
		if issue.Message != "" {
			fmt.Fprintf(w, " %s %s",
				c.Dim("—"),
				c.Dim(issue.Message),
			)
		}
		if issue.Suggestion != "" {
			fmt.Fprintf(w, "\n    %s %s",
				c.Dim("└─"),
				c.Accent("Fix: "+issue.Suggestion),
			)
		}
		fmt.Fprintln(w)
	}
}

// runTriage executes interactive triage mode.
func runTriage(cmd *cobra.Command, client *dot.Client, scanCfg dot.ScanConfig, autoIgnore bool) error {
	triageOpts := dot.TriageOptions{
		AutoIgnoreHighConfidence: autoIgnore,
	}

	result, err := client.Triage(cmd.Context(), scanCfg, triageOpts)
	if err != nil {
		return formatError(err)
	}

	// Display results
	renderTriageResults(cmd.OutOrStdout(), result)

	return nil
}

// renderTriageResults displays the triage operation results.
func renderTriageResults(w io.Writer, result dot.TriageResult) {
	colorize := shouldUseColor()
	c := render.NewColorizer(colorize)

	fmt.Fprintln(w)
	fmt.Fprintln(w, c.Success("Triage Complete"))
	fmt.Fprintln(w)

	if len(result.Ignored) > 0 {
		fmt.Fprintf(w, "%s %d links ignored:\n", c.Info("•"), len(result.Ignored))
		for _, link := range result.Ignored {
			fmt.Fprintf(w, "  %s\n", c.Dim(link))
		}
		fmt.Fprintln(w)
	}

	if len(result.Patterns) > 0 {
		fmt.Fprintf(w, "%s %d patterns added:\n", c.Info("•"), len(result.Patterns))
		for _, pattern := range result.Patterns {
			fmt.Fprintf(w, "  %s\n", c.Dim(pattern))
		}
		fmt.Fprintln(w)
	}

	if len(result.Adopted) > 0 {
		fmt.Fprintf(w, "%s %d links marked for adoption:\n", c.Info("•"), len(result.Adopted))
		for link, pkg := range result.Adopted {
			fmt.Fprintf(w, "  %s %s %s\n", c.Dim(link), c.Dim("→"), c.Bold(pkg))
		}
		fmt.Fprintln(w)
	}

	if len(result.Skipped) > 0 {
		fmt.Fprintf(w, "%s %d links skipped\n", c.Dim("•"), len(result.Skipped))
		fmt.Fprintln(w)
	}

	if len(result.Errors) > 0 {
		fmt.Fprintf(w, "%s %d errors:\n", c.Error("✗"), len(result.Errors))
		for link, err := range result.Errors {
			fmt.Fprintf(w, "  %s %s %s\n", c.Bold(link), c.Dim("—"), c.Error(err.Error()))
		}
		fmt.Fprintln(w)
	}

	totalProcessed := len(result.Ignored) + len(result.Adopted) + len(result.Skipped)
	if totalProcessed == 0 && len(result.Patterns) == 0 && len(result.Errors) == 0 {
		fmt.Fprintln(w, c.Dim("No orphaned links found to triage"))
	}
}

// NewDoctorCommand creates the doctor command.
func NewDoctorCommand(cfg *dot.Config) *cobra.Command {
	var format string
	var color string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Perform health checks on the installation",
		Long: `Run comprehensive health checks on the dot installation.

Checks for:
  - Broken symlinks in managed packages (links pointing to non-existent targets)
  - Orphaned symlinks not in manifest (unmanaged links in target directory)
  - Broken unmanaged symlinks (orphaned links with non-existent targets)
  - Permission issues
  - Manifest inconsistencies

Orphan Detection:
  By default, doctor uses scoped scanning to find unmanaged symlinks in
  directories containing managed links. This efficiently detects leftover
  symlinks from previously managed packages.

  Use --scan-mode=off to disable orphan detection for faster checks.
  Use --scan-mode=deep for thorough scanning of entire target directory.

Triage Mode:
  Use --triage to interactively process orphaned symlinks. Triage mode groups
  orphaned links by category and allows you to ignore, adopt, or handle them
  individually. This is useful for cleaning up after uninstalling packages or
  managing symlinks created by other tools.

Exit codes:
  0 - Healthy (no issues found)
  1 - Warnings detected (e.g., orphaned links)
  2 - Errors detected (e.g., broken links)`,
		Example: `  # Run health check with default scoped scanning
  dot doctor

  # Run health check without orphan detection (faster)
  dot doctor --scan-mode=off

  # Run thorough scan of entire home directory
  dot doctor --scan-mode=deep

  # Interactive triage mode for orphaned symlinks
  dot doctor --triage

  # Run health check with JSON output
  dot doctor --format=json

  # Run health check without colors
  dot doctor --color=never`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Placeholder - will be overridden by newDoctorCommand
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json, yaml, table)")
	cmd.Flags().StringVar(&color, "color", "auto", "Colorize output (auto, always, never)")
	cmd.Flags().String("scan-mode", "scoped", "Orphan detection mode (off, scoped, deep)")
	cmd.Flags().Int("max-depth", 10, "Maximum recursion depth for deep scan")
	cmd.Flags().Bool("triage", false, "Interactive triage mode for orphaned symlinks")
	cmd.Flags().Bool("auto-ignore", false, "Automatically ignore high-confidence categories in triage mode")
	cmd.Flags().String("mode", "fast", "Diagnostic mode (fast, deep)")
	cmd.Flags().Bool("verbose", false, "Show detailed diagnostic output")

	return cmd
}
