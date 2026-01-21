package schema

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/c4a8-azure/marinatemd/internal/hclparse"
	"gopkg.in/yaml.v3"
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
	Children map[string]*Node `yaml:"children,omitempty"` // Children nodes (not inlined to avoid field name conflicts)

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

// UnmarshalYAML implements custom YAML unmarshaling for Node.
// This handles both the new format (with explicit "children" key) and
// the old format (with inline children) for backward compatibility.
func (n *Node) UnmarshalYAML(value *yaml.Node) error {
	// Define a type alias to avoid recursion
	type nodeAlias Node

	// First, try to unmarshal using the standard struct approach
	var aux nodeAlias
	if err := value.Decode(&aux); err != nil {
		return err
	}

	// Copy the decoded fields to the actual node
	*n = Node(aux)

	// If Children is nil, initialize it
	if n.Children == nil {
		n.Children = make(map[string]*Node)
	}

	// Now check if there are any additional fields at the same level
	// that are not standard fields - these would be inline children from old format
	if value.Kind == yaml.MappingNode {
		knownFields := map[string]bool{
			"_meta":        true,
			"children":     true,
			"type":         true,
			"required":     true,
			"description":  true,
			"example":      true,
			"element_type": true,
			"value_type":   true,
		}

		// Iterate through the mapping node to find inline children
		for i := 0; i < len(value.Content); i += 2 {
			keyNode := value.Content[i]
			valueNode := value.Content[i+1]

			fieldName := keyNode.Value

			// If this is not a known field, it's an inline child (old format)
			if !knownFields[fieldName] {
				var childNode Node
				if err := valueNode.Decode(&childNode); err != nil {
					return fmt.Errorf("failed to decode inline child %s: %w", fieldName, err)
				}
				n.Children[fieldName] = &childNode
			}
		}
	}

	return nil
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
		return b.parseObjectType(typeExpr, nodes)
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
func (b *Builder) parseObjectType(typeExpr string, nodes map[string]*Node) error {
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
		if parseErr := b.parseFieldType(fieldType, node, fieldName); parseErr != nil {
			return fmt.Errorf("failed to parse field %s: %w", fieldName, parseErr)
		}

		nodes[fieldName] = node
	}

	return nil
}

// parseFieldType parses the type of a single field and populates the node.
func (b *Builder) parseFieldType(typeExpr string, node *Node, _fieldName string) error {
	typeExpr = strings.TrimSpace(typeExpr)

	// Handle optional wrapper
	if strings.HasPrefix(typeExpr, "optional(") {
		innerType := extractFunctionArg(typeExpr, "optional")
		// optional() can have a second argument (default value), extract only the first
		innerType = extractFirstArg(innerType)
		return b.parseFieldType(innerType, node, _fieldName)
	}

	// Handle object type
	if strings.HasPrefix(typeExpr, "object(") {
		return b.parseObjectFieldType(typeExpr, node)
	}

	// Handle list type
	if strings.HasPrefix(typeExpr, "list(") {
		return b.parseListFieldType(typeExpr, node)
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
		return b.parseMapFieldType(typeExpr, node)
	}

	// Simple types
	node.Type = typeExpr
	return nil
}

// parseObjectFieldType parses an object type and its nested fields.
func (b *Builder) parseObjectFieldType(typeExpr string, node *Node) error {
	node.Type = "object"
	content := extractFunctionArg(typeExpr, "object")
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		content = content[1 : len(content)-1]
	}
	fields, err := b.parseObjectFields(content)
	if err != nil {
		return err
	}
	return b.populateChildNodes(node, fields)
}

// parseListFieldType parses a list type and its element type.
func (b *Builder) parseListFieldType(typeExpr string, node *Node) error {
	node.Type = "list"
	innerType := extractFunctionArg(typeExpr, "list")
	node.ElementType = b.simplifyType(innerType)
	// If list contains objects, parse them as children
	if strings.HasPrefix(innerType, "object(") {
		return b.parseNestedObjectChildren(innerType, node)
	}
	return nil
}

// parseMapFieldType parses a map type and its value type.
func (b *Builder) parseMapFieldType(typeExpr string, node *Node) error {
	node.Type = "map"
	innerType := extractFunctionArg(typeExpr, "map")
	node.ValueType = b.simplifyType(innerType)
	// If map contains objects, parse them as children
	if strings.HasPrefix(innerType, "object(") {
		return b.parseNestedObjectChildren(innerType, node)
	}
	return nil
}

