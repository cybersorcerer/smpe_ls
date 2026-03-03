package semantic

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

func newTestProvider(t *testing.T) (*parser.Parser, *Provider) {
	t.Helper()
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := parser.NewParser(store.Statements)
	sp := NewProvider(store.Statements)
	return p, sp
}

// --- encodeTokens unit tests ---

func TestEncodeTokens_Empty(t *testing.T) {
	sp := NewProvider(map[string]data.MCSStatement{})
	result := sp.encodeTokens([]Token{})
	if len(result) != 0 {
		t.Errorf("Expected empty slice for no tokens, got %v", result)
	}
}

func TestEncodeTokens_SingleToken(t *testing.T) {
	sp := NewProvider(map[string]data.MCSStatement{})
	tokens := []Token{
		{Line: 0, StartChar: 2, Length: 8, Type: TokenTypeKeyword, Modifiers: TokenModifierNone},
	}
	result := sp.encodeTokens(tokens)
	// Single token: [deltaLine=0, deltaChar=2, length=8, type=0, modifiers=0]
	if len(result) != 5 {
		t.Fatalf("Expected 5 integers for one token, got %d: %v", len(result), result)
	}
	if result[0] != 0 {
		t.Errorf("deltaLine: expected 0, got %d", result[0])
	}
	if result[1] != 2 {
		t.Errorf("deltaChar: expected 2, got %d", result[1])
	}
	if result[2] != 8 {
		t.Errorf("length: expected 8, got %d", result[2])
	}
	if result[3] != int(TokenTypeKeyword) {
		t.Errorf("type: expected %d (Keyword), got %d", int(TokenTypeKeyword), result[3])
	}
	if result[4] != int(TokenModifierNone) {
		t.Errorf("modifiers: expected 0, got %d", result[4])
	}
}

func TestEncodeTokens_TwoTokensSameLine(t *testing.T) {
	sp := NewProvider(map[string]data.MCSStatement{})
	tokens := []Token{
		{Line: 0, StartChar: 0, Length: 9, Type: TokenTypeKeyword},
		{Line: 0, StartChar: 10, Length: 4, Type: TokenTypeFunction},
	}
	result := sp.encodeTokens(tokens)
	if len(result) != 10 {
		t.Fatalf("Expected 10 integers for two tokens, got %d", len(result))
	}
	// Second token on same line: deltaLine=0, deltaChar=10-0=10
	if result[5] != 0 {
		t.Errorf("Token2 deltaLine: expected 0, got %d", result[5])
	}
	if result[6] != 10 {
		t.Errorf("Token2 deltaChar: expected 10, got %d", result[6])
	}
	if result[7] != 4 {
		t.Errorf("Token2 length: expected 4, got %d", result[7])
	}
	if result[8] != int(TokenTypeFunction) {
		t.Errorf("Token2 type: expected %d (Function), got %d", int(TokenTypeFunction), result[8])
	}
}

func TestEncodeTokens_TwoTokensDifferentLines(t *testing.T) {
	sp := NewProvider(map[string]data.MCSStatement{})
	tokens := []Token{
		{Line: 2, StartChar: 5, Length: 9, Type: TokenTypeKeyword},
		{Line: 4, StartChar: 5, Length: 4, Type: TokenTypeFunction},
	}
	result := sp.encodeTokens(tokens)
	if len(result) != 10 {
		t.Fatalf("Expected 10 integers for two tokens, got %d", len(result))
	}
	// First token: deltaLine=2 (from line 0), deltaChar=5
	if result[0] != 2 {
		t.Errorf("Token1 deltaLine: expected 2, got %d", result[0])
	}
	if result[1] != 5 {
		t.Errorf("Token1 deltaChar: expected 5, got %d", result[1])
	}
	// Second token on different line: deltaLine=2 (line 4 - line 2), deltaChar=5 (absolute on new line)
	if result[5] != 2 {
		t.Errorf("Token2 deltaLine: expected 2, got %d", result[5])
	}
	if result[6] != 5 {
		t.Errorf("Token2 deltaChar: expected 5 (absolute on new line), got %d", result[6])
	}
}

// --- BuildTokensFromAST integration tests ---

func TestBuildTokensFromAST_SimpleStatement(t *testing.T) {
	p, sp := newTestProvider(t)
	input := "++USERMOD(LJS2012) .\n"
	doc := p.Parse(input)
	result := sp.BuildTokensFromAST(doc, input)

	if len(result) == 0 {
		t.Fatal("Expected tokens from AST, got none")
	}
	if len(result)%5 != 0 {
		t.Errorf("Token count must be multiple of 5, got %d", len(result))
	}
	// First token: ++USERMOD keyword at line 0, char 0, length 9
	if result[0] != 0 {
		t.Errorf("First token deltaLine: expected 0, got %d", result[0])
	}
	if result[1] != 0 {
		t.Errorf("First token deltaChar: expected 0, got %d", result[1])
	}
	if result[2] != 9 {
		t.Errorf("First token length: expected 9 (++USERMOD), got %d", result[2])
	}
	if result[3] != int(TokenTypeKeyword) {
		t.Errorf("First token type: expected %d (Keyword), got %d", int(TokenTypeKeyword), result[3])
	}
}

