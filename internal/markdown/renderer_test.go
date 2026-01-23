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
				Type:     "string",
				Required: false,
				Marinate: &schema.MarinateInfo{
					Description: "Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.",
				},
				Attributes: map[string]*schema.Node{},
			},
			"default_action": {
				Type:     "string",
				Required: true,
				Marinate: &schema.MarinateInfo{
					Description: "Specifies the default action of allow or deny when no other rules match.",
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
			"database": {
				Type:     "object",
				Required: false,
				Marinate: &schema.MarinateInfo{
					Description: "Database configuration settings",
				},
				Attributes: map[string]*schema.Node{
					"host": {
						Type:     "string",
						Required: true,
						Marinate: &schema.MarinateInfo{
							Description: "The database host address",
						},
						Attributes: map[string]*schema.Node{},
					},
					"port": {
						Type:     "number",
						Required: false,
						Marinate: &schema.MarinateInfo{
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
				Type:     "string",
				Required: true,
				Marinate: &schema.MarinateInfo{
					Description: "A test field",
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
			"zebra": {
				Type:     "string",
				Required: false,
				Marinate: &schema.MarinateInfo{
					Description: "Last alphabetically",
				},
				Attributes: map[string]*schema.Node{},
			},
			"alpha": {
				Type:     "string",
				Required: false,
				Marinate: &schema.MarinateInfo{
					Description: "First alphabetically",
				},
				Attributes: map[string]*schema.Node{},
			},
			"middle": {
				Type:     "string",
				Required: false,
				Marinate: &schema.MarinateInfo{
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

func TestRenderSchema_WithObjectSeparators(t *testing.T) {
	tests := []struct {
		name         string
		separators   []ObjectSeparator
		expectedText string
		notExpected  string
	}{
		{
			name:         "no separators",
			separators:   []ObjectSeparator{},
			expectedText: "`parent`",
			notExpected:  "---",
		},
		{
			name: "blank separator at level 1",
			separators: []ObjectSeparator{
				{Level: 1, Style: SeparatorStyleBlank, Count: 1},
			},
			expectedText: "`parent`",
			notExpected:  "---", // Blank lines, not horizontal rules
		},
		{
			name: "line separator at level 1",
			separators: []ObjectSeparator{
				{Level: 1, Style: SeparatorStyleLine},
			},
			expectedText: "---", // Should have horizontal rule between child objects
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TemplateConfig{
				AttributeTemplate: "{attribute} - ({required}) {description}",
				RequiredText:      "Required",
				OptionalText:      "Optional",
				EscapeMode:        "inline_code",
				IndentStyle:       "bullets",
				IndentSize:        2,
				ObjectSeparators:  tt.separators,
			}

			// Create schema with nested objects
			s := &schema.Schema{
				Variable: "test_var",
				Version:  "1",
				SchemaNodes: map[string]*schema.Node{
					"parent": {
						Type:     "object",
						Required: false,
						Marinate: &schema.MarinateInfo{
							Description: "Parent object",
						},
						Attributes: map[string]*schema.Node{
							"child1": {
								Type:     "object",
								Required: false,
								Marinate: &schema.MarinateInfo{
									Description: "First child object",
								},
								Attributes: map[string]*schema.Node{},
							},
							"child2": {
								Type:     "object",
								Required: false,
								Marinate: &schema.MarinateInfo{
									Description: "Second child object",
								},
								Attributes: map[string]*schema.Node{},
							},
						},
					},
				},
			}

			r := NewRendererWithTemplate(cfg)
			result, err := r.RenderSchema(s)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !strings.Contains(result, tt.expectedText) {
				t.Errorf("Expected to find %q in result:\n%s", tt.expectedText, result)
			}

			if tt.notExpected != "" && strings.Contains(result, tt.notExpected) {
				t.Errorf("Did not expect to find %q in result:\n%s", tt.notExpected, result)
			}
		})
	}
}
