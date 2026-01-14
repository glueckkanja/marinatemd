package schema

import (
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/hclparse"
)

func TestBuildFromHCL_SimpleTypes(t *testing.T) {
	tests := []struct {
		name     string
		variable *hclparse.Variable
		want     *Schema
	}{
		{
			name: "simple string type",
			variable: &hclparse.Variable{
				Name:        "app_name",
				Type:        "string",
				Description: "<!-- MARINATED: app_name -->",
				MarinatedID: "app_name",
			},
			want: &Schema{
				Variable:    "app_name",
				Version:     "1",
				SchemaNodes: map[string]*Node{},
			},
		},
		{
			name: "list of strings",
			variable: &hclparse.Variable{
				Name:        "tags",
				Type:        "list(string)",
				Description: "<!-- MARINATED: tags -->",
				MarinatedID: "tags",
			},
			want: &Schema{
				Variable: "tags",
				Version:  "1",
				SchemaNodes: map[string]*Node{
					"_root": {
						Type:        "list",
						ElementType: "string",
						Required:    true,
						Meta: &MetaInfo{
							Description: "# TODO: Add description for tags",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			got, err := b.BuildFromVariable(tt.variable)
			if err != nil {
				t.Fatalf("BuildFromVariable() error = %v", err)
			}

			if got.Variable != tt.want.Variable {
				t.Errorf("Variable = %v, want %v", got.Variable, tt.want.Variable)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.want.Version)
			}
		})
	}
}

func TestBuildFromHCL_ComplexObject(t *testing.T) {
	variable := &hclparse.Variable{
		Name: "app_config",
		Type: `object({
    database = optional(object({
      host     = string
      port     = optional(number, 5432)
      ssl_mode = optional(string, "require")
    }))
    cache = optional(object({
      redis_url = string
      ttl       = optional(number, 3600)
    }))
  })`,
		Description: "<!-- MARINATED: app_config -->",
		MarinatedID: "app_config",
	}

	b := NewBuilder()
	schema, err := b.BuildFromVariable(variable)
	if err != nil {
		t.Fatalf("BuildFromVariable() error = %v", err)
	}

	if schema.Variable != "app_config" {
		t.Errorf("Variable = %v, want app_config", schema.Variable)
	}

	// Check that database node exists
	database, ok := schema.SchemaNodes["database"]
	if !ok {
		t.Fatal("expected 'database' node in schema")
	}

	// Database should be optional (required: false) and type: object
	if database.Required {
		t.Error("expected database to be optional (required: false)")
	}
	if database.Type != "object" {
		t.Errorf("expected database type to be 'object', got %v", database.Type)
	}

	// Database should have _meta
	if database.Meta == nil {
		t.Error("expected database to have _meta")
	}

	// Database should have children (host, port, ssl_mode)
	if database.Children == nil || len(database.Children) == 0 {
		t.Fatal("expected database to have children")
	}

	// Check host field
	host, ok := database.Children["host"]
	if !ok {
		t.Fatal("expected 'host' field in database")
	}
	if !host.Required {
		t.Error("expected host to be required")
	}
	if host.Type != "string" {
		t.Errorf("expected host type to be 'string', got %v", host.Type)
	}

	// Check port field
	port, ok := database.Children["port"]
	if !ok {
		t.Fatal("expected 'port' field in database")
	}
	if port.Required {
		t.Error("expected port to be optional (required: false)")
	}
	if port.Type != "number" {
		t.Errorf("expected port type to be 'number', got %v", port.Type)
	}
}

func TestBuildFromHCL_MapOfObjects(t *testing.T) {
	variable := &hclparse.Variable{
		Name: "local_user",
		Type: `map(object({
    name            = string
    ssh_key_enabled = optional(bool)
  }))`,
		Description: "<!-- MARINATED: local_user -->",
		MarinatedID: "local_user",
	}

	b := NewBuilder()
	schema, err := b.BuildFromVariable(variable)
	if err != nil {
		t.Fatalf("BuildFromVariable() error = %v", err)
	}

	if schema.Variable != "local_user" {
		t.Errorf("Variable = %v, want local_user", schema.Variable)
	}

	// The map itself should be represented somehow
	// We expect the object structure to be documented as children
	if len(schema.SchemaNodes) == 0 {
		t.Error("expected schema nodes to be populated")
	}
}

func TestBuildFromHCL_NestedOptionalObjects(t *testing.T) {
	variable := &hclparse.Variable{
		Name: "network_rules",
		Type: `object({
    private_link_access = optional(list(object({
      endpoint_resource_id = string
      endpoint_tenant_id   = optional(string)
    })))
  })`,
		Description: "<!-- MARINATED: network_rules -->",
		MarinatedID: "network_rules",
	}

	b := NewBuilder()
	schema, err := b.BuildFromVariable(variable)
	if err != nil {
		t.Fatalf("BuildFromVariable() error = %v", err)
	}

	// Check private_link_access exists
	pla, ok := schema.SchemaNodes["private_link_access"]
	if !ok {
		t.Fatal("expected 'private_link_access' node")
	}

	// Should be optional list
	if pla.Required {
		t.Error("expected private_link_access to be optional")
	}
	if pla.Type != "list" {
		t.Errorf("expected type 'list', got %v", pla.Type)
	}
	if pla.ElementType != "object" {
		t.Errorf("expected element_type 'object', got %v", pla.ElementType)
	}

	// Should have children for the object structure
	if pla.Children == nil || len(pla.Children) == 0 {
		t.Fatal("expected private_link_access to have children describing object structure")
	}

	// Check endpoint_resource_id
	eri, ok := pla.Children["endpoint_resource_id"]
	if !ok {
		t.Fatal("expected 'endpoint_resource_id' in children")
	}
	if !eri.Required {
		t.Error("expected endpoint_resource_id to be required")
	}
	if eri.Type != "string" {
		t.Errorf("expected type 'string', got %v", eri.Type)
	}

	// Check endpoint_tenant_id
	eti, ok := pla.Children["endpoint_tenant_id"]
	if !ok {
		t.Fatal("expected 'endpoint_tenant_id' in children")
	}
	if eti.Required {
		t.Error("expected endpoint_tenant_id to be optional")
	}
	if eti.Type != "string" {
		t.Errorf("expected type 'string', got %v", eti.Type)
	}
}

func TestMergeWithExisting_PreserveDescriptions(t *testing.T) {
	// Existing schema with user descriptions
	existing := &Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*Node{
			"database": {
				Type:     "object",
				Required: false,
				Meta: &MetaInfo{
					Description: "User-written description for database",
				},
				Children: map[string]*Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "The database hostname",
					},
					"port": {
						Type:        "number",
						Required:    false,
						Description: "The database port number",
					},
				},
			},
		},
	}

	// New schema from updated HCL (same structure)
	newSchema := &Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*Node{
			"database": {
				Type:     "object",
				Required: false,
				Meta: &MetaInfo{
					Description: "# TODO: Add description for database",
				},
				Children: map[string]*Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "# TODO: Add description for host",
					},
					"port": {
						Type:        "number",
						Required:    false,
						Description: "# TODO: Add description for port",
					},
				},
			},
		},
	}

	b := NewBuilder()
	merged, err := b.MergeWithExisting(newSchema, existing)
	if err != nil {
		t.Fatalf("MergeWithExisting() error = %v", err)
	}

	// Check that user descriptions are preserved
	db := merged.SchemaNodes["database"]
	if db.Meta.Description != "User-written description for database" {
		t.Errorf("expected user description to be preserved, got %v", db.Meta.Description)
	}

	host := db.Children["host"]
	if host.Description != "The database hostname" {
		t.Errorf("expected user description to be preserved, got %v", host.Description)
	}
}

