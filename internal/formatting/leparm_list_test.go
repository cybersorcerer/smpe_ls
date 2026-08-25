package formatting

import (
	"strings"
	"testing"
)

// TestFormatLeparmPreservesCommas is a regression test for a formatting bug
// found after fixing LEPARM completion/parsing (see LEPARM_BUG.md): LEPARM
// is a sub-operand container (its Values entries — AC, ALIGN2, AMODE, etc.
// — are comma-separated per the syntax diagram's comma loop in
// syntax_diagrams/leparm.png), but formatOperandIndented's sub-operand
// branch joined its children with a hardcoded space, discarding any commas
// the user wrote. SMP/E MCS list elements accept either blanks or commas as
// separators; the user's original choice of comma must be preserved.
func TestFormatLeparmPreservesCommas(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++MOD(MYPROG)\n    LEPARM(AC(1),ALIGN2,AMODE(24))\n    DISTLIB(D1)\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, ",") {
		t.Errorf("expected commas to be preserved between LEPARM values, got:\n%s", result)
	}
}

// TestFormatLeparmWrapsWhenManyValues verifies LEPARM is wrapped onto
// multiple lines once it exceeds WrapListsAfterN values, the same way
// PRE/SUP/other type="list" operands already are (see
// TestFormatListOperandIsWrapped). LEPARM itself is type="string" in
// smpe.json (it needs Values for completion, not list semantics), but its
// contents are still a list per the syntax diagram, so the sub-operand
// formatting path must apply the same wrap threshold.
func TestFormatLeparmWrapsWhenManyValues(t *testing.T) {
	p, fp := newTestFormatter(t)
	// 5 values, well above the test config's WrapListsAfterN=2.
	input := "++MOD(MYPROG)\n    LEPARM(AC(1),ALIGN2,AMODE(24),COMPAT(LKED),FETCHOPT(PACK))\n    DISTLIB(D1)\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	for _, v := range []string{"AC(1)", "ALIGN2", "AMODE(24)", "COMPAT(LKED)", "FETCHOPT(PACK)"} {
		if !strings.Contains(result, v) {
			t.Errorf("value %q missing from formatted output:\n%s", v, result)
		}
	}

	lines := strings.Split(result, "\n")
	acLine, fetchoptLine := -1, -1
	for i, line := range lines {
		if strings.Contains(line, "AC(1)") {
			acLine = i
		}
		if strings.Contains(line, "FETCHOPT(PACK)") {
			fetchoptLine = i
		}
	}
	if acLine == -1 || fetchoptLine == -1 {
		t.Fatalf("could not locate AC(1) and FETCHOPT(PACK) in output:\n%s", result)
	}
	if acLine == fetchoptLine {
		t.Errorf("expected AC(1) and FETCHOPT(PACK) on different lines (list wrapping), got both on line %d:\n%s", acLine, result)
	}
}

// TestFormatLeparmShortStaysSingleLine is a baseline/guardrail: a LEPARM
// with only 2 values (at or below WrapListsAfterN) must NOT be wrapped, and
// commas must still be preserved. The fix for the two bugs above must not
// force wrapping onto short LEPARM lists.
func TestFormatLeparmShortStaysSingleLine(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++MOD(MYPROG)\n    LEPARM(AC(1),ALIGN2)\n    DISTLIB(D1)\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, "LEPARM(AC(1),ALIGN2)") {
		t.Errorf("expected short LEPARM to stay on one line with commas preserved, got:\n%s", result)
	}
}

// TestFormatFromdsStaysSpaceSeparated is a baseline/guardrail: FROMDS's
// sub-operands (DSN, VOL, UNIT, NUMBER) are documented as space-separated
// in syntax_diagrams/mod_add_replace.png (no comma loop, unlike LEPARM).
// The fix for LEPARM's comma handling must not force commas into FROMDS.
func TestFormatFromdsStaysSpaceSeparated(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++MOD(MYPROG)\n    FROMDS(DSN(MY.DATASET) NUMBER(1))\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, "FROMDS(DSN(MY.DATASET) NUMBER(1))") {
		t.Errorf("expected FROMDS sub-operands to stay space-separated, got:\n%s", result)
	}
}

// TestFormatLeparmWrappedIsIdempotent verifies that formatting a wrapped,
// multi-line LEPARM a second time produces the same output. This caught a
// real bug during development: subOperandsCommaSeparated originally bailed
// out to a space-separated fallback whenever two sub-operands were on
// different lines, which is exactly the shape produced by the wrap branch
// itself — so a second format pass silently lost the commas and re-joined
// everything onto one line with spaces. Formatting must be a fixed point.
func TestFormatLeparmWrappedIsIdempotent(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++MOD(MYPROG)\n    LEPARM(AC(1),ALIGN2,AMODE(24),COMPAT(LKED),FETCHOPT(PACK))\n    DISTLIB(D1)\n.\n"

	r1 := formatOnce(t, p, fp, input)
	r2 := formatOnce(t, p, fp, r1)

	if r1 != r2 {
		t.Errorf("formatting is not idempotent:\nfirst pass:\n%s\nsecond pass:\n%s", r1, r2)
	}
	if !strings.Contains(r1, "AC(1),") {
		t.Errorf("expected wrapped LEPARM to keep trailing commas, got:\n%s", r1)
	}
}

