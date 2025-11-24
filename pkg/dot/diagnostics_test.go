package dot_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestHealthStatus_String(t *testing.T) {
	tests := []struct {
		status dot.HealthStatus
		want   string
	}{
		{dot.HealthOK, "healthy"},
		{dot.HealthWarnings, "warnings"},
		{dot.HealthErrors, "errors"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.status.String())
	}
}

func TestHealthStatus_MarshalJSON(t *testing.T) {
	status := dot.HealthOK
	data, err := json.Marshal(status)
	require.NoError(t, err)
	assert.Equal(t, `"healthy"`, string(data))
}

func TestHealthStatus_MarshalYAML(t *testing.T) {
	status := dot.HealthWarnings
	data, err := yaml.Marshal(status)
	require.NoError(t, err)
	assert.Contains(t, string(data), "warnings")
}

func TestIssueSeverity_String(t *testing.T) {
	tests := []struct {
		severity dot.IssueSeverity
		want     string
	}{
		{dot.SeverityInfo, "info"},
		{dot.SeverityWarning, "warning"},
		{dot.SeverityError, "error"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.severity.String())
	}
}

func TestIssueSeverity_MarshalJSON(t *testing.T) {
	severity := dot.SeverityError
	data, err := json.Marshal(severity)
	require.NoError(t, err)
	assert.Equal(t, `"error"`, string(data))
}

func TestIssueSeverity_MarshalYAML(t *testing.T) {
	severity := dot.SeverityInfo
	data, err := yaml.Marshal(severity)
	require.NoError(t, err)
	assert.Contains(t, string(data), "info")
}

func TestIssueType_String(t *testing.T) {
	tests := []struct {
		issueType dot.IssueType
		want      string
	}{
		{dot.IssueBrokenLink, "broken_link"},
		{dot.IssueOrphanedLink, "orphaned_link"},
		{dot.IssueWrongTarget, "wrong_target"},
		{dot.IssuePermission, "permission"},
		{dot.IssueCircular, "circular"},
		{dot.IssueManifestInconsistency, "manifest_inconsistency"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.issueType.String())
	}
}

func TestIssueType_MarshalJSON(t *testing.T) {
	issueType := dot.IssueBrokenLink
	data, err := json.Marshal(issueType)
	require.NoError(t, err)
	assert.Equal(t, `"broken_link"`, string(data))
}

func TestIssueType_MarshalYAML(t *testing.T) {
	issueType := dot.IssuePermission
	data, err := yaml.Marshal(issueType)
	require.NoError(t, err)
	assert.Contains(t, string(data), "permission")
}

func TestDiagnosticReport(t *testing.T) {
	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthOK,
		Issues:        []dot.Issue{},
		Statistics: dot.DiagnosticStats{
			TotalLinks: 10,
		},
	}

	assert.Equal(t, dot.HealthOK, report.OverallHealth)
	assert.Empty(t, report.Issues)
	assert.Equal(t, 10, report.Statistics.TotalLinks)
}

func TestDiagnosticReport_JSON(t *testing.T) {
	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthWarnings,
		Issues: []dot.Issue{
			{
				Severity:   dot.SeverityWarning,
				Type:       dot.IssueOrphanedLink,
				Path:       "/tmp/test",
				Message:    "Test message",
				Suggestion: "Test suggestion",
			},
		},
		Statistics: dot.DiagnosticStats{
			TotalLinks:    5,
			BrokenLinks:   1,
			OrphanedLinks: 2,
			ManagedLinks:  3,
		},
	}

	data, err := json.Marshal(report)
	require.NoError(t, err)

	// Verify JSON structure (unmarshaling enums not supported yet)
	assert.Contains(t, string(data), `"warnings"`)
	assert.Contains(t, string(data), `"warning"`)
	assert.Contains(t, string(data), `"orphaned_link"`)
	assert.Contains(t, string(data), `"total_links":5`)
}
