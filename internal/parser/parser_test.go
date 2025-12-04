package parser

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
)

func TestParseUsermod(t *testing.T) {
	// Create test data matching smpe.json structure
	statements := map[string]data.MCSStatement{
		"++USERMOD": {
			Name:        "++USERMOD",
			Description: "Define a USERMOD",
			Parameter:   "sysmod_id",
			Type:        "sysmod",
			Operands: []data.Operand{
				{Name: "DESC", Parameter: "description", Type: "string"},
				{Name: "REWORK", Parameter: "date", Type: "date"},
			},
		},
	}

	parser := NewParser(statements)

	// Test simple statement with parameter
	text := "++USERMOD(LJS2012) REWORK(2022056)"
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]
	if stmt.Name != "++USERMOD" {
		t.Errorf("Expected statement name ++USERMOD, got %s", stmt.Name)
	}

	// Check statement parameter (LJS2012)
	if len(stmt.Children) < 1 {
		t.Fatalf("Expected at least 1 child (statement parameter), got %d", len(stmt.Children))
	}

	stmtParam := stmt.Children[0]
	if stmtParam.Type != NodeTypeParameter {
		t.Errorf("Expected first child to be NodeTypeParameter, got %v", stmtParam.Type)
	}
	if stmtParam.Value != "LJS2012" {
		t.Errorf("Expected statement parameter value LJS2012, got %s", stmtParam.Value)
	}

	// Check REWORK operand
	reworkFound := false
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand && child.Name == "REWORK" {
			reworkFound = true
			if len(child.Children) != 1 {
				t.Errorf("Expected REWORK to have 1 parameter child, got %d", len(child.Children))
			} else {
				param := child.Children[0]
				if param.Value != "2022056" {
					t.Errorf("Expected REWORK parameter value 2022056, got %s", param.Value)
				}
			}
			break
		}
	}

	if !reworkFound {
		t.Error("REWORK operand not found")
	}
}

func TestParseFromDS(t *testing.T) {
	// Create test data with sub-operands
	statements := map[string]data.MCSStatement{
		"++JAR": {
			Name:        "++JAR",
			Description: "Define a JAR file",
			Parameter:   "jar_name",
			Type:        "data_element",
			Operands: []data.Operand{
				{
					Name:      "FROMDS",
					Parameter: "DSN(dsname) VOLUME(volser) UNIT(unit) NUMBER(number)",
					Type:      "composite",
					Values: []data.AllowedValue{
						{Name: "DSN", Description: "Dataset name"},
						{Name: "VOLUME", Description: "Volume serial"},
						{Name: "UNIT", Description: "Unit type"},
						{Name: "NUMBER", Description: "Sequence number"},
					},
				},
			},
		},
	}

	parser := NewParser(statements)

	// Test FROMDS with sub-operands
	text := "++JAR(MYJAR) FROMDS(DSN(MY.DATASET) VOLUME(VOL001))"
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]

	// Find FROMDS operand
	fromdsFound := false
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand && child.Name == "FROMDS" {
			fromdsFound = true

			// FROMDS should have sub-operands (DSN, VOLUME), not a simple parameter
			if len(child.Children) < 2 {
				t.Errorf("Expected FROMDS to have at least 2 sub-operands, got %d", len(child.Children))
			}

			// Check for DSN sub-operand
			dsnFound := false
			volumeFound := false
			for _, subOp := range child.Children {
				if subOp.Name == "DSN" {
					dsnFound = true
					if len(subOp.Children) != 1 {
						t.Errorf("Expected DSN to have 1 parameter, got %d", len(subOp.Children))
					} else if subOp.Children[0].Value != "MY.DATASET" {
						t.Errorf("Expected DSN parameter MY.DATASET, got %s", subOp.Children[0].Value)
					}
				}
				if subOp.Name == "VOLUME" {
					volumeFound = true
					if len(subOp.Children) != 1 {
						t.Errorf("Expected VOLUME to have 1 parameter, got %d", len(subOp.Children))
					} else if subOp.Children[0].Value != "VOL001" {
						t.Errorf("Expected VOLUME parameter VOL001, got %s", subOp.Children[0].Value)
					}
				}
			}

			if !dsnFound {
				t.Error("DSN sub-operand not found in FROMDS")
			}
			if !volumeFound {
				t.Error("VOLUME sub-operand not found in FROMDS")
			}

			break
		}
	}

	if !fromdsFound {
		t.Error("FROMDS operand not found")
	}
}

