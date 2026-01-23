package markdown

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// DefaultIndentSize is the default number of spaces per indent level.
	DefaultIndentSize = 2

	// SeparatorStyleNone means no separation between nested objects.
	SeparatorStyleNone = "none"

	// SeparatorStyleLine inserts a horizontal line (---) between objects.
	SeparatorStyleLine = "line"

	// SeparatorStyleBlank inserts blank lines between objects.
	SeparatorStyleBlank = "blank"

	// SeparatorStyleFence inserts a fence/divider (---) between objects.
	SeparatorStyleFence = "fence"
)

// TemplateConfig defines how markdown is generated from schema fields.
type TemplateConfig struct {
	// AttributeTemplate defines the format for rendering individual attributes.
	// Supports placeholders: {attribute}, {required}, {description}, {type}, {default}, {example}
	// Default: "{attribute} - ({required}) {description}"
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

	// ObjectSeparators defines visual separation rules for nested objects at specific depths.
	// If nil or empty, no separators are inserted (default behavior).
	// Rules are applied in order, with later rules overriding earlier ones for the same level.
	ObjectSeparators []ObjectSeparator `mapstructure:"object_separators" yaml:"object_separators,omitempty"`
}

// ObjectSeparator defines how to visually separate nested objects at a specific depth level.
type ObjectSeparator struct {
	// Level is the nesting depth at which to apply this separator (0 = top-level).
	// Use -1 to apply to all levels, or >= 0 for specific levels.
	Level int `mapstructure:"level" yaml:"level"`

	// Style defines the type of separation to insert.
	// Options: "none", "blank", "line", "fence"
	//   - "none": No separation (removes any default)
	//   - "blank": Insert blank line(s)
	//   - "line": Insert horizontal line (---)
	//   - "fence": Insert fence/divider (---)
	// Default: "none"
	Style string `mapstructure:"style" yaml:"style"`

	// Count specifies how many times to repeat the separator (for blank lines).
	// Only applies when Style is "blank". Default: 1
	Count int `mapstructure:"count" yaml:"count,omitempty"`
}

// DefaultTemplateConfig returns the default template configuration.
func DefaultTemplateConfig() *TemplateConfig {
	return &TemplateConfig{
		AttributeTemplate: "{attribute} - ({required}) {description}",
		RequiredText:      "Required",
		OptionalText:      "Optional",
		EscapeMode:        "inline_code",
		IndentStyle:       "bullets",
		IndentSize:        DefaultIndentSize,
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
		return errors.New("attribute_template must contain {attribute} placeholder")
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

	// Validate object separators
	for i, sep := range tc.ObjectSeparators {
		if err := validateSeparator(&sep, i); err != nil {
			return fmt.Errorf("object_separators[%d]: %w", i, err)
		}
	}

	return nil
}

// validateSeparator checks if an ObjectSeparator configuration is valid.
func validateSeparator(sep *ObjectSeparator, _ int) error {
	// Validate level (must be >= -1)
	if sep.Level < -1 {
		return fmt.Errorf("level must be >= -1 (got %d)", sep.Level)
	}

	// Validate style
	validStyles := map[string]bool{
		SeparatorStyleNone:  true,
		SeparatorStyleBlank: true,
		SeparatorStyleLine:  true,
		SeparatorStyleFence: true,
	}
	if sep.Style == "" {
		sep.Style = SeparatorStyleNone // Default
	}
	if !validStyles[sep.Style] {
		return fmt.Errorf("invalid style: %s (valid options: none, blank, line, fence)", sep.Style)
	}

	// Validate count (must be positive for blank style)
	if sep.Style == SeparatorStyleBlank {
		if sep.Count <= 0 {
			sep.Count = 1 // Default to 1 blank line
		}
	}

	return nil
}

// GetSeparatorForLevel returns the separator configuration for a given nesting depth.
// Returns nil if no separator is configured for that level.
func (tc *TemplateConfig) GetSeparatorForLevel(depth int) *ObjectSeparator {
	if len(tc.ObjectSeparators) == 0 {
		return nil
	}

	var matchedSep *ObjectSeparator

	// Apply rules in order, allowing later rules to override earlier ones
	for i := range tc.ObjectSeparators {
		sep := &tc.ObjectSeparators[i]

		// Check if this rule applies to the current depth
		if sep.Level == -1 || sep.Level == depth {
			matchedSep = sep
		}
	}

	return matchedSep
}

// RenderSeparator returns the markdown string for a separator based on its style.
func (tc *TemplateConfig) RenderSeparator(sep *ObjectSeparator) string {
	if sep == nil || sep.Style == SeparatorStyleNone {
		return ""
	}

	switch sep.Style {
	case SeparatorStyleBlank:
		count := sep.Count
		if count <= 0 {
			count = 1
		}
		return strings.Repeat("\n", count)
	case SeparatorStyleLine, SeparatorStyleFence:
		return "\n---\n\n"
	default:
		return ""
	}
}
