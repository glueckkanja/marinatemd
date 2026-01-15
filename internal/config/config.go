package config

import (
	"github.com/c4a8-azure/marinatemd/internal/markdown"
	"github.com/spf13/viper"
)

// Config represents the application configuration loaded from .marinated.yml.
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

	// MarkdownTemplate configures how markdown is generated from schema
	MarkdownTemplate *markdown.TemplateConfig `mapstructure:"markdown_template"`

	// Split configures the split command behavior
	Split *SplitConfig `mapstructure:"split"`
}

// SplitConfig represents configuration for the split command.
type SplitConfig struct {
	// InputPath is the default input markdown file to split (relative to docs_path)
	InputPath string `mapstructure:"input_path"`

	// OutputDir is the default output directory for split files (relative to docs_path)
	OutputDir string `mapstructure:"output_dir"`

	// HeaderFile is the path to the header file to prepend to each split file
	HeaderFile string `mapstructure:"header_file"`

	// FooterFile is the path to the footer file to append to each split file
	FooterFile string `mapstructure:"footer_file"`
}

// Load returns the configuration loaded from viper.
// This should be called after Viper has been initialized by Cobra.
func Load() (*Config, error) {
	cfg := &Config{
		// Set defaults
		DocsPath:         "docs",
		VariablesPath:    ".",
		ReadmePath:       "README.md",
		Verbose:          false,
		MarkdownTemplate: markdown.DefaultTemplateConfig(),
		Split: &SplitConfig{
			InputPath:  "README.md",
			OutputDir:  "variables",
			HeaderFile: "",
			FooterFile: "",
		},
	}

	// Unmarshal viper config into struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Validate markdown template configuration
	if err := cfg.MarkdownTemplate.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SetDefaults configures default values for viper
// SetDefaults sets default configuration values.
// This should be called during initialization before config file is read.
func SetDefaults() {
	viper.SetDefault("docs_path", "docs")
	viper.SetDefault("variables_path", ".")
	viper.SetDefault("readme_path", "README.md")
	viper.SetDefault("verbose", false)

	// Set markdown template defaults
	defaultTemplate := markdown.DefaultTemplateConfig()
	viper.SetDefault("markdown_template.attribute_template", defaultTemplate.AttributeTemplate)
	viper.SetDefault("markdown_template.required_text", defaultTemplate.RequiredText)
	viper.SetDefault("markdown_template.optional_text", defaultTemplate.OptionalText)
	viper.SetDefault("markdown_template.escape_mode", defaultTemplate.EscapeMode)
	viper.SetDefault("markdown_template.indent_style", defaultTemplate.IndentStyle)
	viper.SetDefault("markdown_template.indent_size", defaultTemplate.IndentSize)

	// Set split command defaults
	viper.SetDefault("split.input_path", "README.md")
	viper.SetDefault("split.output_dir", "variables")
	viper.SetDefault("split.header_file", "")
	viper.SetDefault("split.footer_file", "")
}
