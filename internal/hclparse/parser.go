package hclparse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Common errors.
var (
	ErrNotImplemented = errors.New("not yet implemented")
)

// Parser handles parsing of HCL files (variables.tf) to extract variable definitions.
type Parser struct {
	variables []*Variable
}

// NewParser creates a new HCL parser instance.
func NewParser() *Parser {
	return &Parser{
		variables: make([]*Variable, 0),
	}
}

// ParseVariables parses all variables.*.tf files in the given directory
// ParseVariables scans the module path for variables.tf files
// and extracts variable definitions, particularly those marked with MARINATED comments.
func (p *Parser) ParseVariables(modulePath string) error {
	// Find all variables.*.tf files
	pattern := filepath.Join(modulePath, "variables*.tf")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob for variables files: %w", err)
	}

	parser := hclparse.NewParser()

	for _, filename := range matches {
		fileContent, readErr := os.ReadFile(filename)
		if readErr != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, readErr)
		}

		file, diags := parser.ParseHCL(fileContent, filename)
		if diags.HasErrors() {
			return fmt.Errorf("failed to parse HCL in %s: %w", filename, diags)
		}

		// Extract variable blocks
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			return fmt.Errorf("unexpected body type in %s", filename)
		}

		for _, block := range body.Blocks {
			if block.Type != "variable" {
				continue
			}

			if len(block.Labels) != 1 {
				continue
			}

			variable, parseErr := p.parseVariableBlock(block)
			if parseErr != nil {
				return fmt.Errorf("failed to parse variable %s in %s: %w", block.Labels[0], filename, parseErr)
			}

			p.variables = append(p.variables, variable)
		}
	}

	return nil
}

// parseVariableBlock extracts a Variable from an HCL variable block.
func (p *Parser) parseVariableBlock(block *hclsyntax.Block) (*Variable, error) {
	varName := block.Labels[0]

	variable := &Variable{
		Name: varName,
	}

	// Extract attributes
	for name, attr := range block.Body.Attributes {
		switch name {
		case "type":
			// Store the raw type expression as a string
			typeStr := extractTypeString(attr.Expr)
			variable.Type = typeStr

		case "description":
			// Extract description value
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to evaluate description: %w", diags)
			}
			if val.Type() == cty.String {
				variable.Description = val.AsString()
			}

		case "default":
			// We're skipping default handling per requirements
			continue
		}
	}

	// Check for MARINATED marker
	if variable.Description != "" {
		marinatedID, found := ExtractMarinatedID(variable.Description)
		if found {
			variable.Marinated = true
			variable.MarinatedID = marinatedID
		}
	}

	return variable, nil
}

// extractTypeString converts an HCL type expression to a string representation.
func extractTypeString(expr hclsyntax.Expression) string {
	// Get the source bytes from the expression's range
	srcRange := expr.Range()
	if srcRange.Filename == "" {
		// Fallback: try to reconstruct from the expression
		return reconstructTypeExpr(expr)
	}

	// Read the file and extract the bytes
	content, err := os.ReadFile(srcRange.Filename)
	if err != nil {
		return reconstructTypeExpr(expr)
	}

	return string(srcRange.SliceBytes(content))
}

// reconstructTypeExpr attempts to reconstruct a type expression from its AST.
func reconstructTypeExpr(expr hclsyntax.Expression) string {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		// Simple type like "string", "number", "bool"
		return e.Traversal.RootName()
	case *hclsyntax.FunctionCallExpr:
		// Function call like "optional(...)", "list(...)", "object(...)"
		args := make([]string, len(e.Args))
		for i, arg := range e.Args {
			args[i] = reconstructTypeExpr(arg)
		}
		return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
	case *hclsyntax.ObjectConsExpr:
		// Object construction like { key = value }
		items := make([]string, 0, len(e.Items))
		for _, item := range e.Items {
			keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
			if !ok {
				continue
			}
			// The key is typically wrapped in an ObjectConsKeyExpr
			key := ""
			if scopeTraversal, isScope := keyExpr.Wrapped.(*hclsyntax.ScopeTraversalExpr); isScope {
				key = scopeTraversal.Traversal.RootName()
			}
			valueStr := reconstructTypeExpr(item.ValueExpr)
			items = append(items, fmt.Sprintf("%s = %s", key, valueStr))
		}
		return fmt.Sprintf("{%s}", strings.Join(items, ", "))
	default:
		// Fallback
		return fmt.Sprintf("%T", expr)
	}
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
	marinated := make([]*Variable, 0)
	for _, v := range p.variables {
		if v.Marinated {
			marinated = append(marinated, v)
		}
	}
	return marinated, nil
}

// ExtractMarinatedID extracts the ID from a MARINATED marker in a description.
// Returns the ID and true if found, empty string and false otherwise.
func ExtractMarinatedID(description string) (string, bool) {
	// Pattern: <!-- MARINATED: <id> -->
	// Allow for spaces around the ID and handle escaped underscores (\_)
	re := regexp.MustCompile(`<!--\s*MARINATED:\s*([a-zA-Z0-9_\\]+)\s*-->`)
	matches := re.FindStringSubmatch(description)
	if len(matches) >= 2 && strings.TrimSpace(matches[1]) != "" {
		// Remove backslash escapes from the ID (e.g., configure\_adds\_resources -> configure_adds_resources)
		id := strings.ReplaceAll(strings.TrimSpace(matches[1]), `\`, "")
		return id, true
	}
	return "", false
}