func TestParseMultiline(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++USERMOD": {
			Name:        "++USERMOD",
			Description: "Define a USERMOD",
			Parameter:   "sysmod_id",
			Type:        "sysmod",
			Operands: []data.Operand{
				{Name: "DESC", Parameter: "description", Type: "string"},
				{Name: "REWORK", Parameter: "date", Type: "date"},
			},
		},
	}

	parser := NewParser(statements)

	// Test multiline statement
	text := `++USERMOD(LJS2012) REWORK(2022056)
  DESC("Test description")`

	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]

	// Should have statement parameter + REWORK + DESC
	operandCount := 0
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand {
			operandCount++
		}
	}

	if operandCount != 2 {
		t.Errorf("Expected 2 operands (REWORK, DESC), got %d", operandCount)
	}
}

func TestParseInlineData(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++JCLIN": {
			Name:        "++JCLIN",
			Description: "JCL inline data",
			Type:        "inline",
			InlineData:  true,
		},
		"++VER": {
			Name:        "++VER",
			Description: "Version statement",
			Type:        "control",
			InlineData:  false,
		},
	}

	parser := NewParser(statements)

	// Test inline data - JCL lines should be skipped, but next statement should be parsed
	text := `++JCLIN .
//LINK EXEC LINKS
//SYSLIN DD *
  This is JCL data
/*
++VER(Z038)`

	doc := parser.Parse(text)

	// Should parse ++JCLIN and ++VER, skipping JCL lines in between
	if len(doc.Statements) != 2 {
		t.Fatalf("Expected 2 statements (++JCLIN, ++VER), got %d", len(doc.Statements))
	}

	if doc.Statements[0].Name != "++JCLIN" {
		t.Errorf("Expected first statement to be ++JCLIN, got %s", doc.Statements[0].Name)
	}

	if doc.Statements[1].Name != "++VER" {
		t.Errorf("Expected second statement to be ++VER, got %s", doc.Statements[1].Name)
	}
}

func TestParseCommentsAfterBlockComment(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++MAC": {
			Name:        "++MAC",
			Description: "Macro statement",
			Type:        "control",
			InlineData:  true,
			Parameter:   "macro-name",
			Operands: []data.Operand{
				{Name: "RELFILE", Parameter: "file_id", Description: "Relative file number"},
				{Name: "TXLIB", Parameter: "library_name", Description: "Text library"},
			},
		},
	}

	parser := NewParser(statements)

	// Test that comments within statements are correctly parsed
	// Comments only count when they're part of a statement region
	text := `++MAC(MYMAC) RELFILE(1) /* Block comment
line 2
line 3
*/ .

++MAC(MYMAC2) /* Single line comment */ TXLIB(ATXLIB) .`

	// Debug: print the text
	t.Logf("Input text:\n%s", text)
	t.Logf("Text length: %d", len(text))

	doc := parser.Parse(text)

	// Debug output
	t.Logf("Got %d statements:", len(doc.Statements))
	for i, stmt := range doc.Statements {
		t.Logf("  [%d] %s at line %d", i, stmt.Name, stmt.Position.Line)
	}
	t.Logf("Got %d comments:", len(doc.Comments))
	for i, comment := range doc.Comments {
		t.Logf("  [%d] at line %d, char %d, len %d", i, comment.Position.Line, comment.Position.Character, comment.Position.Length)
	}

	// Should have 2 comments (both within statement regions)
	if len(doc.Comments) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(doc.Comments))
	}

	// First comment should be block comment
	if doc.Comments[0].Type != NodeTypeComment {
		t.Errorf("Expected first node to be a comment")
	}

	// Second comment should be single-line comment
	if doc.Comments[1].Type != NodeTypeComment {
		t.Errorf("Expected second node to be a comment")
	}

	// Both statements should be parsed
	if len(doc.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(doc.Statements))
	}
}

