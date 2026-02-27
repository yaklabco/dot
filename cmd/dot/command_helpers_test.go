package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		singular string
		plural   string
		expected string
	}{
		{
			name:     "zero items",
			count:    0,
			singular: "item",
			plural:   "items",
			expected: "0 items",
		},
		{
			name:     "one item",
			count:    1,
			singular: "item",
			plural:   "items",
			expected: "1 item",
		},
		{
			name:     "two items",
			count:    2,
			singular: "item",
			plural:   "items",
			expected: "2 items",
		},
		{
			name:     "five items",
			count:    5,
			singular: "item",
			plural:   "items",
			expected: "5 items",
		},
		{
			name:     "one file",
			count:    1,
			singular: "file",
			plural:   "files",
			expected: "1 file",
		},
		{
			name:     "multiple files",
			count:    10,
			singular: "file",
			plural:   "files",
			expected: "10 files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCount(tt.count, tt.singular, tt.plural)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatSuccessMessage(t *testing.T) {
	t.Run("contains title case verb for single package", func(t *testing.T) {
		var buf bytes.Buffer
		formatSuccessMessage(&buf, "remanaged", 1, false)
		output := buf.String()

		assert.Contains(t, output, "Remanaged")
		assert.Contains(t, output, "1 package")
	})

	t.Run("pluralizes for multiple packages", func(t *testing.T) {
		var buf bytes.Buffer
		formatSuccessMessage(&buf, "managed", 3, false)
		output := buf.String()

		assert.Contains(t, output, "Managed")
		assert.Contains(t, output, "3 packages")
	})
}
