package hover

import (
	"strings"
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
				{
					Name:        "REWORK",
					Parameter:   "rework_id",
					Description: "Rework identifier",
					Type:        "string",
					Length:      7,
				},
				{
					Name:        "DESC",
					Parameter:   "description",
					Description: "Description of the modification",
				},
			},
		},
		"++APAR": {
			Name:        "++APAR",
			Description: "Identifies an APAR fix",
			Parameter:   "apar_id",
			Operands: []data.Operand{
				{
					Name:        "FILES",
					Parameter:   "number",
					Description: "Number of files",
					Type:        "integer",
				},
				{
					Name:        "RFDSNPFX",
					Parameter:   "prefix",
					Description: "Dataset name prefix",
					AllowedIf:   "FILES",
				},
			},
		},
		"++MAC": {
			Name:        "++MAC",
			Description: "Defines a macro",
			Parameter:   "member_name",
			Operands: []data.Operand{
				{
					Name:        "DISTLIB",
					Parameter:   "dataset_name",
					Description: "Distribution library",
				},
				{
					Name:               "DELETE",
					Description:        "Delete mode",
					Type:               "boolean",
					MutuallyExclusive:  "DISTLIB",
				},
				{
					Name:        "FROMDS",
					Parameter:   "DSN(dsname) VOL(volser) UNIT(unit) NUMBER(number)",
					Description: "Source dataset",
					Values: []data.AllowedValue{
						{Name: "DSN", Description: "Dataset name", Type: "string", Length: 44},
						{Name: "VOL", Description: "Volume serial", Type: "string", Length: 6},
						{Name: "UNIT", Description: "Unit type", Type: "string", Length: 8},
						{Name: "NUMBER", Description: "File number", Type: "integer"},
					},
				},
			},
		},
	}

	statementList := []data.MCSStatement{
		statements["++USERMOD"],
		statements["++APAR"],
		statements["++MAC"],
	}

	store := &data.Store{
		Statements: statements,
		List:       statementList,
	}

	p := parser.NewParser(statements)
	hp := NewProvider(store)

	return store, p, hp
}

// Test: Hover on ++USERMOD statement
func TestHoverOnStatement(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++USERMOD(UZ12345) REWORK(2024001)."
	doc := p.Parse(text)

	// Hover on "++USERMOD" (position 2, line 0)
	hover := hp.GetHoverAST(doc, 0, 2)

	if hover == nil {
		t.Fatal("Expected hover info for ++USERMOD, got nil")
	}

	content := hover.Contents.Value

	// Check for statement name
	if !strings.Contains(content, "++USERMOD") {
		t.Errorf("Expected hover to contain statement name '++USERMOD', got: %s", content)
	}

	// Check for description
	if !strings.Contains(content, "Identifies a user modification") {
		t.Errorf("Expected hover to contain description, got: %s", content)
	}

	// Check for parameter info
	if !strings.Contains(content, "usermod_name") {
		t.Errorf("Expected hover to contain parameter info, got: %s", content)
	}

	// Check for operands list
	if !strings.Contains(content, "REWORK") {
		t.Errorf("Expected hover to contain operand 'REWORK', got: %s", content)
	}

	t.Logf("Hover content:\n%s", content)
}

// Test: Hover on operand
func TestHoverOnOperand(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++USERMOD(UZ12345) REWORK(2024001)."
	doc := p.Parse(text)

	// Hover on "REWORK" (position 20, line 0)
	hover := hp.GetHoverAST(doc, 0, 20)

	if hover == nil {
		t.Fatal("Expected hover info for REWORK operand, got nil")
	}

	content := hover.Contents.Value

	// Check for operand name
	if !strings.Contains(content, "REWORK") {
		t.Errorf("Expected hover to contain operand name 'REWORK', got: %s", content)
	}

	// Check for description
	if !strings.Contains(content, "Rework identifier") {
		t.Errorf("Expected hover to contain operand description, got: %s", content)
	}

	// Check for type
	if !strings.Contains(content, "string") {
		t.Errorf("Expected hover to contain type 'string', got: %s", content)
	}

	// Check for length
	if !strings.Contains(content, "7") {
		t.Errorf("Expected hover to contain length '7', got: %s", content)
	}

	t.Logf("Hover content:\n%s", content)
}

// Test: Hover on operand with mutually_exclusive
func TestHoverOnOperandWithMutuallyExclusive(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++MAC(MYMAC) DELETE."
	doc := p.Parse(text)

	// Hover on "DELETE" (position 13, line 0)
	hover := hp.GetHoverAST(doc, 0, 13)

	if hover == nil {
		t.Fatal("Expected hover info for DELETE operand, got nil")
	}

	content := hover.Contents.Value

	// Check for mutually exclusive info
	if !strings.Contains(content, "Cannot be used with") {
		t.Errorf("Expected hover to contain 'Cannot be used with', got: %s", content)
	}

	if !strings.Contains(content, "DISTLIB") {
		t.Errorf("Expected hover to contain 'DISTLIB' as mutually exclusive, got: %s", content)
	}

	t.Logf("Hover content:\n%s", content)
}

