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
	exportPath string // Base path for export/variables/ directory
}

// NewReader creates a new YAML reader.
//
// The exportPath should be the parent directory that contains the "variables" folder.
// For example, if your YAML files are in "/path/to/project/docs/variables/", you should
// pass "/path/to/project/docs" as the exportPath. The Reader will automatically append
// "variables" to construct the full path to the YAML files.
//
// This design allows the Reader to work with the standard directory structure where
// all schema YAML files are stored in a "variables" subdirectory.
func NewReader(exportPath string) *Reader {
	return &Reader{
		exportPath: exportPath,
	}
}

// ReadSchema reads a YAML schema file for the given variable name.
// Returns nil, nil if the file doesn't exist (not an error condition).
func (r *Reader) ReadSchema(variableName string) (*schema.Schema, error) {
	// Construct path: {exportPath}/variables/{variableName}.yaml
	yamlPath := filepath.Join(r.exportPath, "variables", variableName+".yaml")

	// Check if file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		//nolint:nilnil // Intentional: nil schema with nil error indicates file doesn't exist yet
		return nil, nil // Not an error - file just doesn't exist yet
	}

	// Read file
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", yamlPath, err)
	}

	// Parse YAML
	var s schema.Schema
	if unmarshalErr := yaml.Unmarshal(content, &s); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML from %s: %w", yamlPath, unmarshalErr)
	}

	return &s, nil
}

// SchemaExists checks if a YAML schema file exists for the given variable.
func (r *Reader) SchemaExists(variableName string) (bool, error) {
	yamlPath := filepath.Join(r.exportPath, "variables", variableName+".yaml")
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
	exportPath string // Base path for export/variables/ directory
}

// NewWriter creates a new YAML writer.
func NewWriter(exportPath string) *Writer {
	return &Writer{
		exportPath: exportPath,
	}
}

// WriteSchema writes a schema to a YAML file.
func (w *Writer) WriteSchema(s *schema.Schema) error {
	// Ensure export/variables/ directory exists
	varDir := filepath.Join(w.exportPath, "variables")
	if err := os.MkdirAll(varDir, 0750); err != nil {
		return fmt.Errorf("failed to create variables directory: %w", err)
	}

	// Marshal schema to YAML
	yamlBytes, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal schema to YAML: %w", err)
	}

	// Write to file: {exportPath}/variables/{schema.Variable}.yaml
	yamlPath := filepath.Join(varDir, s.Variable+".yaml")
	if writeErr := os.WriteFile(yamlPath, yamlBytes, 0600); writeErr != nil {
		return fmt.Errorf("failed to write YAML file %s: %w", yamlPath, writeErr)
	}

	return nil
}