// parseNestedObjectChildren parses object fields within a list or map.
func (b *Builder) parseNestedObjectChildren(objectTypeExpr string, node *Node) error {
	content := extractFunctionArg(objectTypeExpr, "object")
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		content = content[1 : len(content)-1]
	}
	fields, err := b.parseObjectFields(content)
	if err != nil {
		return err
	}
	return b.populateChildNodes(node, fields)
}

// populateChildNodes creates child nodes from parsed fields.
func (b *Builder) populateChildNodes(node *Node, fields map[string]string) error {
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

// extractFirstArg extracts only the first argument from a comma-separated list.
func extractFirstArg(args string) string {
	depth := 0
	for i, ch := range args {
		switch ch {
		case '(', '{':
			depth++
		case ')', '}':
			depth--
		case ',':
			if depth == 0 {
				return strings.TrimSpace(args[:i])
			}
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

	parser := &fieldParser{
		content: content,
		fields:  fields,
	}
	return parser.parse()
}

// fieldParser helps parse object field definitions.
type fieldParser struct {
	content      string
	fields       map[string]string
	currentField string
	currentValue strings.Builder
	depth        int
	inField      bool
}

// parse processes the content and extracts field definitions.
func (fp *fieldParser) parse() (map[string]string, error) {
	for i := range len(fp.content) {
		ch := fp.content[i]

		switch {
		case ch == '(' || ch == '{':
			fp.handleOpenBracket(ch)
		case ch == ')' || ch == '}':
			fp.handleCloseBracket(ch)
		case ch == '=' && fp.depth == 0 && !fp.inField:
			fp.handleAssignment(i)
		case fp.depth == 0 && fp.inField && (ch == '\n' || ch == '\r'):
			fp.handleNewline(i)
		case fp.inField:
			fp.currentValue.WriteByte(ch)
		}
	}

	// Save last field
	fp.saveCurrentField()
	return fp.fields, nil
}

// handleOpenBracket processes opening brackets/parentheses.
func (fp *fieldParser) handleOpenBracket(ch byte) {
	fp.depth++
	if fp.inField {
		fp.currentValue.WriteByte(ch)
	}
}

// handleCloseBracket processes closing brackets/parentheses.
func (fp *fieldParser) handleCloseBracket(ch byte) {
	fp.depth--
	if fp.inField {
		fp.currentValue.WriteByte(ch)
	}
}

// handleAssignment processes field assignment operators.
func (fp *fieldParser) handleAssignment(pos int) {
	// Extract field name backwards
	j := pos - 1
	for j >= 0 && isWhitespace(fp.content[j]) {
		j--
	}
	end := j + 1
	for j >= 0 && isIdentChar(fp.content[j]) {
		j--
	}
	start := j + 1
	fp.currentField = strings.TrimSpace(fp.content[start:end])
	fp.inField = true
	fp.currentValue.Reset()
}

// handleNewline processes newline characters and determines field boundaries.
func (fp *fieldParser) handleNewline(pos int) {
	// Check if next non-whitespace is a new field or end
	j := pos + 1
	for j < len(fp.content) && isWhitespace(fp.content[j]) {
		j++
	}
	if j >= len(fp.content) || (j < len(fp.content) && isIdentStart(fp.content[j])) {
		// End of this field
		fp.saveCurrentField()
	} else if fp.inField {
		// Continuation
		fp.currentValue.WriteByte(' ')
	}
}

// saveCurrentField saves the current field if one is being processed.
func (fp *fieldParser) saveCurrentField() {
	if fp.currentField != "" && fp.inField {
		fp.fields[fp.currentField] = strings.TrimSpace(fp.currentValue.String())
		fp.currentField = ""
		fp.currentValue.Reset()
		fp.inField = false
	}
}

// isWhitespace returns true if ch is a whitespace character.
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
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
			if parseErr5 := b.parseFieldType(fieldType, childNode, name); parseErr5 != nil {
				return parseErr5
			}
			childNode.Description = fmt.Sprintf("# TODO: Add description for %s", name)
			node.Children[name] = childNode
		}
	}

	nodes["_root"] = node
	return nil
}

// extractFunctionArg extracts the argument(s) from a function call.
// For "optional(...)" it returns "...".
func extractFunctionArg(expr, funcName string) string {
	if !strings.HasPrefix(expr, funcName+"(") {
		return ""
	}
	content := expr[len(funcName)+1:]
	// Find matching closing paren
	depth := 1
	for i, ch := range content {
		switch ch {
		case '(':
			depth++
		case ')':
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
