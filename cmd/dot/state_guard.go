package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yaklabco/dot/internal/cli/terminal"
	"github.com/yaklabco/dot/internal/config"
	"github.com/yaklabco/dot/internal/stateguard"
)

// runStateGuard checks for existing dot state on first mutating command
// and lets the user choose how to proceed. Returns nil if no action needed.
func runStateGuard(cmd *cobra.Command) error {
	manifestDir, targetDir, configPath, homeDir := resolveGuardPaths()

	skip := GetCLIFlags().batch || !terminal.IsInteractive()

	result, err := stateguard.Check(context.Background(), stateguard.GuardOptions{
		In:           cmd.InOrStdin(),
		Out:          cmd.ErrOrStderr(),
		Skip:         skip,
		ColorEnabled: shouldUseColor(),
		ManifestDir:  manifestDir,
		TargetDir:    targetDir,
		ConfigPath:   configPath,
		HomeDir:      homeDir,
	})
	if err != nil {
		return err
	}

	// Silence is golden for noop and already-acknowledged
	_ = result
	return nil
}

// resolveGuardPaths determines the manifest dir, target dir, config path,
// and home dir from config and flags.
func resolveGuardPaths() (manifestDir, targetDir, configPath, homeDir string) {
	homeDir, _ = os.UserHomeDir()
	if homeDir == "" {
		homeDir, _ = filepath.Abs(".")
	}

	configPath = getConfigFilePath()
	extCfg, _ := loadConfigWithRepoPriority(GetCLIFlags().packageDir, configPath)

	if extCfg != nil && extCfg.Directories.Manifest != "" {
		manifestDir = extCfg.Directories.Manifest
	} else {
		manifestDir = config.GetXDGDataPath("dot/manifest")
	}

	if extCfg != nil && extCfg.Directories.Target != "" && filepath.IsAbs(extCfg.Directories.Target) {
		targetDir = extCfg.Directories.Target
	} else {
		targetDir = homeDir
	}

	return
}