// Test: Hover on operand with sub-operands (FROMDS)
func TestHoverOnOperandWithSubOperands(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++MAC(MYMAC) FROMDS(DSN(MY.LIB) VOL(DISK01))."
	doc := p.Parse(text)

	// Hover on "FROMDS" (position 13, line 0)
	hover := hp.GetHoverAST(doc, 0, 13)

	if hover == nil {
		t.Fatal("Expected hover info for FROMDS operand, got nil")
	}

	content := hover.Contents.Value

	// Check for operand name
	if !strings.Contains(content, "FROMDS") {
		t.Errorf("Expected hover to contain 'FROMDS', got: %s", content)
	}

	// Check for sub-operands
	if !strings.Contains(content, "Allowed Values") {
		t.Errorf("Expected hover to contain 'Allowed Values' section, got: %s", content)
	}

	if !strings.Contains(content, "DSN") {
		t.Errorf("Expected hover to contain sub-operand 'DSN', got: %s", content)
	}

	if !strings.Contains(content, "VOL") {
		t.Errorf("Expected hover to contain sub-operand 'VOL', got: %s", content)
	}

	if !strings.Contains(content, "Dataset name") {
		t.Errorf("Expected hover to contain 'Dataset name' description, got: %s", content)
	}

	t.Logf("Hover content:\n%s", content)
}

// Test: No hover on parameter value
func TestNoHoverOnParameterValue(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++USERMOD(UZ12345) REWORK(2024001)."
	doc := p.Parse(text)

	// Hover on parameter value "UZ12345" (position 11, line 0)
	hover := hp.GetHoverAST(doc, 0, 11)

	if hover != nil {
		t.Errorf("Expected no hover info for parameter value, got: %v", hover.Contents.Value)
	}

	t.Log("Correctly returns nil for parameter value")
}

// Test: No hover on empty position
func TestNoHoverOnEmptyPosition(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++USERMOD(UZ12345) REWORK(2024001)."
	doc := p.Parse(text)

	// Hover on whitespace (position 18, line 0)
	hover := hp.GetHoverAST(doc, 0, 18)

	if hover != nil {
		t.Errorf("Expected no hover info for whitespace, got: %v", hover.Contents.Value)
	}

	t.Log("Correctly returns nil for whitespace")
}

// Test: Hover on multiline statement
func TestHoverOnMultilineStatement(t *testing.T) {
	_, p, hp := createTestProviders()

	text := `++USERMOD(UZ12345)
          REWORK(2024001)
          DESC('Test modification').`
	doc := p.Parse(text)

	// Hover on "DESC" on line 2 (position 10, line 2)
	hover := hp.GetHoverAST(doc, 2, 10)

	if hover == nil {
		t.Fatal("Expected hover info for DESC operand in multiline statement, got nil")
	}

	content := hover.Contents.Value

	if !strings.Contains(content, "DESC") {
		t.Errorf("Expected hover to contain 'DESC', got: %s", content)
	}

	if !strings.Contains(content, "Description of the modification") {
		t.Errorf("Expected hover to contain operand description, got: %s", content)
	}

	t.Logf("Hover content:\n%s", content)
}

// Test: Hover on operand with allowed_if
func TestHoverOnOperandWithAllowedIf(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++APAR(UZ12345) FILES(100) RFDSNPFX(MY.PREFIX)."
	doc := p.Parse(text)

	// Hover on "RFDSNPFX" (position 27, line 0)
	hover := hp.GetHoverAST(doc, 0, 27)

	if hover == nil {
		t.Fatal("Expected hover info for RFDSNPFX operand, got nil")
	}

	content := hover.Contents.Value

	// Check for operand name
	if !strings.Contains(content, "RFDSNPFX") {
		t.Errorf("Expected hover to contain 'RFDSNPFX', got: %s", content)
	}

	// Check for description
	if !strings.Contains(content, "Dataset name prefix") {
		t.Errorf("Expected hover to contain description, got: %s", content)
	}

	t.Logf("Hover content:\n%s", content)
}

// Test: Hover returns markdown content
func TestHoverReturnsMarkdown(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++USERMOD(UZ12345) REWORK(2024001)."
	doc := p.Parse(text)

	hover := hp.GetHoverAST(doc, 0, 2)

	if hover == nil {
		t.Fatal("Expected hover info, got nil")
	}

	if hover.Contents.Kind != "markdown" {
		t.Errorf("Expected markdown content kind, got: %s", hover.Contents.Kind)
	}

	// Check for markdown formatting
	content := hover.Contents.Value
	if !strings.Contains(content, "**") {
		t.Errorf("Expected markdown bold formatting (**), got: %s", content)
	}

	t.Log("Hover correctly returns markdown content")
}

// Test: Hover on nil document
func TestHoverOnNilDocument(t *testing.T) {
	_, _, hp := createTestProviders()

	hover := hp.GetHoverAST(nil, 0, 0)

	if hover != nil {
		t.Errorf("Expected nil for nil document, got: %v", hover)
	}

	t.Log("Correctly handles nil document")
}

// Test: Hover on out-of-bounds position
func TestHoverOnOutOfBoundsPosition(t *testing.T) {
	_, p, hp := createTestProviders()

	text := "++USERMOD(UZ12345)."
	doc := p.Parse(text)

	// Hover beyond text length
	hover := hp.GetHoverAST(doc, 0, 100)

	if hover != nil {
		t.Errorf("Expected nil for out-of-bounds position, got: %v", hover)
	}

	t.Log("Correctly handles out-of-bounds position")
}
