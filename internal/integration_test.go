package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"github.com/c4a8-azure/marinatemd/internal/schema"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
)

func TestEndToEndPipeline(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test HCL file with MARINATED marker
	hclContent := `
variable "app_config" {
  type = object({
    database = optional(object({
      host     = string
      port     = optional(number, 5432)
      ssl_mode = optional(string, "require")
    }))
    cache = optional(object({
      redis_url = string
      ttl       = optional(number, 3600)
    }))
  })
  description = "<!-- MARINATED: app_config --> Application configuration"
}

variable "plain_var" {
  type        = string
  description = "This is not marinated"
}
`

	hclFile := filepath.Join(tmpDir, "variables.tf")
	if err := os.WriteFile(hclFile, []byte(hclContent), 0644); err != nil {
		t.Fatalf("failed to write test HCL file: %v", err)
	}

	// Step 1: Parse HCL
	parser := hclparse.NewParser()
	if err := parser.ParseVariables(tmpDir); err != nil {
		t.Fatalf("ParseVariables() error = %v", err)
	}

	marinatedVars, err := parser.ExtractMarinatedVars()
	if err != nil {
		t.Fatalf("ExtractMarinatedVars() error = %v", err)
	}

	if len(marinatedVars) != 1 {
		t.Fatalf("expected 1 marinated variable, got %d", len(marinatedVars))
	}

	t.Logf("✓ Parsed %d marinated variable(s)", len(marinatedVars))

	// Step 2: Build schema
	builder := schema.NewBuilder()
	s, err := builder.BuildFromVariable(marinatedVars[0])
	if err != nil {
		t.Fatalf("BuildFromVariable() error = %v", err)
	}

	if s.Variable != "app_config" {
		t.Errorf("expected variable name 'app_config', got %s", s.Variable)
	}

	if len(s.SchemaNodes) != 2 {
		t.Errorf("expected 2 top-level nodes (database, cache), got %d", len(s.SchemaNodes))
	}

	t.Logf("✓ Built schema for '%s' with %d top-level nodes", s.Variable, len(s.SchemaNodes))

	// Step 3: Write YAML
	writer := yamlio.NewWriter(tmpDir)
	if writeErr := writer.WriteSchema(s); writeErr != nil {
		t.Fatalf("WriteSchema() error = %v", writeErr)
	}

	yamlPath := filepath.Join(tmpDir, "variables", s.Variable+".yaml")
	if _, statErr := os.Stat(yamlPath); os.IsNotExist(statErr) {
		t.Fatalf("expected YAML file to exist at %s", yamlPath)
	}

	t.Logf("✓ Wrote YAML schema to %s", yamlPath)

	// Step 4: Read it back
	reader := yamlio.NewReader(tmpDir)
	readBack, readErr := reader.ReadSchema(s.Variable)
	if readErr != nil {
		t.Fatalf("ReadSchema() error = %v", readErr)
	}

	if readBack.Variable != s.Variable {
		t.Errorf("variable mismatch: got %s, want %s", readBack.Variable, s.Variable)
	}

	// Verify structure
	if _, ok := readBack.SchemaNodes["database"]; !ok {
		t.Error("expected 'database' node in read-back schema")
	}
	if _, ok := readBack.SchemaNodes["cache"]; !ok {
		t.Error("expected 'cache' node in read-back schema")
	}

	t.Logf("✓ Successfully read back schema from YAML")

	// Step 5: Test merging - simulate user editing YAML
	// Modify the schema to add user descriptions
	readBack.SchemaNodes["database"].Meta.Description = "User-written database description"
	readBack.SchemaNodes["database"].Children["host"].Description = "The database hostname or IP"

	// Write the modified version
	if writeErr2 := writer.WriteSchema(readBack); writeErr2 != nil {
		t.Fatalf("WriteSchema() (modified) error = %v", writeErr2)
	}

	// Parse again (simulate code change)
	newSchema, err := builder.BuildFromVariable(marinatedVars[0])
	if err != nil {
		t.Fatalf("BuildFromVariable() (second time) error = %v", err)
	}

	// Read existing YAML
	existing, readExistingErr := reader.ReadSchema("app_config")
	if readExistingErr != nil {
		t.Fatalf("ReadSchema() (existing) error = %v", readExistingErr)
	}

	// Merge
	merged, err := builder.MergeWithExisting(newSchema, existing)
	if err != nil {
		t.Fatalf("MergeWithExisting() error = %v", err)
	}

	// Verify user descriptions were preserved
	if merged.SchemaNodes["database"].Meta.Description != "User-written database description" {
		t.Errorf("expected user description to be preserved, got %s",
			merged.SchemaNodes["database"].Meta.Description)
	}

	if merged.SchemaNodes["database"].Children["host"].Description != "The database hostname or IP" {
		t.Errorf("expected user host description to be preserved, got %s",
			merged.SchemaNodes["database"].Children["host"].Description)
	}

	t.Logf("✓ Merge preserved user descriptions")
	t.Log("\n✓✓✓ END-TO-END PIPELINE TEST PASSED ✓✓✓")
}
