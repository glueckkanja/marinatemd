package markdown //nolint:testpackage // tests need access to unexported types

import (
	"fmt"
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/schema"
)

// TestDemo_NetworkRulesExample demonstrates rendering with the default template.
// This test serves as a demonstration of the template system output.
func TestDemo_NetworkRulesExample(t *testing.T) {
	// Create a sample schema matching the network_rules example from the user's file
	s := &schema.Schema{
		Variable: "network_rules",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"bypass": {
				Marinate: &schema.MarinateInfo{
					Description: "Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.",
				},
				Attributes: map[string]*schema.Node{},
			},
			"default_action": {
				Marinate: &schema.MarinateInfo{
					Description: "Specifies the default action of allow or deny when no other rules match. Valid options are `Deny` or `Allow`.",
				},
				Attributes: map[string]*schema.Node{},
			},
			"ip_rules": {
				Marinate: &schema.MarinateInfo{
					Description: "List of public IP or IP ranges in CIDR Format. Only IPv4 addresses are allowed.",
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	// Render with default template
	r := NewRenderer()
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	fmt.Println("\n=== DEFAULT TEMPLATE OUTPUT ===")
	fmt.Print(result)
	fmt.Print("=== END OUTPUT ===\n")

	// Verify the output contains expected elements
	if result == "" {
		t.Error("Expected non-empty output")
	}
}

// TestDemo_CustomTemplateExample demonstrates rendering with a custom template.
func TestDemo_CustomTemplateExample(t *testing.T) {
	s := &schema.Schema{
		Variable: "network_rules",
		Version:  "1",
		SchemaNodes: map[string]*schema.Node{
			"bypass": {
				Marinate: &schema.MarinateInfo{
					Description: "Specifies whether traffic is bypassed for Logging/Metrics/AzureServices.",
				},
				Attributes: map[string]*schema.Node{},
			},
		},
	}

	// Custom template similar to table format
	customCfg := &TemplateConfig{
		AttributeTemplate: "**{attribute}** | `{type}` | {required} | {description}",
		RequiredText:      "✓",
		OptionalText:      "✗",
		EscapeMode:        "bold",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
	}

	r := NewRendererWithTemplate(customCfg)
	result, err := r.RenderSchema(s)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	fmt.Println("\n=== CUSTOM TEMPLATE OUTPUT ===")
	fmt.Print(result)
	fmt.Print("=== END OUTPUT ===\n")
}
