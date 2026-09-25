package formatting

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Typing "/* " closes the comment on the spot, so nobody has to remember the
// "*/". The marker has to fit within column 72 - SMP/E ignores anything beyond
// it, so a "*/" out there would look like it closes the comment without doing
// so.

// typeSpace applies what the server answers for a space typed at the end of
// line 0 of text, and returns the resulting line.
func typeSpace(t *testing.T, text string) string {
	t.Helper()
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	doc := parser.NewParser(store.Statements).Parse(text)

	lines := strings.Split(text, "\n")
	pos := lsp.Position{Line: 0, Character: len([]rune(lines[0]))}

	fp := NewProvider()
	edits := fp.CloseComment(doc, text, pos, " ")
	if len(edits) == 0 {
		return lines[0]
	}
	runes := []rune(lines[0])
	e := edits[0]
	return string(runes[:e.Range.Start.Character]) + e.NewText + string(runes[e.Range.End.Character:])
}

func TestClosesCommentAfterSlashStarSpace(t *testing.T) {
	got := typeSpace(t, "++USERMOD(U1) /* ")
	want := "++USERMOD(U1) /*  */"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Only a space directly behind "/*" opens a comment.
func TestDoesNotCloseElsewhere(t *testing.T) {
	for _, text := range []string{
		"++USERMOD(U1) ",         // no comment at all
		"++USERMOD(U1) /* text ", // space further inside the comment
		"++USERMOD(U1) /*text ",  // no space after the marker
		"++USERMOD(U1) * ",       // not a comment marker
	} {
		if got := typeSpace(t, text); got != text {
			t.Errorf("%q was changed to %q, expected no edit", text, got)
		}
	}
}

// A comment that is already closed further along the line must not get a
// second marker - that would end it in the wrong place.
func TestDoesNotCloseAnAlreadyClosedComment(t *testing.T) {
	text := "++USERMOD(U1) /* already */"
	// Cursor right after the opening "/* ", not at the end of the line.
	store, _ := data.Load("../../data/smpe.json")
	doc := parser.NewParser(store.Statements).Parse(text)
	pos := lsp.Position{Line: 0, Character: 17} // just past "/* "
	if got := NewProvider().CloseComment(doc, text, pos, " "); len(got) != 0 {
		t.Errorf("inserted %+v although the comment is already closed", got)
	}
}

// The closing marker must fit up to column 72, counted from where it starts.
func TestRespectsColumn72(t *testing.T) {
	// Build a line whose "/* " ends exactly at a chosen column.
	lineEndingAt := func(col int) string {
		pad := col - len("/* ")
		return strings.Repeat("X", pad) + "/* "
	}

	// Ends at 69: the three characters of " */" occupy 70, 71, 72 - fits.
	fits := lineEndingAt(69)
	if got := typeSpace(t, fits); got != fits+closeCommentInsert {
		t.Errorf("at column 69 the marker should fit, got %q", got[max(0, len(got)-12):])
	}

	// Ends at 70: " */" would reach column 73 - too far.
	tooFar := lineEndingAt(70)
	if got := typeSpace(t, tooFar); got != tooFar {
		t.Errorf("at column 70 the marker must not be inserted, got %q", got[max(0, len(got)-12):])
	}
}

// Inline data is the element's own text. A "/*" there opens a REXX program,
// not an MCS comment, and the server keeps out.
func TestLeavesInlineDataAlone(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	text := "++SRC(MYEXEC) DISTLIB(AOSB3) .\n/* "
	doc := parser.NewParser(store.Statements).Parse(text)
	pos := lsp.Position{Line: 1, Character: 3}

	if got := NewProvider().CloseComment(doc, text, pos, " "); len(got) != 0 {
		t.Errorf("inserted %+v into inline data", got)
	}
}

// Any other character leaves the document untouched.
func TestOnlyReactsToASpace(t *testing.T) {
	store, _ := data.Load("../../data/smpe.json")
	text := "++USERMOD(U1) /* "
	doc := parser.NewParser(store.Statements).Parse(text)
	pos := lsp.Position{Line: 0, Character: len([]rune(text))}
	for _, ch := range []string{"*", "/", "x", ""} {
		if got := NewProvider().CloseComment(doc, text, pos, ch); len(got) != 0 {
			t.Errorf("character %q produced %+v", ch, got)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
