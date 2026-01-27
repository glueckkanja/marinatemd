# ShowDescription Feature - Visual Examples

## Example 1: Default Behavior (show_description omitted or true)

### YAML Schema:
```yaml
schema:
  database_host:
    _marinate:
      description: "The database server hostname or IP address"
      type: string
      required: true
```

### Rendered Markdown:
```markdown
- `database_host` (Required) - The database server hostname or IP address
```

---

## Example 2: Hidden Description (show_description: false)

### YAML Schema:
```yaml
schema:
  database_host:
    _marinate:
      description: "The database server hostname or IP address"
      show_description: false
      type: string
      required: true
```

### Rendered Markdown:
```markdown
- `database_host` (Required)
```

**Note:** The attribute is still rendered with its name and required status, but the description text is omitted.

---

## Example 3: Complex Object with Mixed Settings

### YAML Schema:
```yaml
schema:
  database:
    _marinate:
      description: "Database connection configuration"
      show_description: false  # Hide this description
      type: object
      required: false
    host:
      _marinate:
        description: "Database hostname"
        type: string
        required: true
        # show_description omitted - will show description
    port:
      _marinate:
        description: "Database port number"
        show_description: false  # Hide this description
        type: number
        required: false
        default: 5432
    ssl_enabled:
      _marinate:
        description: "Enable SSL/TLS connection"
        show_description: true  # Explicitly show
        type: bool
        required: false
        default: true
```

### Rendered Markdown:
```markdown
- `database` (Optional)
  - `host` (Required) - Database hostname
  - `port` (Optional)
  - `ssl_enabled` (Optional) - Enable SSL/TLS connection
```

**Note:** 
- `database`: Rendered without description (show_description: false)
- `host`: Full rendering with description (default behavior)
- `port`: Rendered without description (show_description: false)
- `ssl_enabled`: Full rendering with description (show_description: true)

---

## Example 4: When to Use show_description: false

### Use Case: Auto-generated placeholders

When you parse Terraform variables but haven't written descriptions yet:

```yaml
schema:
  new_feature_flag:
    _marinate:
      description: "# TODO: Add description for new_feature_flag"
      show_description: false  # Hide the TODO until description is ready
      type: bool
      required: false
      default: false
```

**Rendered:**
```markdown
- `new_feature_flag` (Optional)
```

This way, the structure is documented but users don't see the incomplete TODO message.

---

## Example 5: Compact Documentation Style

For very simple or self-explanatory fields:

```yaml
schema:
  enabled:
    _marinate:
      description: "Whether the feature is enabled"
      show_description: false  # Name is self-explanatory
      type: bool
      required: false
      default: true
  
  count:
    _marinate:
      description: "Number of instances"
      show_description: false  # Name is self-explanatory
      type: number
      required: false
      default: 1
```

**Rendered:**
```markdown
- `enabled` (Optional)
- `count` (Optional)
```

Clean and minimal output while maintaining structure.
