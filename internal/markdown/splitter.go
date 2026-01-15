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

	// Regular expression to match MARINATED markers
	marinatedMarkerRe := regexp.MustCompile(`<!-- MARINATED:\s*(\S+?)\s*-->`)

	lines := strings.Split(content, "\n")
	var currentSection *VariableSection
	var sectionLines []string
	inSection := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this is a new variable heading (###)
		if strings.HasPrefix(trimmedLine, "### ") {
			// Save previous section if exists
			if currentSection != nil && len(sectionLines) > 0 {
				currentSection.Content = strings.Join(sectionLines, "\n")
				sections = append(sections, *currentSection)
			}

			// Start a new section
			currentSection = nil
			sectionLines = []string{line}
			inSection = true

			// Check if this section has a MARINATED marker in the next few lines
			// (typically in the Description line)
			for j := i + 1; j < len(lines) && j < i+10; j++ {
				if matches := marinatedMarkerRe.FindStringSubmatch(lines[j]); matches != nil {
					// Found a MARINATED marker
					variableName := strings.ReplaceAll(matches[1], "\\_", "_")
					currentSection = &VariableSection{
						VariableName: variableName,
					}
					break
				}
				// Stop looking if we hit another heading
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "###") ||
					strings.HasPrefix(strings.TrimSpace(lines[j]), "##") {
					break
				}
			}
			continue
		}

		// Check if we hit a new major section (##) or higher
		if strings.HasPrefix(trimmedLine, "## ") || strings.HasPrefix(trimmedLine, "# ") {
			// Save current section if it's a MARINATED variable
			if currentSection != nil && len(sectionLines) > 0 {
				currentSection.Content = strings.Join(sectionLines, "\n")
				sections = append(sections, *currentSection)
			}
			currentSection = nil
			sectionLines = nil
			inSection = false
			continue
		}

		// Accumulate lines if we're in a MARINATED section
		if inSection && currentSection != nil {
			sectionLines = append(sectionLines, line)
		}
	}

	// Don't forget the last section
	if currentSection != nil && len(sectionLines) > 0 {
		currentSection.Content = strings.Join(sectionLines, "\n")
		sections = append(sections, *currentSection)
	}

	return sections, nil
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content.String()), 0644); err != nil {
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

		if err := s.WriteSection(outputPath, section); err != nil {
			return createdFiles, fmt.Errorf("failed to write section for %s: %w", section.VariableName, err)
		}

		createdFiles = append(createdFiles, outputPath)
	}

	return createdFiles, nil
}
