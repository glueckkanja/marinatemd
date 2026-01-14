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
	absRoot, cfg, err := setupExportEnvironment(args)
	if err != nil {
		return err
	}

	marinatedVars, err := parseAndExtractVariables(absRoot)
	if err != nil {
		return err
	}

	docsPath := filepath.Join(absRoot, cfg.DocsPath)
	variablesDir := filepath.Join(docsPath, "variables")
	if mkdirErr := os.MkdirAll(variablesDir, 0750); mkdirErr != nil {
		return fmt.Errorf("failed to create variables directory: %w", mkdirErr)
	}

	if processErr := processMarinatedVariables(marinatedVars, docsPath, variablesDir); processErr != nil {
		return processErr
	}

	printExportSummary(len(marinatedVars), variablesDir)
	return nil
}

func setupExportEnvironment(args []string) (string, *config.Config, error) {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve module path: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return "", nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Exporting variables from: %s\n", absRoot)
	fmt.Printf("Documentation path: %s\n", filepath.Join(absRoot, cfg.DocsPath))

	return absRoot, cfg, nil
}

func parseAndExtractVariables(absRoot string) ([]*hclparse.Variable, error) {
	fmt.Println("\nParsing Terraform variables...")
	parser := hclparse.NewParser()
	if err := parser.ParseVariables(absRoot); err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	marinatedVars, err := parser.ExtractMarinatedVars()
	if err != nil {
		return nil, fmt.Errorf("failed to extract marinated variables: %w", err)
	}

	if len(marinatedVars) == 0 {
		fmt.Println("WARNING: No MARINATED variables found in module")
		fmt.Println("   Add <!-- MARINATED: variable_name --> to variable descriptions to enable documentation")
		return nil, nil
	}

	fmt.Printf("Found %d MARINATED variable(s)\n", len(marinatedVars))
	return marinatedVars, nil
}

func processMarinatedVariables(marinatedVars []*hclparse.Variable, docsPath, variablesDir string) error {
	builder := schema.NewBuilder()
	reader := yamlio.NewReader(docsPath)
	writer := yamlio.NewWriter(docsPath)

	fmt.Println("\nProcessing variables...")
	for _, variable := range marinatedVars {
		if err := processVariable(variable, builder, reader, writer, variablesDir); err != nil {
			return err
		}
	}
	return nil
}

func processVariable(
	variable *hclparse.Variable,
	builder *schema.Builder,
	reader *yamlio.Reader,
	writer *yamlio.Writer,
	variablesDir string,
) error {
	fmt.Printf("\n  Processing '%s' (ID: %s)...\n", variable.Name, variable.MarinatedID)

	newSchema, err := builder.BuildFromVariable(variable)
	if err != nil {
		return fmt.Errorf("failed to build schema for variable %s: %w", variable.Name, err)
	}

	existingSchema, err := reader.ReadSchema(variable.MarinatedID)
	if err != nil {
		return fmt.Errorf("failed to read existing schema for %s: %w", variable.MarinatedID, err)
	}

	finalSchema, err := mergeOrUseNewSchema(newSchema, existingSchema, builder)
	if err != nil {
		return fmt.Errorf("failed to merge schemas for %s: %w", variable.MarinatedID, err)
	}

	yamlPath := filepath.Join(variablesDir, variable.MarinatedID+".yaml")
	if writeErr := writer.WriteSchema(finalSchema); writeErr != nil {
		return fmt.Errorf("failed to write schema for %s: %w", variable.MarinatedID, writeErr)
	}

	fmt.Printf("    Written to %s\n", yamlPath)
	return nil
}

func mergeOrUseNewSchema(newSchema, existingSchema *schema.Schema, builder *schema.Builder) (*schema.Schema, error) {
	if existingSchema != nil {
		fmt.Printf("    Merging with existing schema...\n")
		return builder.MergeWithExisting(newSchema, existingSchema)
	}
	fmt.Printf("    Creating new schema...\n")
	return newSchema, nil
}

func printExportSummary(count int, variablesDir string) {
	fmt.Printf("\nSuccessfully exported %d variable(s)\n", count)
	fmt.Printf("   YAML schemas written to: %s\n", variablesDir)
	fmt.Println("\nNext steps:")
	fmt.Println("   1. Review and edit the generated YAML files to add descriptions")
	fmt.Println("   2. Run 'marinatemd inject' to update documentation")
}
