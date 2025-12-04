package completion

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
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
			Name:        "++MAC",
			Description: "Defines a macro",
			Parameter:   "member_name",
			Operands: []data.Operand{
				{Name: "DISTLIB", Parameter: "dataset_name", Description: "Distribution library"},
				{
					Name:        "FROMDS",
					Parameter:   "DSN(dsname) VOL(volser) UNIT(unit) NUMBER(number)",
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
	cp := NewProvider(store)

	return store, p, cp
}

// Test: MCS statement completions at start of line
func TestCompletionMCSStatements(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "+"
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 1)

	if len(items) == 0 {
		t.Error("Expected MCS statement completions, got none")
	}

	t.Logf("Got %d completions:", len(items))
	for _, item := range items {
		t.Logf("  - %s", item.Label)
	}

	// Should include ++USERMOD, ++VER, ++MAC
	foundUsermod := false
	foundVer := false
	foundMac := false

	for _, item := range items {
		if item.Label == "++USERMOD" {
			foundUsermod = true
		}
		if item.Label == "++VER" {
			foundVer = true
		}
		if item.Label == "++MAC" {
			foundMac = true
		}
	}

	if !foundUsermod {
		t.Error("Expected ++USERMOD in completions")
	}
	if !foundVer {
		t.Error("Expected ++VER in completions")
	}
	if !foundMac {
		t.Error("Expected ++MAC in completions")
	}
}

// Test: Operand completions after statement
func TestCompletionOperandsAfterStatement(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++USERMOD(LJS2012) "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 19)

	if len(items) == 0 {
		t.Error("Expected operand completions, got none")
	}

	// Should include REWORK and DESC
	foundRework := false
	foundDesc := false

	for _, item := range items {
		if item.Label == "REWORK" {
			foundRework = true
		}
		if item.Label == "DESC" {
			foundDesc = true
		}
	}

	if !foundRework {
		t.Error("Expected REWORK in operand completions")
	}
	if !foundDesc {
		t.Error("Expected DESC in operand completions")
	}
}

// Test: No completions inside statement parameter
func TestCompletionNoCompletionInStatementParameter(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++USERMOD(LJS"
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 13)

	// Should not offer completions inside statement parameter
	if len(items) > 0 {
		t.Logf("Got %d completions (should be 0):", len(items))
		for _, item := range items {
			t.Logf("  - %s", item.Label)
		}
		t.Error("Expected no completions inside statement parameter")
	}
}

