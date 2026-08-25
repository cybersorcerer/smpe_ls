package parser

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
)

// leparmLikeStatements returns a minimal statement map that reproduces the
// real ++MOD/LEPARM shape from data/smpe.json: a top-level operand (LEPARM)
// whose Parameter field contains "(" (marking it as having sub-operands),
// and a Values list of simple, parenthesized sub-operands (AC, AMODE) plus
// one bare boolean flag (ALIGN2) — matching the documented bug report in
// LEPARM_BUG.md.
func leparmLikeStatements() map[string]data.MCSStatement {
	return map[string]data.MCSStatement{
		"++MOD": {
			Name:      "++MOD",
			Parameter: "name",
			Operands: []data.Operand{
				{
					Name:      "LEPARM",
					Parameter: "AC(n) ALIGN2 AMODE(mode)",
					Type:      "string",
					Values: []data.AllowedValue{
						{Name: "AC", Parameter: "1", Type: "integer"},
						{Name: "ALIGN2", Type: "boolean"},
						{Name: "AMODE|AMOD", Parameter: "24|31|64|ANY|MIN", Type: "string"},
					},
				},
				{Name: "DISTLIB", Parameter: "ddname", Type: "string"},
			},
		},
	}
}

// TestASTInvariants_SimpleOperand runs the structural invariant suite
// (no duplicate children, no self-referential parameter wrapping, valid
// positions, correct parent linkage) against the most basic case: a
// top-level operand with a single simple value.
func TestASTInvariants_SimpleOperand(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++VER": {
			Name:      "++VER",
			Parameter: "sysrel",
			Operands: []data.Operand{
				{Name: "FMID", Parameter: "sysmod_id", Type: "string"},
			},
		},
	}
	p := NewParser(statements)
	doc := p.Parse("++VER(I700)\n    FMID(CTSA400)\n.")
	checkDocumentInvariants(t, doc)
}

// TestASTInvariants_SubOperandSingleValue reproduces the exact bug shape
// from LEPARM_BUG.md: a sub-operand (AC) with a single, simple parenthesized
// value (AC(1)) inside a parent operand that has sub-operands (LEPARM).
// Before the fix, AC's parameter wrapper node had one child that was an
// identical copy of itself (same Type/Value/Position) — this is exactly
// what checkNoSelfReferentialValue and checkNoDuplicateChildren detect.
func TestASTInvariants_SubOperandSingleValue(t *testing.T) {
	p := NewParser(leparmLikeStatements())
	doc := p.Parse("++MOD(MYPROG)\n    LEPARM(AC(1))\n    DISTLIB(D1)\n.")
	checkDocumentInvariants(t, doc)
}

// TestASTInvariants_SubOperandMultipleCommaSeparated covers the full
// reported case: multiple comma-separated sub-operands, some with values
// (AC, AMODE) and one bare boolean (ALIGN2), inside LEPARM.
func TestASTInvariants_SubOperandMultipleCommaSeparated(t *testing.T) {
	p := NewParser(leparmLikeStatements())
	doc := p.Parse("++MOD(MYPROG)\n    LEPARM(AC(1),ALIGN2,AMODE(24))\n    DISTLIB(D1)\n.")
	checkDocumentInvariants(t, doc)
}

// TestASTInvariants_AcrossAllOperandKinds parses every operand kind found
// in the leparm-like fixture individually to make sure the invariant check
// itself doesn't accidentally only cover one shape.
func TestASTInvariants_AcrossAllOperandKinds(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"bare_boolean", "++MOD(P)\n    LEPARM(ALIGN2)\n    DISTLIB(D1)\n."},
		{"single_value", "++MOD(P)\n    LEPARM(AC(1))\n    DISTLIB(D1)\n."},
		{"aliased_operand", "++MOD(P)\n    LEPARM(AMODE(31))\n    DISTLIB(D1)\n."},
		{"top_level_simple", "++MOD(P)\n    DISTLIB(D1)\n."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser(leparmLikeStatements())
			doc := p.Parse(tc.src)
			checkDocumentInvariants(t, doc)
		})
	}
}

