package codeactions

import (
	"regexp"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

func diag(code string, data any, line, char, length int) lsp.Diagnostic {
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: line, Character: char},
			End:   lsp.Position{Line: line, Character: char + length},
		},
		Code: code,
		Data: data,
	}
}

func TestMissingTerminatorFix(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{diag(diagnostics.CodeMissingTerminator, nil, 0, 0, 6)},
	}
	actions := p.GetCodeActions(uri, ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Add statement terminator" {
		t.Errorf("unexpected title: %q", actions[0].Title)
	}
	edits := actions[0].Edit.Changes[uri]
	if len(edits) != 1 || edits[0].NewText != "." {
		t.Errorf("expected single '.' edit, got %+v", edits)
	}
}

func TestReworkFix(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeEmptyOperandParameter, map[string]any{"operand": "REWORK"}, 1, 4, 6),
		},
	}
	actions := p.GetCodeActions(uri, ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Set REWORK to current date" {
		t.Errorf("unexpected title: %q", actions[0].Title)
	}
	edits := actions[0].Edit.Changes[uri]
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if !regexp.MustCompile(`^REWORK\(\d{7}\)$`).MatchString(edits[0].NewText) {
		t.Errorf("expected REWORK(yyyyddd), got %q", edits[0].NewText)
	}
}

func TestEmptyOperandNonReworkNoFix(t *testing.T) {
	p := NewProvider()
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeEmptyOperandParameter, map[string]any{"operand": "FMID"}, 1, 4, 4),
		},
	}
	if got := p.GetCodeActions("file:///t.smpe", ctx); len(got) != 0 {
		t.Errorf("expected no actions for non-REWORK empty operand, got %d", len(got))
	}
}

func TestSingleMissingOperandFix(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeMissingRequiredOperand, map[string]any{"operand": "SOURCEID"}, 0, 0, 8),
		},
	}
	actions := p.GetCodeActions(uri, ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Insert operand SOURCEID" {
		t.Errorf("unexpected title: %q", actions[0].Title)
	}
	if got := actions[0].Edit.Changes[uri][0].NewText; got != "\n    SOURCEID()" {
		t.Errorf("unexpected edit text: %q", got)
	}
}

func TestMissingOperandNilDataNoFix(t *testing.T) {
	p := NewProvider()
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeMissingRequiredOperand, nil, 0, 0, 8),
		},
	}
	if got := p.GetCodeActions("file:///t.smpe", ctx); len(got) != 0 {
		t.Errorf("expected no actions for missing-operand with nil Data, got %d", len(got))
	}
}

func TestMultipleMissingOperandsAddInsertAll(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeMissingRequiredOperand, map[string]any{"operand": "SOURCEID"}, 0, 0, 8),
			diag(diagnostics.CodeMissingRequiredOperand, map[string]any{"operand": "TO"}, 0, 0, 8),
		},
	}
	actions := p.GetCodeActions(uri, ctx)
	if len(actions) != 3 { // 2 single + 1 insert-all
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}
	var all *lsp.CodeAction
	for i := range actions {
		if actions[i].Title == "Insert all required operands" {
			all = &actions[i]
		}
	}
	if all == nil {
		t.Fatal("missing 'Insert all required operands' action")
	}
	if got := all.Edit.Changes[uri][0].NewText; got != "\n    SOURCEID()\n    TO()" {
		t.Errorf("unexpected insert-all text: %q", got)
	}
}
