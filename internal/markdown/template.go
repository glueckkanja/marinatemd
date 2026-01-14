package markdown

import (
	"fmt"
	"strings"
)

// TemplateConfig defines how markdown is generated from schema fields.
type TemplateConfig struct {
	// AttributeTemplate defines the format for rendering individual attributes.
	// Supports placeholders: {attribute}, {required}, {description}, {type}, {default}, {example}
	// Default: "`{attribute}` - ({required}) {description}"
	AttributeTemplate string `mapstructure:"attribute_template" yaml:"attribute_template"`

	// RequiredText is the text to display when an attribute is required.
	// Default: "Required"
	RequiredText string `mapstructure:"required_text" yaml:"required_text"`

	// OptionalText is the text to display when an attribute is optional.
	// Default: "Optional"
	OptionalText string `mapstructure:"optional_text" yaml:"optional_text"`

	// EscapeMode determines how values are escaped in the output.
	// Options: "inline_code", "none", "bold", "italic"
	// Default: "inline_code" (wraps in backticks)
	EscapeMode string `mapstructure:"escape_mode" yaml:"escape_mode"`

	// IndentStyle defines how nested attributes are indented.
	// Options: "spaces", "bullets"
	// Default: "bullets"
	IndentStyle string `mapstructure:"indent_style" yaml:"indent_style"`

	// IndentSize defines the number of spaces per indent level (when IndentStyle is "spaces").
	// Default: 2
	IndentSize int `mapstructure:"indent_size" yaml:"indent_size"`
}

// DefaultTemplateConfig returns the default template configuration.
func DefaultTemplateConfig() *TemplateConfig {
	return &TemplateConfig{
		AttributeTemplate: "{attribute} - ({required}) {description}",
		RequiredText:      "Required",
		OptionalText:      "Optional",
		EscapeMode:        "inline_code",
		IndentStyle:       "bullets",
		IndentSize:        2,
	}
}

// TemplateContext holds the data for rendering a single attribute.
type TemplateContext struct {
	Attribute   string
	Required    bool
	Description string
	Type        string
	Default     string
	Example     string
}

// RenderAttribute applies the template to a context and returns the formatted string.
func (tc *TemplateConfig) RenderAttribute(ctx TemplateContext) string {
	template := tc.AttributeTemplate

	// Determine required/optional text
	requiredText := tc.OptionalText
	if ctx.Required {
		requiredText = tc.RequiredText
	}

	// Apply escaping to attribute name based on escape mode
	escapedAttribute := tc.escape(ctx.Attribute)

	// Replace placeholders
	replacements := map[string]string{
		"{attribute}":   escapedAttribute,
		"{required}":    requiredText,
		"{description}": ctx.Description,
		"{type}":        ctx.Type,
		"{default}":     ctx.Default,
		"{example}":     ctx.Example,
	}

	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// escape applies the configured escape mode to a string.
func (tc *TemplateConfig) escape(s string) string {
	switch tc.EscapeMode {
	case "inline_code":
		return fmt.Sprintf("`%s`", s)
	case "bold":
		return fmt.Sprintf("**%s**", s)
	case "italic":
		return fmt.Sprintf("*%s*", s)
	case "none":
		return s
	default:
		return fmt.Sprintf("`%s`", s) // Default to inline_code
	}
}

// FormatIndent returns the appropriate indentation string for a given depth.
func (tc *TemplateConfig) FormatIndent(depth int) string {
	if tc.IndentStyle == "bullets" {
		return strings.Repeat("  ", depth) + "- "
	}
	// For "spaces" style
	return strings.Repeat(" ", depth*tc.IndentSize)
}

// Validate checks if the template configuration is valid.
func (tc *TemplateConfig) Validate() error {
	// Check for required placeholders in template
	if !strings.Contains(tc.AttributeTemplate, "{attribute}") {
		return fmt.Errorf("attribute_template must contain {attribute} placeholder")
	}

	// Validate escape mode
	validEscapeModes := map[string]bool{
		"inline_code": true,
		"none":        true,
		"bold":        true,
		"italic":      true,
	}
	if !validEscapeModes[tc.EscapeMode] {
		return fmt.Errorf("invalid escape_mode: %s (valid options: inline_code, none, bold, italic)", tc.EscapeMode)
	}

	// Validate indent style
	validIndentStyles := map[string]bool{
		"bullets": true,
		"spaces":  true,
	}
	if !validIndentStyles[tc.IndentStyle] {
		return fmt.Errorf("invalid indent_style: %s (valid options: bullets, spaces)", tc.IndentStyle)
	}

	// Validate indent size
	if tc.IndentSize < 0 {
		return fmt.Errorf("indent_size must be non-negative")
	}

	return nil
}
