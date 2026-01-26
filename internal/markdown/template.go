package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

const (
	// DefaultIndentSize is the default number of spaces per indent level.
	DefaultIndentSize = 2
)

// TemplateConfig defines how markdown is generated from schema fields.
type TemplateConfig struct {
	// AttributeTemplate defines the format for rendering individual attributes.
	// Supports Go template syntax with conditionals and functions.
	// Available fields: .Attribute, .Required, .Description, .Type, .Default, .Example
	// Available booleans: .IsRequired, .HasDefault, .HasExample, .HasType
	//
	// Simple placeholders (legacy, auto-converted):
	//   {attribute}, {required}, {description}, {type}, {default}, {example}
	//
	// Go template syntax (recommended for conditionals):
	//   {{.Attribute}} - ({{.Required}}) {{.Description}}{{if .HasDefault}} - Default: {{.Default}}{{end}}
	//
	// Default: "{{.Attribute}} - ({{.Required}}) {{.Description}}"
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

	// SeparatorIndents specifies the indent levels at which to insert "---" separators between nodes.
	// Empty list or nil means no separators. Can specify multiple levels.
	// Example: [0, 2] will add separators at depth 0 and depth 2
	SeparatorIndents []int `mapstructure:"separator_indents" yaml:"separator_indents"`

	// compiledTemplate holds the parsed Go template (internal use)
	compiledTemplate *template.Template
}

// DefaultTemplateConfig returns the default template configuration.
func DefaultTemplateConfig() *TemplateConfig {
	cfg := &TemplateConfig{
		AttributeTemplate: "{{.Attribute}} - ({{.Required}}) {{.Description}}",
		RequiredText:      "Required",
		OptionalText:      "Optional",
		EscapeMode:        "inline_code",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
		SeparatorIndents:  []int{}, // No separators by default
	}
	// Compile the template immediately
	_ = cfg.compileTemplate()
	return cfg
}

// TemplateContext holds the data for rendering a single attribute.
type TemplateContext struct {
	Attribute   string
	Required    string // String representation ("Required" or "Optional")
	IsRequired  bool   // Boolean flag for conditional checks
	Description string
	Type        string
	Default     string
	Example     string
	HasDefault  bool // Helper for conditionals
	HasExample  bool // Helper for conditionals
	HasType     bool // Helper for conditionals
}

// compileTemplate compiles the attribute template into a Go template.
// It supports both legacy {placeholder} syntax and Go template {{.Field}} syntax.
func (tc *TemplateConfig) compileTemplate() error {
	// Convert legacy placeholders to Go template syntax
	templateStr := tc.convertLegacyPlaceholders(tc.AttributeTemplate)

	// Create template with helper functions
	tmpl := template.New("attribute").Funcs(template.FuncMap{
		"escape": tc.escape,
	})

	var err error
	tc.compiledTemplate, err = tmpl.Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to compile template: %w", err)
	}

	return nil
}

// convertLegacyPlaceholders converts old {placeholder} syntax to {{.Field}} syntax
// for backward compatibility.
func (tc *TemplateConfig) convertLegacyPlaceholders(tmpl string) string {
	// Only convert if the template uses legacy syntax
	if !strings.Contains(tmpl, "{{") && strings.Contains(tmpl, "{") {
		replacements := map[string]string{
			"{attribute}":   "{{.Attribute}}",
			"{required}":    "{{.Required}}",
			"{description}": "{{.Description}}",
			"{type}":        "{{.Type}}",
			"{default}":     "{{.Default}}",
			"{example}":     "{{.Example}}",
		}
		for old, new := range replacements {
			tmpl = strings.ReplaceAll(tmpl, old, new)
		}
	}
	return tmpl
}

// RenderAttribute applies the template to a context and returns the formatted string.
func (tc *TemplateConfig) RenderAttribute(ctx TemplateContext) string {
	// Ensure template is compiled
	if tc.compiledTemplate == nil {
		if err := tc.compileTemplate(); err != nil {
			// Fallback to simple replacement if template compilation fails
			return tc.renderLegacy(ctx)
		}
	}

	// Apply escaping to attribute name
	ctx.Attribute = tc.escape(ctx.Attribute)

	// Execute template
	var buf bytes.Buffer
	if err := tc.compiledTemplate.Execute(&buf, ctx); err != nil {
		// Fallback to legacy rendering on error
		return tc.renderLegacy(ctx)
	}

	return buf.String()
}

// renderLegacy provides fallback rendering using simple string replacement.
func (tc *TemplateConfig) renderLegacy(ctx TemplateContext) string {
	result := tc.AttributeTemplate

	// Apply escaping to attribute name
	escapedAttribute := tc.escape(ctx.Attribute)

	// Replace placeholders
	replacements := map[string]string{
		"{attribute}":      escapedAttribute,
		"{{.Attribute}}":   escapedAttribute,
		"{required}":       ctx.Required,
		"{{.Required}}":    ctx.Required,
		"{description}":    ctx.Description,
		"{{.Description}}": ctx.Description,
		"{type}":           ctx.Type,
		"{{.Type}}":        ctx.Type,
		"{default}":        ctx.Default,
		"{{.Default}}":     ctx.Default,
		"{example}":        ctx.Example,
		"{{.Example}}":     ctx.Example,
	}

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
	// Try to compile the template (this also does legacy conversion)
	if err := tc.compileTemplate(); err != nil {
		return fmt.Errorf("invalid attribute_template: %w", err)
	}

	// Check for required placeholders AFTER legacy conversion
	// Use convertLegacyPlaceholders to get the actual template that will be used
	convertedTemplate := tc.convertLegacyPlaceholders(tc.AttributeTemplate)
	hasAttribute := strings.Contains(convertedTemplate, "{{.Attribute}}")
	if !hasAttribute {
		return errors.New("attribute_template must contain {attribute} or {{.Attribute}} placeholder")
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
		return errors.New("indent_size must be non-negative")
	}

	return nil
}
