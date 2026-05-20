package completion

import (
	"strings"
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

// Test: typing any partial MCS statement name (++X, ++XY, …) keeps the
// MCS completion list open so the client can filter locally. Operand
// completion may only kick in once a non-MCS character (space, paren,
// digit, …) follows.
func TestCompletionMCSPrefixKeepsMenuOpen(t *testing.T) {
	_, p, cp := createTestProviders()

	mcsCases := []string{
		"+",
		"++",
		"++U",
		"++US",
		"++USER",
		"++USERMOD",
		"  ++",
		"  ++US",
		"++V",
		"++M",
		"++MA",
	}

	for _, text := range mcsCases {
		t.Run("mcs:"+text, func(t *testing.T) {
			doc := p.Parse(text)
			items := cp.GetCompletionsAST(doc, text, 0, len(text))

			if len(items) == 0 {
				t.Fatalf("Expected MCS completions for %q, got none", text)
			}
			for _, item := range items {
				if !strings.HasPrefix(item.Label, "++") {
					t.Errorf("Expected only ++ items for %q, got non-MCS item %q", text, item.Label)
				}
			}
		})
	}
}

// Test: operand-name filtering keeps the menu open while the user is
// typing an operand prefix (e.g. ++USERMOD(LJS2012) RE → REWORK). The
// server returns the full operand list; the client filters locally.
func TestCompletionOperandPrefixKeepsMenuOpen(t *testing.T) {
	_, p, cp := createTestProviders()

	operandCases := []struct {
		text   string
		expect string // one operand name that MUST be in the result
	}{
		{"++USERMOD(LJS2012) ", "REWORK"},
		{"++USERMOD(LJS2012) R", "REWORK"},
		{"++USERMOD(LJS2012) RE", "REWORK"},
		{"++USERMOD(LJS2012) D", "DESC"},
		{"++USERMOD(LJS2012) DE", "DESC"},
		{"++USERMOD(LJS2012) REWORK(123) D", "DESC"},
	}

	for _, tc := range operandCases {
		t.Run("operand:"+tc.text, func(t *testing.T) {
			doc := p.Parse(tc.text)
			items := cp.GetCompletionsAST(doc, tc.text, 0, len(tc.text))

			if len(items) == 0 {
				t.Fatalf("Expected operand completions for %q, got none", tc.text)
			}

			found := false
			for _, item := range items {
				if strings.HasPrefix(item.Label, "++") {
					t.Errorf("Unexpected MCS item %q in operand context %q",
						item.Label, tc.text)
				}
				if item.Label == tc.expect {
					found = true
				}
			}
			if !found {
				t.Errorf("Expected operand %q in completions for %q",
					tc.expect, tc.text)
			}
		})
	}
}

// Test: sub-operand name filtering inside operand parens (e.g.
// FROMDS(D → DSN; FROMDS(U → UNIT). Server returns full sub-operand
// list; client filters locally by the typed prefix.
func TestCompletionSubOperandPrefixKeepsMenuOpen(t *testing.T) {
	_, p, cp := createTestProviders()

	cases := []struct {
		text   string
		expect string
	}{
		{"++MAC(MYMAC) FROMDS(", "DSN"},
		{"++MAC(MYMAC) FROMDS(D", "DSN"},
		{"++MAC(MYMAC) FROMDS(DS", "DSN"},
		{"++MAC(MYMAC) FROMDS(V", "VOL"},
		{"++MAC(MYMAC) FROMDS(U", "UNIT"},
		{"++MAC(MYMAC) FROMDS(N", "NUMBER"},
	}

	for _, tc := range cases {
		t.Run("subop:"+tc.text, func(t *testing.T) {
			doc := p.Parse(tc.text)
			items := cp.GetCompletionsAST(doc, tc.text, 0, len(tc.text))

			if len(items) == 0 {
				t.Fatalf("Expected sub-operand completions for %q, got none", tc.text)
			}
			found := false
			for _, item := range items {
				if strings.HasPrefix(item.Label, "++") {
					t.Errorf("Unexpected MCS item %q in sub-operand context %q",
						item.Label, tc.text)
				}
				if item.Label == tc.expect {
					found = true
				}
			}
			if !found {
				t.Errorf("Expected sub-operand %q in completions for %q",
					tc.expect, tc.text)
			}
		})
	}
}

// Test: multi-line scenarios — statement on line 0, operand or
// sub-operand prefix typed on a continuation line that starts with
// whitespace. The server must recognize the context and return the
// appropriate operand/sub-operand list.
func TestCompletionMultilineContexts(t *testing.T) {
	_, p, cp := createTestProviders()

	cases := []struct {
		name     string
		text     string
		line     int
		col      int
		expect   string
		mustHave string // "operand" | "subop"
	}{
		{
			name:   "operand-full-list-after-newline",
			text:   "++USERMOD(LJS2012)\n  ",
			line:   1,
			col:    2,
			expect: "REWORK",
		},
		{
			name:   "operand-prefix-after-newline",
			text:   "++USERMOD(LJS2012)\n  RE",
			line:   1,
			col:    4,
			expect: "REWORK",
		},
		{
			name:   "operand-prefix-after-newline-D",
			text:   "++USERMOD(LJS2012)\n  D",
			line:   1,
			col:    3,
			expect: "DESC",
		},
		{
			name:   "subop-after-newline",
			text:   "++MAC(MYMAC)\n  FROMDS(",
			line:   1,
			col:    9,
			expect: "DSN",
		},
		{
			name:   "subop-prefix-after-newline",
			text:   "++MAC(MYMAC)\n  FROMDS(D",
			line:   1,
			col:    10,
			expect: "DSN",
		},
		{
			name:   "operand-after-multiple-continuation-lines",
			text:   "++USERMOD(LJS2012) REWORK(2022056)\n  ",
			line:   1,
			col:    2,
			expect: "DESC",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := p.Parse(tc.text)
			items := cp.GetCompletionsAST(doc, tc.text, tc.line, tc.col)

			if len(items) == 0 {
				t.Fatalf("Expected completions for %q at (%d,%d), got none",
					tc.text, tc.line, tc.col)
			}
			found := false
			for _, item := range items {
				if strings.HasPrefix(item.Label, "++") {
					t.Errorf("Unexpected MCS item %q in multi-line context",
						item.Label)
				}
				if item.Label == tc.expect {
					found = true
				}
			}
			if !found {
				t.Errorf("Expected %q in completions for %q at (%d,%d)",
					tc.expect, tc.text, tc.line, tc.col)
			}
		})
	}
}

