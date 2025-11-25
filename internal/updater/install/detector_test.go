package install

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProbe is a test probe implementation.
type mockProbe struct {
	name      string
	platforms []string
	info      *Info
	err       error
}

func (p *mockProbe) Name() string        { return p.name }
func (p *mockProbe) Platforms() []string { return p.platforms }
func (p *mockProbe) Detect(_ context.Context, _ string) (*Info, error) {
	return p.info, p.err
}

func TestNewDetector_DefaultProbes(t *testing.T) {
	d := NewDetector()
	require.NotNil(t, d)
}

func TestNewDetector_WithOptions(t *testing.T) {
	mockFS := &MockFileSystem{}
	customProbe := &mockProbe{name: "custom"}

	d := NewDetector(
		WithFileSystem(mockFS),
		WithProbes(customProbe),
		WithVersion("1.0.0"),
	)

	require.NotNil(t, d)
}

func TestDetector_Detect_FirstMatchWins(t *testing.T) {
	probe1 := &mockProbe{
		name: "probe1",
		info: &Info{Source: SourceHomebrew, Version: "1.0.0"},
	}
	probe2 := &mockProbe{
		name: "probe2",
		info: &Info{Source: SourceApt, Version: "2.0.0"},
	}

	d := NewDetector(
		WithProbes(probe1, probe2),
		WithExecResolver(func() (string, error) { return "/usr/bin/dot", nil }),
	)

	info, err := d.Detect(context.Background())

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, SourceHomebrew, info.Source) // First probe wins
	assert.Equal(t, "1.0.0", info.Version)
}

func TestDetector_Detect_FallbackToManual(t *testing.T) {
	probe1 := &mockProbe{name: "probe1", info: nil} // No match
	probe2 := &mockProbe{name: "probe2", info: nil} // No match

	d := NewDetector(
		WithProbes(probe1, probe2),
		WithExecResolver(func() (string, error) { return "/usr/bin/dot", nil }),
		WithVersion("1.2.3"),
	)

	info, err := d.Detect(context.Background())

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, SourceManual, info.Source)
	assert.Equal(t, "1.2.3", info.Version)
	assert.False(t, info.CanAutoUpgrade)
}

func TestDetector_Detect_SkipFailedProbes(t *testing.T) {
	probe1 := &mockProbe{
		name: "failing",
		info: nil,
		err:  assert.AnError,
	}
	probe2 := &mockProbe{
		name: "working",
		info: &Info{Source: SourceApt, Version: "1.0.0"},
	}

	d := NewDetector(
		WithProbes(probe1, probe2),
		WithExecResolver(func() (string, error) { return "/usr/bin/dot", nil }),
	)

	info, err := d.Detect(context.Background())

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, SourceApt, info.Source) // Second probe matched
}

func TestDetector_Detect_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	probe := &mockProbe{
		name: "slow",
		info: &Info{Source: SourceHomebrew},
	}

	d := NewDetector(
		WithProbes(probe),
		WithExecResolver(func() (string, error) { return "/usr/bin/dot", nil }),
	)

	_, err := d.Detect(ctx)

	assert.Error(t, err)
}

func TestDetector_ManualFallback(t *testing.T) {
	det := &detector{version: "2.0.0"}
	info := det.manualFallback("/usr/local/bin/dot")

	assert.Equal(t, SourceManual, info.Source)
	assert.Equal(t, "2.0.0", info.Version)
	assert.Equal(t, "/usr/local/bin/dot", info.ExecutablePath)
	assert.False(t, info.CanAutoUpgrade)
	assert.Contains(t, info.UpgradeInstructions, "GitHub")
}

func TestDefaultProbes_Darwin(t *testing.T) {
	// Just verify the function doesn't panic
	probes := defaultProbes(OSFileSystem{}, "1.0.0")
	assert.NotEmpty(t, probes)
}
