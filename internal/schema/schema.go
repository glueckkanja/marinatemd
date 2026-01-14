package schema

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/c4a8-azure/marinatemd/internal/hclparse"
)

// Common errors.
var (
	ErrNotImplemented = errors.New("not yet implemented")
)

// Schema represents the internal schema model for a Terraform variable.
type Schema struct {
	Variable    string           `yaml:"variable"`
	Version     string           `yaml:"version"`
	SchemaNodes map[string]*Node `yaml:"schema"`
}

// Node represents a node in the schema tree.
// Can be a complex object with nested children or a leaf attribute.
// According to Option B: nodes can have both leaf fields AND children for nested objects.
type Node struct {
	Meta     *MetaInfo        `yaml:"_meta,omitempty"`
	Children map[string]*Node `yaml:",inline"` // Inline children at same level as other fields

	// Fields present at both leaf AND parent nodes
	Type        string `yaml:"type,omitempty"`        // Type information (string, number, bool, object, list, map, etc.)
	Required    bool   `yaml:"required,omitempty"`    // Whether this field is required
	Description string `yaml:"description,omitempty"` // Description for leaf nodes
	Example     any    `yaml:"example,omitempty"`     // Example value for documentation

	// Additional type information for complex types
	ElementType string `yaml:"element_type,omitempty"` // For list/set types, the element type
	ValueType   string `yaml:"value_type,omitempty"`   // For map types, the value type
}

// MetaInfo contains metadata for complex objects (non-leaf nodes).
type MetaInfo struct {
	Description string `yaml:"description,omitempty"`
}

// Builder creates schema models from parsed HCL variables.
type Builder struct {
}

// NewBuilder creates a new schema builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildFromVariable converts an HCL variable to a Schema model.
func (b *Builder) BuildFromVariable(variable *hclparse.Variable) (*Schema, error) {
	schema := &Schema{
		Variable:    variable.MarinatedID,
		Version:     "1",
		SchemaNodes: make(map[string]*Node),
	}

	// Parse the type expression and build the schema tree
	if err := b.parseType(variable.Type, schema.SchemaNodes, variable.MarinatedID); err != nil {
		return nil, fmt.Errorf("failed to parse type for variable %s: %w", variable.Name, err)
	}

	return schema, nil
}

// parseType recursively parses a type expression and populates the schema nodes.
func (b *Builder) parseType(typeExpr string, nodes map[string]*Node, contextName string) error {
	typeExpr = strings.TrimSpace(typeExpr)

	// Check for object type
	if strings.HasPrefix(typeExpr, "object(") {
		return b.parseObjectType(typeExpr, nodes, contextName)
	}

	// Check for optional wrapper
	if strings.HasPrefix(typeExpr, "optional(") {
		return b.parseOptionalType(typeExpr, nodes, contextName)
	}

	// Check for list type
	if strings.HasPrefix(typeExpr, "list(") {
		return b.parseListType(typeExpr, nodes, contextName)
	}

	// Check for set type
	if strings.HasPrefix(typeExpr, "set(") {
		return b.parseSetType(typeExpr, nodes, contextName)
	}

	// Check for map type
	if strings.HasPrefix(typeExpr, "map(") {
		return b.parseMapType(typeExpr, nodes, contextName)
	}

	// Simple types (string, number, bool, any)
	// These typically don't add nodes unless explicitly needed
	return nil
}

// parseObjectType parses an object type expression.
func (b *Builder) parseObjectType(typeExpr string, nodes map[string]*Node, contextName string) error {
	// Extract the object definition: object({...})
	if !strings.HasPrefix(typeExpr, "object(") || !strings.HasSuffix(typeExpr, ")") {
		return fmt.Errorf("invalid object type: %s", typeExpr)
	}

	// Extract content between object( and )
	content := typeExpr[len("object("):]
	content = content[:len(content)-1] // Remove trailing )
	content = strings.TrimSpace(content)

	// Remove outer braces if present
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		content = content[1 : len(content)-1]
	}

	// Parse fields
	fields, err := b.parseObjectFields(content)
	if err != nil {
		return err
	}

	// Create nodes for each field
	for fieldName, fieldType := range fields {
		node := &Node{
			Meta: &MetaInfo{
				Description: fmt.Sprintf("# TODO: Add description for %s", fieldName),
			},
			Children: make(map[string]*Node),
		}

		// Determine if field is optional
		isOptional := strings.HasPrefix(fieldType, "optional(")
		node.Required = !isOptional

		// Parse the field type
		if err := b.parseFieldType(fieldType, node, fieldName); err != nil {
			return fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}

		nodes[fieldName] = node
	}

	return nil
}

