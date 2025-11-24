package adopt

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestNewInteractiveAdopter(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, true, fs, "/tmp/test-config")

	assert.NotNil(t, adopter)
	assert.Equal(t, input, adopter.input)
	assert.Equal(t, output, adopter.output)
	assert.True(t, adopter.colorize)
}

func TestRun_NoCandidates(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	ctx := context.Background()
	candidates := []DotfileCandidate{}

	groups, err := adopter.Run(ctx, candidates)

	require.NoError(t, err)
	assert.Nil(t, groups)
	assert.Contains(t, output.String(), "No adoptable dotfiles found")
}

func TestRun_NoSelection(t *testing.T) {
	t.Skip("Skipping: Bubble Tea interactive tests require TTY or complex mocking")

	// User quits without selecting
	input := strings.NewReader("q")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	ctx := context.Background()
	now := time.Now()
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			Size:         100,
			ModTime:      now,
			Category:     "shell",
			SuggestedPkg: "bash",
		},
	}

	groups, err := adopter.Run(ctx, candidates)

	require.NoError(t, err)
	assert.Nil(t, groups)
}

func TestRun_Cancellation(t *testing.T) {
	t.Skip("Skipping: Bubble Tea interactive tests require TTY or complex mocking")

	// User quits
	input := strings.NewReader("q")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	ctx := context.Background()
	now := time.Now()
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			Size:         100,
			ModTime:      now,
			Category:     "shell",
			SuggestedPkg: "bash",
		},
	}

	groups, err := adopter.Run(ctx, candidates)

	require.NoError(t, err)
	assert.Nil(t, groups)
}

func TestRun_FullWorkflow(t *testing.T) {
	t.Skip("Skipping: Bubble Tea interactive tests require TTY or complex mocking")

	// User quits
	input := strings.NewReader("q")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	ctx := context.Background()
	now := time.Now()
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			Size:         100,
			ModTime:      now,
			Category:     "shell",
			SuggestedPkg: "bash",
		},
		{
			Path:         "/home/user/.bash_profile",
			RelPath:      ".bash_profile",
			Size:         50,
			ModTime:      now,
			Category:     "shell",
			SuggestedPkg: "bash",
		},
	}

	groups, err := adopter.Run(ctx, candidates)

	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "bash", groups[0].PackageName)
	assert.Len(t, groups[0].Files, 2)
}

func TestRun_EditPackageName(t *testing.T) {
	t.Skip("Interactive Bubble Tea UI requires a real TTY and cannot be tested with mock input")

	// User selects file, edits package name, and confirms
	// Input: select item 1, edit (edit), new name (custom-bash), confirm (y)
	input := strings.NewReader("1\nedit\ncustom-bash\ny\n")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	ctx := context.Background()
	now := time.Now()
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			Size:         100,
			ModTime:      now,
			Category:     "shell",
			SuggestedPkg: "bash",
		},
	}

	groups, err := adopter.Run(ctx, candidates)

	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "custom-bash", groups[0].PackageName)
}

func TestRun_SkipPackage(t *testing.T) {
	t.Skip("Interactive Bubble Tea UI requires a real TTY and cannot be tested with mock input")

	// User selects multiple files but skips one package
	// Input: select items 1,2, skip bash (n), accept vim (y), confirm (y)
	input := strings.NewReader("1,2\nn\ny\ny\n")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	ctx := context.Background()
	now := time.Now()
	candidates := []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			Size:         100,
			ModTime:      now,
			Category:     "shell",
			SuggestedPkg: "bash",
		},
		{
			Path:         "/home/user/.vimrc",
			RelPath:      ".vimrc",
			Size:         200,
			ModTime:      now,
			Category:     "editor",
			SuggestedPkg: "vim",
		},
	}

	groups, err := adopter.Run(ctx, candidates)

	require.NoError(t, err)
	require.Len(t, groups, 1)
	// bash was skipped (first 'n'), vim was accepted (first 'y')
	assert.Equal(t, "vim", groups[0].PackageName)
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 100, "100B"},
		{"kilobytes", 1024, "1.0KB"},
		{"megabytes", 1024 * 1024, "1.0MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0GB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024, "1.0TB"},
		{"fractional KB", 1536, "1.5KB"},
		{"fractional MB", 1024*1024 + 512*1024, "1.5MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrganizeIntoPackages_EmptyInput(t *testing.T) {
	input := strings.NewReader("\n") // Empty response triggers loop again, so we need at least one input
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	now := time.Now()
	adopter.candidates = []DotfileCandidate{
		{
			Path:         "/home/user/.bashrc",
			RelPath:      ".bashrc",
			Size:         100,
			ModTime:      now,
			Category:     "shell",
			SuggestedPkg: "bash",
		},
	}

	// Providing empty input followed by 'y' to accept default
	input = strings.NewReader("\ny\n")
	adopter.input = input

	selections := []int{0}
	groups, err := adopter.organizeIntoPackages(selections)

	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "bash", groups[0].PackageName)
}

func TestConfirmAdoption_Accept(t *testing.T) {
	input := strings.NewReader("y\n")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	groups := []AdoptGroup{
		{
			PackageName: "bash",
			Files:       []string{"/home/user/.bashrc"},
			Category:    "shell",
		},
	}

	result := adopter.confirmAdoption(groups)
	assert.True(t, result)
}

func TestConfirmAdoption_Reject(t *testing.T) {
	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	groups := []AdoptGroup{
		{
			PackageName: "bash",
			Files:       []string{"/home/user/.bashrc"},
			Category:    "shell",
		},
	}

	result := adopter.confirmAdoption(groups)
	assert.False(t, result)
}

func TestConfirmAdoption_DefaultNo(t *testing.T) {
	input := strings.NewReader("\n") // Empty input should default to no
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	groups := []AdoptGroup{
		{
			PackageName: "bash",
			Files:       []string{"/home/user/.bashrc"},
			Category:    "shell",
		},
	}

	result := adopter.confirmAdoption(groups)
	assert.False(t, result)
}

func TestConfirmAdoption_DisplaysPreview(t *testing.T) {
	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	adopter := NewInteractiveAdopter(input, output, false, fs, "/tmp/test-config")

	groups := []AdoptGroup{
		{
			PackageName: "bash",
			Files:       []string{"/home/user/.bashrc", "/home/user/.bash_profile"},
			Category:    "shell",
		},
		{
			PackageName: "vim",
			Files:       []string{"/home/user/.vimrc"},
			Category:    "editor",
		},
	}

	adopter.confirmAdoption(groups)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Adoption Preview")
	assert.Contains(t, outputStr, "bash")
	assert.Contains(t, outputStr, "vim")
	assert.Contains(t, outputStr, "3 files")
	assert.Contains(t, outputStr, "2 packages")
	assert.Contains(t, outputStr, "Files will be moved")
}