func TestBuildTokensFromAST_StatementWithOperand(t *testing.T) {
	p, sp := newTestProvider(t)
	input := "++USERMOD(LJS2012) DESC(my-description) .\n"
	doc := p.Parse(input)
	result := sp.BuildTokensFromAST(doc, input)

	if len(result)%5 != 0 {
		t.Errorf("Token count must be multiple of 5, got %d", len(result))
	}
	tokenCount := len(result) / 5

	// Expect: statement keyword + statement parameter + operand + operand parameter = at least 4
	if tokenCount < 4 {
		t.Errorf("Expected at least 4 tokens for statement with operand, got %d", tokenCount)
	}

	foundFunction := false
	for i := 0; i < tokenCount; i++ {
		if result[i*5+3] == int(TokenTypeFunction) {
			foundFunction = true
			break
		}
	}
	if !foundFunction {
		t.Error("Expected at least one Function-type token for operand DESC")
	}
}

func TestBuildTokensFromAST_MultipleStatements(t *testing.T) {
	p, sp := newTestProvider(t)
	input := "++USERMOD(LJS2012) .\n++PTF(UA99999) .\n"
	doc := p.Parse(input)
	result := sp.BuildTokensFromAST(doc, input)

	if len(result)%5 != 0 {
		t.Fatalf("Token count must be multiple of 5, got %d", len(result))
	}
	tokenCount := len(result) / 5

	if tokenCount < 2 {
		t.Errorf("Expected at least 2 keyword tokens for 2 statements, got %d", tokenCount)
	}

	// All deltaLine values must be >= 0
	for i := 0; i < tokenCount; i++ {
		if result[i*5] < 0 {
			t.Errorf("Token %d has negative deltaLine %d", i, result[i*5])
		}
	}
}

func TestBuildTokensFromAST_InlineComment(t *testing.T) {
	p, sp := newTestProvider(t)
	input := "++VER(Z038) /* a comment */ .\n"
	doc := p.Parse(input)
	result := sp.BuildTokensFromAST(doc, input)

	if len(result)%5 != 0 {
		t.Fatalf("Token count must be multiple of 5, got %d", len(result))
	}
	tokenCount := len(result) / 5

	foundComment := false
	for i := 0; i < tokenCount; i++ {
		if result[i*5+3] == int(TokenTypeComment) {
			foundComment = true
			break
		}
	}
	if !foundComment {
		t.Error("Expected a Comment-type token for inline comment")
	}
}

func TestBuildTokensFromAST_CommentInsideStatement(t *testing.T) {
	p, sp := newTestProvider(t)
	// Valid: comment inside a statement (between ++ and terminator .)
	input := "++VER(Z038)\n  /* comment inside statement */\n  PRE(UA12345)\n.\n"
	doc := p.Parse(input)
	result := sp.BuildTokensFromAST(doc, input)

	if len(result)%5 != 0 {
		t.Fatalf("Token count must be multiple of 5, got %d", len(result))
	}
	tokenCount := len(result) / 5

	foundComment := false
	foundKeyword := false
	for i := 0; i < tokenCount; i++ {
		switch result[i*5+3] {
		case int(TokenTypeComment):
			foundComment = true
		case int(TokenTypeKeyword):
			foundKeyword = true
		}
	}
	if !foundComment {
		t.Error("Expected Comment-type token for comment inside statement")
	}
	if !foundKeyword {
		t.Error("Expected Keyword-type token for ++VER statement")
	}
}

func TestBuildTokensFromAST_EmptyInput(t *testing.T) {
	p, sp := newTestProvider(t)
	input := ""
	doc := p.Parse(input)
	result := sp.BuildTokensFromAST(doc, input)

	if len(result) != 0 {
		t.Errorf("Expected no tokens for empty input, got %v", result)
	}
}

func TestBuildTokensFromAST_TokensAreSorted(t *testing.T) {
	p, sp := newTestProvider(t)
	input := "/* comment line 0 */\n++USERMOD(LJS2012) DESC(my-desc) REWORK(2022056) .\n"
	doc := p.Parse(input)
	result := sp.BuildTokensFromAST(doc, input)

	if len(result)%5 != 0 {
		t.Fatalf("Token count must be multiple of 5, got %d", len(result))
	}
	tokenCount := len(result) / 5

	// Reconstruct absolute positions and verify non-decreasing order
	absLine := 0
	absChar := 0
	for i := 0; i < tokenCount; i++ {
		deltaLine := result[i*5]
		deltaChar := result[i*5+1]

		prevLine := absLine
		prevChar := absChar

		if deltaLine == 0 {
			absChar += deltaChar
		} else {
			absLine += deltaLine
			absChar = deltaChar
		}

		if i > 0 {
			if absLine < prevLine || (absLine == prevLine && absChar < prevChar) {
				t.Errorf("Token %d is out of order: abs line %d char %d comes before prev line %d char %d",
					i, absLine, absChar, prevLine, prevChar)
			}
		}
	}
}