// parseFieldType parses the type of a single field and populates the node.
func (b *Builder) parseFieldType(typeExpr string, node *Node, fieldName string) error {
	typeExpr = strings.TrimSpace(typeExpr)

	// Handle optional wrapper
	if strings.HasPrefix(typeExpr, "optional(") {
		innerType := extractFunctionArg(typeExpr, "optional")
		// optional() can have a second argument (default value), extract only the first
		innerType = extractFirstArg(innerType)
		return b.parseFieldType(innerType, node, fieldName)
	}

	// Handle object type
	if strings.HasPrefix(typeExpr, "object(") {
		node.Type = "object"
		// Parse nested object fields
		content := extractFunctionArg(typeExpr, "object")
		if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
			content = content[1 : len(content)-1]
		}
		fields, err := b.parseObjectFields(content)
		if err != nil {
			return err
		}
		for name, fieldType := range fields {
			childNode := &Node{
				Children: make(map[string]*Node),
			}
			isOptional := strings.HasPrefix(fieldType, "optional(")
			childNode.Required = !isOptional
			if err := b.parseFieldType(fieldType, childNode, name); err != nil {
				return err
			}
			childNode.Description = fmt.Sprintf("# TODO: Add description for %s", name)
			node.Children[name] = childNode
		}
		return nil
	}

	// Handle list type
	if strings.HasPrefix(typeExpr, "list(") {
		node.Type = "list"
		innerType := extractFunctionArg(typeExpr, "list")
		node.ElementType = b.simplifyType(innerType)
		// If list contains objects, parse them as children
		if strings.HasPrefix(innerType, "object(") {
			content := extractFunctionArg(innerType, "object")
			if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
				content = content[1 : len(content)-1]
			}
			fields, err := b.parseObjectFields(content)
			if err != nil {
				return err
			}
			for name, fieldType := range fields {
				childNode := &Node{
					Children: make(map[string]*Node),
				}
				isOptional := strings.HasPrefix(fieldType, "optional(")
				childNode.Required = !isOptional
				if err := b.parseFieldType(fieldType, childNode, name); err != nil {
					return err
				}
				childNode.Description = fmt.Sprintf("# TODO: Add description for %s", name)
				node.Children[name] = childNode
			}
		}
		return nil
	}

	// Handle set type
	if strings.HasPrefix(typeExpr, "set(") {
		node.Type = "set"
		innerType := extractFunctionArg(typeExpr, "set")
		node.ElementType = b.simplifyType(innerType)
		return nil
	}

	// Handle map type
	if strings.HasPrefix(typeExpr, "map(") {
		node.Type = "map"
		innerType := extractFunctionArg(typeExpr, "map")
		node.ValueType = b.simplifyType(innerType)
		// If map contains objects, parse them as children
		if strings.HasPrefix(innerType, "object(") {
			content := extractFunctionArg(innerType, "object")
			if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
				content = content[1 : len(content)-1]
			}
			fields, err := b.parseObjectFields(content)
			if err != nil {
				return err
			}
			for name, fieldType := range fields {
				childNode := &Node{
					Children: make(map[string]*Node),
				}
				isOptional := strings.HasPrefix(fieldType, "optional(")
				childNode.Required = !isOptional
				if err := b.parseFieldType(fieldType, childNode, name); err != nil {
					return err
				}
				childNode.Description = fmt.Sprintf("# TODO: Add description for %s", name)
				node.Children[name] = childNode
			}
		}
		return nil
	}

	// Simple types
	node.Type = typeExpr
	return nil
}

// extractFirstArg extracts only the first argument from a comma-separated list.
func extractFirstArg(args string) string {
	depth := 0
	for i, ch := range args {
		if ch == '(' || ch == '{' {
			depth++
		} else if ch == ')' || ch == '}' {
			depth--
		} else if ch == ',' && depth == 0 {
			return strings.TrimSpace(args[:i])
		}
	}
	return strings.TrimSpace(args)
}

