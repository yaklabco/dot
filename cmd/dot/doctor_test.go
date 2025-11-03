package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/internal/cli/render"
	"github.com/jamesainslie/dot/pkg/dot"
)

func TestNewDoctorCommand(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewDoctorCommand(cfg)

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "doctor")
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestDoctorCommand_Flags(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewDoctorCommand(cfg)

	// Check that format flag exists
	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	// Check that color flag exists
	colorFlag := cmd.Flags().Lookup("color")
	require.NotNil(t, colorFlag)
	assert.Equal(t, "auto", colorFlag.DefValue)
}

func TestDoctorCommand_Help(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewDoctorCommand(cfg)

	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Long, "Exit codes")
}

func TestGetHealthDisplay(t *testing.T) {
	tests := []struct {
		name         string
		health       dot.HealthStatus
		expectedIcon string
		expectedText string
	}{
		{"healthy", dot.HealthOK, "✓", "Healthy"},
		{"warnings", dot.HealthWarnings, "⚠", "Warnings detected"},
		{"errors", dot.HealthErrors, "✗", "Errors detected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon, text, colorFunc := getHealthDisplay(tt.health)
			assert.Contains(t, icon, tt.expectedIcon)
			assert.Equal(t, tt.expectedText, text)
			assert.NotNil(t, colorFunc)
		})
	}
}

func TestFilterIssuesBySeverity(t *testing.T) {
	issues := []dot.Issue{
		{Severity: dot.SeverityError, Path: "error1"},
		{Severity: dot.SeverityWarning, Path: "warning1"},
		{Severity: dot.SeverityError, Path: "error2"},
		{Severity: dot.SeverityInfo, Path: "info1"},
	}

	t.Run("filter errors", func(t *testing.T) {
		errors := filterIssuesBySeverity(issues, dot.SeverityError)
		assert.Len(t, errors, 2)
		assert.Equal(t, "error1", errors[0].Path)
		assert.Equal(t, "error2", errors[1].Path)
	})

	t.Run("filter warnings", func(t *testing.T) {
		warnings := filterIssuesBySeverity(issues, dot.SeverityWarning)
		assert.Len(t, warnings, 1)
		assert.Equal(t, "warning1", warnings[0].Path)
	})

	t.Run("filter info", func(t *testing.T) {
		infos := filterIssuesBySeverity(issues, dot.SeverityInfo)
		assert.Len(t, infos, 1)
		assert.Equal(t, "info1", infos[0].Path)
	})

	t.Run("no matches", func(t *testing.T) {
		// Use a severity that doesn't exist in the test data
		filtered := filterIssuesBySeverity([]dot.Issue{}, dot.SeverityError)
		assert.Len(t, filtered, 0)
	})
}

func TestRenderSuccinctDiagnostics(t *testing.T) {
	// Save and restore NO_COLOR
	orig := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", orig)
		}
	}()
	os.Setenv("NO_COLOR", "1")

	t.Run("healthy report", func(t *testing.T) {
		report := dot.DiagnosticReport{
			OverallHealth: dot.HealthOK,
			Issues:        []dot.Issue{},
			Statistics: dot.DiagnosticStats{
				TotalLinks:   10,
				ManagedLinks: 10,
			},
		}

		var buf bytes.Buffer
		renderSuccinctDiagnostics(&buf, report)
		output := buf.String()

		assert.Contains(t, output, "Healthy")
		assert.Contains(t, output, "No issues found")
	})

	t.Run("report with errors and warnings", func(t *testing.T) {
		report := dot.DiagnosticReport{
			OverallHealth: dot.HealthErrors,
			Issues: []dot.Issue{
				{Severity: dot.SeverityError, Path: "/path/to/error", Message: "error message"},
				{Severity: dot.SeverityWarning, Path: "/path/to/warning", Message: "warning message"},
			},
			Statistics: dot.DiagnosticStats{
				TotalLinks:   10,
				BrokenLinks:  1,
				ManagedLinks: 9,
			},
		}

		var buf bytes.Buffer
		renderSuccinctDiagnostics(&buf, report)
		output := buf.String()

		assert.Contains(t, output, "1 errors:")
		assert.Contains(t, output, "1 warnings:")
		assert.Contains(t, output, "/path/to/error")
		assert.Contains(t, output, "/path/to/warning")
	})

	t.Run("empty statistics", func(t *testing.T) {
		report := dot.DiagnosticReport{
			OverallHealth: dot.HealthOK,
			Issues:        []dot.Issue{},
			Statistics:    dot.DiagnosticStats{},
		}

		var buf bytes.Buffer
		renderSuccinctDiagnostics(&buf, report)
		output := buf.String()

		assert.Contains(t, output, "Healthy")
	})
}

func TestRenderIssueList(t *testing.T) {
	// Save and restore NO_COLOR
	orig := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", orig)
		}
	}()
	os.Setenv("NO_COLOR", "1")

	issues := []dot.Issue{
		{Path: "/path/to/file", Message: "test message"},
		{Path: "/another/path", Message: ""},
	}

	colorize := shouldUseColor()
	c := render.NewColorizer(colorize)

	var buf bytes.Buffer
	renderIssueList(&buf, issues, c.Dim)
	output := buf.String()

	assert.Contains(t, output, "/path/to/file")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "/another/path")
}
