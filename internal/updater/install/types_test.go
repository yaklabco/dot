package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSource_String(t *testing.T) {
	tests := []struct {
		source   Source
		expected string
	}{
		{SourceHomebrew, "homebrew"},
		{SourceApt, "apt"},
		{SourcePacman, "pacman"},
		{SourceChocolatey, "chocolatey"},
		{SourceGoInstall, "go-install"},
		{SourceBuild, "source"},
		{SourceManual, "manual"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.source.String())
		})
	}
}

func TestInfo_Fields(t *testing.T) {
	info := &Info{
		Source:              SourceHomebrew,
		Version:             "1.0.0",
		ExecutablePath:      "/opt/homebrew/Cellar/dot/1.0.0/bin/dot",
		Metadata:            map[string]string{"tap": "yaklabco/dot"},
		CanAutoUpgrade:      true,
		UpgradeInstructions: "brew upgrade yaklabco/dot/dot",
	}

	assert.Equal(t, SourceHomebrew, info.Source)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "/opt/homebrew/Cellar/dot/1.0.0/bin/dot", info.ExecutablePath)
	assert.Equal(t, "yaklabco/dot", info.Metadata["tap"])
	assert.True(t, info.CanAutoUpgrade)
	assert.Equal(t, "brew upgrade yaklabco/dot/dot", info.UpgradeInstructions)
}

func TestUpgradeResult_Fields(t *testing.T) {
	result := &UpgradeResult{
		Success:         true,
		PreviousVersion: "1.0.0",
		NewVersion:      "1.1.0",
		Output:          "Upgraded successfully",
		Error:           nil,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "1.0.0", result.PreviousVersion)
	assert.Equal(t, "1.1.0", result.NewVersion)
	assert.Equal(t, "Upgraded successfully", result.Output)
	assert.Nil(t, result.Error)
}
