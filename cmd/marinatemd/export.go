package marinatemd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"github.com/c4a8-azure/marinatemd/internal/logger"
	"github.com/c4a8-azure/marinatemd/internal/paths"
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
	moduleRoot, cfg, err := paths.SetupEnvironment(args)
	if err != nil {
		return err
	}

	// Resolve paths using config
	exportPath := paths.ResolveExportPath(moduleRoot, cfg)

	logger.Log.Info("exporting variables", "moduleRoot", moduleRoot, "exportPath", exportPath)

	marinatedVars, err := parseAndExtractVariables(moduleRoot)
	if err != nil {
		return err
	}

	variablesDir := filepath.Join(exportPath, "variables")
	if mkdirErr := os.MkdirAll(variablesDir, 0750); mkdirErr != nil {
		return fmt.Errorf("failed to create variables directory: %w", mkdirErr)
	}

	if processErr := processMarinatedVariables(marinatedVars, exportPath, variablesDir); processErr != nil {
		return processErr
	}

	printExportSummary(len(marinatedVars), variablesDir)
	return nil
}

func parseAndExtractVariables(variablesPath string) ([]*hclparse.Variable, error) {
	logger.Log.Debug("parsing terraform variables", "path", variablesPath)
	parser := hclparse.NewParser()
	if err := parser.ParseVariables(variablesPath); err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	marinatedVars, err := parser.ExtractMarinatedVars()
	if err != nil {
		return nil, fmt.Errorf("failed to extract marinated variables: %w", err)
	}

	if len(marinatedVars) == 0 {
		logger.Log.Warn("no MARINATED variables found in module",
			"help", "Add <!-- MARINATED: variable_name --> to variable descriptions to enable documentation")
		return nil, nil
	}

	logger.Log.Info("found marinated variables", "count", len(marinatedVars))
	return marinatedVars, nil
}

func processMarinatedVariables(marinatedVars []*hclparse.Variable, docsPath, variablesDir string) error {
	builder := schema.NewBuilder()
	reader := yamlio.NewReader(docsPath)
	writer := yamlio.NewWriter(docsPath)

	logger.Log.Debug("processing variables", "count", len(marinatedVars))
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
	logger.Log.Debug("processing variable", "name", variable.Name, "id", variable.MarinatedID)

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

	logger.Log.Info("exported variable", "name", variable.Name, "path", yamlPath)
	return nil
}

func mergeOrUseNewSchema(newSchema, existingSchema *schema.Schema, builder *schema.Builder) (*schema.Schema, error) {
	if existingSchema != nil {
		logger.Log.Debug("merging with existing schema", "variable", newSchema.Variable)
		return builder.MergeWithExisting(newSchema, existingSchema)
	}
	logger.Log.Debug("creating new schema", "variable", newSchema.Variable)
	return newSchema, nil
}

func printExportSummary(count int, variablesDir string) {
	logger.Log.Info("export complete", "count", count, "directory", variablesDir)
}
