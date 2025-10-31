package completion

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides code completion
type Provider struct {
	// TODO: Load from smpe.json
}

// NewProvider creates a new completion provider
func NewProvider() *Provider {
	return &Provider{}
}

// GetCompletions returns completion items for the given position
func (p *Provider) GetCompletions(text string, line, character int) []lsp.CompletionItem {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return nil
	}

	// Get the text before cursor
	textBefore := currentLine[:character]

	// Check if we're at the start of a line (MCS statement)
	if strings.TrimSpace(textBefore) == "" || strings.HasPrefix(strings.TrimSpace(textBefore), "+") {
		return p.getMCSCompletions()
	}

	// Check if we're after a MCS statement (operands)
	if strings.Contains(textBefore, "++") {
		return p.getOperandCompletions(textBefore)
	}

	return nil
}

// getMCSCompletions returns MCS statement completions
func (p *Provider) getMCSCompletions() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{
			Label:         "++APAR",
			Kind:          lsp.CompletionItemKindKeyword,
			Detail:        "Service SYSMOD",
			Documentation: "Identifies a temporary fix (APAR)",
			InsertText:    "++APAR($1)",
		},
		{
			Label:         "++ASSIGN",
			Kind:          lsp.CompletionItemKindKeyword,
			Detail:        "Source ID Assignment",
			Documentation: "Assigns source IDs to SYSMODs",
			InsertText:    "++ASSIGN SOURCEID($1) TO($2)",
		},
		{
			Label:         "++DELETE",
			Kind:          lsp.CompletionItemKindKeyword,
			Detail:        "Delete Load Module",
			Documentation: "Deletes load modules from target libraries",
			InsertText:    "++DELETE($1) SYSLIB($2)",
		},
		{
			Label:         "++FEATURE",
			Kind:          lsp.CompletionItemKindKeyword,
			Detail:        "SYSMOD Set Description",
			Documentation: "Describes a set of SYSMODs as a feature",
			InsertText:    "++FEATURE($1) DESCRIPTION($2)",
		},
		{
			Label:         "++FUNCTION",
			Kind:          lsp.CompletionItemKindKeyword,
			Detail:        "Function SYSMOD",
			Documentation: "Identifies a base or dependent function SYSMOD",
			InsertText:    "++FUNCTION($1)",
		},
		{
			Label:         "++HOLD",
			Kind:          lsp.CompletionItemKindKeyword,
			Detail:        "Exception Status",
			Documentation: "Places a SYSMOD in exception status",
			InsertText:    "++HOLD($1) FMID($2) REASON($3)",
		},
	}
}

// getOperandCompletions returns operand completions based on context
func (p *Provider) getOperandCompletions(textBefore string) []lsp.CompletionItem {
	// Determine which MCS statement we're in
	if strings.Contains(textBefore, "++APAR") {
		return p.getAparOperands()
	} else if strings.Contains(textBefore, "++ASSIGN") {
		return p.getAssignOperands()
	} else if strings.Contains(textBefore, "++DELETE") {
		return p.getDeleteOperands()
	} else if strings.Contains(textBefore, "++FEATURE") {
		return p.getFeatureOperands()
	} else if strings.Contains(textBefore, "++FUNCTION") {
		return p.getFunctionOperands()
	} else if strings.Contains(textBefore, "++HOLD") {
		return p.getHoldOperands()
	}

	return nil
}

func (p *Provider) getAparOperands() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "DESCRIPTION", Kind: lsp.CompletionItemKindProperty, InsertText: "DESCRIPTION($1)"},
		{Label: "FILES", Kind: lsp.CompletionItemKindProperty, InsertText: "FILES($1)"},
		{Label: "RFDSNPFX", Kind: lsp.CompletionItemKindProperty, InsertText: "RFDSNPFX($1)"},
		{Label: "REWORK", Kind: lsp.CompletionItemKindProperty, InsertText: "REWORK($1)"},
	}
}

func (p *Provider) getAssignOperands() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "SOURCEID", Kind: lsp.CompletionItemKindProperty, InsertText: "SOURCEID($1)"},
		{Label: "TO", Kind: lsp.CompletionItemKindProperty, InsertText: "TO($1)"},
	}
}

func (p *Provider) getDeleteOperands() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "SYSLIB", Kind: lsp.CompletionItemKindProperty, InsertText: "SYSLIB($1)"},
		{Label: "ALIAS", Kind: lsp.CompletionItemKindProperty, InsertText: "ALIAS($1)"},
	}
}

func (p *Provider) getFeatureOperands() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "DESCRIPTION", Kind: lsp.CompletionItemKindProperty, InsertText: "DESCRIPTION($1)"},
		{Label: "FMID", Kind: lsp.CompletionItemKindProperty, InsertText: "FMID($1)"},
		{Label: "PRODUCT", Kind: lsp.CompletionItemKindProperty, InsertText: "PRODUCT($1)"},
		{Label: "REWORK", Kind: lsp.CompletionItemKindProperty, InsertText: "REWORK($1)"},
	}
}

func (p *Provider) getFunctionOperands() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "DESCRIPTION", Kind: lsp.CompletionItemKindProperty, InsertText: "DESCRIPTION($1)"},
		{Label: "FESN", Kind: lsp.CompletionItemKindProperty, InsertText: "FESN($1)"},
		{Label: "FILES", Kind: lsp.CompletionItemKindProperty, InsertText: "FILES($1)"},
		{Label: "RFDSNPFX", Kind: lsp.CompletionItemKindProperty, InsertText: "RFDSNPFX($1)"},
		{Label: "REWORK", Kind: lsp.CompletionItemKindProperty, InsertText: "REWORK($1)"},
	}
}

func (p *Provider) getHoldOperands() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "FMID", Kind: lsp.CompletionItemKindProperty, InsertText: "FMID($1)"},
		{Label: "REASON", Kind: lsp.CompletionItemKindProperty, InsertText: "REASON($1)"},
		{Label: "ERROR", Kind: lsp.CompletionItemKindProperty, InsertText: "ERROR"},
		{Label: "FIXCAT", Kind: lsp.CompletionItemKindProperty, InsertText: "FIXCAT"},
		{Label: "SYSTEM", Kind: lsp.CompletionItemKindProperty, InsertText: "SYSTEM"},
		{Label: "USER", Kind: lsp.CompletionItemKindProperty, InsertText: "USER"},
		{Label: "CATEGORY", Kind: lsp.CompletionItemKindProperty, InsertText: "CATEGORY($1)"},
		{Label: "RESOLVER", Kind: lsp.CompletionItemKindProperty, InsertText: "RESOLVER($1)"},
		{Label: "CLASS", Kind: lsp.CompletionItemKindProperty, InsertText: "CLASS($1)"},
		{Label: "DATE", Kind: lsp.CompletionItemKindProperty, InsertText: "DATE($1)"},
		{Label: "COMMENT", Kind: lsp.CompletionItemKindProperty, InsertText: "COMMENT($1)"},
	}
}
