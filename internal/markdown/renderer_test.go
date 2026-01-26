package markdown //nolint:testpackage // tests need access to unexported types

import (
	"strings"
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/schema"
)

func TestNewRenderer(t *testing.T) {
	r := NewRenderer()
	if r == nil {
		t.Fatal("Expected renderer, got nil")
	}
	if r.templateCfg == nil {
		t.Fatal("Expected template config, got nil")
	}
}

func TestNewRendererWithTemplate(t *testing.T) {
	customCfg := &TemplateConfig{
		AttributeTemplate: "**{attribute}**: {description}",
		RequiredText:      "Mandatory",
		OptionalText:      "Optional",
		EscapeMode:        "bold",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
	}

	r := NewRendererWithTemplate(customCfg)
	if r == nil {
		t.Fatal("Expected renderer, got nil")
	}
	if r.templateCfg != customCfg {
		t.Fatal("Expected custom template config")
	}

	// Test with nil config - should use default
	r2 := NewRendererWithTemplate(nil)
	if r2.templateCfg == nil {
		t.Fatal("Expected default template config when nil is passed")
	}
}

func TestRenderSchema_SimpleAttributes(t *testing.T) {
	s := &schema.Schema{
		Variable: "network_rules",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"bypass": {
				Marinate: &schema.MarinateInfo{
					Description: "Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.",
					Required:    false,
				},
				Attributes: map[string]*schema.Node{},
			},
			"default_action": {
				Marinate: &schema.MarinateInfo{
					Description: "Specifies the default action of allow or deny when no other rules match.",
					Required:    true,
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that both attributes are present
	if !strings.Contains(result, "`bypass`") {
		t.Error("Expected bypass attribute in output")
	}
	if !strings.Contains(result, "(Optional)") {
		t.Error("Expected Optional marker for bypass")
	}
	if !strings.Contains(result, "`default_action`") {
		t.Error("Expected default_action attribute in output")
	}
	if !strings.Contains(result, "(Required)") {
		t.Error("Expected Required marker for default_action")
	}
}

func TestRenderSchema_NestedObjects(t *testing.T) {
	s := &schema.Schema{
		Variable: "app_config",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {Marinate: &schema.MarinateInfo{
				Description: "Database configuration settings",
			},
				Attributes: map[string]*schema.Node{
					"host": {Marinate: &schema.MarinateInfo{
						Description: "The database host address",
					},
						Attributes: map[string]*schema.Node{},
					},
					"port": {Marinate: &schema.MarinateInfo{
						Description: "The database port number",
					},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check hierarchical structure
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}

	// First line should be the parent object
	if !strings.Contains(lines[0], "`database`") {
		t.Error("Expected database as first item")
	}

	// Children should be indented
	if !strings.HasPrefix(lines[1], "  - ") {
		t.Error("Expected child to be indented with bullet")
	}
	if !strings.Contains(result, "`host`") {
		t.Error("Expected host attribute")
	}
	if !strings.Contains(result, "`port`") {
		t.Error("Expected port attribute")
	}
}

func TestRenderSchema_CustomTemplate(t *testing.T) {
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"field1": {
				Marinate: &schema.MarinateInfo{
					Description: "A test field",
					Required:    true,
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	customCfg := &TemplateConfig{
		AttributeTemplate: "**{attribute}** [{required}]: {description}",
		RequiredText:      "Mandatory",
		OptionalText:      "Optional",
		EscapeMode:        "bold",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
	}

	r := NewRendererWithTemplate(customCfg)
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "**field1**") {
		t.Error("Expected bold-formatted field1")
	}
	if !strings.Contains(result, "[Mandatory]") {
		t.Error("Expected custom required text 'Mandatory'")
	}
}

func TestRenderSchema_NilSchema(t *testing.T) {
	r := NewRenderer()
	_, err := r.RenderSchema(nil)
	if err == nil {
		t.Error("Expected error for nil schema")
	}
}

func TestRenderSchema_EmptySchema(t *testing.T) {
	s := &schema.Schema{
		Variable:    "empty",
		Version:     "1",
		SchemaNodes: map[string]*schema.Node{},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "" {
		t.Error("Expected empty output for schema with no nodes")
	}
}

func TestRenderSchema_DeterministicOrder(t *testing.T) {
	s := &schema.Schema{
		Variable: "test",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"zebra": {Marinate: &schema.MarinateInfo{
				Description: "Last alphabetically",
			},
				Attributes: map[string]*schema.Node{},
			},
			"alpha": {Marinate: &schema.MarinateInfo{
				Description: "First alphabetically",
			},
				Attributes: map[string]*schema.Node{},
			},
			"middle": {Marinate: &schema.MarinateInfo{
				Description: "Middle alphabetically",
			},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	r := NewRenderer()
	result1, err1 := r.RenderSchema(s)
	if err1 != nil {
		t.Fatalf("Unexpected error: %v", err1)
	}

	result2, err2 := r.RenderSchema(s)
	if err2 != nil {
		t.Fatalf("Unexpected error: %v", err2)
	}

	// Results should be identical
	if result1 != result2 {
		t.Error("Expected deterministic output across multiple renders")
	}

	// Check alphabetical order
	lines := strings.Split(strings.TrimSpace(result1), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "`alpha`") {
		t.Error("Expected alpha first")
	}
	if !strings.Contains(lines[1], "`middle`") {
		t.Error("Expected middle second")
	}
	if !strings.Contains(lines[2], "`zebra`") {
		t.Error("Expected zebra last")
	}
}

// TODO: Add tests for injecting content into docs
func TestInjectIntoFile(t *testing.T) {
	t.Skip("Not implemented yet")
}

// TODO: Add tests for finding MARINATED markers
func TestFindMarkers(t *testing.T) {
	t.Skip("Not implemented yet")
}

func TestRenderSchema_ShowDescriptionDefault(t *testing.T) {
	// When ShowDescription is nil (omitted), description should be shown by default
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"field1": {
				Marinate: &schema.MarinateInfo{
					Description:     "This description should be visible",
					ShowDescription: nil, // Omitted - defaults to true
					Required:        true,
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Description should be present in output
	if !strings.Contains(result, "This description should be visible") {
		t.Error("Expected description to be visible when ShowDescription is nil (default)")
	}
	if !strings.Contains(result, "`field1`") {
		t.Error("Expected field1 attribute in output")
	}
}

func TestRenderSchema_ShowDescriptionExplicitlyFalse(t *testing.T) {
	// When ShowDescription is explicitly false, description should be hidden
	showDesc := false
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"field1": {
				Marinate: &schema.MarinateInfo{
					Description:     "This description should be hidden",
					ShowDescription: &showDesc,
					Required:        true,
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Description should NOT be present in output
	if strings.Contains(result, "This description should be hidden") {
		t.Error("Expected description to be hidden when ShowDescription is false")
	}
	// The attribute itself should still be rendered (with name and required status)
	if !strings.Contains(result, "`field1`") {
		t.Error("Expected field1 attribute to be rendered even when description is hidden")
	}
	if !strings.Contains(result, "(Required)") {
		t.Error("Expected Required marker to be present even when description is hidden")
	}
}

func TestRenderSchema_ShowDescriptionExplicitlyTrue(t *testing.T) {
	// When ShowDescription is explicitly true, description should be shown
	showDesc := true
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"field1": {
				Marinate: &schema.MarinateInfo{
					Description:     "This description should be visible",
					ShowDescription: &showDesc,
					Required:        false,
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Description should be present in output
	if !strings.Contains(result, "This description should be visible") {
		t.Error("Expected description to be visible when ShowDescription is true")
	}
	if !strings.Contains(result, "`field1`") {
		t.Error("Expected field1 attribute in output")
	}
}

func TestRenderSchema_ShowDescriptionWithNestedAttributes(t *testing.T) {
	// Test that hiding parent description doesn't hide children
	showDesc := false
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"parent": {
				Marinate: &schema.MarinateInfo{
					Description:     "Parent description should be hidden",
					ShowDescription: &showDesc,
				},
				Attributes: map[string]*schema.Node{
					"child": {
						Marinate: &schema.MarinateInfo{
							Description: "Child description should be visible",
							Required:    true,
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Parent description should be hidden
	if strings.Contains(result, "Parent description should be hidden") {
		t.Error("Expected parent description to be hidden")
	}

	// Parent attribute itself should still be rendered (for structure)
	// Note: Depending on template, parent may or may not appear if it has no description
	// But children should definitely be visible

	// Child description should be visible
	if !strings.Contains(result, "Child description should be visible") {
		t.Error("Expected child description to be visible")
	}
	if !strings.Contains(result, "`child`") {
		t.Error("Expected child attribute in output")
	}
}

func TestRenderSchema_ShowDescriptionWithDefaults(t *testing.T) {
	// Test that attribute with default value is rendered even without description
	showDesc := false
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"port": {
				Marinate: &schema.MarinateInfo{
					Description:     "Database port number",
					ShowDescription: &showDesc,
					Type:            "number",
					Required:        false,
					Default:         5432,
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Description should NOT be in output
	if strings.Contains(result, "Database port number") {
		t.Error("Expected description to be hidden")
	}

	// But the attribute should still be rendered
	if !strings.Contains(result, "`port`") {
		t.Error("Expected port attribute to be rendered")
	}
}

func TestRenderSchema_WithDefaultAndExample(t *testing.T) {
	// Test that default and example values are rendered correctly
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"database": {
				Marinate: &schema.MarinateInfo{
					Description: "Database configuration",
					Required:    false,
				},
				Attributes: map[string]*schema.Node{
					"port": {
						Marinate: &schema.MarinateInfo{
							Description: "Database port",
							Required:    false,
							Type:        "number",
							Default:     5432,
							Example:     3306,
						},
						Attributes: map[string]*schema.Node{},
					},
					"host": {
						Marinate: &schema.MarinateInfo{
							Description: "Database hostname",
							Required:    true,
							Type:        "string",
							Default:     "localhost",
							Example:     "db.example.com",
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	// Create a custom template that uses {default} and {example}
	templateCfg := &TemplateConfig{
		AttributeTemplate: "{attribute} - ({required}) {description} [Default: {default}] [Example: {example}]",
		RequiredText:      "Required",
		OptionalText:      "Optional",
		EscapeMode:        "inline_code",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
	}

	r := NewRendererWithTemplate(templateCfg)
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that default and example values appear in output
	if !strings.Contains(result, "[Default: 5432]") {
		t.Error("Expected default value 5432 for port")
	}
	if !strings.Contains(result, "[Example: 3306]") {
		t.Error("Expected example value 3306 for port")
	}
	if !strings.Contains(result, "[Default: localhost]") {
		t.Error("Expected default value 'localhost' for host")
	}
	if !strings.Contains(result, "[Example: db.example.com]") {
		t.Error("Expected example value 'db.example.com' for host")
	}
}

func TestRenderSchema_WithEmptyStringDefault(t *testing.T) {
	// Test that empty string default values are rendered correctly
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"config": {
				Marinate: &schema.MarinateInfo{
					Description: "Configuration object",
					Required:    false,
				},
				Attributes: map[string]*schema.Node{
					"prefix": {
						Marinate: &schema.MarinateInfo{
							Description: "Resource name prefix",
							Required:    false,
							Type:        "string",
							Default:     "", // Empty string is a valid default
						},
						Attributes: map[string]*schema.Node{},
					},
					"suffix": {
						Marinate: &schema.MarinateInfo{
							Description: "Resource name suffix",
							Required:    false,
							Type:        "string",
							Default:     "prod", // Non-empty default for comparison
						},
						Attributes: map[string]*schema.Node{},
					},
					"no_default": {
						Marinate: &schema.MarinateInfo{
							Description: "Field without default",
							Required:    true,
							Type:        "string",
							Default:     nil, // No default value
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	// Create a custom template that uses {default} conditionally
	templateCfg := &TemplateConfig{
		AttributeTemplate: "{{.Attribute}} - ({{.Required}}) {{.Description}}{{if .HasDefault}} [Default: {{.Default}}]{{end}}",
		RequiredText:      "Required",
		OptionalText:      "Optional",
		EscapeMode:        "inline_code",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
	}

	r := NewRendererWithTemplate(templateCfg)
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that empty string default is shown as ""
	if !strings.Contains(result, `[Default: ""]`) {
		t.Errorf("Expected empty string default shown as \"\" for prefix. Got:\n%s", result)
	}

	// Check that non-empty default is shown
	if !strings.Contains(result, `[Default: prod]`) {
		t.Errorf("Expected default value 'prod' for suffix. Got:\n%s", result)
	}

	// Check that field without default has no [Default: ...] text
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, "`no_default`") {
			if strings.Contains(line, "[Default:") {
				t.Errorf("Expected no default for no_default field. Got line: %s", line)
			}
			break
		}
	}
}

func TestRenderSchema_WithEmptyMapDefault(t *testing.T) {
	// Test that empty map/object defaults (map[]) are rendered as {}
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"config": {
				Marinate: &schema.MarinateInfo{
					Description: "Configuration settings",
					Required:    false,
				},
				Attributes: map[string]*schema.Node{
					"tags": {
						Marinate: &schema.MarinateInfo{
							Description: "Resource tags",
							Required:    false,
							Type:        "map(string)",
							Default:     map[string]interface{}{}, // Empty map
						},
						Attributes: map[string]*schema.Node{},
					},
					"metadata": {
						Marinate: &schema.MarinateInfo{
							Description: "Additional metadata",
							Required:    false,
							Type:        "object",
							Example:     map[string]interface{}{}, // Empty object as example
						},
						Attributes: map[string]*schema.Node{},
					},
				},
			},
		},
	}

	templateCfg := &TemplateConfig{
		AttributeTemplate: "{{.Attribute}} - {{.Description}}{{if .HasDefault}} [Default: {{.Default}}]{{end}}{{if .HasExample}} [Example: {{.Example}}]{{end}}",
		RequiredText:      "Required",
		OptionalText:      "Optional",
		EscapeMode:        "inline_code",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
	}

	r := NewRendererWithTemplate(templateCfg)
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that empty map default is shown as {}
	if !strings.Contains(result, "[Default: {}]") {
		t.Errorf("Expected empty map default shown as {} for tags. Got:\n%s", result)
	}

	// Check that empty object example is shown as {}
	if !strings.Contains(result, "[Example: {}]") {
		t.Errorf("Expected empty object example shown as {} for metadata. Got:\n%s", result)
	}

	// Ensure it's not showing map[]
	if strings.Contains(result, "map[]") {
		t.Errorf("Expected {} but found map[] in output:\n%s", result)
	}
}

func TestRenderSchema_TrimsTrailingWhitespace(t *testing.T) {
	// Test that trailing whitespace is trimmed from rendered lines
	s := &schema.Schema{
		Variable: "test_var",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"field1": {
				Marinate: &schema.MarinateInfo{
					Description: "First field",
					Required:    true,
				},
				Attributes: map[string]*schema.Node{},
			},
			"field2": {
				Marinate: &schema.MarinateInfo{
					Description: "Second field",
					Required:    false,
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	// Template that might produce trailing spaces
	templateCfg := &TemplateConfig{
		AttributeTemplate: "{{.Attribute}} - ({{.Required}}) {{.Description}} ",
		RequiredText:      "Required",
		OptionalText:      "Optional",
		EscapeMode:        "inline_code",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
	}

	r := NewRendererWithTemplate(templateCfg)
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Split into lines and check that none have trailing whitespace
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if line != "" && line != strings.TrimRight(line, " \t") {
			t.Errorf("Line %d has trailing whitespace: %q", i, line)
		}
	}
}
