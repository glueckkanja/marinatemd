package yamlio

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/schema"
	"gopkg.in/yaml.v3"
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
// Returns nil, nil if the file doesn't exist (not an error condition).
func (r *Reader) ReadSchema(variableName string) (*schema.Schema, error) {
	// Construct path: {docsPath}/variables/{variableName}.yaml
	yamlPath := filepath.Join(r.docsPath, "variables", variableName+".yaml")

	// Check if file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		return nil, nil // Not an error - file just doesn't exist yet
	}

	// Read file
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", yamlPath, err)
	}

	// Parse YAML
	var s schema.Schema
	if err := yaml.Unmarshal(content, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML from %s: %w", yamlPath, err)
	}

	return &s, nil
}

// SchemaExists checks if a YAML schema file exists for the given variable.
func (r *Reader) SchemaExists(variableName string) (bool, error) {
	yamlPath := filepath.Join(r.docsPath, "variables", variableName+".yaml")
	_, err := os.Stat(yamlPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check if schema exists: %w", err)
	}
	return true, nil
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
func (w *Writer) WriteSchema(s *schema.Schema) error {
	// Ensure docs/variables/ directory exists
	varDir := filepath.Join(w.docsPath, "variables")
	if err := os.MkdirAll(varDir, 0755); err != nil {
		return fmt.Errorf("failed to create variables directory: %w", err)
	}

	// Marshal schema to YAML
	yamlBytes, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal schema to YAML: %w", err)
	}

	// Write to file: {docsPath}/variables/{schema.Variable}.yaml
	yamlPath := filepath.Join(varDir, s.Variable+".yaml")
	if err := os.WriteFile(yamlPath, yamlBytes, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file %s: %w", yamlPath, err)
	}

	return nil
}
