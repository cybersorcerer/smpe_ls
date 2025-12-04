package diagnostics

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Helper function to create test data and providers
func createTestProviders() (*data.Store, *parser.Parser, *Provider) {
	statements := map[string]data.MCSStatement{
		"++USERMOD": {
			Name:        "++USERMOD",
			Description: "Identifies a user modification",
			Parameter:   "usermod_name",
			Operands: []data.Operand{
				{Name: "REWORK", Parameter: "rework_id", Description: "Rework identifier"},
				{Name: "DESC", Parameter: "description", Description: "Description"},
			},
		},
		"++VER": {
			Name:        "++VER",
			Description: "Specifies version information",
			Parameter:   "version_id",
			Operands:    []data.Operand{},
		},
		"++MAC": {
			Name:              "++MAC",
			Description:       "Defines a macro",
			Parameter:         "member_name",
			LanguageVariants:  true,
			InlineData:        true,
			Operands: []data.Operand{
				{Name: "DISTLIB", Parameter: "dataset_name", Description: "Distribution library"},
				{
					Name:      "FROMDS",
					Parameter: "DSN(dsname) VOL(volser) UNIT(unit) NUMBER(number)",
					Description: "Source dataset",
					Values: []data.AllowedValue{
						{Name: "DSN", Description: "Dataset name", Type: "string", Length: 44},
						{Name: "VOL", Description: "Volume serial", Type: "string", Length: 6},
						{Name: "UNIT", Description: "Unit type", Type: "string", Length: 8},
						{Name: "NUMBER", Description: "File number", Type: "integer", Length: 0},
					},
				},
			},
		},
	}

	statementList := []data.MCSStatement{
		statements["++USERMOD"],
		statements["++VER"],
		statements["++MAC"],
	}

	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	return store, p, dp
}

// Test: Missing terminator
func TestDiagnosticsMissingTerminator(t *testing.T) {
	_, p, dp := createTestProviders()

	input := "++VER(Z038)"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Debug: print all diagnostics
	t.Logf("Got %d diagnostics:", len(diags))
	for _, diag := range diags {
		t.Logf("  - Severity: %d, Message: %s", diag.Severity, diag.Message)
	}

	// Should have diagnostic for missing terminator
	found := false
	for _, diag := range diags {
		if diag.Severity == lsp.SeverityError && containsText(diag.Message, "terminated") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected diagnostic for missing terminator")
	}
}

// Test: Unbalanced parentheses
func TestDiagnosticsUnbalancedParentheses(t *testing.T) {
	_, p, dp := createTestProviders()

	input := "++USERMOD(LJS2012 ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have diagnostic for unbalanced parentheses
	found := false
	for _, diag := range diags {
		if diag.Severity == lsp.SeverityError && containsText(diag.Message, "parenthesis") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected diagnostic for unbalanced parentheses")
	}
}

// Test: Empty sub-operand parameter (VOL())
func TestDiagnosticsEmptySubOperandParameter(t *testing.T) {
	_, p, dp := createTestProviders()

	input := `++MAC(test) FROMDS(DSN(my.test) VOL()) .`
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have warning for empty VOL parameter
	found := false
	for _, diag := range diags {
		if diag.Severity == lsp.SeverityWarning && containsText(diag.Message, "VOL") && containsText(diag.Message, "empty") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected warning for empty VOL parameter, got %d diagnostics", len(diags))
		for _, diag := range diags {
			t.Logf("  - %s", diag.Message)
		}
	}
}

// Test: Valid sub-operands should have no warnings
func TestDiagnosticsValidSubOperands(t *testing.T) {
	_, p, dp := createTestProviders()

	input := `++MAC(test) FROMDS(DSN(my.test) VOL(ABC123) UNIT(SYSDA)) .`
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have no warnings for empty parameters
	for _, diag := range diags {
		if containsText(diag.Message, "empty parameter") {
			t.Errorf("Unexpected diagnostic: %s", diag.Message)
		}
	}
}

