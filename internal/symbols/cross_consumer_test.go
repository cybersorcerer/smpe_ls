package symbols

import (
	"strings"
	"testing"
)

// TestGetStatementEndPosition_ConsistentAcrossConsumers guards against the
// exact class of bug found and fixed in this project (see LEPARM_BUG.md and
// the v1.3.8 changelog entry): GetStatementEndPosition/GetSymbolKind used to
// be duplicated in cmd/smpe_outl/main.go as a separate copy of the same
// logic. A bug fixed in one copy silently kept failing in the other,
// because no test exercised both consumers against the same input.
//
// cmd/smpe_outl now calls this Provider directly instead of carrying its
// own copy, which structurally prevents that specific class of drift. This
// test locks that in: GetDocumentSymbols (used by the LSP server) and
// GetStatementEndPosition/GetSymbolKind (called directly by smpe_outl) must
// agree on the same statement's end position and kind for identical input.
// If someone reintroduces a second copy of this logic anywhere, this test
// gives no protection by itself — but it does mean the two current call
// sites can never quietly disagree without failing here first.
func TestGetStatementEndPosition_ConsistentAcrossConsumers(t *testing.T) {
	src := strings.Join([]string{
		"++USERMOD(U1)",
		"    REWORK(2026233)",
		"    DESC(text)",
		"  /*",
		"  | CHGS: 19.11.25/XV880AJ- some change",
		"  */",
		".",
		"",
	}, "\n")

	p := testParser(t)
	doc := p.Parse(src)
	if len(doc.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(doc.Statements))
	}
	stmt := doc.Statements[0]
	lines := strings.Split(src, "\n")

	prov := NewProvider()

	// Consumer 1: direct call, as cmd/smpe_outl/main.go does.
	directEndLine, directEndChar := prov.GetStatementEndPosition(stmt, lines)
	directKind := prov.GetSymbolKind(stmt.Name)

	// Consumer 2: via GetDocumentSymbols, as the LSP server's
	// textDocument/documentSymbol handler does.
	docSyms := prov.GetDocumentSymbols(doc, lines)
	if len(docSyms) != 1 {
		t.Fatalf("expected 1 document symbol, got %d", len(docSyms))
	}
	viaDocSymEndLine := docSyms[0].Range.End.Line
	viaDocSymEndChar := docSyms[0].Range.End.Character
	viaDocSymKind := docSyms[0].Kind

	if directEndLine != viaDocSymEndLine || directEndChar != viaDocSymEndChar {
		t.Errorf("end position disagreement: direct=(%d,%d) via GetDocumentSymbols=(%d,%d)",
			directEndLine, directEndChar, viaDocSymEndLine, viaDocSymEndChar)
	}
	if directKind != viaDocSymKind {
		t.Errorf("symbol kind disagreement: direct=%v via GetDocumentSymbols=%v", directKind, viaDocSymKind)
	}

	// Sanity: both must find the real terminator (line 6), not the date
	// inside the comment (line 4) — this is the v1.3.8 bug this whole
	// fixture is modeled on.
	if directEndLine != 6 {
		t.Errorf("expected terminator at line 6, got line %d (line content: %q)", directEndLine, lines[directEndLine])
	}
}
