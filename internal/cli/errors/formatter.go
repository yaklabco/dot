package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/yaklabco/dot/internal/domain"
	"golang.org/x/term"
)

// Formatter converts domain errors to user-friendly messages.
type Formatter struct {
	colorEnabled bool
	verbosity    int
	width        int
}

// NewFormatter creates a formatter with the given options.
func NewFormatter(colorEnabled bool, verbosity int) *Formatter {
	width := 80 // Default width
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		width = w
	}
	return &Formatter{
		colorEnabled: colorEnabled,
		verbosity:    verbosity,
		width:        width,
	}
}

// Format converts an error to a formatted message.
func (f *Formatter) Format(err error) string {
	return f.FormatWithContext(err, ErrorContext{})
}

// FormatWithContext adds contextual information to error formatting.
func (f *Formatter) FormatWithContext(err error, ctx ErrorContext) string {
	if err == nil {
		return ""
	}

	// Handle ErrMultiple specially
	var errMultiple domain.ErrMultiple
	if errors.As(err, &errMultiple) {
		return f.formatMultiple(errMultiple, ctx)
	}

	// Get template for this error type
	tmpl := f.getTemplate(err, ctx)
	if tmpl == nil {
		// Fallback to simple error string
		return err.Error()
	}

	// Render template
	return tmpl.Render(f.colorEnabled, f.width)
}

// formatMultiple formats multiple errors as a list.
func (f *Formatter) formatMultiple(err domain.ErrMultiple, ctx ErrorContext) string {
	if len(err.Errors) == 0 {
		return "no errors"
	}

	if len(err.Errors) == 1 {
		return f.FormatWithContext(err.Errors[0], ctx)
	}

	var b strings.Builder
	if f.colorEnabled {
		b.WriteString(colorRed)
	}
	fmt.Fprintf(&b, "Multiple errors occurred (%d total)", len(err.Errors))
	if f.colorEnabled {
		b.WriteString(colorReset)
	}
	b.WriteString("\n\n")

	for i, e := range err.Errors {
		fmt.Fprintf(&b, "%d. ", i+1)
		b.WriteString(f.FormatWithContext(e, ctx))
		if i < len(err.Errors)-1 {
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

// getTemplate returns the appropriate template for an error type.
func (f *Formatter) getTemplate(err error, ctx ErrorContext) *Template {
	var suggestion SuggestionEngine
	suggestion.context = ctx

	// Domain Errors
	var invalidPath domain.ErrInvalidPath
	if errors.As(err, &invalidPath) {
		return &Template{
			Title:       "Invalid Path",
			Description: fmt.Sprintf("The path %q is not valid", invalidPath.Path),
			Details:     []string{invalidPath.Reason},
			Suggestions: suggestion.Generate(err),
		}
	}

	var pkgNotFound domain.ErrPackageNotFound
	if errors.As(err, &pkgNotFound) {
		return &Template{
			Title:       "Package Not Found",
			Description: fmt.Sprintf("Package %q does not exist", pkgNotFound.Package),
			Suggestions: suggestion.Generate(err),
		}
	}

	var conflict domain.ErrConflict
	if errors.As(err, &conflict) {
		return &Template{
			Title:       "Conflict Detected",
			Description: fmt.Sprintf("Cannot proceed at %q", conflict.Path),
			Details:     []string{conflict.Reason},
			Suggestions: suggestion.Generate(err),
		}
	}

	var cyclicDep domain.ErrCyclicDependency
	if errors.As(err, &cyclicDep) {
		return &Template{
			Title:       "Circular Dependency Detected",
			Description: "Operations form a dependency cycle",
			Details:     []string{strings.Join(cyclicDep.Cycle, " -> ")},
			Suggestions: suggestion.Generate(err),
		}
	}

	// Infrastructure Errors
	var permDenied domain.ErrPermissionDenied
	if errors.As(err, &permDenied) {
		return &Template{
			Title:       "Permission Denied",
			Description: fmt.Sprintf("Cannot %s %q", permDenied.Operation, permDenied.Path),
			Suggestions: suggestion.Generate(err),
		}
	}

	var fsOp domain.ErrFilesystemOperation
	if errors.As(err, &fsOp) {
		return &Template{
			Title:       "Filesystem Operation Failed",
			Description: fmt.Sprintf("Operation %q failed at %q", fsOp.Operation, fsOp.Path),
			Details:     []string{fsOp.Err.Error()},
			Suggestions: suggestion.Generate(err),
		}
	}

	// Executor Errors
	var emptyPlan domain.ErrEmptyPlan
	if errors.As(err, &emptyPlan) {
		return &Template{
			Title:       "Empty Plan",
			Description: "No operations to execute",
			Suggestions: []string{"Verify the package exists and contains files"},
		}
	}

	var execFailed domain.ErrExecutionFailed
	if errors.As(err, &execFailed) {
		details := []string{
			fmt.Sprintf("%d operations succeeded", execFailed.Executed),
			fmt.Sprintf("%d operations failed", execFailed.Failed),
		}
		if execFailed.RolledBack > 0 {
			details = append(details, fmt.Sprintf("%d operations rolled back", execFailed.RolledBack))
		}
		return &Template{
			Title:       "Execution Failed",
			Description: "Some operations could not be completed",
			Details:     details,
			Suggestions: suggestion.Generate(err),
		}
	}

	var srcNotFound domain.ErrSourceNotFound
	if errors.As(err, &srcNotFound) {
		return &Template{
			Title:       "Source Not Found",
			Description: fmt.Sprintf("Source file does not exist: %q", srcNotFound.Path),
			Suggestions: suggestion.Generate(err),
		}
	}

	var parentNotFound domain.ErrParentNotFound
	if errors.As(err, &parentNotFound) {
		return &Template{
			Title:       "Parent Directory Not Found",
			Description: fmt.Sprintf("Parent directory does not exist: %q", parentNotFound.Path),
			Suggestions: []string{"Ensure parent directories exist before creating files"},
		}
	}

	var checkpointNotFound domain.ErrCheckpointNotFound
	if errors.As(err, &checkpointNotFound) {
		return &Template{
			Title:       "Checkpoint Not Found",
			Description: fmt.Sprintf("Checkpoint %q does not exist", checkpointNotFound.ID),
			Suggestions: []string{"Verify checkpoint ID or list available checkpoints"},
		}
	}

	var notImpl domain.ErrNotImplemented
	if errors.As(err, &notImpl) {
		return &Template{
			Title:       "Not Implemented",
			Description: fmt.Sprintf("%s is not implemented", notImpl.Feature),
			Suggestions: []string{"Use an alternative supported operation or file an issue"},
		}
	}

	return nil
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)