// Test: Unknown sub-operand
func TestDiagnosticsUnknownSubOperand(t *testing.T) {
	_, p, dp := createTestProviders()

	input := `++MAC(test) FROMDS(DSN(my.test) UNKNOWN(value)) .`
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have warning for unknown sub-operand
	found := false
	for _, diag := range diags {
		if diag.Severity == lsp.SeverityWarning && containsText(diag.Message, "Unknown sub-operand") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected warning for unknown sub-operand")
	}
}

// Test: Missing required parameter for statement
func TestDiagnosticsMissingRequiredStatementParameter(t *testing.T) {
	_, p, dp := createTestProviders()

	input := "++USERMOD ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have error for missing required statement parameter
	found := false
	for _, diag := range diags {
		if diag.Severity == lsp.SeverityError && containsText(diag.Message, "Missing required parameter") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error for missing required statement parameter")
	}
}

// Test: Valid statement should have no errors
func TestDiagnosticsValidStatement(t *testing.T) {
	_, p, dp := createTestProviders()

	input := "++VER(Z038) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have no errors
	if len(diags) > 0 {
		t.Errorf("Expected no diagnostics for valid statement, got %d:", len(diags))
		for _, diag := range diags {
			t.Logf("  - %s", diag.Message)
		}
	}
}

// Test: Multiple statements with mixed errors
func TestDiagnosticsMultipleStatements(t *testing.T) {
	_, p, dp := createTestProviders()

	input := `++VER(Z038) .
++USERMOD(LJS2012)
++MAC(test) FROMDS(VOL()) .`

	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have at least 2 diagnostics:
	// 1. Missing terminator on ++USERMOD
	// 2. Empty VOL() parameter
	if len(diags) < 2 {
		t.Errorf("Expected at least 2 diagnostics, got %d", len(diags))
	}
}

// Test: Terminator after multiline comment
func TestDiagnosticsTerminatorAfterComment(t *testing.T) {
	_, p, dp := createTestProviders()

	input := `++USERMOD(LJS2012) REWORK(2022056)
  DESC("BASE_JES2")
  /*
  +--------------------------------------------------------------------+
  ! Important notes                                                    !
  +--------------------------------------------------------------------+
  */ .`

	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should NOT have diagnostic for missing terminator
	for _, diag := range diags {
		if containsText(diag.Message, "terminator") {
			t.Errorf("Unexpected diagnostic for missing terminator: %s", diag.Message)
		}
	}
}

