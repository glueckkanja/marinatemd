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
// It looks for <!-- MARINATED: variable_name --> markers and replaces them with the provided markdown content.
// The file is read, modified, and written back atomically.
func (i *Injector) InjectIntoFile(filePath string, variableName string, markdownContent string) error {
	// Read the entire file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Convert to string for easier manipulation
	fileContent := string(content)

	// Build the marker to find - try both with escaped and unescaped underscores
	marker := fmt.Sprintf("<!-- MARINATED: %s -->", variableName)
	escapedMarker := fmt.Sprintf("<!-- MARINATED: %s -->", strings.ReplaceAll(variableName, "_", "\\_"))

	// Check if either marker exists
	foundMarker := marker
	if !strings.Contains(fileContent, marker) {
		if strings.Contains(fileContent, escapedMarker) {
			foundMarker = escapedMarker
		} else {
			return fmt.Errorf("marker %s not found in file", marker)
		}
	}

	// Use a more sophisticated approach: find the marker and replace just the marker line
	// while preserving everything else
	lines := strings.Split(fileContent, "\n")
	var result strings.Builder

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check if this line contains our marker
		if strings.Contains(line, foundMarker) {
			// Extract any prefix (e.g., "Description: ")
			markerIdx := strings.Index(line, "<!--")
			prefix := ""
			if markerIdx > 0 {
				prefix = line[:markerIdx]
			}

			// Write the prefix and marker on the same line, then content on new lines
			result.WriteString(prefix)
			result.WriteString(foundMarker)
			result.WriteString("\n\n")
			result.WriteString(strings.TrimSpace(markdownContent))

			// Skip any existing content that was previously injected
			// Look ahead to find where the next section starts (e.g., "Type:", "Default:", or next "###")
			i++
			for i < len(lines) {
				nextLine := strings.TrimSpace(lines[i])
				// Stop when we hit the next significant markdown section
				if strings.HasPrefix(nextLine, "Type:") ||
					strings.HasPrefix(nextLine, "Default:") ||
					strings.HasPrefix(nextLine, "###") ||
					strings.HasPrefix(nextLine, "##") ||
					strings.HasPrefix(nextLine, "<!--") {
					i-- // Back up so we don't skip this line
					break
				}
				// Skip lines that are part of the old injected content
				if nextLine == "" {
					// Keep one blank line for spacing
					if i+1 < len(lines) {
						next := strings.TrimSpace(lines[i+1])
						if strings.HasPrefix(next, "Type:") ||
							strings.HasPrefix(next, "Default:") ||
							strings.HasPrefix(next, "###") {
							break
						}
					}
				}
				i++
			}
		} else {
			result.WriteString(line)
		}

		// Add newline except for the last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	// Write the modified content back to the file
	if err := os.WriteFile(filePath, []byte(result.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
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
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		// Look for <!-- MARINATED: variable_name -->
		if strings.Contains(line, "<!-- MARINATED:") {
			// Extract the variable name
			start := strings.Index(line, "<!-- MARINATED:")
			if start == -1 {
				continue
			}

			remaining := line[start+len("<!-- MARINATED:"):]
			end := strings.Index(remaining, "-->")
			if end == -1 {
				continue
			}

			variableName := strings.TrimSpace(remaining[:end])
			// Handle escaped underscores in markdown (e.g., app\_config -> app_config)
			variableName = strings.ReplaceAll(variableName, "\\_", "_")
			if variableName != "" {
				markers = append(markers, variableName)
			}
		}
	}

	return markers, nil
}
