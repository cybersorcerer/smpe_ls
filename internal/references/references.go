package references

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// SymbolType represents the type of a symbol
type SymbolType int

const (
	SymbolTypeSYSMOD   SymbolType = iota // PTF, APAR, USERMOD, FUNCTION
	SymbolTypeFMID                       // FUNCTION identifier
	SymbolTypeElement                    // MAC, SRC, MOD, etc.
)

// Symbol represents a symbol in the document
type Symbol struct {
	Name       string
	Type       SymbolType
	Position   lsp.Position
	Length     int
	IsDefinition bool
	Context    string // e.g., "PRE", "REQ", "SUP", "FMID", statement name
}

// Provider provides go-to-definition and find-references functionality
type Provider struct{}

// NewProvider creates a new references provider
func NewProvider() *Provider {
	return &Provider{}
}

// GetDefinition returns the definition location for a symbol at the given position
func (p *Provider) GetDefinition(doc *parser.Document, text string, line, character int) *lsp.Location {
	if doc == nil {
		return nil
	}

	// Find symbol at cursor position
	symbol := p.findSymbolAtPosition(doc, line, character)
	if symbol == nil || symbol.IsDefinition {
		return nil // Already at definition or no symbol found
	}

	// Find the definition of this symbol
	definition := p.findDefinition(doc, symbol.Name, symbol.Type)
	if definition == nil {
		return nil
	}

	return &lsp.Location{
		URI: "", // Will be set by handler
		Range: lsp.Range{
			Start: definition.Position,
			End: lsp.Position{
				Line:      definition.Position.Line,
				Character: definition.Position.Character + definition.Length,
			},
		},
	}
}

// GetReferences returns all references to a symbol at the given position
func (p *Provider) GetReferences(doc *parser.Document, text string, line, character int, includeDeclaration bool) []lsp.Location {
	if doc == nil {
		return nil
	}

	// Find symbol at cursor position
	symbol := p.findSymbolAtPosition(doc, line, character)
	if symbol == nil {
		return nil
	}

	// Find all references to this symbol
	symbols := p.findAllSymbols(doc)
	var locations []lsp.Location

	for _, s := range symbols {
		if s.Name == symbol.Name && s.Type == symbol.Type {
			// Skip definition if not requested
			if !includeDeclaration && s.IsDefinition {
				continue
			}

			locations = append(locations, lsp.Location{
				URI: "", // Will be set by handler
				Range: lsp.Range{
					Start: s.Position,
					End: lsp.Position{
						Line:      s.Position.Line,
						Character: s.Position.Character + s.Length,
					},
				},
			})
		}
	}

	return locations
}