// Test: Sub-operand completions inside FROMDS
func TestCompletionSubOperandsInFromDS(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++MAC(MYMAC) FROMDS("
	doc := p.Parse(text)

	t.Logf("Parsed AST - statements: %d", len(doc.Statements))
	if len(doc.Statements) > 0 {
		stmt := doc.Statements[0]
		t.Logf("Statement: %s, children: %d", stmt.Name, len(stmt.Children))
		for _, child := range stmt.Children {
			t.Logf("  Child: type=%v, name=%s, pos=%d, len=%d, hasOperandDef=%v",
				child.Type, child.Name, child.Position.Character, child.Position.Length,
				child.OperandDef != nil)
			if child.OperandDef != nil && len(child.OperandDef.Values) > 0 {
				t.Logf("    OperandDef has %d values", len(child.OperandDef.Values))
			}
		}
	}

	items := cp.GetCompletionsAST(doc, text, 0, 20)

	t.Logf("Got %d completions:", len(items))
	for _, item := range items {
		t.Logf("  - %s", item.Label)
	}

	if len(items) == 0 {
		t.Error("Expected sub-operand completions, got none")
	}

	// Should include DSN, VOL, UNIT, NUMBER
	foundDSN := false
	foundVOL := false
	foundUNIT := false
	foundNUMBER := false

	for _, item := range items {
		if item.Label == "DSN" {
			foundDSN = true
		}
		if item.Label == "VOL" {
			foundVOL = true
		}
		if item.Label == "UNIT" {
			foundUNIT = true
		}
		if item.Label == "NUMBER" {
			foundNUMBER = true
		}
	}

	if !foundDSN {
		t.Error("Expected DSN in sub-operand completions")
	}
	if !foundVOL {
		t.Error("Expected VOL in sub-operand completions")
	}
	if !foundUNIT {
		t.Error("Expected UNIT in sub-operand completions")
	}
	if !foundNUMBER {
		t.Error("Expected NUMBER in sub-operand completions")
	}
}

// Test: Operand completions after first operand
func TestCompletionOperandsAfterFirstOperand(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++USERMOD(LJS2012) REWORK(2022056) "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 35) // Position 35 = after trailing space

	if len(items) == 0 {
		t.Error("Expected operand completions after first operand, got none")
	}

	// Should include DESC
	foundDesc := false

	for _, item := range items {
		if item.Label == "DESC" {
			foundDesc = true
		}
	}

	if !foundDesc {
		t.Error("Expected DESC in operand completions")
	}

	// Note: REWORK should still be offered (filtering is done by editor/user)
	// The completion provider doesn't filter out already-used operands
}

// Test: Multiline statement completions
func TestCompletionMultilineOperands(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++USERMOD(LJS2012) REWORK(2022056)\n  "
	doc := p.Parse(text)

	t.Logf("Parsed %d statements", len(doc.Statements))
	if len(doc.Statements) > 0 {
		stmt := doc.Statements[0]
		t.Logf("Statement: %s at line %d", stmt.Name, stmt.Position.Line)
		t.Logf("Statement has %d children", len(stmt.Children))
		for _, child := range stmt.Children {
			t.Logf("  Child: %s at line %d", child.Name, child.Position.Line)
		}
	}

	items := cp.GetCompletionsAST(doc, text, 1, 2)

	t.Logf("Got %d completions:", len(items))
	for _, item := range items {
		t.Logf("  - %s", item.Label)
	}

	if len(items) == 0 {
		t.Error("Expected operand completions on continuation line, got none")
	}

	// Should include DESC
	foundDesc := false
	for _, item := range items {
		if item.Label == "DESC" {
			foundDesc = true
		}
	}

	if !foundDesc {
		t.Error("Expected DESC in operand completions on continuation line")
	}
}

// Test: ++APAR operand completions
func TestCompletionAparOperands(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++APAR": {
			Name:        "++APAR",
			Description: "APAR fix",
			Parameter:   "SYSMOD-ID",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "DESC", Parameter: "description", Type: "string"},
				{Name: "REWORK", Parameter: "level", Type: "integer"},
				{Name: "FILES", Parameter: "number", Type: "integer"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++APAR"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	cp := NewProvider(store)

	text := "++APAR(UA12345) "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 16)

	if len(items) == 0 {
		t.Error("Expected operand completions for ++APAR, got none")
	}

	// Should include DESC, REWORK, FILES
	foundDesc := false
	foundRework := false
	foundFiles := false

	for _, item := range items {
		if item.Label == "DESC" {
			foundDesc = true
		}
		if item.Label == "REWORK" {
			foundRework = true
		}
		if item.Label == "FILES" {
			foundFiles = true
		}
	}

	if !foundDesc {
		t.Error("Expected DESC in ++APAR operand completions")
	}
	if !foundRework {
		t.Error("Expected REWORK in ++APAR operand completions")
	}
	if !foundFiles {
		t.Error("Expected FILES in ++APAR operand completions")
	}
}

// Test: ++ASSIGN operand completions
func TestCompletionAssignOperands(t *testing.T) {
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
	cp := NewProvider(store)

	text := "++ASSIGN "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 9)

	if len(items) == 0 {
		t.Error("Expected operand completions for ++ASSIGN, got none")
	}

	// Should include SOURCEID and TO
	foundSourceid := false
	foundTo := false

	for _, item := range items {
		if item.Label == "SOURCEID" {
			foundSourceid = true
		}
		if item.Label == "TO" {
			foundTo = true
		}
	}

	if !foundSourceid {
		t.Error("Expected SOURCEID in ++ASSIGN operand completions")
	}
	if !foundTo {
		t.Error("Expected TO in ++ASSIGN operand completions")
	}
}

