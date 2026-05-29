package codeactions

import (
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
