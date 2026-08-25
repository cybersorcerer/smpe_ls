package completion

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// leparmLikeStatements mirrors the real ++MOD/LEPARM shape from
// data/smpe.json: a top-level operand (LEPARM) whose Parameter field
// contains "(" (marking it as having sub-operands), with Values entries
// for AC (a sub-operand that itself takes a simple integer value, no
// further sub-operands), ALIGN2 (bare boolean), and AMODE (a sub-operand
// with an enumerated string value).
func leparmLikeStatements() (*data.Store, *parser.Parser, *Provider) {
	statements := map[string]data.MCSStatement{
		"++MOD": {
			Name:      "++MOD",
			Parameter: "name",
			Operands: []data.Operand{
				{
					Name:      "LEPARM",
					Parameter: "AC(n) ALIGN2 AMODE(mode)",
					Type:      "string",
					Values: []data.AllowedValue{
						{Name: "AC", Parameter: "1", Type: "integer", Description: "Authorization code"},
						{Name: "ALIGN2", Type: "boolean", Description: "Align on doubleword"},
						{Name: "AMODE|AMOD", Parameter: "24|31|64|ANY|MIN", Type: "string", Description: "Addressing mode"},
					},
				},
				{Name: "DISTLIB", Parameter: "ddname", Type: "string"},
			},
		},
	}
	store := &data.Store{Statements: statements, List: []data.MCSStatement{statements["++MOD"]}}
	p := parser.NewParser(statements)
	cp := NewProvider(store)
	return store, p, cp
}

// TestCompletionInsideLeparmSubOperandParens is a regression test for the
// context-detection bug found while investigating LEPARM support (see
// LEPARM_BUG.md): when the cursor is inside a sub-operand's own
// parentheses — e.g. LEPARM(AC(│) — findOperandAtPosition returned the
// outer LEPARM operand instead of the inner AC sub-operand, because it
// returned as soon as the cursor fell within the outer operand's full
// parenthesized range, before ever reaching the recursive sub-operand
// search that would have found AC.
//
// The practical effect: completion inside AC(...) offered all 3 LEPARM
// top-level sub-operand names (AC, ALIGN2, AMODE) instead of AC's own
// completions (AC takes a plain integer, so no further completions are
// expected there — but the bug is that LEPARM's names leaked in at all).
func TestCompletionInsideLeparmSubOperandParens(t *testing.T) {
	_, p, cp := leparmLikeStatements()

	// Cursor right after "LEPARM(AC(" — inside AC's own parentheses.
	src := "++MOD(MYPROG)\n    LEPARM(AC(),ALIGN2)\n    DISTLIB(D1)\n."
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 1, 14) // position right after "AC("

	for _, item := range items {
		if item.Label == "ALIGN2" || item.Label == "AMODE" || item.Label == "AMOD" {
			t.Errorf("completion inside AC(...) leaked outer LEPARM sub-operand %q — "+
				"findOperandAtPosition should have resolved to AC, not LEPARM. Got %d items total.",
				item.Label, len(items))
		}
	}
}

// TestCompletionAfterLeparmCommaStillOffersSiblings is a baseline/guardrail
// test: the case that already worked correctly (cursor right after a comma
// at LEPARM's own level, e.g. LEPARM(AC(1),│) must keep offering the
// remaining LEPARM sub-operand names. The fix for the bug above must not
// break this.
func TestCompletionAfterLeparmCommaStillOffersSiblings(t *testing.T) {
	_, p, cp := leparmLikeStatements()

	src := "++MOD(MYPROG)\n    LEPARM(AC(1),)\n    DISTLIB(D1)\n."
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 1, 17) // position right after the comma

	found := map[string]bool{}
	for _, item := range items {
		found[item.Label] = true
	}
	for _, want := range []string{"ALIGN2", "AMODE"} {
		if !found[want] {
			t.Errorf("expected %q among LEPARM sibling completions after comma, got %v", want, found)
		}
	}
}

// TestCompletionLeparmDoesNotReofferUsedSubOperand is a regression test for
// a bug reported after fixing the nested-context bug above: sub-operand
// completion inside LEPARM was not context-sensitive about which
// sub-operands were already used, unlike top-level operand completion
// (getOperandCompletionsAST, which filters via presentOperands). After
// LEPARM(AC(1),│ the cursor is back at LEPARM's own level, offering
// sibling sub-operands is correct — but AC itself, already used, must not
// be offered again.
func TestCompletionLeparmDoesNotReofferUsedSubOperand(t *testing.T) {
	_, p, cp := leparmLikeStatements()

	src := "++MOD(MYPROG)\n    LEPARM(AC(1),)\n."
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 1, 17) // position right after the comma

	for _, item := range items {
		if item.Label == "AC" {
			t.Errorf("expected AC to be excluded from completions (already used), got %v", symbolLabels(items))
		}
	}
}

