package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/spf13/viper"
)

func TestLoad_Defaults(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	config.SetDefaults()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check main defaults
	if cfg.DocsPath != "docs" {
		t.Errorf("DocsPath = %s, want docs", cfg.DocsPath)
	}
	if cfg.VariablesPath != "." {
		t.Errorf("VariablesPath = %s, want .", cfg.VariablesPath)
	}
	if cfg.ReadmePath != "README.md" {
		t.Errorf("ReadmePath = %s, want README.md", cfg.ReadmePath)
	}

	// Check split defaults
	if cfg.Split == nil {
		t.Fatal("Split config is nil")
	}
	if cfg.Split.InputPath != "README.md" {
		t.Errorf("Split.InputPath = %s, want README.md", cfg.Split.InputPath)
	}
	if cfg.Split.OutputDir != "variables" {
		t.Errorf("Split.OutputDir = %s, want variables", cfg.Split.OutputDir)
	}
	if cfg.Split.HeaderFile != "" {
		t.Errorf("Split.HeaderFile = %s, want empty string", cfg.Split.HeaderFile)
	}
	if cfg.Split.FooterFile != "" {
		t.Errorf("Split.FooterFile = %s, want empty string", cfg.Split.FooterFile)
	}
}

func TestLoad_FromConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".marinated.yml")

	configContent := `docs_path: custom_docs
variables_path: terraform
readme_path: VARIABLES.md
split:
  input_path: docs/all_vars.md
  output_dir: split_vars
  header_file: templates/_header.md
  footer_file: templates/_footer.md
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Reset viper and set config file
	viper.Reset()
	config.SetDefaults()
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check that values were loaded from config file
	if cfg.DocsPath != "custom_docs" {
		t.Errorf("DocsPath = %s, want custom_docs", cfg.DocsPath)
	}
	if cfg.VariablesPath != "terraform" {
		t.Errorf("VariablesPath = %s, want terraform", cfg.VariablesPath)
	}
	if cfg.ReadmePath != "VARIABLES.md" {
		t.Errorf("ReadmePath = %s, want VARIABLES.md", cfg.ReadmePath)
	}

	// Check split config
	if cfg.Split == nil {
		t.Fatal("Split config is nil")
	}
	if cfg.Split.InputPath != "docs/all_vars.md" {
		t.Errorf("Split.InputPath = %s, want docs/all_vars.md", cfg.Split.InputPath)
	}
	if cfg.Split.OutputDir != "split_vars" {
		t.Errorf("Split.OutputDir = %s, want split_vars", cfg.Split.OutputDir)
	}
	if cfg.Split.HeaderFile != "templates/_header.md" {
		t.Errorf("Split.HeaderFile = %s, want templates/_header.md", cfg.Split.HeaderFile)
	}
	if cfg.Split.FooterFile != "templates/_footer.md" {
		t.Errorf("Split.FooterFile = %s, want templates/_footer.md", cfg.Split.FooterFile)
	}
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	config.SetDefaults()

	// Check that all defaults are set in viper
	if viper.GetString("docs_path") != "docs" {
		t.Errorf("Default docs_path not set correctly")
	}
	if viper.GetString("split.input_path") != "README.md" {
		t.Errorf("Default split.input_path not set correctly")
	}
	if viper.GetString("split.output_dir") != "variables" {
		t.Errorf("Default split.output_dir not set correctly")
	}
}
