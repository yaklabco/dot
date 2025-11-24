package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yaklabco/dot/internal/cli/adopt"
	"github.com/yaklabco/dot/internal/cli/output"
	"github.com/yaklabco/dot/internal/cli/render"
	"github.com/yaklabco/dot/internal/cli/terminal"
	"github.com/yaklabco/dot/internal/config"
	"github.com/yaklabco/dot/internal/scanner"
	"github.com/yaklabco/dot/pkg/dot"
)

// newAdoptCommand creates the adopt command.
func newAdoptCommand() *cobra.Command {
	var scanDirs []string
	var excludeDirs []string
	var maxSize string

	cmd := &cobra.Command{
		Use:   "adopt [PACKAGE] FILE [FILE...]",
		Short: "Move existing files into package then link",
		Long: `Move existing files into a package and create symlinks.

Interactive Mode (no arguments):
  dot adopt                  # Discover and select dotfiles interactively

Traditional Mode:

Single File (auto-naming):
  dot adopt .ssh              # Creates package "ssh"
  dot adopt .vimrc            # Creates package "vimrc"

Multiple Files (explicit package):
  dot adopt ssh .ssh .ssh/config
  dot adopt vim .vimrc .vim

Path Resolution:
  ./file or ../file  → Resolved from current directory
  file or .config/x  → Resolved from target directory ($HOME)
  /abs or ~/file     → Used as absolute path

Interactive Mode Options:
  --scan-dirs       Additional directories to scan
  --exclude-dirs    Directories to exclude from discovery
  --max-size        Maximum file size (default: 10M)

For shell glob expansion, specify package name:
  dot adopt git .git*         # Package "git" with all .git* files`,
		Args: cobra.ArbitraryArgs, // Accept 0 or more arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdoptCommand(cmd, args, scanDirs, excludeDirs, maxSize)
		},
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

	// Interactive mode options (only used when no arguments provided)
	cmd.Flags().StringSliceVar(&scanDirs, "scan-dirs", nil,
		"additional directories to scan (interactive mode)")
	cmd.Flags().StringSliceVar(&excludeDirs, "exclude-dirs", nil,
		"directories to exclude from discovery (interactive mode)")
	cmd.Flags().StringVar(&maxSize, "max-size", "10M",
		"maximum file size to adopt (interactive mode)")

	return cmd
}

// runAdoptCommand routes to interactive or traditional mode based on arguments.
func runAdoptCommand(cmd *cobra.Command, args []string, scanDirs, excludeDirs []string, maxSizeStr string) error {
	// No arguments → Interactive mode
	if len(args) == 0 {
		return runAdoptInteractive(cmd, scanDirs, excludeDirs, maxSizeStr)
	}

	// Has arguments → Traditional mode
	return runAdoptTraditional(cmd, args)
}

// runAdoptInteractive handles interactive discovery and adoption.
func runAdoptInteractive(cmd *cobra.Command, scanDirs, excludeDirs []string, maxSizeStr string) error {
	// Build config
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return formatError(err)
	}

	// Check if we're in a TTY (interactive terminal)
	if !terminal.IsInteractive() {
		// Silence usage on this error since it's not a usage problem
		cmd.SilenceUsage = true
		fmt.Fprintln(cmd.ErrOrStderr(), "Error: interactive mode requires a terminal (TTY)")
		fmt.Fprintln(cmd.ErrOrStderr(), "")
		fmt.Fprintln(cmd.ErrOrStderr(), "To adopt files in non-interactive mode:")
		fmt.Fprintln(cmd.ErrOrStderr(), "  dot adopt FILE                    # Auto-name package from file")
		fmt.Fprintln(cmd.ErrOrStderr(), "  dot adopt PACKAGE FILE [FILE...]  # Specify package name")
		return fmt.Errorf("not a terminal")
	}

	// Parse max size
	maxSize, err := parseSize(maxSizeStr)
	if err != nil {
		return fmt.Errorf("invalid max-size: %w", err)
	}

	// Default scan directories: $HOME and $HOME/.config
	if len(scanDirs) == 0 {
		scanDirs = adopt.DefaultScanDirs(cfg.TargetDir)
	}

	// Default exclude directories
	if len(excludeDirs) == 0 {
		excludeDirs = adopt.DefaultExcludeDirs()
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Create client
	client, err := dot.NewClient(cfg)
	if err != nil {
		return formatError(err)
	}

	// Build DiscoveryOptions
	opts := adopt.DiscoveryOptions{
		ScanDirs:    scanDirs,
		ExcludeDirs: excludeDirs,
		MaxFileSize: maxSize,
	}

	// Discover dotfiles with progress spinner
	candidates, err := adopt.ScanWithProgress(ctx, cfg.FS, opts, client, cfg.TargetDir)
	if err != nil {
		return formatError(err)
	}

	// Handle no candidates found
	if len(candidates) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No adoptable dotfiles found.")
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Tip: Use flags to customize discovery:")
		fmt.Fprintln(cmd.OutOrStdout(), "  --scan-dirs      Scan additional directories")
		fmt.Fprintln(cmd.OutOrStdout(), "  --exclude-dirs   Exclude specific directories")
		fmt.Fprintln(cmd.OutOrStdout(), "  --max-size       Change maximum file size")
		return nil
	}

	// Run interactive session
	colorize := shouldUseColor()
	configDir := config.GetConfigPath("dot")
	adopter := adopt.NewInteractiveAdopter(
		cmd.InOrStdin(),
		cmd.OutOrStdout(),
		colorize,
		cfg.FS,
		configDir,
	)

	groups, err := adopter.Run(ctx, candidates)
	if err != nil {
		return formatError(err)
	}

	if len(groups) == 0 {
		return nil // User cancelled or nothing selected
	}

	// Execute adoptions
	colorizer := render.NewColorizer(colorize)
	formatter := output.NewFormatter(cmd.OutOrStdout(), colorize)

	totalFiles := 0
	for _, group := range groups {
		// Check for potential secrets before adopting
		displaySecretsWarning(cmd.ErrOrStderr(), group.Files)

		if err := client.Adopt(ctx, group.Files, group.PackageName); err != nil {
			return formatError(fmt.Errorf("adopt %s: %w", group.PackageName, err))
		}

		totalFiles += len(group.Files)
		fmt.Fprintf(cmd.OutOrStdout(), "%s Adopted %d files into %s\n",
			colorizer.Success("✓"),
			len(group.Files),
			colorizer.Accent(group.PackageName))
	}

	formatter.BlankLine()
	fmt.Fprintf(cmd.OutOrStdout(), "Total: %s files into %s packages\n",
		colorizer.Accent(strconv.Itoa(totalFiles)),
		colorizer.Accent(strconv.Itoa(len(groups))))

	return nil
}