// simplifyType extracts the base type name from a type expression.
func (b *Builder) simplifyType(typeExpr string) string {
	typeExpr = strings.TrimSpace(typeExpr)
	if strings.HasPrefix(typeExpr, "optional(") {
		return b.simplifyType(extractFunctionArg(typeExpr, "optional"))
	}
	if strings.HasPrefix(typeExpr, "object(") {
		return "object"
	}
	if strings.HasPrefix(typeExpr, "list(") {
		return "list"
	}
	if strings.HasPrefix(typeExpr, "set(") {
		return "set"
	}
	if strings.HasPrefix(typeExpr, "map(") {
		return "map"
	}
	return typeExpr
}

// parseObjectFields parses the fields of an object from its body.
func (b *Builder) parseObjectFields(content string) (map[string]string, error) {
	fields := make(map[string]string)
	content = strings.TrimSpace(content)

	if content == "" {
		return fields, nil
	}

	// Use a more robust regex-based approach to extract fields
	// Pattern: fieldname = type_expression
	// We need to handle nested structures carefully

	var currentField string
	var currentValue strings.Builder
	depth := 0
	inField := false

	i := 0
	for i < len(content) {
		ch := content[i]

		// Track depth for nested structures
		if ch == '(' || ch == '{' {
			depth++
			if inField {
				currentValue.WriteByte(ch)
			}
		} else if ch == ')' || ch == '}' {
			depth--
			if inField {
				currentValue.WriteByte(ch)
			}
		} else if ch == '=' && depth == 0 && !inField {
			// Found field assignment at top level
			// Extract field name backwards
			j := i - 1
			for j >= 0 && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n' || content[j] == '\r') {
				j--
			}
			end := j + 1
			for j >= 0 && (isIdentChar(content[j])) {
				j--
			}
			start := j + 1
			currentField = strings.TrimSpace(content[start:end])
			inField = true
			currentValue.Reset()
		} else if depth == 0 && inField && (ch == '\n' || ch == '\r') {
			// Check if next non-whitespace is a new field or end
			j := i + 1
			for j < len(content) && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n' || content[j] == '\r') {
				j++
			}
			if j >= len(content) || (j < len(content) && isIdentStart(content[j])) {
				// End of this field
				if currentField != "" {
					fields[currentField] = strings.TrimSpace(currentValue.String())
					currentField = ""
					currentValue.Reset()
					inField = false
				}
			} else {
				// Continuation
				if inField {
					currentValue.WriteByte(' ')
				}
			}
		} else if inField {
			currentValue.WriteByte(ch)
		}

		i++
	}

	// Save last field
	if currentField != "" && inField {
		fields[currentField] = strings.TrimSpace(currentValue.String())
	}

	return fields, nil
}

// isIdentChar returns true if ch can be part of an identifier.
func isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

// isIdentStart returns true if ch can start an identifier.
func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// parseOptionalType parses an optional type wrapper.
func (b *Builder) parseOptionalType(typeExpr string, nodes map[string]*Node, contextName string) error {
	innerType := extractFunctionArg(typeExpr, "optional")
	return b.parseType(innerType, nodes, contextName)
}

// parseListType parses a list type.
func (b *Builder) parseListType(typeExpr string, nodes map[string]*Node, contextName string) error {
	node := &Node{
		Type: "list",
		Meta: &MetaInfo{
			Description: fmt.Sprintf("# TODO: Add description for %s", contextName),
		},
		Children: make(map[string]*Node),
		Required: true,
	}

	innerType := extractFunctionArg(typeExpr, "list")
	node.ElementType = b.simplifyType(innerType)

	nodes["_root"] = node
	return nil
}

// parseSetType parses a set type.
func (b *Builder) parseSetType(typeExpr string, nodes map[string]*Node, contextName string) error {
	node := &Node{
		Type: "set",
		Meta: &MetaInfo{
			Description: fmt.Sprintf("# TODO: Add description for %s", contextName),
		},
		Children: make(map[string]*Node),
		Required: true,
	}

	innerType := extractFunctionArg(typeExpr, "set")
	node.ElementType = b.simplifyType(innerType)

	nodes["_root"] = node
	return nil
}

