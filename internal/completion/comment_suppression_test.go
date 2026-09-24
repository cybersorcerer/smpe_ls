package completion

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Comment text is prose, not MCS. Offering statements or operands inside it
// puts a popup in the way of someone writing a sentence - every space in a
// comment would open the list again.

// posOf returns the position just after the given marker text.
func posOf(t *testing.T, text, marker string) (int, int) {
	t.Helper()
	for i, line := range strings.Split(text, "\n") {
		if idx := strings.Index(line, marker); idx >= 0 {
			return i, len([]rune(line[:idx+len(marker)]))
		}
	}
	t.Fatalf("marker %q not found", marker)
	return 0, 0
}

const commentDoc = "++USERMOD(U1)\n" +
	"    /* a comment with several   blanks   here */\n" +
	"    DESC(Test)\n" +
	"    /*\n" +
	"     * a longer comment   with blanks\n" +
	"     */\n" +
	"    .\n"

func TestNoCompletionInsideSingleLineComment(t *testing.T) {
	line, char := posOf(t, commentDoc, "a comment with several")
	for _, tk := range []int{lsp.CompletionTriggerInvoked, lsp.CompletionTriggerCharacter} {
		if items := completionsAtKind(t, commentDoc, line, char, tk); len(items) != 0 {
			t.Errorf("triggerKind %d: %d completions inside a comment, want none: %v",
				tk, len(items), items)
		}
	}
}

func TestNoCompletionInsideMultiLineComment(t *testing.T) {
	line, char := posOf(t, commentDoc, "a longer comment")
	for _, tk := range []int{lsp.CompletionTriggerInvoked, lsp.CompletionTriggerCharacter} {
		if items := completionsAtKind(t, commentDoc, line, char, tk); len(items) != 0 {
			t.Errorf("triggerKind %d: %d completions inside a comment, want none: %v",
				tk, len(items), items)
		}
	}
}

// The comment markers themselves count as comment, so typing right after "/*"
// stays quiet too.
func TestNoCompletionRightAfterCommentStart(t *testing.T) {
	text := "++USERMOD(U1)\n    /* \n    .\n"
	line, char := posOf(t, text, "/* ")
	if items := completionsAtKind(t, text, line, char, lsp.CompletionTriggerCharacter); len(items) != 0 {
		t.Errorf("%d completions right after '/*', want none: %v", len(items), items)
	}
}

// Outside the comment the operand completion has to keep working, otherwise
// the fix would take away what people rely on.
func TestCompletionStillWorksOutsideComments(t *testing.T) {
	line, char := posOf(t, commentDoc, "DESC(Test)")
	if items := completionsAtKind(t, commentDoc, line, char, lsp.CompletionTriggerInvoked); len(items) == 0 {
		t.Error("No completions after an operand, expected the remaining operands")
	}
}

// A comment that ended earlier on the same line must not silence the rest.
func TestCompletionAfterCommentEndsOnSameLine(t *testing.T) {
	text := "++USERMOD(U1)\n    /* note */ \n    .\n"
	line, char := posOf(t, text, "/* note */ ")
	if items := completionsAtKind(t, text, line, char, lsp.CompletionTriggerInvoked); len(items) == 0 {
		t.Error("No completions after a closed comment, expected the remaining operands")
	}
}

// The tests above use the small fixture from completion_test.go, where ++VER
// deliberately has no operands. This one runs against the real smpe.json so a
// suppression that silences everything cannot pass unnoticed.
func TestCommentSuppressionAgainstRealData(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := parser.NewParser(store.Statements)
	cp := NewProvider(store)

	text := "++VER(Z038)\n" +
		"    /* a comment with several   blanks */\n" +
		"    FMID(HBB77E0)\n" +
		"    .\n"
	doc := p.Parse(text)

	// The end of the comment line, i.e. just past the closing "*/".
	afterComment := len([]rune(strings.Split(text, "\n")[1]))

	cases := []struct {
		line, char int
		wantAny    bool
		what       string
	}{
		{1, 20, false, "inside the comment"},
		{1, 34, false, "after several blanks in the comment"},
		{1, afterComment, true, "after the comment ended"},
		{2, 17, true, "after FMID(HBB77E0)"},
	}
	for _, tc := range cases {
		got := cp.GetCompletionsAST(doc, text, tc.line, tc.char, lsp.CompletionTriggerInvoked)
		if tc.wantAny && len(got) == 0 {
			t.Errorf("%s: no completions, expected the remaining operands", tc.what)
		}
		if !tc.wantAny && len(got) != 0 {
			t.Errorf("%s: %d completions, want none", tc.what, len(got))
		}
	}
}

// Whether the statement list opens, and what it replaces when an item is
// accepted, must not depend on the case being typed. The offered names are
// upper case either way; only the range to replace is at stake, and without
// it the typed prefix would stay in front of the inserted name.
func TestLowerCasePrefixBehavesLikeUpperCase(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	p := parser.NewParser(store.Statements)
	cp := NewProvider(store)

	for _, pair := range []struct{ upper, lower string }{
		{"++D", "++d"},
		{"++SRC", "++src"},
	} {
		up := cp.GetCompletionsAST(p.Parse(pair.upper), pair.upper, 0, len(pair.upper), lsp.CompletionTriggerCharacter)
		lo := cp.GetCompletionsAST(p.Parse(pair.lower), pair.lower, 0, len(pair.lower), lsp.CompletionTriggerCharacter)

		if len(up) != len(lo) {
			t.Errorf("%q offers %d items, %q offers %d", pair.upper, len(up), pair.lower, len(lo))
			continue
		}
		if len(lo) == 0 {
			t.Errorf("%q offers nothing", pair.lower)
			continue
		}
		if lo[0].TextEdit == nil {
			t.Errorf("%q has no replacement range, the typed prefix would stay in the line", pair.lower)
			continue
		}
		if up[0].TextEdit == nil {
			t.Fatalf("%q has no replacement range either - the test premise is wrong", pair.upper)
		}
		if lo[0].TextEdit.Range != up[0].TextEdit.Range {
			t.Errorf("%q replaces %+v, %q replaces %+v",
				pair.lower, lo[0].TextEdit.Range, pair.upper, up[0].TextEdit.Range)
		}
	}
}

func TestIsTypingMCSPrefixAcceptsBothCases(t *testing.T) {
	cases := map[string]bool{
		"+": true, "++": true, "++S": true, "++s": true,
		"++ASSIGN": true, "++assign": true, "++AsSiGn": true,
		"++ASSIGN ": false, "++++": false, "abc": false, "": false,
	}
	for in, want := range cases {
		if got := isTypingMCSPrefix(in); got != want {
			t.Errorf("isTypingMCSPrefix(%q) = %v, want %v", in, got, want)
		}
	}
}
