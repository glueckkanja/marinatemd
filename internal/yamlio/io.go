package yamlio

import (
	"errors"

	"github.com/c4a8-azure/marinatemd/internal/schema"
)

// Common errors.
var (
	ErrNotImplemented = errors.New("not yet implemented")
)

// Reader handles reading YAML schema files from disk.
type Reader struct {
	docsPath string // Base path for docs/variables/ directory
}

// NewReader creates a new YAML reader.
func NewReader(docsPath string) *Reader {
	return &Reader{
		docsPath: docsPath,
	}
}

// ReadSchema reads a YAML schema file for the given variable name.
func (r *Reader) ReadSchema(_ string) (*schema.Schema, error) {
	// TODO: Implement YAML reading.
	// - Construct path: {docsPath}/variables/{variableName}.yaml
	// - Parse YAML into Schema struct
	// - Handle file not found gracefully (return nil, nil for new schemas)
	return nil, ErrNotImplemented
}

// Writer handles writing YAML schema files to disk.
type Writer struct {
	docsPath string // Base path for docs/variables/ directory
}

// NewWriter creates a new YAML writer.
func NewWriter(docsPath string) *Writer {
	return &Writer{
		docsPath: docsPath,
	}
}

// WriteSchema writes a schema to a YAML file.
func (w *Writer) WriteSchema(_ *schema.Schema) error {
	// TODO: Implement YAML writing
	// - Ensure docs/variables/ directory exists
	// - Marshal schema to YAML with proper formatting
	// - Write to {docsPath}/variables/{schema.Variable}.yaml
	// - Preserve formatting and comments where possible
	return nil
}

// SchemaExists checks if a YAML schema file exists for the given variable.
func (r *Reader) SchemaExists(_ string) (bool, error) {
	// TODO: Check if YAML file exists.
	return false, ErrNotImplemented
}
