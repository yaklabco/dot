package dot

import "github.com/yaklabco/dot/internal/config"

// ExtendedConfig contains all application configuration.
// It is an alias to the internal ExtendedConfig to provide a stable API.
type ExtendedConfig = config.ExtendedConfig

// DefaultExtendedConfig returns extended configuration with sensible defaults.
func DefaultExtendedConfig() *ExtendedConfig {
	return config.DefaultExtended()
}

// LoadExtendedFromFile loads extended configuration from specified file.
func LoadExtendedFromFile(path string) (*ExtendedConfig, error) {
	return config.LoadExtendedFromFile(path)
}

// ConfigLoader handles configuration loading with precedence.
type ConfigLoader struct {
	loader *config.Loader
}

// NewConfigLoader creates a new configuration loader.
func NewConfigLoader(appName, configPath string) *ConfigLoader {
	return &ConfigLoader{
		loader: config.NewLoader(appName, configPath),
	}
}

// LoadWithEnv loads configuration with environment variable overrides.
func (l *ConfigLoader) LoadWithEnv() (*ExtendedConfig, error) {
	return l.loader.LoadWithEnv()
}
