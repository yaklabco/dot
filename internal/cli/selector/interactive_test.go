package selector

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInteractiveSelector_Select_SingleChoice(t *testing.T) {
	input := strings.NewReader("1\n")
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{"dot-vim", "dot-zsh", "dot-tmux"}

	selected, err := selector.Select(context.Background(), packages)
	require.NoError(t, err)

	assert.Equal(t, []string{"dot-vim"}, selected)
	assert.Contains(t, output.String(), "Package Selection")
	assert.Contains(t, output.String(), "packages available")
	assert.Contains(t, output.String(), "dot-vim")
	assert.Contains(t, output.String(), "dot-zsh")
	assert.Contains(t, output.String(), "dot-tmux")
}

func TestInteractiveSelector_Select_MultipleChoices(t *testing.T) {
	input := strings.NewReader("1,3\n")
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{"dot-vim", "dot-zsh", "dot-tmux"}

	selected, err := selector.Select(context.Background(), packages)
	require.NoError(t, err)

	assert.Equal(t, []string{"dot-vim", "dot-tmux"}, selected)
}

func TestInteractiveSelector_Select_Range(t *testing.T) {
	input := strings.NewReader("1-3\n")
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{"dot-vim", "dot-zsh", "dot-tmux", "dot-git"}

	selected, err := selector.Select(context.Background(), packages)
	require.NoError(t, err)

	assert.Equal(t, []string{"dot-vim", "dot-zsh", "dot-tmux"}, selected)
}

func TestInteractiveSelector_Select_All(t *testing.T) {
	input := strings.NewReader("all\n")
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{"dot-vim", "dot-zsh", "dot-tmux"}

	selected, err := selector.Select(context.Background(), packages)
	require.NoError(t, err)

	assert.Equal(t, []string{"dot-vim", "dot-zsh", "dot-tmux"}, selected)
}

func TestInteractiveSelector_Select_None(t *testing.T) {
	input := strings.NewReader("none\n")
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{"dot-vim", "dot-zsh", "dot-tmux"}

	selected, err := selector.Select(context.Background(), packages)
	require.NoError(t, err)

	assert.Empty(t, selected)
}

func TestInteractiveSelector_Select_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"out of range", "99\n"},
		{"invalid format", "abc\n"},
		{"invalid range", "3-1\n"},
		{"zero", "0\n"},
		{"negative", "-1\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Provide invalid input followed by valid input
			input := strings.NewReader(tt.input + "1\n")
			output := &bytes.Buffer{}

			selector := NewInteractiveSelector(input, output)
			packages := []string{"dot-vim", "dot-zsh"}

			selected, err := selector.Select(context.Background(), packages)
			require.NoError(t, err)

			// Should recover and select based on second input
			assert.Equal(t, []string{"dot-vim"}, selected)
			assert.Contains(t, output.String(), "Invalid selection")
		})
	}
}

func TestInteractiveSelector_Select_EmptyPackageList(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{}

	selected, err := selector.Select(context.Background(), packages)
	require.NoError(t, err)
	assert.Empty(t, selected)
}

func TestInteractiveSelector_Select_ContextCancellation(t *testing.T) {
	input := strings.NewReader("") // No input
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{"dot-vim"}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := selector.Select(ctx, packages)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestInteractiveSelector_Select_MixedInput(t *testing.T) {
	input := strings.NewReader("1, 3-4, 6\n")
	output := &bytes.Buffer{}

	selector := NewInteractiveSelector(input, output)
	packages := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}

	selected, err := selector.Select(context.Background(), packages)
	require.NoError(t, err)

	// Should select indices 1, 3, 4, 6 (packages p1, p3, p4, p6)
	assert.Equal(t, []string{"p1", "p3", "p4", "p6"}, selected)
}

func TestParseSelection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxIndex int
		want     []int
		wantErr  bool
	}{
		{
			name:     "single number",
			input:    "1",
			maxIndex: 5,
			want:     []int{0},
		},
		{
			name:     "multiple numbers",
			input:    "1,3,5",
			maxIndex: 5,
			want:     []int{0, 2, 4},
		},
		{
			name:     "range",
			input:    "1-3",
			maxIndex: 5,
			want:     []int{0, 1, 2},
		},
		{
			name:     "mixed",
			input:    "1, 3-4, 6",
			maxIndex: 7,
			want:     []int{0, 2, 3, 5},
		},
		{
			name:     "all",
			input:    "all",
			maxIndex: 3,
			want:     []int{0, 1, 2},
		},
		{
			name:     "none",
			input:    "none",
			maxIndex: 3,
			want:     []int{},
		},
		{
			name:     "out of range",
			input:    "10",
			maxIndex: 5,
			wantErr:  true,
		},
		{
			name:     "invalid format",
			input:    "abc",
			maxIndex: 5,
			wantErr:  true,
		},
		{
			name:     "zero",
			input:    "0",
			maxIndex: 5,
			wantErr:  true,
		},
		{
			name:     "negative",
			input:    "-1",
			maxIndex: 5,
			wantErr:  true,
		},
		{
			name:     "invalid range",
			input:    "3-1",
			maxIndex: 5,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSelection(tt.input, tt.maxIndex)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
