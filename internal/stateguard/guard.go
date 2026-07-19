package stateguard

import (
	"context"
	"fmt"
	"io"

	"github.com/yaklabco/dot/internal/cli/output"
	"github.com/yaklabco/dot/internal/cli/prompt"
	"github.com/yaklabco/dot/internal/cli/render"
)

// CheckResult indicates the outcome of the state guard check.
type CheckResult int

const (
	// ResultNoop means no existing state was found (true first run).
	ResultNoop CheckResult = iota
	// ResultAlreadyAcknowledged means the marker file already exists.
	ResultAlreadyAcknowledged
	// ResultContinue means the user chose to continue with existing state.
	ResultContinue
	// ResultFresh means the user chose to start fresh.
	ResultFresh
	// ResultBackupAndFresh means the user chose to back up then start fresh.
	ResultBackupAndFresh
)

// GuardOptions configures the state guard check.
type GuardOptions struct {
	In           io.Reader
	Out          io.Writer
	Skip         bool   // Auto-continue without prompting (batch mode / non-TTY)
	ColorEnabled bool   // Whether to use colored output
	ManifestDir  string // Directory containing the manifest file
	TargetDir    string // Target directory for symlinks
	ConfigPath   string // Path to the config file (for backup)
	HomeDir      string // Home directory (for backup location)
}

// Check is the main entry point for the first-run state guard.
// It detects existing state and prompts the user for action.
func Check(ctx context.Context, opts GuardOptions) (CheckResult, error) {
	if MarkerExists() {
		return ResultAlreadyAcknowledged, nil
	}

	state, err := DetectExistingState(ctx, opts.ManifestDir, opts.TargetDir)
	if err != nil {
		return ResultNoop, fmt.Errorf("detect existing state: %w", err)
	}
	if state == nil {
		return ResultNoop, nil
	}

	// Skip mode (batch / non-TTY): auto-continue
	if opts.Skip {
		return ResultContinue, ActionContinue()
	}

	f := output.NewFormatter(opts.Out, opts.ColorEnabled)
	c := render.NewColorizer(opts.ColorEnabled)

	// Show summary
	f.BlankLine()
	f.Warning("Existing dot installation detected")
	f.BlankLine()
	f.Bullet(fmt.Sprintf("%s  %s  %s",
		c.Bold(fmt.Sprintf("%d %s", state.PackageCount, pluralize(state.PackageCount, "package", "packages"))),
		c.Dim("·"),
		c.Bold(fmt.Sprintf("%d %s", state.LinkCount, pluralize(state.LinkCount, "symlink", "symlinks"))),
	))
	f.Bullet(fmt.Sprintf("%s %s", c.Dim("manifest:"), c.Dim(state.ManifestPath)))
	f.BlankLine()

	// Prompt
	options := []string{
		fmt.Sprintf("%s — use the existing installation as-is", c.Bold("Continue")),
		fmt.Sprintf("%s — remove all managed symlinks and manifest", c.Bold("Start fresh")),
		fmt.Sprintf("%s — archive state then reset", c.Bold("Back up and start fresh")),
	}

	p := prompt.New(opts.In, opts.Out)
	choice, err := p.SelectWithDefault("", options, 0)
	if err != nil {
		return ResultNoop, fmt.Errorf("prompt: %w", err)
	}

	f.BlankLine()

	switch choice {
	case 0:
		f.SuccessSimple("Continuing with existing installation")
		f.BlankLine()
		return ResultContinue, ActionContinue()
	case 1:
		if err := ActionFresh(opts.ManifestDir, opts.TargetDir); err != nil {
			return ResultFresh, fmt.Errorf("start fresh: %w", err)
		}
		f.SuccessSimple("Removed managed symlinks and manifest")
		f.BlankLine()
		return ResultFresh, nil
	case 2:
		backupDir, err := ActionBackupAndFresh(opts.ManifestDir, opts.TargetDir, opts.ConfigPath, opts.HomeDir)
		if err != nil {
			return ResultBackupAndFresh, fmt.Errorf("backup and start fresh: %w", err)
		}
		f.SuccessSimple(fmt.Sprintf("Backed up to %s", c.Accent(backupDir)))
		f.BlankLine()
		return ResultBackupAndFresh, nil
	default:
		// Invalid selection or cancelled: default to continue
		f.SuccessSimple("Continuing with existing installation")
		f.BlankLine()
		return ResultContinue, ActionContinue()
	}
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
