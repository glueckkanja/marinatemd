package schema

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/glueckkanja/marinatemd/internal/hclparse"
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
// Each node can have _marinate metadata (all schema information) and nested attribute nodes.
// All schema metadata is stored under _marinate for clean separation.
// Child attributes are stored as inline YAML fields to avoid extra nesting.
type Node struct {
	Marinate   *MarinateInfo    `yaml:"_marinate,omitempty"` // All schema metadata
	Attributes map[string]*Node `yaml:",inline"`             // Child attributes inlined
}

// MarinateInfo contains all schema metadata for a node.
// This includes both user-editable documentation and technical schema fields.
type MarinateInfo struct {
	Description     string `yaml:"description,omitempty"`      // User-editable description
	ShowDescription *bool  `yaml:"show_description,omitempty"` // Control visibility of description in rendered output (nil = true by default)
	Example         any    `yaml:"example,omitempty"`          // Example value for documentation
	Type            string `yaml:"type,omitempty"`             // Type information (string, number, bool, object, list, map, etc.)
	Required        bool   `yaml:"required,omitempty"`         // Whether this field is required
	ElementType     string `yaml:"element_type,omitempty"`     // For list/set types, the element type
	ValueType       string `yaml:"value_type,omitempty"`       // For map types, the value type
	Default         any    `yaml:"default,omitempty"`          // Default value for optional fields
}

// UnmarshalYAML implements custom YAML unmarshaling for Node.
// This separates _marinate metadata from child attributes and direct fields.
func (n *Node) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", value.Kind)
	}

	// Initialize attributes map
	n.Attributes = make(map[string]*Node)

	// Known fields that are part of the Node struct
	knownFields := map[string]bool{
		"_marinate":    true,
		"type":         true,
		"required":     true,
		"element_type": true,
		"value_type":   true,
		"default":      true,
	}

	// Iterate through the mapping node
	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valueNode := value.Content[i+1]
		fieldName := keyNode.Value

		switch fieldName {
		case "_marinate":
			// Unmarshal the _marinate metadata
			var marinate MarinateInfo
			if err := valueNode.Decode(&marinate); err != nil {
				return fmt.Errorf("failed to decode _marinate: %w", err)
			}
			n.Marinate = &marinate
		default:
			// All other fields are child attributes
			if !knownFields[fieldName] {
				var childNode Node
				if err := valueNode.Decode(&childNode); err != nil {
					return fmt.Errorf("failed to decode attribute %s: %w", fieldName, err)
				}
				n.Attributes[fieldName] = &childNode
			}
		}
	}

	return nil
}

// MarshalYAML implements custom YAML marshaling for Node.
// This ensures _marinate and direct fields are marshaled first, followed by attributes.
func (n *Node) MarshalYAML() (any, error) {
	// Create a yaml.Node to represent this Node
	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	// Add _marinate first if it exists
	if n.Marinate != nil {
		marinateKey := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "_marinate",
		}

		marinateValue := &yaml.Node{}
		if err := marinateValue.Encode(n.Marinate); err != nil {
			return nil, fmt.Errorf("failed to encode _marinate: %w", err)
		}

		node.Content = append(node.Content, marinateKey, marinateValue)
	}

	// Add attributes in sorted order for deterministic output
	if len(n.Attributes) > 0 {
		for _, name := range sortedKeys(n.Attributes) {
			attr := n.Attributes[name]

			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: name,
			}

			valueNode := &yaml.Node{}
			if err := valueNode.Encode(attr); err != nil {
				return nil, fmt.Errorf("failed to encode attribute %s: %w", name, err)
			}

			node.Content = append(node.Content, keyNode, valueNode)
		}
	}

	return node, nil
}

// sortedKeys returns a sorted slice of keys from the Attributes map.
// This ensures deterministic YAML output.
func sortedKeys(attributes map[string]*Node) []string {
	keys := make([]string, 0, len(attributes))
	for name := range attributes {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
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
			Marinate: &MarinateInfo{
				Description: fmt.Sprintf("# TODO: Add description for %s", fieldName),
			},
			Attributes: make(map[string]*Node),
		}

		// Determine if field is optional
		isOptional := strings.HasPrefix(fieldType, "optional(")
		node.Marinate.Required = !isOptional

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
		fullArgs := extractFunctionArg(typeExpr, "optional")
		// optional() can have a second argument (default value)
		innerType := extractFirstArg(fullArgs)
		// Try to extract default value (second argument)
		defaultValue := extractSecondArg(fullArgs)
		if defaultValue != "" {
			node.Marinate.Default = parseDefaultValue(defaultValue)
		}
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
		node.Marinate.Type = "set"
		innerType := extractFunctionArg(typeExpr, "set")
		node.Marinate.ElementType = b.simplifyType(innerType)
		return nil
	}

	// Handle map type
	if strings.HasPrefix(typeExpr, "map(") {
		return b.parseMapFieldType(typeExpr, node)
	}

	// Simple types
	node.Marinate.Type = typeExpr
	return nil
}

