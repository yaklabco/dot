package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "bytes",
			input:    "100",
			expected: 100,
		},
		{
			name:     "bytes with B suffix",
			input:    "100B",
			expected: 100,
		},
		{
			name:     "kilobytes",
			input:    "10K",
			expected: 10 * 1024,
		},
		{
			name:     "kilobytes with B suffix",
			input:    "10KB",
			expected: 10 * 1024,
		},
		{
			name:     "megabytes",
			input:    "5M",
			expected: 5 * 1024 * 1024,
		},
		{
			name:     "megabytes with B suffix",
			input:    "5MB",
			expected: 5 * 1024 * 1024,
		},
		{
			name:     "gigabytes",
			input:    "2G",
			expected: 2 * 1024 * 1024 * 1024,
		},
		{
			name:     "gigabytes with B suffix",
			input:    "2GB",
			expected: 2 * 1024 * 1024 * 1024,
		},
		{
			name:     "fractional megabytes",
			input:    "1.5M",
			expected: int64(1.5 * 1024 * 1024),
		},
		{
			name:     "fractional gigabytes",
			input:    "0.5G",
			expected: int64(0.5 * 1024 * 1024 * 1024),
		},
		{
			name:     "lowercase k",
			input:    "10k",
			expected: 10 * 1024,
		},
		{
			name:     "lowercase m",
			input:    "10m",
			expected: 10 * 1024 * 1024,
		},
		{
			name:     "with whitespace",
			input:    "  10M  ",
			expected: 10 * 1024 * 1024,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:    "invalid format",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid number",
			input:   "M",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSize(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
