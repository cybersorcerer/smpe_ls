package diagnostics

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// standaloneCommentDiags runs only the standalone-comment check over text.
func standaloneCommentDiags(t *testing.T, text string) []lsp.Diagnostic {
	t.Helper()
	_, p, dp := loadRealStore(t)
	doc := p.Parse(text)
	return dp.AnalyzeASTWithConfigAndText(doc, &Config{StandaloneCommentBetweenMCS: true}, text)
}

// A dot inside a multi-line block comment ended the statement, so a comment
// following the real terminator looked like it stood between two statements.
func TestStandaloneCommentNotReportedAfterDotInBlockComment(t *testing.T) {
	text := "++USERMOD(LBCP042) REWORK(2026259)\n" +
		"   DESC(CONVERT THE OPERLOG) .\n" +
		"++VER(Z038) FMID(HBB77E0)\n" +
		" /**\n" +
		"  * readable syslog lines.\n" +
		"  */\n" +
		" /*\n" +
		" +-- GITLAB-META-START ---\n" +
		" */\n" +
		" .\n" +
		"++JCLIN.\n"
	if diags := standaloneCommentDiags(t, text); len(diags) != 0 {
		t.Errorf("Comment belongs to the statement, not between statements: %+v", diags)
	}
}

// The same for a dot inside an operand value, which the terminator search
// must not mistake for the end of the statement either.
func TestStandaloneCommentNotReportedAfterDotInOperandValue(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"   DESC(R+V IIQ. SMF Exit 83)\n" +
		" /* a comment belonging to this statement */\n" +
		" .\n" +
		"++VER(Z038) FMID(F1) .\n"
	if diags := standaloneCommentDiags(t, text); len(diags) != 0 {
		t.Errorf("Comment belongs to the statement, not between statements: %+v", diags)
	}
}

// A comment that really does stand between two terminated statements is still
// reported - the check must not be defanged.
func TestStandaloneCommentStillReportedBetweenStatements(t *testing.T) {
	text := "++USERMOD(U1) REWORK(2026259) .\n" +
		" /* this one really is between statements */\n" +
		"++VER(Z038) FMID(F1) .\n"
	diags := standaloneCommentDiags(t, text)
	if len(diags) != 1 {
		t.Fatalf("Expected exactly 1 diagnostic, got %d: %+v", len(diags), diags)
	}
	if diags[0].Range.Start.Line != 1 {
		t.Errorf("Expected line 1, got %d", diags[0].Range.Start.Line)
	}
}
