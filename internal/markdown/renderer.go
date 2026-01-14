package markdown

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/c4a8-azure/marinatemd/internal/schema"
)

// Common errors.
var (
	ErrNotImplemented = errors.New("not yet implemented")
)

// Renderer generates hierarchical markdown from schema models.
type Renderer struct {
	templateCfg *TemplateConfig
}

// NewRenderer creates a new markdown renderer with default template configuration.
func NewRenderer() *Renderer {
	return &Renderer{
		templateCfg: DefaultTemplateConfig(),
	}
}

// NewRendererWithTemplate creates a new markdown renderer with custom template configuration.
func NewRendererWithTemplate(templateCfg *TemplateConfig) *Renderer {
	if templateCfg == nil {
		templateCfg = DefaultTemplateConfig()
	}
	return &Renderer{
		templateCfg: templateCfg,
	}
}

// RenderSchema converts a schema to hierarchical markdown documentation.
func (r *Renderer) RenderSchema(s *schema.Schema) (string, error) {
	if s == nil {
		return "", errors.New("schema cannot be nil")
	}

	var builder strings.Builder

	// Render each top-level node in sorted order for deterministic output
	nodeNames := make([]string, 0, len(s.SchemaNodes))
	for name := range s.SchemaNodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	for _, nodeName := range nodeNames {
		node := s.SchemaNodes[nodeName]
		if err := r.renderNode(nodeName, node, 0, &builder); err != nil {
			return "", fmt.Errorf("failed to render node %s: %w", nodeName, err)
		}
	}

	return builder.String(), nil
}

// renderNode recursively renders a node and its children.
func (r *Renderer) renderNode(name string, node *schema.Node, depth int, builder *strings.Builder) error {
	if node == nil {
		return nil
	}

	// Decide whether this node should be rendered as an attribute entry.
	//
	// - Nodes with an explicit description are always rendered as attributes,
	//   even if they have children.
	// - Leaf nodes (no children) are rendered as attributes so their type and
	//   required/default metadata are visible even without a description.
	hasDescription := node.Description != ""
	isLeaf := len(node.Children) == 0

	if hasDescription || isLeaf {
		ctx := TemplateContext{
			Attribute:   name,
			Required:    node.Required,
			Description: node.Description,
			Type:        node.Type,
		}

		indent := r.templateCfg.FormatIndent(depth)
		rendered := r.templateCfg.RenderAttribute(ctx)
		builder.WriteString(indent)
		builder.WriteString(rendered)
		builder.WriteString("\n")
	} else if node.Meta != nil && node.Meta.Description != "" {
		// For complex objects with only meta description, render the meta
		indent := r.templateCfg.FormatIndent(depth)
		ctx := TemplateContext{
			Attribute:   name,
			Required:    node.Required,
			Description: node.Meta.Description,
			Type:        node.Type,
		}
		rendered := r.templateCfg.RenderAttribute(ctx)
		builder.WriteString(indent)
		builder.WriteString(rendered)
		builder.WriteString("\n")
	}

	// Render children recursively
	if len(node.Children) > 0 {
		childNames := make([]string, 0, len(node.Children))
		for childName := range node.Children {
			childNames = append(childNames, childName)
		}
		sort.Strings(childNames)

		for _, childName := range childNames {
			child := node.Children[childName]
			if err := r.renderNode(childName, child, depth+1, builder); err != nil {
				return err
			}
		}
	}

	return nil
}

// Injector handles injecting generated markdown into documentation files.
type Injector struct{}

// NewInjector creates a new markdown injector.
func NewInjector() *Injector {
	return &Injector{}
}

