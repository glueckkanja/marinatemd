# MarinateMD

<img width="300" height="300" alt="martinate-md-logo" src="https://github.com/user-attachments/assets/4f628219-e026-4624-9d72-544b96d58b50" />

## What's This All About?

If you've ever stared at a terraform-docs output showing a complex `object({...})` variable and thought "well, that's completely useless", you're in the right place.

The problem: terraform-docs (and similar tools) show you the type structure but give you exactly zero context about what those nested fields actually do. Your carefully crafted object with 20 nested attributes gets dumped as a wall of HCL syntax that's technically correct and practically useless.

The usual workaround involves manually maintaining multiline markdown descriptions in your `variables.tf` files. Good luck keeping those in sync when your schema changes. Been there, done that, got the merge conflicts.

**MarinateMD** takes a different approach.

It extracts the structure from your HCL variables, generates YAML schema files where you document each field once, then injects clean markdown into your docs. When your variable schema changes, it merges the updates intelligently without nuking your descriptions.

**The workflow:**

1. **Extract**: Parse HCL and generate structured YAML schemas from complex types
2. **Document**: Fill in descriptions in YAML (persisted, not embedded in HCL)
3. **Inject**: Generate hierarchical markdown and inject it into README.md or Terraform files
4. **Split**: Optionally break monolithic docs into per-variable files

This is particularly useful for platform teams shipping reusable Terraform modules with complex configuration objects. Instead of forcing users to decode `object({database = optional(object({host = string, port = optional(number, 5432)}))})`, they get actual documentation.

Built in Go, works with Terraform and OpenTofu, plays nice with terraform-docs.

## How It Works

### Step 1: Extract the Schema

When MarinateMD encounters a variable specially marked like this in your `variables.tf`:

```hcl
variable "app_config" {
  type = object({
    database = optional(object({
      host     = string
      port     = optional(number, 5432)
      ssl_mode = optional(string, "require")
    }))
    cache = optional(object({
      redis_url = string
      ttl       = optional(number, 3600)
    }))
  })
  description = "<!-- MARINATED: app_config -->"
}
```

It generates a schema file `docs/variables/app_config.yaml`:

```yaml
variable: app_config
version: "1"

schema:
  database:
    _marinate:
      description: '# TODO: Add description for database'
      type: object
    host:
      _marinate:
        description: '# TODO: Add description for host'
        type: string
        required: true
    port:
      _marinate:
        description: '# TODO: Add description for port'
        type: number
        required: false
        default: 5432
    ssl_mode:
      _marinate:
        description: '# TODO: Add description for ssl_mode'
        type: string
        required: false
        default: "require"
  
  cache:
    _marinate:
      description: '# TODO: Add description for cache'
      type: object
    redis_url:
      _marinate:
        description: '# TODO: Add description for redis_url'
        type: string
        required: true
    ttl:
      _marinate:
        description: '# TODO: Add description for ttl'
        type: number
        required: false
        default: 3600
```

### Step 2: Marinate with Documentation

Now edit the YAML file and replace the TODO placeholders with actual documentation:

```yaml
schema:
  database:
    _marinate:
      description: |
        Database connection configuration.
        When omitted, the application runs in memory-only mode.
      type: object
    host:
      _marinate:
        description: Database server hostname or IP address
        type: string
        required: true
        example: "db.example.com"
    port:
      _marinate:
        description: PostgreSQL port number
        type: number
        required: false
        default: 5432
```

All schema metadata lives under the `_marinate` key. This keeps your documentation data cleanly separated from the nested attribute structure.

### Step 3: Generate Documentation

If you're using terraform-docs for basic variable documentation, run it now to generate your base markdown.

### Step 4: Inject Structured Markdown

Run the inject command to read your YAML schemas and inject hierarchical markdown:

```bash
marinate inject
```

This generates markdown like:

```markdown
### app_config

Description: <!-- MARINATED: app_config -->

- **database** - (Optional) Database connection configuration. When omitted, the application runs in memory-only mode.
  - **host** - (Required) Database server hostname or IP address
  - **port** - (Optional) PostgreSQL port number
  - **ssl_mode** - (Optional) SSL connection mode for database connections
- **cache** - (Optional) Redis cache configuration
  - **redis_url** - (Required) Redis connection URL
  - **ttl** - (Optional) Default cache TTL in seconds

<!-- /MARINATED: app_config -->
```

