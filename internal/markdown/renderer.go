package markdown

import (
	"errors"

	"github.com/c4a8-azure/marinatemd/internal/schema"
)

// Common errors.
var (
	ErrNotImplemented = errors.New("not yet implemented")
)

// Renderer generates hierarchical markdown from schema models.
type Renderer struct {
	// TODO: Add configuration for markdown rendering style.
}

// NewRenderer creates a new markdown renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderSchema converts a schema to hierarchical markdown documentation.
func (r *Renderer) RenderSchema(_ *schema.Schema) (string, error) {
	// TODO: Implement markdown generation
	// - Render variable name and overall description from _meta
	// - Create nested headings/lists for object hierarchy
	// - Include required/optional flags, defaults, examples
	// - Format tables or definition lists for attributes
	// - Ensure deterministic output (stable ordering)
	return "", nil
}

// Injector handles injecting generated markdown into documentation files.
type Injector struct {
	// TODO: Add state for tracking injection points.
}

// NewInjector creates a new markdown injector.
func NewInjector() *Injector {
	return &Injector{}
}

// InjectIntoFile replaces content between markers in a documentation file
// InjectIntoFile injects generated markdown into a documentation file.
// Looks for <!-- MARINATED: variable_name --> and <!-- /MARINATED: variable_name -->.
func (i *Injector) InjectIntoFile(_ string, _ string, _ string) error {
	// TODO: Implement injection logic
	// - Read file content
	// - Find marker pairs for the given variable
	// - Replace content between markers
	// - Write back to file
	// - Preserve surrounding content exactly
	return nil
}

// FindMarkers scans a file and returns all MARINATED markers found.
func (i *Injector) FindMarkers(_ string) ([]string, error) {
	// TODO: Parse file and extract all <!-- MARINATED: name --> markers.
	return nil, ErrNotImplemented
}
