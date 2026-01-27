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
	if sshKey.Marinate.Type != "list" {
		t.Errorf("expected type 'list', got %v", sshKey.Marinate.Type)
	}
	if sshKey.Marinate.ElementType != "object" {
		t.Errorf("expected element_type 'object', got %v", sshKey.Marinate.ElementType)
	}

	// Check that it has the "description" attribute
	description, ok := sshKey.Attributes["description"]
	if !ok {
		t.Fatal("expected 'description' field in ssh_authorized_key attributes")
	}

	// description field should be optional string
	if description.Marinate.Required {
		t.Error("expected description to be optional")
	}
	if description.Marinate.Type != "string" {
		t.Errorf("expected type 'string', got %v", description.Marinate.Type)
	}

	// Check that it has the "key" attribute
	key, ok := sshKey.Attributes["key"]
	if !ok {
		t.Fatal("expected 'key' field in ssh_authorized_key attributes")
	}
	if !key.Marinate.Required {
		t.Error("expected key to be required")
	}
	if key.Marinate.Type != "string" {
		t.Errorf("expected type 'string', got %v", key.Marinate.Type)
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
				Marinate: &schema.MarinateInfo{
					Type:        "list",
					ElementType: "object",
					Required:    false,
					Description: "List of SSH keys",
				},
				Attributes: map[string]*schema.Node{
					"description": {
						Marinate: &schema.MarinateInfo{
							Type:        "string",
							Required:    false,
							Description: "Description of the SSH key",
						},
						Attributes: map[string]*schema.Node{},
					},
					"key": {
						Marinate: &schema.MarinateInfo{
							Type:        "string",
							Required:    true,
							Description: "The SSH key value",
						},
						Attributes: map[string]*schema.Node{},
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

	// Verify the "description" attribute exists and is correct
	desc, ok := sshKey.Attributes["description"]
	if !ok {
		t.Fatal("expected 'description' attribute")
	}
	if desc.Marinate.Type != "string" {
		t.Errorf("expected type 'string', got %v", desc.Marinate.Type)
	}
	if desc.Marinate.Description != "Description of the SSH key" {
		t.Errorf("expected description preserved, got %v", desc.Marinate.Description)
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
