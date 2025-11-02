package prompt

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "yes",
			input:    "y\n",
			expected: true,
		},
		{
			name:     "Yes",
			input:    "Yes\n",
			expected: true,
		},
		{
			name:     "YES",
			input:    "YES\n",
			expected: true,
		},
		{
			name:     "no",
			input:    "n\n",
			expected: false,
		},
		{
			name:     "No",
			input:    "No\n",
			expected: false,
		},
		{
			name:     "empty (defaults to no)",
			input:    "\n",
			expected: false,
		},
		{
			name:     "invalid input",
			input:    "maybe\n",
			expected: false,
		},
		{
			name:     "whitespace around yes",
			input:    "  yes  \n",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			p := New(in, out)

			result, err := p.Confirm("Continue?")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
			assert.Contains(t, out.String(), "Continue?")
			assert.Contains(t, out.String(), "[y/N]:")
		})
	}
}

func TestConfirmWithDefault(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		expected   bool
	}{
		{
			name:       "yes with default yes",
			input:      "y\n",
			defaultYes: true,
			expected:   true,
		},
		{
			name:       "empty with default yes",
			input:      "\n",
			defaultYes: true,
			expected:   true,
		},
		{
			name:       "no with default yes",
			input:      "n\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "yes with default no",
			input:      "y\n",
			defaultYes: false,
			expected:   true,
		},
		{
			name:       "empty with default no",
			input:      "\n",
			defaultYes: false,
			expected:   false,
		},
		{
			name:       "no with default no",
			input:      "n\n",
			defaultYes: false,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			p := New(in, out)

			result, err := p.ConfirmWithDefault("Continue?", tt.defaultYes)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			if tt.defaultYes {
				assert.Contains(t, out.String(), "[Y/n]:")
			} else {
				assert.Contains(t, out.String(), "[y/N]:")
			}
		})
	}
}

func TestInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "hello\n",
			expected: "hello",
		},
		{
			name:     "text with spaces",
			input:    "hello world\n",
			expected: "hello world",
		},
		{
			name:     "empty input",
			input:    "\n",
			expected: "",
		},
		{
			name:     "trimmed whitespace",
			input:    "  hello  \n",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			p := New(in, out)

			result, err := p.Input("Enter name")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
			assert.Contains(t, out.String(), "Enter name:")
		})
	}
}

func TestSelect(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "select first option",
			input:    "1\n",
			expected: 0,
		},
		{
			name:     "select second option",
			input:    "2\n",
			expected: 1,
		},
		{
			name:     "select third option",
			input:    "3\n",
			expected: 2,
		},
		{
			name:     "invalid selection (0)",
			input:    "0\n",
			expected: -1,
		},
		{
			name:     "invalid selection (out of range)",
			input:    "4\n",
			expected: -1,
		},
		{
			name:     "invalid selection (text)",
			input:    "abc\n",
			expected: -1,
		},
		{
			name:     "empty input",
			input:    "\n",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			p := New(in, out)

			result, err := p.Select("Choose an option", options)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
			assert.Contains(t, out.String(), "Choose an option")
			assert.Contains(t, out.String(), "1) Option 1")
			assert.Contains(t, out.String(), "2) Option 2")
			assert.Contains(t, out.String(), "3) Option 3")
		})
	}
}

func TestConfirm_EOF(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}
	p := New(in, out)

	result, err := p.Confirm("Continue?")
	require.NoError(t, err)
	assert.False(t, result)
}

func TestInput_EOF(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}
	p := New(in, out)

	result, err := p.Input("Enter name")
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestSelect_EOF(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}
	p := New(in, out)

	result, err := p.Select("Choose", []string{"A", "B"})
	require.NoError(t, err)
	assert.Equal(t, -1, result)
}
