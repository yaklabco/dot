package main

import (
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
