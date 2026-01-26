# ShowDescription Feature Documentation

## Overview

The `show_description` field in the `_marinate` node allows you to control whether a description text is rendered in the final markdown output. This is useful when you want to show the attribute structure (name, required/optional status, default values) but omit the description text itself.

## Behavior

- **Omitted (nil)**: When `show_description` is not specified in the YAML schema, the description will be shown by default
- **Explicitly `true`**: The description text will be included in the rendered markdown
- **Explicitly `false`**: The attribute line is still rendered (with name, required/optional, defaults) but the description text is omitted

## Use Cases

1. **Placeholder attributes**: Show the attribute exists but don't provide description yet
2. **Auto-generated documentation**: Display structure without verbose descriptions
3. **Compact output**: Reduce markdown size by omitting obvious or redundant descriptions
4. **Progressive documentation**: Start with attribute names and add descriptions later

## YAML Schema Syntax

```yaml
schema:
  my_field:
    _marinate:
      description: "This is the field description"
      show_description: false  # Hides the description from rendered output
      type: string
      required: true
```

## Examples

### Example 1: Default Behavior (Show Description)

```yaml
schema:
  app_name:
    _marinate:
      description: "The application name"
      type: string
      required: true
      # show_description is omitted - defaults to showing the description
```

**Rendered output:**
```markdown
- `app_name` (Required) - The application name
```

### Example 2: Explicitly Hide Description

```yaml
schema:
  internal_config:
    _marinate:
      description: "Internal configuration - not for public documentation"
      show_description: false
      type: object
      required: false
```

**Rendered output:**
```markdown
- `internal_config` (Optional)
```

Note: The attribute is still rendered with its name and optional status, but without the description text.

### Example 3: Hide Parent Description, Show Children

```yaml
schema:
  credentials:
    _marinate:
      description: "Sensitive credentials"
      show_description: false  # Hide parent description text
      type: object
      required: false
    username:
      _marinate:
        description: "Database username"
        # Omitted - shows by default
        type: string
        required: true
    password:
      _marinate:
        description: "Database password"
        # Omitted - shows by default
        type: string
        required: true
```

**Rendered output:**
```markdown
- `credentials` (Optional)
  - `username` (Required) - Database username
  - `password` (Required) - Database password
```

Note: The parent `credentials` is rendered without description text, but child fields show their descriptions.

### Example 4: Mixed Visibility

```yaml
schema:
  database:
    _marinate:
      description: "Database configuration"
      # Omitted - shows by default
      type: object
    host:
      _marinate:
        description: "Database host"
        show_description: true  # Explicitly shown
        type: string
        required: true
    port:
      _marinate:
        description: "Internal port configuration - not documented"
        show_description: false  # Hidden
        type: number
        default: 5432
    ssl_mode:
      _marinate:
        description: "SSL connection mode"
        # Omitted - shows by default
        type: string
        default: "require"
```

**Rendered output:**
```markdown
- `database` - Database configuration
  - `host` (Required) - Database host
  - `ssl_mode` - SSL connection mode
```

Note: The `port` field is rendered as `- port (Optional)` without the description text.

## Implementation Details

### Schema Structure

The `show_description` field is defined as a pointer to bool (`*bool`) in the `MarinateInfo` struct:

```go
type MarinateInfo struct {
    Description     string `yaml:"description,omitempty"`
    ShowDescription *bool  `yaml:"show_description,omitempty"`
    // ... other fields
}
```

Using a pointer allows us to distinguish between three states:
- `nil` (not set) - defaults to showing the description
- `true` - explicitly show the description
- `false` - explicitly hide the description

### Rendering Logic

The markdown renderer checks the `show_description` field before rendering content:

```go
func (r *Renderer) renderNodeContent(name string, node *schema.Node, depth int, builder *strings.Builder) {
    // Determine the description to use
    description := node.Marinate.Description
    
    // Default to showing description if not specified
    showDescription := true
    if node.Marinate.ShowDescription != nil {
        showDescription = *node.Marinate.ShowDescription
    }

    // If hidden, use empty description but still render the attribute
    if !showDescription {
        description = ""
    }

    // Render the attribute with (possibly empty) description
    ctx := TemplateContext{
        Attribute:   name,
        Required:    node.Marinate.Required,
        Description: description,  // Empty string if hidden
        Type:        node.Marinate.Type,
    }
    // ... render the attribute line
}
```

## Testing

The feature includes comprehensive tests:

1. **Default behavior**: Verify descriptions are shown when `show_description` is omitted
2. **Explicit false**: Verify attribute is rendered but description text is empty when set to `false`
3. **Explicit true**: Verify descriptions are shown when set to `true`
4. **Nested structures**: Verify parent and child visibility settings work independently

See `internal/schema/schema_test.go` and `internal/markdown/renderer_test.go` for test implementations.

## Migration Guide

Existing schemas without `show_description` will continue to work unchanged - all descriptions will be shown by default, maintaining backward compatibility.

To hide a description text (while keeping the attribute):

1. Open your YAML schema file
2. Add `show_description: false` to the `_marinate` section of any node
3. Run the export command to regenerate markdown

```bash
marinatemd export
```

The attribute will be rendered as: `- attribute_name (Required/Optional)` without the description text.

## Best Practices

1. **Use for structure**: Good for showing what attributes exist without verbose descriptions
2. **Combine with defaults**: Show attribute with default value but skip redundant description
3. **Progressive documentation**: Start with `show_description: false` and add descriptions incrementally
4. **Test your output**: Always verify the rendered markdown shows the attribute metadata correctly
