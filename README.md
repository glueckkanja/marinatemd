# MarinateMD

## 🧊 What's This All About?

Ever tried documenting complex Terraform variables and felt like you're wrestling with a giant blob of HCL? You know the pain: your beautifully structured `object({...})` variables get flattened into unreadable type definitions in terraform-docs, leaving your team scratching their heads about what each nested attribute actually does.

**MarinateMD** is your escape hatch from documentation hell.

This Go-powered companion tool transforms the way you document complex Terraform/OpenTofu variables by:

🔍 **Extracting** the hidden structure from your complex variable types  
📝 **Generating** human-friendly YAML schemas where you can document every single attribute  
🔄 **Merging** updates intelligently - your custom descriptions survive schema changes  
🎯 **Injecting** beautiful, structured markdown back into your README.md  

Think of it as terraform-docs on steroids, specifically designed for those gnarly object variables with dozens of nested attributes. Instead of showing developers a wall of type definitions, they get clean, hierarchical documentation that actually explains what each field does, what's required, and what the defaults are.

Perfect for platform teams building complex modules, infrastructure engineers tired of undocumented variables, and anyone who believes that good docs are as important as good code (and a good tofu).

**TL;DR:** Stop letting complex Terraform variables be documentation black holes. Marinate them in markdown instead.

## 🥒 How It Works

### Step 1: Extract the Schema

When MarinateMD encounters a complex variable like this in your `variables.tf`:

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

It generates `docs/variables/app_config.yaml`:

```yaml
variable: app_config
version: 1

schema:
  _meta:
    description: |
      # TODO: Add description for app_config
  
  database:
    _meta:
      description: |
        # TODO: Add description for database configuration
    host:
      description: |
        # TODO: Add description for host
      required: true
    port:
      description: |
        # TODO: Add description for port
      required: false
      default: 5432
    ssl_mode:
      description: |
        # TODO: Add description for ssl_mode
      required: false
      default: "require"
  
  cache:
    _meta:
      description: |
        # TODO: Add description for cache configuration
    redis_url:
      description: |
        # TODO: Add description for redis_url
      required: true
    ttl:
      description: |
        # TODO: Add description for ttl
      required: false
      default: 3600
```

### Step 2: Marinate with Documentation

You fill in the descriptions:

```yaml
schema:
  _meta:
    description: |
      Application configuration object containing database and cache settings.
  
  database:
    _meta:
      description: |
        Database connection configuration. When omitted, the app runs in memory-only mode.
    host:
      description: |
        Database server hostname or IP address
      required: true
      example: "db.example.com"
    port:
      description: |
        Database server port number
      required: false
      default: 5432
```

### Step 3: Inject Beautiful Markdown

## MarinateMD Structured Markdown Output

MarinateMD automatically transforms complex object type definitions into clean, hierarchical markdown documentation. This eliminates the need to manually write multiline markdown descriptions for complex data structures.

### Benefits

- **Automatic Generation**: No need to manually craft verbose markdown documentation for object types
- **Clear Structure**: Replaces cryptic `object({...})` type annotations with readable, nested documentation
- **Team-Friendly**: Provides meaningful explanations of each field's purpose and structure
- **Reduced Maintenance**: Documentation stays in sync with your code without manual markdown editing

### How It Works

MarinateMD processes your variable descriptions and automatically generates structured markdown that replaces placeholder content, transforming unclear type definitions into comprehensive documentation that your team can actually understand and use - just add `<!-- MARINATED: variable_name -->` and let the tool handle the rest.