// findSymbolAtPosition finds the symbol at the given position
func (p *Provider) findSymbolAtPosition(doc *parser.Document, line, character int) *Symbol {
	for _, stmt := range doc.Statements {
		// Check if cursor is on statement parameter (SYSMOD definition)
		if p.isSYSMODStatement(stmt.Name) {
			for _, child := range stmt.Children {
				if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
					if p.isPositionInRange(line, character, child.Position) {
						return &Symbol{
							Name:         child.Value,
							Type:         p.getSYSMODType(stmt.Name),
							Position:     lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
							Length:       len(child.Value),
							IsDefinition: true,
							Context:      stmt.Name,
						}
					}
				}
			}
		}

		// Check operands for references
		for _, child := range stmt.Children {
			if child.Type != parser.NodeTypeOperand {
				continue
			}

			// Check operands that reference SYSMODs: PRE, REQ, SUP, IF
			if p.isSYSMODReferenceOperand(child.Name) {
				for _, param := range child.Children {
					if param.Type == parser.NodeTypeParameter {
						// Handle comma-separated values
						refs := p.parseParameterReferences(param.Value)
						offset := 0
						for _, ref := range refs {
							refStart := param.Position.Character + offset
							refLen := len(ref)
							if p.isPositionInRangeWithLength(line, character, param.Position.Line, refStart, refLen) {
								return &Symbol{
									Name:         ref,
									Type:         SymbolTypeSYSMOD,
									Position:     lsp.Position{Line: param.Position.Line, Character: refStart},
									Length:       refLen,
									IsDefinition: false,
									Context:      child.Name,
								}
							}
							offset += refLen + 1 // +1 for comma
						}
					}
				}
			}

			// Check FMID operand (references ++FUNCTION)
			if child.Name == "FMID" {
				for _, param := range child.Children {
					if param.Type == parser.NodeTypeParameter {
						if p.isPositionInRange(line, character, param.Position) {
							return &Symbol{
								Name:         param.Value,
								Type:         SymbolTypeFMID,
								Position:     lsp.Position{Line: param.Position.Line, Character: param.Position.Character},
								Length:       len(param.Value),
								IsDefinition: false,
								Context:      "FMID",
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// findDefinition finds the definition of a symbol
func (p *Provider) findDefinition(doc *parser.Document, name string, symbolType SymbolType) *Symbol {
	for _, stmt := range doc.Statements {
		switch symbolType {
		case SymbolTypeSYSMOD:
			if p.isSYSMODStatement(stmt.Name) {
				for _, child := range stmt.Children {
					if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
						if child.Value == name {
							return &Symbol{
								Name:         child.Value,
								Type:         symbolType,
								Position:     lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
								Length:       len(child.Value),
								IsDefinition: true,
								Context:      stmt.Name,
							}
						}
					}
				}
			}
		case SymbolTypeFMID:
			if stmt.Name == "++FUNCTION" {
				for _, child := range stmt.Children {
					if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
						if child.Value == name {
							return &Symbol{
								Name:         child.Value,
								Type:         symbolType,
								Position:     lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
								Length:       len(child.Value),
								IsDefinition: true,
								Context:      stmt.Name,
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// findAllSymbols finds all symbols in the document
func (p *Provider) findAllSymbols(doc *parser.Document) []Symbol {
	var symbols []Symbol

	for _, stmt := range doc.Statements {
		// Collect SYSMOD definitions
		if p.isSYSMODStatement(stmt.Name) {
			for _, child := range stmt.Children {
				if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
					symbols = append(symbols, Symbol{
						Name:         child.Value,
						Type:         p.getSYSMODType(stmt.Name),
						Position:     lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
						Length:       len(child.Value),
						IsDefinition: true,
						Context:      stmt.Name,
					})
				}
			}
		}

		// Collect references from operands
		for _, child := range stmt.Children {
			if child.Type != parser.NodeTypeOperand {
				continue
			}

			// SYSMOD references
			if p.isSYSMODReferenceOperand(child.Name) {
				for _, param := range child.Children {
					if param.Type == parser.NodeTypeParameter {
						refs := p.parseParameterReferences(param.Value)
						offset := 0
						for _, ref := range refs {
							symbols = append(symbols, Symbol{
								Name:         ref,
								Type:         SymbolTypeSYSMOD,
								Position:     lsp.Position{Line: param.Position.Line, Character: param.Position.Character + offset},
								Length:       len(ref),
								IsDefinition: false,
								Context:      child.Name,
							})
							offset += len(ref) + 1
						}
					}
				}
			}

			// FMID references
			if child.Name == "FMID" {
				for _, param := range child.Children {
					if param.Type == parser.NodeTypeParameter {
						symbols = append(symbols, Symbol{
							Name:         param.Value,
							Type:         SymbolTypeFMID,
							Position:     lsp.Position{Line: param.Position.Line, Character: param.Position.Character},
							Length:       len(param.Value),
							IsDefinition: false,
							Context:      "FMID",
						})
					}
				}
			}
		}
	}

	return symbols
}

// isSYSMODStatement checks if the statement defines a SYSMOD
func (p *Provider) isSYSMODStatement(name string) bool {
	switch name {
	case "++PTF", "++APAR", "++USERMOD", "++FUNCTION":
		return true
	}
	return false
}

// getSYSMODType returns the symbol type for a SYSMOD statement
func (p *Provider) getSYSMODType(stmtName string) SymbolType {
	if stmtName == "++FUNCTION" {
		return SymbolTypeFMID
	}
	return SymbolTypeSYSMOD
}

// isSYSMODReferenceOperand checks if the operand references a SYSMOD
func (p *Provider) isSYSMODReferenceOperand(name string) bool {
	switch name {
	case "PRE", "REQ", "SUP", "IF":
		return true
	}
	return false
}

// parseParameterReferences parses comma-separated references from a parameter value
func (p *Provider) parseParameterReferences(value string) []string {
	// Handle both simple values and comma-separated lists
	// e.g., "UJ12345" or "UJ12345,UJ12346"
	parts := strings.Split(value, ",")
	var refs []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			refs = append(refs, trimmed)
		}
	}
	return refs
}

// isPositionInRange checks if a position is within a node's range
func (p *Provider) isPositionInRange(line, character int, pos parser.Position) bool {
	if line != pos.Line {
		return false
	}
	return character >= pos.Character && character < pos.Character+pos.Length
}

// isPositionInRangeWithLength checks if a position is within a specific range
func (p *Provider) isPositionInRangeWithLength(line, character, targetLine, targetChar, length int) bool {
	if line != targetLine {
		return false
	}
	return character >= targetChar && character < targetChar+length
}
