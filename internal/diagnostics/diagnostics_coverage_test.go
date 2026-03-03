package diagnostics

// Tests for diagnostic types not covered by diagnostics_test.go:
// DuplicateOperand, MissingRequiredOperand, DependencyViolation,
// MutuallyExclusive, RequiredGroup, ContentBeyondColumn72,
// StandaloneCommentBetweenMCS, MissingInlineData, UnknownStatement

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

const realSMPEJSON = "../../data/smpe.json"

// loadRealStore loads the real smpe.json for tests that need actual definitions
func loadRealStore(t *testing.T) (*data.Store, *parser.Parser, *Provider) {
	t.Helper()
	store, err := data.Load(realSMPEJSON)
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := parser.NewParser(store.Statements)
	dp := NewProvider(store)
	return store, p, dp
}

func hasDiagnostic(diags []lsp.Diagnostic, severity int, substr string) bool {
	for _, d := range diags {
		if d.Severity == severity && containsText(d.Message, substr) {
			return true
		}
	}
	return false
}

func noDiagnosticWith(diags []lsp.Diagnostic, substr string) bool {
	for _, d := range diags {
		if containsText(d.Message, substr) {
			return false
		}
	}
	return true
}

// --- DuplicateOperand ---

func TestDiagnosticsDuplicateOperand(t *testing.T) {
	_, p, dp := createTestProviders()
	input := "++USERMOD(LJS2012) REWORK(2022056) REWORK(2022099) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	if !hasDiagnostic(diags, lsp.SeverityHint, "Duplicate") {
		t.Errorf("Expected hint for duplicate operand REWORK, got: %v", diags)
	}
}

func TestDiagnosticsNoDuplicateOnDistinctOperands(t *testing.T) {
	_, p, dp := createTestProviders()
	input := "++USERMOD(LJS2012) REWORK(2022056) DESC(my-desc) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	if !noDiagnosticWith(diags, "Duplicate") {
		t.Error("Unexpected duplicate operand diagnostic for distinct operands")
	}
}

// --- MissingRequiredOperand ---

func TestDiagnosticsMissingRequiredOperand_ASSIGN(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// ++ASSIGN requires SOURCEID and TO
	input := "++ASSIGN TO(UA12345) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	// MissingRequiredOperand is reported as Warning
	if !hasDiagnostic(diags, lsp.SeverityWarning, "SOURCEID") {
		t.Error("Expected warning for missing required operand SOURCEID")
	}
}

func TestDiagnosticsMissingRequiredOperand_IF(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// ++IF requires FMID and REQ
	input := "++IF REQ(UA12345) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	// MissingRequiredOperand is reported as Warning
	if !hasDiagnostic(diags, lsp.SeverityWarning, "FMID") {
		t.Error("Expected warning for missing required operand FMID in ++IF")
	}
}

// --- MutuallyExclusive ---

func TestDiagnosticsMutuallyExclusive(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// ++RELEASE: ERROR and USER are mutually exclusive
	input := "++RELEASE(UZ12345) ERROR USER FMID(FXY1040) REASON(CPU0A) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	if !hasDiagnostic(diags, lsp.SeverityError, "mutually exclusive") {
		t.Error("Expected error for mutually exclusive operands ERROR and USER")
	}
}

func TestDiagnosticsNoMutuallyExclusiveViolation(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// Only one of the mutually exclusive group
	input := "++RELEASE(UZ12345) USER FMID(FXY1040) REASON(CPU0A) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	if !noDiagnosticWith(diags, "mutually exclusive") {
		t.Errorf("Unexpected mutually exclusive diagnostic: %v", diags)
	}
}

// --- RequiredGroup ---

func TestDiagnosticsRequiredGroup_RELEASE(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// ++RELEASE requires one of ERROR|FIXCAT|SYSTEM|USER
	input := "++RELEASE(UZ12345) FMID(FXY1040) REASON(CPU0A) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	if !hasDiagnostic(diags, lsp.SeverityError, "One of") {
		t.Error("Expected error for missing required group (ERROR|FIXCAT|SYSTEM|USER)")
	}
}

