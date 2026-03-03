package data

import (
	"os"
	"path/filepath"
	"testing"
)

const smpeJSONPath = "../../data/smpe.json"

func TestLoadValid(t *testing.T) {
	store, err := Load(smpeJSONPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if store == nil {
		t.Fatal("Store is nil")
	}
	if len(store.Statements) == 0 {
		t.Error("Statements map is empty")
	}
	if len(store.List) == 0 {
		t.Error("Statements list is empty")
	}
	if len(store.Statements) != len(store.List) {
		t.Errorf("Map and list out of sync: map=%d list=%d", len(store.Statements), len(store.List))
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("./nonexistent_file_that_does_not_exist.json")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("[not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_, err := Load(bad)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadMapKeyMatchesName(t *testing.T) {
	store, err := Load(smpeJSONPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	for key, stmt := range store.Statements {
		if key != stmt.Name {
			t.Errorf("Map key %q does not match statement name %q", key, stmt.Name)
		}
	}
}

func TestLoadKnownStatementsExist(t *testing.T) {
	store, err := Load(smpeJSONPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	required := []string{"++APAR", "++PTF", "++USERMOD", "++FUNCTION", "++VER", "++MOD", "++MAC", "++JCLIN"}
	for _, name := range required {
		if _, ok := store.Statements[name]; !ok {
			t.Errorf("Expected statement %q not found in store", name)
		}
	}
}

func TestLoadStatementFields(t *testing.T) {
	store, err := Load(smpeJSONPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// ++USERMOD must have a parameter
	usermod, ok := store.Statements["++USERMOD"]
	if !ok {
		t.Fatal("++USERMOD not found")
	}
	if usermod.Parameter == "" {
		t.Error("++USERMOD should have a non-empty Parameter field")
	}

	// ++MAC must have InlineData=true
	mac, ok := store.Statements["++MAC"]
	if !ok {
		t.Fatal("++MAC not found")
	}
	if !mac.InlineData {
		t.Error("++MAC should have InlineData=true")
	}

	// ++JCLIN must have InlineData=true
	jclin, ok := store.Statements["++JCLIN"]
	if !ok {
		t.Fatal("++JCLIN not found")
	}
	if !jclin.InlineData {
		t.Error("++JCLIN should have InlineData=true")
	}

	// ++VER must have operands (PRE, REQ, etc.)
	ver, ok := store.Statements["++VER"]
	if !ok {
		t.Fatal("++VER not found")
	}
	if len(ver.Operands) == 0 {
		t.Error("++VER should have operands")
	}
}

func TestLoadHFSStatementsHaveCorrectOperands(t *testing.T) {
	store, err := Load(smpeJSONPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// All HFS/Data Element statements should have SHSCRIPT, SYMPATH, RMID
	hfsStatements := []string{
		"++HFS",
		"++AIX1", "++AIX2", "++AIX3", "++AIX4", "++AIX5",
		"++CLIENT1", "++CLIENT2", "++CLIENT3", "++CLIENT4", "++CLIENT5",
		"++OS21", "++OS22", "++OS23", "++OS24", "++OS25",
		"++UNIX1", "++UNIX2", "++UNIX3", "++UNIX4", "++UNIX5",
		"++WIN1", "++WIN2", "++WIN3", "++WIN4", "++WIN5",
	}
	requiredOps := []string{"SHSCRIPT", "SYMPATH", "RMID", "DISTLIB", "SYSLIB", "BINARY", "TEXT", "DELETE"}

	for _, stmtName := range hfsStatements {
		stmt, ok := store.Statements[stmtName]
		if !ok {
			t.Errorf("Statement %q not found", stmtName)
			continue
		}
		opNames := make(map[string]bool)
		for _, op := range stmt.Operands {
			opNames[op.Name] = true
		}
		for _, opName := range requiredOps {
			if !opNames[opName] {
				t.Errorf("%s: missing operand %q", stmtName, opName)
			}
		}
	}
}

func TestLoadAllStatementsHaveNameAndType(t *testing.T) {
	store, err := Load(smpeJSONPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	for _, stmt := range store.List {
		if stmt.Name == "" {
			t.Error("Found statement with empty Name")
		}
		if stmt.Type == "" {
			t.Errorf("Statement %q has empty Type", stmt.Name)
		}
	}
}

func TestLoadNewFormatRefResolution(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "smpe_new.json")
	content := `{
		"templates": {
			"test_operands": [
				{"name": "DISTLIB", "type": "string", "description": "Distribution library"},
				{"name": "SYSLIB",  "type": "string", "description": "Target library"}
			]
		},
		"statements": [
			{
				"name": "++TESTSTMT",
				"type": "HFS",
				"description": "Test statement",
				"operands": [{"$ref": "test_operands"}]
			},
			{
				"name": "++TESTSTMT2",
				"type": "HFS",
				"description": "Test statement 2",
				"operands": [{"$ref": "test_operands"}]
			}
		]
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp: %v", err)
	}

	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load new format failed: %v", err)
	}

	for _, name := range []string{"++TESTSTMT", "++TESTSTMT2"} {
		stmt, ok := store.Statements[name]
		if !ok {
			t.Errorf("Statement %q not found", name)
			continue
		}
		if len(stmt.Operands) != 2 {
			t.Errorf("%s: expected 2 operands after $ref resolution, got %d", name, len(stmt.Operands))
			continue
		}
		for _, op := range stmt.Operands {
			if op.Name == "" {
				t.Errorf("%s: operand with empty Name — $ref may not have resolved", name)
			}
		}
	}

	// Verify the two statements got independent copies (not shared slice)
	s1 := store.Statements["++TESTSTMT"]
	s2 := store.Statements["++TESTSTMT2"]
	if len(s1.Operands) > 0 && len(s2.Operands) > 0 {
		s1.Operands[0].Name = "MUTATED"
		if s2.Operands[0].Name == "MUTATED" {
			t.Error("$ref operands are shared (not copied) — mutation of one affects the other")
		}
	}
}

func TestLoadNewFormatLegacyFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.json")
	content := `[{"name":"++LEGACYSTMT","type":"MCS","description":"legacy test"}]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load legacy format failed: %v", err)
	}
	if _, ok := store.Statements["++LEGACYSTMT"]; !ok {
		t.Error("++LEGACYSTMT not found — legacy format fallback failed")
	}
}

func TestLoadNewFormatUnknownRefReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "badref.json")
	content := `{
		"templates": {},
		"statements": [
			{"name":"++BADREF","type":"HFS","description":"x","operands":[{"$ref":"nonexistent"}]}
		]
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Error("Expected error for unknown $ref, got nil")
	}
}
