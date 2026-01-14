// Package schema provides types and utilities for managing variable schemas.
// This is the bridge between HCL parsing and YAML/markdown generation.
package schema

import "errors"

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
type Node struct {
	Meta     *MetaInfo        `yaml:"_meta,omitempty"`
	Children map[string]*Node `yaml:"children,omitempty"`

	// Leaf node fields (only populated for leaf attributes)
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Default     any    `yaml:"default,omitempty"`
	Example     any    `yaml:"example,omitempty"`
}

// MetaInfo contains metadata for complex objects (non-leaf nodes).
type MetaInfo struct {
	Description string `yaml:"description,omitempty"`
}

// Builder creates schema models from parsed HCL variables.
type Builder struct {
	// TODO: Add state for schema building.
}

// NewBuilder creates a new schema builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildFromHCL converts an HCL variable definition to a Schema model.
func (b *Builder) BuildFromHCL(_ string, _ string) (*Schema, error) {
	// TODO: Implement conversion from HCL type to Schema
	// - Parse object({...}) structures
	// - Handle optional() wrappers
	// - Extract default values
	// - Determine required vs optional attributes
	return nil, ErrNotImplemented
}

// MergeWithExisting merges new schema structure with existing YAML schema
// MergeWithExisting merges a new schema with an existing one.
// Preserves user-written descriptions while updating structure.
func (b *Builder) MergeWithExisting(_, _ *Schema) (*Schema, error) {
	// TODO: Implement intelligent merging
	// - Keep existing descriptions where fields still exist
	// - Add new fields from updated HCL
	// - Remove fields that no longer exist in HCL
	return nil, ErrNotImplemented
}