// runAdoptTraditional handles the traditional file-based adoption.
func runAdoptTraditional(cmd *cobra.Command, args []string) error {
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

	// Check for potential secrets before adopting
	displaySecretsWarning(cmd.ErrOrStderr(), files)

	if err := client.Adopt(ctx, files, pkg); err != nil {
		return formatError(err)
	}

	if !cfg.DryRun {
		// Determine colorization from global flag
		colorize := shouldUseColor()

		// Create formatter for consistent output
		formatter := output.NewFormatter(cmd.OutOrStdout(), colorize)
		colorizer := render.NewColorizer(colorize)

		// Print success message
		if len(files) == 1 {
			fmt.Fprintf(cmd.OutOrStdout(), "%s Adopted %s into %s\n",
				colorizer.Success("✓"),
				files[0],
				colorizer.Accent(pkg))
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s Adopted %d files into %s\n",
				colorizer.Success("✓"),
				len(files),
				colorizer.Accent(pkg))
		}
		formatter.BlankLine()
	}

	return nil
}

// parseSize parses human-readable size strings into bytes.
// Supports formats like "10M", "1.5G", "500K", "100B"
func parseSize(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}

	s = strings.TrimSpace(strings.ToUpper(s))

	// Parse multiplier
	multiplier := int64(1)
	if len(s) > 0 {
		switch s[len(s)-1] {
		case 'K':
			multiplier = 1024
			s = s[:len(s)-1]
		case 'M':
			multiplier = 1024 * 1024
			s = s[:len(s)-1]
		case 'G':
			multiplier = 1024 * 1024 * 1024
			s = s[:len(s)-1]
		case 'B':
			// Optional 'B' suffix
			if len(s) > 1 {
				switch s[len(s)-2] {
				case 'K':
					multiplier = 1024
					s = s[:len(s)-2]
				case 'M':
					multiplier = 1024 * 1024
					s = s[:len(s)-2]
				case 'G':
					multiplier = 1024 * 1024 * 1024
					s = s[:len(s)-2]
				default:
					s = s[:len(s)-1]
				}
			} else {
				s = s[:len(s)-1]
			}
		}
	}

	// Parse number
	var value float64
	_, err := fmt.Sscanf(s, "%f", &value)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", s)
	}

	return int64(value * float64(multiplier)), nil
}

// displaySecretsWarning checks files for potential secrets and displays warnings.
func displaySecretsWarning(w interface{ Write([]byte) (int, error) }, files []string) {
	warnings := checkFilesForSecrets(files)
	if len(warnings) == 0 {
		return
	}

	fmt.Fprintf(w, "\nWarning: Files may contain secrets:\n")
	for _, warning := range warnings {
		fmt.Fprintf(w, "  - %s (%s)\n", warning.Path, warning.Reason)
	}
	fmt.Fprintf(w, "\nConsider using dedicated secrets management. See 'dot help secrets'.\n\n")
}
