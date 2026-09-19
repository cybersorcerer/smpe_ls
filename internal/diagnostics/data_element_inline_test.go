package diagnostics

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Every statement that carries element data needs either inline data or one
// of FROMDS/RELFILE/TXLIB/LKLIB. The inline_data flag was missing for the
// Data Element MCS statements ++CLIST, ++DATA and ++DATA1 to ++DATA5 and for
// ++ZAP, ++PROGRAM, ++JAR and ++JARUPD, so none of the inline data handling
// applied to them.

// diagsWith runs the analysis with only the given config over text.
func diagsWith(t *testing.T, cfg *Config, text string) []lsp.Diagnostic {
	t.Helper()
	_, p, dp := loadRealStore(t)
	return dp.AnalyzeASTWithConfigAndText(p.Parse(text), cfg, text)
}

// Without inline data and without an external source the element has no
// content at all, which is what the diagnostic is for.
func TestDataElementMissingInlineDataReported(t *testing.T) {
	cfg := &Config{MissingInlineData: true}
	for _, name := range []string{"++CLIST", "++DATA", "++DATA1", "++DATA2", "++DATA3", "++DATA4", "++DATA5"} {
		text := name + "(E1) DISTLIB(ALIB) .\n++VER(Z038) FMID(F1) .\n"
		if diags := diagsWith(t, cfg, text); len(diags) != 1 {
			t.Errorf("%s: expected 1 missing-inline-data diagnostic, got %d: %+v", name, len(diags), diags)
		}
	}
}

// ++ZAP has no operand naming an external source at all - its IMASPZAP
// control statements can only follow inline. ++PROGRAM, ++JAR and ++JARUPD
// carry element data just like the data elements do.
func TestElementCarryingMCSMissingInlineDataReported(t *testing.T) {
	cfg := &Config{MissingInlineData: true}
	cases := []string{
		"++ZAP(MYMOD) DISTLIB(ALIB) .",
		"++PROGRAM(MYPGM) DISTLIB(ALIB) .",
		"++JAR(MYJAR) DISTLIB(ALIB) .",
		"++JARUPD(MYJAR) .",
	}
	for _, stmt := range cases {
		text := stmt + "\n++VER(Z038) FMID(F1) .\n"
		if diags := diagsWith(t, cfg, text); len(diags) != 1 {
			t.Errorf("%s: expected 1 missing-inline-data diagnostic, got %d: %+v", stmt, len(diags), diags)
		}
	}
}

// One of FROMDS, RELFILE, TXLIB or LKLIB replaces the inline data.
func TestDataElementExternalSourceSuppressesDiagnostic(t *testing.T) {
	cfg := &Config{MissingInlineData: true}
	for _, operand := range []string{"TXLIB(MYLIB)", "RELFILE(1)", "LKLIB(MYLIB)"} {
		text := "++CLIST(E1) DISTLIB(ALIB) " + operand + " .\n++VER(Z038) FMID(F1) .\n"
		if diags := diagsWith(t, cfg, text); len(diags) != 0 {
			t.Errorf("%s: element data comes from %s, no diagnostic expected: %+v", operand, operand, diags)
		}
	}

	for _, stmt := range []string{
		"++PROGRAM(MYPGM) DISTLIB(ALIB) LKLIB(MYLIB) .",
		"++JAR(MYJAR) DISTLIB(ALIB) TXLIB(MYLIB) .",
		"++JARUPD(MYJAR) RELFILE(1) .",
	} {
		text := stmt + "\n++VER(Z038) FMID(F1) .\n"
		if diags := diagsWith(t, cfg, text); len(diags) != 0 {
			t.Errorf("%s: element data comes from an external source: %+v", stmt, diags)
		}
	}
}

// Present inline data suppresses it just as well.
func TestDataElementWithInlineDataNotReported(t *testing.T) {
	cfg := &Config{MissingInlineData: true}
	text := "++DATA3(E1) DISTLIB(ALIB) .\nSOME DATA LINE\n++VER(Z038) FMID(F1) .\n"
	if diags := diagsWith(t, cfg, text); len(diags) != 0 {
		t.Errorf("Statement has inline data, no diagnostic expected: %+v", diags)
	}
}

// The column 1 check skips inline data. That exception only works when the
// statement is known to carry inline data.
func TestCommentInColumn1SkippedInDataElementInlineData(t *testing.T) {
	cfg := &Config{CommentInColumn1: true}
	text := "++CLIST(E1) DISTLIB(ALIB) .\n/* this is CLIST content, not a comment */\n++VER(Z038) FMID(F1) .\n"
	if diags := diagsWith(t, cfg, text); len(diags) != 0 {
		t.Errorf("Line belongs to the inline data and must not be reported: %+v", diags)
	}
}
