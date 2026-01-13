package config

import (
	"github.com/spf13/viper"
)

// Config represents the application configuration loaded from .marinated.yml
type Config struct {
	// Mock configuration fields for initialization
	// These will be expanded as features are implemented

	// DocsPath is the path where documentation and YAML schemas are stored
	DocsPath string `mapstructure:"docs_path"`

	// VariablesPath is the relative path to find variables.*.tf files
	VariablesPath string `mapstructure:"variables_path"`

	// ReadmePath is the path to the README or documentation file to inject into
	ReadmePath string `mapstructure:"readme_path"`

	// Verbose enables verbose logging
	Verbose bool `mapstructure:"verbose"`
}

// Load returns the configuration loaded from viper
// This should be called after Viper has been initialized by Cobra
func Load() (*Config, error) {
	cfg := &Config{
		// Set defaults
		DocsPath:      "docs",
		VariablesPath: ".",
		ReadmePath:    "README.md",
		Verbose:       false,
	}

	// Unmarshal viper config into struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SetDefaults configures default values for viper
// This should be called during initialization before config file is read
func SetDefaults() {
	viper.SetDefault("docs_path", "docs")
	viper.SetDefault("variables_path", ".")
	viper.SetDefault("readme_path", "README.md")
	viper.SetDefault("verbose", false)
}
