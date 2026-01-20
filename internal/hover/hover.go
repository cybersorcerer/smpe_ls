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
	// Header with statement name, parameter syntax and type
	content := fmt.Sprintf("**%s**", stmt.Name)
	if stmt.Parameter != "" {
		content += fmt.Sprintf(" `(%s)`", stmt.Parameter)
	}
	if stmt.Type != "" {
		content += fmt.Sprintf(" — *%s*", stmt.Type)
	}
	content += "\n\n"

	// Description (may contain markdown from smpe.json)
	content += stmt.Description + "\n\n"

	// Collect required and optional operands
	var requiredOps []data.Operand
	var optionalOps []data.Operand

	for _, op := range stmt.Operands {
		if op.Required {
			requiredOps = append(requiredOps, op)
		} else {
			optionalOps = append(optionalOps, op)
		}
	}

	// Required operands section
	if len(requiredOps) > 0 {
		content += "---\n\n"
		content += "**Required Operands:**\n"
		for _, op := range requiredOps {
			content += p.formatOperandListItem(op)
		}
		content += "\n"
	}

	// Optional operands section (limited to first 10)
	if len(optionalOps) > 0 {
		if len(requiredOps) == 0 {
			content += "---\n\n"
		}
		content += "**Optional Operands:**\n"
		displayCount := len(optionalOps)
		if displayCount > 10 {
			displayCount = 10
		}
		for i := 0; i < displayCount; i++ {
			content += p.formatOperandListItem(optionalOps[i])
		}
		if len(optionalOps) > 10 {
			content += fmt.Sprintf("- *... and %d more*\n", len(optionalOps)-10)
		}
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupKindMarkdown,
			Value: content,
		},
	}
}

// formatOperandListItem formats a single operand for the operand list
func (p *Provider) formatOperandListItem(op data.Operand) string {
	// Use primary name (before any | alias separator)
	primaryName := op.Name
	if idx := len(op.Name); idx > 0 {
		for i, c := range op.Name {
			if c == '|' {
				primaryName = op.Name[:i]
				break
			}
		}
	}

	var item string
	if op.Parameter != "" {
		item = fmt.Sprintf("- `%s(%s)`\n", primaryName, op.Parameter)
	} else {
		item = fmt.Sprintf("- `%s`\n", primaryName)
	}
	return item
}

// createOperandHover creates hover info for an operand
func (p *Provider) createOperandHover(operand data.Operand) *lsp.Hover {
	// Header with operand name and parameter syntax
	content := fmt.Sprintf("**%s**", operand.Name)
	if operand.Parameter != "" {
		content += fmt.Sprintf(" `(%s)`", operand.Parameter)
	}
	if operand.Required {
		content += " — *required*"
	}
	content += "\n\n"

	// Description (may contain markdown from smpe.json)
	if operand.Description != "" {
		content += operand.Description + "\n\n"
	}

	// Type and length info on one line
	if operand.Type != "" || operand.Length > 0 {
		content += "---\n\n"
		if operand.Type != "" && operand.Length > 0 {
			content += fmt.Sprintf("**Type:** %s (max %d chars)\n\n", operand.Type, operand.Length)
		} else if operand.Type != "" {
			content += fmt.Sprintf("**Type:** %s\n\n", operand.Type)
		} else if operand.Length > 0 {
			content += fmt.Sprintf("**Max Length:** %d\n\n", operand.Length)
		}
	}

	// Mutually exclusive operands as list
	if operand.MutuallyExclusive != "" {
		content += "**Cannot be used with:**\n"
		exclusives := splitByPipe(operand.MutuallyExclusive)
		for _, ex := range exclusives {
			content += fmt.Sprintf("- `%s`\n", ex)
		}
		content += "\n"
	}

	// Allowed values
	if len(operand.Values) > 0 {
		content += "**Allowed Values:**\n"
		displayCount := len(operand.Values)
		if displayCount > 8 {
			displayCount = 8
		}
		for i := 0; i < displayCount; i++ {
			val := operand.Values[i]
			if val.Description != "" {
				// Truncate long descriptions
				desc := val.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				content += fmt.Sprintf("- `%s` — %s\n", val.Name, desc)
			} else {
				content += fmt.Sprintf("- `%s`\n", val.Name)
			}
		}
		if len(operand.Values) > 8 {
			content += fmt.Sprintf("- *... and %d more*\n", len(operand.Values)-8)
		}
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupKindMarkdown,
			Value: content,
		},
	}
}

// splitByPipe splits a string by pipe character
func splitByPipe(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == '|' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
