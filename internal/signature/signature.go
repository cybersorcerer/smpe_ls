// Package signature provides LSP signature help for SMP/E operand parameters.
package signature

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider builds signature help from the AST and smpe.json operand metadata.
type Provider struct {
	statements map[string]data.MCSStatement
	enabled    bool
}

// NewProvider creates a signature help provider. Enabled by default; the client
// can turn it off via the smpe.signatureHelp.enabled setting.
func NewProvider(store *data.Store) *Provider {
	return &Provider{statements: store.Statements, enabled: true}
}

// SetEnabled toggles the feature (driven by the client setting).
func (p *Provider) SetEnabled(enabled bool) { p.enabled = enabled }

// GetSignatureHelp returns help for the operand parameter at the cursor, or nil
// when disabled, when the cursor is not inside an operand's parentheses, or when
// the operand takes no parameter (boolean flag).
func (p *Provider) GetSignatureHelp(doc *parser.Document, text string, line, character int) *lsp.SignatureHelp {
	if !p.enabled || doc == nil {
		return nil
	}
	node := p.findOperandAtCursor(doc, text, line, character)
	if node == nil || node.Type != parser.NodeTypeOperand || node.OperandDef == nil {
		return nil
	}
	op := *node.OperandDef
	if op.Parameter == "" || op.Type == "boolean" {
		return nil
	}
	return buildOperandSignature(op)
}

// buildOperandSignature renders the box content: "NAME(PARAM)" plus the first
// sentence of the description and the type, kept short so the floating box stays
// readable while typing.
func buildOperandSignature(op data.Operand) *lsp.SignatureHelp {
	label := op.Name + "(" + op.Parameter + ")"
	doc := strings.TrimSpace(firstSentence(op.Description))
	if op.Type != "" {
		if doc != "" {
			doc += " "
		}
		doc += "(" + op.Type + ")"
	}
	return &lsp.SignatureHelp{
		Signatures:      []lsp.SignatureInformation{{Label: label, Documentation: doc}},
		ActiveSignature: 0,
		ActiveParameter: 0,
	}
}

// firstSentence returns text up to and including the first ". " boundary, or the
// whole string if there is no such boundary.
func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i >= 0 {
		return s[:i+1]
	}
	return s
}

// findOperandAtCursor finds the operand whose parentheses the cursor sits
// strictly inside. The parser's operand Position.Length spans only the operand
// name, not its parentheses, so a name-length heuristic cannot tell "DISTLIB(|)"
// (inside) from "DISTLIB() |" (after the closing paren). We therefore inspect
// the actual line text:
//  1. Scan the line from column 0 up to the cursor, tracking paren depth.
//     Depth <= 0 at the cursor means the cursor is not inside an open group.
//  2. The most recently opened, still-open '(' identifies the group the cursor
//     is in. Read the identifier (letters) immediately preceding that '(' — that
//     is the owning operand's name, starting at nameStart.
//  3. Return the operand child node at (line, nameStart).
func (p *Provider) findOperandAtCursor(doc *parser.Document, text string, line, character int) *parser.Node {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}
	ln := lines[line]
	if character > len(ln) {
		character = len(ln)
	}

	// Track paren depth and the columns of currently-open '(' groups.
	openStack := make([]int, 0, 4)
	for i := 0; i < character; i++ {
		switch ln[i] {
		case '(':
			openStack = append(openStack, i)
		case ')':
			if len(openStack) > 0 {
				openStack = openStack[:len(openStack)-1]
			}
		}
	}
	if len(openStack) == 0 {
		return nil
	}
	openCol := openStack[len(openStack)-1]

	// Read the identifier (letters) immediately before the opening paren.
	end := openCol
	start := end
	for start > 0 && isIdentLetter(ln[start-1]) {
		start--
	}
	if start == end {
		return nil
	}

	for _, stmt := range doc.Statements {
		for _, child := range stmt.Children {
			if child.Type != parser.NodeTypeOperand || child.OperandDef == nil {
				continue
			}
			if child.Position.Line == line && child.Position.Character == start {
				return child
			}
		}
	}
	return nil
}

// isIdentLetter reports whether b is part of an operand identifier (letters only,
// matching the SMP/E operand keyword grammar).
func isIdentLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
