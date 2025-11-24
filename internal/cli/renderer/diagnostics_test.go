package renderer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestJSONRenderer_RenderDiagnostics(t *testing.T) {
	r := &JSONRenderer{pretty: true}

	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthOK,
		Issues:        []dot.Issue{},
		Statistics: dot.DiagnosticStats{
			TotalLinks: 10,
		},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"overall_health"`)
	assert.Contains(t, output, `"healthy"`)
	assert.Contains(t, output, `"total_links"`)
}

func TestYAMLRenderer_RenderDiagnostics(t *testing.T) {
	r := &YAMLRenderer{indent: 2}

	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthWarnings,
		Issues: []dot.Issue{
			{
				Severity:   dot.SeverityWarning,
				Type:       dot.IssueOrphanedLink,
				Path:       "/test",
				Message:    "Test",
				Suggestion: "Fix it",
			},
		},
		Statistics: dot.DiagnosticStats{},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "overall_health")
	assert.Contains(t, output, "warnings")
	assert.Contains(t, output, "issues")
}

func TestTableRenderer_RenderDiagnostics(t *testing.T) {
	r := &TableRenderer{
		colorize: false,
		scheme:   ColorScheme{},
		width:    80,
	}

	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthErrors,
		Issues: []dot.Issue{
			{
				Severity: dot.SeverityError,
				Type:     dot.IssueBrokenLink,
				Path:     "/test/path",
				Message:  "Broken",
			},
		},
		Statistics: dot.DiagnosticStats{
			TotalLinks:  5,
			BrokenLinks: 1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "errors")
	assert.Contains(t, output, "Total Links: 5")
	assert.Contains(t, output, "Broken Links: 1")
}
