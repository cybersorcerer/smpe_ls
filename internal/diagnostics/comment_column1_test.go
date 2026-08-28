package diagnostics

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// commentColumn1Diags runs only the comment-in-column-1 check over text.
func commentColumn1Diags(t *testing.T, text string) []lsp.Diagnostic {
	t.Helper()
	_, p, dp := loadRealStore(t)
	doc := p.Parse(text)
	return dp.AnalyzeASTWithConfigAndText(doc, &Config{CommentInColumn1: true}, text)
}

func TestCommentInColumn1IsReported(t *testing.T) {
	text := "++VER(Z038)\n" +
		"/* comment starting in column 1 */\n" +
		"    FMID(FXY1040)\n" +
		".\n"
	diags := commentColumn1Diags(t, text)

	if len(diags) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d: %+v", len(diags), diags)
	}
	d := diags[0]
	if d.Severity != lsp.SeverityError {
		t.Errorf("Expected error severity, got %d", d.Severity)
	}
	if d.Code != CodeCommentInColumn1 {
		t.Errorf("Expected code %q, got %q", CodeCommentInColumn1, d.Code)
	}
	if d.Range.Start.Line != 1 || d.Range.Start.Character != 0 {
		t.Errorf("Expected range at 1:0, got %d:%d", d.Range.Start.Line, d.Range.Start.Character)
	}
}

func TestIndentedCommentIsNotReported(t *testing.T) {
	text := "++VER(Z038)\n" +
		"  /* properly indented comment */\n" +
		"    FMID(FXY1040)\n" +
		".\n"
	if diags := commentColumn1Diags(t, text); len(diags) != 0 {
		t.Errorf("Expected no diagnostics, got %+v", diags)
	}
}

// Every line opening with "/*" in column 1 is reported, including a
// continuation line inside a block comment - the reader sees the end-of-data-set
// marker on any line.
func TestCommentInColumn1ReportsEveryLine(t *testing.T) {
	text := "++VER(Z038)\n" +
		"/* block opens here\n" +
		"   payload line\n" +
		"/* another one in column 1\n" +
		"   closes */\n" +
		"    FMID(FXY1040)\n" +
		".\n"
	diags := commentColumn1Diags(t, text)
	if len(diags) != 2 {
		t.Fatalf("Expected 2 diagnostics, got %d: %+v", len(diags), diags)
	}
	if diags[0].Range.Start.Line != 1 || diags[1].Range.Start.Line != 3 {
		t.Errorf("Expected lines 1 and 3, got %d and %d",
			diags[0].Range.Start.Line, diags[1].Range.Start.Line)
	}
}

// The block range travels in Data so the quick fix can shift the whole comment.
func TestCommentInColumn1CarriesBlockRange(t *testing.T) {
	text := "++VER(Z038)\n" +
		"/* block opens\n" +
		"   middle\n" +
		"   closes */\n" +
		"    FMID(FXY1040)\n" +
		".\n"
	diags := commentColumn1Diags(t, text)
	if len(diags) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d", len(diags))
	}
	m, ok := diags[0].Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected Data payload, got %T", diags[0].Data)
	}
	if m["blockStart"] != 1 || m["blockEnd"] != 3 {
		t.Errorf("Expected block 1..3, got %v..%v", m["blockStart"], m["blockEnd"])
	}
}

// Inline data is checked too: a "/*" in column 1 there truncates the input the
// same way, which is exactly the REXX-source trap.
func TestCommentInColumn1CoversInlineData(t *testing.T) {
	text := "++SRC(MYEXEC) DISTLIB(AOSB3) .\n" +
		"/* REXX */\n" +
		"say 'hello'\n"
	diags := commentColumn1Diags(t, text)
	if len(diags) != 1 {
		t.Fatalf("Expected 1 diagnostic in inline data, got %d: %+v", len(diags), diags)
	}
	if diags[0].Range.Start.Line != 1 {
		t.Errorf("Expected line 1, got %d", diags[0].Range.Start.Line)
	}
}

func TestCommentInColumn1CanBeDisabled(t *testing.T) {
	text := "++VER(Z038)\n" +
		"/* comment starting in column 1 */\n" +
		"    FMID(FXY1040)\n" +
		".\n"
	_, p, dp := loadRealStore(t)
	doc := p.Parse(text)
	diags := dp.AnalyzeASTWithConfigAndText(doc, &Config{CommentInColumn1: false}, text)
	if len(diags) != 0 {
		t.Errorf("Expected no diagnostics when disabled, got %+v", diags)
	}
}
