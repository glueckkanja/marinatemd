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
	// Determine the module root
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("failed to resolve module path: %w", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine input file path
	inputPath := splitInputFile
	if inputPath == "" {
		// Default to docs/README.md
		docsPath := filepath.Join(absRoot, cfg.DocsPath)
		inputPath = filepath.Join(docsPath, "README.md")
	} else if !filepath.IsAbs(inputPath) {
		// Make relative paths relative to the module root
		inputPath = filepath.Join(absRoot, inputPath)
	}

	// Determine output directory
	outputDir := splitOutputDir
	if outputDir == "" {
		// Default to docs/variables
		docsPath := filepath.Join(absRoot, cfg.DocsPath)
		outputDir = filepath.Join(docsPath, "variables")
	} else if !filepath.IsAbs(outputDir) {
		// Make relative paths relative to the module root
		outputDir = filepath.Join(absRoot, outputDir)
	}

	// Resolve header and footer paths if provided
	var headerPath, footerPath string
	if splitHeaderFile != "" {
		if !filepath.IsAbs(splitHeaderFile) {
			headerPath = filepath.Join(absRoot, splitHeaderFile)
		} else {
			headerPath = splitHeaderFile
		}
	}
	if splitFooterFile != "" {
		if !filepath.IsAbs(splitFooterFile) {
			footerPath = filepath.Join(absRoot, splitFooterFile)
		} else {
			footerPath = splitFooterFile
		}
	}

	// Create splitter with templates if provided
	var splitter *markdown.Splitter
	if headerPath != "" || footerPath != "" {
		splitter, err = markdown.NewSplitterWithTemplate(headerPath, footerPath)
		if err != nil {
			return fmt.Errorf("failed to create splitter with templates: %w", err)
		}
		fmt.Printf("Using header: %s\n", headerPath)
		fmt.Printf("Using footer: %s\n", footerPath)
	} else {
		splitter = markdown.NewSplitter()
	}

	// Split the file
	fmt.Printf("Splitting %s...\n", inputPath)
	createdFiles, err := splitter.SplitToFiles(inputPath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to split file: %w", err)
	}

	// Print summary
	fmt.Printf("\n✓ Successfully split into %d files:\n", len(createdFiles))
	for _, filePath := range createdFiles {
		relPath, err := filepath.Rel(absRoot, filePath)
		if err != nil {
			relPath = filePath
		}
		fmt.Printf("  - %s\n", relPath)
	}

	return nil
}
