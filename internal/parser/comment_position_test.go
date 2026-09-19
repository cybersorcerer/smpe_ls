package parser

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
)

// Known bug, not fixed yet: positions must address the original text, not the
// text with comments cut out.
//
// Parse() builds a comment-free copy of every line and computes all positions
// on it. A comment is replaced by a single space, so everything to its right
// moves left by the length of the comment minus one. An operand behind an
// inline comment therefore reports the column of the comment text:
//
//	++MAC(A) /* c */ FROMDS(DSN(MY.DATA)) .   FROMDS reported at column 11
//	++MAC(A) FROMDS(DSN(MY.DATA)) .           FROMDS reported at column 9
//
// Diagnostics then underline the comment instead of the operand, and
// smpe_outl reports "FROMDS" without its parameter because the '(' is no
// longer where the operand ends.
//
// Widening the comment to spaces instead of cutting it out fixes every
// position, but the parameter values are read from those same lines, so they
// would carry the blanked columns into the formatter, which prints parameter
// values verbatim by design. ++ASSIGN TO(A1, /* c */ A2) then formats with
// long runs of spaces and an unwanted line break. The fix therefore needs a
// position mapping that leaves the values alone, which is scheduled for a
// release of its own.
//
// These tests describe the wanted behaviour and are skipped until then.

const positionBugSkip = "known bug: positions are computed on comment-free lines, see file comment"

// posText returns the original text at a node's position, which must be the
// node's own name.
func posText(t *testing.T, text string, line, char, length int) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		t.Fatalf("line %d beyond text", line)
	}
	l := []rune(lines[line])
	if char+length > len(l) {
		return "<out of range>"
	}
	return string(l[char : char+length])
}

func parseOne(t *testing.T, text string) *Node {
	t.Helper()
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	doc := NewParser(store.Statements).Parse(text)
	if len(doc.Statements) == 0 {
		t.Fatalf("No statement parsed from %q", text)
	}
	return doc.Statements[0]
}

func TestOperandPositionAfterInlineComment(t *testing.T) {
	t.Skip(positionBugSkip)

	cases := []struct {
		name    string
		text    string
		operand string
	}{
		{"no comment", "++MAC(A) FROMDS(DSN(MY.DATA)) .", "FROMDS"},
		{"comment before operand", "++MAC(A) /* c */ FROMDS(DSN(MY.DATA)) .", "FROMDS"},
		{"long comment before operand", "++MAC(A) /* a much longer comment */ FROMDS(DSN(MY.DATA)) .", "FROMDS"},
		{"two comments", "++MAC(A) /* one */ /* two */ FROMDS(DSN(MY.DATA)) .", "FROMDS"},
		{"comment between operands", "++MAC(A) DISTLIB(ALIB) /* c */ FROMDS(DSN(MY.DATA)) .", "FROMDS"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stmt := parseOne(t, tc.text)
			for _, child := range stmt.Children {
				if child.Type != NodeTypeOperand || child.Name != tc.operand {
					continue
				}
				got := posText(t, tc.text, child.Position.Line, child.Position.Character, child.Position.Length)
				if got != tc.operand {
					t.Errorf("position points at %q, want %q\n  text: %s\n  char=%d len=%d",
						got, tc.operand, tc.text, child.Position.Character, child.Position.Length)
				}
				return
			}
			t.Fatalf("operand %s not found in %q", tc.operand, tc.text)
		})
	}
}

// A comment that ends on a later line shifts the remainder of that line.
func TestOperandPositionAfterMultiLineComment(t *testing.T) {
	t.Skip(positionBugSkip)

	text := "++MAC(A)\n" +
		"   /* a comment\n" +
		"      spanning lines */ FROMDS(DSN(MY.DATA))\n" +
		"   .\n"
	stmt := parseOne(t, text)
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand && child.Name == "FROMDS" {
			got := posText(t, text, child.Position.Line, child.Position.Character, child.Position.Length)
			if got != "FROMDS" {
				t.Errorf("position points at %q, want %q (line %d char %d)",
					got, "FROMDS", child.Position.Line, child.Position.Character)
			}
			return
		}
	}
	t.Fatal("operand FROMDS not found")
}

// The operand parameter has to be reachable from the operand's end, which is
// what smpe_outl relies on to report "FROMDS(DSN(...))" rather than "FROMDS".
func TestOperandParameterFollowsOperandAfterComment(t *testing.T) {
	t.Skip(positionBugSkip)

	text := "++MAC(A) /* c */ FROMDS(DSN(MY.DATA)) ."
	stmt := parseOne(t, text)
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand && child.Name == "FROMDS" {
			end := child.Position.Character + child.Position.Length
			rest := []rune(text)[end:]
			if len(rest) == 0 || rest[0] != '(' {
				t.Errorf("text after operand is %q, want it to start with '('", string(rest))
			}
			return
		}
	}
	t.Fatal("operand FROMDS not found")
}
