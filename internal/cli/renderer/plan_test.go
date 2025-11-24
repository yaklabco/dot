package renderer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestJSONRenderer_RenderPlan(t *testing.T) {
	r := &JSONRenderer{pretty: true}

	plan := dot.Plan{
		Operations: []dot.Operation{
			dot.NewLinkCreate("op1", dot.MustParsePath("/src/file"), dot.MustParseTargetPath("/dst/file")),
		},
		Metadata: dot.PlanMetadata{
			PackageCount:   1,
			OperationCount: 1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderPlan(&buf, plan)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Operations")
	assert.Contains(t, output, "Metadata")
}

func TestYAMLRenderer_RenderPlan(t *testing.T) {
	r := &YAMLRenderer{indent: 2}

	plan := dot.Plan{
		Operations: []dot.Operation{
			dot.NewLinkCreate("op1", dot.MustParsePath("/src"), dot.MustParseTargetPath("/dst")),
		},
		Metadata: dot.PlanMetadata{
			PackageCount:   1,
			OperationCount: 1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderPlan(&buf, plan)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "packagecount")
}

func TestTextRenderer_RenderPlan(t *testing.T) {
	r := &TextRenderer{
		colorize: false,
		scheme:   ColorScheme{},
		width:    80,
	}

	plan := dot.Plan{
		Operations: []dot.Operation{
			dot.NewLinkCreate("op1", dot.MustParsePath("/src/vimrc"), dot.MustParseTargetPath("/target/.vimrc")),
			dot.NewDirCreate("op2", dot.MustParsePath("/target/dir")),
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
	assert.Contains(t, output, "Plan:")
	assert.Contains(t, output, "Summary:")
}

func TestTableRenderer_RenderPlan(t *testing.T) {
	r := &TableRenderer{}

	plan := dot.Plan{
		Operations: []dot.Operation{
			dot.NewLinkCreate("op1", dot.MustParsePath("/s"), dot.MustParseTargetPath("/t")),
		},
		Metadata: dot.PlanMetadata{
			OperationCount: 1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderPlan(&buf, plan)
	require.NoError(t, err)

	output := buf.String()
	assert.NotEmpty(t, output)
}