// TestFormatLeparmSpaceSeparatedWraps is a regression test for a real
// user-reported bug: a space-separated (no comma) LEPARM with many values
// spanning two source lines was left completely untouched by the
// sub-operand formatting branch (which only wrapped when a comma was
// detected), producing a >72-column single-line string. That overlong
// line was then blindly cut by wrapLineAt72 at column 72 — landing the
// break mid-value instead of between sub-operands, with the original
// mid-statement line break still baked in. Wrapping at WrapListsAfterN
// must apply to space-separated sub-operand containers too — only the
// separator (space vs comma) is controlled by what the user wrote, not
// whether wrapping happens.
func TestFormatLeparmSpaceSeparatedWraps(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++MOD(MYMOD)\n    DISTLIB(D$DIST)\n    LEPARM(AC(1) ALIGN2 AMODE(31) COMPAT(PM1) NCAL FETCHOPT(PACK)\n    MAXBLK(0) PACK HOBSET NE)\n    LKLIB(LLIB)\n    LMOD(LMOD01)\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	for _, v := range []string{"AC(1)", "ALIGN2", "AMODE(31)", "COMPAT(PM1)", "NCAL", "FETCHOPT(PACK)", "MAXBLK(0)", "PACK", "HOBSET", "NE"} {
		if !strings.Contains(result, v) {
			t.Errorf("value %q missing from formatted output:\n%s", v, result)
		}
	}

	// No line may exceed 72 columns.
	for _, line := range strings.Split(result, "\n") {
		if len([]rune(line)) > 72 {
			t.Errorf("line exceeds 72 columns (%d): %q", len([]rune(line)), line)
		}
	}

	// Structural check: LEPARM must be wrapped as "LEPARM(" on its own
	// line, one value per following line, and ")" alone on the closing
	// line — the same shape TestFormatLeparmWrapsWhenManyValues expects
	// for the comma case, just with no trailing comma per line. This is
	// the check that actually distinguishes real wrapping from the
	// original mid-statement source line break being passed through
	// unchanged (which coincidentally also puts AC(1) and NE on different
	// lines, but with a ragged, non-wrapped shape and no per-item indent).
	lines := strings.Split(result, "\n")
	leparmOpenLine := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "LEPARM(" {
			leparmOpenLine = i
			break
		}
	}
	if leparmOpenLine == -1 {
		t.Fatalf("expected a line consisting solely of \"LEPARM(\" (wrapped list opening), got:\n%s", result)
	}

	wantValueLines := []string{"AC(1)", "ALIGN2", "AMODE(31)", "COMPAT(PM1)", "NCAL", "FETCHOPT(PACK)", "MAXBLK(0)", "PACK", "HOBSET", "NE"}
	for i, want := range wantValueLines {
		lineIdx := leparmOpenLine + 1 + i
		if lineIdx >= len(lines) {
			t.Fatalf("output ended before all wrapped LEPARM values were found; missing %q onward", want)
		}
		if strings.TrimSpace(lines[lineIdx]) != want {
			t.Errorf("expected wrapped value %q at line %d, got %q\nfull output:\n%s", want, lineIdx, lines[lineIdx], result)
		}
	}
	closeLineIdx := leparmOpenLine + 1 + len(wantValueLines)
	if closeLineIdx >= len(lines) || strings.TrimSpace(lines[closeLineIdx]) != ")" {
		t.Errorf("expected \")\" alone on the line after the wrapped values, got %q\nfull output:\n%s",
			lines[min(closeLineIdx, len(lines)-1)], result)
	}

	// Commas must NOT be introduced — the user wrote spaces.
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ",") {
			t.Errorf("unexpected comma introduced into space-separated LEPARM: %q", line)
		}
	}
}

// TestFormatFromdsWrapsWhenManyValues verifies the wrap-independent-of-
// separator fix also applies to FROMDS: with all 4 optional sub-operands
// present (DSN, VOL, UNIT, NUMBER — one more than WrapListsAfterN=2), it
// must wrap one value per line, space-separated (no comma introduced),
// matching FROMDS's documented space-separated syntax
// (syntax_diagrams/mod_add_replace.png has no comma loop for FROMDS).
func TestFormatFromdsWrapsWhenManyValues(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++MOD(MYMOD)\n    FROMDS(DSN(MY.DATASET) VOL(VOL001) UNIT(SYSDA) NUMBER(1))\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	lines := strings.Split(result, "\n")
	openLine := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "FROMDS(" {
			openLine = i
			break
		}
	}
	if openLine == -1 {
		t.Fatalf("expected a line consisting solely of \"FROMDS(\" (wrapped list opening), got:\n%s", result)
	}

	want := []string{"DSN(MY.DATASET)", "VOL(VOL001)", "UNIT(SYSDA)", "NUMBER(1)"}
	for i, w := range want {
		idx := openLine + 1 + i
		if idx >= len(lines) || strings.TrimSpace(lines[idx]) != w {
			t.Errorf("expected %q at line %d, got %q\nfull output:\n%s", w, idx, lines[min(idx, len(lines)-1)], result)
		}
	}

	for _, line := range lines {
		if strings.HasSuffix(strings.TrimSpace(line), ",") {
			t.Errorf("unexpected comma introduced into space-separated FROMDS: %q", line)
		}
	}
}