// parseObjectFieldType parses an object type and its nested fields.
func (b *Builder) parseObjectFieldType(typeExpr string, node *Node) error {
	node.Marinate.Type = "object"
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
	node.Marinate.Type = "list"
	innerType := extractFunctionArg(typeExpr, "list")
	node.Marinate.ElementType = b.simplifyType(innerType)
	// If list contains objects, parse them as children
	if strings.HasPrefix(innerType, "object(") {
		return b.parseNestedObjectChildren(innerType, node)
	}
	return nil
}

// parseMapFieldType parses a map type and its value type.
func (b *Builder) parseMapFieldType(typeExpr string, node *Node) error {
	node.Marinate.Type = "map"
	innerType := extractFunctionArg(typeExpr, "map")
	node.Marinate.ValueType = b.simplifyType(innerType)
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
			Marinate: &MarinateInfo{
				Description: fmt.Sprintf("# TODO: Add description for %s", name),
			},
			Attributes: make(map[string]*Node),
		}
		isOptional := strings.HasPrefix(fieldType, "optional(")
		childNode.Marinate.Required = !isOptional
		if err := b.parseFieldType(fieldType, childNode, name); err != nil {
			return err
		}
		node.Attributes[name] = childNode
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
		Marinate: &MarinateInfo{
			Description: fmt.Sprintf("# TODO: Add description for %s", contextName),
			Type:        "list",
			Required:    true,
		},
		Attributes: make(map[string]*Node),
	}

	innerType := extractFunctionArg(typeExpr, "list")
	node.Marinate.ElementType = b.simplifyType(innerType)

	nodes["_root"] = node
	return nil
}

// parseSetType parses a set type.
func (b *Builder) parseSetType(typeExpr string, nodes map[string]*Node, contextName string) error {
	node := &Node{
		Marinate: &MarinateInfo{
			Description: fmt.Sprintf("# TODO: Add description for %s", contextName),
			Type:        "set",
			Required:    true,
		},
		Attributes: make(map[string]*Node),
	}

	innerType := extractFunctionArg(typeExpr, "set")
	node.Marinate.ElementType = b.simplifyType(innerType)

	nodes["_root"] = node
	return nil
}

// parseMapType parses a map type.
func (b *Builder) parseMapType(typeExpr string, nodes map[string]*Node, contextName string) error {
	node := &Node{
		Marinate: &MarinateInfo{
			Description: fmt.Sprintf("# TODO: Add description for %s", contextName),
			Type:        "map",
			Required:    true,
		},
		Attributes: make(map[string]*Node),
	}

	innerType := extractFunctionArg(typeExpr, "map")
	node.Marinate.ValueType = b.simplifyType(innerType)

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
				Marinate: &MarinateInfo{
					Description: fmt.Sprintf("# TODO: Add description for %s", name),
				},
				Attributes: make(map[string]*Node),
			}
			isOptional := strings.HasPrefix(fieldType, "optional(")
			childNode.Marinate.Required = !isOptional
			if parseErr5 := b.parseFieldType(fieldType, childNode, name); parseErr5 != nil {
				return parseErr5
			}
			node.Attributes[name] = childNode
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
		Attributes: make(map[string]*Node),
	}

	// Merge Marinate metadata
	merged.Marinate = b.mergeMarinateInfo(newNode.Marinate, existingNode.Marinate)

	// Merge attributes
	for attrName, newAttr := range newNode.Attributes {
		if existingAttr, ok := existingNode.Attributes[attrName]; ok {
			merged.Attributes[attrName] = b.mergeNodes(newAttr, existingAttr)
		} else {
			merged.Attributes[attrName] = newAttr
		}
	}

	return merged
}

