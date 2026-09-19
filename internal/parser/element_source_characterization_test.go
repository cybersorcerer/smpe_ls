package parser

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
)

// Characterization test for the "which operand replaces inline data" rule as
// the parser applies it when filling Document.StatementsExpectingInline.
//
// The rule is restated here rather than imported, so the test keeps its value
// when the production code moves the list into smpe.json.

func elementSourceVariants() []struct {
	label          string
	operand        string
	replacesInline bool
} {
	return []struct {
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
}

// A statement only lands in StatementsExpectingInline when nothing supplies
// its data from elsewhere and it is not being deleted.
func TestCharacterizeStatementsExpectingInline(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := NewParser(store.Statements)

	checked := 0
	for name, def := range store.Statements {
		if !def.InlineData {
			continue
		}
		for _, v := range elementSourceVariants() {
			head := name + "(E1) DISTLIB(ALIB)"
			if v.operand != "" {
				head += " " + v.operand
			}
			text := head + " .\nDATA LINE\n++VER(Z038) FMID(F1) .\n"

			doc := p.Parse(text)
			expecting := false
			for _, s := range doc.StatementsExpectingInline {
				if s.Name == name {
					expecting = true
					break
				}
			}

			if want := !v.replacesInline; expecting != want {
				t.Errorf("%s + %s: in StatementsExpectingInline = %v, want %v",
					name, v.label, expecting, want)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("No statement was checked - the matrix would be empty")
	}
	t.Logf("checked %d statement/operand combinations", checked)
}

// Statements that smpe.json does not mark as carrying inline data never end
// up in the list, whatever operands they have.
func TestCharacterizeNonInlineStatementsNeverExpectInline(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := NewParser(store.Statements)

	for name, def := range store.Statements {
		if def.InlineData {
			continue
		}
		text := name + "(E1) .\nDATA LINE\n++VER(Z038) FMID(F1) .\n"
		doc := p.Parse(text)
		for _, s := range doc.StatementsExpectingInline {
			if s.Name == name {
				t.Errorf("%s carries no inline_data but expects inline data", name)
			}
		}
	}
}