// TestCompletionLeparmAliasUsageExcludesBothNames verifies alias handling:
// using AMOD (an alias of AMODE|AMOD) must exclude both AMODE and AMOD from
// further completions at that level, matching how getOperandCompletionsAST
// treats aliases for top-level operands.
func TestCompletionLeparmAliasUsageExcludesBothNames(t *testing.T) {
	_, p, cp := leparmLikeStatements()

	// Indices in line 1 ("    LEPARM(AMOD(31),)"): '(' after AMOD at 15,
	// ')' at 18, ',' at 19, final ')' at 20. Right after the comma is 20.
	src := "++MOD(MYPROG)\n    LEPARM(AMOD(31),)\n."
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 1, 20) // position right after the comma

	for _, item := range items {
		if item.Label == "AMODE" || item.Label == "AMOD" {
			t.Errorf("expected AMODE/AMOD to be excluded after AMOD was used, got %v", symbolLabels(items))
		}
	}
	found := map[string]bool{}
	for _, item := range items {
		found[item.Label] = true
	}
	if !found["AC"] || !found["ALIGN2"] {
		t.Errorf("expected AC and ALIGN2 to remain offered, got %v", symbolLabels(items))
	}
}

func symbolLabels(items []lsp.CompletionItem) []string {
	labels := make([]string, len(items))
	for i, it := range items {
		labels[i] = it.Label
	}
	return labels
}

// TestCompletionInsideFromdsSubOperandParens covers the same bug shape for
// FROMDS, the other real sub-operand-container operand confirmed affected
// during diagnosis (see LEPARM_BUG.md). Cursor inside DSN(...) must not
// offer VOL/UNIT/NUMBER (FROMDS's other sub-operands).
func TestCompletionInsideFromdsSubOperandParens(t *testing.T) {
	_, p, cp := createTestProviders()

	// Indices: ...FROMDS( at 19, DSN( at 20-23 ('(' is index 23), ')' at 24, ')' at 25.
	src := "++MAC(MYMAC) FROMDS(DSN())"
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 0, 24) // position right after "DSN(", i.e. inside DSN's own parens

	for _, item := range items {
		if item.Label == "VOL" || item.Label == "UNIT" || item.Label == "NUMBER" {
			t.Errorf("completion inside DSN(...) leaked outer FROMDS sub-operand %q", item.Label)
		}
	}
}

// TestCompletionPipeSeparatedParameterValues is a regression test for a bug
// found while fixing the FETCHOPT data gap in data/smpe.json (see
// LEPARM_BUG.md): an operand whose OperandDef has no Values entries but
// whose Parameter field is itself a pipe-separated enumeration (e.g.
// AMODE's "24|31|64|ANY|MIN", UPCASE's "YES|NO") offered zero completion
// items when the cursor was inside that operand's own parentheses —
// getOperandValueCompletionsAST only handled the Values-populated case and
// fell through to a bare "return nil" otherwise. This affected every such
// operand, not just the one that surfaced it (FETCHOPT), including AMODE,
// COMPAT, UPCASE, RMODE, and RMODEX.
func TestCompletionPipeSeparatedParameterValues(t *testing.T) {
	_, p, cp := leparmLikeStatements()

	// AMODE(│) — Parameter is "24|31|64|ANY|MIN", no Values entries.
	src := "++MOD(MYPROG)\n    LEPARM(AMODE())\n."
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 1, 17) // inside AMODE(

	found := map[string]bool{}
	for _, item := range items {
		found[item.Label] = true
	}
	for _, want := range []string{"24", "31", "64", "ANY", "MIN"} {
		if !found[want] {
			t.Errorf("expected %q among AMODE value completions, got %v", want, found)
		}
	}
}

// TestCompletionPipeSeparatedValuesDoNotLeakSiblings verifies the pipe-
// value fix scopes correctly: completions inside AC(...) — a sub-operand
// with a plain, non-enumerated parameter ("1", an integer) — must still be
// empty, not accidentally pick up LEPARM's or any other operand's values.
func TestCompletionPipeSeparatedValuesDoNotLeakSiblings(t *testing.T) {
	_, p, cp := leparmLikeStatements()

	src := "++MOD(MYPROG)\n    LEPARM(AC())\n."
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 1, 14) // inside AC(

	for _, item := range items {
		t.Errorf("expected no completions inside AC(...) (plain integer parameter), got %q", item.Label)
	}
}

// TestCompletionFromdsDoesNotReofferUsedSubOperand verifies the "already
// present" filter also applies to FROMDS, the other real sub-operand
// container operand: once DSN is set, it must not be offered again, while
// the remaining optional sub-operands (VOL, UNIT, NUMBER) still are.
func TestCompletionFromdsDoesNotReofferUsedSubOperand(t *testing.T) {
	_, p, cp := createTestProviders()

	// Indices: "++MAC(MYMAC) FROMDS(DSN(MY.DATASET) )" is 38 chars (0-37);
	// the space before the closing ')' is at index 35, so the position
	// right after it (still inside FROMDS's parens) is 36.
	src := "++MAC(MYMAC) FROMDS(DSN(MY.DATASET) )"
	doc := p.Parse(src)
	items := cp.GetCompletionsAST(doc, src, 0, 36)

	found := map[string]bool{}
	for _, item := range items {
		found[item.Label] = true
	}
	if found["DSN"] {
		t.Errorf("expected DSN to be excluded (already used), got %v", symbolLabels(items))
	}
	for _, want := range []string{"VOL", "UNIT", "NUMBER"} {
		if !found[want] {
			t.Errorf("expected %q to remain offered, got %v", want, symbolLabels(items))
		}
	}
}
