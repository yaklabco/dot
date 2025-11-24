package dot

import (
	"io"

	"github.com/yaklabco/dot/internal/updater"
)

// UpdateCheckResult contains the result of an update check.
type UpdateCheckResult = updater.CheckResult

// StartupChecker checks for updates at startup.
// It wraps the internal StartupChecker.
type StartupChecker struct {
	checker *updater.StartupChecker
}

// NewStartupChecker creates a new startup update checker.
func NewStartupChecker(currentVersion string, cfg *ExtendedConfig, configDir string, output io.Writer) *StartupChecker {
	return &StartupChecker{
		checker: updater.NewStartupChecker(currentVersion, cfg, configDir, output),
	}
}

// Check performs the update check.
func (c *StartupChecker) Check() (*UpdateCheckResult, error) {
	return c.checker.Check()
}

// ShowNotification displays an update notification if available.
func (c *StartupChecker) ShowNotification(result *UpdateCheckResult) {
	c.checker.ShowNotification(result)
}