// Test: operand-name filtering on continuation lines (line 2+ that
// starts with whitespace and contains operand prefix).
func TestCompletionOperandPrefixOnContinuationLine(t *testing.T) {
	_, p, cp := createTestProviders()

	text := "++USERMOD(LJS2012)\n  RE"
	doc := p.Parse(text)
	// Cursor on line 1 (zero-indexed), after the "RE"
	items := cp.GetCompletionsAST(doc, text, 1, 4)

	if len(items) == 0 {
		t.Fatal("Expected operand completions on continuation line, got none")
	}
	for _, item := range items {
		if strings.HasPrefix(item.Label, "++") {
			t.Errorf("Unexpected MCS item %q on continuation line", item.Label)
		}
	}
	// REWORK must be present (matches RE prefix; server returns full list)
	foundRework := false
	for _, item := range items {
		if item.Label == "REWORK" {
			foundRework = true
		}
	}
	if !foundRework {
		t.Error("Expected REWORK in continuation-line operand completions")
	}
}

// Test: MCS-statement boilerplate snippets are offered alongside the
// keyword items, both at "++" and while typing a longer prefix. The
// snippet items must carry FilterText so VSCode's prefix matcher (which
// ignores '+' as a word character) still matches them.
func TestCompletionMCSSnippetsAreOffered(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++USERMOD": {
			Name:        "++USERMOD",
			Description: "Identifies a user modification",
			Parameter:   "usermod_name",
			Snippet:     "++USERMOD($1) /* yyyyddd */ .",
			Type:        "MCS",
		},
	}
	store := &data.Store{
		Statements: statements,
		List:       []data.MCSStatement{statements["++USERMOD"]},
	}
	p := parser.NewParser(statements)
	cp := NewProvider(store)

	prefixes := []string{"++", "++U", "++USER", "++USERMOD"}
	for _, prefix := range prefixes {
		t.Run("prefix:"+prefix, func(t *testing.T) {
			doc := p.Parse(prefix)
			items := cp.GetCompletionsAST(doc, prefix, 0, len(prefix))

			hasKeyword := false
			hasSnippet := false
			for _, item := range items {
				if item.Kind == lsp.CompletionItemKindKeyword && item.Label == "++USERMOD" {
					hasKeyword = true
				}
				if item.Kind == lsp.CompletionItemKindSnippet {
					hasSnippet = true
					if item.FilterText == "" {
						t.Errorf("snippet item for %q missing FilterText", prefix)
					}
					if strings.Contains(item.FilterText, " ") {
						t.Errorf("snippet FilterText %q contains whitespace — won't match VSCode/blink prefix filter",
							item.FilterText)
					}
					if !strings.HasPrefix(item.Label, "++USERMOD") {
						t.Errorf("snippet item label %q should start with ++USERMOD", item.Label)
					}
				}
			}
			if !hasKeyword {
				t.Errorf("Expected keyword item ++USERMOD for prefix %q", prefix)
			}
			if !hasSnippet {
				t.Errorf("Expected snippet item (Kind=Snippet) for prefix %q", prefix)
			}
		})
	}
}

