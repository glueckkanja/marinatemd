package marinatemd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"github.com/c4a8-azure/marinatemd/internal/logger"
	"github.com/c4a8-azure/marinatemd/internal/markdown"
	"github.com/c4a8-azure/marinatemd/internal/paths"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
	"github.com/spf13/cobra"
)

const (
	injectTypeMarkdown  = "markdown"
	injectTypeTerraform = "terraform"
	injectTypeBoth      = "both"
)

var (
	markdownFile    string
	injectType      string
	terraformModule string
)

// injectCmd represents the inject command that reads YAML schemas and injects markdown into documentation.
var injectCmd = &cobra.Command{
	Use:   "inject [schema-path]",
	Short: "Inject YAML schema documentation into markdown or Terraform files",
	Long: `Read YAML schema files and inject rendered markdown documentation into markdown and/or Terraform files.

Arguments:
  [schema-path]  Optional path to directory containing YAML schema files (*.yaml).
                 Defaults to <current-dir>/docs/variables

Flags:
  --inject-type        Type of injection: "markdown" (default), "terraform", or "both".
                       - markdown: inject into markdown files only
                       - terraform: inject into Terraform variable files only
                       - both: inject into both markdown and Terraform files
  --markdown-file      Path to the markdown file to inject into.
                       Can be absolute or relative to current working directory.
                       Defaults to <current-dir>/README.md
                       Required when inject-type is "markdown" or "both".
  --terraform-module   Path to the Terraform module directory containing variables*.tf files.
                       Can be absolute or relative to current working directory.
                       Required when inject-type is "terraform" or "both".

Examples:
  # 1. Use default paths (./docs/variables/*.yaml → ./README.md)
  marinatemd inject

  # 2. Inject into Terraform files only
  marinatemd inject --inject-type terraform --terraform-module ./terraform

  # 3. Inject into both markdown and Terraform files
  marinatemd inject --inject-type both --terraform-module ./terraform --markdown-file README.md

  # 4. Custom schema path and custom markdown file
  marinatemd inject /path/to/variables --markdown-file docs/API.md
  marinatemd inject ./docs/variables --markdown-file /abs/path/to/doc.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInject,
}

func init() {
	rootCmd.AddCommand(injectCmd)

	injectCmd.Flags().StringVar(
		&markdownFile,
		"markdown-file",
		"",
		"markdown file to inject into (absolute or relative to current directory)",
	)

	injectCmd.Flags().StringVar(
		&injectType,
		"inject-type",
		"markdown",
		"type of injection: markdown, terraform, or both",
	)

	injectCmd.Flags().StringVar(
		&terraformModule,
		"terraform-module",
		"",
		"path to Terraform module directory (required for terraform or both inject types)",
	)
}

func runInject(_ *cobra.Command, args []string) error {
	// Load configuration (for template settings)
	moduleRoot, cfg, err := paths.SetupEnvironment(args)
	if err != nil {
		return err
	}

	logger.Log.Debug("loaded configuration", "moduleRoot", moduleRoot)

	// Validate inject type
	if validateErr := validateInjectType(); validateErr != nil {
		return validateErr
	}

	schemaBasePath, markdownPath, terraformPath, err := resolveInjectPaths(args)
	if err != nil {
		return err
	}

	logger.Log.Info("injecting documentation",
		"schemaBasePath", schemaBasePath,
		"injectType", injectType,
		"markdownPath", markdownPath,
		"terraformPath", terraformPath)

	// Handle markdown injection
	if injectType == injectTypeMarkdown || injectType == injectTypeBoth {
		if mdErr := injectMarkdown(schemaBasePath, markdownPath, cfg); mdErr != nil {
			return mdErr
		}
	}

	// Handle Terraform injection
	if injectType == injectTypeTerraform || injectType == injectTypeBoth {
		if tfErr := injectTerraform(schemaBasePath, terraformPath, cfg); tfErr != nil {
			return tfErr
		}
	}

	return nil
}

// validateInjectType validates the inject-type flag value.
func validateInjectType() error {
	validTypes := map[string]bool{
		injectTypeMarkdown:  true,
		injectTypeTerraform: true,
		injectTypeBoth:      true,
	}

	if !validTypes[injectType] {
		return fmt.Errorf("invalid inject-type: %s (must be markdown, terraform, or both)", injectType)
	}

	return validateTerraformModuleFlag()
}

func validateTerraformModuleFlag() error {
	// Validate that terraform-module is provided when needed
	if requiresTerraformModule() && terraformModule == "" {
		return fmt.Errorf("--terraform-module is required when inject-type is %s", injectType)
	}
	return nil
}

func requiresTerraformModule() bool {
	return injectType == injectTypeTerraform || injectType == injectTypeBoth
}

// injectMarkdown handles markdown injection logic.
func injectMarkdown(schemaBasePath, markdownPath string, cfg *config.Config) error {
	logger.Log.Info("injecting into markdown", "path", markdownPath)

	// Verify markdown file exists
	if _, statErr := os.Stat(markdownPath); statErr != nil {
		return fmt.Errorf("markdown file not found: %s", markdownPath)
	}
	logger.Log.Debug("markdown file found", "path", markdownPath)

	injector := markdown.NewInjector()
	markers, err := findAndValidateMarkers(injector, markdownPath)
	if err != nil {
		return err
	}
	if len(markers) == 0 {
		return nil
	}

	// Create renderer with template config from configuration
	renderer := markdown.NewRendererWithTemplate(cfg.MarkdownTemplate)
	reader := yamlio.NewReader(schemaBasePath)
	successCount := processInjectMarkers(markers, markdownPath, renderer, injector, reader)
	printInjectSummary("markdown", successCount, len(markers))
	return nil
}

// injectTerraform handles Terraform injection logic.
func injectTerraform(schemaBasePath, terraformPath string, cfg *config.Config) error {
	logger.Log.Info("injecting into Terraform", "path", terraformPath)

	// Verify terraform module directory exists
	if _, statErr := os.Stat(terraformPath); statErr != nil {
		return fmt.Errorf("terraform module directory not found: %s", terraformPath)
	}
	logger.Log.Debug("terraform module found", "path", terraformPath)

	tfInjector := hclparse.NewTerraformInjector(terraformPath)
	markers, err := tfInjector.FindMarkers()
	if err != nil {
		return fmt.Errorf("failed to find markers in Terraform files: %w", err)
	}

	if len(markers) == 0 {
		logger.Log.Warn("no MARINATED markers found in Terraform variables",
			"path", terraformPath,
			"help", "Add <!-- MARINATED: variable_name --> to variable descriptions")
		return nil
	}

	logger.Log.Info("found markers in Terraform", "count", len(markers))

	// Create renderer with template config from configuration
	renderer := markdown.NewRendererWithTemplate(cfg.MarkdownTemplate)
	reader := yamlio.NewReader(schemaBasePath)
	successCount := processTerraformMarkers(markers, tfInjector, renderer, reader)
	printInjectSummary("Terraform", successCount, len(markers))
	return nil
}

// processTerraformMarkers processes each marker for Terraform injection.
func processTerraformMarkers(
	markers []string,
	tfInjector *hclparse.TerraformInjector,
	renderer *markdown.Renderer,
	reader *yamlio.Reader,
) int {
	successCount := 0
	for _, markerID := range markers {
		if processTerraformMarker(markerID, tfInjector, renderer, reader) {
			successCount++
		}
	}
	return successCount
}

// processTerraformMarker processes a single marker for Terraform injection.
func processTerraformMarker(
	markerID string,
	tfInjector *hclparse.TerraformInjector,
	renderer *markdown.Renderer,
	reader *yamlio.Reader,
) bool {
	logger.Log.Debug("injecting Terraform documentation", "marker", markerID)

	// Find the file containing this variable
	filePath, _, err := tfInjector.FindVariableFile(markerID)
	if err != nil {
		logger.Log.Warn("could not find variable file", "marker", markerID, "error", err)
		return false
	}

	schema, err := reader.ReadSchema(markerID)
	if err != nil {
		logger.Log.Warn("could not read schema", "marker", markerID, "error", err)
		return false
	}

	if schema == nil {
		logger.Log.Warn("no schema found",
			"marker", markerID,
			"help", "Run 'marinatemd export' first to generate YAML schemas")
		return false
	}

	renderedMarkdown, err := renderer.RenderSchema(schema)
	if err != nil {
		logger.Log.Warn("could not render markdown", "marker", markerID, "error", err)
		return false
	}

	if injectErr := tfInjector.InjectIntoFile(filePath, markerID, renderedMarkdown); injectErr != nil {
		logger.Log.Warn("could not inject into Terraform file", "marker", markerID, "error", injectErr)
		return false
	}

	logger.Log.Info("injected Terraform documentation", "marker", markerID, "file", filepath.Base(filePath))
	return true
}

// resolveInjectPaths determines the schema path and markdown file path based on arguments and flags.
// The schema path points directly to the directory containing YAML schema files.
// Returns: (schemaPath, markdownPath, terraformPath, error).
func resolveInjectPaths(args []string) (string, string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get current directory: %w", err)
	}
	logger.Log.Debug("current working directory", "path", cwd)

	schemaBasePath, err := resolveSchemaBasePath(cwd, args)
	if err != nil {
		return "", "", "", err
	}

	markdownPath := resolveMarkdownPath(cwd)
	terraformPath, err := resolveTerraformPath(cwd)
	if err != nil {
		return "", "", "", err
	}

	return schemaBasePath, markdownPath, terraformPath, nil
}

func resolveSchemaBasePath(cwd string, args []string) (string, error) {
	var schemaPath string
	if len(args) > 0 {
		// User provided a path - use it as-is (expecting it to point to directory with YAML files)
		if filepath.IsAbs(args[0]) {
			schemaPath = args[0]
		} else {
			schemaPath = filepath.Join(cwd, args[0])
		}
		logger.Log.Debug("using schema path from argument", "input", args[0], "resolved", schemaPath)
	} else {
		// Default: use ./docs/variables
		schemaPath = filepath.Join(cwd, "docs", "variables")
		logger.Log.Debug("using default schema path", "path", schemaPath)
	}

	// Verify schema directory exists
	if _, statErr := os.Stat(schemaPath); statErr != nil {
		return "", fmt.Errorf(
			"schema directory not found: %s\n   "+
				"Ensure YAML schema files exist or run 'marinatemd export' first",
			schemaPath)
	}
	logger.Log.Debug("found schema directory", "path", schemaPath)

	// Important: yamlio.Reader expects parent directory and appends "variables"
	// If the path ends with "variables", use the parent directory
	if filepath.Base(schemaPath) == "variables" {
		parentPath := filepath.Dir(schemaPath)
		logger.Log.Debug(
			"schema path ends with 'variables', using parent directory",
			"input",
			schemaPath,
			"parent",
			parentPath,
		)
		return parentPath, nil
	}

	// Otherwise, assume the path is the parent directory containing a "variables" subdirectory
	logger.Log.Debug("using path as parent directory (expects variables/ subdirectory)", "path", schemaPath)
	return schemaPath, nil
}

func resolveMarkdownPath(cwd string) string {
	if injectType != injectTypeMarkdown && injectType != injectTypeBoth {
		return ""
	}

	if markdownFile != "" {
		if filepath.IsAbs(markdownFile) {
			logger.Log.Debug("using markdown file from flag", "input", markdownFile, "resolved", markdownFile)
			return markdownFile
		}
		resolved := filepath.Join(cwd, markdownFile)
		logger.Log.Debug("using markdown file from flag", "input", markdownFile, "resolved", resolved)
		return resolved
	}

	defaultPath := filepath.Join(cwd, "README.md")
	logger.Log.Debug("using default markdown file", "path", defaultPath)
	return defaultPath
}

func resolveTerraformPath(cwd string) (string, error) {
	if injectType != injectTypeTerraform && injectType != injectTypeBoth {
		return "", nil
	}

	if terraformModule == "" {
		return "", fmt.Errorf("--terraform-module is required when inject-type is %s", injectType)
	}

	if filepath.IsAbs(terraformModule) {
		logger.Log.Debug("using terraform module from flag", "input", terraformModule, "resolved", terraformModule)
		return terraformModule, nil
	}

	resolved := filepath.Join(cwd, terraformModule)
	logger.Log.Debug("using terraform module from flag", "input", terraformModule, "resolved", resolved)
	return resolved, nil
}

func findAndValidateMarkers(injector *markdown.Injector, markdownPath string) ([]string, error) {
	logger.Log.Debug("scanning for MARINATED markers", "file", markdownPath)
	markers, err := injector.FindMarkers(markdownPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find markers in documentation file: %w", err)
	}

	if len(markers) == 0 {
		logger.Log.Warn("no MARINATED markers found in documentation",
			"file", markdownPath,
			"help", "Add <!-- MARINATED: variable_name --> to your documentation")
		return nil, nil
	}

	logger.Log.Info("found markers", "count", len(markers), "file", filepath.Base(markdownPath))
	return markers, nil
}

func processInjectMarkers(
	markers []string,
	markdownPath string,
	renderer *markdown.Renderer,
	injector *markdown.Injector,
	reader *yamlio.Reader,
) int {
	successCount := 0
	for _, markerID := range markers {
		if processMarker(markerID, markdownPath, renderer, injector, reader) {
			successCount++
		}
	}
	return successCount
}

func processMarker(
	markerID, markdownPath string,
	renderer *markdown.Renderer,
	injector *markdown.Injector,
	reader *yamlio.Reader,
) bool {
	logger.Log.Debug("injecting documentation", "marker", markerID)

	schema, err := reader.ReadSchema(markerID)
	if err != nil {
		logger.Log.Warn("could not read schema", "marker", markerID, "error", err)
		return false
	}

	if schema == nil {
		logger.Log.Warn("no schema found",
			"marker", markerID,
			"help", "Run 'marinatemd export' first to generate YAML schemas")
		return false
	}

	renderedMarkdown, err := renderer.RenderSchema(schema)
	if err != nil {
		logger.Log.Warn("could not render markdown", "marker", markerID, "error", err)
		return false
	}

	if injectErr := injector.InjectIntoFile(markdownPath, markerID, renderedMarkdown); injectErr != nil {
		logger.Log.Warn("could not inject markdown", "marker", markerID, "error", injectErr)
		return false
	}

	logger.Log.Info("injected documentation", "marker", markerID)
	return true
}

func printInjectSummary(targetType string, successCount, totalCount int) {
	logger.Log.Info("injection complete",
		"target", targetType,
		"success", successCount,
		"total", totalCount)
	if successCount < totalCount {
		logger.Log.Warn("some variables were not injected",
			"help", "Run 'marinatemd export' to generate missing YAML schemas")
	}
}
