# Markdown Template System - Implementation Summary

## Overview
Implemented a flexible templating configuration system for generating structured markdown from Terraform variable schemas. The system allows users to customize how attributes are rendered in documentation while maintaining consistency and supporting hierarchical structures.

## Key Features

### 1. Template Placeholders
The system supports the following placeholders in templates:
- `{attribute}` - The attribute name (with configurable escaping)
- `{required}` - "Required" or "Optional" text (customizable)
- `{description}` - The attribute description
- `{type}` - The attribute type (string, number, object, etc.)
- `{default}` - Default value (if present)
- `{example}` - Example value (if provided)

### 2. Escape Modes
Control how attribute names are formatted:
- `inline_code` (default): Wraps in backticks → `` `attribute` ``
- `bold`: Wraps in double asterisks → `**attribute**`
- `italic`: Wraps in single asterisks → `*attribute*`
- `none`: No formatting → `attribute`

### 3. Indent Styles
Support for hierarchical attribute rendering:
- `bullets` (default): Uses markdown bullet lists with proper indentation
- `spaces`: Uses plain space indentation

### 4. Configurable Text
Customize the text used for required/optional indicators to match your documentation style.

## Default Template

The default configuration produces output matching the user's example:

**Template**: `{attribute} - ({required}) {description}`
**Escape Mode**: `inline_code`
**Output**: 
```markdown
- `bypass` - (Optional) Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.
```

## Implementation Details

### Files Created
1. **`internal/markdown/template.go`** (177 lines)
   - `TemplateConfig` struct with all configuration options
   - `DefaultTemplateConfig()` function returning default settings
   - `TemplateContext` struct for rendering data
   - `RenderAttribute()` method applying templates
   - `escape()` method for formatting attribute names
   - `FormatIndent()` method for hierarchical indentation
   - `Validate()` method ensuring configuration validity

2. **`internal/markdown/template_test.go`** (310 lines)
   - Comprehensive test suite covering:
     - Default configuration values
     - Attribute rendering with various templates
     - All escape modes
     - Indent formatting at different depths
     - Configuration validation
     - Edge cases and error conditions

3. **`internal/markdown/demo_test.go`** (87 lines)
   - Demonstration tests showing real-world usage
   - Examples of default and custom template output

### Files Modified
1. **`internal/config/config.go`**
   - Added `MarkdownTemplate *markdown.TemplateConfig` field
   - Updated `Load()` to initialize with defaults and validate
   - Extended `SetDefaults()` to include template defaults

2. **`internal/markdown/renderer.go`**
   - Added `templateCfg` field to `Renderer` struct
   - Implemented `NewRendererWithTemplate()` constructor
   - Implemented `RenderSchema()` method using templates
   - Added `renderNode()` recursive method for hierarchical rendering
   - Ensures deterministic output with sorted keys

3. **`internal/markdown/renderer_test.go`**
   - Replaced placeholder tests with comprehensive test suite
   - Added tests for simple attributes, nested objects, custom templates
   - Tests for deterministic ordering and error handling

### Documentation Created
1. **`docs/markdown-templates.md`** - Complete user guide covering:
   - Configuration format
   - Available placeholders
   - Escape modes with examples
   - Indent styles with examples
   - Multiple real-world examples
   - Implementation details

2. **`examples/.marinated.yml`** - Example configuration file with:
   - All configuration options documented
   - Default values shown
   - Comments explaining each setting

## Configuration Example

```yaml
markdown_template:
  attribute_template: "{attribute} - ({required}) {description}"
  required_text: "Required"
  optional_text: "Optional"
  escape_mode: "inline_code"
  indent_style: "bullets"
  indent_size: 2
```

## Test Results
All tests pass successfully:
- 18 template tests (all passing)
- 8 renderer tests (all passing)
- 2 demo tests (all passing)
- All existing tests remain passing

## Usage
Users can now:
1. Use the default template (matches the example in the user's request)
2. Customize templates in `.marinated.yml`
3. Choose from multiple escape modes and indent styles
4. Create custom formatting patterns for their documentation needs

## Benefits
- **Flexibility**: Users can match their documentation style guidelines
- **Consistency**: Templates ensure uniform formatting across all generated docs
- **Maintainability**: Template configuration is separate from code
- **Validation**: Configuration is validated on load to catch errors early
- **Extensibility**: Easy to add new placeholders or formatting options in the future
