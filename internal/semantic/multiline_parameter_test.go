package semantic

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

// A semantic token must stay inside one line. The LSP spec has no way to
// express a token that spans a line break, and VS Code drops such a token
// entirely - the text then keeps the editor's default colour.
//
// An operand written across lines produces exactly that:
//
//	SUP(
//	    LBCP034
//	)
//
// The parameter node covers everything between the parentheses, newlines
// included, so its single token ran past the end of its line and the value
// lost its highlighting while FMID(HBB77E0) on one line kept it.

// decodedToken is one token read back from the LSP wire format.
type decodedToken struct {
	line, startChar, length, tokType int
}

// decodeTokens turns the delta-encoded LSP data back into absolute positions.
func decodeTokens(data []int) []decodedToken {
	var out []decodedToken
	line, char := 0, 0
	for i := 0; i+4 < len(data)+1 && i+5 <= len(data); i += 5 {
		deltaLine, deltaChar := data[i], data[i+1]
		line += deltaLine
		if deltaLine == 0 {
			char += deltaChar
		} else {
			char = deltaChar
		}
		out = append(out, decodedToken{line, char, data[i+2], data[i+3]})
	}
	return out
}

func buildTokens(t *testing.T, text string) ([]decodedToken, []string) {
	t.Helper()
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	doc := parser.NewParser(store.Statements).Parse(text)
	prov := NewProvider(store.Statements)
	return decodeTokens(prov.BuildTokensFromAST(doc, text)), strings.Split(text, "\n")
}

// No token may reach past the end of its own line.
func TestTokensStayWithinTheirLine(t *testing.T) {
	text := "++VER(Z038)\n" +
		"    FMID(HBB77E0)\n" +
		"    SUP(\n" +
		"        LBCP034\n" +
		"    )\n" +
		"    .\n"

	tokens, lines := buildTokens(t, text)
	for _, tok := range tokens {
		if tok.line >= len(lines) {
			t.Errorf("Token on line %d, but the text has %d lines", tok.line, len(lines))
			continue
		}
		width := len([]rune(lines[tok.line]))
		if tok.startChar+tok.length > width {
			t.Errorf("Token reaches past the line end: line %d, char %d, length %d, line is %d wide (%q)",
				tok.line, tok.startChar, tok.length, width, lines[tok.line])
		}
	}
}

// The value itself must carry a parameter token, wherever it is written.
func TestMultilineOperandValueIsHighlighted(t *testing.T) {
	text := "++VER(Z038)\n" +
		"    FMID(HBB77E0)\n" +
		"    SUP(\n" +
		"        LBCP034\n" +
		"    )\n" +
		"    .\n"

	tokens, lines := buildTokens(t, text)

	// LBCP034 sits on line 3, indented by 8.
	const wantLine, wantChar, wantLen = 3, 8, 7
	if got := string([]rune(lines[wantLine])[wantChar : wantChar+wantLen]); got != "LBCP034" {
		t.Fatalf("Test text moved: expected LBCP034 at line %d char %d, found %q", wantLine, wantChar, got)
	}

	for _, tok := range tokens {
		if tok.line == wantLine && tok.startChar == wantChar &&
			tok.length == wantLen && tok.tokType == int(TokenTypeParameter) {
			return
		}
	}
	t.Errorf("No parameter token on LBCP034 (line %d, char %d, length %d).\nTokens: %+v",
		wantLine, wantChar, wantLen, tokens)
}

// A value on the same line as its operand keeps working as before.
func TestSingleLineOperandValueStillHighlighted(t *testing.T) {
	text := "++VER(Z038)\n    FMID(HBB77E0)\n    .\n"
	tokens, _ := buildTokens(t, text)

	for _, tok := range tokens {
		if tok.line == 1 && tok.startChar == 9 && tok.length == 7 &&
			tok.tokType == int(TokenTypeParameter) {
			return
		}
	}
	t.Errorf("No parameter token on HBB77E0 (line 1, char 9, length 7).\nTokens: %+v", tokens)
}

// A list written across lines gets one token per line, on the values only.
func TestMultilineListValuesEachGetAToken(t *testing.T) {
	text := "++ASSIGN\n" +
		"    TO(\n" +
		"        A12345,\n" +
		"        A23456\n" +
		"    )\n" +
		"    .\n"

	tokens, lines := buildTokens(t, text)
	for _, want := range []struct {
		line int
		text string
	}{{2, "A12345"}, {3, "A23456"}} {
		found := false
		for _, tok := range tokens {
			if tok.line != want.line || tok.tokType != int(TokenTypeParameter) {
				continue
			}
			got := string([]rune(lines[tok.line])[tok.startChar : tok.startChar+tok.length])
			if got == want.text {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("No parameter token covering %q on line %d.\nTokens: %+v", want.text, want.line, tokens)
		}
	}
}
