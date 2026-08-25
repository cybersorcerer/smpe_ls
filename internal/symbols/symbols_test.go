package symbols

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

func testParser(t *testing.T) *parser.Parser {
	t.Helper()
	statements := map[string]data.MCSStatement{
		"++USERMOD": {
			Name:      "++USERMOD",
			Parameter: "usermod_name",
			Operands: []data.Operand{
				{Name: "REWORK", Parameter: "rework_id"},
				{Name: "DESC", Parameter: "description"},
			},
		},
	}
	return parser.NewParser(statements)
}

// TestGetStatementEndPositionSkipsDatesInComments verifies a '.' inside a
// /* ... */ block comment (e.g. a German-style "dd.mm.yy" date) is not
// mistaken for the statement terminator, and the real terminator further
// down is found instead. Regression test for the range/outline mismatch
// found with a GitLab CI metadata comment block containing dated entries
// like "19.11.25" and "22.04.22".
func TestGetStatementEndPositionSkipsDatesInComments(t *testing.T) {
	src := strings.Join([]string{
		"++USERMOD(U1)",
		"    REWORK(2026233)",
		"    DESC(test)",
		"  /*",
		"  | CHGS: 19.11.25/XV880AJ- some change",
		"  | CHGS: 22.04.22/XV880AJ- another change",
		"  */",
		".",
		"",
	}, "\n")

	p := testParser(t)
	doc := p.Parse(src)
	if len(doc.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(doc.Statements))
	}
	lines := strings.Split(src, "\n")

	prov := NewProvider()
	endLine, endChar := prov.GetStatementEndPosition(doc.Statements[0], lines)

	// The real terminator '.' is alone on line 7 (0-based).
	if endLine != 7 || endChar != 1 {
		t.Errorf("expected terminator at line 7 char 1, got line %d char %d (line content: %q)",
			endLine, endChar, lines[endLine])
	}
}

// TestGetStatementEndPositionNoComment verifies the ordinary case (no
// comment involved) still finds the terminator correctly.
func TestGetStatementEndPositionNoComment(t *testing.T) {
	src := strings.Join([]string{
		"++USERMOD(U1)",
		"    REWORK(2026233)",
		".",
		"",
	}, "\n")

	p := testParser(t)
	doc := p.Parse(src)
	lines := strings.Split(src, "\n")

	prov := NewProvider()
	endLine, endChar := prov.GetStatementEndPosition(doc.Statements[0], lines)

	if endLine != 2 || endChar != 1 {
		t.Errorf("expected terminator at line 2 char 1, got line %d char %d", endLine, endChar)
	}
}
