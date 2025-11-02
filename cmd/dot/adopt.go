package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newAdoptCommand creates the adopt command.
func newAdoptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adopt [PACKAGE] FILE [FILE...]",
		Short: "Move existing files into package then link",
		Long: `Move existing files into a package and create symlinks.

Single File (auto-naming):
  dot adopt .ssh              # Creates package "ssh"
  dot adopt .vimrc            # Creates package "vimrc"

Multiple Files (explicit package):
  dot adopt ssh .ssh .ssh/config
  dot adopt vim .vimrc .vim

For shell glob expansion, specify package name:
  dot adopt git .git*         # Package "git" with all .git* files`,
		Args: argsWithUsage(cobra.MinimumNArgs(1)),
		RunE: runAdopt,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// For auto-naming mode, complete with files
			// For explicit mode, first arg is package, rest are files
			if len(args) == 0 {
				// Could be package name or file - suggest both packages and files
				return getAvailablePackages(), cobra.ShellCompDirectiveDefault
			}
			// Subsequent arguments: complete with files
			return nil, cobra.ShellCompDirectiveDefault
		},
	}

	return cmd
}

// runAdopt handles the adopt command execution.
func runAdopt(cmd *cobra.Command, args []string) error {
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return formatError(err)
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		return formatError(err)
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	var pkg string
	var files []string

	if len(args) == 1 {
		// Auto-naming: derive package from single file
		files = []string{args[0]}
		pkg = derivePackageName(args[0])
		if pkg == "" {
			return fmt.Errorf("cannot derive package name from: %s", args[0])
		}
		// Apply dotfile translation to package name
		// ".ssh" → "dot-ssh", "README.md" → "README.md"
		pkg = scanner.UntranslateDotfile(pkg)
	} else {
		// Explicit mode: first arg is package name
		pkg = args[0]
		files = args[1:]
	}

	if err := client.Adopt(ctx, files, pkg); err != nil {
		return formatError(err)
	}

	if !cfg.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "Adopted %s into %s\n", formatCount(len(files), "file", "files"), pkg)
		fmt.Fprintln(cmd.OutOrStdout()) // Blank line for terminal spacing
	}

	return nil
}
