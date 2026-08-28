package codeactions

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// column1Diag builds a comment-in-column-1 diagnostic for the given block range.
func column1Diag(line, blockStart, blockEnd int) lsp.Diagnostic {
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: line, Character: 0},
			End:   lsp.Position{Line: line, Character: 2},
		},
		Severity: lsp.SeverityError,
		Code:     diagnostics.CodeCommentInColumn1,
		Message:  "Comment must not begin in column 1",
		Data: map[string]interface{}{
			"blockStart": blockStart,
			"blockEnd":   blockEnd,
		},
	}
}

// applyEdits applies pure-insertion edits to text (bottom-up so positions hold).
func applyEdits(text string, edits []lsp.TextEdit) string {
	lines := strings.Split(text, "\n")
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		l := e.Range.Start.Line
		if l >= len(lines) {
			continue
		}
		c := e.Range.Start.Character
		lines[l] = lines[l][:c] + e.NewText + lines[l][c:]
	}
	return strings.Join(lines, "\n")
}

func TestCommentColumn1SingleLineFix(t *testing.T) {
	text := "++VER(Z038)\n/* comment in column 1 */\n    FMID(FXY1040)\n.\n"
	p := NewProvider()
	d := column1Diag(1, 1, 1)

	actions := p.GetCodeActions("file:///t.smpe", text, nil, lsp.Range{},
		lsp.CodeActionContext{Diagnostics: []lsp.Diagnostic{d}})

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action for a single-line comment, got %d", len(actions))
	}
	if !strings.Contains(actions[0].Title, "this comment line") {
		t.Errorf("Unexpected title: %q", actions[0].Title)
	}

	got := applyEdits(text, actions[0].Edit.Changes["file:///t.smpe"])
	want := "++VER(Z038)\n  /* comment in column 1 */\n    FMID(FXY1040)\n.\n"
	if got != want {
		t.Errorf("Fix produced:\n%q\nwant:\n%q", got, want)
	}
}

func TestCommentColumn1OffersBothFixesForBlock(t *testing.T) {
	text := "++VER(Z038)\n/* block opens\n   middle line\n   closes */\n    FMID(FXY1040)\n.\n"
	p := NewProvider()
	d := column1Diag(1, 1, 3)

	actions := p.GetCodeActions("file:///t.smpe", text, nil, lsp.Range{},
		lsp.CodeActionContext{Diagnostics: []lsp.Diagnostic{d}})

	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions for a multi-line comment, got %d", len(actions))
	}
	if !strings.Contains(actions[0].Title, "this comment line") {
		t.Errorf("Expected line fix first, got %q", actions[0].Title)
	}
	if !strings.Contains(actions[1].Title, "whole comment block") {
		t.Errorf("Expected block fix second, got %q", actions[1].Title)
	}

	// Line fix touches only the reported line
	gotLine := applyEdits(text, actions[0].Edit.Changes["file:///t.smpe"])
	wantLine := "++VER(Z038)\n  /* block opens\n   middle line\n   closes */\n    FMID(FXY1040)\n.\n"
	if gotLine != wantLine {
		t.Errorf("Line fix produced:\n%q\nwant:\n%q", gotLine, wantLine)
	}

	// Block fix shifts every line of the comment by the same amount
	gotBlock := applyEdits(text, actions[1].Edit.Changes["file:///t.smpe"])
	wantBlock := "++VER(Z038)\n  /* block opens\n     middle line\n     closes */\n    FMID(FXY1040)\n.\n"
	if gotBlock != wantBlock {
		t.Errorf("Block fix produced:\n%q\nwant:\n%q", gotBlock, wantBlock)
	}
}

// A blank line inside the block must not gain trailing whitespace.
func TestCommentColumn1BlockFixSkipsBlankLines(t *testing.T) {
	text := "++VER(Z038)\n/* block opens\n\n   closes */\n.\n"
	p := NewProvider()
	d := column1Diag(1, 1, 3)

	actions := p.GetCodeActions("file:///t.smpe", text, nil, lsp.Range{},
		lsp.CodeActionContext{Diagnostics: []lsp.Diagnostic{d}})
	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(actions))
	}

	got := applyEdits(text, actions[1].Edit.Changes["file:///t.smpe"])
	want := "++VER(Z038)\n  /* block opens\n\n     closes */\n.\n"
	if got != want {
		t.Errorf("Block fix produced:\n%q\nwant:\n%q", got, want)
	}
}
