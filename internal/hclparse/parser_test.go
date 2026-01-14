package hclparse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_ParseVariables(t *testing.T) {
	tests := []struct {
		name       string
		hclContent string
		wantErr    bool
		validate   func(t *testing.T, p *Parser)
	}{
		{
			name: "simple string variable with MARINATED marker",
			hclContent: `
variable "app_name" {
  type        = string
  description = "<!-- MARINATED: app_name --> The application name"
}
`,
			wantErr: false,
			validate: func(t *testing.T, p *Parser) {
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
			},
		},
		{
			name: "complex object with optional nested fields",
			hclContent: `
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
`,
			wantErr: false,
			validate: func(t *testing.T, p *Parser) {
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
				// Type should be preserved as the full HCL expression
				if v.Type == "" {
					t.Error("expected Type to be set")
				}
			},
		},
		{
			name: "map of objects",
			hclContent: `
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
`,
			wantErr: false,
			validate: func(t *testing.T, p *Parser) {
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
			},
		},
		{
			name: "multiple variables with mixed MARINATED markers",
			hclContent: `
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
`,
			wantErr: false,
			validate: func(t *testing.T, p *Parser) {
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
			},
		},
		{
			name: "list and set types",
			hclContent: `
variable "tags" {
  type        = list(string)
  description = "<!-- MARINATED: tags -->"
}

variable "allowed_ips" {
  type        = set(string)
  description = "<!-- MARINATED: allowed_ips -->"
}
`,
			wantErr: false,
			validate: func(t *testing.T, p *Parser) {
				vars, err := p.ExtractMarinatedVars()
				if err != nil {
					t.Fatalf("ExtractMarinatedVars() error = %v", err)
				}
				if len(vars) != 2 {
					t.Fatalf("expected 2 variables, got %d", len(vars))
				}
			},
		},
		{
			name: "invalid HCL syntax",
			hclContent: `
variable "broken" {
  type = this is not valid HCL
  description = "<!-- MARINATED: broken -->"
}
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory with test HCL file
			tmpDir := t.TempDir()
			hclFile := filepath.Join(tmpDir, "variables.tf")
			if err := os.WriteFile(hclFile, []byte(tt.hclContent), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			p := NewParser()
			err := p.ParseVariables(tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, p)
			}
		})
	}
}

func TestParser_ParseVariables_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple variables.*.tf files
	files := map[string]string{
		"variables.tf": `
variable "var_one" {
  type        = string
  description = "<!-- MARINATED: var_one -->"
}
`,
		"variables.network.tf": `
variable "var_two" {
  type        = string
  description = "<!-- MARINATED: var_two -->"
}
`,
		"variables.storage.tf": `
variable "var_three" {
  type        = string
  description = "<!-- MARINATED: var_three -->"
}
`,
		"main.tf": `
# This file should be ignored
resource "null_resource" "test" {}
`,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", filename, err)
		}
	}

	p := NewParser()
	if err := p.ParseVariables(tmpDir); err != nil {
		t.Fatalf("ParseVariables() error = %v", err)
	}

	vars, err := p.ExtractMarinatedVars()
	if err != nil {
		t.Fatalf("ExtractMarinatedVars() error = %v", err)
	}

	if len(vars) != 3 {
		t.Errorf("expected 3 variables from multiple files, got %d", len(vars))
	}
}

func TestExtractMarinatedID(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantID      string
		wantFound   bool
	}{
		{
			name:        "exact marker",
			description: "<!-- MARINATED: app_config -->",
			wantID:      "app_config",
			wantFound:   true,
		},
		{
			name:        "marker with prefix text",
			description: "Database config <!-- MARINATED: app_config -->",
			wantID:      "app_config",
			wantFound:   true,
		},
		{
			name:        "marker with suffix text",
			description: "<!-- MARINATED: app_config --> for the application",
			wantID:      "app_config",
			wantFound:   true,
		},
		{
			name:        "marker with surrounding text",
			description: "This is <!-- MARINATED: app_config --> a config",
			wantID:      "app_config",
			wantFound:   true,
		},
		{
			name:        "no marker",
			description: "Just a regular description",
			wantID:      "",
			wantFound:   false,
		},
		{
			name:        "malformed marker - no ID",
			description: "<!-- MARINATED: -->",
			wantID:      "",
			wantFound:   false,
		},
		{
			name:        "marker with underscores and numbers",
			description: "<!-- MARINATED: config_v2_final -->",
			wantID:      "config_v2_final",
			wantFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotFound := extractMarinatedID(tt.description)
			if gotID != tt.wantID {
				t.Errorf("extractMarinatedID() gotID = %v, want %v", gotID, tt.wantID)
			}
			if gotFound != tt.wantFound {
				t.Errorf("extractMarinatedID() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}
