package marinatemd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"github.com/c4a8-azure/marinatemd/internal/markdown"
	"github.com/c4a8-azure/marinatemd/internal/schema"
	"github.com/c4a8-azure/marinatemd/internal/yamlio"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	moduleRoot string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "marinatemd [module-path]",
	Short: "Generate structured documentation for complex Terraform/OpenTofu variables",
	Long: `MarinateMD parses Terraform/OpenTofu variables marked with MARINATED comments,
generates or updates YAML schema files, and injects hierarchical markdown documentation
into README.md or other documentation files.

Example:
  marinatemd .
  marinatemd /path/to/terraform/module
  marinatemd --config .marinated.yml .`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
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
		fmt.Printf("Processing module at: %s\n", absRoot)
		if viper.ConfigFileUsed() != "" {
			fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())
		}

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

		// Step 5: Find README.md or configured documentation file
		readmePath := filepath.Join(docsPath, "README.md")
		if _, err := os.Stat(readmePath); err != nil {
			fmt.Printf("\nWARNING: README.md not found at %s\n", readmePath)
			fmt.Printf("         Skipping markdown injection step\n")
		} else {
			// Step 6: Inject rendered markdown into README.md
			fmt.Printf("\nInjecting markdown into documentation...\n")
			renderer := markdown.NewRenderer()
			injector := markdown.NewInjector()

			// Find all markers in the README
			markers, findErr := injector.FindMarkers(readmePath)
			if findErr != nil {
				return fmt.Errorf("failed to find markers in README: %w", findErr)
			}

			if len(markers) == 0 {
				fmt.Printf("   No MARINATED markers found in %s\n", readmePath)
			} else {
				fmt.Printf("   Found %d marker(s) in README.md\n", len(markers))

				// Process each marker
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
				}
			}
		}

		// Success summary
		fmt.Printf("\nSuccessfully processed %d variable(s)\n", len(marinatedVars))
		fmt.Printf("   YAML schemas written to: %s\n", variablesDir)
		fmt.Println("\nNext steps:")
		fmt.Println("   1. Review and edit the generated YAML files to add descriptions")
		fmt.Println("   2. Run marinatemd again to regenerate markdown")

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		"config file (default searches for .marinated.yml in multiple locations)",
	)

	// Local flags (only for root command)
	rootCmd.Flags().StringVar(&moduleRoot, "module-root", ".", "root directory of the Terraform/OpenTofu module")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set defaults first
	config.SetDefaults()

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in multiple locations (following terraform-docs pattern)
		// Priority order:
		// 1. root of module directory (if specified)
		// 2. .config/ folder at root of module directory
		// 3. current directory
		// 4. .config/ folder at current directory
		// 5. $HOME/.marinated.d/

		viper.SetConfigName(".marinated")
		viper.SetConfigType("yml")

		// Current directory
		viper.AddConfigPath(".")
		viper.AddConfigPath(".config")

		// Module root (if different from current)
		if moduleRoot != "" && moduleRoot != "." {
			viper.AddConfigPath(moduleRoot)
			viper.AddConfigPath(moduleRoot + "/.config")
		}

		// Home directory
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home + "/.marinated.d")
		}
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("MARINATED")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	// Try to read the config file (ignore errors silently).
	_ = viper.ReadInConfig()
}
