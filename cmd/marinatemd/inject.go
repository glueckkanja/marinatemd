package marinatemd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/c4a8-azure/marinatemd/internal/logger"
	"github.com/c4a8-azure/marinatemd/internal/markdown"
	"github.com/c4a8-azure/marinatemd/internal/paths"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
	"github.com/spf13/cobra"
)

var (
	docsFile string
)

// injectCmd represents the inject command that reads YAML schemas and injects markdown into documentation.
var injectCmd = &cobra.Command{
	Use:   "inject [module-path]",
	Short: "Inject YAML schema documentation into markdown files",
	Long: `Read YAML schema files from docs/variables/ and inject rendered markdown
documentation into README.md or other documentation files at MARINATED markers.

This command:
  1. Scans the documentation file for <!-- MARINATED: variable_name --> markers
  2. Reads corresponding YAML schema files
  3. Renders hierarchical markdown from the schemas
  4. Injects the markdown between start and end markers
  5. Preserves all other content in the documentation

Example:
  marinatemd inject .
  marinatemd inject /path/to/terraform/module
  marinatemd inject --docs-file docs/VARIABLES.md .`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInject,
}

func init() {
	rootCmd.AddCommand(injectCmd)

	injectCmd.Flags().StringVar(
		&docsFile,
		"docs-file",
		"README.md",
		"documentation file to inject into (relative to docs path)",
	)
}

func runInject(_ *cobra.Command, args []string) error {
	moduleRoot, cfg, docsFilePath, err := setupInjectEnvironment(args)
	if err != nil {
		return err
	}

	exportPath := paths.ResolveExportPath(moduleRoot, cfg)
	injector := markdown.NewInjector()

	markers, err := findAndValidateMarkers(injector, docsFilePath)
	if err != nil {
		return err
	}
	if len(markers) == 0 {
		return nil
	}

	renderer := markdown.NewRenderer()
	reader := yamlio.NewReader(exportPath)
	successCount := processInjectMarkers(markers, docsFilePath, renderer, injector, reader)
	printInjectSummary(successCount, len(markers))
	return nil
}

func setupInjectEnvironment(args []string) (string, *config.Config, string, error) {
	moduleRoot, cfg, err := paths.SetupEnvironment(args)
	if err != nil {
		return "", nil, "", err
	}

	exportPath := paths.ResolveExportPath(moduleRoot, cfg)

	logger.Log.Info("injecting documentation", "moduleRoot", moduleRoot, "exportPath", exportPath)

	docsFilePath := filepath.Join(exportPath, docsFile)
	if _, statErr := os.Stat(docsFilePath); statErr != nil {
		return "", nil, "", fmt.Errorf(
			"documentation file not found: %s\n   Use --docs-file flag to specify a different file",
			docsFilePath)
	}

	logger.Log.Debug("target documentation file", "path", docsFilePath)
	return moduleRoot, cfg, docsFilePath, nil
}

func findAndValidateMarkers(injector *markdown.Injector, docsFilePath string) ([]string, error) {
	logger.Log.Debug("scanning for MARINATED markers", "file", docsFilePath)
	markers, err := injector.FindMarkers(docsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to find markers in documentation file: %w", err)
	}

	if len(markers) == 0 {
		logger.Log.Warn("no MARINATED markers found in documentation",
			"file", docsFilePath,
			"help", "Add <!-- MARINATED: variable_name --> to your documentation")
		return nil, nil
	}

	logger.Log.Info("found markers", "count", len(markers), "file", docsFile)
	return markers, nil
}

func processInjectMarkers(
	markers []string,
	docsFilePath string,
	renderer *markdown.Renderer,
	injector *markdown.Injector,
	reader *yamlio.Reader,
) int {
	successCount := 0
	for _, markerID := range markers {
		if processMarker(markerID, docsFilePath, renderer, injector, reader) {
			successCount++
		}
	}
	return successCount
}

func processMarker(
	markerID, docsFilePath string,
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

	if injectErr := injector.InjectIntoFile(docsFilePath, markerID, renderedMarkdown); injectErr != nil {
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