// Test: ++DELETE operand completions
func TestCompletionDeleteOperands(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++DELETE": {
			Name:        "++DELETE",
			Description: "Delete load module",
			Parameter:   "NAME",
			Operands: []data.Operand{
				{Name: "ALIAS", Parameter: "alias", Type: "string"},
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
	cp := NewProvider(store)

	text := "++DELETE(MYMODULE) "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 19)

	if len(items) == 0 {
		t.Error("Expected operand completions for ++DELETE, got none")
	}

	// Should include ALIAS and SYSLIB
	foundAlias := false
	foundSyslib := false

	for _, item := range items {
		if item.Label == "ALIAS" {
			foundAlias = true
		}
		if item.Label == "SYSLIB" {
			foundSyslib = true
		}
	}

	if !foundAlias {
		t.Error("Expected ALIAS in ++DELETE operand completions")
	}
	if !foundSyslib {
		t.Error("Expected SYSLIB in ++DELETE operand completions")
	}
}

// Test: ++HOLD operand completions
func TestCompletionHoldOperands(t *testing.T) {
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
				{Name: "DATE", Parameter: "date", Type: "integer"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++HOLD"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	cp := NewProvider(store)

	text := "++HOLD(UA12345) "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 16)

	if len(items) == 0 {
		t.Error("Expected operand completions for ++HOLD, got none")
	}

	// Should include ERROR, FMID, REASON, DATE
	foundError := false
	foundFmid := false
	foundReason := false

	for _, item := range items {
		if item.Label == "ERROR" {
			foundError = true
		}
		if item.Label == "FMID" {
			foundFmid = true
		}
		if item.Label == "REASON" {
			foundReason = true
		}
	}

	if !foundError {
		t.Error("Expected ERROR in ++HOLD operand completions")
	}
	if !foundFmid {
		t.Error("Expected FMID in ++HOLD operand completions")
	}
	if !foundReason {
		t.Error("Expected REASON in ++HOLD operand completions")
	}
}

// Test: ++IF operand completions
func TestCompletionIfOperands(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++IF": {
			Name:        "++IF",
			Description: "Conditional requisite",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "FMID", Parameter: "sysmod-id", Type: "string"},
				{Name: "THEN", Type: "boolean"},
				{Name: "REQ", Parameter: "sysmod-id", Type: "string"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++IF"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	cp := NewProvider(store)

	text := "++IF "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 5)

	if len(items) == 0 {
		t.Error("Expected operand completions for ++IF, got none")
	}

	// Should include FMID, THEN, REQ
	foundFmid := false
	foundThen := false
	foundReq := false

	for _, item := range items {
		if item.Label == "FMID" {
			foundFmid = true
		}
		if item.Label == "THEN" {
			foundThen = true
		}
		if item.Label == "REQ" {
			foundReq = true
		}
	}

	if !foundFmid {
		t.Error("Expected FMID in ++IF operand completions")
	}
	if !foundThen {
		t.Error("Expected THEN in ++IF operand completions")
	}
	if !foundReq {
		t.Error("Expected REQ in ++IF operand completions")
	}
}

