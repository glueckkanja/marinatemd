package marinatemd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/c4a8-azure/marinatemd/internal/markdown"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
	"github.com/spf13/cobra"
)

var (
	readmeFile string
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
  marinatemd inject --readme docs/VARIABLES.md .`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInject,
}

func init() {
	rootCmd.AddCommand(injectCmd)

	injectCmd.Flags().StringVar(
		&readmeFile,
		"readme",
		"README.md",
		"documentation file to inject into (relative to docs path)",
	)
}

func runInject(_ *cobra.Command, args []string) error {
	absRoot, cfg, readmePath, err := setupInjectEnvironment(args)
	if err != nil {
		return err
	}

	docsPath := filepath.Join(absRoot, cfg.DocsPath)
	injector := markdown.NewInjector()

	markers, err := findAndValidateMarkers(injector, readmePath)
	if err != nil {
		return err
	}
	if len(markers) == 0 {
		return nil
	}

	renderer := markdown.NewRenderer()
	reader := yamlio.NewReader(docsPath)
	successCount := processInjectMarkers(markers, readmePath, renderer, injector, reader)
	printInjectSummary(successCount, len(markers))
	return nil
}

func setupInjectEnvironment(args []string) (string, *config.Config, string, error) {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to resolve module path: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Injecting documentation into: %s\n", absRoot)
	docsPath := filepath.Join(absRoot, cfg.DocsPath)
	fmt.Printf("Documentation path: %s\n", docsPath)

	readmePath := filepath.Join(docsPath, readmeFile)
	if _, statErr := os.Stat(readmePath); statErr != nil {
		return "", nil, "", fmt.Errorf(
			"documentation file not found: %s\n   Use --readme flag to specify a different file",
			readmePath)
	}

	fmt.Printf("Target file: %s\n", readmePath)
	return absRoot, cfg, readmePath, nil
}

func findAndValidateMarkers(injector *markdown.Injector, readmePath string) ([]string, error) {
	fmt.Println("\nInjecting markdown into documentation...")
	markers, err := injector.FindMarkers(readmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to find markers in documentation file: %w", err)
	}

	if len(markers) == 0 {
		fmt.Printf("   No MARINATED markers found in %s\n", readmePath)
		fmt.Println("\nAdd markers to your documentation file:")
		fmt.Println("   Description: <!-- MARINATED: variable_name -->")
		return nil, nil
	}

	fmt.Printf("   Found %d marker(s) in %s\n", len(markers), readmeFile)
	return markers, nil
}

func processInjectMarkers(
	markers []string,
	readmePath string,
	renderer *markdown.Renderer,
	injector *markdown.Injector,
	reader *yamlio.Reader,
) int {
	successCount := 0
	for _, markerID := range markers {
		if processMarker(markerID, readmePath, renderer, injector, reader) {
			successCount++
		}
	}
	return successCount
}

func processMarker(
	markerID, readmePath string,
	renderer *markdown.Renderer,
	injector *markdown.Injector,
	reader *yamlio.Reader,
) bool {
	fmt.Printf("   Injecting documentation for '%s'...\n", markerID)

	schema, err := reader.ReadSchema(markerID)
	if err != nil {
		fmt.Printf("      WARNING: Could not read schema for %s: %v\n", markerID, err)
		return false
	}

	if schema == nil {
		fmt.Printf("      WARNING: No schema found for %s\n", markerID)
		fmt.Printf("             Run 'marinatemd export' first to generate YAML schemas\n")
		return false
	}

	renderedMarkdown, err := renderer.RenderSchema(schema)
	if err != nil {
		fmt.Printf("      WARNING: Could not render markdown for %s: %v\n", markerID, err)
		return false
	}

	if injectErr := injector.InjectIntoFile(readmePath, markerID, renderedMarkdown); injectErr != nil {
		fmt.Printf("      WARNING: Could not inject markdown for %s: %v\n", markerID, injectErr)
		return false
	}

	fmt.Printf("      ✓ Injected successfully\n")
	return true
}

func printInjectSummary(successCount, totalCount int) {
	fmt.Printf("\nSuccessfully injected %d of %d variable(s)\n", successCount, totalCount)
	if successCount < totalCount {
		fmt.Println("\nSome variables were not injected. Check warnings above.")
		fmt.Println("   Run 'marinatemd export' to generate missing YAML schemas")
	}
}
