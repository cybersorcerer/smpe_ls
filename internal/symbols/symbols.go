package symbols

import (
	"strings"

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
	kind := p.GetSymbolKind(stmt.Name)

	// Calculate range (from statement start to terminator or last operand)
	endLine, endChar := p.GetStatementEndPosition(stmt, lines)

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

// GetSymbolKind returns the appropriate SymbolKind for a statement
func (p *Provider) GetSymbolKind(stmtName string) lsp.SymbolKind {
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

// GetStatementEndPosition finds the end position of a statement: its '.'
// terminator. Only a '.' at parenthesis depth 0, outside /* ... */ block
// comments and single-quoted strings counts. A dot inside an operand value -
// a free-text DESC, a dotted dataset name, a "dd.mm.yy" date in a comment -
// is not the terminator and must not cut the range short.
func (p *Provider) GetStatementEndPosition(stmt *parser.Node, lines []string) (int, int) {
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

	// Look for the terminator. The scan starts at the statement's own line so
	// parenthesis and string state are complete: an operand may open its
	// parenthesis on an earlier line than the one carrying a dot.
	inBlockComment := false
	inQuote := false
	parenDepth := 0
	// With unbalanced parentheses the depth is meaningless - the statement is
	// already reported as malformed by its own diagnostic - so the terminator
	// is accepted at any depth rather than never being found at all.
	ignoreDepth := stmt.UnbalancedParens != 0
	for i := stmt.Position.Line; i < len(lines); i++ {
		line := lines[i]

		// A following statement ends the search. Only a "++" at the start of
		// the line counts: "++APAR" inside a comment is text, not a statement.
		// Checked before scanning the line so the next statement's terminator
		// is never taken for this one's. inBlockComment still holds the state
		// at the start of this line, and a line inside an open block comment
		// cannot begin a statement.
		if i > stmt.Position.Line && !inBlockComment && strings.HasPrefix(strings.TrimSpace(line), "++") {
			return endLine, endChar
		}

		for j := 0; j < len(line); j++ {
			if inBlockComment {
				if j+1 < len(line) && line[j] == '*' && line[j+1] == '/' {
					inBlockComment = false
					j++
				}
				continue
			}
			if inQuote {
				if line[j] == '\'' {
					inQuote = false
				}
				continue
			}
			if j+1 < len(line) && line[j] == '/' && line[j+1] == '*' {
				inBlockComment = true
				j++
				continue
			}
			switch line[j] {
			case '\'':
				// Only track strings outside parentheses; inside them the
				// depth already keeps dots from counting.
				if parenDepth == 0 {
					inQuote = true
				}
				continue
			case '(':
				parenDepth++
				continue
			case ')':
				if parenDepth > 0 {
					parenDepth--
				}
				continue
			}
			if line[j] == '.' && (parenDepth == 0 || ignoreDepth) {
				return i, j + 1
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