The markdown gets injected between the `<!-- MARINATED: app_config -->` markers in your README.md (or directly into your `variables.tf` if you prefer).

No more cryptic `object({...})` type dumps. Just clean, hierarchical documentation that humans can actually parse.

## Commands

### `export` - Extract Variable Schemas

Parses `variables.tf` files for variables marked with `<!-- MARINATED: name -->` comments and generates structured YAML schema files.

**Basic usage:**

```bash
marinate export .
marinate export /path/to/terraform/module
```

**What it does:**

1. Scans for `.tf` files containing variable declarations
2. Identifies variables with `description = "<!-- MARINATED: variable_name -->"`
3. Parses the HCL type structure (handles objects, optionals, lists, maps, etc.)
4. Generates or updates YAML files in `docs/variables/`
5. Merges intelligently - preserves existing descriptions when schema changes

**Schema output:**

Each variable gets a `docs/variables/<variable_name>.yaml` file containing:
- Variable name and version
- Hierarchical schema with `_marinate` metadata blocks
- Type information, required flags, defaults
- Placeholders for descriptions (or preserved existing ones)

If you modify your HCL variable structure and re-run export, it updates the schema structure while keeping your documentation intact.

### `inject` - Update Documentation

Reads YAML schemas and renders them as hierarchical markdown, injecting the output into README.md and/or Terraform variable files.

**Basic usage:**

```bash
# Default: inject from ./docs/variables/*.yaml into ./README.md
marinate inject

# Custom schema directory
marinate inject /path/to/yaml/schemas

# Inject into Terraform variable files instead
marinate inject --inject-type terraform --terraform-module ./terraform

# Inject into both README and Terraform files
marinate inject --inject-type both --terraform-module ./terraform
```

**Arguments:**

- `[schema-path]` - Directory containing YAML schema files (defaults to `./docs/variables`)
  - If path ends with `variables`, the parent directory is used automatically
  - Can be absolute or relative

**Flags:**

- `--inject-type` - Where to inject: `markdown` (default), `terraform`, or `both`
- `--markdown-file` - Target markdown file (default: `./README.md`)
- `--terraform-module` - Terraform module directory (required for `terraform` or `both` types)

**Injection targets:**

**Markdown mode:** Injects between `<!-- MARINATED: name -->` and `<!-- /MARINATED: name -->` markers in your README.md.

**Terraform mode:** Injects directly into the variable's `description` field in your `variables.tf` files, replacing the MARINATED marker comment.

**Both mode:** Does both of the above in one pass.

**Examples:**

```bash
# Default paths
marinate inject

# Custom markdown target
marinate inject --markdown-file docs/VARIABLES.md

# Update Terraform files with rendered docs
marinate inject --inject-type terraform --terraform-module .

# Update both README and Terraform files
marinate inject --inject-type both --terraform-module . --markdown-file README.md

# Custom schema location
marinate inject /custom/schemas --markdown-file docs/API.md
```

**Path resolution:**

All paths can be absolute or relative to the current directory. The tool resolves parent directories intelligently if you point at a `variables/` subdirectory.

### `split` - Post-Process Documentation

Takes a markdown file with multiple MARINATED variable sections and splits it into separate files, one per variable.

**Basic usage:**

```bash
# Default: split docs/README.md into docs/variables/*.md
marinate split .

# Custom paths
marinate split --input docs/README.md --output docs/split .

# Add header/footer templates
marinate split --header _header.md --footer _footer.md .
```

**Flags:**

- `--input` - Source markdown file (default: `docs/README.md`)
- `--output` - Output directory (default: `docs/variables`)
- `--header` - Header template to prepend to each file
- `--footer` - Footer template to append to each file

**What it does:**

1. Scans the input markdown for sections bounded by `<!-- MARINATED: name -->` markers
2. Extracts each complete variable section (heading, description, type, defaults)
3. Writes `<variable_name>.md` files to the output directory
4. Optionally wraps each file with header/footer content

