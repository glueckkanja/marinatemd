package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// VariableSection represents an extracted section for a MARINATED variable.
type VariableSection struct {
	VariableName string
	Content      string // The full content for this variable including heading, description, type, default
}

// Splitter handles splitting a markdown file by MARINATED variables.
type Splitter struct {
	headerContent string
	footerContent string
}

// NewSplitter creates a new markdown splitter.
func NewSplitter() *Splitter {
	return &Splitter{}
}

// NewSplitterWithTemplate creates a new markdown splitter with header and footer templates.
func NewSplitterWithTemplate(headerPath, footerPath string) (*Splitter, error) {
	s := &Splitter{}

	if headerPath != "" {
		content, err := os.ReadFile(headerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read header file: %w", err)
		}
		s.headerContent = string(content)
	}

	if footerPath != "" {
		content, err := os.ReadFile(footerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read footer file: %w", err)
		}
		s.footerContent = string(content)
	}

	return s, nil
}

// SetHeader sets the header content to prepend to each split file.
func (s *Splitter) SetHeader(header string) {
	s.headerContent = header
}

// SetFooter sets the footer content to append to each split file.
func (s *Splitter) SetFooter(footer string) {
	s.footerContent = footer
}

// ExtractSections parses a markdown file and extracts all MARINATED variable sections.
// Each section includes the variable heading, description (with MARINATED markers),
// type, default, and any other related content.
func (s *Splitter) ExtractSections(filePath string) ([]VariableSection, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return s.extractSectionsFromContent(string(content))
}

// extractSectionsFromContent parses markdown content and extracts variable sections.
func (s *Splitter) extractSectionsFromContent(content string) ([]VariableSection, error) {
	var sections []VariableSection
	marinatedMarkerRe := regexp.MustCompile(`<!-- MARINATED:\s*(\S+?)\s*-->`)

	lines := strings.Split(content, "\n")
	var currentSection *VariableSection
	var sectionLines []string
	inSection := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if isVariableHeading(trimmedLine) {
			sections = saveCurrentSection(currentSection, sectionLines, sections)
			currentSection, sectionLines, inSection = startNewSection(line, lines, i, marinatedMarkerRe)
			continue
		}

		if isMajorHeading(trimmedLine) {
			sections = saveCurrentSection(currentSection, sectionLines, sections)
			currentSection = nil
			sectionLines = nil
			inSection = false
			continue
		}

		if inSection && currentSection != nil {
			sectionLines = append(sectionLines, line)
		}
	}

	// Save the last section
	sections = saveCurrentSection(currentSection, sectionLines, sections)
	return sections, nil
}

func isVariableHeading(line string) bool {
	return strings.HasPrefix(line, "### ")
}

func isMajorHeading(line string) bool {
	return strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ")
}

func saveCurrentSection(
	currentSection *VariableSection,
	sectionLines []string,
	sections []VariableSection,
) []VariableSection {
	if currentSection != nil && len(sectionLines) > 0 {
		currentSection.Content = strings.Join(sectionLines, "\n")
		sections = append(sections, *currentSection)
	}
	return sections
}

func startNewSection(
	line string,
	lines []string,
	currentIndex int,
	marinatedMarkerRe *regexp.Regexp,
) (*VariableSection, []string, bool) {
	sectionLines := []string{line}
	currentSection := findMarinatedMarker(lines, currentIndex, marinatedMarkerRe)
	return currentSection, sectionLines, true
}

func findMarinatedMarker(lines []string, startIndex int, marinatedMarkerRe *regexp.Regexp) *VariableSection {
	maxLookAhead := 10
	endIndex := min(startIndex+1+maxLookAhead, len(lines))

	for j := startIndex + 1; j < endIndex; j++ {
		if matches := marinatedMarkerRe.FindStringSubmatch(lines[j]); matches != nil {
			variableName := strings.ReplaceAll(matches[1], "\\_", "_")
			return &VariableSection{VariableName: variableName}
		}
		if isVariableHeading(strings.TrimSpace(lines[j])) || isMajorHeading(strings.TrimSpace(lines[j])) {
			break
		}
	}
	return nil
}

// WriteSection writes a single variable section to a file with optional header and footer.
func (s *Splitter) WriteSection(outputPath string, section VariableSection) error {
	var content strings.Builder

	// Add header if configured
	if s.headerContent != "" {
		content.WriteString(s.headerContent)
		if !strings.HasSuffix(s.headerContent, "\n") {
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Add the section content
	content.WriteString(strings.TrimSpace(section.Content))
	content.WriteString("\n")

	// Add footer if configured
	if s.footerContent != "" {
		content.WriteString("\n")
		content.WriteString(s.footerContent)
		if !strings.HasSuffix(s.footerContent, "\n") {
			content.WriteString("\n")
		}
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SplitToFiles splits a markdown file into separate files for each MARINATED variable.
// Each output file is named <variable_name>.md and placed in the outputDir.
func (s *Splitter) SplitToFiles(inputPath string, outputDir string) ([]string, error) {
	sections, err := s.ExtractSections(inputPath)
	if err != nil {
		return nil, err
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("no MARINATED variables found in %s", inputPath)
	}

	var createdFiles []string

	for _, section := range sections {
		// Create output filename: <variable_name>.md
		outputFilename := fmt.Sprintf("%s.md", section.VariableName)
		outputPath := filepath.Join(outputDir, outputFilename)

		if writeErr := s.WriteSection(outputPath, section); writeErr != nil {
			return createdFiles, fmt.Errorf("failed to write section for %s: %w", section.VariableName, writeErr)
		}

		createdFiles = append(createdFiles, outputPath)
	}

	return createdFiles, nil
}
