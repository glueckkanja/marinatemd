package markdown

import (
	"testing"
)

func TestDefaultTemplateConfig(t *testing.T) {
	cfg := DefaultTemplateConfig()

	if cfg.AttributeTemplate != "{attribute} - ({required}) {description}" {
		t.Errorf("Expected default attribute template, got: %s", cfg.AttributeTemplate)
	}

	if cfg.RequiredText != "Required" {
		t.Errorf("Expected 'Required', got: %s", cfg.RequiredText)
	}

	if cfg.OptionalText != "Optional" {
		t.Errorf("Expected 'Optional', got: %s", cfg.OptionalText)
	}

	if cfg.EscapeMode != "inline_code" {
		t.Errorf("Expected 'inline_code', got: %s", cfg.EscapeMode)
	}

	if cfg.IndentStyle != "bullets" {
		t.Errorf("Expected 'bullets', got: %s", cfg.IndentStyle)
	}

	if cfg.IndentSize != 2 {
		t.Errorf("Expected indent size 2, got: %d", cfg.IndentSize)
	}
}

func TestRenderAttribute_DefaultTemplate(t *testing.T) {
	cfg := DefaultTemplateConfig()

	tests := []struct {
		name     string
		ctx      TemplateContext
		expected string
	}{
		{
			name: "required attribute",
			ctx: TemplateContext{
				Attribute:   "bypass",
				Required:    true,
				Description: "Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.",
			},
			expected: "`bypass` - (Required) Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.",
		},
		{
			name: "optional attribute",
			ctx: TemplateContext{
				Attribute:   "bypass",
				Required:    false,
				Description: "Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.",
			},
			expected: "`bypass` - (Optional) Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.",
		},
		{
			name: "attribute with type",
			ctx: TemplateContext{
				Attribute:   "port",
				Required:    false,
				Description: "The port number",
				Type:        "number",
			},
			expected: "`port` - (Optional) The port number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.RenderAttribute(tt.ctx)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRenderAttribute_CustomTemplate(t *testing.T) {
	cfg := &TemplateConfig{
		AttributeTemplate: "**{attribute}** ({type}): {description} [Required: {required}]",
		RequiredText:      "Yes",
		OptionalText:      "No",
		EscapeMode:        "bold",
		IndentStyle:       "bullets",
		IndentSize:        2,
	}

	ctx := TemplateContext{
		Attribute:   "database_url",
		Required:    true,
		Description: "The database connection string",
		Type:        "string",
	}

	result := cfg.RenderAttribute(ctx)
	expected := "****database_url**** (string): The database connection string [Required: Yes]"

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		name       string
		escapeMode string
		input      string
		expected   string
	}{
		{
			name:       "inline_code",
			escapeMode: "inline_code",
			input:      "my_attribute",
			expected:   "`my_attribute`",
		},
		{
			name:       "bold",
			escapeMode: "bold",
			input:      "my_attribute",
			expected:   "**my_attribute**",
		},
		{
			name:       "italic",
			escapeMode: "italic",
			input:      "my_attribute",
			expected:   "*my_attribute*",
		},
		{
			name:       "none",
			escapeMode: "none",
			input:      "my_attribute",
			expected:   "my_attribute",
		},
		{
			name:       "invalid defaults to inline_code",
			escapeMode: "invalid",
			input:      "my_attribute",
			expected:   "`my_attribute`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TemplateConfig{EscapeMode: tt.escapeMode}
			result := cfg.escape(tt.input)
			if result != tt.expected {
				t.Errorf("Expected: %s, Got: %s", tt.expected, result)
			}
		})
	}
}

func TestFormatIndent(t *testing.T) {
	tests := []struct {
		name        string
		indentStyle string
		indentSize  int
		depth       int
		expected    string
	}{
		{
			name:        "bullets depth 0",
			indentStyle: "bullets",
			indentSize:  2,
			depth:       0,
			expected:    "- ",
		},
		{
			name:        "bullets depth 1",
			indentStyle: "bullets",
			indentSize:  2,
			depth:       1,
			expected:    "  - ",
		},
		{
			name:        "bullets depth 2",
			indentStyle: "bullets",
			indentSize:  2,
			depth:       2,
			expected:    "    - ",
		},
		{
			name:        "spaces depth 0",
			indentStyle: "spaces",
			indentSize:  4,
			depth:       0,
			expected:    "",
		},
		{
			name:        "spaces depth 1",
			indentStyle: "spaces",
			indentSize:  4,
			depth:       1,
			expected:    "    ",
		},
		{
			name:        "spaces depth 2",
			indentStyle: "spaces",
			indentSize:  2,
			depth:       2,
			expected:    "    ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TemplateConfig{
				IndentStyle: tt.indentStyle,
				IndentSize:  tt.indentSize,
			}
			result := cfg.FormatIndent(tt.depth)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *TemplateConfig
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			cfg:       DefaultTemplateConfig(),
			wantError: false,
		},
		{
			name: "missing attribute placeholder",
			cfg: &TemplateConfig{
				AttributeTemplate: "({required}) {description}",
				EscapeMode:        "inline_code",
				IndentStyle:       "bullets",
			},
			wantError: true,
			errorMsg:  "attribute_template must contain {attribute} placeholder",
		},
		{
			name: "invalid escape mode",
			cfg: &TemplateConfig{
				AttributeTemplate: "{attribute} - ({required}) {description}",
				EscapeMode:        "invalid_mode",
				IndentStyle:       "bullets",
			},
			wantError: true,
			errorMsg:  "invalid escape_mode",
		},
		{
			name: "invalid indent style",
			cfg: &TemplateConfig{
				AttributeTemplate: "{attribute} - ({required}) {description}",
				EscapeMode:        "inline_code",
				IndentStyle:       "tabs",
			},
			wantError: true,
			errorMsg:  "invalid indent_style",
		},
		{
			name: "negative indent size",
			cfg: &TemplateConfig{
				AttributeTemplate: "{attribute} - ({required}) {description}",
				EscapeMode:        "inline_code",
				IndentStyle:       "spaces",
				IndentSize:        -1,
			},
			wantError: true,
			errorMsg:  "indent_size must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %s", err.Error())
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
