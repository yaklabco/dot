package renderer

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestTextRenderer_RenderDiagnostics(t *testing.T) {
	r := &TextRenderer{
		colorize: false,
		scheme:   ColorScheme{},
		width:    80,
	}

	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthOK,
		Issues:        []dot.Issue{},
		Statistics: dot.DiagnosticStats{
			TotalLinks:   10,
			ManagedLinks: 10,
		},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "healthy")
	assert.Contains(t, output, "Total Links: 10")
	assert.Contains(t, output, "No issues found")
}

func TestTextRenderer_RenderDiagnostics_WithIssues(t *testing.T) {
	r := &TextRenderer{
		colorize: false,
		scheme:   ColorScheme{},
		width:    80,
	}

	report := dot.DiagnosticReport{
		OverallHealth: dot.HealthErrors,
		Issues: []dot.Issue{
			{
				Severity:   dot.SeverityError,
				Type:       dot.IssueBrokenLink,
				Path:       "/tmp/test",
				Message:    "Link is broken",
				Suggestion: "Fix the link",
			},
		},
		Statistics: dot.DiagnosticStats{
			TotalLinks:   10,
			ManagedLinks: 10,
			BrokenLinks:  1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "errors")
	assert.Contains(t, output, "Broken Links: 1")
	assert.Contains(t, output, "Link is broken")
	assert.Contains(t, output, "Fix the link")
}

func TestTextRenderer_RenderPlan_NoColorOmitsANSICodes(t *testing.T) {
	r := &TextRenderer{
		colorize: false,
		scheme:   DefaultColorScheme(),
		width:    80,
	}

	plan := dot.Plan{
		Operations: []dot.Operation{
			dot.NewDirCreate("op1", dot.MustParsePath("/target/vim")),
			dot.NewLinkCreate("op2", dot.MustParsePath("/src/vimrc"), dot.MustParseTargetPath("/target/.vimrc")),
		},
		Metadata: dot.PlanMetadata{
			PackageCount:   1,
			OperationCount: 2,
			LinkCount:      1,
			DirCount:       1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderPlan(&buf, plan)
	require.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "\033[", "no-color output should not contain ANSI escape codes")
	assert.Contains(t, output, "Dry run mode")
	assert.Contains(t, output, "Create symlink")
	assert.Contains(t, output, "Create directory")
}

func TestTextRenderer_RenderPlan_WithColorIncludesANSICodes(t *testing.T) {
	r := &TextRenderer{
		colorize: true,
		scheme:   DefaultColorScheme(),
		width:    80,
	}

	plan := dot.Plan{
		Operations: []dot.Operation{
			dot.NewLinkCreate("op1", dot.MustParsePath("/src/vimrc"), dot.MustParseTargetPath("/target/.vimrc")),
		},
		Metadata: dot.PlanMetadata{
			PackageCount:   1,
			OperationCount: 1,
			LinkCount:      1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderPlan(&buf, plan)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "\033[", "color output should contain ANSI escape codes")
}

func TestTextRenderer_RenderStatus_Empty(t *testing.T) {
	r := &TextRenderer{
		colorize: false,
		scheme:   ColorScheme{},
		width:    80,
	}

	status := dot.Status{
		Packages: []dot.PackageInfo{},
	}

	var buf bytes.Buffer
	err := r.RenderStatus(&buf, status)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "No packages installed")
}

func TestTextRenderer_RenderStatus_WithPackages(t *testing.T) {
	r := &TextRenderer{
		colorize: true,
		scheme:   DefaultColorScheme(),
		width:    80,
	}

	status := dot.Status{
		Packages: []dot.PackageInfo{
			{
				Name:        "vim",
				InstalledAt: time.Now(),
				LinkCount:   5,
				Links:       []string{".vimrc", ".vim/colors", ".vim/syntax"},
			},
		},
	}

	var buf bytes.Buffer
	err := r.RenderStatus(&buf, status)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "vim")
	assert.Contains(t, output, "Links: 5")
	assert.Contains(t, output, ".vimrc")
}
