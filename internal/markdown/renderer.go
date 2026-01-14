package markdown

import (
	"errors"
	"fmt"
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
		return "", fmt.Errorf("schema cannot be nil")
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

	// Render this node as an attribute if it has a description or is a leaf node
	if node.Description != "" || len(node.Children) == 0 {
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
type Injector struct {
	// TODO: Add state for tracking injection points.
}

// NewInjector creates a new markdown injector.
func NewInjector() *Injector {
	return &Injector{}
}

// InjectIntoFile replaces content between markers in a documentation file
// InjectIntoFile injects generated markdown into a documentation file.
// Looks for <!-- MARINATED: variable_name --> and <!-- /MARINATED: variable_name -->.
func (i *Injector) InjectIntoFile(_ string, _ string, _ string) error {
	// TODO: Implement injection logic
	// - Read file content
	// - Find marker pairs for the given variable
	// - Replace content between markers
	// - Write back to file
	// - Preserve surrounding content exactly
	return nil
}

// FindMarkers scans a file and returns all MARINATED markers found.
func (i *Injector) FindMarkers(_ string) ([]string, error) {
	// TODO: Parse file and extract all <!-- MARINATED: name --> markers.
	return nil, ErrNotImplemented
}
