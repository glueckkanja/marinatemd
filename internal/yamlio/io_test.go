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
				Marinate: &schema.MarinateInfo{
					Description: "Database configuration",
				},
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "Database host",
						},
						Attributes: map[string]*schema.Node{},
					},
					"port": {
						Marinate: &schema.MarinateInfo{
							Description: "Database port",
						},
						Attributes: map[string]*schema.Node{},
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
    _marinate:
      description: "Database configuration"
      type: object
      required: false
    host:
      _marinate:
        description: "Database host"
        type: string
        required: true
    port:
      _marinate:
        description: "Database port"
        type: number
        required: false
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
	if db.Marinate.Type != "object" {
		t.Errorf("database type = %v, want object", db.Marinate.Type)
	}
	if db.Marinate.Required {
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
				Marinate: &schema.MarinateInfo{
					Description: "Private link access rules",
					Type:        "list",
					ElementType: "object",
					Required:    false,
				},
				Attributes: map[string]*schema.Node{
					"endpoint_resource_id": {
						Marinate: &schema.MarinateInfo{
							Description: "Resource ID",
							Type:        "string",
							Required:    true,
						},
						Attributes: map[string]*schema.Node{},
					},
					"endpoint_tenant_id": {
						Marinate: &schema.MarinateInfo{
							Description: "Tenant ID",
							Type:        "string",
							Required:    false,
						},
						Attributes: map[string]*schema.Node{},
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
	if pla.Marinate.Type != "list" {
		t.Errorf("type = %v, want list", pla.Marinate.Type)
	}
	if pla.Marinate.ElementType != "object" {
		t.Errorf("element_type = %v, want object", pla.Marinate.ElementType)
	}

	// Check attributes
	eri, ok := pla.Attributes["endpoint_resource_id"]
	if !ok {
		t.Fatal("expected endpoint_resource_id attribute")
	}
	if !eri.Marinate.Required {
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
				Marinate: &schema.MarinateInfo{
					Description: "User-written database description",
				},
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "User-written host description",
						},
						Attributes: map[string]*schema.Node{},
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
	if db.Marinate == nil || db.Marinate.Description != "User-written database description" {
		t.Errorf("expected description to be preserved, got %v", db.Marinate)
	}

	host := db.Attributes["host"]
	if host.Marinate == nil || host.Marinate.Description != "User-written host description" {
		t.Errorf("expected host description to be preserved, got %v", host.Marinate)
	}
}

// TestWriter_FieldNamedDescription tests that we can write and read
// a schema with a field named "description" without conflicts.
func TestWriter_FieldNamedDescription(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a schema with a node that has an attribute named "description"
	originalSchema := &schema.Schema{
		Variable: "ssh_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"ssh_authorized_key": {
				Marinate: &schema.MarinateInfo{
					Description: "SSH authorized keys configuration",
				},
				Attributes: map[string]*schema.Node{
					"description": {
						Marinate: &schema.MarinateInfo{
							Description: "Description of the SSH key",
							Type:        "string",
							Required:    false,
						},
						Attributes: map[string]*schema.Node{},
					},
					"key": {
						Marinate: &schema.MarinateInfo{
							Description: "The SSH public key",
							Type:        "string",
							Required:    true,
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	// Write the schema
	writer := yamlio.NewWriter(tmpDir)
	if err := writer.WriteSchema(originalSchema); err != nil {
		t.Fatalf("WriteSchema() error = %v", err)
	}

	// Read the YAML file content for inspection
	yamlPath := filepath.Join(tmpDir, "variables", "ssh_config.yaml")
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("failed to read YAML file: %v", err)
	}
	t.Logf("Generated YAML:\n%s", string(content))

	// Read it back
	reader := yamlio.NewReader(tmpDir)
	readSchema, err := reader.ReadSchema("ssh_config")
	if err != nil {
		t.Fatalf("ReadSchema() error = %v", err)
	}

	// Verify the schema structure
	sshKey := readSchema.SchemaNodes["ssh_authorized_key"]
	if sshKey == nil {
		t.Fatal("expected ssh_authorized_key node")
	}

	// Verify the "description" attribute exists
	desc, ok := sshKey.Attributes["description"]
	if !ok {
		t.Fatal("expected 'description' attribute")
	}
	if desc.Marinate.Type != "string" {
		t.Errorf("expected type 'string', got %v", desc.Marinate.Type)
	}
	if desc.Marinate == nil || desc.Marinate.Description != "Description of the SSH key" {
		t.Errorf("expected description preserved, got %v", desc.Marinate)
	}

	// Verify the "key" attribute exists
	key, ok := sshKey.Attributes["key"]
	if !ok {
		t.Fatal("expected 'key' attribute")
	}
	if !key.Marinate.Required {
		t.Error("expected key to be required")
	}
}
