package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/yaklabco/dot/pkg/dot"
)

// newRemanageCommand creates the remanage command.
func newRemanageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remanage PACKAGE [PACKAGE...]",
		Short: "Reinstall packages with incremental updates",
		Long: `Reinstall one or more packages by removing old symlinks and 
creating new ones.`,
		Args:              argsWithUsage(cobra.MinimumNArgs(1)),
		RunE:              runRemanage,
		ValidArgsFunction: packageCompletion(true), // Complete with installed packages
	}

	return cmd
}

// runRemanage handles the remanage command execution.
func runRemanage(cmd *cobra.Command, args []string) error {
	return executePackageCommand(cmd, args, func(client *dot.Client, ctx context.Context, packages []string) error {
		return client.Remanage(ctx, packages...)
	}, "remanaged")
}
