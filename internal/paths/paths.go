package paths

import (
	"fmt"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/config"
)

// SetupEnvironment is a shared function that resolves the module path and loads configuration.
// It handles both absolute and relative paths correctly.
// Returns: (moduleRoot, config, error).
func SetupEnvironment(args []string) (string, *config.Config, error) {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	// filepath.Abs handles both absolute and relative paths correctly:
	// - Absolute paths (e.g., /srv/bla) are returned as-is
	// - Relative paths (e.g., ./subdir or subdir) are joined with current directory
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve module path: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return "", nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return absRoot, cfg, nil
}

// ResolveExportPath returns the absolute path to the export directory.
// Uses cfg.ExportPath relative to moduleRoot.
func ResolveExportPath(moduleRoot string, cfg *config.Config) string {
	if filepath.IsAbs(cfg.ExportPath) {
		return cfg.ExportPath
	}
	return filepath.Join(moduleRoot, cfg.ExportPath)
}
