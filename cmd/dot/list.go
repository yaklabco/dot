package main

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/cli/render"
	"github.com/jamesainslie/dot/internal/cli/renderer"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newListCommand creates the list command with configuration from global flags.
func newListCommand() *cobra.Command {
	cmd := NewListCommand(&dot.Config{})

	// Override RunE to build config from global flags
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfigWithCmd(cmd)
		if err != nil {
			return err
		}

		// Load extended config for table_style
		configPath := getConfigFilePath()
		extCfg, _ := loadConfigWithRepoPriority(configPath)

		// Get flags
		format, _ := cmd.Flags().GetString("format")
		color, _ := cmd.Flags().GetString("color")
		sortBy, _ := cmd.Flags().GetString("sort")
		showTarget, _ := cmd.Flags().GetBool("show-target")

		// Create client
		client, err := dot.NewClient(cfg)
		if err != nil {
			return formatError(err)
		}

		// Get list of packages
		packages, err := client.List(cmd.Context())
		if err != nil {
			return formatError(err)
		}

		// Sort packages
		sortPackages(packages, sortBy)

		// Create status from packages
		status := dot.Status{
			Packages: packages,
		}

		// Determine colorization
		colorize := shouldColorize(color)

		// Use clean text format by default, structured formats for others
		if format == "text" {
			renderCleanList(cmd.OutOrStdout(), packages, cfg.PackageDir, cfg.TargetDir, showTarget)
		} else {
			// Print context header for table formats
			if format == "table" {
				fmt.Fprintf(cmd.OutOrStdout(), "Package directory: %s\n", cfg.PackageDir)
				fmt.Fprintf(cmd.OutOrStdout(), "Target directory:  %s\n", cfg.TargetDir)
				if cfg.ManifestDir != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Manifest:          %s\n", cfg.ManifestDir)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Manifest:          %s/.dot-manifest.json\n", cfg.TargetDir)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}

			// Create renderer with table_style from config
			tableStyle := ""
			if extCfg != nil {
				tableStyle = extCfg.Output.TableStyle
			}
			r, err := renderer.NewRenderer(format, colorize, tableStyle)
			if err != nil {
				return fmt.Errorf("invalid format: %w", err)
			}

			// Render list
			if err := r.RenderStatus(cmd.OutOrStdout(), status); err != nil {
				return fmt.Errorf("render failed: %w", err)
			}

			// Add newline after output for better terminal spacing
			if format == "table" {
				fmt.Fprintln(cmd.OutOrStdout())
			}
		}

		return nil
	}

	return cmd
}

// NewListCommand creates the list command.
func NewListCommand(cfg *dot.Config) *cobra.Command {
	var format string
	var color string
	var sortBy string
	var showTarget bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all installed packages",
		Long: `Display information about all installed packages.

Shows package name, link count, and installation timestamp for all
packages currently managed by dot. The list can be sorted by various
fields and displayed in multiple output formats.`,
		Example: `  # List all packages
  dot list

  # List packages sorted by link count
  dot list --sort=links

  # Show target directory
  dot list --show-target

  # List packages in JSON format
  dot list --format=json

  # List packages without colors
  dot list --color=never`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Placeholder - will be overridden by newListCommand
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json, yaml, table)")
	cmd.Flags().StringVar(&color, "color", "auto", "Colorize output (auto, always, never)")
	cmd.Flags().StringVar(&sortBy, "sort", "name", "Sort by field (name, links, date)")
	cmd.Flags().BoolVar(&showTarget, "show-target", false, "Show target directory in output")

	return cmd
}

// renderCleanList renders a clean, minimalist package list with subtle colorization.
func renderCleanList(w io.Writer, packages []dot.PackageInfo, packageDir, targetDir string, showTarget bool) {
	if len(packages) == 0 {
		fmt.Fprintf(w, "No packages installed\n")
		return
	}

	// Determine if we should use colors
	colorize := shouldUseColor()
	colorizer := render.NewColorizer(colorize)

	// Header
	pluralS := ""
	if len(packages) != 1 {
		pluralS = "s"
	}
	fmt.Fprintf(w, "Packages: %d package%s in %s\n", len(packages), pluralS, packageDir)

	// Show target directory if requested
	if showTarget {
		fmt.Fprintf(w, "Target:   %s\n", targetDir)
	}

	fmt.Fprintln(w)

	// Calculate column widths for alignment
	maxNameWidth := 0
	maxLinkTextWidth := 0
	for _, pkg := range packages {
		if len(pkg.Name) > maxNameWidth {
			maxNameWidth = len(pkg.Name)
		}
		linkText := fmt.Sprintf("(%d link", pkg.LinkCount)
		if pkg.LinkCount != 1 {
			linkText += "s"
		}
		linkText += ")"
		if len(linkText) > maxLinkTextWidth {
			maxLinkTextWidth = len(linkText)
		}
	}

	// List packages with aligned columns and subtle colorization
	for _, pkg := range packages {
		linkText := fmt.Sprintf("(%d link", pkg.LinkCount)
		if pkg.LinkCount != 1 {
			linkText += "s"
		}
		linkText += ")"

		timeAgo := formatTimeAgo(pkg.InstalledAt)

		// Package name in accent color (dark blue/purple)
		fmt.Fprintf(w, "%s  ",
			colorizer.Accent(fmt.Sprintf("%-*s", maxNameWidth, pkg.Name)))

		// Link count in dim color
		fmt.Fprintf(w, "%s  ",
			colorizer.Dim(fmt.Sprintf("%-*s", maxLinkTextWidth, linkText)))

		// Time in dim color
		fmt.Fprintf(w, "%s\n",
			colorizer.Dim("installed "+timeAgo))
	}
}

// formatTimeAgo formats a time as a human-readable "time ago" string.
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// sortPackages sorts packages by the specified field.
func sortPackages(packages []dot.PackageInfo, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})
	case "links":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].LinkCount > packages[j].LinkCount // Descending
		})
	case "date":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].InstalledAt.After(packages[j].InstalledAt) // Most recent first
		})
	default:
		// Default to name sorting
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})
	}
}
