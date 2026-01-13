package schema

// Schema represents the internal schema model for a Terraform variable
// This is the bridge between HCL parsing and YAML/markdown generation
type Schema struct {
	Variable    string                 `yaml:"variable"`
	Version     string                 `yaml:"version"`
	SchemaNodes map[string]*SchemaNode `yaml:"schema"`
}

// SchemaNode represents a node in the variable schema tree
// Can be a complex object with nested children or a leaf attribute
type SchemaNode struct {
	Meta     *MetaInfo              `yaml:"_meta,omitempty"`
	Children map[string]*SchemaNode `yaml:",inline,omitempty"`

	// Leaf node fields (only populated for leaf attributes)
	Description string      `yaml:"description,omitempty"`
	Required    bool        `yaml:"required,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Example     interface{} `yaml:"example,omitempty"`
}

// MetaInfo contains metadata for complex objects (non-leaf nodes)
type MetaInfo struct {
	Description string `yaml:"description,omitempty"`
}

// Builder creates schema models from parsed HCL variables
type Builder struct {
	// TODO: Add state for schema building
}

// NewBuilder creates a new schema builder
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildFromHCL converts an HCL variable definition to a Schema model
func (b *Builder) BuildFromHCL(varName string, hclType string) (*Schema, error) {
	// TODO: Implement conversion from HCL type to Schema
	// - Parse object({...}) structures
	// - Handle optional() wrappers
	// - Extract default values
	// - Determine required vs optional attributes
	return nil, nil
}

// MergeWithExisting merges new schema structure with existing YAML schema
// Preserves user-written descriptions while updating structure
func (b *Builder) MergeWithExisting(newSchema, existingSchema *Schema) (*Schema, error) {
	// TODO: Implement intelligent merging
	// - Keep existing descriptions where fields still exist
	// - Add new fields from updated HCL
	// - Remove fields that no longer exist in HCL
	return nil, nil
}
