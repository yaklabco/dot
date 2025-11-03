package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatter_Success(t *testing.T) {
	tests := []struct {
		name     string
		verb     string
		count    int
		singular string
		plural   string
		want     string
	}{
		{
			name:     "single item",
			verb:     "managed",
			count:    1,
			singular: "package",
			plural:   "packages",
			want:     "Managed 1 package",
		},
		{
			name:     "multiple items",
			verb:     "managed",
			count:    3,
			singular: "package",
			plural:   "packages",
			want:     "Managed 3 packages",
		},
		{
			name:     "zero items",
			verb:     "adopted",
			count:    0,
			singular: "file",
			plural:   "files",
			want:     "Adopted 0 files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewFormatter(&buf, false) // Disable colors for testing

			f.Success(tt.verb, tt.count, tt.singular, tt.plural)

			output := buf.String()
			assert.Contains(t, output, "✓")
			assert.Contains(t, output, tt.want)
		})
	}
}

func TestFormatter_SuccessSimple(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.SuccessSimple("Upgrade completed")

	output := buf.String()
	assert.Contains(t, output, "✓")
	assert.Contains(t, output, "Upgrade completed")
}

func TestFormatter_Error(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.Error("Operation failed")

	output := buf.String()
	assert.Contains(t, output, "✗")
	assert.Contains(t, output, "Operation failed")
}

func TestFormatter_Warning(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.Warning("Potential issue detected")

	output := buf.String()
	assert.Contains(t, output, "⚠")
	assert.Contains(t, output, "Potential issue detected")
}

func TestFormatter_Info(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.Info("Additional information")

	output := buf.String()
	assert.Contains(t, output, "ℹ")
	assert.Contains(t, output, "Additional information")
}

func TestFormatter_Bullet(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.Bullet("Item one")

	output := buf.String()
	assert.Contains(t, output, "•")
	assert.Contains(t, output, "Item one")
	assert.True(t, strings.HasPrefix(output, "  ")) // Check indentation
}

func TestFormatter_BulletWithDetail(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.BulletWithDetail("main text", "detail text")

	output := buf.String()
	assert.Contains(t, output, "•")
	assert.Contains(t, output, "main text")
	assert.Contains(t, output, "—")
	assert.Contains(t, output, "detail text")
}

func TestFormatter_Header(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.Header("Section Title")

	output := buf.String()
	assert.Contains(t, output, "Section Title")
}

func TestFormatter_WithColors(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, true) // Enable colors

	f.Success("managed", 2, "package", "packages")

	output := buf.String()
	assert.Contains(t, output, "✓")
	assert.Contains(t, output, "Managed 2 packages")
	// Should contain ANSI codes when colors enabled
	assert.Contains(t, output, "\033[")
}

func TestFormatter_Indent(t *testing.T) {
	tests := []struct {
		name  string
		level int
		text  string
		want  string
	}{
		{
			name:  "no indentation",
			level: 0,
			text:  "text",
			want:  "text",
		},
		{
			name:  "single level",
			level: 1,
			text:  "text",
			want:  "  text",
		},
		{
			name:  "multiple levels",
			level: 3,
			text:  "text",
			want:  "      text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewFormatter(&buf, false)

			f.Indent(tt.level, tt.text)

			output := strings.TrimRight(buf.String(), "\n")
			assert.Equal(t, tt.want, output)
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		singular string
		plural   string
		want     string
	}{
		{
			name:     "singular",
			count:    1,
			singular: "package",
			plural:   "packages",
			want:     "package",
		},
		{
			name:     "plural",
			count:    2,
			singular: "package",
			plural:   "packages",
			want:     "packages",
		},
		{
			name:     "zero is plural",
			count:    0,
			singular: "file",
			plural:   "files",
			want:     "files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pluralize(tt.count, tt.singular, tt.plural)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatter_BlankLine(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.BlankLine()

	assert.Equal(t, "\n", buf.String())
}
