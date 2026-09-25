package completion

import (
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// An accepted item has to replace the fragment already typed. Without an
// explicit range the editor works it out, and it only recognises the fragment
// when it matches the item case-sensitively - typing "s" and accepting SUP
// produced "sSUP" while "S" worked. Shipped that way in 1.3.16.

func realProviders(t *testing.T) (*parser.Parser, *Provider) {
	t.Helper()
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	return parser.NewParser(store.Statements), NewProvider(store)
}

// applied returns the line as it would read after accepting the named item.
func applied(t *testing.T, text, label string) string {
	t.Helper()
	p, cp := realProviders(t)
	runes := []rune(text)
	items := cp.GetCompletionsAST(p.Parse(text), text, 0, len(runes), lsp.CompletionTriggerCharacter)
	for _, it := range items {
		if it.Label != label {
			continue
		}
		if it.TextEdit == nil {
			t.Fatalf("%q: item %q carries no replacement range", text, label)
		}
		return string(runes[:it.TextEdit.Range.Start.Character]) + it.TextEdit.NewText
	}
	t.Fatalf("%q: item %q not offered among %d items", text, label, len(items))
	return ""
}

func TestAcceptingAnOperandReplacesTheTypedFragment(t *testing.T) {
	for _, tc := range []struct{ text, label, want string }{
		{"++VER(Z038) S", "SUP", "++VER(Z038) SUP($0)"},
		{"++VER(Z038) s", "SUP", "++VER(Z038) SUP($0)"},
		{"++VER(Z038) su", "SUP", "++VER(Z038) SUP($0)"},
		{"++USERMOD(U1) d", "DESCRIPTION", "++USERMOD(U1) DESCRIPTION($0)"},
	} {
		if got := applied(t, tc.text, tc.label); got != tc.want {
			t.Errorf("%q + %s = %q, want %q", tc.text, tc.label, got, tc.want)
		}
	}
}

func TestAcceptingAValueReplacesTheTypedFragment(t *testing.T) {
	for _, tc := range []struct{ text, label, want string }{
		{"++HOLD(H1) CLASS(E", "ERREL", "++HOLD(H1) CLASS(ERREL"},
		{"++HOLD(H1) CLASS(e", "ERREL", "++HOLD(H1) CLASS(ERREL"},
	} {
		if got := applied(t, tc.text, tc.label); got != tc.want {
			t.Errorf("%q + %s = %q, want %q", tc.text, tc.label, got, tc.want)
		}
	}
}

// With nothing typed there is nothing to replace, and the item may come
// without a range - the editor then simply inserts at the cursor.
func TestNoFragmentMeansNoRange(t *testing.T) {
	p, cp := realProviders(t)
	text := "++VER(Z038) "
	items := cp.GetCompletionsAST(p.Parse(text), text, 0, len([]rune(text)), lsp.CompletionTriggerCharacter)
	if len(items) == 0 {
		t.Fatal("no operands offered")
	}
	for _, it := range items {
		if it.TextEdit != nil {
			t.Errorf("item %q carries a range although nothing was typed: %+v", it.Label, it.TextEdit.Range)
		}
	}
}

func TestTypedWordRange(t *testing.T) {
	cases := []struct {
		text       string
		line, char int
		wantStart  int
		wantNil    bool
	}{
		{"++VER(Z038) SUP", 0, 15, 12, false},
		{"++VER(Z038) sup", 0, 15, 12, false},
		{"++VER(Z038) ", 0, 12, 0, true},
		{"++HOLD(H1) CLASS(er", 0, 19, 17, false},
		{"", 0, 0, 0, true},
	}
	for _, tc := range cases {
		got := typedWordRange(tc.text, tc.line, tc.char)
		if tc.wantNil {
			if got != nil {
				t.Errorf("typedWordRange(%q, %d) = %+v, want nil", tc.text, tc.char, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("typedWordRange(%q, %d) = nil, want start %d", tc.text, tc.char, tc.wantStart)
			continue
		}
		if got.Start.Character != tc.wantStart || got.End.Character != tc.char {
			t.Errorf("typedWordRange(%q, %d) = %d-%d, want %d-%d",
				tc.text, tc.char, got.Start.Character, got.End.Character, tc.wantStart, tc.char)
		}
	}
}
