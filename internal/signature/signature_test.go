package signature

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

func createTestProviders() (*data.Store, *parser.Parser, *Provider) {
	statements := map[string]data.MCSStatement{
		"++MOVE": {
			Name:        "++MOVE",
			Description: "Moves an element",
			Parameter:   "element_name",
			Operands: []data.Operand{
				{Name: "DISTLIB", Parameter: "DDNAME", Type: "string",
					Description: "Specifies the ddname of the distribution library. Extra trailing text."},
				{Name: "FMID", Parameter: "FMID-LIST", Type: "list",
					Description: "Specifies the FMID that owns the element."},
				{Name: "MAC", Type: "boolean",
					Description: "Specifies that the member to be moved is a macro."},
			},
		},
	}
	store := &data.Store{Statements: statements}
	p := parser.NewParser(statements)
	return store, p, NewProvider(store)
}

// colAfterParen returns (line, character) just inside an operand's "(".
func colAfterParen(text, operand string) (int, int) {
	for li, ln := range strings.Split(text, "\n") {
		idx := strings.Index(ln, operand+"(")
		if idx >= 0 {
			return li, idx + len(operand) + 1
		}
	}
	return -1, -1
}

func TestSignatureStringOperand(t *testing.T) {
	_, p, sp := createTestProviders()
	text := "++MOVE(M1)\n    DISTLIB()\n."
	line, char := colAfterParen(text, "DISTLIB")
	doc := p.Parse(text)
	sh := sp.GetSignatureHelp(doc, text, line, char)
	if sh == nil {
		t.Fatal("expected signature help for DISTLIB, got nil")
	}
	if len(sh.Signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(sh.Signatures))
	}
	if sh.Signatures[0].Label != "DISTLIB(DDNAME)" {
		t.Errorf("unexpected label: %q", sh.Signatures[0].Label)
	}
	if !strings.Contains(sh.Signatures[0].Documentation, "string") {
		t.Errorf("doc should mention type 'string', got %q", sh.Signatures[0].Documentation)
	}
	if strings.Contains(sh.Signatures[0].Documentation, "Extra trailing text") {
		t.Errorf("doc not truncated to first sentence: %q", sh.Signatures[0].Documentation)
	}
}

func TestSignatureBooleanOperandNil(t *testing.T) {
	_, p, sp := createTestProviders()
	text := "++MOVE(M1)\n    MAC()\n."
	line, char := colAfterParen(text, "MAC")
	doc := p.Parse(text)
	if sh := sp.GetSignatureHelp(doc, text, line, char); sh != nil {
		t.Errorf("expected nil for boolean operand MAC, got %+v", sh)
	}
}

func TestSignatureDisabledNil(t *testing.T) {
	_, p, sp := createTestProviders()
	sp.SetEnabled(false)
	text := "++MOVE(M1)\n    DISTLIB()\n."
	line, char := colAfterParen(text, "DISTLIB")
	doc := p.Parse(text)
	if sh := sp.GetSignatureHelp(doc, text, line, char); sh != nil {
		t.Errorf("expected nil when disabled, got %+v", sh)
	}
}

func TestSignatureOutsideOperandNil(t *testing.T) {
	_, p, sp := createTestProviders()
	text := "++MOVE(M1)\n    DISTLIB()\n."
	doc := p.Parse(text)
	if sh := sp.GetSignatureHelp(doc, text, 0, 2); sh != nil {
		t.Errorf("expected nil outside operand, got %+v", sh)
	}
}

func TestSignatureAfterClosingParenNil(t *testing.T) {
	_, p, sp := createTestProviders()
	text := "++MOVE(M1)\n    DISTLIB() \n."
	doc := p.Parse(text)
	// cursor on line 1 at the trailing space AFTER "DISTLIB()" (col = len("    DISTLIB() "))
	line, char := 1, len("    DISTLIB() ")
	if sh := sp.GetSignatureHelp(doc, text, line, char); sh != nil {
		t.Errorf("expected nil when cursor is after the closing paren, got %+v", sh)
	}
}

func TestSignatureInsideParensWithValue(t *testing.T) {
	_, p, sp := createTestProviders()
	text := "++MOVE(M1)\n    DISTLIB(SMPLTS)\n."
	doc := p.Parse(text)
	// cursor between the parens, in the middle of the value: after "DISTLIB(SMP"
	line, char := 1, len("    DISTLIB(SMP")
	sh := sp.GetSignatureHelp(doc, text, line, char)
	if sh == nil {
		t.Fatal("expected signature help while typing inside DISTLIB(...), got nil")
	}
	if sh.Signatures[0].Label != "DISTLIB(DDNAME)" {
		t.Errorf("unexpected label: %q", sh.Signatures[0].Label)
	}
}
