package formatting

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

// newTestFormatter creates a Provider with real smpe.json data and standard config
func newTestFormatter(t *testing.T) (*parser.Parser, *Provider) {
	t.Helper()
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := parser.NewParser(store.Statements)
	fp := NewProvider()
	fp.SetConfig(&Config{
		Enabled:             true,
		IndentContinuation:  4,
		OneOperandPerLine:   true,
		WrapListsAfterN:     2,
		MoveLeadingComments: true,
	})
	return p, fp
}

// formatOnce parses and formats input, returns the formatted text
func formatOnce(t *testing.T, p *parser.Parser, fp *Provider, input string) string {
	t.Helper()
	doc := p.Parse(input)
	edits := fp.FormatDocument(doc, input)
	if len(edits) == 0 {
		return input // already formatted
	}
	return edits[0].NewText
}

// --- Output structure tests ---

func TestFormatOutputStructure_SingleStatement(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++RELEASE(UZ12345) FMID(FXY1040) USER REASON(CPU0A) .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, "++RELEASE(UZ12345)") {
		t.Error("Header line missing")
	}
	if !strings.Contains(result, "    FMID(FXY1040)") {
		t.Error("FMID not indented correctly (expected 4 spaces)")
	}
	if !strings.Contains(result, "    USER") {
		t.Error("USER not indented correctly")
	}
	if !strings.Contains(result, "    REASON(CPU0A)") {
		t.Error("REASON not indented correctly")
	}
	// Terminator on own line
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine != "." {
		t.Errorf("Expected terminator '.' on last line, got %q", lastLine)
	}
	// Header line must not contain operands
	headerLine := lines[0]
	if strings.Contains(headerLine, "FMID") || strings.Contains(headerLine, "USER") {
		t.Errorf("Header line should not contain operands: %q", headerLine)
	}
}

func TestFormatOutputStructure_ListWrapping(t *testing.T) {
	p, fp := newTestFormatter(t)
	// PRE with 3 items exceeds WrapListsAfterN=2
	input := "++VER(Z038) PRE(U001 U002 U003) .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	// All three values must appear
	for _, v := range []string{"U001", "U002", "U003"} {
		if !strings.Contains(result, v) {
			t.Errorf("Value %q missing from formatted output", v)
		}
	}
	// With wrapping, values should be on separate lines
	u001Line := -1
	u003Line := -1
	for i, line := range strings.Split(result, "\n") {
		if strings.Contains(line, "U001") {
			u001Line = i
		}
		if strings.Contains(line, "U003") {
			u003Line = i
		}
	}
	if u001Line == u003Line {
		t.Error("U001 and U003 should be on different lines (list wrapping)")
	}
}

func TestFormatStatementWithNoOperands(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++VER(Z038) .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	// Should be exactly 2 lines: header + terminator
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines for statement with no operands, got %d:\n%s", len(lines), result)
	}
	if !strings.HasPrefix(lines[0], "++VER(Z038)") {
		t.Errorf("First line should be ++VER(Z038), got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "." {
		t.Errorf("Second line should be '.', got %q", lines[1])
	}
}

// --- Regression tests for dot-in-string and dot-in-comment bugs ---

func TestFormatDotInSingleQuotedString(t *testing.T) {
	// Regression: LINK('../path') caused formatter to treat the dot in '../'
	// as statement terminator, producing duplicate PARM output
	p, fp := newTestFormatter(t)
	input := "++HFS(MYHFS)\n  SYSLIB(MYHFSLIB)\n  LINK('../myfile')\n  PARM(PATHMODE(0,7,5,5))\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	// PARM must appear exactly once
	count := strings.Count(result, "PARM(")
	if count != 1 {
		t.Errorf("PARM( should appear exactly once, got %d times:\n%s", count, result)
	}
	// The link value must be preserved intact
	if !strings.Contains(result, "LINK('../myfile')") {
		t.Errorf("LINK value with dot in single-quoted string not preserved:\n%s", result)
	}
	// Terminator must appear exactly once at end
	terminatorCount := strings.Count(result, "\n.")
	if terminatorCount != 1 {
		t.Errorf("Terminator should appear exactly once, got %d times:\n%s", terminatorCount, result)
	}
	// Must be idempotent
	result2 := formatOnce(t, p, fp, result)
	if result != result2 {
		t.Errorf("Not idempotent!\nFirst:\n%s\nSecond:\n%s", result, result2)
	}
}

func TestFormatDotInInlineComment(t *testing.T) {
	// Regression: /* version 12.7 */ caused inStatement=false too early,
	// making subsequent statement lines lose their syntax highlighting
	p, fp := newTestFormatter(t)
	input := "++RELEASE(UZ12345) FMID(FXY1040) /* comment v1.0 */ USER REASON(CPU0A) .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	// All operands must be in the output
	for _, op := range []string{"FMID(FXY1040)", "USER", "REASON(CPU0A)"} {
		if !strings.Contains(result, op) {
			t.Errorf("Operand %q missing from formatted output:\n%s", op, result)
		}
	}
	// Must be idempotent
	result2 := formatOnce(t, p, fp, result)
	if result != result2 {
		t.Errorf("Not idempotent!\nFirst:\n%s\nSecond:\n%s", result, result2)
	}
}

