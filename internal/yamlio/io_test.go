package yamlio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/schema"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
)

func TestWriter_WriteSchema(t *testing.T) {
	tmpDir := t.TempDir()

	schema := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Type:     "object",
				Required: false,
				Meta: &schema.MetaInfo{
					Description: "Database configuration",
				},
				Children: map[string]*schema.Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "Database host",
					},
					"port": {
						Type:        "number",
						Required:    false,
						Description: "Database port",
					},
				},
			},
		},
	}

	writer := yamlio.NewWriter(tmpDir)
	if err := writer.WriteSchema(schema); err != nil {
		t.Fatalf("WriteSchema() error = %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(tmpDir, "variables", "app_config.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected YAML file to be created at %s", expectedPath)
	}

	// Read it back and verify
	reader := yamlio.NewReader(tmpDir)
	readSchema, err := reader.ReadSchema("app_config")
	if err != nil {
		t.Fatalf("ReadSchema() error = %v", err)
	}

	if readSchema.Variable != schema.Variable {
		t.Errorf("Variable = %v, want %v", readSchema.Variable, schema.Variable)
	}
	if readSchema.Version != schema.Version {
		t.Errorf("Version = %v, want %v", readSchema.Version, schema.Version)
	}
}

func TestReader_ReadSchema_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	reader := yamlio.NewReader(tmpDir)

	// Reading non-existent schema should return nil, nil
	schema, err := reader.ReadSchema("nonexistent")
	if err != nil {
		t.Errorf("expected no error for missing file, got %v", err)
	}
	if schema != nil {
		t.Error("expected nil schema for missing file")
	}
}

func TestReader_ReadSchema_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sample YAML file
	yamlContent := `variable: app_config
version: "1"
schema:
  database:
    type: object
    required: false
    _meta:
      description: "Database configuration"
    host:
      type: string
      required: true
      description: "Database host"
    port:
      type: number
      required: false
      description: "Database port"
`

	// Create directory
	varDir := filepath.Join(tmpDir, "variables")
	if err := os.MkdirAll(varDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Write file
	yamlPath := filepath.Join(varDir, "app_config.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write YAML file: %v", err)
	}

	// Read it
	reader := yamlio.NewReader(tmpDir)
	schema, err := reader.ReadSchema("app_config")
	if err != nil {
		t.Fatalf("ReadSchema() error = %v", err)
	}

	if schema == nil {
		t.Fatal("expected schema to be read")
	}
	if schema.Variable != "app_config" {
		t.Errorf("Variable = %v, want app_config", schema.Variable)
	}
	if schema.Version != "1" {
		t.Errorf("Version = %v, want 1", schema.Version)
	}

	// Check database node
	db, ok := schema.SchemaNodes["database"]
	if !ok {
		t.Fatal("expected database node")
	}
	if db.Type != "object" {
		t.Errorf("database type = %v, want object", db.Type)
	}
	if db.Required {
		t.Error("expected database to be optional")
	}
}

func TestWriter_WriteSchema_ComplexNested(t *testing.T) {
	tmpDir := t.TempDir()

	schema := &schema.Schema{
		Variable: "network_rules",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"private_link_access": {
				Type:        "list",
				Required:    false,
				ElementType: "object",
				Meta: &schema.MetaInfo{
					Description: "Private link access rules",
				},
				Children: map[string]*schema.Node{
					"endpoint_resource_id": {
						Type:        "string",
						Required:    true,
						Description: "Resource ID",
					},
					"endpoint_tenant_id": {
						Type:        "string",
						Required:    false,
						Description: "Tenant ID",
					},
				},
			},
		},
	}

	writer := yamlio.NewWriter(tmpDir)
	if err := writer.WriteSchema(schema); err != nil {
		t.Fatalf("WriteSchema() error = %v", err)
	}

	// Read it back
	reader := yamlio.NewReader(tmpDir)
	readSchema, err := reader.ReadSchema("network_rules")
	if err != nil {
		t.Fatalf("ReadSchema() error = %v", err)
	}

	// Verify structure
	pla := readSchema.SchemaNodes["private_link_access"]
	if pla.Type != "list" {
		t.Errorf("type = %v, want list", pla.Type)
	}
	if pla.ElementType != "object" {
		t.Errorf("element_type = %v, want object", pla.ElementType)
	}

	// Check children
	eri, ok := pla.Children["endpoint_resource_id"]
	if !ok {
		t.Fatal("expected endpoint_resource_id child")
	}
	if !eri.Required {
		t.Error("expected endpoint_resource_id to be required")
	}
}

func TestSchemaExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a schema file
	schema := &schema.Schema{
		Variable:    "existing_var",
		Version:     "1",
		SchemaNodes: map[string]*schema.Node{},
	}

	writer := yamlio.NewWriter(tmpDir)
	if err := writer.WriteSchema(schema); err != nil {
		t.Fatalf("WriteSchema() error = %v", err)
	}

	reader := yamlio.NewReader(tmpDir)

	// Check existing file
	exists, err := reader.SchemaExists("existing_var")
	if err != nil {
		t.Errorf("SchemaExists() error = %v", err)
	}
	if !exists {
		t.Error("expected schema to exist")
	}

	// Check non-existing file
	exists, err = reader.SchemaExists("nonexistent")
	if err != nil {
		t.Errorf("SchemaExists() error = %v", err)
	}
	if exists {
		t.Error("expected schema to not exist")
	}
}

func TestWriter_PreserveExistingDescriptions(t *testing.T) {
	tmpDir := t.TempDir()

	// Write initial schema with user descriptions
	initialSchema := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Type:     "object",
				Required: false,
				Meta: &schema.MetaInfo{
					Description: "User-written database description",
				},
				Children: map[string]*schema.Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "User-written host description",
					},
				},
			},
		},
	}

	writer := yamlio.NewWriter(tmpDir)
	if err := writer.WriteSchema(initialSchema); err != nil {
		t.Fatalf("WriteSchema() error = %v", err)
	}

	// Read it back to verify it was written correctly
	reader := yamlio.NewReader(tmpDir)
	readBack, err := reader.ReadSchema("app_config")
	if err != nil {
		t.Fatalf("ReadSchema() error = %v", err)
	}

	// Verify descriptions were preserved
	db := readBack.SchemaNodes["database"]
	if db.Meta.Description != "User-written database description" {
		t.Errorf("expected description to be preserved, got %v", db.Meta.Description)
	}

	host := db.Children["host"]
	if host.Description != "User-written host description" {
		t.Errorf("expected host description to be preserved, got %v", host.Description)
	}
}
