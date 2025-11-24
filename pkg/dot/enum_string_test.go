package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestHealthStatus_String_Unknown(t *testing.T) {
	status := dot.HealthStatus(99)
	assert.Equal(t, "unknown", status.String())
}

func TestIssueSeverity_String_Unknown(t *testing.T) {
	severity := dot.IssueSeverity(99)
	assert.Equal(t, "unknown", severity.String())
}

func TestIssueType_String_Unknown(t *testing.T) {
	issueType := dot.IssueType(99)
	assert.Equal(t, "unknown", issueType.String())
}

func TestNodeType_String_Unknown(t *testing.T) {
	nodeType := dot.NodeType(99)
	assert.Equal(t, "Unknown", nodeType.String())
}