// TestASTInvariants_RealStatementFixtures runs the invariant suite against
// every existing parser test's input pattern, using the same statement
// shapes already exercised in parser_test.go (USERMOD, FROMDS, APAR, HOLD,
// IF, FEATURE, ASSIGN, DELETE), so a future change to any operand-parsing
// path is checked for the duplication bug even if the corresponding
// value-assertion test doesn't happen to look for it.
func TestASTInvariants_RealStatementFixtures(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++USERMOD": {
			Name: "++USERMOD", Parameter: "sysmod_id",
			Operands: []data.Operand{
				{Name: "DESC", Parameter: "description", Type: "string"},
				{Name: "REWORK", Parameter: "date", Type: "date"},
			},
		},
		"++MOD": {
			Name: "++MOD", Parameter: "name",
			Operands: []data.Operand{
				{Name: "DISTLIB", Parameter: "ddname", Type: "string"},
				{
					Name:      "FROMDS",
					Parameter: "DSN(dsname) VOL(volser) UNIT(unit) NUMBER(number)",
					Type:      "string",
					Values: []data.AllowedValue{
						{Name: "DSN", Parameter: "dsname", Type: "string"},
						{Name: "VOL", Parameter: "volser", Type: "string"},
						{Name: "UNIT", Parameter: "unit", Type: "string"},
						{Name: "NUMBER", Parameter: "number", Type: "integer"},
					},
				},
			},
		},
	}
	cases := []string{
		"++USERMOD(LJS2012) REWORK(2022056)",
		"++MOD(MYMOD)\n    FROMDS(DSN(MY.DATASET) VOL(VOL001) UNIT(SYSDA) NUMBER(1))\n.",
	}
	p := NewParser(statements)
	for _, src := range cases {
		doc := p.Parse(src)
		checkDocumentInvariants(t, doc)
	}
}

// TestMultiValueParameterListPreserved is a regression guard for the fix to
// the parameter-node-duplication bug: PRE(UJ001,UJ002,UJ003)-style operands
// (comma-separated multi-value lists, no sub-operand Values definition —
// hasSubOperands is false) must keep producing one child node per value
// under the parameter wrapper, for per-value highlighting/navigation. The
// duplication fix only removes the redundant single-child wrapper case; it
// must not collapse this legitimate multi-value case down to zero or one
// child.
func TestMultiValueParameterListPreserved(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++PTF": {
			Name: "++PTF", Parameter: "sysmod_id",
			Operands: []data.Operand{
				{Name: "PRE", Parameter: "SYSMOD_IDs", Type: "list"},
			},
		},
	}
	p := NewParser(statements)
	doc := p.Parse("++PTF(UA12345)\n    PRE(UJ001,UJ002,UJ003)\n.")
	checkDocumentInvariants(t, doc)

	if len(doc.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(doc.Statements))
	}
	stmt := doc.Statements[0]

	var preOperand *Node
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand && child.Name == "PRE" {
			preOperand = child
			break
		}
	}
	if preOperand == nil {
		t.Fatal("PRE operand not found")
	}
	if len(preOperand.Children) != 1 {
		t.Fatalf("expected PRE to have 1 wrapper child, got %d", len(preOperand.Children))
	}
	wrapper := preOperand.Children[0]

	want := []string{"UJ001", "UJ002", "UJ003"}
	if len(wrapper.Children) != len(want) {
		t.Fatalf("expected wrapper to have %d value children, got %d", len(want), len(wrapper.Children))
	}
	for i, w := range want {
		if wrapper.Children[i].Value != w {
			t.Errorf("value child %d: expected %q, got %q", i, w, wrapper.Children[i].Value)
		}
	}
}
