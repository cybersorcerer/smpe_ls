package diagnostics

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// LKLIB names a ddname holding the module in load module format, so the module
// is not packaged inline. Per the ++MOD usage notes it is mutually exclusive
// with inline packaging, exactly like FROMDS, RELFILE and TXLIB.
func TestMissingInlineDataNotReportedForLklib(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++MOD(MYMOD) DISTLIB(AOS12) LKLIB(MYLKLIB) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	if !noDiagnosticWith(diags, "inline data") {
		t.Errorf("LKLIB supplies the module, no inline data should be demanded: %v", diags)
	}
}

// The same when another statement follows, which is the branch reporting
// "before next statement".
func TestMissingInlineDataNotReportedForLklibBeforeNextStatement(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++MOD(MYMOD1) DISTLIB(AOS12) LKLIB(MYLKLIB) .\n" +
		"++MOD(MYMOD2) DISTLIB(AOS12) TXLIB(MYTXLIB) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	if !noDiagnosticWith(diags, "inline data") {
		t.Errorf("Neither LKLIB nor TXLIB should demand inline data: %v", diags)
	}
}

// A ++MOD without any external source still has to carry inline data.
func TestMissingInlineDataStillReportedWithoutExternalSource(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++MOD(MYMOD) DISTLIB(AOS12) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	if !hasDiagnostic(diags, lsp.SeverityWarning, "inline data") {
		t.Errorf("Expected a warning for ++MOD without inline data: %v", diags)
	}
}

// The message lists the alternatives to inline packaging, so LKLIB belongs
// there for statements that accept it.
func TestMissingInlineDataMessageListsLklib(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++MOD(MYMOD) DISTLIB(AOS12) .\n"
	doc := p.Parse(input)

	for _, d := range dp.AnalyzeAST(doc) {
		if strings.Contains(d.Message, "inline data") {
			if !strings.Contains(d.Message, "LKLIB") {
				t.Errorf("Expected LKLIB among the alternatives, got: %s", d.Message)
			}
			return
		}
	}
	t.Fatal("No inline data diagnostic found")
}