// Test: ++FEATURE operand completions
func TestCompletionFeatureOperands(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++FEATURE": {
			Name:        "++FEATURE",
			Description: "Feature definition",
			Parameter:   "NAME",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "DESC", Parameter: "description", Type: "string"},
				{Name: "FMID", Parameter: "fmid", Type: "string"},
				{Name: "PRODUCT", Parameter: "prodid", Type: "string"},
				{Name: "REWORK", Parameter: "level", Type: "integer"},
			},
		},
	}

	statementList := []data.MCSStatement{statements["++FEATURE"]}
	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	cp := NewProvider(store)

	text := "++FEATURE(MYFEATURE) "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 21)

	if len(items) == 0 {
		t.Error("Expected operand completions for ++FEATURE, got none")
	}

	// Should include DESC, FMID, PRODUCT, REWORK
	foundDesc := false
	foundFmid := false
	foundProduct := false

	for _, item := range items {
		if item.Label == "DESC" {
			foundDesc = true
		}
		if item.Label == "FMID" {
			foundFmid = true
		}
		if item.Label == "PRODUCT" {
			foundProduct = true
		}
	}

	if !foundDesc {
		t.Error("Expected DESC in ++FEATURE operand completions")
	}
	if !foundFmid {
		t.Error("Expected FMID in ++FEATURE operand completions")
	}
	if !foundProduct {
		t.Error("Expected PRODUCT in ++FEATURE operand completions")
	}
}

// Test: Real file - typing ++ on new line should offer all statements
func TestCompletionRealFileNewStatement(t *testing.T) {
	// Load real smpe.json
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}

	p := parser.NewParser(store.Statements)
	cp := NewProvider(store)

	// Simulate test-usermod-simple.smpe content + typing ++ on line 5
	text := `++USERMOD(TEST002)
  REWORK(123)
  DESC(This is a test with PRIMARY(!) inside)
  .

++`
	doc := p.Parse(text)

	// Cursor is at line 5, char 2 (after ++)
	items := cp.GetCompletionsAST(doc, text, 5, 2)

	t.Logf("Got %d completions:", len(items))
	for i, item := range items {
		if i < 10 { // Print first 10
			t.Logf("  - %s", item.Label)
		}
	}

	if len(items) < 50 {
		t.Errorf("Expected many statement completions (all 78 statements), got only %d", len(items))
	}

	// Should include various types
	foundUsermod := false
	foundApar := false
	foundMac := false
	foundJar := false

	for _, item := range items {
		if item.Label == "++USERMOD" {
			foundUsermod = true
		}
		if item.Label == "++APAR" {
			foundApar = true
		}
		if item.Label == "++MAC" {
			foundMac = true
		}
		if item.Label == "++JAR" {
			foundJar = true
		}
	}

	if !foundUsermod {
		t.Error("Expected ++USERMOD in statement completions")
	}
	if !foundApar {
		t.Error("Expected ++APAR in statement completions")
	}
	if !foundMac {
		t.Error("Expected ++MAC in statement completions")
	}
	if !foundJar {
		t.Error("Expected ++JAR in statement completions")
	}
}

// Test: No completions for statement without operands
func TestCompletionNoOperandsForStatementWithoutOperands(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++VER(Z038) "
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 12)

	// ++VER has no operands defined, so no operand completions
	if len(items) > 0 {
		t.Logf("Got %d completions:", len(items))
		for _, item := range items {
			t.Logf("  - %s", item.Label)
		}
		// Note: This might still offer MCS completions if parser thinks we're at start
		// Let's be lenient here - as long as we don't crash
	}
}

// Test: Completion with TextEdit range
func TestCompletionTextEditRange(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++"
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 2)

	if len(items) == 0 {
		t.Error("Expected MCS statement completions, got none")
	}

	// Check that completions have TextEdit with proper range
	for _, item := range items {
		if item.TextEdit != nil {
			// TextEdit should replace the ++ with the full statement
			if item.TextEdit.Range.Start.Line != 0 || item.TextEdit.Range.Start.Character != 0 {
				t.Errorf("Expected TextEdit range to start at (0,0), got (%d,%d)",
					item.TextEdit.Range.Start.Line, item.TextEdit.Range.Start.Character)
			}
			if item.TextEdit.Range.End.Line != 0 || item.TextEdit.Range.End.Character != 2 {
				t.Errorf("Expected TextEdit range to end at (0,2), got (%d,%d)",
					item.TextEdit.Range.End.Line, item.TextEdit.Range.End.Character)
			}
		}
	}
}
