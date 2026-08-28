package formatting

import (
	"strings"
	"testing"
)

// Tests for the inline data boundary: comment scanning must never reach into
// the data lines that follow a statement expecting inline data.

// An unterminated comment after the terminator must not make the formatter
// reach into the following inline data lines.
func TestUnclosedTrailingCommentDoesNotConsumeInlineData(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++SRC(MYEXEC) DISTLIB(AOSB3) . /* unclosed trailing\n" +
		"/* REXX */\n" +
		"say 'hello'\n" +
		"++VER(Z038) FMID(FXY1040) .\n"

	doc := p.Parse(input)
	edits := fp.FormatDocument(doc, input)
	if len(edits) == 0 {
		t.Fatal("Expected at least one edit")
	}
	t.Logf("edit[0]: %d:%d -> %d:%d\n%s", edits[0].Range.Start.Line, edits[0].Range.Start.Character,
		edits[0].Range.End.Line, edits[0].Range.End.Character, edits[0].NewText)

	if edits[0].Range.End.Line != 0 {
		t.Errorf("Edit range reaches into inline data: ends at line %d, want 0", edits[0].Range.End.Line)
	}
	if strings.Contains(edits[0].NewText, "REXX") {
		t.Errorf("Inline data pulled into statement text:\n%s", edits[0].NewText)
	}
}

// getStatementEndLine must stop at the terminator line for statements that
// expect inline data, even when a comment after the dot is left unclosed.
func TestGetStatementEndLine_UnclosedCommentWithInlineData(t *testing.T) {
	p, fp := newTestFormatter(t)

	input := "++SRC(MYEXEC) DISTLIB(AOSB3) . /* unclosed trailing\n" +
		"/* REXX */\n" +
		"say 'hello'\n"
	doc := p.Parse(input)
	if len(doc.Statements) == 0 {
		t.Fatal("Expected a parsed statement")
	}

	lines := strings.Split(input, "\n")
	if got := fp.getStatementEndLine(doc.Statements[0], lines); got != 0 {
		t.Errorf("Expected endLine=0, got %d (scan ran into inline data)", got)
	}
}
