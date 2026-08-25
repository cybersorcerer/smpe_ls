package main

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

// leparmLikeStatements mirrors the real ++MOD/LEPARM shape from
// data/smpe.json closely enough to reproduce the bug documented in
// LEPARM_BUG.md: a top-level operand (LEPARM) whose Parameter field
// contains "(" (sub-operands), with Values entries for AC (parenthesized
// value), ALIGN2 (bare boolean), and AMODE (parenthesized value).
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

func symbolNames(syms []OutlineSymbol) []string {
	names := make([]string, len(syms))
	for i, s := range syms {
		names[i] = s.Name
	}
	return names
}

func contains(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

// TestBuildOperandSymbols_LeparmMissing is a regression test for the bug
// documented in LEPARM_BUG.md ("Auswirkung auf smpe_outl", point 3):
// an operand with sub-operands (like LEPARM) instead of a single direct
// value was silently dropped from buildOperandSymbols because it looked
// for a direct NodeTypeParameter child and gave up when it found none.
// Fixed by falling back to the raw source text between the operand's
// parentheses (extractParenthesizedText) when no direct parameter value
// exists.
func TestBuildOperandSymbols_LeparmMissing(t *testing.T) {
	src := "++MOD(MYPROG)\n    LEPARM(AC(1))\n    DISTLIB(D1)\n."
	p := parser.NewParser(leparmLikeStatements())
	doc := p.Parse(src)
	if len(doc.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(doc.Statements))
	}
	lines := strings.Split(src, "\n")

	syms := buildSymbols(doc, lines, false, false)
	if len(syms) != 1 {
		t.Fatalf("expected 1 top-level symbol, got %d", len(syms))
	}

	names := symbolNames(syms[0].Children)
	if !contains(names, "DISTLIB(D1)") {
		t.Errorf("expected DISTLIB(D1) among children, got %v", names)
	}
	if !contains(names, "LEPARM(AC(1))") {
		t.Errorf("expected LEPARM(AC(1)) among children, got %v", names)
	}
}

// TestBuildOperandSymbols_LeparmMultiValue covers the reported case with
// multiple comma-separated sub-operands, some with values and one bare
// boolean flag — the exact shape from the original bug report.
func TestBuildOperandSymbols_LeparmMultiValue(t *testing.T) {
	src := "++MOD(MYPROG)\n    LEPARM(AC(1),ALIGN2,AMODE(24))\n."
	p := parser.NewParser(leparmLikeStatements())
	doc := p.Parse(src)
	lines := strings.Split(src, "\n")

	syms := buildSymbols(doc, lines, false, false)
	names := symbolNames(syms[0].Children)
	want := "LEPARM(AC(1),ALIGN2,AMODE(24))"
	if !contains(names, want) {
		t.Errorf("expected %q among children, got %v", want, names)
	}
}

// TestBuildOperandSymbols_LeparmMultiLine verifies a LEPARM whose value
// spans multiple source lines is reconstructed as a single-line, comma-
// separated value with internal line breaks and their leading indentation
// collapsed into single spaces — not with all the original whitespace
// preserved verbatim.
func TestBuildOperandSymbols_LeparmMultiLine(t *testing.T) {
	src := "++MOD(MYPROG)\n    LEPARM(AC(1),\n           ALIGN2,\n           AMODE(24))\n."
	p := parser.NewParser(leparmLikeStatements())
	doc := p.Parse(src)
	lines := strings.Split(src, "\n")

	syms := buildSymbols(doc, lines, true, false)
	names := symbolNames(syms[0].Children)
	want := "LEPARM(AC(1), ALIGN2, AMODE(24))"
	if !contains(names, want) {
		t.Errorf("expected %q among children, got %v", want, names)
	}

	// Also verify the range spans from the operand start to the closing
	// paren on the last line (line 3, 0-based), not just the first line.
	var leparm *OutlineSymbol
	for i := range syms[0].Children {
		if syms[0].Children[i].Name == want {
			leparm = &syms[0].Children[i]
			break
		}
	}
	if leparm == nil || leparm.Range == nil {
		t.Fatal("expected LEPARM symbol with a Range")
	}
	if leparm.Range.End.Line != 3 {
		t.Errorf("expected range to end on line 3 (last LEPARM source line), got line %d", leparm.Range.End.Line)
	}
}

// TestBuildOperandSymbols_FromdsNotDropped verifies the fix also covers
// FROMDS, the other real-world sub-operand container operand in
// data/smpe.json, confirmed during diagnosis to have the exact same
// "silently dropped" bug as LEPARM.
func TestBuildOperandSymbols_FromdsNotDropped(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++MOD": {
			Name: "++MOD", Parameter: "name",
			Operands: []data.Operand{
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
	src := "++MOD(MYPROG)\n    FROMDS(DSN(MY.DATASET) VOL(VOL001) UNIT(SYSDA) NUMBER(1))\n."
	p := parser.NewParser(statements)
	doc := p.Parse(src)
	lines := strings.Split(src, "\n")

	syms := buildSymbols(doc, lines, false, false)
	names := symbolNames(syms[0].Children)
	want := "FROMDS(DSN(MY.DATASET) VOL(VOL001) UNIT(SYSDA) NUMBER(1))"
	if !contains(names, want) {
		t.Errorf("expected %q among children, got %v", want, names)
	}
}

// TestBuildOperandSymbols_RangeMatchesActualText is a regression test for
// the second issue in LEPARM_BUG.md point 3: the Range.End for an operand
// symbol is reconstructed algebraically as
// child.Position.Length + len(paramValue) + 2, instead of being read from
// the actual parameter node's position. This drifts whenever the operand's
// source text isn't exactly "NAME(VALUE)" with no extra characters — e.g.
// when leading whitespace or multiple values are involved.
//
// This test uses a simple, well-behaved case (DISTLIB(D1), no LEPARM
// involved) to establish the expected/working baseline; it should PASS
// today. It exists as a guardrail so a future refactor of the Range
// computation can't silently break the common, simple case while fixing
// the LEPARM one.
func TestBuildOperandSymbols_RangeMatchesActualText(t *testing.T) {
	src := "++MOD(MYPROG)\n    DISTLIB(D1)\n."
	p := parser.NewParser(leparmLikeStatements())
	doc := p.Parse(src)
	lines := strings.Split(src, "\n")

	syms := buildSymbols(doc, lines, true, false)
	if len(syms) != 1 || len(syms[0].Children) != 1 {
		t.Fatalf("expected 1 statement with 1 child, got %d statements", len(syms))
	}

	child := syms[0].Children[0]
	if child.Range == nil {
		t.Fatal("expected Range to be set")
	}
	// "    DISTLIB(D1)" -> DISTLIB starts at char 4, "DISTLIB(D1)" is 11
	// chars long, so the range should end at char 4+11=15 on line 1.
	wantLine, wantChar := 1, 15
	if child.Range.End.Line != wantLine || child.Range.End.Character != wantChar {
		t.Errorf("expected range end line=%d char=%d, got line=%d char=%d",
			wantLine, wantChar, child.Range.End.Line, child.Range.End.Character)
	}
}
