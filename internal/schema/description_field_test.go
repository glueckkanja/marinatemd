package schema_test

import (
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"github.com/c4a8-azure/marinatemd/internal/schema"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
)

// TestBuildFromHCL_FieldNamedDescription tests that we can handle
// an attribute named "description" without conflicts.
func TestBuildFromHCL_FieldNamedDescription(t *testing.T) {
	// This reproduces the bug from the issue report
	// The ssh_authorized_key object has a field named "description"
	variable := &hclparse.Variable{
		Name: "ssh_config",
		Type: `object({
			ssh_authorized_key = optional(list(object({
				description = optional(string)
				key         = string
			})))
		})`,
		Description: "<!-- MARINATED: ssh_config -->",
		MarinatedID: "ssh_config",
	}

	b := schema.NewBuilder()
	s, err := b.BuildFromVariable(variable)
	if err != nil {
		t.Fatalf("BuildFromVariable() error = %v", err)
	}

	if s.Variable != "ssh_config" {
		t.Errorf("Variable = %v, want ssh_config", s.Variable)
	}

	// Check that ssh_authorized_key node exists
	sshKey, ok := s.SchemaNodes["ssh_authorized_key"]
	if !ok {
		t.Fatal("expected 'ssh_authorized_key' node in schema")
	}

	// It should be a list of objects
	if sshKey.Type != "list" {
		t.Errorf("expected type 'list', got %v", sshKey.Type)
	}
	if sshKey.ElementType != "object" {
		t.Errorf("expected element_type 'object', got %v", sshKey.ElementType)
	}

	// Check that it has the "description" child
	description, ok := sshKey.Children["description"]
	if !ok {
		t.Fatal("expected 'description' field in ssh_authorized_key children")
	}

	// description field should be optional string
	if description.Required {
		t.Error("expected description to be optional")
	}
	if description.Type != "string" {
		t.Errorf("expected type 'string', got %v", description.Type)
	}

	// Check that it has the "key" child
	key, ok := sshKey.Children["key"]
	if !ok {
		t.Fatal("expected 'key' field in ssh_authorized_key children")
	}
	if !key.Required {
		t.Error("expected key to be required")
	}
	if key.Type != "string" {
		t.Errorf("expected type 'string', got %v", key.Type)
	}
}

// TestYAMLMarshalUnmarshal_FieldNamedDescription tests that we can
// marshal and unmarshal a schema with a field named "description".
func TestYAMLMarshalUnmarshal_FieldNamedDescription(t *testing.T) {
	// Create a schema with a node that has a child named "description"
	original := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"ssh_key": {
				Type:        "list",
				ElementType: "object",
				Required:    false,
				Description: "List of SSH keys",
				Children: map[string]*schema.Node{
					"description": {
						Type:        "string",
						Required:    false,
						Description: "Description of the SSH key",
					},
					"key": {
						Type:        "string",
						Required:    true,
						Description: "The SSH key value",
					},
				},
			},
		},
	}

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Write the schema
	writer := yamlio.NewWriter(tmpDir)
	err := writer.WriteSchema(original)
	if err != nil {
		t.Fatalf("WriteSchema() error = %v", err)
	}

	// Read it back
	reader := yamlio.NewReader(tmpDir)
	read, err := reader.ReadSchema("test_var")
	if err != nil {
		t.Fatalf("ReadSchema() error = %v", err)
	}

	// Verify the schema is correct
	if read.Variable != original.Variable {
		t.Errorf("Variable = %v, want %v", read.Variable, original.Variable)
	}

	sshKey, ok := read.SchemaNodes["ssh_key"]
	if !ok {
		t.Fatal("expected 'ssh_key' node")
	}

	// Verify the "description" child exists and is correct
	desc, ok := sshKey.Children["description"]
	if !ok {
		t.Fatal("expected 'description' child")
	}
	if desc.Type != "string" {
		t.Errorf("expected type 'string', got %v", desc.Type)
	}
	if desc.Description != "Description of the SSH key" {
		t.Errorf("expected description preserved, got %v", desc.Description)
	}

	// Verify the "key" child exists
	key, ok := sshKey.Children["key"]
	if !ok {
		t.Fatal("expected 'key' child")
	}
	if !key.Required {
		t.Error("expected key to be required")
	}
}