func TestMergeWithExisting_AddNewFields(t *testing.T) {
	// Existing schema
	existing := &Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*Node{
			"database": {
				Type:     "object",
				Required: false,
				Children: map[string]*Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "The database hostname",
					},
				},
			},
		},
	}

	// New schema with additional field
	newSchema := &Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*Node{
			"database": {
				Type:     "object",
				Required: false,
				Children: map[string]*Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "# TODO: Add description for host",
					},
					"port": {
						Type:        "number",
						Required:    false,
						Description: "# TODO: Add description for port",
					},
				},
			},
		},
	}

	b := NewBuilder()
	merged, err := b.MergeWithExisting(newSchema, existing)
	if err != nil {
		t.Fatalf("MergeWithExisting() error = %v", err)
	}

	// Check that new field is added
	db := merged.SchemaNodes["database"]
	if _, ok := db.Children["port"]; !ok {
		t.Error("expected new 'port' field to be added")
	}

	// Check that existing description is preserved
	if db.Children["host"].Description != "The database hostname" {
		t.Error("expected existing description to be preserved")
	}
}

func TestMergeWithExisting_RemoveDeletedFields(t *testing.T) {
	// Existing schema with a field that will be removed
	existing := &Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*Node{
			"database": {
				Type:     "object",
				Required: false,
				Children: map[string]*Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "The database hostname",
					},
					"old_field": {
						Type:        "string",
						Required:    false,
						Description: "This field no longer exists in HCL",
					},
				},
			},
		},
	}

	// New schema without old_field
	newSchema := &Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*Node{
			"database": {
				Type:     "object",
				Required: false,
				Children: map[string]*Node{
					"host": {
						Type:        "string",
						Required:    true,
						Description: "# TODO: Add description for host",
					},
				},
			},
		},
	}

	b := NewBuilder()
	merged, err := b.MergeWithExisting(newSchema, existing)
	if err != nil {
		t.Fatalf("MergeWithExisting() error = %v", err)
	}

	// Check that old field is removed
	db := merged.SchemaNodes["database"]
	if _, ok := db.Children["old_field"]; ok {
		t.Error("expected 'old_field' to be removed")
	}

	// Check that existing field is preserved
	if _, ok := db.Children["host"]; !ok {
		t.Error("expected 'host' field to be preserved")
	}
}
