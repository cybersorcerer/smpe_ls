// Package codeactions provides LSP quick fixes for SMP/E diagnostics.
package codeactions

import (
	"fmt"
	"strings"
	"time"

	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider builds code actions from diagnostics.
type Provider struct{}

// NewProvider creates a code actions provider.
func NewProvider() *Provider {
	return &Provider{}
}

// GetCodeActions returns quick fixes for the diagnostics in the request context.
// It never returns nil (serializes as [] not null).
func (p *Provider) GetCodeActions(uri string, ctx lsp.CodeActionContext) []lsp.CodeAction {
	actions := []lsp.CodeAction{}
	for _, d := range ctx.Diagnostics {
		switch d.Code {
		case diagnostics.CodeMissingTerminator:
			actions = append(actions, p.terminatorFix(uri, d))
		case diagnostics.CodeEmptyOperandParameter:
			if a, ok := p.reworkFix(uri, d); ok {
				actions = append(actions, a)
			}
		}
	}
	return actions
}

// operandName reads the operand name from the diagnostic Data payload.
// Data round-trips through JSON as map[string]interface{}.
func operandName(d lsp.Diagnostic) string {
	m, ok := d.Data.(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := m["operand"].(string)
	return name
}

// reworkFix replaces an empty REWORK() operand with the current Julian date.
// Returns false when the empty operand is not REWORK.
func (p *Provider) reworkFix(uri string, d lsp.Diagnostic) (lsp.CodeAction, bool) {
	if !strings.EqualFold(operandName(d), "REWORK") {
		return lsp.CodeAction{}, false
	}
	now := time.Now()
	newText := fmt.Sprintf("REWORK(%d%03d)", now.Year(), now.YearDay())
	edit := lsp.TextEdit{Range: d.Range, NewText: newText}
	return lsp.CodeAction{
		Title:       "Set REWORK to current date",
		Kind:        "quickfix",
		Diagnostics: []lsp.Diagnostic{d},
		Edit:        &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
		IsPreferred: true,
	}, true
}

// terminatorFix inserts a '.' at the end of the statement range.
func (p *Provider) terminatorFix(uri string, d lsp.Diagnostic) lsp.CodeAction {
	pos := d.Range.End
	edit := lsp.TextEdit{
		Range:   lsp.Range{Start: pos, End: pos},
		NewText: ".",
	}
	return lsp.CodeAction{
		Title:       "Add statement terminator",
		Kind:        "quickfix",
		Diagnostics: []lsp.Diagnostic{d},
		Edit:        &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
		IsPreferred: true,
	}
}
