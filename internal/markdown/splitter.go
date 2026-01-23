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
// Only MARINATED variables are extracted, but the entire variable section is captured
// including Type:, Default:, and any subsections like "### Advanced Settings".
func (s *Splitter) extractSectionsFromContent(content string) ([]VariableSection, error) {
	var sections []VariableSection
	marinatedMarkerRe := regexp.MustCompile(`<!-- MARINATED:\s*(\S+?)\s*-->`)

	lines := strings.Split(content, "\n")
	var currentSection *VariableSection
	var sectionLines []string
	inSection := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this is a potential new variable definition (### heading)
		if isVariableHeading(trimmedLine) {
			// Look ahead to see if this is actually a NEW variable (has Type:/Default:)
			// or just a subsection within the current variable (like "### Advanced Settings")
			if isNewVariableStart(lines, i) {
				// Save the previous section if it exists
				sections = saveCurrentSection(currentSection, sectionLines, sections)
				currentSection, sectionLines, inSection = startNewSection(line, lines, i, marinatedMarkerRe)
				continue
			}
			// Otherwise, it's a subsection - keep collecting if we're in a section
		}

		// Major headings always end the current section
		if isMajorHeading(trimmedLine) {
			sections = saveCurrentSection(currentSection, sectionLines, sections)
			currentSection = nil
			sectionLines = nil
			inSection = false
			continue
		}

		// Collect all lines when in a section
		if inSection && currentSection != nil {
			sectionLines = append(sectionLines, line)
		}
	}

	// Save the last section
	sections = saveCurrentSection(currentSection, sectionLines, sections)
	return sections, nil
}

// isNewVariableStart checks if a ### heading is the start of a new variable definition.
// A true variable heading is followed by "Description:" within a few lines.
// Subsection headings (like "### Advanced Settings") are NOT followed by "Description:".
func isNewVariableStart(lines []string, headingIndex int) bool {
	// Look for "Description:" which terraform-docs puts right after variable headings
	maxLookAhead := 15 // Variables have Description: very close to the heading
	endIndex := min(headingIndex+maxLookAhead, len(lines))

	for j := headingIndex + 1; j < endIndex; j++ {
		trimmed := strings.TrimSpace(lines[j])

		// Real variables have "Description:" within a few lines
		if strings.HasPrefix(trimmed, "Description:") {
			return true
		}

		// If we hit another heading before finding Description:, it's not a variable
		if isVariableHeading(trimmed) || isMajorHeading(trimmed) {
			return false
		}
	}

	// If we didn't find Description: within reasonable distance, it's a subsection
	return false
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
	maxLookAhead := 50 // Increased to handle sections with subsections like "### Attributes"
	endIndex := min(startIndex+1+maxLookAhead, len(lines))

	for j := startIndex + 1; j < endIndex; j++ {
		trimmed := strings.TrimSpace(lines[j])

		if matches := marinatedMarkerRe.FindStringSubmatch(lines[j]); matches != nil {
			variableName := strings.ReplaceAll(matches[1], "\\_", "_")
			return &VariableSection{VariableName: variableName}
		}

		// Stop searching if we hit another variable heading (one with Description:)
		// or a major heading - this means we've moved past the current variable
		if isVariableHeading(trimmed) && isNewVariableStart(lines, j) {
			break
		}
		if isMajorHeading(trimmed) {
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
