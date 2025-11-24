package renderer

import (
	"encoding/json"
	"io"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/pkg/dot"
)

// JSONRenderer renders output as JSON.
type JSONRenderer struct {
	pretty bool
}

// newEncoder creates a new JSON encoder with configured settings.
func (r *JSONRenderer) newEncoder(w io.Writer) *json.Encoder {
	encoder := json.NewEncoder(w)
	if r.pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder
}

// RenderStatus renders installation status as JSON.
func (r *JSONRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	return r.newEncoder(w).Encode(status)
}

// RenderDiagnostics renders diagnostic report as JSON.
func (r *JSONRenderer) RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error {
	return r.newEncoder(w).Encode(report)
}

// RenderPlan renders an execution plan as JSON.
func (r *JSONRenderer) RenderPlan(w io.Writer, plan domain.Plan) error {
	return r.newEncoder(w).Encode(plan)
}
