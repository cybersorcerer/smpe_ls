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
func (p *Provider) GetSignatureHelp(doc *parser.Document, line, character int) *lsp.SignatureHelp {
	if !p.enabled || doc == nil {
		return nil
	}
	node := p.findOperandAtCursor(doc, line, character)
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

// findOperandAtCursor finds the operand the cursor sits in. The parser's
// operand Position.Length spans only the operand name (e.g. "DISTLIB", len 7),
// not its parentheses, so the cursor inside "DISTLIB(|)" lands just past the
// node span. We therefore select, per statement, the operand child on the
// cursor's line whose name starts at or before the cursor with the greatest
// start column — that is the operand the cursor is typing inside.
func (p *Provider) findOperandAtCursor(doc *parser.Document, line, character int) *parser.Node {
	var best *parser.Node
	for _, stmt := range doc.Statements {
		for _, child := range stmt.Children {
			if child.Type != parser.NodeTypeOperand || child.OperandDef == nil {
				continue
			}
			if cursorInsideOperandParens(child, line, character) {
				if best == nil || child.Position.Character > best.Position.Character {
					best = child
				}
			}
		}
	}
	return best
}

// cursorInsideOperandParens reports whether (line, character) is at or after the
// operand name's start on the operand's line. The cursor must be strictly after
// the name start so positions left of the operand (e.g. statement keyword) do
// not match.
func cursorInsideOperandParens(op *parser.Node, line, character int) bool {
	if line != op.Position.Line {
		return false
	}
	return character >= op.Position.Character+op.Position.Length
}
