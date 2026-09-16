package symbols

import (
	"strings"
	"testing"
)

// endLineFor parses src and returns the end line the provider computes for the
// first statement.
func endLineFor(t *testing.T, src string) int {
	t.Helper()
	doc := testParser(t).Parse(src)
	if len(doc.Statements) == 0 {
		t.Fatalf("no statement parsed from %q", src)
	}
	line, _ := NewProvider().GetStatementEndPosition(doc.Statements[0], strings.Split(src, "\n"))
	return line
}

// A dot inside an operand value is not the statement terminator. Free-text
// operands such as DESC routinely contain one, which cut the symbol range
// short at that operand instead of the real terminator.
func TestStatementEndIgnoresDotInOperandValue(t *testing.T) {
	src := "++USERMOD(LIIQ101)\n" +
		"    REWORK(2026239)\n" +
		"    DESC(R+V IIQ. SMF Exit 83)\n" +
		".\n"
	if got := endLineFor(t, src); got != 3 {
		t.Errorf("Expected end line 3 (the terminator), got %d", got)
	}
}

// The same for a dotted dataset name in a nested operand.
func TestStatementEndIgnoresDotInNestedValue(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"    DESC(HLQ.MID.LLQ)\n" +
		".\n"
	if got := endLineFor(t, src); got != 2 {
		t.Errorf("Expected end line 2, got %d", got)
	}
}

// A dot inside a single-quoted value must not terminate either.
func TestStatementEndIgnoresDotInQuotedValue(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"    DESC('version 1.5 of the exit')\n" +
		".\n"
	if got := endLineFor(t, src); got != 2 {
		t.Errorf("Expected end line 2, got %d", got)
	}
}

// Regression guard for the 1.3.8 fix: a dot in a block comment stays ignored.
func TestStatementEndIgnoresDotInComment(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"    REWORK(2026239) /* changed 22.04.22 */\n" +
		".\n"
	if got := endLineFor(t, src); got != 2 {
		t.Errorf("Expected end line 2, got %d", got)
	}
}

// The terminator on the operand's own line is still found.
func TestStatementEndFindsTerminatorOnOperandLine(t *testing.T) {
	src := "++USERMOD(U1) REWORK(2026239) .\n"
	if got := endLineFor(t, src); got != 0 {
		t.Errorf("Expected end line 0, got %d", got)
	}
}

// A "++" inside a comment is text, not the start of the next statement. Only a
// "++" at the beginning of a line ends the search.
func TestStatementEndIgnoresPlusPlusInComment(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"  /* Test file for ++USERMOD MCS statement */\n" +
		"  /* Valid: Minimal ++USERMOD */\n" +
		".\n"
	if got := endLineFor(t, src); got != 3 {
		t.Errorf("Expected end line 3 (the terminator), got %d", got)
	}
}

// The next statement still ends the search when it starts the line.
func TestStatementEndStopsAtNextStatement(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"++VER(Z038) FMID(F1) .\n"
	if got := endLineFor(t, src); got != 0 {
		t.Errorf("Expected end line 0, got %d", got)
	}
}

// With unbalanced parentheses the depth is meaningless: the statement is
// already reported as malformed, and the terminator must still be found so no
// follow-up diagnostics are produced.
func TestStatementEndWithUnbalancedParens(t *testing.T) {
	src := "++USERMOD(U44444\n" +
		"/* malformed parentheses */\n" +
		"    DESC(Missing closing paren)\n" +
		"    .\n"
	if got := endLineFor(t, src); got != 3 {
		t.Errorf("Expected end line 3 (the terminator), got %d", got)
	}
}

// Leading whitespace does not hide a statement: what counts is a "++" as the
// first non-whitespace character of the line, not literally column 1.
func TestStatementEndStopsAtIndentedNextStatement(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"     ++VER(Z038) FMID(F1) .\n"
	if got := endLineFor(t, src); got != 0 {
		t.Errorf("Expected end line 0, got %d", got)
	}
}

// An indented statement that does have a terminator still finds it.
func TestStatementEndFindsTerminatorOfIndentedStatement(t *testing.T) {
	src := "  ++USERMOD(U1)\n" +
		"      REWORK(2026259)\n" +
		"  .\n"
	if got := endLineFor(t, src); got != 2 {
		t.Errorf("Expected end line 2, got %d", got)
	}
}

// A statement written inside a comment is text. Neither the "++" nor the
// terminator in "/* ++APAR(A1) . */" may end the enclosing statement.
func TestStatementEndIgnoresStatementInsideComment(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"/* ++APAR(A1) . */\n" +
		"    REWORK(2026259)\n" +
		".\n"
	if got := endLineFor(t, src); got != 3 {
		t.Errorf("Expected end line 3 (the terminator), got %d", got)
	}
}

// The same across lines: a line inside an open block comment may start with
// "++" and still is not a statement. This is why the check runs before the
// line is scanned, while inBlockComment still holds the state at line start.
func TestStatementEndIgnoresStatementInsideMultiLineComment(t *testing.T) {
	src := "++USERMOD(U1)\n" +
		"/* example:\n" +
		"++APAR(A1) .\n" +
		"*/\n" +
		"    REWORK(2026259)\n" +
		".\n"
	if got := endLineFor(t, src); got != 5 {
		t.Errorf("Expected end line 5 (the terminator), got %d", got)
	}
}
