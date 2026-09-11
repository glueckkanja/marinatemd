# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Run tests (all)
go test -v -race ./...

# Run a single test
go test -v -run TestName ./internal/package/...

# Lint
golangci-lint run --timeout=5m

# Format
gofmt -s -w .

# Vet
go vet ./...
```

## Architecture

MarinateMD is a Go CLI tool that improves documentation for complex Terraform/OpenTofu variables. The core pipeline is:

**Parse HCL → Build Schema → Merge YAML → Render Markdown → Inject into docs**

### Package Responsibilities

- **`cmd/marinate/main.go`** - Entrypoint, delegates to `cmd/marinatemd`
- **`cmd/marinatemd/`** - Cobra CLI commands (`root.go`, `export.go`, plus inject/split commands). Uses Viper for config/flags.
- **`internal/config/`** - `Config` struct mapped from `.marinated.yml` via Viper. `SetDefaults()` called at init, `Load()` called per-command.
- **`internal/hclparse/`** - Parses `variables.tf` files using `hashicorp/hcl/v2`. `Parser.ParseVariables()` + `ExtractMarinatedVars()` finds variables with `description = "<!-- MARINATED: name -->"`. `injector.go` handles injecting rendered markdown back into HCL files.
- **`internal/schema/`** - `Builder` converts parsed HCL variable types into an internal `Schema` struct. `MergeWithExisting()` preserves user-written descriptions when schema changes.
- **`internal/yamlio/`** - `Reader`/`Writer` for YAML schema files under `docs/variables/*.yaml`. The YAML format uses `_marinate` keys for metadata within the nested attribute structure.
- **`internal/markdown/`** - `renderer.go` turns a Schema into hierarchical markdown using Go templates. `injector.go` injects rendered output between `<!-- MARINATED: name -->` / `<!-- /MARINATED: name -->` markers. `template.go` defines `TemplateConfig` controlling output format.
- **`internal/paths/`** - Path resolution utilities shared across commands.
- **`internal/logger/`** - Thin wrapper around `charmbracelet/log`.

### Key Conventions

- **Marinated marker**: Variables opt-in via `description = "<!-- MARINATED: variable_name -->"` in HCL. The name after `MARINATED:` is the schema ID and maps to `docs/variables/<id>.yaml`.
- **YAML schema structure**: Each field has a `_marinate` subkey with `description`, `type`, `required`, `default`, `example`. The nesting mirrors the HCL object structure. User descriptions under `_marinate` are preserved across schema updates.
- **Config priority**: CLI flags > `.marinated.yml` > built-in defaults. Config is searched in `.`, `.config/`, and `~/.marinated.d/`. Env vars prefixed `MARINATED_` are also supported.
- **Deterministic output**: Generated YAML and markdown must be stable across runs (consistent ordering) to avoid noisy diffs.
- **Test packages**: Tests use a separate `_test` package (enforced by `testpackage` linter).
- **No global loggers**: `sloglint` enforces no global logger usage; pass logger via `internal/logger` package.
- **Line length**: 120 chars max (enforced by `golines`).
