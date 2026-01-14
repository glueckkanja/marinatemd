# Markdown Template Configuration

The `marinatemd` tool supports customizable markdown templates for generating documentation from Terraform variable schemas. This allows you to control how attributes are formatted in the generated documentation.

## Default Template

By default, attributes are rendered using this format:

```
`bypass` - (Optional) Specifies whether traffic is bypassed for Logging/Metrics/AzureServices.
```

This corresponds to the default template:
```
{attribute} - ({required}) {description}
```

with `escape_mode: inline_code` which wraps the attribute name in backticks.

## Configuration

You can customize the markdown template in your `.marinated.yml` configuration file:

```yaml
markdown_template:
  # Template for rendering individual attributes
  # Available placeholders: {attribute}, {required}, {description}, {type}, {default}, {example}
  attribute_template: "{attribute} - ({required}) {description}"
  
  # Text to display for required fields
  required_text: "Required"
  
  # Text to display for optional fields
  optional_text: "Optional"
  
  # How to escape the attribute name
  # Options: "inline_code", "bold", "italic", "none"
  escape_mode: "inline_code"
  
  # How to indent nested attributes
  # Options: "bullets", "spaces"
  indent_style: "bullets"
  
  # Number of spaces per indent level (when indent_style is "spaces")
  indent_size: 2
```

## Available Placeholders

The following placeholders can be used in your `attribute_template`:

- `{attribute}` - The name of the attribute (will be escaped according to `escape_mode`)
- `{required}` - "Required" or "Optional" text based on whether the field is required
- `{description}` - The description text from the schema
- `{type}` - The type of the attribute (string, number, object, etc.)
- `{default}` - The default value (if any)
- `{example}` - An example value (if provided)

## Escape Modes

### `inline_code` (default)
Wraps the attribute name in backticks: `` `bypass` ``

### `bold`
Wraps the attribute name in double asterisks: `**bypass**`

### `italic`
Wraps the attribute name in single asterisks: `*bypass*`

### `none`
No escaping: `bypass`

## Indent Styles

### `bullets` (default)
Uses markdown list syntax with bullets:
```
- `database` - (Optional) Database configuration
  - `host` - (Required) Database host
  - `port` - (Optional) Database port
```

### `spaces`
Uses plain indentation without bullets:
```
database - (Optional) Database configuration
  host - (Required) Database host
  port - (Optional) Database port
```

## Examples

### Terraform-style Documentation

```yaml
markdown_template:
  attribute_template: "- `{attribute}` - ({required}) {description}"
  required_text: "Required"
  optional_text: "Optional"
  escape_mode: "inline_code"
  indent_style: "bullets"
```

Output:
```markdown
- `bypass` - (Optional) Specifies whether traffic is bypassed.
- `default_action` - (Required) Specifies the default action.
```

### Table-style Documentation

```yaml
markdown_template:
  attribute_template: "**{attribute}** | {type} | {required} | {description}"
  required_text: "Yes"
  optional_text: "No"
  escape_mode: "bold"
  indent_style: "spaces"
  indent_size: 2
```

Output:
```markdown
**bypass** | string | No | Specifies whether traffic is bypassed.
**default_action** | string | Yes | Specifies the default action.
```

### Compact Format

```yaml
markdown_template:
  attribute_template: "{attribute} ({required}): {description}"
  required_text: "req"
  optional_text: "opt"
  escape_mode: "none"
  indent_style: "bullets"
```

Output:
```markdown
- bypass (opt): Specifies whether traffic is bypassed.
- default_action (req): Specifies the default action.
```

## Implementation Details

The template system is implemented in the `internal/markdown` package:

- `template.go` - Defines the `TemplateConfig` struct and rendering logic
- `template_test.go` - Tests for template parsing and rendering
- `renderer.go` - Main markdown renderer that uses the template configuration

Templates are validated when the configuration is loaded, ensuring that:
- The `{attribute}` placeholder is present
- The `escape_mode` is valid
- The `indent_style` is valid
- The `indent_size` is non-negative