func TestParseInlineDataWithMacro(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++MAC": {
			Name:        "++MAC",
			Description: "Macro statement",
			Type:        "control",
			InlineData:  true,
			Parameter:   "macro-name",
		},
		"++VER": {
			Name:        "++VER",
			Description: "Version statement",
			Type:        "control",
			InlineData:  false,
		},
	}

	parser := NewParser(statements)

	// Test with correct inline data (macro source following the statement)
	text := `++MAC(MYMAC) DISTLIB(AMACLIB) .
         MACRO
         MYMAC
         MEND
++VER(Z038)`

	doc := parser.Parse(text)

	// Should parse both statements, skipping macro source lines
	if len(doc.Statements) != 2 {
		t.Fatalf("Expected 2 statements (++MAC, ++VER), got %d", len(doc.Statements))
	}

	if doc.Statements[0].Name != "++MAC" {
		t.Errorf("Expected first statement to be ++MAC, got %s", doc.Statements[0].Name)
	}

	if doc.Statements[1].Name != "++VER" {
		t.Errorf("Expected second statement to be ++VER, got %s", doc.Statements[1].Name)
	}
}

func TestParseMissingInlineData(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++MAC": {
			Name:        "++MAC",
			Description: "Macro statement",
			Type:        "control",
			InlineData:  true,
			Parameter:   "macro-name",
		},
	}

	parser := NewParser(statements)

	// Test missing inline data - statement expects inline data but next line is another statement
	// This should parse both statements, but diagnostics should warn about missing inline data
	text := `++MAC(MYMAC) DISTLIB(AMACLIB) .
++MAC(MYMAC2) DISTLIB(AMACLIB) .`

	doc := parser.Parse(text)

	// Should parse both statements
	if len(doc.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(doc.Statements))
	}

	if doc.Statements[0].Name != "++MAC" {
		t.Errorf("Expected first statement to be ++MAC, got %s", doc.Statements[0].Name)
	}

	if doc.Statements[1].Name != "++MAC" {
		t.Errorf("Expected second statement to be ++MAC, got %s", doc.Statements[1].Name)
	}

	// Note: Diagnostic for missing inline data will be generated by diagnostics_ast.go
	// The parser should still parse the structure correctly
}

// Test: ++APAR statement parsing
func TestParseApar(t *testing.T) {
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
				{Name: "RFDSNPFX", Parameter: "prefix", Type: "string"},
			},
		},
	}

	parser := NewParser(statements)

	text := "++APAR(UA12345) REWORK(2024001) DESC(\"Test APAR fix\") ."
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]
	if stmt.Name != "++APAR" {
		t.Errorf("Expected statement name ++APAR, got %s", stmt.Name)
	}

	// Check statement parameter (UA12345)
	if len(stmt.Children) < 1 {
		t.Fatalf("Expected at least 1 child (statement parameter), got %d", len(stmt.Children))
	}

	stmtParam := stmt.Children[0]
	if stmtParam.Type != NodeTypeParameter {
		t.Errorf("Expected first child to be NodeTypeParameter, got %v", stmtParam.Type)
	}
	if stmtParam.Value != "UA12345" {
		t.Errorf("Expected statement parameter value UA12345, got %s", stmtParam.Value)
	}

	// Check REWORK operand
	reworkFound := false
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand && child.Name == "REWORK" {
			reworkFound = true
			if len(child.Children) != 1 {
				t.Errorf("Expected REWORK to have 1 parameter child, got %d", len(child.Children))
			}
			break
		}
	}
	if !reworkFound {
		t.Error("REWORK operand not found")
	}
}

// Test: ++ASSIGN statement parsing
func TestParseAssign(t *testing.T) {
	statements := map[string]data.MCSStatement{
		"++ASSIGN": {
			Name:        "++ASSIGN",
			Description: "Assign source ID to SYSMODs",
			Type:        "MCS",
			Operands: []data.Operand{
				{Name: "SOURCEID", Parameter: "source-id", Type: "string"},
				{Name: "TO", Parameter: "sysmod-ids", Type: "string"},
			},
		},
	}

	parser := NewParser(statements)

	text := "++ASSIGN SOURCEID(MYSOURCE) TO(UA12345,UA12346) ."
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]
	if stmt.Name != "++ASSIGN" {
		t.Errorf("Expected statement name ++ASSIGN, got %s", stmt.Name)
	}

	// Check SOURCEID operand
	sourceidFound := false
	toFound := false
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand {
			if child.Name == "SOURCEID" {
				sourceidFound = true
			}
			if child.Name == "TO" {
				toFound = true
			}
		}
	}

	if !sourceidFound {
		t.Error("SOURCEID operand not found")
	}
	if !toFound {
		t.Error("TO operand not found")
	}
}

