# MarinateMD – Copilot Instructions

These instructions are for AI coding agents (GitHub Copilot, etc.) working in this repo.

## Project Purpose and Big Picture

- **Goal**: Provide a Go-based companion tool that improves documentation for complex Terraform/OpenTofu variables.
- **Core flow**: Parse `variables.tf` → derive structured variable schema → persist/editable YAML under `docs/variables/*.yaml` → regenerate **hierarchical markdown** that is injected into `README.md` (or other docs) at annotated locations.
- **Key idea**: Treat complex `object({...})` variables as schemas, not just type strings, so teams can document every nested attribute clearly while keeping docs in sync with code.

## Core Concepts and Conventions

- **Marinated variables**: Any Terraform/OpenTofu variable that should be documented by this tool is marked in `variables.tf` with a description comment of the form:
  - `description = "<!-- MARINATED: app_config -->"`
  - The name after `MARINATED:` (e.g. `app_config`) is the logical variable identifier and usually maps to `docs/variables/app_config.yaml`.
- **YAML schema files**:
  - Reside under `docs/variables/` and are named `<variable>.yaml` (e.g. `app_config.yaml`).
  - Top-level keys typically include `variable`, `version`, and a `schema` section.
  - `schema` contains:
    - Optional `_meta` node with a multi-line `description` for the overall variable.
    - One entry per attribute (e.g. `database`, `cache`), each possibly containing its own `_meta` and leaf fields.
    - Leaf fields track `description`, `required`, and optional `default`/`example` values.
- **Generated markdown**:
  - The tool is responsible for turning the YAML schema into **nested markdown** that replaces placeholder content in `README.md`.
  - Markdown is structured (headings, lists, or tables) to mirror the object hierarchy and avoids exposing raw `object({...})` syntax to end users.

## Expected Code Structure (when implementing)

- **Language**: Go.
- **CLI libraries**: Use `cobra` for the command-line structure and `viper` for configuration and flags.
- **Likely packages to create** (if not present yet):
  - `internal/hclparse` or similar: parse HCL from `variables.tf` and extract complex variable definitions.
  - `internal/schema`: map HCL variable types to an internal schema representation (objects, optionals, defaults, required flags).
  - `internal/yamlio`: read/write `docs/variables/*.yaml`, preserving existing descriptions while updating structure on schema changes.
  - `internal/markdown`: render hierarchical markdown from the schema model.
  - `cmd/marinatemd`: CLI entrypoint wiring together parsing, YAML sync, and markdown injection using Cobra/Viper.
- **CLI behavior** (to follow when adding commands):
  - Accept a module root path (default current directory) and locate `variables.*.tf` / `README.md` there.
  - Scan for `<!-- MARINATED: name -->` markers in variable descriptions and the docs.
  - For each variable:
    - Generate or update a YAML schema file.
    - Merge new schema structure with existing YAML descriptions (do not discard user-written docs).
    - Regenerate markdown and inject it into `README.md` at the appropriate location.

## Project-Specific Practices
- **Preserve user content**: When updating YAML schemas or markdown, never overwrite human-written descriptions if the corresponding field still exists in the HCL schema; only add/remove entries when the schema changes.
- **Stable file layout**: Keep YAML under `docs/variables/` and do not introduce alternative locations unless you add configuration flags and update docs.
- **Terraform/OpenTofu compatibility**: Write parsing logic against generic HCL where possible so it can support both Terraform and OpenTofu modules.
- **Deterministic output**: Ensure generated YAML and markdown are stable across runs (consistent field ordering, heading levels, and list ordering) to minimize noisy diffs.
 - **Testing-first mindset**: Treat tests as a core part of development. Add or update Go tests when changing HCL parsing, schema modeling, YAML IO, markdown rendering, or CLI behavior, and keep `go test ./...` passing.
- **Linting and formatting**: Follow existing Go code style and conventions. Use `golangci-lint` as configured in `.golangci.yml` to catch issues before committing.

## How to Work Effectively as an AI Agent
- When adding new features, **extend the existing extraction → YAML → markdown pipeline** instead of introducing separate, parallel flows.
- Prefer small, composable Go packages with clear responsibilities (parse, model, IO, render) rather than monolithic functions.
- Whenever you touch the YAML or markdown generation logic, update the `README.md` examples to reflect any format changes.
- If you introduce new configuration (flags, env vars), document it in `README.md` under a usage or configuration section and reference how it affects schema generation or markdown injection.
