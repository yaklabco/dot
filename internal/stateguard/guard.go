package stateguard

import (
	"context"
	"fmt"
	"io"

	"github.com/yaklabco/dot/internal/cli/prompt"
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
	In          io.Reader
	Out         io.Writer
	Skip        bool   // Auto-continue without prompting (batch mode / non-TTY)
	ManifestDir string // Directory containing the manifest file
	TargetDir   string // Target directory for symlinks
	ConfigPath  string // Path to the config file (for backup)
	HomeDir     string // Home directory (for backup location)
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

	// Show summary and prompt
	PrintSummary(opts.Out, state)

	options := []string{
		"Continue — use the existing installation as-is",
		"Start fresh — remove all managed symlinks and manifest",
		"Back up and start fresh — archive state then reset",
	}

	p := prompt.New(opts.In, opts.Out)
	choice, err := p.SelectWithDefault("", options, 0)
	if err != nil {
		return ResultNoop, fmt.Errorf("prompt: %w", err)
	}

	switch choice {
	case 0:
		return ResultContinue, ActionContinue()
	case 1:
		if err := ActionFresh(opts.ManifestDir, opts.TargetDir); err != nil {
			return ResultFresh, fmt.Errorf("start fresh: %w", err)
		}
		return ResultFresh, nil
	case 2:
		backupDir, err := ActionBackupAndFresh(opts.ManifestDir, opts.TargetDir, opts.ConfigPath, opts.HomeDir)
		if err != nil {
			return ResultBackupAndFresh, fmt.Errorf("backup and start fresh: %w", err)
		}
		fmt.Fprintf(opts.Out, "Backed up to: %s\n", backupDir)
		return ResultBackupAndFresh, nil
	default:
		// Invalid selection or cancelled: default to continue
		return ResultContinue, ActionContinue()
	}
}
