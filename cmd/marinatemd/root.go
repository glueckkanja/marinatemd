package marinatemd

import (
	"fmt"
	"os"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	moduleRoot string
)

// rootCmd represents the base command when called without any subcommands
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine module root
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// TODO: Implement main logic here
		// - Parse HCL variables from root
		// - Generate/update YAML schemas
		// - Generate markdown and inject into docs

		fmt.Printf("Processing module at: %s\n", root)
		if viper.ConfigFileUsed() != "" {
			fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())
		}
		fmt.Printf("Docs path: %s\n", cfg.DocsPath)
		fmt.Printf("README path: %s\n", cfg.ReadmePath)

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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default searches for .marinated.yml in multiple locations)")

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
	if err := viper.ReadInConfig(); err == nil {
		// Config file found and successfully parsed
		// Silently continue - we'll show config file used in verbose mode
	}
}