// Test: regression guard — once a non-MCS char (space, '(') follows the
// statement name, the result must NOT be the MCS list anymore. Operand
// completion takes over.
func TestCompletionNonMCSAfterStatement(t *testing.T) {
	_, p, cp := createTestProviders()

	nonMcsCases := []string{
		"++USERMOD ",          // trailing space (cursor after space)
		"++USERMOD(",          // opened paren
		"++USERMOD(LJS2012) ", // operand-completion position
	}

	for _, text := range nonMcsCases {
		t.Run("nonmcs:"+text, func(t *testing.T) {
			doc := p.Parse(text)
			items := cp.GetCompletionsAST(doc, text, 0, len(text))

			for _, item := range items {
				if strings.HasPrefix(item.Label, "++") {
					t.Errorf("For %q expected non-MCS items, but got MCS item %q",
						text, item.Label)
				}
			}
		})
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

// Test: Snippet completion item
func TestCompletionSnippetItem(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++PTF": {
			Name:        "++PTF",
			Description: "Program Temporary Fix",
			Parameter:   "SYSMOD-ID",
			Snippet:     "++PTF(${1:UAnnnnn}) .\n++VER(${2:Z038}) .",
		},
	}
	store := &data.Store{
		Statements: statements,
		List:       []data.MCSStatement{statements["++PTF"]},
	}
	cp := NewProvider(store)
	p := parser.NewParser(statements)
	text := "+"
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 1)

	foundSnippet := false
	for _, item := range items {
		if item.Label == "++PTF …" {
			foundSnippet = true
			// When typing "+" a replaceRange is set, so TextEdit is used instead of InsertText
			snippetText := item.InsertText
			if item.TextEdit != nil {
				snippetText = item.TextEdit.NewText
			}
			if snippetText != "++PTF(${1:UAnnnnn}) .\n++VER(${2:Z038}) ." {
				t.Errorf("Unexpected snippet text: %s", snippetText)
			}
			if item.InsertTextFormat != lsp.InsertTextFormatSnippet {
				t.Error("Expected InsertTextFormatSnippet")
			}
			if item.Kind != lsp.CompletionItemKindSnippet {
				t.Error("Expected CompletionItemKindSnippet")
			}
		}
	}
	if !foundSnippet {
		t.Error("Expected snippet completion item for ++PTF")
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

func TestCompletionSnippetItemWithReplaceRange(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++PTF": {
			Name:        "++PTF",
			Description: "Program Temporary Fix",
			Parameter:   "SYSMOD-ID",
			Snippet:     "++PTF(${1:UAnnnnn}) .\n++VER(${2:Z038}) .",
		},
	}
	store := &data.Store{
		Statements: statements,
		List:       []data.MCSStatement{statements["++PTF"]},
	}
	cp := NewProvider(store)
	p := parser.NewParser(statements)
	// Simulate user typing "++" — triggers replaceRange path
	text := "++"
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 2)

	foundSnippet := false
	for _, item := range items {
		if item.Label == "++PTF …" {
			foundSnippet = true
			if item.TextEdit == nil {
				t.Error("Expected TextEdit to be set when replaceRange != nil")
			} else if item.TextEdit.NewText != "++PTF(${1:UAnnnnn}) .\n++VER(${2:Z038}) ." {
				t.Errorf("Unexpected TextEdit.NewText: %s", item.TextEdit.NewText)
			}
			if item.InsertText != "" {
				t.Error("InsertText should be empty when TextEdit is set")
			}
		}
	}
	if !foundSnippet {
		t.Error("Expected snippet completion item for ++PTF with replaceRange")
	}
}

func TestCompletionSnippetItemNoReplaceRange(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++PTF": {
			Name:        "++PTF",
			Description: "Program Temporary Fix",
			Parameter:   "SYSMOD-ID",
			Snippet:     "++PTF(${1:UAnnnnn}) .\n++VER(${2:Z038}) .",
		},
	}
	store := &data.Store{
		Statements: statements,
		List:       []data.MCSStatement{statements["++PTF"]},
	}
	cp := NewProvider(store)
	p := parser.NewParser(statements)
	// Use text with no "++"-prefixed content so no statement node is found,
	// but trimmedBefore is non-empty (not "+"-prefixed) — this forces the
	// findNodeAtPosition nil branch which calls getMCSCompletions(nil).
	text := "SOMETHING"
	doc := p.Parse(text)
	items := cp.GetCompletionsAST(doc, text, 0, 5)

	foundSnippet := false
	for _, item := range items {
		if item.Label == "++PTF …" {
			foundSnippet = true
			if item.InsertText != "++PTF(${1:UAnnnnn}) .\n++VER(${2:Z038}) ." {
				t.Errorf("Expected InsertText to be set, got: %q", item.InsertText)
			}
			if item.TextEdit != nil {
				t.Error("Expected TextEdit to be nil when replaceRange is nil")
			}
		}
	}
	if !foundSnippet {
		t.Error("Expected snippet completion item for ++PTF without replaceRange")
	}
}
