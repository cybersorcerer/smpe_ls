package symbols

import (
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides document symbol functionality
type Provider struct{}

// NewProvider creates a new symbol provider
func NewProvider() *Provider {
	return &Provider{}
}

// GetDocumentSymbols returns all symbols in a document
func (p *Provider) GetDocumentSymbols(doc *parser.Document, lines []string) []lsp.DocumentSymbol {
	if doc == nil {
		return nil
	}

	var symbols []lsp.DocumentSymbol

	for _, stmt := range doc.Statements {
		symbol := p.createStatementSymbol(stmt, lines)
		if symbol != nil {
			symbols = append(symbols, *symbol)
		}
	}

	return symbols
}

// createStatementSymbol creates a DocumentSymbol for a statement
func (p *Provider) createStatementSymbol(stmt *parser.Node, lines []string) *lsp.DocumentSymbol {
	if stmt == nil || stmt.Type != parser.NodeTypeStatement {
		return nil
	}

	// Get statement parameter (e.g., "LJS2012" from "++USERMOD(LJS2012)")
	stmtParam := ""
	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
			stmtParam = child.Value
			break
		}
	}

	// Build symbol name
	name := stmt.Name
	if stmtParam != "" {
		name = stmt.Name + "(" + stmtParam + ")"
	}

	// Determine symbol kind based on statement type
	kind := p.getSymbolKind(stmt.Name)

	// Calculate range (from statement start to terminator or last operand)
	endLine, endChar := p.getStatementEndPosition(stmt, lines)

	symbol := &lsp.DocumentSymbol{
		Name:   name,
		Detail: p.getStatementDetail(stmt),
		Kind:   kind,
		Range: lsp.Range{
			Start: lsp.Position{Line: stmt.Position.Line, Character: stmt.Position.Character},
			End:   lsp.Position{Line: endLine, Character: endChar},
		},
		SelectionRange: lsp.Range{
			Start: lsp.Position{Line: stmt.Position.Line, Character: stmt.Position.Character},
			End:   lsp.Position{Line: stmt.Position.Line, Character: stmt.Position.Character + stmt.Position.Length},
		},
	}

	// Add child symbols for key operands
	symbol.Children = p.getOperandSymbols(stmt)

	return symbol
}

// getSymbolKind returns the appropriate SymbolKind for a statement
func (p *Provider) getSymbolKind(stmtName string) lsp.SymbolKind {
	switch stmtName {
	case "++FUNCTION", "++USERMOD", "++PTF", "++APAR":
		// SYSMOD definitions - like classes/modules
		return lsp.SymbolKindClass
	case "++VER":
		// Verification - like a method/function
		return lsp.SymbolKindMethod
	case "++IF":
		// Conditional - like an operator
		return lsp.SymbolKindOperator
	case "++MAC", "++SRC", "++MOD":
		// Data elements - like structs/objects
		return lsp.SymbolKindStruct
	case "++MACUPD", "++SRCUPD":
		// Updates - like events
		return lsp.SymbolKindEvent
	case "++JCLIN":
		// JCL - like a file
		return lsp.SymbolKindFile
	default:
		// Default to function for other statements
		return lsp.SymbolKindFunction
	}
}

// getStatementDetail returns a detail string for the statement
func (p *Provider) getStatementDetail(stmt *parser.Node) string {
	if stmt.StatementDef != nil {
		return stmt.StatementDef.Description
	}
	return ""
}

// getStatementEndPosition finds the end position of a statement
func (p *Provider) getStatementEndPosition(stmt *parser.Node, lines []string) (int, int) {
	endLine := stmt.Position.Line
	endChar := stmt.Position.Character + stmt.Position.Length

	// Check children for the furthest position
	for _, child := range stmt.Children {
		if child.Position.Line > endLine {
			endLine = child.Position.Line
			endChar = child.Position.Character + child.Position.Length
		} else if child.Position.Line == endLine && child.Position.Character+child.Position.Length > endChar {
			endChar = child.Position.Character + child.Position.Length
		}

		// Check grandchildren
		for _, grandchild := range child.Children {
			if grandchild.Position.Line > endLine {
				endLine = grandchild.Position.Line
				endChar = grandchild.Position.Character + grandchild.Position.Length
			} else if grandchild.Position.Line == endLine && grandchild.Position.Character+grandchild.Position.Length > endChar {
				endChar = grandchild.Position.Character + grandchild.Position.Length
			}
		}
	}

	// Look for terminator
	for i := endLine; i < len(lines); i++ {
		line := lines[i]
		for j := 0; j < len(line); j++ {
			if line[j] == '.' {
				return i, j + 1
			}
		}
		// Stop if we hit another statement
		if i > stmt.Position.Line {
			for j := 0; j < len(line)-1; j++ {
				if line[j] == '+' && line[j+1] == '+' {
					return endLine, endChar
				}
			}
		}
	}

	return endLine, endChar
}

// getOperandSymbols returns child symbols for key operands
func (p *Provider) getOperandSymbols(stmt *parser.Node) []lsp.DocumentSymbol {
	var children []lsp.DocumentSymbol

	for _, child := range stmt.Children {
		if child.Type != parser.NodeTypeOperand {
			continue
		}

		// Only include operands with parameters as child symbols
		paramValue := ""
		for _, grandchild := range child.Children {
			if grandchild.Type == parser.NodeTypeParameter {
				paramValue = grandchild.Value
				break
			}
		}

		if paramValue == "" {
			continue
		}

		// Create symbol for this operand
		name := child.Name + "(" + paramValue + ")"

		childSymbol := lsp.DocumentSymbol{
			Name: name,
			Kind: lsp.SymbolKindProperty,
			Range: lsp.Range{
				Start: lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
				End:   lsp.Position{Line: child.Position.Line, Character: child.Position.Character + child.Position.Length + len(paramValue) + 2},
			},
			SelectionRange: lsp.Range{
				Start: lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
				End:   lsp.Position{Line: child.Position.Line, Character: child.Position.Character + child.Position.Length},
			},
		}

		children = append(children, childSymbol)
	}

	return children
}
