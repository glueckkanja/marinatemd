package marinatemd

import (
	"fmt"
	"path/filepath"

	"github.com/c4a8-azure/marinatemd/internal/config"
	"github.com/c4a8-azure/marinatemd/internal/markdown"
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
		"input markdown file to split (defaults to docs/README.md)",
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
	absRoot, cfg, err := setupSplitEnvironment(args)
	if err != nil {
		return err
	}

	inputPath := resolveInputPath(absRoot, cfg)
	outputDir := resolveOutputDir(absRoot, cfg)
	headerPath, footerPath := resolveTemplatePaths(absRoot, cfg)

	splitter, err := createSplitter(headerPath, footerPath)
	if err != nil {
		return err
	}

	return executeSplit(splitter, inputPath, outputDir, absRoot)
}

func setupSplitEnvironment(args []string) (string, *config.Config, error) {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

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

func resolveInputPath(absRoot string, cfg *config.Config) string {
	switch {
	case splitInputFile != "" && filepath.IsAbs(splitInputFile):
		return splitInputFile
	case splitInputFile != "":
		return filepath.Join(absRoot, splitInputFile)
	case cfg.Split != nil && cfg.Split.InputPath != "":
		docsPath := filepath.Join(absRoot, cfg.DocsPath)
		return filepath.Join(docsPath, cfg.Split.InputPath)
	default:
		docsPath := filepath.Join(absRoot, cfg.DocsPath)
		return filepath.Join(docsPath, "README.md")
	}
}

func resolveOutputDir(absRoot string, cfg *config.Config) string {
	switch {
	case splitOutputDir != "" && filepath.IsAbs(splitOutputDir):
		return splitOutputDir
	case splitOutputDir != "":
		return filepath.Join(absRoot, splitOutputDir)
	case cfg.Split != nil && cfg.Split.OutputDir != "":
		docsPath := filepath.Join(absRoot, cfg.DocsPath)
		return filepath.Join(docsPath, cfg.Split.OutputDir)
	default:
		docsPath := filepath.Join(absRoot, cfg.DocsPath)
		return filepath.Join(docsPath, "variables")
	}
}

func resolveTemplatePaths(absRoot string, cfg *config.Config) (string, string) {
	headerPath := resolveTemplatePath(absRoot, splitHeaderFile, cfg.Split, func(s *config.SplitConfig) string {
		return s.HeaderFile
	})
	footerPath := resolveTemplatePath(absRoot, splitFooterFile, cfg.Split, func(s *config.SplitConfig) string {
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
	if headerPath != "" || footerPath != "" {
		splitter, err := markdown.NewSplitterWithTemplate(headerPath, footerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create splitter with templates: %w", err)
		}
		fmt.Printf("Using header: %s\n", headerPath)
		fmt.Printf("Using footer: %s\n", footerPath)
		return splitter, nil
	}
	return markdown.NewSplitter(), nil
}

func executeSplit(splitter *markdown.Splitter, inputPath, outputDir, absRoot string) error {
	fmt.Printf("Splitting %s...\n", inputPath)
	createdFiles, err := splitter.SplitToFiles(inputPath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to split file: %w", err)
	}

	printSplitSummary(createdFiles, absRoot)
	return nil
}

func printSplitSummary(createdFiles []string, absRoot string) {
	fmt.Printf("\n✓ Successfully split into %d files:\n", len(createdFiles))
	for _, filePath := range createdFiles {
		relPath, relErr := filepath.Rel(absRoot, filePath)
		if relErr != nil {
			relPath = filePath
		}
		fmt.Printf("  - %s\\n", relPath)
	}
}