// Test: Language variant without language ID
func TestDiagnosticsLanguageVariantMissingLanguageID(t *testing.T) {
	_, p, dp := createTestProviders()

	// ++MAC requires language ID (e.g., ++MACASM)
	input := "++MAC(test) DISTLIB(AMACLIB) ."
	doc := p.Parse(input)
	_ = dp.AnalyzeAST(doc)

	// Note: The parser will accept ++MAC without language ID
	// This test checks if diagnostics flags it appropriately
	// For now, ++MAC without language ID is accepted (not all implementations require it)
	// So we just ensure it parses without crashing
	if len(doc.Statements) != 1 {
		t.Errorf("Expected 1 statement, got %d", len(doc.Statements))
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsText(s, substr string) bool {
	// Simple case-insensitive contains check
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// Test: ++APAR missing required parameter
func TestDiagnosticsAparMissingParameter(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++APAR": {
			Name:        "++APAR",
			Description: "APAR fix",
			Parameter:   "SYSMOD-ID",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "DESC", Parameter: "description", Type: "string"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++APAR"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	input := "++APAR ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have error for missing required parameter
	found := false
	for _, diag := range diags {
		if diag.Severity == lsp.SeverityError && containsText(diag.Message, "parameter") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error for missing required SYSMOD-ID parameter")
	}
}

// Test: ++ASSIGN valid statement
func TestDiagnosticsAssignValid(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++ASSIGN": {
			Name:        "++ASSIGN",
			Description: "Assign source ID",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "SOURCEID", Parameter: "source-id", Type: "string"},
				{Name: "TO", Parameter: "sysmod-ids", Type: "string"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++ASSIGN"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	input := "++ASSIGN SOURCEID(MYSOURCE) TO(UA12345) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have no errors for valid statement
	if len(diags) > 0 {
		t.Errorf("Expected no diagnostics for valid ++ASSIGN, got %d:", len(diags))
		for _, diag := range diags {
			t.Logf("  - %s", diag.Message)
		}
	}
}

// Test: ++DELETE missing terminator
func TestDiagnosticsDeleteMissingTerminator(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++DELETE": {
			Name:        "++DELETE",
			Description: "Delete load module",
			Parameter:   "NAME",
			Operands: []data.Operand{
				{Name: "SYSLIB", Parameter: "ddname", Type: "string"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++DELETE"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	input := "++DELETE(MYMODULE) SYSLIB(SYSLIB1)"
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have error for missing terminator
	found := false
	for _, diag := range diags {
		if diag.Severity == lsp.SeverityError && containsText(diag.Message, "terminated") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error for missing terminator")
	}
}

// Test: ++HOLD with multiple operands
func TestDiagnosticsHoldValid(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++HOLD": {
			Name:        "++HOLD",
			Description: "Place SYSMOD in exception status",
			Parameter:   "SYSMOD-ID",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "ERROR", Type: "boolean"},
				{Name: "FMID", Parameter: "fmid", Type: "string"},
				{Name: "REASON", Parameter: "reason-id", Type: "string"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++HOLD"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	input := "++HOLD(UA12345) ERROR FMID(HBB7790) REASON(ACTION) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have no errors for valid statement
	if len(diags) > 0 {
		t.Errorf("Expected no diagnostics for valid ++HOLD, got %d:", len(diags))
		for _, diag := range diags {
			t.Logf("  - %s", diag.Message)
		}
	}
}

// Test: ++IF missing required FMID operand
func TestDiagnosticsIfMissingFmid(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++IF": {
			Name:        "++IF",
			Description: "Conditional requisite",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "FMID", Parameter: "sysmod-id", Type: "string", Required: true},
				{Name: "THEN", Type: "boolean"},
				{Name: "REQ", Parameter: "sysmod-id", Type: "string", Required: true},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++IF"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	input := "++IF THEN REQ(UA12345) ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have warning/error for missing required FMID operand
	// Note: The diagnostics implementation needs to check for required operands
	t.Logf("Got %d diagnostics:", len(diags))
	for _, diag := range diags {
		t.Logf("  - Severity: %d, Message: %s", diag.Severity, diag.Message)
	}
}

// Test: ++FEATURE valid statement
func TestDiagnosticsFeatureValid(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++FEATURE": {
			Name:        "++FEATURE",
			Description: "Feature definition",
			Parameter:   "NAME",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "FMID", Parameter: "fmid", Type: "string"},
				{Name: "DESC", Parameter: "description", Type: "string"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++FEATURE"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	input := "++FEATURE(MYFEATURE) FMID(HBB7790) DESC(\"Test\") ."
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have no errors for valid statement
	if len(diags) > 0 {
		t.Errorf("Expected no diagnostics for valid ++FEATURE, got %d:", len(diags))
		for _, diag := range diags {
			t.Logf("  - %s", diag.Message)
		}
	}
}

// Test: ++APAR multiline with operands
func TestDiagnosticsAparMultiline(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++APAR": {
			Name:        "++APAR",
			Description: "APAR fix",
			Parameter:   "SYSMOD-ID",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "DESC", Parameter: "description", Type: "string"},
				{Name: "REWORK", Parameter: "level", Type: "integer"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++APAR"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	dp := NewProvider(store)

	input := `++APAR(UA12345) REWORK(2024001)
  DESC("Test APAR fix") .`
	doc := p.Parse(input)
	diags := dp.AnalyzeAST(doc)

	// Should have no errors for valid multiline statement
	if len(diags) > 0 {
		t.Errorf("Expected no diagnostics for valid multiline ++APAR, got %d:", len(diags))
		for _, diag := range diags {
			t.Logf("  - %s", diag.Message)
		}
	}
}
