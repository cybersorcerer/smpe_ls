package formatting

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

// Characterization test for the "which operand replaces inline data" rule as
// the formatter applies it. Inline data must never be touched, so the rule
// decides whether the lines after the terminator are data or MCS text.
//
// The rule is restated here rather than imported, so the test keeps its value
// when the production code moves the list into smpe.json.

// A line that looks like an operand but is really inline data. If the
// formatter considers the statement to carry inline data it must leave this
// line alone; otherwise it reformats it as part of the statement.
const probeDataLine = "distlib(alib)   txlib(mylib)"

func TestCharacterizeFormatterInlineDataBoundary(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := parser.NewParser(store.Statements)
	fp := NewProvider()
	fp.SetConfig(&Config{
		Enabled:             true,
		IndentContinuation:  4,
		OneOperandPerLine:   true,
		WrapListsAfterN:     2,
		MoveLeadingComments: true,
	})

	variants := []struct {
		label          string
		operand        string
		replacesInline bool
	}{
		{"none", "", false},
		{"FROMDS", "FROMDS(DSN(MY.SOURCE.DS))", true},
		{"RELFILE", "RELFILE(1)", true},
		{"TXLIB", "TXLIB(MYLIB)", true},
		{"LKLIB", "LKLIB(MYLIB)", true},
		{"DELETE", "DELETE", true},
	}

	checked := 0
	for name, def := range store.Statements {
		if !def.InlineData {
			continue
		}
		for _, v := range variants {
			head := name + "(E1) DISTLIB(ALIB)"
			if v.operand != "" {
				head += " " + v.operand
			}
			input := head + " .\n" + probeDataLine + "\n++VER(Z038) FMID(F1) .\n"

			doc := p.Parse(input)
			stmt := doc.Statements[0]
			got := fp.stmtExpectsInlineData(stmt)
			if want := !v.replacesInline; got != want {
				t.Errorf("%s + %s: stmtExpectsInlineData = %v, want %v",
					name, v.label, got, want)
			}

			// The edits must never reach past the terminator line when the
			// statement carries inline data.
			for _, e := range fp.FormatDocument(doc, input) {
				if e.Range.Start.Line == 0 && !v.replacesInline && e.Range.End.Line > 0 {
					t.Errorf("%s + %s: edit reaches into inline data (line %d)",
						name, v.label, e.Range.End.Line)
				}
				if strings.Contains(e.NewText, probeDataLine) && !v.replacesInline {
					t.Errorf("%s + %s: inline data pulled into the statement:\n%s",
						name, v.label, e.NewText)
				}
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("No statement was checked - the matrix would be empty")
	}
	t.Logf("checked %d statement/operand combinations", checked)
}
