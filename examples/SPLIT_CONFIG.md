# Split Command Configuration Examples

This directory demonstrates how to configure the `marinatemd split` command.

## Method 1: Configuration File

Create a `.marinated.yml` file in your module root:

```yaml
split:
  input_path: README.md
  output_dir: split_output
  header_file: _header.md
  footer_file: _footer.md
```

Then run without flags:

```bash
marinatemd split .
```

## Method 2: CLI Flags

Override settings via command-line flags:

```bash
marinatemd split --input docs/README.md \
                 --output docs/variables \
                 --header templates/_header.md \
                 --footer templates/_footer.md \
                 .
```

## Priority Order

Settings are applied in this order (highest to lowest priority):

1. **CLI flags** - Highest priority
2. **Config file** (`.marinated.yml`)
3. **Built-in defaults** - Lowest priority

### Example Priority Behavior

If `.marinated.yml` contains:
```yaml
split:
  output_dir: config_output
```

And you run:
```bash
marinatemd split --output cli_output .
```

Result: Files are written to `cli_output` (CLI flag wins)

## Files in This Example

- `.marinated.yml` - Configuration file with split settings
- `_header.md` - Header template (prepended to each split file)
- `_footer.md` - Footer template (appended to each split file)
- `docs/README.md` - Input file containing MARINATED variables
