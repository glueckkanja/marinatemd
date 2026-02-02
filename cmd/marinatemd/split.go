package marinatemd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/glueckkanja/marinatemd/internal/config"
	"github.com/glueckkanja/marinatemd/internal/logger"
	"github.com/glueckkanja/marinatemd/internal/markdown"
	"github.com/glueckkanja/marinatemd/internal/paths"
	"github.com/glueckkanja/marinatemd/internal/yamlio"
	"github.com/spf13/cobra"
)

var (
	splitInputFile  string
	splitOutputDir  string
	splitHeaderFile string
	splitFooterFile string
)

// splitCmd represents the split command that post-processes markdown files.
var splitCmd = &cobra.Command{
	Use:   "split [module-path]",
	Short: "Split a markdown file into separate files for each MARINATED variable",
	Long: `Post-process a markdown file by splitting it into separate files for each MARINATED variable.

This command:
  1. Scans the input markdown file for MARINATED variable sections
  2. Extracts each section including heading, description, type, and default
  3. Creates a separate .md file for each variable in the output directory
  4. Optionally prepends a header and/or appends a footer to each file

This is useful when you want individual documentation files for each variable
instead of a single monolithic README.

Example:
  marinatemd split .
  marinatemd split --input docs/README.md --output docs/variables .
  marinatemd split --header _header.md --footer _footer.md .`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSplit,
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().StringVar(
		&splitInputFile,
		"input",
		"",
		"input markdown file to split (defaults to docs_file from configuration)",
	)

	splitCmd.Flags().StringVar(
		&splitOutputDir,
		"output",
		"",
		"output directory for split files (defaults to docs/variables)",
	)

	splitCmd.Flags().StringVar(
		&splitHeaderFile,
		"header",
		"",
		"path to header file to prepend to each split file",
	)

	splitCmd.Flags().StringVar(
		&splitFooterFile,
		"footer",
		"",
		"path to footer file to append to each split file",
	)
}

func runSplit(_ *cobra.Command, args []string) error {
	moduleRoot, cfg, err := paths.SetupEnvironment(args)
	if err != nil {
		return err
	}

	inputPath := resolveInputPath(moduleRoot, cfg)
	outputDir := resolveOutputDir(moduleRoot, cfg)
	headerPath, footerPath := resolveTemplatePaths(moduleRoot, cfg)

	splitter, err := createSplitter(headerPath, footerPath)
	if err != nil {
		return err
	}

	if appErr := applyConfigNameOverrides(splitter, moduleRoot, cfg); appErr != nil {
		return appErr
	}

	return executeSplit(splitter, inputPath, outputDir, moduleRoot)
}

func resolveInputPath(moduleRoot string, cfg *config.Config) string {
	switch {
	case splitInputFile != "" && filepath.IsAbs(splitInputFile):
		return splitInputFile
	case splitInputFile != "":
		return filepath.Join(moduleRoot, splitInputFile)
	case cfg.Split != nil && cfg.Split.InputPath != "":
		exportPath := paths.ResolveExportPath(moduleRoot, cfg)
		return filepath.Join(exportPath, cfg.Split.InputPath)
	default:
		return resolveDefaultDocsFile(moduleRoot, cfg)
	}
}

func resolveDefaultDocsFile(moduleRoot string, cfg *config.Config) string {
	docsFile := "README.md"
	if cfg != nil && cfg.DocsFile != "" {
		docsFile = cfg.DocsFile
	}

	if filepath.IsAbs(docsFile) {
		return docsFile
	}

	return filepath.Join(moduleRoot, docsFile)
}

func resolveOutputDir(moduleRoot string, cfg *config.Config) string {
	switch {
	case splitOutputDir != "" && filepath.IsAbs(splitOutputDir):
		return splitOutputDir
	case splitOutputDir != "":
		return filepath.Join(moduleRoot, splitOutputDir)
	case cfg.Split != nil && cfg.Split.OutputDir != "":
		exportPath := paths.ResolveExportPath(moduleRoot, cfg)
		return filepath.Join(exportPath, cfg.Split.OutputDir)
	default:
		exportPath := paths.ResolveExportPath(moduleRoot, cfg)
		return filepath.Join(exportPath, "variables")
	}
}

func resolveTemplatePaths(moduleRoot string, cfg *config.Config) (string, string) {
	headerPath := resolveTemplatePath(moduleRoot, splitHeaderFile, cfg.Split, func(s *config.SplitConfig) string {
		return s.HeaderFile
	})
	footerPath := resolveTemplatePath(moduleRoot, splitFooterFile, cfg.Split, func(s *config.SplitConfig) string {
		return s.FooterFile
	})
	return headerPath, footerPath
}

func resolveTemplatePath(
	absRoot, cliFlag string,
	splitCfg *config.SplitConfig,
	getter func(*config.SplitConfig) string,
) string {
	switch {
	case cliFlag != "" && filepath.IsAbs(cliFlag):
		return cliFlag
	case cliFlag != "":
		return filepath.Join(absRoot, cliFlag)
	case splitCfg != nil && getter(splitCfg) != "":
		return filepath.Join(absRoot, getter(splitCfg))
	default:
		return ""
	}
}

func createSplitter(headerPath, footerPath string) (*markdown.Splitter, error) {
	if headerPath == "" && footerPath == "" {
		return markdown.NewSplitter(), nil
	}

	splitter := markdown.NewSplitter()

	if headerPath != "" {
		headerContent, err := os.ReadFile(headerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read header file: %w", err)
		}
		splitter.SetHeader(string(headerContent))
	}

	if footerPath != "" {
		footerContent, err := os.ReadFile(footerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read footer file: %w", err)
		}
		splitter.SetFooter(string(footerContent))
	}

	logger.Log.Debug("using templates", "header", headerPath, "footer", footerPath)
	return splitter, nil
}

func executeSplit(splitter *markdown.Splitter, inputPath, outputDir, absRoot string) error {
	logger.Log.Debug("splitting file", "input", inputPath, "output", outputDir)
	createdFiles, err := splitter.SplitToFiles(inputPath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to split file: %w", err)
	}

	printSplitSummary(createdFiles, absRoot)
	return nil
}

func printSplitSummary(createdFiles []string, absRoot string) {
	logger.Log.Info("split complete", "files", len(createdFiles))
	for _, filePath := range createdFiles {
		relPath, relErr := filepath.Rel(absRoot, filePath)
		if relErr != nil {
			relPath = filePath
		}
		logger.Log.Debug("created file", "path", relPath)
	}
}

func applyConfigNameOverrides(splitter *markdown.Splitter, moduleRoot string, cfg *config.Config) error {
	exportPath := paths.ResolveExportPath(moduleRoot, cfg)
	reader := yamlio.NewReader(exportPath)

	files, err := os.ReadDir(filepath.Join(exportPath, "variables"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to list schema files: %w", err)
	}

	for _, entry := range files {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		variable := strings.TrimSuffix(entry.Name(), ".yaml")
		schemaFile, readErr := reader.ReadSchema(variable)
		if readErr != nil {
			return fmt.Errorf("failed to read schema for %s: %w", variable, readErr)
		}
		if schemaFile == nil || schemaFile.Config == nil || schemaFile.Config.Name == "" {
			continue
		}

		splitter.SetNameOverride(variable, schemaFile.Config.Name)
	}

	return nil
}
