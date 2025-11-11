package hover

import (
	"fmt"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides hover information
type Provider struct {
	statements map[string]data.MCSStatement
}

// NewProvider creates a new hover provider with shared data
func NewProvider(store *data.Store) *Provider {
	return &Provider{
		statements: store.Statements,
	}
}

// GetHoverAST returns hover information using AST-based lookup
func (p *Provider) GetHoverAST(doc *parser.Document, line, character int) *lsp.Hover {
	if doc == nil {
		return nil
	}

	// Find the node at the cursor position
	node := p.findNodeAtPosition(doc, line, character)
	if node == nil {
		return nil
	}

	logger.Debug("Hover node type: %v, name: %s", node.Type, node.Name)

	switch node.Type {
	case parser.NodeTypeStatement:
		if node.StatementDef != nil {
			stmt := *node.StatementDef
			// Add language ID info if present
			if node.LanguageID != "" {
				stmt.Description = fmt.Sprintf("%s\n\n**Language:** %s", stmt.Description, node.LanguageID)
			}
			return p.createStatementHover(stmt)
		}
	case parser.NodeTypeOperand:
		if node.OperandDef != nil {
			return p.createOperandHover(*node.OperandDef)
		}
	case parser.NodeTypeParameter:
		// No hover for parameter values
		return nil
	}

	return nil
}

// findNodeAtPosition finds the AST node at the given position
func (p *Provider) findNodeAtPosition(doc *parser.Document, line, character int) *parser.Node {
	// Search through all statements
	for _, stmt := range doc.Statements {
		if node := p.findNodeInTree(stmt, line, character); node != nil {
			return node
		}
	}
	return nil
}

// findNodeInTree recursively searches for a node at the given position
func (p *Provider) findNodeInTree(node *parser.Node, line, character int) *parser.Node {
	if node == nil {
		return nil
	}

	// Check if position is within this node's range
	if line == node.Position.Line {
		nodeEnd := node.Position.Character + node.Position.Length
		if character >= node.Position.Character && character < nodeEnd {
			// Check children first (more specific match)
			for _, child := range node.Children {
				if childNode := p.findNodeInTree(child, line, character); childNode != nil {
					return childNode
				}
			}
			// Return this node if no child matched
			return node
		}
	}

	// Check children even if line doesn't match (for multiline statements)
	for _, child := range node.Children {
		if childNode := p.findNodeInTree(child, line, character); childNode != nil {
			return childNode
		}
	}

	return nil
}

// createStatementHover creates hover info for a statement
func (p *Provider) createStatementHover(stmt data.MCSStatement) *lsp.Hover {
	content := fmt.Sprintf("**%s**\n\n%s\n\n", stmt.Name, stmt.Description)

	if stmt.Parameter != "" {
		content += fmt.Sprintf("**Parameter:** %s\n\n", stmt.Parameter)
	}

	if len(stmt.Operands) > 0 {
		content += "**Operands:**\n"
		for _, op := range stmt.Operands {
			content += fmt.Sprintf("- `%s`: %s\n", op.Name, op.Description)
		}
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupKindMarkdown,
			Value: content,
		},
	}
}

// createOperandHover creates hover info for an operand
func (p *Provider) createOperandHover(operand data.Operand) *lsp.Hover {
	content := fmt.Sprintf("**%s**\n\n%s\n\n", operand.Name, operand.Description)

	if operand.Parameter != "" {
		content += fmt.Sprintf("**Parameter:** %s\n\n", operand.Parameter)
	}

	if operand.Type != "" {
		content += fmt.Sprintf("**Type:** %s\n\n", operand.Type)
	}

	if operand.Length > 0 {
		content += fmt.Sprintf("**Length:** %d\n\n", operand.Length)
	}

	if operand.MutuallyExclusive != "" {
		content += fmt.Sprintf("**Mutually Exclusive with:** %s\n\n", operand.MutuallyExclusive)
	}

	if len(operand.Values) > 0 {
		content += "**Allowed Values:**\n"
		for _, val := range operand.Values {
			if val.Description != "" {
				content += fmt.Sprintf("- `%s`: %s\n", val.Name, val.Description)
			} else {
				content += fmt.Sprintf("- `%s`\n", val.Name)
			}
		}
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupKindMarkdown,
			Value: content,
		},
	}
}
