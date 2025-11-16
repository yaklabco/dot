package scanner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInteractivePrompter(t *testing.T) {
	prompter := NewInteractivePrompter()
	assert.NotNil(t, prompter)
	assert.NotNil(t, prompter.input)
	assert.NotNil(t, prompter.output)
	assert.False(t, prompter.skipAll)
}

func TestNewInteractivePrompterWithIO(t *testing.T) {
	input := strings.NewReader("test")
	output := &bytes.Buffer{}

	prompter := NewInteractivePrompterWithIO(input, output)
	assert.NotNil(t, prompter)
	assert.Equal(t, input, prompter.input)
	assert.Equal(t, output, prompter.output)
	assert.False(t, prompter.skipAll)
}

func TestInteractivePrompter_ShouldInclude_Include(t *testing.T) {
	input := strings.NewReader("i\n")
	output := &bytes.Buffer{}
	prompter := NewInteractivePrompterWithIO(input, output)

	result := prompter.ShouldInclude("/path/to/large/file.txt", 10*1024*1024, 5*1024*1024)

	assert.True(t, result)
	assert.Contains(t, output.String(), "Large file detected:")
	assert.Contains(t, output.String(), "/path/to/large/file.txt")
	assert.Contains(t, output.String(), "10.0 MB")
	assert.Contains(t, output.String(), "5.0 MB")
	assert.Contains(t, output.String(), "i) Include this file")
}

func TestInteractivePrompter_ShouldInclude_Skip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"explicit skip", "s\n"},
		{"empty input", "\n"},
		{"invalid choice", "x\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			prompter := NewInteractivePrompterWithIO(input, output)

			result := prompter.ShouldInclude("/path/to/file.txt", 1024*1024, 512*1024)

			assert.False(t, result)
			assert.False(t, prompter.skipAll)
		})
	}
}

func TestInteractivePrompter_ShouldInclude_SkipAll(t *testing.T) {
	input := strings.NewReader("a\n")
	output := &bytes.Buffer{}
	prompter := NewInteractivePrompterWithIO(input, output)

	result := prompter.ShouldInclude("/path/to/file1.txt", 1024*1024, 512*1024)

	assert.False(t, result)
	assert.True(t, prompter.skipAll)
}

func TestInteractivePrompter_ShouldInclude_SkipAllPersists(t *testing.T) {
	// First prompt: user chooses "skip all"
	input := strings.NewReader("a\n")
	output := &bytes.Buffer{}
	prompter := NewInteractivePrompterWithIO(input, output)

	result1 := prompter.ShouldInclude("/path/to/file1.txt", 1024*1024, 512*1024)
	assert.False(t, result1)
	assert.True(t, prompter.skipAll)

	// Second call: should skip without prompting
	output.Reset()
	result2 := prompter.ShouldInclude("/path/to/file2.txt", 2*1024*1024, 512*1024)
	assert.False(t, result2)
	// Should not have printed prompt again
	assert.Empty(t, output.String())
}

func TestInteractivePrompter_ShouldInclude_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		skipAll  bool
	}{
		{"uppercase I", "I\n", true, false},
		{"uppercase S", "S\n", false, false},
		{"uppercase A", "A\n", false, true},
		{"mixed case", "iNcLuDe\n", false, false}, // Only first char matters
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			prompter := NewInteractivePrompterWithIO(input, output)

			result := prompter.ShouldInclude("/path/to/file.txt", 1024, 512)

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.skipAll, prompter.skipAll)
		})
	}
}

func TestInteractivePrompter_ShouldInclude_ReadError(t *testing.T) {
	// Use a reader that immediately returns EOF
	input := strings.NewReader("")
	output := &bytes.Buffer{}
	prompter := NewInteractivePrompterWithIO(input, output)

	result := prompter.ShouldInclude("/path/to/file.txt", 1024, 512)

	// On error, should default to skip
	assert.False(t, result)
}

func TestNewBatchPrompter(t *testing.T) {
	prompter := NewBatchPrompter()
	assert.NotNil(t, prompter)
}

func TestBatchPrompter_ShouldInclude(t *testing.T) {
	prompter := NewBatchPrompter()

	// Batch prompter always returns false
	result := prompter.ShouldInclude("/path/to/file.txt", 10*1024*1024, 5*1024*1024)
	assert.False(t, result)

	// Multiple calls still return false
	result = prompter.ShouldInclude("/another/file.txt", 100*1024*1024, 5*1024*1024)
	assert.False(t, result)
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"1023 bytes", 1023, "1023 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1024 * 1024, "1.0 MB"},
		{"10 MB", 10 * 1024 * 1024, "10.0 MB"},
		{"1 GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"1.5 GB", 1536 * 1024 * 1024, "1.5 GB"},
		{"1 TB", 1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{"5.2 TB", 5632 * 1024 * 1024 * 1024, "5.5 TB"},
		{"1 PB", 1024 * 1024 * 1024 * 1024 * 1024, "1.0 PB"},
		{"2.5 PB", 2621440 * 1024 * 1024 * 1024, "2.5 PB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatSize_Precision(t *testing.T) {
	// Test that we get one decimal place
	result := formatSize(1536) // 1.5 KB
	assert.Equal(t, "1.5 KB", result)

	result = formatSize(1587) // ~1.55 KB, should round
	assert.Contains(t, result, "1.5 KB")
}

func TestIsInteractive(t *testing.T) {
	// This test just verifies the function can be called
	// The actual result depends on the test environment
	result := IsInteractive()
	assert.IsType(t, false, result) // Returns a bool
}

func TestInteractivePrompter_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"leading space", " i\n", true},
		{"trailing space", "i \n", true},
		{"both spaces", " i \n", true},
		{"tabs", "\ti\t\n", true},
		{"newline only", "\n", false},
		{"spaces only", "   \n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			prompter := NewInteractivePrompterWithIO(input, output)

			result := prompter.ShouldInclude("/file.txt", 1024, 512)
			assert.Equal(t, tt.expected, result)
		})
	}
}
