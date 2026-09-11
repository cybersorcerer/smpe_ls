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