**Use case:**

Useful when you want per-variable documentation files instead of a monolithic README. Particularly handy if you're generating docs with terraform-docs first, then splitting them for a documentation site or wiki.

**Example workflow:**

```bash
# Generate base docs with terraform-docs
terraform-docs markdown table . > docs/README.md

# Inject MARINATED documentation
marinate inject --markdown-file docs/README.md

# Split into individual files with custom header/footer
marinate split --header templates/header.md --footer templates/footer.md
```

## Configuration

Create a `.marinated.yml` file in your module root to configure default behavior. All settings are optional and can be overridden via CLI flags.

### Example Configuration

```yaml
# .marinated.yml

# Base paths
export_path: docs              # Where YAML schemas and docs live
docs_file: README.md           # Default markdown target for inject

# Split command defaults
split:
  input_path: README.md        # Input file (relative to export_path)
  output_dir: variables        # Output directory (relative to export_path)
  header_file: _header.md      # Header template
  footer_file: _footer.md      # Footer template

# Markdown rendering configuration
markdown_template:
  # Template for rendering each attribute line
  # Supports Go template syntax with conditionals
  attribute_template: "{{.Attribute}} - ({{.Required}}) {{.Description}}"
  
  # Text labels
  required_text: "Required"
  optional_text: "Optional"
  
  # How to escape values in output
  escape_mode: inline_code     # Options: inline_code, none, bold, italic
  
  # Indentation style for nested attributes
  indent_style: bullets        # Options: bullets, spaces
  indent_size: 2               # Spaces per indent level (when indent_style=spaces)
  
  # Optional: Add separators between top-level attributes
  separator_indents: [0]       # Depths at which to insert "---" separators
```

### Configuration Reference

**Base Settings:**

| Setting       | Description                              | Default     |
| ------------- | ---------------------------------------- | ----------- |
| `export_path` | Directory for YAML schemas and docs      | `docs`      |
| `docs_file`   | Default markdown file for inject command | `README.md` |

**Split Configuration (`split`):**

| Setting       | Description                                   | Default     |
| ------------- | --------------------------------------------- | ----------- |
| `input_path`  | Input markdown file (relative to export_path) | `README.md` |
| `output_dir`  | Output directory (relative to export_path)    | `variables` |
| `header_file` | Header template to prepend                    | _(none)_    |
| `footer_file` | Footer template to append                     | _(none)_    |

**Markdown Template (`markdown_template`):**

| Setting              | Description                                     | Default                                             |
| -------------------- | ----------------------------------------------- | --------------------------------------------------- |
| `attribute_template` | Go template for rendering attribute lines       | `{{.Attribute}} - ({{.Required}}) {{.Description}}` |
| `required_text`      | Label for required fields                       | `Required`                                          |
| `optional_text`      | Label for optional fields                       | `Optional`                                          |
| `escape_mode`        | Value escaping: inline_code, none, bold, italic | `inline_code`                                       |
| `indent_style`       | Indentation: bullets or spaces                  | `bullets`                                           |
| `indent_size`        | Spaces per indent level (when using spaces)     | `2`                                                 |
| `separator_indents`  | Depths at which to insert "---" separators      | `[]` (no separators)                                |

**Template Customization:**

The `attribute_template` field supports Go template syntax with these variables:

- `{{.Attribute}}` - Field name
- `{{.Required}}` - "Required" or "Optional" text
- `{{.Description}}` - User-provided description
- `{{.Type}}` - HCL type (string, number, object, etc.)
- `{{.Default}}` - Default value (if any)
- `{{.Example}}` - Example value (if provided)

Conditionals:
- `{{if .IsRequired}}...{{end}}`
- `{{if .HasDefault}}...{{end}}`
- `{{if .HasExample}}...{{end}}`

Example with conditionals:

```yaml
markdown_template:
  attribute_template: "{{.Attribute}}{{if .IsRequired}}*{{end}} - {{.Description}}{{if .HasDefault}} (default: {{.Default}}){{end}}"
```

**Priority Order:**

CLI flags > `.marinated.yml` > built-in defaults
