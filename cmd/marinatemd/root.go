package marinatemd

import (
	"os"

	"github.com/glueckkanja/marinatemd/internal/config"
	"github.com/glueckkanja/marinatemd/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	debug   bool
)

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
  marinatemd inject --docs-file docs/VARIABLES.md .`,
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
	cobra.OnInitialize(initLogger, initConfig)

	// Set viper defaults before config file is read
	config.SetDefaults()

	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		"config file (default searches for .marinated.yml in multiple locations)",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"enable verbose output (info level)",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&debug,
		"debug",
		"d",
		false,
		"enable debug output (most detailed)",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		logger.Log.Debug("using config file from flag", "path", cfgFile)
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		logger.Log.Debug("searching for config file in default locations")
		// Search for config in multiple locations
		viper.SetConfigName(".marinated")
		viper.SetConfigType("yml")

		// Current directory
		viper.AddConfigPath(".")
		viper.AddConfigPath(".config")
		logger.Log.Debug("added config search paths", "paths", "., .config")

		// Home directory
		if home, err := os.UserHomeDir(); err == nil {
			homePath := home + "/.marinated.d"
			viper.AddConfigPath(homePath)
			logger.Log.Debug("added home config path", "path", homePath)
		}
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("MARINATED")
	viper.AutomaticEnv()
	logger.Log.Debug("environment variables enabled", "prefix", "MARINATED")

	// If a config file is found, read it in (ignore errors silently).
	if err := viper.ReadInConfig(); err == nil {
		logger.Log.Debug("config file loaded", "path", viper.ConfigFileUsed())
	} else {
		logger.Log.Debug("no config file found, using defaults", "error", err)
	}
}

// initLogger sets up the logger based on verbose and debug flags.
func initLogger() {
	logger.Setup(verbose, debug)
}
