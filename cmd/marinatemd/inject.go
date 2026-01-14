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
	// Determine module root
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	// Convert to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("failed to resolve module path: %w", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Print configuration info
	fmt.Printf("Injecting documentation into: %s\n", absRoot)

	// Resolve docs path relative to module root
	docsPath := filepath.Join(absRoot, cfg.DocsPath)
	fmt.Printf("Documentation path: %s\n", docsPath)

	// Resolve README path
	readmePath := filepath.Join(docsPath, readmeFile)
	if _, err := os.Stat(readmePath); err != nil {
		return fmt.Errorf("documentation file not found: %s\n   Use --readme flag to specify a different file", readmePath)
	}

	fmt.Printf("Target file: %s\n", readmePath)

	// Initialize markdown renderer and injector
	fmt.Println("\nInjecting markdown into documentation...")
	renderer := markdown.NewRenderer()
	injector := markdown.NewInjector()
	reader := yamlio.NewReader(docsPath)

	// Find all markers in the README
	markers, findErr := injector.FindMarkers(readmePath)
	if findErr != nil {
		return fmt.Errorf("failed to find markers in documentation file: %w", findErr)
	}

	if len(markers) == 0 {
		fmt.Printf("   No MARINATED markers found in %s\n", readmePath)
		fmt.Println("\nAdd markers to your documentation file:")
		fmt.Println("   Description: <!-- MARINATED: variable_name -->")
		return nil
	}

	fmt.Printf("   Found %d marker(s) in %s\n", len(markers), readmeFile)

	// Process each marker
	successCount := 0
	for _, markerID := range markers {
		fmt.Printf("   Injecting documentation for '%s'...\n", markerID)

		// Read the schema from YAML
		schema, readErr := reader.ReadSchema(markerID)
		if readErr != nil {
			fmt.Printf("      WARNING: Could not read schema for %s: %v\n", markerID, readErr)
			continue
		}

		if schema == nil {
			fmt.Printf("      WARNING: No schema found for %s\n", markerID)
			fmt.Printf("             Run 'marinatemd export' first to generate YAML schemas\n")
			continue
		}

		// Render markdown from schema
		renderedMarkdown, renderErr := renderer.RenderSchema(schema)
		if renderErr != nil {
			fmt.Printf("      WARNING: Could not render markdown for %s: %v\n", markerID, renderErr)
			continue
		}

		// Inject into README
		if injectErr := injector.InjectIntoFile(readmePath, markerID, renderedMarkdown); injectErr != nil {
			fmt.Printf("      WARNING: Could not inject markdown for %s: %v\n", markerID, injectErr)
			continue
		}

		fmt.Printf("      ✓ Injected successfully\n")
		successCount++
	}

	// Success summary
	fmt.Printf("\nSuccessfully injected %d of %d variable(s)\n", successCount, len(markers))
	if successCount < len(markers) {
		fmt.Println("\nSome variables were not injected. Check warnings above.")
		fmt.Println("   Run 'marinatemd export' to generate missing YAML schemas")
	}

	return nil
}
