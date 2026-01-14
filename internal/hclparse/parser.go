package hclparse

import "errors"

// Common errors.
var (
	ErrNotImplemented = errors.New("not yet implemented")
)

// Parser handles parsing of HCL files (variables.tf) to extract variable definitions.
type Parser struct {
	// TODO: Add fields for HCL parsing state
}

// NewParser creates a new HCL parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// ParseVariables parses all variables.*.tf files in the given directory
// ParseVariables scans the module path for variables.tf files
// and extracts variable definitions, particularly those marked with MARINATED comments.
func (p *Parser) ParseVariables(_ string) error {
	// TODO: Implement HCL parsing logic
	// - Scan for variables.*.tf files
	// - Parse HCL structure
	// - Extract variable definitions with type information
	// - Identify MARINATED markers in descriptions
	return nil
}

// Variable represents a parsed Terraform/OpenTofu variable.
type Variable struct {
	Name        string
	Type        string // HCL type expression
	Description string
	Default     any
	Marinated   bool   // Whether this variable has a MARINATED marker
	MarinatedID string // The ID after "MARINATED:" in the description
}

// ExtractMarinatedVars returns only variables marked with MARINATED comments.
func (p *Parser) ExtractMarinatedVars() ([]*Variable, error) {
	// TODO: Filter and return only marinated variables.
	return nil, ErrNotImplemented
}
