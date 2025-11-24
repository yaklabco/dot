package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yaklabco/dot/internal/cli/output"
	"github.com/yaklabco/dot/internal/cli/renderer"
	"github.com/yaklabco/dot/pkg/dot"
)

// newManageCommand creates the manage command.
func newManageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage PACKAGE [PACKAGE...]",
		Short: "Install packages by creating symlinks",
		Long: `Install one or more packages by creating symlinks from the package 
directory to the target directory.`,
		Args:              argsWithUsage(cobra.MinimumNArgs(1)),
		RunE:              runManage,
		ValidArgsFunction: packageCompletion(false), // Complete with available packages
	}

	return cmd
}

// runManage handles the manage command execution.
func runManage(cmd *cobra.Command, args []string) error {
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	// Load extended config for table_style
	configPath := getConfigFilePath()
	extCfg, _ := loadConfigWithRepoPriority(cliFlags.packageDir, configPath)

	client, err := dot.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	packages := args

	// Check for potential secrets in packages before managing
	if warnings := checkPackagesForSecrets(ctx, client, packages); len(warnings) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "\nWarning: Potential secrets detected:\n")
		for _, w := range warnings {
			fmt.Fprintf(cmd.ErrOrStderr(), "  - %s (%s)\n", w.Path, w.Reason)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "\nThese files are ignored by default. See 'dot help secrets' for details.\n\n")
	}

	// If dry-run mode, render the plan instead of executing
	if cfg.DryRun {
		plan, err := client.PlanManage(ctx, packages...)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			return err
		}

		// Create renderer and render the plan with table_style from config
		tableStyle := ""
		if extCfg != nil {
			tableStyle = extCfg.Output.TableStyle
		}
		rend, err := renderer.NewRenderer("text", true, tableStyle)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			return err
		}

		if err := rend.RenderPlan(os.Stdout, plan); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			return err
		}

		return nil
	}

	// Normal execution
	if err := client.Manage(ctx, packages...); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	// Determine colorization from global flag
	colorize := shouldUseColor()

	// Create formatter and print success message
	formatter := output.NewFormatter(cmd.OutOrStdout(), colorize)
	formatter.Success("managed", len(packages), "package", "packages")
	formatter.BlankLine()

	return nil
}
