package marinatemd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"github.com/c4a8-azure/marinatemd/internal/schema"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command that parses HCL and generates/merges YAML schemas.
var exportCmd = &cobra.Command{
	Use:   "export [module-path]",
	Short: "Export Terraform variables to YAML schema files",
	Long: `Parse Terraform/OpenTofu variables marked with MARINATED comments and
generate or update YAML schema files in the docs/variables/ directory.

This command:
  1. Parses variables.tf files for MARINATED markers
  2. Generates structured YAML schemas for complex variable types
  3. Merges with existing YAML files to preserve user descriptions
  4. Creates new YAML files for newly discovered variables

Example:
  marinatemd export .
  marinatemd export /path/to/terraform/module
  marinatemd export --config .marinated.yml .`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExport(_ *cobra.Command, args []string) error {
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
	fmt.Printf("Exporting variables from: %s\n", absRoot)

	// Resolve docs path relative to module root
	docsPath := filepath.Join(absRoot, cfg.DocsPath)
	fmt.Printf("Documentation path: %s\n", docsPath)

	// Step 1: Parse HCL variables from the module
	fmt.Println("\nParsing Terraform variables...")
	parser := hclparse.NewParser()
	if parseErr := parser.ParseVariables(absRoot); parseErr != nil {
		return fmt.Errorf("failed to parse variables: %w", parseErr)
	}

	// Step 2: Extract variables marked with MARINATED
	marinatedVars, extractErr := parser.ExtractMarinatedVars()
	if extractErr != nil {
		return fmt.Errorf("failed to extract marinated variables: %w", extractErr)
	}

	if len(marinatedVars) == 0 {
		fmt.Println("WARNING: No MARINATED variables found in module")
		fmt.Println("   Add <!-- MARINATED: variable_name --> to variable descriptions to enable documentation")
		return nil
	}

	fmt.Printf("Found %d MARINATED variable(s)\n", len(marinatedVars))

	// Step 3: Create docs/variables/ directory structure
	variablesDir := filepath.Join(docsPath, "variables")
	fmt.Printf("\nCreating directory structure: %s\n", variablesDir)
	if mkdirErr := os.MkdirAll(variablesDir, 0750); mkdirErr != nil {
		return fmt.Errorf("failed to create variables directory: %w", mkdirErr)
	}

	// Step 4: Process each MARINATED variable
	builder := schema.NewBuilder()
	reader := yamlio.NewReader(docsPath)
	writer := yamlio.NewWriter(docsPath)

	fmt.Println("\nProcessing variables...")
	for _, variable := range marinatedVars {
		fmt.Printf("\n  Processing '%s' (ID: %s)...\n", variable.Name, variable.MarinatedID)

		// Build schema from HCL variable
		newSchema, buildErr := builder.BuildFromVariable(variable)
		if buildErr != nil {
			return fmt.Errorf("failed to build schema for variable %s: %w", variable.Name, buildErr)
		}

		// Check if YAML schema already exists
		existingSchema, readErr := reader.ReadSchema(variable.MarinatedID)
		if readErr != nil {
			return fmt.Errorf("failed to read existing schema for %s: %w", variable.MarinatedID, readErr)
		}

		var finalSchema *schema.Schema
		if existingSchema != nil {
			// Merge new schema with existing to preserve user descriptions
			fmt.Printf("    Merging with existing schema...\n")
			var mergeErr error
			finalSchema, mergeErr = builder.MergeWithExisting(newSchema, existingSchema)
			if mergeErr != nil {
				return fmt.Errorf("failed to merge schemas for %s: %w", variable.MarinatedID, mergeErr)
			}
		} else {
			// No existing schema, use new one
			fmt.Printf("    Creating new schema...\n")
			finalSchema = newSchema
		}

		// Write the schema to YAML file
		yamlPath := filepath.Join(variablesDir, variable.MarinatedID+".yaml")
		if writeErr := writer.WriteSchema(finalSchema); writeErr != nil {
			return fmt.Errorf("failed to write schema for %s: %w", variable.MarinatedID, writeErr)
		}

		fmt.Printf("    Written to %s\n", yamlPath)
	}

	// Success summary
	fmt.Printf("\nSuccessfully exported %d variable(s)\n", len(marinatedVars))
	fmt.Printf("   YAML schemas written to: %s\n", variablesDir)
	fmt.Println("\nNext steps:")
	fmt.Println("   1. Review and edit the generated YAML files to add descriptions")
	fmt.Println("   2. Run 'marinatemd inject' to update documentation")

	return nil
}
