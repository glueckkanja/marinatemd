package hclparse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TerraformInjector handles injecting markdown documentation into Terraform variable files.
type TerraformInjector struct {
	modulePath string
}

// NewTerraformInjector creates a new Terraform injector for the given module path.
func NewTerraformInjector(modulePath string) *TerraformInjector {
	return &TerraformInjector{
		modulePath: modulePath,
	}
}

// FindVariableFile locates the Terraform file containing a variable with the given marinated ID.
// Returns the file path and the variable name, or an error if not found.
func (ti *TerraformInjector) FindVariableFile(marinatedID string) (string, string, error) {
	parser := NewParser()
	if err := parser.ParseVariables(ti.modulePath); err != nil {
		return "", "", fmt.Errorf("failed to parse variables: %w", err)
	}

	marinatedVars, err := parser.ExtractMarinatedVars()
	if err != nil {
		return "", "", fmt.Errorf("failed to extract marinated variables: %w", err)
	}

	for _, v := range marinatedVars {
		if v.MarinatedID == marinatedID {
			// Find the file containing this variable
			pattern := filepath.Join(ti.modulePath, "variables*.tf")
			matches, globErr := filepath.Glob(pattern)
			if globErr != nil {
				return "", "", fmt.Errorf("failed to glob for variables files: %w", globErr)
			}

			for _, filename := range matches {
				if containsVariableDefinition(filename, v.Name) {
					return filename, v.Name, nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("variable with marinated ID %s not found", marinatedID)
}

// containsVariableDefinition checks if a file contains a variable definition.
func containsVariableDefinition(filename, variableName string) bool {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	// Look for variable "name" { pattern
	pattern := fmt.Sprintf("variable \"%s\"", variableName)
	return strings.Contains(string(content), pattern)
}

// InjectIntoFile injects markdown documentation inside the description string of a Terraform variable.
// It looks for the MARINATED marker in the description and injects content inside the description using HTML comments.
func (ti *TerraformInjector) InjectIntoFile(filePath, marinatedID, markdownContent string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)
	// Handle both escaped and unescaped underscores in markers
	escapedID := strings.ReplaceAll(marinatedID, "_", `\_`)
	startComment := fmt.Sprintf("<!-- MARINATED: %s -->", marinatedID)
	escapedStartComment := fmt.Sprintf("<!-- MARINATED: %s -->", escapedID)
	endComment := fmt.Sprintf("<!-- /MARINATED: %s -->", marinatedID)
	escapedEndComment := fmt.Sprintf("<!-- /MARINATED: %s -->", escapedID)

	// Check for either version of the marker
	if !strings.Contains(fileContent, startComment) && !strings.Contains(fileContent, escapedStartComment) {
		return fmt.Errorf("MARINATED marker %s not found in file", startComment)
	}

	// Use the version that exists in the file
	actualStartComment := startComment
	actualEndComment := endComment
	if strings.Contains(fileContent, escapedStartComment) {
		actualStartComment = escapedStartComment
		actualEndComment = escapedEndComment
	}

	modified, err := processFileContent(fileContent, marinatedID, markdownContent, actualStartComment, actualEndComment)
	if err != nil {
		return err
	}

	if writeErr := os.WriteFile(filePath, []byte(modified), 0600); writeErr != nil {
		return fmt.Errorf("failed to write file: %w", writeErr)
	}
	return nil
}

func processFileContent(fileContent, marinatedID, markdownContent, startComment, endComment string) (string, error) {
	lines := strings.Split(fileContent, "\n")
	var result strings.Builder
	foundMarker := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if isDescriptionLine(line) {
			hasMarker := checkForMarker(lines, i, startComment)
			if !hasMarker {
				result.WriteString(line)
				result.WriteString("\n")
				continue
			}

			foundMarker = true
			i = processDescription(lines, i, &result, line, marinatedID, markdownContent, startComment, endComment)
			continue
		}

		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	if !foundMarker {
		return "", errors.New("could not find description with MARINATED marker")
	}
	return result.String(), nil
}

func isDescriptionLine(line string) bool {
	return strings.Contains(line, "description") && strings.Contains(line, "=")
}

func checkForMarker(lines []string, startIdx int, startComment string) bool {
	if strings.Contains(lines[startIdx], startComment) {
		return true
	}

	// For heredoc, check until we find the closing delimiter or another variable
	// Extract the heredoc delimiter if present
	delimiter := extractHeredocDelimiter(lines[startIdx])
	if delimiter == "" {
		return false
	}

	// Search through the heredoc content
	for j := startIdx + 1; j < len(lines); j++ {
		if strings.Contains(lines[j], startComment) {
			return true
		}
		// Stop if we find the closing delimiter
		if isClosingDelimiterSimple(lines[j], delimiter) {
			break
		}
		// Stop if we find another variable
		if strings.Contains(lines[j], "variable") {
			break
		}
	}
	return false
}

// isClosingDelimiterSimple checks if a line contains just the delimiter.
func isClosingDelimiterSimple(line, delimiter string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == delimiter
}

func processDescription(
	lines []string,
	idx int,
	result *strings.Builder,
	line, marinatedID, markdownContent, startComment, endComment string,
) int {
	if isSingleLineDescription(line) {
		result.WriteString(convertToMultilineDescription(line, marinatedID, markdownContent))
		result.WriteString("\n")
		return idx
	}

	// Check for any heredoc format (<<DELIMITER or <<-DELIMITER)
	if delimiter := extractHeredocDelimiter(line); delimiter != "" {
		return processHeredoc(lines, idx, result, line, markdownContent, startComment, endComment, delimiter)
	}
	return idx
}

// extractHeredocDelimiter extracts the delimiter from a heredoc line.
// Returns the delimiter name or empty string if not a heredoc.
func extractHeredocDelimiter(line string) string {
	// Match <<DELIMITER or <<-DELIMITER
	re := regexp.MustCompile(`<<-?([A-Z_]+)`)
	matches := re.FindStringSubmatch(line)
	const minMatchGroups = 2
	if len(matches) >= minMatchGroups {
		return matches[1]
	}
	return ""
}

func processHeredoc(
	lines []string,
	idx int,
	result *strings.Builder,
	line, markdownContent, startComment, endComment, delimiter string,
) int {
	indent := getIndentation(line)
	result.WriteString(line)
	result.WriteString("\n")

	idx++
	foundStartMarker := false
	foundEndMarker := false
	var contentLines []string

	// Collect all heredoc content
	for idx < len(lines) {
		currentLine := lines[idx]

		// Check if we reached the closing delimiter
		if isClosingDelimiter(currentLine, indent, delimiter) {
			// Process collected content
			processHeredocContent(
				result,
				contentLines,
				foundStartMarker,
				foundEndMarker,
				markdownContent,
				startComment,
				endComment,
			)

			result.WriteString(currentLine)
			result.WriteString("\n")
			break
		}

		// Track if we've seen the markers
		if strings.Contains(currentLine, startComment) {
			foundStartMarker = true
		}
		if strings.Contains(currentLine, endComment) {
			foundEndMarker = true
		}

		contentLines = append(contentLines, currentLine)
		idx++
	}
	return idx
}

func processHeredocContent(
	result *strings.Builder,
	contentLines []string,
	foundStartMarker, foundEndMarker bool,
	markdownContent, startComment, endComment string,
) {
	if !foundStartMarker {
		// No marker found - shouldn't happen if checkForMarker worked correctly
		for _, line := range contentLines {
			result.WriteString(line)
			result.WriteString("\n")
		}
		return
	}

	// Write content before start marker
	startIdx := -1
	endIdx := len(contentLines)

	for i, line := range contentLines {
		if strings.Contains(line, startComment) {
			startIdx = i
			break
		}
	}

	// Write everything before the start marker
	for i := range startIdx {
		result.WriteString(contentLines[i])
		result.WriteString("\n")
	}

	// Write the injected content
	result.WriteString(startComment)
	result.WriteString("\n\n")
	result.WriteString(markdownContent)
	result.WriteString("\n\n")
	result.WriteString(endComment)
	result.WriteString("\n")

	// If there was an end marker, find it and write content after it
	if foundEndMarker {
		for i := startIdx + 1; i < len(contentLines); i++ {
			if strings.Contains(contentLines[i], endComment) {
				endIdx = i
				break
			}
		}
		// Write content after end marker
		for i := endIdx + 1; i < len(contentLines); i++ {
			result.WriteString(contentLines[i])
			result.WriteString("\n")
		}
	}
}

// isClosingDelimiter checks if a line is the closing delimiter for a heredoc.
func isClosingDelimiter(line, indent, delimiter string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == delimiter || trimmed == indent+delimiter
}

// isSingleLineDescription checks if a description is in single-line format.
func isSingleLineDescription(line string) bool {
	trimmed := strings.TrimSpace(line)
	// Single line if it has both description = "..." on one line
	return strings.Contains(trimmed, "description") &&
		strings.Contains(trimmed, "=") &&
		strings.Contains(trimmed, "\"") &&
		!strings.Contains(trimmed, "<<")
}

// convertToMultilineDescription converts a single-line description to multiline heredoc format with injection.
func convertToMultilineDescription(line, marinatedID, markdownContent string) string {
	// Extract indentation
	indent := getIndentation(line)

	// Build the multiline description
	var result strings.Builder
	result.WriteString(indent)
	result.WriteString("description = <<-EOT\n")
	result.WriteString(fmt.Sprintf("<!-- MARINATED: %s -->\n\n", marinatedID))
	result.WriteString(markdownContent)
	result.WriteString("\n\n")
	result.WriteString(fmt.Sprintf("<!-- /MARINATED: %s -->\n", marinatedID))
	result.WriteString(indent)
	result.WriteString("EOT")

	return result.String()
}

// getIndentation extracts the leading whitespace from a line.
func getIndentation(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return ""
}

// FindMarkers scans the Terraform module for all MARINATED markers.
// Returns a slice of marinated IDs found.
func (ti *TerraformInjector) FindMarkers() ([]string, error) {
	parser := NewParser()
	if err := parser.ParseVariables(ti.modulePath); err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	marinatedVars, err := parser.ExtractMarinatedVars()
	if err != nil {
		return nil, fmt.Errorf("failed to extract marinated variables: %w", err)
	}

	markers := make([]string, 0, len(marinatedVars))
	for _, v := range marinatedVars {
		markers = append(markers, v.MarinatedID)
	}

	return markers, nil
}