// mergeMarinateInfo merges Marinate metadata from new and existing nodes.
// Preserves user-written descriptions if they're not TODO placeholders.
func (b *Builder) mergeMarinateInfo(newInfo, existingInfo *MarinateInfo) *MarinateInfo {
	if newInfo == nil && existingInfo == nil {
		return nil
	}

	merged := &MarinateInfo{}

	// Start with new node's values
	if newInfo != nil {
		merged.Description = newInfo.Description
		merged.Example = newInfo.Example
		merged.Type = newInfo.Type
		merged.Required = newInfo.Required
		merged.ElementType = newInfo.ElementType
		merged.ValueType = newInfo.ValueType
		merged.Default = newInfo.Default
	}

	// Preserve existing user-written descriptions if they're not TODO placeholders
	if existingInfo != nil {
		if existingInfo.Description != "" && !b.isTODO(existingInfo.Description) {
			merged.Description = existingInfo.Description
		}
		if existingInfo.Example != nil {
			merged.Example = existingInfo.Example
		}
	}

	return merged
}

// isTODO checks if a description is a TODO placeholder.
func (b *Builder) isTODO(desc string) bool {
	re := regexp.MustCompile(`(?i)#\s*TODO`)
	return re.MatchString(desc)
}

// extractSecondArg extracts the second argument from a comma-separated list.
// Returns empty string if there's no second argument.
func extractSecondArg(args string) string {
	depth := 0
	for i, ch := range args {
		switch ch {
		case '(', '{', '[':
			depth++
		case ')', '}', ']':
			depth--
		case ',':
			if depth == 0 {
				// Found the comma, return everything after it (trimmed)
				secondArg := strings.TrimSpace(args[i+1:])
				return secondArg
			}
		}
	}
	return ""
}

// parseDefaultValue converts a default value string from HCL to a Go value.
// This handles strings, numbers, bools, lists, and maps.
func parseDefaultValue(defaultStr string) any {
	defaultStr = strings.TrimSpace(defaultStr)

	// Handle null
	if defaultStr == "null" {
		return nil
	}

	// Handle boolean
	if defaultStr == "true" {
		return true
	}
	if defaultStr == "false" {
		return false
	}

	// Handle strings (quoted)
	if strings.HasPrefix(defaultStr, "\"") && strings.HasSuffix(defaultStr, "\"") {
		// Unescape the string
		unquoted := defaultStr[1 : len(defaultStr)-1]
		return unquoted
	}

	// Handle empty list
	if defaultStr == "[]" {
		return []any{}
	}

	// Handle list
	if strings.HasPrefix(defaultStr, "[") && strings.HasSuffix(defaultStr, "]") {
		return parseListDefault(defaultStr)
	}

	// Handle empty map/object
	if defaultStr == "{}" {
		return map[string]any{}
	}

	// Handle map/object
	if strings.HasPrefix(defaultStr, "{") && strings.HasSuffix(defaultStr, "}") {
		return parseMapDefault(defaultStr)
	}

	// Try to parse as number
	// For simplicity, we'll return the string as-is since YAML can handle numeric strings
	// This avoids complex float/int parsing
	return defaultStr
}

// parseListDefault parses a list default value like ["a", "b"] or [1, 2].
func parseListDefault(listStr string) []any {
	listStr = strings.TrimSpace(listStr)
	if listStr == "[]" {
		return []any{}
	}

	// Remove brackets
	content := listStr[1 : len(listStr)-1]
	content = strings.TrimSpace(content)

	if content == "" {
		return []any{}
	}

	// Split by comma, respecting nested structures
	items := splitByComma(content)
	result := make([]any, 0, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)
		result = append(result, parseDefaultValue(item))
	}

	return result
}

// parseMapDefault parses a map/object default value.
func parseMapDefault(mapStr string) map[string]any {
	mapStr = strings.TrimSpace(mapStr)
	if mapStr == "{}" {
		return map[string]any{}
	}

	// Remove braces
	content := mapStr[1 : len(mapStr)-1]
	content = strings.TrimSpace(content)

	if content == "" {
		return map[string]any{}
	}

	// For now, return a simple representation
	// Full parsing would require more complex logic
	return map[string]any{}
}

// splitByComma splits a string by commas, respecting nested brackets.
func splitByComma(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		switch ch {
		case '(', '{', '[':
			depth++
			current.WriteRune(ch)
		case ')', '}', ']':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Add the last item
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