// parseMapType parses a map type.
func (b *Builder) parseMapType(typeExpr string, nodes map[string]*Node, contextName string) error {
	node := &Node{
		Type: "map",
		Meta: &MetaInfo{
			Description: fmt.Sprintf("# TODO: Add description for %s", contextName),
		},
		Children: make(map[string]*Node),
		Required: true,
	}

	innerType := extractFunctionArg(typeExpr, "map")
	node.ValueType = b.simplifyType(innerType)

	// If map contains objects, parse them
	if strings.HasPrefix(innerType, "object(") {
		content := extractFunctionArg(innerType, "object")
		if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
			content = content[1 : len(content)-1]
		}
		fields, err := b.parseObjectFields(content)
		if err != nil {
			return err
		}
		for name, fieldType := range fields {
			childNode := &Node{
				Children: make(map[string]*Node),
			}
			isOptional := strings.HasPrefix(fieldType, "optional(")
			childNode.Required = !isOptional
			if err := b.parseFieldType(fieldType, childNode, name); err != nil {
				return err
			}
			childNode.Description = fmt.Sprintf("# TODO: Add description for %s", name)
			node.Children[name] = childNode
		}
	}

	nodes["_root"] = node
	return nil
}

// extractFunctionArg extracts the argument(s) from a function call.
// For "optional(...)" it returns "..."
func extractFunctionArg(expr, funcName string) string {
	if !strings.HasPrefix(expr, funcName+"(") {
		return ""
	}
	content := expr[len(funcName)+1:]
	// Find matching closing paren
	depth := 1
	for i, ch := range content {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				return strings.TrimSpace(content[:i])
			}
		}
	}
	return strings.TrimSpace(content)
}

// MergeWithExisting merges a new schema with an existing one.
// Preserves user-written descriptions while updating structure.
func (b *Builder) MergeWithExisting(newSchema, existing *Schema) (*Schema, error) {
	merged := &Schema{
		Variable:    newSchema.Variable,
		Version:     newSchema.Version,
		SchemaNodes: make(map[string]*Node),
	}

	// Merge each node from new schema
	for nodeName, newNode := range newSchema.SchemaNodes {
		if existingNode, ok := existing.SchemaNodes[nodeName]; ok {
			// Node exists in both - merge them
			merged.SchemaNodes[nodeName] = b.mergeNodes(newNode, existingNode)
		} else {
			// New node - add it
			merged.SchemaNodes[nodeName] = newNode
		}
	}

	return merged, nil
}

// mergeNodes merges two nodes, preserving descriptions from existing.
func (b *Builder) mergeNodes(newNode, existingNode *Node) *Node {
	merged := &Node{
		Type:        newNode.Type,
		Required:    newNode.Required,
		ElementType: newNode.ElementType,
		ValueType:   newNode.ValueType,
		Children:    make(map[string]*Node),
	}

	// Preserve existing descriptions if they're not TODO placeholders
	if existingNode.Description != "" && !b.isTODO(existingNode.Description) {
		merged.Description = existingNode.Description
	} else {
		merged.Description = newNode.Description
	}

	if existingNode.Example != nil {
		merged.Example = existingNode.Example
	} else {
		merged.Example = newNode.Example
	}

	// Merge Meta
	if newNode.Meta != nil || existingNode.Meta != nil {
		merged.Meta = &MetaInfo{}
		if newNode.Meta != nil {
			merged.Meta.Description = newNode.Meta.Description
		}
		if existingNode.Meta != nil && existingNode.Meta.Description != "" && !b.isTODO(existingNode.Meta.Description) {
			merged.Meta.Description = existingNode.Meta.Description
		}
	}

	// Merge children
	for childName, newChild := range newNode.Children {
		if existingChild, ok := existingNode.Children[childName]; ok {
			merged.Children[childName] = b.mergeNodes(newChild, existingChild)
		} else {
			merged.Children[childName] = newChild
		}
	}

	return merged
}

// isTODO checks if a description is a TODO placeholder.
func (b *Builder) isTODO(desc string) bool {
	re := regexp.MustCompile(`(?i)#\s*TODO`)
	return re.MatchString(desc)
}
