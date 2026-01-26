package schema_test

import (
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"github.com/c4a8-azure/marinatemd/internal/schema"
)

func TestBuildFromHCL_SimpleTypes(t *testing.T) {
	tests := []struct {
		name     string
		variable *hclparse.Variable
		want     *schema.Schema
	}{
		{
			name: "simple string type",
			variable: &hclparse.Variable{
				Name:        "app_name",
				Description: "<!-- MARINATED: app_name -->",
				MarinatedID: "app_name",
			},
			want: &schema.Schema{
				Variable:    "app_name",
				Version:     "1",
				SchemaNodes: map[string]*schema.Node{},
			},
		},
		{
			name: "list of strings",
			variable: &hclparse.Variable{
				Name:        "tags",
				Description: "<!-- MARINATED: tags -->",
				MarinatedID: "tags",
			},
			want: &schema.Schema{
				Variable: "tags",
				Version:  "1",
				SchemaNodes: map[string]*schema.Node{
					"_root": {
						Marinate: &schema.MarinateInfo{
							Description: "# TODO: Add description for tags",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := schema.NewBuilder()
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

	b := schema.NewBuilder()
	s, err := b.BuildFromVariable(variable)
	if err != nil {
		t.Fatalf("BuildFromVariable() error = %v", err)
	}

	if s.Variable != "app_config" {
		t.Errorf("Variable = %v, want app_config", s.Variable)
	}

	// Check that database node exists
	database, ok := s.SchemaNodes["database"]
	if !ok {
		t.Fatal("expected 'database' node in schema")
	}

	// Database should be optional (required: false) and type: object
	if database.Marinate.Required {
		t.Error("expected database to be optional (required: false)")
	}
	if database.Marinate.Type != "object" {
		t.Errorf("expected database type to be 'object', got %v", database.Marinate.Type)
	}

	// Database should have _meta
	if database.Marinate == nil {
		t.Error("expected database to have _marinate")
	}

	// Database should have attributes (host, port, ssl_mode)
	if len(database.Attributes) == 0 {
		t.Fatal("expected database to have attributes")
	}

	// Check host field
	host, ok := database.Attributes["host"]
	if !ok {
		t.Fatal("expected 'host' field in database")
	}
	if !host.Marinate.Required {
		t.Error("expected host to be required")
	}
	if host.Marinate.Type != "string" {
		t.Errorf("expected host type to be 'string', got %v", host.Marinate.Type)
	}

	// Check port field
	port, ok := database.Attributes["port"]
	if !ok {
		t.Fatal("expected 'port' field in database")
	}
	if port.Marinate.Required {
		t.Error("expected port to be optional (required: false)")
	}
	if port.Marinate.Type != "number" {
		t.Errorf("expected port type to be 'number', got %v", port.Marinate.Type)
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

	b := schema.NewBuilder()
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

	b := schema.NewBuilder()
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
	if pla.Marinate.Required {
		t.Error("expected private_link_access to be optional")
	}
	if pla.Marinate.Type != "list" {
		t.Errorf("expected type 'list', got %v", pla.Marinate.Type)
	}
	if pla.Marinate.ElementType != "object" {
		t.Errorf("expected element_type 'object', got %v", pla.Marinate.ElementType)
	}

	// Should have attributes for the object structure
	if len(pla.Attributes) == 0 {
		t.Fatal("expected private_link_access to have attributes describing object structure")
	}

	// Check endpoint_resource_id
	eri, ok := pla.Attributes["endpoint_resource_id"]
	if !ok {
		t.Fatal("expected 'endpoint_resource_id' in children")
	}
	if !eri.Marinate.Required {
		t.Error("expected endpoint_resource_id to be required")
	}
	if eri.Marinate.Type != "string" {
		t.Errorf("expected type 'string', got %v", eri.Marinate.Type)
	}

	// Check endpoint_tenant_id
	eti, ok := pla.Attributes["endpoint_tenant_id"]
	if !ok {
		t.Fatal("expected 'endpoint_tenant_id' in children")
	}
	if eti.Marinate.Required {
		t.Error("expected endpoint_tenant_id to be optional")
	}
	if eti.Marinate.Type != "string" {
		t.Errorf("expected type 'string', got %v", eti.Marinate.Type)
	}
}

func TestMergeWithExisting_PreserveDescriptions(t *testing.T) {
	// Existing schema with user descriptions
	existing := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Marinate: &schema.MarinateInfo{
					Description: "User-written description for database",
				},
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "The database hostname",
						},
						Attributes: map[string]*schema.Node{},
					},
					"port": {
						Marinate: &schema.MarinateInfo{
							Description: "The database port number",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	// New schema from updated HCL (same structure)
	newSchema := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Marinate: &schema.MarinateInfo{
					Description: "# TODO: Add description for database",
				},
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "# TODO: Add description for host",
						},
						Attributes: map[string]*schema.Node{},
					},
					"port": {
						Marinate: &schema.MarinateInfo{
							Description: "# TODO: Add description for port",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	b := schema.NewBuilder()
	merged, err := b.MergeWithExisting(newSchema, existing)
	if err != nil {
		t.Fatalf("MergeWithExisting() error = %v", err)
	}

	// Check that user descriptions are preserved
	db := merged.SchemaNodes["database"]
	if db.Marinate == nil || db.Marinate.Description != "User-written description for database" {
		if db.Marinate == nil {
			t.Error("expected database to have Marinate")
		} else {
			t.Errorf("expected user description to be preserved, got %v", db.Marinate.Description)
		}
	}

	host := db.Attributes["host"]
	if host.Marinate == nil || host.Marinate.Description != "The database hostname" {
		if host.Marinate == nil {
			t.Error("expected host to have Marinate")
		} else {
			t.Errorf("expected user description to be preserved, got %v", host.Marinate.Description)
		}
	}
}

func TestMergeWithExisting_AddNewFields(t *testing.T) {
	// Existing schema
	existing := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "The database hostname",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	// New schema with additional field
	newSchema := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "# TODO: Add description for host",
						},
						Attributes: map[string]*schema.Node{},
					},
					"port": {
						Marinate: &schema.MarinateInfo{
							Description: "# TODO: Add description for port",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	b := schema.NewBuilder()
	merged, err := b.MergeWithExisting(newSchema, existing)
	if err != nil {
		t.Fatalf("MergeWithExisting() error = %v", err)
	}

	// Check that new field is added
	db := merged.SchemaNodes["database"]
	if _, ok := db.Attributes["port"]; !ok {
		t.Error("expected new 'port' field to be added")
	}

	// Check that existing description is preserved
	if db.Attributes["host"].Marinate == nil || db.Attributes["host"].Marinate.Description != "The database hostname" {
		t.Error("expected existing description to be preserved")
	}
}

func TestMergeWithExisting_RemoveDeletedFields(t *testing.T) {
	// Existing schema with a field that will be removed
	existing := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "The database hostname",
						},
						Attributes: map[string]*schema.Node{},
					},
					"old_field": {
						Marinate: &schema.MarinateInfo{
							Description: "This field no longer exists in HCL",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	// New schema without old_field
	newSchema := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Attributes: map[string]*schema.Node{
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "# TODO: Add description for host",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	b := schema.NewBuilder()
	merged, err := b.MergeWithExisting(newSchema, existing)
	if err != nil {
		t.Fatalf("MergeWithExisting() error = %v", err)
	}

	// Check that old field is removed
	db := merged.SchemaNodes["database"]
	if _, ok := db.Attributes["old_field"]; ok {
		t.Error("expected 'old_field' to be removed")
	}

	// Check that existing field is preserved
	if _, ok := db.Attributes["host"]; !ok {
		t.Error("expected 'host' field to be preserved")
	}
}
