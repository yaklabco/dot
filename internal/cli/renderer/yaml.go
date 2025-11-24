package renderer

import (
	"io"

	"gopkg.in/yaml.v3"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/pkg/dot"
)

// YAMLRenderer renders output as YAML.
type YAMLRenderer struct {
	indent int
}

// newEncoder creates a new YAML encoder with configured settings.
func (r *YAMLRenderer) newEncoder(w io.Writer) *yaml.Encoder {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(r.indent)
	return encoder
}

// RenderStatus renders installation status as YAML.
func (r *YAMLRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	encoder := r.newEncoder(w)
	defer encoder.Close()
	return encoder.Encode(status)
}

// RenderDiagnostics renders diagnostic report as YAML.
func (r *YAMLRenderer) RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error {
	encoder := r.newEncoder(w)
	defer encoder.Close()
	return encoder.Encode(report)
}

// RenderPlan renders an execution plan as YAML.
func (r *YAMLRenderer) RenderPlan(w io.Writer, plan domain.Plan) error {
	encoder := r.newEncoder(w)
	defer encoder.Close()
	return encoder.Encode(plan)
}
