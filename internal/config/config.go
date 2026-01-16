package config

import (
	"github.com/c4a8-azure/marinatemd/internal/logger"
	"github.com/c4a8-azure/marinatemd/internal/markdown"
	"github.com/spf13/viper"
)

// Config represents the application configuration loaded from .marinated.yml.
type Config struct {
	// Mock configuration fields for initialization
	// These will be expanded as features are implemented

	// ExportPath is the path where YAML schemas and documentation are exported
	ExportPath string `mapstructure:"export_path"`

	// DocsFile is the path to the main documentation file to inject into
	DocsFile string `mapstructure:"docs_file"`

	// Verbose enables verbose logging
	Verbose bool `mapstructure:"verbose"`

	// MarkdownTemplate configures how markdown is generated from schema
	MarkdownTemplate *markdown.TemplateConfig `mapstructure:"markdown_template"`

	// Split configures the split command behavior
	Split *SplitConfig `mapstructure:"split"`
}

// SplitConfig represents configuration for the split command.
type SplitConfig struct {
	// InputPath is the input markdown file to split (relative to export_path).
	// If empty, defaults to docs_file.
	InputPath string `mapstructure:"input_path"`

	// OutputDir is the output directory for split files (relative to export_path)
	OutputDir string `mapstructure:"output_dir"`

	// HeaderFile is the path to the header file to prepend to each split file
	HeaderFile string `mapstructure:"header_file"`

	// FooterFile is the path to the footer file to append to each split file
	FooterFile string `mapstructure:"footer_file"`
}

// Load returns the configuration loaded from viper.
// This should be called after Viper has been initialized by Cobra.
func Load() (*Config, error) {
	logger.Log.Debug("loading configuration")

	cfg := &Config{
		// Set defaults
		ExportPath:       "docs",
		DocsFile:         "README.md",
		Verbose:          false,
		MarkdownTemplate: markdown.DefaultTemplateConfig(),
		Split: &SplitConfig{
			InputPath:  "", // Empty means use DocsFile
			OutputDir:  "variables",
			HeaderFile: "",
			FooterFile: "",
		},
	}

	logger.Log.Debug("config defaults set",
		"export_path", cfg.ExportPath,
		"docs_file", cfg.DocsFile,
		"split.output_dir", cfg.Split.OutputDir)

	// Unmarshal viper config into struct
	if err := viper.Unmarshal(cfg); err != nil {
		logger.Log.Debug("failed to unmarshal config", "error", err)
		return nil, err
	}

	logger.Log.Debug("config loaded from viper",
		"export_path", cfg.ExportPath,
		"docs_file", cfg.DocsFile,
		"split.input_path", cfg.Split.InputPath,
		"split.output_dir", cfg.Split.OutputDir,
		"split.header_file", cfg.Split.HeaderFile,
		"split.footer_file", cfg.Split.FooterFile)

	// Validate markdown template configuration
	if err := cfg.MarkdownTemplate.Validate(); err != nil {
		logger.Log.Debug("config validation failed", "error", err)
		return nil, err
	}

	logger.Log.Debug("configuration validated successfully")
	return cfg, nil
}

// SetDefaults configures default values for viper
// SetDefaults sets default configuration values.
// This should be called during initialization before config file is read.
func SetDefaults() {
	logger.Log.Debug("setting viper default values")

	viper.SetDefault("export_path", "docs")
	viper.SetDefault("docs_file", "README.md")
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
	viper.SetDefault("split.input_path", "") // Empty means use docs_file
	viper.SetDefault("split.output_dir", "variables")
	viper.SetDefault("split.header_file", "")
	viper.SetDefault("split.footer_file", "")

	logger.Log.Debug("viper defaults configured")
}
