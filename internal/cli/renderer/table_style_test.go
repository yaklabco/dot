package renderer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestTableRenderer_SimpleStyle(t *testing.T) {
	status := dot.Status{
		Packages: []dot.PackageInfo{
			{Name: "test-pkg", LinkCount: 3},
		},
	}

	t.Run("simple style renders plain text table", func(t *testing.T) {
		r := &TableRenderer{
			colorize:   false,
			scheme:     ColorScheme{},
			width:      80,
			tableStyle: "simple",
		}

		var buf bytes.Buffer
		err := r.RenderStatus(&buf, status)
		require.NoError(t, err)

		output := buf.String()
		// Check for simple text formatting (no Unicode box drawing)
		assert.Contains(t, output, "Package")
		assert.Contains(t, output, "Links")
		assert.Contains(t, output, "test-pkg")
		assert.Contains(t, output, "3")
		// Should have dashes for separator
		assert.Contains(t, output, "---")
		// Should NOT have Unicode box drawing characters
		assert.NotContains(t, output, "┌")
		assert.NotContains(t, output, "└")
	})

	t.Run("default style renders modern table", func(t *testing.T) {
		r := &TableRenderer{
			colorize:   false,
			scheme:     ColorScheme{},
			width:      80,
			tableStyle: "default",
		}

		var buf bytes.Buffer
		err := r.RenderStatus(&buf, status)
		require.NoError(t, err)

		output := buf.String()
		// Check for modern table formatting with Unicode box drawing
		assert.Contains(t, output, "PACKAGE")
		assert.Contains(t, output, "LINKS")
		assert.Contains(t, output, "test-pkg")
		// Should have Unicode box drawing characters
		assert.True(t, strings.Contains(output, "┌") || strings.Contains(output, "│"))
	})

	t.Run("empty status with simple style", func(t *testing.T) {
		r := &TableRenderer{
			colorize:   false,
			scheme:     ColorScheme{},
			width:      80,
			tableStyle: "simple",
		}

		emptyStatus := dot.Status{
			Packages: []dot.PackageInfo{},
		}

		var buf bytes.Buffer
		err := r.RenderStatus(&buf, emptyStatus)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No packages installed")
	})
}

func TestTableRenderer_DiagnosticsSimpleStyle(t *testing.T) {
	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthOK,
		Issues: []dot.Issue{
			{
				Severity: dot.SeverityError,
				Type:     dot.IssueBrokenLink,
				Path:     "/test/path",
				Message:  "test message",
			},
		},
		Statistics: dot.DiagnosticStats{
			TotalLinks:   5,
			ManagedLinks: 4,
		},
	}

	t.Run("simple style diagnostics", func(t *testing.T) {
		r := &TableRenderer{
			colorize:   false,
			scheme:     ColorScheme{},
			width:      80,
			tableStyle: "simple",
		}

		var buf bytes.Buffer
		err := r.RenderDiagnostics(&buf, report)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Health Status")
		assert.Contains(t, output, "Statistics")
		assert.Contains(t, output, "test message")
		// Simple style should have dashes
		assert.Contains(t, output, "---")
	})
}

func TestNewRenderer_WithTableStyle(t *testing.T) {
	tests := []struct {
		name       string
		tableStyle string
		expected   string
	}{
		{
			name:       "default style",
			tableStyle: "default",
			expected:   "default",
		},
		{
			name:       "simple style",
			tableStyle: "simple",
			expected:   "simple",
		},
		{
			name:       "empty defaults to default",
			tableStyle: "",
			expected:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRenderer("table", false, tt.tableStyle)
			require.NoError(t, err)

			tableRenderer, ok := r.(*TableRenderer)
			require.True(t, ok, "Expected TableRenderer")
			assert.Equal(t, tt.expected, tableRenderer.tableStyle)
		})
	}
}
