package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jamesainslie/dot/pkg/dot"
)

// newUnmanageCommand creates the unmanage command.
func newUnmanageCommand() *cobra.Command {
	var purge bool
	var noRestore bool
	var cleanup bool
	var all bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "unmanage PACKAGE [PACKAGE...]",
		Short: "Remove packages by deleting symlinks",
		Long: `Remove one or more packages by deleting their symlinks from 
the target directory.

By default, adopted packages (created via 'dot adopt') are restored to 
their original locations. Managed packages only have their symlinks removed.

Cleanup mode removes orphaned packages from the manifest without modifying 
the filesystem - useful when packages no longer exist.

Use --all to remove all managed packages at once. This requires confirmation
unless --yes or --force is specified.`,
		Example: `  # Remove package and restore adopted files
  dot unmanage ssh

  # Remove package and delete package directory
  dot unmanage ssh --purge

  # Remove package without restoring (leave in package dir)
  dot unmanage ssh --no-restore

  # Clean up orphaned manifest entry (no filesystem changes)
  dot unmanage old-package --cleanup

  # Remove all packages (with confirmation)
  dot unmanage --all

  # Remove all packages without confirmation
  dot unmanage --all --yes

  # Preview removing all packages without changes
  dot unmanage --all --dry-run`,
		Args: argsWithUsage(func(cmd *cobra.Command, args []string) error {
			allFlag, _ := cmd.Flags().GetBool("all")
			if allFlag && len(args) > 0 {
				return fmt.Errorf("cannot specify package names with --all flag")
			}
			if !allFlag && len(args) == 0 {
				return fmt.Errorf("requires at least 1 package name or --all flag")
			}
			return nil
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnmanage(cmd, args, purge, noRestore, cleanup, all, yes)
		},
		ValidArgsFunction: packageCompletion(true), // Complete with installed packages
	}

	cmd.Flags().BoolVar(&purge, "purge", false, "Delete package directory instead of restoring files")
	cmd.Flags().BoolVar(&noRestore, "no-restore", false, "Don't restore adopted files (leave in package directory)")
	cmd.Flags().BoolVar(&cleanup, "cleanup", false, "Remove orphaned manifest entries (packages with missing links/directories)")
	cmd.Flags().BoolVar(&all, "all", false, "Remove all managed packages")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&yes, "force", false, "Skip confirmation prompt (alias for --yes)")

	return cmd
}

// runUnmanage handles the unmanage command execution.
func runUnmanage(cmd *cobra.Command, args []string, purge, noRestore, cleanup, all, yes bool) error {
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return err
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Build options
	opts := dot.UnmanageOptions{
		Purge:   purge,
		Restore: !noRestore && !purge, // Default is true unless --no-restore or --purge
		Cleanup: cleanup,
	}

	// Handle --all flag
	if all {
		return runUnmanageAll(cmd, cfg, client, ctx, opts, yes)
	}

	packages := args

	// Execute unmanage with options
	if err := client.UnmanageWithOptions(ctx, opts, packages...); err != nil {
		return err
	}

	if !cfg.DryRun {
		if cleanup {
			if len(packages) > 0 {
				fmt.Printf("Cleaned up %s from manifest\n", formatCount(len(packages), "orphaned package", "orphaned packages"))
			} else {
				fmt.Println("No orphaned packages found in manifest")
			}
		} else if purge {
			fmt.Printf("Unmanaged and purged %s\n", formatCount(len(packages), "package", "packages"))
		} else if opts.Restore {
			fmt.Printf("Unmanaged and restored %s\n", formatCount(len(packages), "package", "packages"))
		} else {
			fmt.Printf("Unmanaged %s\n", formatCount(len(packages), "package", "packages"))
		}
		fmt.Println() // Blank line for terminal spacing
	}

	return nil
}

// runUnmanageAll handles the unmanage --all command execution with confirmation.
func runUnmanageAll(cmd *cobra.Command, cfg dot.Config, client *dot.Client, ctx context.Context, opts dot.UnmanageOptions, skipConfirm bool) error {
	// Get current status to show what will be removed
	status, err := client.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	packageCount := len(status.Packages)
	if packageCount == 0 {
		fmt.Println("No packages to unmanage")
		return nil
	}

	// Show detailed summary in dry-run mode or when confirming
	if cfg.DryRun || !skipConfirm {
		displayUnmanageAllSummary(status.Packages, opts, cfg.PackageDir)
	}

	// Request confirmation unless --yes/--force/--dry-run
	if !skipConfirm && !cfg.DryRun {
		if !isTerminal(cmd) {
			return fmt.Errorf("stdin is not a terminal; use --yes to confirm")
		}
		if !confirmAction(cmd, "Proceed with unmanaging all packages?") {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	// Execute unmanage all (unless dry-run already handled by client)
	count, err := client.UnmanageAll(ctx, opts)
	if err != nil {
		return err
	}

	// Report results
	reportUnmanageAllResults(count, opts, cfg.DryRun)
	return nil
}

// displayUnmanageAllSummary shows what will be unmanaged.
func displayUnmanageAllSummary(packages []dot.PackageInfo, opts dot.UnmanageOptions, packageDir string) {
	fmt.Printf("This will unmanage %s:\n", accent(fmt.Sprintf("%d package(s)", len(packages))))
	for _, pkg := range packages {
		operation := getUnmanageOperation(pkg, opts)
		operationColor := getOperationColor(operation)
		// Format: package (operation, N links: link1, link2, ... | pkg-dir)
		linkList := formatLinkList(pkg.Links)
		pkgPath := filepath.Join(packageDir, pkg.Name)
		fmt.Printf("  %s %s %s %s %s %s\n",
			dim("•"),
			bold(pkg.Name),
			dim("("),
			operationColor(operation),
			dim(fmt.Sprintf("%d links:", len(pkg.Links))),
			dim(linkList+" | "+pkgPath+")"),
		)
	}
	fmt.Println()
}

// getOperationColor returns the appropriate color function for an operation
func getOperationColor(operation string) func(string) string {
	switch operation {
	case "purge":
		return errorText
	case "restore":
		return info
	default:
		return accent
	}
}

// formatLinkList joins links with commas, showing all if few, or truncating if many.
func formatLinkList(links []string) string {
	if len(links) == 0 {
		return "none"
	}
	if len(links) <= 3 {
		return strings.Join(links, ", ")
	}
	return strings.Join(links[:3], ", ") + fmt.Sprintf("... (%d more)", len(links)-3)
}

// getUnmanageOperation determines the operation type for a package.
func getUnmanageOperation(pkg dot.PackageInfo, opts dot.UnmanageOptions) string {
	if opts.Purge {
		return "purge"
	}
	if opts.Restore && pkg.Source == "adopted" {
		return "restore"
	}
	return "remove"
}

// reportUnmanageAllResults displays the final results.
func reportUnmanageAllResults(count int, opts dot.UnmanageOptions, dryRun bool) {
	operation := "unmanage"
	if opts.Purge {
		operation = "unmanage and purge"
	} else if opts.Restore {
		operation = "unmanage and restore"
	}

	packageText := formatCount(count, "package", "packages")

	if dryRun {
		fmt.Printf("%s %s %s\n",
			dim("Would"),
			operation,
			accent(packageText),
		)
	} else {
		fmt.Printf("%s %s\n",
			operation,
			accent(packageText),
		)
	}
}

// isTerminal checks if the command's input stream is a terminal.
func isTerminal(cmd *cobra.Command) bool {
	in := cmd.InOrStdin()
	if f, ok := in.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// confirmAction prompts the user for confirmation using the command's input stream.
func confirmAction(cmd *cobra.Command, prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
