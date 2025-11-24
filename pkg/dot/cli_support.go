package dot

import "github.com/yaklabco/dot/internal/config"

// GetConfigPath returns XDG-compliant configuration directory path.
func GetConfigPath(appName string) string {
	return config.GetConfigPath(appName)
}

// ConfigWriter wraps the internal config writer.
type ConfigWriter struct {
	writer *config.Writer
}

// NewConfigWriter creates a new config writer for the given path.
func NewConfigWriter(path string) *ConfigWriter {
	return &ConfigWriter{
		writer: config.NewWriter(path),
	}
}

// WriteDefault writes default configuration with options.
func (w *ConfigWriter) WriteDefault(opts config.WriteOptions) error {
	return w.writer.WriteDefault(opts)
}

// Update updates a configuration value by key.
func (w *ConfigWriter) Update(key, value string) error {
	return w.writer.Update(key, value)
}

// WriteOptions contains options for writing configuration.
type WriteOptions = config.WriteOptions

// UpgradeConfig upgrades the configuration file to the latest format.
func UpgradeConfig(configPath string, force bool) (string, error) {
	return config.UpgradeConfig(configPath, force)
}
