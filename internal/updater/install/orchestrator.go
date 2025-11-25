package install

import (
	"context"
	"fmt"
	"time"
)

// UpgradeOptions configures upgrade behavior.
type UpgradeOptions struct {
	// DryRun if true, shows what would be done without executing.
	DryRun bool

	// SkipVerify if true, skips post-upgrade version verification.
	SkipVerify bool

	// ExpectedVersion is the version to verify after upgrade.
	ExpectedVersion string

	// Timeout is the maximum time to wait for upgrade completion.
	Timeout time.Duration
}

// DefaultUpgradeOptions returns default upgrade options.
func DefaultUpgradeOptions() UpgradeOptions {
	return UpgradeOptions{
		DryRun:     false,
		SkipVerify: false,
		Timeout:    DefaultTimeout,
	}
}

// UpgradeOrchestrator coordinates the detect-upgrade-verify flow.
type UpgradeOrchestrator struct {
	detector  Detector
	upgraders []Upgrader
	executor  *Executor
}

// OrchestratorOption configures an UpgradeOrchestrator.
type OrchestratorOption func(*UpgradeOrchestrator)

// WithUpgraders sets custom upgraders.
func WithUpgraders(upgraders ...Upgrader) OrchestratorOption {
	return func(o *UpgradeOrchestrator) {
		o.upgraders = upgraders
	}
}

// WithExecutor sets a custom executor.
func WithExecutor(executor *Executor) OrchestratorOption {
	return func(o *UpgradeOrchestrator) {
		o.executor = executor
	}
}

// WithDetector sets a custom detector.
func WithDetector(detector Detector) OrchestratorOption {
	return func(o *UpgradeOrchestrator) {
		o.detector = detector
	}
}

// NewUpgradeOrchestrator creates an orchestrator with default upgraders.
func NewUpgradeOrchestrator(currentVersion string, opts ...OrchestratorOption) *UpgradeOrchestrator {
	o := &UpgradeOrchestrator{}

	for _, opt := range opts {
		opt(o)
	}

	// Set defaults
	if o.executor == nil {
		o.executor = NewExecutor()
	}

	if o.detector == nil {
		o.detector = NewDetector(WithVersion(currentVersion))
	}

	if len(o.upgraders) == 0 {
		o.upgraders = defaultUpgraders(o.executor, o.detector)
	}

	return o
}

// defaultUpgraders returns the standard set of upgraders.
func defaultUpgraders(executor *Executor, detector Detector) []Upgrader {
	return []Upgrader{
		NewBrewUpgrader(executor, detector),
		NewAptUpgrader(executor, detector),
		NewPacmanUpgrader(executor, detector),
		NewChocoUpgrader(executor, detector),
		NewGoInstallUpgrader(executor, detector),
	}
}

// Detect returns installation information.
func (o *UpgradeOrchestrator) Detect(ctx context.Context) (*Info, error) {
	return o.detector.Detect(ctx)
}

// Upgrade performs the upgrade workflow: detect -> upgrade -> verify.
func (o *UpgradeOrchestrator) Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error) {
	// Apply timeout to context
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Configure executor for dry-run if needed
	if opts.DryRun {
		o.executor = NewExecutor(WithDryRun(true))
	}

	// Step 1: Detect installation
	info, err := o.detector.Detect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect installation: %w", err)
	}

	// Check if upgrade is possible
	if !info.CanAutoUpgrade {
		return &UpgradeResult{
			Success:         false,
			PreviousVersion: info.Version,
			Error:           fmt.Errorf("automatic upgrade not supported for %s installations", info.Source),
		}, nil
	}

	// Step 2: Find appropriate upgrader
	var upgrader Upgrader
	for _, u := range o.upgraders {
		if u.CanUpgrade(info) {
			upgrader = u
			break
		}
	}

	if upgrader == nil {
		return &UpgradeResult{
			Success:         false,
			PreviousVersion: info.Version,
			Error:           fmt.Errorf("no upgrader available for source: %s", info.Source),
		}, nil
	}

	// Step 3: Execute upgrade
	result, err := upgrader.Upgrade(ctx, info)
	if err != nil {
		return result, err
	}

	// Step 4: Verify upgrade
	if !opts.SkipVerify && opts.ExpectedVersion != "" && result.Success {
		verified, verifyErr := upgrader.VerifyUpgrade(ctx, opts.ExpectedVersion)
		if verifyErr != nil {
			result.Error = fmt.Errorf("upgrade completed but verification failed: %w", verifyErr)
			return result, nil
		}

		if verified {
			result.NewVersion = opts.ExpectedVersion
		} else {
			// Re-detect to get actual version
			newInfo, detectErr := o.detector.Detect(ctx)
			if detectErr == nil {
				result.NewVersion = newInfo.Version
			}
			result.Error = fmt.Errorf("upgrade completed but version mismatch: expected %s", opts.ExpectedVersion)
		}
	}

	return result, nil
}

// CanUpgrade checks if automatic upgrade is available for the current installation.
func (o *UpgradeOrchestrator) CanUpgrade(ctx context.Context) (bool, *Info, error) {
	info, err := o.detector.Detect(ctx)
	if err != nil {
		return false, nil, err
	}

	if !info.CanAutoUpgrade {
		return false, info, nil
	}

	// Check if we have an upgrader for this source
	for _, u := range o.upgraders {
		if u.CanUpgrade(info) {
			return true, info, nil
		}
	}

	return false, info, nil
}