func TestDiagnosticsRequiredGroup_Satisfied(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++RELEASE(UZ12345) USER FMID(FXY1040) REASON(CPU0A) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	if !noDiagnosticWith(diags, "One of") {
		t.Errorf("Unexpected required group diagnostic when group is satisfied: %v", diags)
	}
}

// --- DependencyViolation ---

func TestDiagnosticsDependencyViolation(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// ++APAR: RFDSNPFX is allowed_if=FILES — using RFDSNPFX without FILES should trigger
	input := "++APAR(UA12345) RFDSNPFX(MYPREFIX) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	if !hasDiagnostic(diags, lsp.SeverityInformation, "requires") {
		t.Error("Expected dependency violation: RFDSNPFX requires FILES")
	}
}

// --- ContentBeyondColumn72 ---

func TestDiagnosticsContentBeyondColumn72(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// Build a line that exceeds 72 characters
	input := "++VER(Z038) FMID(HBB7790 HBB7791 HBB7792 HBB7793 HBB7794 HBB7795 HBB7796) .\n"
	doc := p.Parse(input)
	text := input
	diags := dp.AnalyzeASTWithConfigAndText(doc, &Config{ContentBeyondColumn72: true}, text)
	t.Logf("Diagnostics: %v", diags)

	if !hasDiagnostic(diags, lsp.SeverityError, "column 72") {
		t.Error("Expected error for content beyond column 72")
	}
}

func TestDiagnosticsNoContentBeyondColumn72_ShortLine(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++VER(Z038) FMID(HBB7790) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeASTWithConfigAndText(doc, &Config{ContentBeyondColumn72: true}, input)

	if !noDiagnosticWith(diags, "column 72") {
		t.Errorf("Unexpected column 72 diagnostic for short line: %v", diags)
	}
}

// --- StandaloneCommentBetweenMCS ---

func TestDiagnosticsStandaloneCommentBetweenMCS(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++VER(Z038) .\n/* standalone comment between statements */\n++VER(Z039) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeASTWithConfigAndText(doc, &Config{StandaloneCommentBetweenMCS: true}, input)
	t.Logf("Diagnostics: %v", diags)

	if !hasDiagnostic(diags, lsp.SeverityError, "comment") {
		t.Error("Expected error for standalone comment between MCS statements")
	}
}

func TestDiagnosticsNoStandaloneComment_InlineIsOK(t *testing.T) {
	_, p, dp := loadRealStore(t)
	// Inline comment on same line as statement — not standalone
	input := "++VER(Z038) /* inline comment */ .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeASTWithConfigAndText(doc, &Config{StandaloneCommentBetweenMCS: true}, input)

	if !noDiagnosticWith(diags, "standalone") {
		t.Errorf("Unexpected standalone comment diagnostic for inline comment: %v", diags)
	}
}

// --- MissingInlineData ---

func TestDiagnosticsMissingInlineData(t *testing.T) {
	_, p, dp := loadRealStore(t)

	// Two consecutive ++MAC statements with no inline data between them
	input := "++MAC(MAC1) DISTLIB(AMACLIB) .\n++MAC(MAC2) DISTLIB(AMACLIB) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	if !hasDiagnostic(diags, lsp.SeverityWarning, "inline data") {
		t.Error("Expected warning for missing inline data on ++MAC")
	}
}

// --- UnknownStatement ---

func TestDiagnosticsUnknownStatement(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++UNKNOWNXYZ(PARAM) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)
	t.Logf("Diagnostics: %v", diags)

	if !hasDiagnostic(diags, lsp.SeverityError, "Unknown statement") {
		t.Error("Expected error for unknown statement ++UNKNOWNXYZ")
	}
}

func TestDiagnosticsNoUnknownStatement_KnownStmt(t *testing.T) {
	_, p, dp := loadRealStore(t)
	input := "++PTF(UA12345) .\n"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	if !noDiagnosticWith(diags, "Unknown statement") {
		t.Errorf("Unexpected unknown statement diagnostic for ++PTF: %v", diags)
	}
}
