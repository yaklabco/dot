package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/yaklabco/dot/pkg/dot"
)

// newDoctorIgnoreCommand creates the `doctor ignore` subcommand.
func newDoctorIgnoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ignore [PATH]",
		Short: "Add a symlink path or glob pattern to the doctor ignore list",
		Long: `Add an entry to the doctor ignore list without interactive triage.

Ignored links and patterns are excluded from orphan detection in future
doctor runs. Pass a target-relative symlink path, or --pattern with a glob
matching target-relative paths.

Examples:
  # Ignore a single foreign symlink
  dot doctor ignore .nix-profile --reason "nix managed"

  # Ignore everything under a directory
  dot doctor ignore --pattern "Code/*"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern, _ := cmd.Flags().GetString("pattern")
			reason, _ := cmd.Flags().GetString("reason")

			if err := validatePathPatternExclusive(args, pattern); err != nil {
				return err
			}

			client, err := newDoctorIgnoreClient(cmd)
			if err != nil {
				return err
			}

			if pattern != "" {
				if err := client.DoctorIgnorePattern(cmd.Context(), pattern); err != nil {
					return formatError(err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Added ignore pattern: %s\n", pattern)
				return nil
			}

			if err := client.DoctorIgnoreLink(cmd.Context(), args[0], reason); err != nil {
				return formatError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Ignored link: %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("pattern", "", "Glob pattern to ignore (mutually exclusive with PATH)")
	cmd.Flags().String("reason", "", "Reason for ignoring (recorded in the manifest)")
	return cmd
}

// newDoctorUnignoreCommand creates the `doctor unignore` subcommand.
func newDoctorUnignoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unignore [PATH]",
		Short: "Remove a symlink path or glob pattern from the doctor ignore list",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern, _ := cmd.Flags().GetString("pattern")

			if err := validatePathPatternExclusive(args, pattern); err != nil {
				return err
			}

			client, err := newDoctorIgnoreClient(cmd)
			if err != nil {
				return err
			}

			if pattern != "" {
				if err := client.DoctorUnignorePattern(cmd.Context(), pattern); err != nil {
					return formatError(err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Removed ignore pattern: %s\n", pattern)
				return nil
			}

			if err := client.DoctorUnignoreLink(cmd.Context(), args[0]); err != nil {
				return formatError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Unignored link: %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("pattern", "", "Glob pattern to remove (mutually exclusive with PATH)")
	return cmd
}

// newDoctorIgnoresCommand creates the `doctor ignores` subcommand.
func newDoctorIgnoresCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ignores",
		Short: "List the doctor ignore list",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newDoctorIgnoreClient(cmd)
			if err != nil {
				return err
			}

			links, patterns, err := client.DoctorListIgnored(cmd.Context())
			if err != nil {
				return formatError(err)
			}

			out := cmd.OutOrStdout()
			if len(links) == 0 && len(patterns) == 0 {
				fmt.Fprintln(out, "No ignored links or patterns")
				return nil
			}

			if len(patterns) > 0 {
				fmt.Fprintln(out, "Ignored patterns:")
				sorted := append([]string(nil), patterns...)
				sort.Strings(sorted)
				for _, p := range sorted {
					fmt.Fprintf(out, "  %s\n", p)
				}
			}

			if len(links) > 0 {
				fmt.Fprintln(out, "Ignored links:")
				paths := make([]string, 0, len(links))
				for path := range links {
					paths = append(paths, path)
				}
				sort.Strings(paths)
				for _, path := range paths {
					link := links[path]
					if link.Reason != "" {
						fmt.Fprintf(out, "  %s -> %s (%s)\n", path, link.Target, link.Reason)
					} else {
						fmt.Fprintf(out, "  %s -> %s\n", path, link.Target)
					}
				}
			}
			return nil
		},
	}
}

// validatePathPatternExclusive enforces that exactly one of a positional path
// or a --pattern flag was provided.
func validatePathPatternExclusive(args []string, pattern string) error {
	hasPath := len(args) == 1
	hasPattern := pattern != ""
	if hasPath == hasPattern {
		return fmt.Errorf("provide exactly one of a PATH argument or --pattern")
	}
	return nil
}

// newDoctorIgnoreClient builds a dot client from the command's configuration.
func newDoctorIgnoreClient(cmd *cobra.Command) (*dot.Client, error) {
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return nil, err
	}
	client, err := dot.NewClient(cfg)
	if err != nil {
		return nil, formatError(err)
	}
	return client, nil
}