// Test: ++DELETE statement parsing
func TestParseDelete(t *testing.T) {
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

	parser := NewParser(statements)

	text := "++DELETE(MYMODULE) SYSLIB(SYSLIB1) ."
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]
	if stmt.Name != "++DELETE" {
		t.Errorf("Expected statement name ++DELETE, got %s", stmt.Name)
	}

	// Check statement parameter (MYMODULE)
	if len(stmt.Children) < 1 {
		t.Fatalf("Expected at least 1 child (statement parameter), got %d", len(stmt.Children))
	}

	stmtParam := stmt.Children[0]
	if stmtParam.Value != "MYMODULE" {
		t.Errorf("Expected statement parameter value MYMODULE, got %s", stmtParam.Value)
	}
}

// Test: ++HOLD statement parsing
func TestParseHold(t *testing.T) {
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

	parser := NewParser(statements)

	text := "++HOLD(UA12345) ERROR FMID(HBB7790) REASON(ACTION) DATE(24001) ."
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]
	if stmt.Name != "++HOLD" {
		t.Errorf("Expected statement name ++HOLD, got %s", stmt.Name)
	}

	// Check operands
	errorFound := false
	fmidFound := false
	reasonFound := false
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand {
			if child.Name == "ERROR" {
				errorFound = true
			}
			if child.Name == "FMID" {
				fmidFound = true
			}
			if child.Name == "REASON" {
				reasonFound = true
			}
		}
	}

	if !errorFound {
		t.Error("ERROR operand not found")
	}
	if !fmidFound {
		t.Error("FMID operand not found")
	}
	if !reasonFound {
		t.Error("REASON operand not found")
	}
}

// Test: ++IF statement parsing
func TestParseIf(t *testing.T) {
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

	parser := NewParser(statements)

	text := "++IF FMID(HBB7790) THEN REQ(UA12345) ."
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]
	if stmt.Name != "++IF" {
		t.Errorf("Expected statement name ++IF, got %s", stmt.Name)
	}

	// Check operands
	fmidFound := false
	thenFound := false
	reqFound := false
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand {
			if child.Name == "FMID" {
				fmidFound = true
			}
			if child.Name == "THEN" {
				thenFound = true
			}
			if child.Name == "REQ" {
				reqFound = true
			}
		}
	}

	if !fmidFound {
		t.Error("FMID operand not found")
	}
	if !thenFound {
		t.Error("THEN operand not found")
	}
	if !reqFound {
		t.Error("REQ operand not found")
	}
}

// Test: ++FEATURE statement parsing
func TestParseFeature(t *testing.T) {
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

	parser := NewParser(statements)

	text := "++FEATURE(MYFEATURE) FMID(HBB7790) DESC(\"Test feature\") ."
	doc := parser.Parse(text)

	if len(doc.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(doc.Statements))
	}

	stmt := doc.Statements[0]
	if stmt.Name != "++FEATURE" {
		t.Errorf("Expected statement name ++FEATURE, got %s", stmt.Name)
	}

	// Check statement parameter (MYFEATURE)
	if len(stmt.Children) < 1 {
		t.Fatalf("Expected at least 1 child (statement parameter), got %d", len(stmt.Children))
	}

	stmtParam := stmt.Children[0]
	if stmtParam.Value != "MYFEATURE" {
		t.Errorf("Expected statement parameter value MYFEATURE, got %s", stmtParam.Value)
	}

	// Check FMID operand
	fmidFound := false
	for _, child := range stmt.Children {
		if child.Type == NodeTypeOperand && child.Name == "FMID" {
			fmidFound = true
			break
		}
	}
	if !fmidFound {
		t.Error("FMID operand not found")
	}
}
