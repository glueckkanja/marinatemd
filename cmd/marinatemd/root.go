package marinatemd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "marinatemd",
	Short: "Generate structured documentation for complex Terraform/OpenTofu variables",
	Long: `MarinateMD parses Terraform/OpenTofu variables marked with MARINATED comments,
generates or updates YAML schema files, and injects hierarchical markdown documentation
into README.md or other documentation files.

Commands:
  export  - Parse HCL variables and generate/merge YAML schema files
  inject  - Read YAML schemas and inject markdown into documentation

Example:
  marinatemd export .
  marinatemd inject .
  marinatemd export /path/to/terraform/module
  marinatemd inject --readme docs/VARIABLES.md .`,
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
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in multiple locations
		viper.SetConfigName(".marinated")
		viper.SetConfigType("yml")

		// Current directory
		viper.AddConfigPath(".")
		viper.AddConfigPath(".config")

		// Home directory
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home + "/.marinated.d")
		}
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("MARINATED")
	viper.AutomaticEnv()

	// If a config file is found, read it in (ignore errors silently).
	_ = viper.ReadInConfig()
}
