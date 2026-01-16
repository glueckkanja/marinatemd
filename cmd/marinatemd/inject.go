package marinatemd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/logger"
	"github.com/c4a8-azure/marinatemd/internal/markdown"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
	"github.com/spf13/cobra"
)

var (
	markdownFile string
)

// injectCmd represents the inject command that reads YAML schemas and injects markdown into documentation.
var injectCmd = &cobra.Command{
	Use:   "inject [schema-path]",
	Short: "Inject YAML schema documentation into markdown files",
	Long: `Read YAML schema files and inject rendered markdown documentation into a markdown file.

Arguments:
  [schema-path]  Optional path to directory containing 'variables' subdirectory with YAML schemas.
                 The tool expects schema files at <schema-path>/variables/*.yaml
                 Defaults to <current-dir>/docs

Flags:
  --markdown-file  Path to the markdown file to inject into.
                   Can be absolute or relative to current working directory.
                   Defaults to <current-dir>/README.md

Examples:
  # 1. Use default paths (./docs/variables/*.yaml → ./README.md)
  marinatemd inject

  # 2. Custom schema base path, default markdown file (./README.md)
  #    Reads from /path/to/schemas/variables/*.yaml
  marinatemd inject /path/to/schemas
  marinatemd inject ./docs

  # 3. Custom schema path and custom markdown file
  marinatemd inject /path/to/schemas --markdown-file docs/API.md
  marinatemd inject ./docs --markdown-file /abs/path/to/doc.md`,
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
}

func runInject(_ *cobra.Command, args []string) error {
	schemaBasePath, markdownPath, err := resolveInjectPaths(args)
	if err != nil {
		return err
	}

	logger.Log.Info("injecting documentation", "schemaBasePath", schemaBasePath, "markdownPath", markdownPath)

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

	renderer := markdown.NewRenderer()
	reader := yamlio.NewReader(schemaBasePath)
	successCount := processInjectMarkers(markers, markdownPath, renderer, injector, reader)
	printInjectSummary(successCount, len(markers))
	return nil
}

// resolveInjectPaths determines the schema base path and markdown file path based on arguments and flags.
// The schema base path is the parent directory of 'variables/', not the variables directory itself.
// Returns: (schemaBasePath, markdownPath, error)
func resolveInjectPaths(args []string) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}
	logger.Log.Debug("current working directory", "path", cwd)

	// Determine schema base path (parent of 'variables/')
	var schemaBasePath string
	if len(args) > 0 {
		// Use provided schema base path argument
		if filepath.IsAbs(args[0]) {
			schemaBasePath = args[0]
		} else {
			schemaBasePath = filepath.Join(cwd, args[0])
		}
		logger.Log.Debug("using schema base path from argument", "input", args[0], "resolved", schemaBasePath)
	} else {
		// Default: <cwd>/docs
		schemaBasePath = filepath.Join(cwd, "docs")
		logger.Log.Debug("using default schema base path", "path", schemaBasePath)
	}

	// Verify variables subdirectory exists
	variablesPath := filepath.Join(schemaBasePath, "variables")
	if _, statErr := os.Stat(variablesPath); statErr != nil {
		return "", "", fmt.Errorf("schema directory not found: %s\n   Expected structure: <path>/variables/*.yaml\n   Ensure YAML schema files exist or run 'marinatemd export' first", variablesPath)
	}
	logger.Log.Debug("found variables directory", "path", variablesPath)

	// Determine markdown file path
	var markdownPath string
	if markdownFile != "" {
		// Use provided markdown file flag
		if filepath.IsAbs(markdownFile) {
			markdownPath = markdownFile
		} else {
			markdownPath = filepath.Join(cwd, markdownFile)
		}
		logger.Log.Debug("using markdown file from flag", "input", markdownFile, "resolved", markdownPath)
	} else {
		// Default: <cwd>/README.md
		markdownPath = filepath.Join(cwd, "README.md")
		logger.Log.Debug("using default markdown file", "path", markdownPath)
	}

	return schemaBasePath, markdownPath, nil
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

func printInjectSummary(successCount, totalCount int) {
	logger.Log.Info("injection complete", "success", successCount, "total", totalCount)
	if successCount < totalCount {
		logger.Log.Warn("some variables were not injected",
			"help", "Run 'marinatemd export' to generate missing YAML schemas")
	}
}
