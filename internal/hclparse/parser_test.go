package hclparse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glueckkanja/marinatemd/internal/hclparse"
)

// Helper function to set up test parser with HCL content.
func setupTestParser(t *testing.T, hclContent string) (*hclparse.Parser, error) {
	t.Helper()
	tmpDir := t.TempDir()
	hclFile := filepath.Join(tmpDir, "variables.tf")
	if err := os.WriteFile(hclFile, []byte(hclContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	p := hclparse.NewParser()
	err := p.ParseVariables(tmpDir)
	return p, err
}

func TestParser_SimpleStringVariable(t *testing.T) {
	hclContent := `
variable "app_name" {
  type        = string
  description = "<!-- MARINATED: app_name --> The application name"
}
`
	p, err := setupTestParser(t, hclContent)
	if err != nil {
		t.Fatalf("ParseVariables() error = %v", err)
	}

	vars, err := p.ExtractMarinatedVars()
	if err != nil {
		t.Fatalf("ExtractMarinatedVars() error = %v", err)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(vars))
	}

	v := vars[0]
	if v.Name != "app_name" {
		t.Errorf("expected name 'app_name', got '%s'", v.Name)
	}
	if !v.Marinated {
		t.Error("expected variable to be marked as Marinated")
	}
	if v.MarinatedID != "app_name" {
		t.Errorf("expected MarinatedID 'app_name', got '%s'", v.MarinatedID)
	}
	if v.Description != "<!-- MARINATED: app_name --> The application name" {
		t.Errorf("expected full description to be preserved, got '%s'", v.Description)
	}
}

func TestParser_ComplexObjectWithOptional(t *testing.T) {
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
  description = "<!-- MARINATED: app_config -->"
}
`
	p, err := setupTestParser(t, hclContent)
	if err != nil {
		t.Fatalf("ParseVariables() error = %v", err)
	}

	vars, err := p.ExtractMarinatedVars()
	if err != nil {
		t.Fatalf("ExtractMarinatedVars() error = %v", err)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(vars))
	}

	v := vars[0]
	if v.Name != "app_config" {
		t.Errorf("expected name 'app_config', got '%s'", v.Name)
	}
	if v.MarinatedID != "app_config" {
		t.Errorf("expected MarinatedID 'app_config', got '%s'", v.MarinatedID)
	}
	if v.Type == "" {
		t.Error("expected Type to be set")
	}
}

func TestParser_MapOfObjects(t *testing.T) {
	hclContent := `
variable "local_user" {
  type = map(object({
    name                 = string
    ssh_key_enabled      = optional(bool)
    permission_scope = optional(list(object({
      resource_name = string
      service       = string
      permissions = object({
        create = optional(bool)
        delete = optional(bool)
        read   = optional(bool)
      })
    })))
  }))
  description = "<!-- MARINATED: local_user --> Local user configuration"
}
`
	p, err := setupTestParser(t, hclContent)
	if err != nil {
		t.Fatalf("ParseVariables() error = %v", err)
	}

	vars, err := p.ExtractMarinatedVars()
	if err != nil {
		t.Fatalf("ExtractMarinatedVars() error = %v", err)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(vars))
	}

	v := vars[0]
	if v.Name != "local_user" {
		t.Errorf("expected name 'local_user', got '%s'", v.Name)
	}
	if v.Type == "" {
		t.Error("expected Type to be set")
	}
}

func TestParser_MultipleVariablesMixed(t *testing.T) {
	hclContent := `
variable "plain_var" {
  type        = string
  description = "Not marinated"
}

variable "marinated_one" {
  type        = string
  description = "<!-- MARINATED: marinated_one -->"
}

variable "another_plain" {
  type        = number
  description = "Also not marinated"
}

variable "marinated_two" {
  type = object({
    field = string
  })
  description = "Some text <!-- MARINATED: marinated_two --> more text"
}
`
	p, err := setupTestParser(t, hclContent)
	if err != nil {
		t.Fatalf("ParseVariables() error = %v", err)
	}

	vars, err := p.ExtractMarinatedVars()
	if err != nil {
		t.Fatalf("ExtractMarinatedVars() error = %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 marinated variables, got %d", len(vars))
	}

	names := make(map[string]bool)
	for _, v := range vars {
		names[v.Name] = true
	}
	if !names["marinated_one"] || !names["marinated_two"] {
		t.Error("expected marinated_one and marinated_two to be extracted")
	}
}

func TestParser_ListAndSetTypes(t *testing.T) {
	hclContent := `
variable "tags" {
  type        = list(string)
  description = "<!-- MARINATED: tags -->"
}

variable "allowed_ips" {
  type        = set(string)
  description = "<!-- MARINATED: allowed_ips -->"
}
`
	p, err := setupTestParser(t, hclContent)
	if err != nil {
		t.Fatalf("ParseVariables() error = %v", err)
	}

	vars, err := p.ExtractMarinatedVars()
	if err != nil {
		t.Fatalf("ExtractMarinatedVars() error = %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(vars))
	}
}

func TestParser_InvalidHCLSyntax(t *testing.T) {
	hclContent := `
variable "broken" {
  type = this is not valid HCL
  description = "<!-- MARINATED: broken -->"
}
`
	_, err := setupTestParser(t, hclContent)
	if err == nil {
		t.Error("ParseVariables() expected error for invalid HCL, got nil")
	}
}