// InjectIntoFile replaces content at MARINATED markers in a documentation file.
// It looks for <!-- MARINATED: variable_name --> markers and replaces content between
// the start marker and <!-- /MARINATED: variable_name --> end marker.
// The file is read, modified, and written back atomically.
func (i *Injector) InjectIntoFile(filePath string, variableName string, markdownContent string) error {
	// Read the entire file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Convert to string for easier manipulation
	fileContent := string(content)

	// Build the markers to find - try both with escaped and unescaped underscores
	startMarker := fmt.Sprintf("<!-- MARINATED: %s -->", variableName)
	endMarker := fmt.Sprintf("<!-- /MARINATED: %s -->", variableName)

	escapedStartMarker := fmt.Sprintf("<!-- MARINATED: %s -->", strings.ReplaceAll(variableName, "_", "\\_"))
	escapedEndMarker := fmt.Sprintf("<!-- /MARINATED: %s -->", strings.ReplaceAll(variableName, "_", "\\_"))

	// Check if either marker exists and determine which version we're using
	foundStartMarker := startMarker
	foundEndMarker := endMarker

	if !strings.Contains(fileContent, startMarker) {
		if strings.Contains(fileContent, escapedStartMarker) {
			foundStartMarker = escapedStartMarker
			foundEndMarker = escapedEndMarker
		} else {
			return fmt.Errorf("marker %s not found in file", startMarker)
		}
	}

	// Parse the file line by line
	lines := strings.Split(fileContent, "\n")
	var result strings.Builder
	inMarinatedBlock := false
	foundBlock := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		switch {
		case strings.Contains(line, foundStartMarker):
			foundBlock = true
			inMarinatedBlock = true
			i = writeMarinatedBlock(line, foundStartMarker, foundEndMarker, markdownContent, lines, i, &result)
		case strings.Contains(line, foundEndMarker) && !inMarinatedBlock:
			// Skip orphaned end markers
			continue
		default:
			// Write non-marinated content as-is
			result.WriteString(line)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}

		if inMarinatedBlock {
			inMarinatedBlock = false
		}
	}

	if !foundBlock {
		return fmt.Errorf("marker %s not found in file", startMarker)
	}

	// Write the modified content back to the file
	if writeErr := os.WriteFile(filePath, []byte(result.String()), 0600); writeErr != nil {
		return fmt.Errorf("failed to write file: %w", writeErr)
	}

	return nil
}

func writeMarinatedBlock(
	line, foundStartMarker, foundEndMarker, markdownContent string,
	lines []string,
	idx int,
	result *strings.Builder,
) int {
	// Extract any prefix (e.g., "Description: ")
	prefix, _, _ := strings.Cut(line, "<!--")

	// Write the start marker line
	result.WriteString(prefix)
	result.WriteString(foundStartMarker)
	result.WriteString("\n\n")

	// Write the content with proper spacing
	result.WriteString(strings.TrimSpace(markdownContent))
	result.WriteString("\n\n")

	// Write the end marker
	result.WriteString(foundEndMarker)
	result.WriteString("\n")

	// Skip everything until we find the end marker or a significant section
	idx++
	for idx < len(lines) {
		currentLine := lines[idx]

		// If we find an existing end marker, skip it and continue
		if strings.Contains(currentLine, foundEndMarker) {
			break
		}

		nextLine := strings.TrimSpace(currentLine)
		// Stop when we hit the next significant markdown section
		if strings.HasPrefix(nextLine, "Type:") ||
			strings.HasPrefix(nextLine, "Default:") ||
			strings.HasPrefix(nextLine, "###") ||
			strings.HasPrefix(nextLine, "##") {
			idx-- // Back up so we don't skip this line
			break
		}

		idx++
	}
	return idx
}

// FindMarkers scans a file and returns all MARINATED markers found.
// Returns a slice of variable names extracted from <!-- MARINATED: name --> markers.
func (i *Injector) FindMarkers(filePath string) ([]string, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Find all MARINATED markers using a simple string search
	var markers []string

	for line := range strings.SplitSeq(string(content), "\n") {
		// Look for <!-- MARINATED: variable_name -->
		if strings.Contains(line, "<!-- MARINATED:") {
			// Extract the variable name
			before, after, found := strings.Cut(line, "<!-- MARINATED:")
			if !found {
				continue
			}
			_ = before // Unused

			variableWithEnd, _, found := strings.Cut(after, "-->")
			if !found {
				continue
			}

			variableName := strings.TrimSpace(variableWithEnd)
			// Handle escaped underscores in markdown (e.g., app\_config -> app_config)
			variableName = strings.ReplaceAll(variableName, "\\_", "_")
			if variableName != "" {
				markers = append(markers, variableName)
			}
		}
	}

	return markers, nil
}