func TestFormatDotInMultipleInlineComments(t *testing.T) {
	// Regression: multiple inline comments with dots (like version numbers)
	// Each line is its own /* ... */ comment
	p, fp := newTestFormatter(t)
	input := "++FUNCTION(B1PF127)\n" +
		"  /* BGS SYSTEMS INC. */\n" +
		"  /* MVS PRODUCT FAMILY RELEASE 12.7 */\n" +
		"  /* BEST/1-DATACENTER R3.2 */\n" +
		".\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	// Statement must remain intact (all comment lines preserved)
	if !strings.Contains(result, "++FUNCTION(B1PF127)") {
		t.Error("Statement header missing")
	}
	if !strings.Contains(result, "12.7") {
		t.Error("Version number in comment was lost")
	}
	if !strings.Contains(result, "R3.2") {
		t.Error("Release in comment was lost")
	}
	// Must be idempotent
	result2 := formatOnce(t, p, fp, result)
	if result != result2 {
		t.Errorf("Not idempotent!\nFirst:\n%s\nSecond:\n%s", result, result2)
	}
}

func TestFormatDescWithSpacesNotSplit(t *testing.T) {
	// Regression: DESC(This is a long description) must not be split at spaces
	p, fp := newTestFormatter(t)
	input := "++USERMOD(LJS2012) DESC(This is a long description with many words) .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, "DESC(This is a long description with many words)") {
		t.Errorf("DESC value was split or corrupted:\n%s", result)
	}
}

func TestFormatListOperandIsWrapped(t *testing.T) {
	// PRE is type "list" — space-separated values must still be wrapped when > WrapListsAfterN
	p, fp := newTestFormatter(t)
	input := "++VER(Z038) PRE(U001 U002 U003) .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	for _, v := range []string{"U001", "U002", "U003"} {
		if !strings.Contains(result, v) {
			t.Errorf("Value %q missing from formatted output", v)
		}
	}
	lines := strings.Split(result, "\n")
	u001Line, u003Line := -1, -1
	for i, line := range lines {
		if strings.Contains(line, "U001") {
			u001Line = i
		}
		if strings.Contains(line, "U003") {
			u003Line = i
		}
	}
	if u001Line == u003Line {
		t.Error("U001 and U003 should be on different lines (list wrapping)")
	}
}

func TestFormatDotInStatementParameter(t *testing.T) {
	// Regression: ++PRODUCT(PROD001,01.00.00) — dot inside the statement parameter
	// must not be treated as terminator, causing the body to be duplicated
	p, fp := newTestFormatter(t)
	input := "++PRODUCT(PROD001,01.00.00)\n    SREL(Z038)\n    DESCRIPTION('Test Product')\n.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	srelCount := strings.Count(result, "SREL(Z038)")
	if srelCount != 1 {
		t.Errorf("SREL(Z038) should appear exactly once, got %d times:\n%s", srelCount, result)
	}
	descCount := strings.Count(result, "DESCRIPTION(")
	if descCount != 1 {
		t.Errorf("DESCRIPTION( should appear exactly once, got %d times:\n%s", descCount, result)
	}
	result2 := formatOnce(t, p, fp, result)
	if result != result2 {
		t.Errorf("Not idempotent!\nFirst:\n%s\nSecond:\n%s", result, result2)
	}
}

// --- Direct unit tests for getStatementEndLine ---

func TestGetStatementEndLine_DotInSingleQuotedString(t *testing.T) {
	fp := NewProvider()
	fp.SetConfig(&Config{Enabled: true, IndentContinuation: 4})

	lines := []string{
		"++HFS(MYHFS) LINK('../mypath') .",
	}
	stmt := makeMinimalNode(0)
	result := fp.getStatementEndLine(stmt, lines)
	if result != 0 {
		t.Errorf("Expected terminatorLine=0, got %d (dot in quoted string was treated as terminator)", result)
	}
}

func TestGetStatementEndLine_DotInInlineComment(t *testing.T) {
	fp := NewProvider()
	fp.SetConfig(&Config{Enabled: true, IndentContinuation: 4})

	lines := []string{
		"++VER(Z038) /* version 1.5 */ .",
	}
	stmt := makeMinimalNode(0)
	result := fp.getStatementEndLine(stmt, lines)
	if result != 0 {
		t.Errorf("Expected terminatorLine=0, got %d (dot in comment was treated as terminator)", result)
	}
}

func TestGetStatementEndLine_MultilineStatement(t *testing.T) {
	fp := NewProvider()
	fp.SetConfig(&Config{Enabled: true, IndentContinuation: 4})

	lines := []string{
		"++HFS(MYHFS)",
		"  SYSLIB(MYHFSLIB)",
		"  LINK('../myfile')",
		"  PARM(PATHMODE(0,7,5,5))",
		".",
	}
	stmt := makeMinimalNode(0)
	result := fp.getStatementEndLine(stmt, lines)
	if result != 4 {
		t.Errorf("Expected terminatorLine=4, got %d", result)
	}
}

func TestGetStatementEndLine_TerminatorAfterComment(t *testing.T) {
	fp := NewProvider()
	fp.SetConfig(&Config{Enabled: true, IndentContinuation: 4})

	// Terminator on same line as closing comment
	lines := []string{
		"++VER(Z038) /* comment v1.2 */ .",
	}
	stmt := makeMinimalNode(0)
	result := fp.getStatementEndLine(stmt, lines)
	if result != 0 {
		t.Errorf("Expected terminatorLine=0, got %d", result)
	}
}

// makeMinimalNode creates a minimal parser.Node for testing getStatementEndLine
func makeMinimalNode(startLine int) *parser.Node {
	return &parser.Node{
		Position: parser.Position{Line: startLine},
	}
}
