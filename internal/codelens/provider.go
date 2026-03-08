package codelens

import (
	"fmt"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides CodeLens functionality for SYSMOD and DDDEF z/OSMF queries
type Provider struct{}

// NewProvider creates a new CodeLens provider
func NewProvider() *Provider {
	return &Provider{}
}

// GetCodeLenses returns CodeLens items for the given document
func (p *Provider) GetCodeLenses(doc *parser.Document) []lsp.CodeLens {
	if doc == nil {
		return nil
	}

	var lenses []lsp.CodeLens

	for _, stmt := range doc.Statements {
		// 1. SYSMOD definitions: ++PTF(UA12345), ++APAR(...), ++USERMOD(...), ++FUNCTION(...)
		if isSYSMODStatement(stmt.Name) {
			for _, child := range stmt.Children {
				if child.Type == parser.NodeTypeParameter && child.Parent == stmt && child.Value != "" {
					lenses = append(lenses, makeSysmodLens(child))
				}
			}
		}

		// 2. Operand references
		for _, child := range stmt.Children {
			if child.Type != parser.NodeTypeOperand {
				continue
			}

			// SYSMOD references — one CodeLens per operand covering all SYSMODs in the list
			if isSYSMODReferenceOperand(child.Name) {
				var allRefs []string
				var firstParam *parser.Node
				for _, param := range child.Children {
					if param.Type == parser.NodeTypeParameter {
						if firstParam == nil {
							firstParam = param
						}
						// Use children (individual items) if available, else split Value
						if len(param.Children) > 0 {
							for _, item := range param.Children {
								if item.Type == parser.NodeTypeParameter && item.Value != "" {
									allRefs = append(allRefs, item.Value)
								}
							}
						} else {
							allRefs = append(allRefs, parseList(param.Value)...)
						}
					}
				}
				if len(allRefs) > 0 && firstParam != nil {
					lenses = append(lenses, makeSysmodListLens(child, firstParam, allRefs))
				}
			}

			// DDDEF references: DISTLIB, SYSLIB, TXLIB, RELFILE, FROMDS
			if isDDDEFReferenceOperand(child.Name) {
				for _, param := range child.Children {
					if param.Type == parser.NodeTypeParameter && param.Value != "" {
						lenses = append(lenses, makeDddefLens(param))
					}
				}
			}
		}
	}

	return lenses
}

// isSYSMODStatement checks if the statement defines a SYSMOD
func isSYSMODStatement(name string) bool {
	switch name {
	case "++PTF", "++APAR", "++USERMOD", "++FUNCTION":
		return true
	}
	return false
}

// isSYSMODReferenceOperand checks if the operand references one or more SYSMODs
func isSYSMODReferenceOperand(name string) bool {
	switch name {
	case "DELETE", "FMID", "NPRE", "PRE", "REQ", "RESOLVER", "RMID", "SUP", "TO", "UMID", "VERSION":
		return true
	}
	return false
}

// isDDDEFReferenceOperand checks if the operand references a DDDEF
func isDDDEFReferenceOperand(name string) bool {
	switch name {
	case "DISTLIB", "SYSLIB", "TXLIB", "RELFILE", "FROMDS":
		return true
	}
	return false
}

// parseList splits a list value into trimmed parts.
// In SMP/E, list items can be separated by commas or spaces (both are valid).
func parseList(value string) []string {
	var refs []string
	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	}) {
		if part != "" {
			refs = append(refs, part)
		}
	}
	return refs
}

// makeSysmodLens creates a CodeLens for a SYSMOD definition
func makeSysmodLens(node *parser.Node) lsp.CodeLens {
	return lsp.CodeLens{
		Range: lsp.Range{
			Start: lsp.Position{Line: node.Position.Line, Character: node.Position.Character},
			End:   lsp.Position{Line: node.Position.Line, Character: node.Position.Character + len(node.Value)},
		},
		Command: &lsp.Command{
			Title:     fmt.Sprintf("🔍 Query SYSMOD %s", node.Value),
			Command:   "smpe.codelens.querySysmod",
			Arguments: []interface{}{node.Value},
		},
	}
}

// makeSysmodListLens creates a single CodeLens for all SYSMODs in a list operand.
// The SYSMOD IDs are passed as a string array; the filter is built by the extension.
func makeSysmodListLens(operand *parser.Node, firstParam *parser.Node, refs []string) lsp.CodeLens {
	title := fmt.Sprintf("🔍 Query %s (%d SYSMODs)", operand.Name, len(refs))
	if len(refs) == 1 {
		title = fmt.Sprintf("🔍 Query %s %s", operand.Name, refs[0])
	}

	// Pass each SYSMOD as a separate argument so VSCode spreads them correctly
	args := make([]interface{}, len(refs))
	for i, ref := range refs {
		args[i] = ref
	}

	return lsp.CodeLens{
		Range: lsp.Range{
			Start: lsp.Position{Line: firstParam.Position.Line, Character: firstParam.Position.Character},
			End:   lsp.Position{Line: firstParam.Position.Line, Character: firstParam.Position.Character + len(firstParam.Value)},
		},
		Command: &lsp.Command{
			Title:     title,
			Command:   "smpe.codelens.querySysmod",
			Arguments: args,
		},
	}
}

// makeDddefLens creates a CodeLens for a DDDEF reference
func makeDddefLens(param *parser.Node) lsp.CodeLens {
	return lsp.CodeLens{
		Range: lsp.Range{
			Start: lsp.Position{Line: param.Position.Line, Character: param.Position.Character},
			End:   lsp.Position{Line: param.Position.Line, Character: param.Position.Character + len(param.Value)},
		},
		Command: &lsp.Command{
			Title:     fmt.Sprintf("🔍 Query DDDEF %s", param.Value),
			Command:   "smpe.codelens.queryDddef",
			Arguments: []interface{}{param.Value},
		},
	}
}
