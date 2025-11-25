package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaklabco/dot/internal/cli/markdown"
	"github.com/yaklabco/dot/internal/cli/output"
	"github.com/yaklabco/dot/internal/cli/pager"
	"github.com/yaklabco/dot/internal/cli/render"
	"github.com/yaklabco/dot/internal/updater/install"
	"github.com/yaklabco/dot/pkg/dot"
)

// newUpgradeCommand creates the upgrade command.
func newUpgradeCommand(version string) *cobra.Command {
	var yes bool
	var checkOnly bool
	var dryRun bool
	var showReleaseNotes bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade dot to the latest version",
		Long: `Upgrade dot to the latest version using the detected package manager.

The upgrade command detects how dot was installed (Homebrew, APT, Pacman,
Chocolatey, go install, or source) and uses the appropriate method to
perform the upgrade.

Configuration (in ~/.config/dot/config.yaml):
  update:
    repository: yaklabco/dot
    include_prerelease: false`,
		Example: `  # Check for and install updates
  dot upgrade

  # Check for updates without installing
  dot upgrade --check-only

  # Skip confirmation prompt
  dot upgrade --yes

  # Show what would be done without executing
  dot upgrade --dry-run

  # Show full release notes with pagination
  dot upgrade --release-notes

  # Debug: see what's happening
  dot upgrade -vvv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(version, yes, checkOnly, dryRun, showReleaseNotes)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&checkOnly, "check-only", false, "Check for updates without installing")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	cmd.Flags().BoolVar(&showReleaseNotes, "release-notes", false, "Show full release notes with pagination")

	return cmd
}

// upgradeContext holds state for the upgrade process.
type upgradeContext struct {
	currentVersion   string
	yes              bool
	checkOnly        bool
	dryRun           bool
	showReleaseNotes bool
	colorizer        *render.Colorizer
	formatter        *output.Formatter
	config           *dot.ExtendedConfig
	orchestrator     *install.UpgradeOrchestrator
	installInfo      *dot.InstallInfo
	latestRelease    *dot.GitHubRelease
	logger           *slog.Logger
}

// runUpgrade handles the upgrade command execution.
func runUpgrade(currentVersion string, yes, checkOnly, dryRun, showReleaseNotes bool) error {
	ctx := &upgradeContext{
		currentVersion:   currentVersion,
		yes:              yes,
		checkOnly:        checkOnly,
		dryRun:           dryRun,
		showReleaseNotes: showReleaseNotes,
	}
	return ctx.run()
}

// run executes the upgrade workflow.
func (ctx *upgradeContext) run() error {
	ctx.initializeUI()
	ctx.initializeLogger()
	ctx.loadConfiguration()

	ctx.logger.Debug("starting upgrade command",
		"version", ctx.currentVersion,
		"check_only", ctx.checkOnly,
		"dry_run", ctx.dryRun,
		"show_release_notes", ctx.showReleaseNotes)

	ctx.orchestrator = dot.NewUpgradeOrchestrator(ctx.currentVersion)

	if err := ctx.detectInstallation(); err != nil {
		return err
	}

	hasUpdate, err := ctx.checkForUpdates()
	if err != nil {
		return err
	}

	if !hasUpdate {
		return nil
	}

	if ctx.checkOnly {
		fmt.Printf("Run %s to upgrade.\n", ctx.colorizer.Accent("dot upgrade"))
		return nil
	}

	return ctx.performUpgrade()
}

// initializeUI sets up the colorizer and formatter.
func (ctx *upgradeContext) initializeUI() {
	colorize := shouldUseColor()
	ctx.colorizer = render.NewColorizer(colorize)
	ctx.formatter = output.NewFormatter(os.Stdout, colorize)
}

// initializeLogger sets up the logger based on verbosity flags.
func (ctx *upgradeContext) initializeLogger() {
	level := verbosityToLevel(cliFlags.verbose)
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	ctx.logger = slog.New(handler)

	// Enable pager debug logging at high verbosity (-vvv)
	if cliFlags.verbose >= 3 {
		pager.SetDebugLogger(ctx.logger)
	}

	ctx.logger.Debug("logger initialized",
		"verbosity", cliFlags.verbose,
		"level", level.String())
}

// loadConfiguration loads the config file with defaults.
func (ctx *upgradeContext) loadConfiguration() {
	ctx.logger.Debug("loading configuration")

	cfg, err := loadConfig()
	if err != nil {
		ctx.logger.Debug("failed to load config, using defaults", "error", err)
		cfg = dot.DefaultExtendedConfig()
	}
	ctx.config = cfg

	ctx.logger.Debug("configuration loaded",
		"repository", cfg.Update.Repository,
		"include_prerelease", cfg.Update.IncludePrerelease)
}

// detectInstallation identifies how dot was installed.
func (ctx *upgradeContext) detectInstallation() error {
	ctx.logger.Debug("detecting installation method")
	fmt.Println("Detecting installation method...")

	installInfo, err := ctx.orchestrator.Detect(context.Background())
	if err != nil {
		ctx.logger.Error("detection failed", "error", err)
		return fmt.Errorf("detect installation: %w", err)
	}
	ctx.installInfo = installInfo

	ctx.logger.Debug("installation detected",
		"source", installInfo.Source.String(),
		"version", installInfo.Version,
		"can_auto_upgrade", installInfo.CanAutoUpgrade,
		"upgrade_instructions", installInfo.UpgradeInstructions)

	fmt.Printf("  Installation method: %s\n", ctx.colorizer.Accent(installInfo.Source.String()))
	if installInfo.Version != "" {
		fmt.Printf("  Detected version:    %s\n", ctx.colorizer.Dim(installInfo.Version))
	}
	fmt.Println()
	return nil
}

// checkForUpdates queries GitHub for new releases.
func (ctx *upgradeContext) checkForUpdates() (bool, error) {
	ctx.logger.Debug("checking for updates",
		"repository", ctx.config.Update.Repository,
		"include_prerelease", ctx.config.Update.IncludePrerelease)

	fmt.Println("Checking for updates...")
	checker := dot.NewVersionChecker(ctx.config.Update.Repository)
	latestRelease, hasUpdate, err := checker.CheckForUpdate(ctx.currentVersion, ctx.config.Update.IncludePrerelease)
	if err != nil {
		ctx.logger.Error("update check failed", "error", err)
		return false, fmt.Errorf("check for updates: %w", err)
	}

	ctx.logger.Debug("update check complete",
		"has_update", hasUpdate,
		"latest_version", latestRelease.TagName,
		"current_version", ctx.currentVersion)

	if !hasUpdate {
		fmt.Printf("\n%s You are already running the latest version (%s)\n",
			ctx.colorizer.Success("[ok]"), ctx.currentVersion)
		return false, nil
	}

	ctx.latestRelease = latestRelease
	ctx.displayUpdateInfo()
	return true, nil
}

// displayUpdateInfo shows update information and release notes.
func (ctx *upgradeContext) displayUpdateInfo() {
	c := ctx.colorizer
	release := ctx.latestRelease

	ctx.logger.Debug("displaying update info",
		"tag", release.TagName,
		"body_length", len(release.Body))

	fmt.Printf("\n%s A new version is available\n\n", c.Info("[info]"))
	fmt.Printf("  Current version:  %s\n", c.Accent(ctx.currentVersion))
	fmt.Printf("  Latest version:   %s\n", c.Accent(release.TagName))

	// Show publish date if available
	if !release.PublishedAt.IsZero() {
		fmt.Printf("  Published:        %s\n", c.Dim(release.PublishedAt.Format("Jan 2, 2006")))
	}

	fmt.Printf("  Release URL:      %s\n", c.Dim(release.HTMLURL))

	if release.Body == "" {
		ctx.logger.Debug("no release body to display")
		fmt.Println()
		return
	}

	// Parse release notes into sections
	ctx.logger.Debug("parsing release notes")
	sections := markdown.ParseReleaseSections(release.Body)

	ctx.logger.Debug("release notes parsed",
		"breaking", len(sections.Breaking),
		"features", len(sections.Features),
		"fixes", len(sections.Fixes),
		"other", len(sections.Other),
		"is_empty", sections.IsEmpty())

	if ctx.showReleaseNotes {
		// Show full release notes with pagination
		ctx.displayFullReleaseNotes(sections)
	} else {
		// Show compact summary
		ctx.displayReleaseSummary(sections)
	}
}

// displayReleaseSummary shows a compact summary of the release notes.
func (ctx *upgradeContext) displayReleaseSummary(sections *markdown.ReleaseSections) {
	c := ctx.colorizer

	ctx.logger.Debug("displaying release summary")

	fmt.Println()

	if sections.IsEmpty() {
		ctx.logger.Debug("sections empty, falling back to simple display")
		// Fallback to simple display if parsing found nothing
		ctx.displaySimpleReleaseNotes()
		return
	}

	// Show summary statistics
	fmt.Print(sections.RenderSummary(c))

	// Show brief details (max 3 per section)
	fmt.Print(sections.RenderDetailed(c, 3))

	// Hint about full release notes
	stats := sections.Stats()
	if stats.TotalItems > 9 {
		fmt.Printf("\n  %s\n", c.Dim("Use --release-notes to view all changes"))
	}

	fmt.Println()
}

// displayFullReleaseNotes shows full release notes with pagination.
func (ctx *upgradeContext) displayFullReleaseNotes(sections *markdown.ReleaseSections) {
	c := ctx.colorizer

	ctx.logger.Debug("displaying full release notes with pagination")

	fmt.Println()
	fmt.Println(c.Bold("Release Notes:"))
	fmt.Println()

	// Build the full content
	var content strings.Builder

	if !sections.IsEmpty() {
		// Use section-based display with high limit
		ctx.logger.Debug("rendering section-based release notes")
		content.WriteString(sections.RenderDetailed(c, 100))
		ctx.logger.Debug("section rendering complete", "content_length", content.Len())
	} else {
		// Fallback to rendered markdown
		ctx.logger.Debug("rendering raw markdown release notes")
		renderer := markdown.NewRenderer(c, 80)
		rendered := renderer.Render(ctx.latestRelease.Body)
		// Indent each line
		for _, line := range strings.Split(rendered, "\n") {
			content.WriteString("  ")
			content.WriteString(line)
			content.WriteString("\n")
		}
		ctx.logger.Debug("markdown rendering complete", "content_length", content.Len())
	}

	ctx.logger.Debug("creating pager")

	// Use pager for long content
	p := pager.New()

	ctx.logger.Debug("pager created, splitting content")

	lines := strings.Split(content.String(), "\n")

	ctx.logger.Debug("pager state",
		"line_count", len(lines),
		"terminal_height", p.Height(),
		"terminal_width", p.Width(),
		"interactive", p.IsInteractive(),
		"needs_paging", p.NeedsPaging(len(lines)))

	if p.IsInteractive() && p.NeedsPaging(len(lines)) {
		fmt.Printf("%s\n", c.Dim("(Press Enter to scroll, q to quit)"))
		fmt.Println()
	}

	ctx.logger.Debug("starting pager display")
	if err := p.DisplayLines(lines); err != nil {
		ctx.logger.Error("pager error, falling back to direct output", "error", err)
		// Fallback to non-paginated display on error
		fmt.Print(content.String())
	}
	ctx.logger.Debug("pager display complete")

	fmt.Println()
}

// displaySimpleReleaseNotes shows a basic line-limited release notes display.
func (ctx *upgradeContext) displaySimpleReleaseNotes() {
	c := ctx.colorizer

	ctx.logger.Debug("displaying simple release notes")

	fmt.Println(c.Bold("Release Notes:"))
	lines := strings.Split(ctx.latestRelease.Body, "\n")
	maxLines := 10

	if len(lines) > maxLines {
		for i := 0; i < maxLines; i++ {
			fmt.Printf("  %s\n", c.Dim(lines[i]))
		}
		fmt.Printf("  %s\n", c.Dim(fmt.Sprintf("... and %d more lines", len(lines)-maxLines)))
		fmt.Printf("  %s\n", c.Dim("Use --release-notes to view all"))
	} else {
		for _, line := range lines {
			fmt.Printf("  %s\n", c.Dim(line))
		}
	}
}

// performUpgrade executes the upgrade process.
func (ctx *upgradeContext) performUpgrade() error {
	c := ctx.colorizer

	ctx.logger.Debug("performing upgrade",
		"can_auto_upgrade", ctx.installInfo.CanAutoUpgrade)

	if !ctx.installInfo.CanAutoUpgrade {
		displayUpgradeInstructions(ctx.installInfo, ctx.latestRelease.HTMLURL)
		return nil
	}

	fmt.Printf("Upgrade method: %s\n", c.Dim(ctx.installInfo.UpgradeInstructions))
	fmt.Println()

	if ctx.dryRun {
		fmt.Printf("%s Dry run - no changes made\n", c.Info("[dry-run]"))
		fmt.Printf("  Would execute: %s\n", c.Accent(ctx.installInfo.UpgradeInstructions))
		return nil
	}

	if !ctx.yes && !confirmUpgrade() {
		fmt.Println("Upgrade cancelled.")
		return nil
	}

	return ctx.executeUpgrade()
}

// executeUpgrade runs the upgrade and displays results.
func (ctx *upgradeContext) executeUpgrade() error {
	c := ctx.colorizer
	fmt.Printf("\n%s Upgrading...\n\n", c.Info("->"))

	ctx.logger.Debug("executing upgrade",
		"expected_version", ctx.latestRelease.TagName)

	opts := dot.DefaultUpgradeOptions()
	opts.DryRun = ctx.dryRun
	opts.ExpectedVersion = ctx.latestRelease.TagName

	result, err := ctx.orchestrator.Upgrade(context.Background(), opts)
	if err != nil {
		ctx.logger.Error("upgrade execution failed", "error", err)
		return fmt.Errorf("upgrade failed: %w", err)
	}

	ctx.logger.Debug("upgrade completed",
		"success", result.Success,
		"new_version", result.NewVersion)

	return ctx.displayUpgradeResult(result, opts.ExpectedVersion)
}

// displayUpgradeResult shows the outcome of the upgrade.
func (ctx *upgradeContext) displayUpgradeResult(result *dot.UpgradeResult, expectedVersion string) error {
	c := ctx.colorizer
	fmt.Println()

	if !result.Success {
		if result.Error != nil {
			return fmt.Errorf("upgrade failed: %w", result.Error)
		}
		return fmt.Errorf("upgrade failed")
	}

	switch {
	case result.NewVersion != "" && result.NewVersion == expectedVersion:
		ctx.formatter.SuccessSimple(fmt.Sprintf("Upgrade completed to version %s", result.NewVersion))
	case result.NewVersion != "":
		ctx.formatter.Warning(fmt.Sprintf("Upgrade completed, but version is %s (expected %s)", result.NewVersion, expectedVersion))
	default:
		ctx.formatter.SuccessSimple("Upgrade completed")
	}

	fmt.Fprintf(os.Stdout, "Run %s to verify the new version.\n", c.Accent("dot --version"))
	ctx.formatter.BlankLine()
	return nil
}

// loadConfig loads the configuration from the config file.
func loadConfig() (*dot.ExtendedConfig, error) {
	configPath := getConfigFilePath()
	loader := dot.NewConfigLoader("dot", configPath)
	return loader.LoadWithEnv()
}

// displayUpgradeInstructions shows instructions for installations that cannot auto-upgrade.
func displayUpgradeInstructions(info *dot.InstallInfo, releaseURL string) {
	colorize := shouldUseColor()
	c := render.NewColorizer(colorize)

	fmt.Println(c.Bold("Upgrade Instructions:"))
	fmt.Printf("\n  Installation type: %s\n", c.Accent(info.Source.String()))
	fmt.Printf("  Automatic upgrade is not available for this installation type.\n\n")

	if info.UpgradeInstructions != "" {
		fmt.Printf("  To upgrade:\n")
		fmt.Printf("    %s\n\n", c.Accent(info.UpgradeInstructions))
	} else {
		fmt.Printf("  Visit the release page to download the latest version:\n")
		fmt.Printf("    %s\n\n", c.Accent(releaseURL))
	}
}

// confirmUpgrade prompts the user for upgrade confirmation.
func confirmUpgrade() bool {
	fmt.Printf("Do you want to upgrade now? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
