package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/config"
)

func TestValidateDotfile_AllCases(t *testing.T) {
	t.Run("translate false no validation", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Dotfile.Translate = false
		cfg.Dotfile.Prefix = "" // Should be ok when translate is false

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("translate true with empty prefix", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Dotfile.Translate = true
		cfg.Dotfile.Prefix = ""

		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("translate true with empty prefix errors", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Dotfile.Translate = true
		cfg.Dotfile.Prefix = "" // Empty prefix should error

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prefix")
	})

	t.Run("translate true with valid prefix", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Dotfile.Translate = true
		cfg.Dotfile.Prefix = "dot-"

		err := cfg.Validate()
		assert.NoError(t, err)
	})
}

func TestValidateSymlinks_AllModes(t *testing.T) {
	tests := []struct {
		mode    string
		wantErr bool
	}{
		{"relative", false},
		{"absolute", false},
		{"invalid-mode", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Symlinks.Mode = tt.mode

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLogging_AllDestinations(t *testing.T) {
	tests := []struct {
		dest    string
		wantErr bool
	}{
		{"stderr", false},
		{"stdout", false},
		{"file", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.dest, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Logging.Destination = tt.dest

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIgnore_UseDefaultsCombinations(t *testing.T) {
	t.Run("use defaults true", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Ignore.UseDefaults = true
		cfg.Ignore.Patterns = []string{}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("use defaults false with patterns", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Ignore.UseDefaults = false
		cfg.Ignore.Patterns = []string{".git", ".svn"}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("use defaults false without patterns", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Ignore.UseDefaults = false
		cfg.Ignore.Patterns = []string{}

		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
