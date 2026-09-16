package completion

import (
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// completionsAt returns the labels offered at line/character in text, as if
// the user had asked explicitly (Ctrl+Space).
func completionsAt(t *testing.T, text string, line, character int) []string {
	t.Helper()
	return completionsAtKind(t, text, line, character, lsp.CompletionTriggerInvoked)
}

// completionsAtKind is completionsAt with an explicit trigger kind.
func completionsAtKind(t *testing.T, text string, line, character, kind int) []string {
	t.Helper()
	_, p, cp := createTestProviders()
	doc := p.Parse(text)
	var labels []string
	for _, item := range cp.GetCompletionsAST(doc, text, line, character, kind) {
		labels = append(labels, item.Label)
	}
	return labels
}

func containsLabel(labels []string, want string) bool {
	for _, l := range labels {
		if l == want {
			return true
		}
	}
	return false
}

// After a terminated statement the cursor is no longer inside it, so indenting
// a fresh line must offer MCS statements - not the operands of the statement
// above. The continuation check only asked whether any statement started
// earlier, never whether it was still open.
func TestCompletionAfterTerminatorOffersStatements(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"    REWORK(2026259)\n" +
		".\n" +
		"   \n"
	labels := completionsAt(t, text, 3, 3)
	t.Logf("labels: %v", labels)

	if containsLabel(labels, "REWORK") || containsLabel(labels, "DESC") {
		t.Errorf("Operands of the finished statement were offered: %v", labels)
	}
	if !containsLabel(labels, "++USERMOD") {
		t.Errorf("Expected MCS statements after a terminated statement, got %v", labels)
	}
}

// Typing "++" after leading whitespace must switch to statement completion.
func TestCompletionAfterTerminatorWithPlusPlus(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"    REWORK(2026259)\n" +
		".\n" +
		"   ++\n"
	labels := completionsAt(t, text, 3, 5)
	t.Logf("labels: %v", labels)

	if !containsLabel(labels, "++USERMOD") {
		t.Errorf("Expected MCS statements while typing ++, got %v", labels)
	}
}

// An indented line inside a statement that has no terminator yet is a real
// continuation line and must keep offering that statement's operands.
func TestCompletionInsideOpenStatementOffersOperands(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"    \n"
	labels := completionsAt(t, text, 1, 4)
	t.Logf("labels: %v", labels)

	if !containsLabel(labels, "REWORK") {
		t.Errorf("Expected operands inside the open statement, got %v", labels)
	}
	if containsLabel(labels, "++USERMOD") {
		t.Errorf("MCS statements offered inside an open statement: %v", strings.Join(labels, ","))
	}
}

// Typing a space on a free line must not pop up the statement list: the client
// asked only because the space is a trigger character.
func TestCompletionSpaceOnEmptyLineStaysQuiet(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"    REWORK(2026259)\n" +
		".\n" +
		"   \n"
	labels := completionsAtKind(t, text, 3, 3, lsp.CompletionTriggerCharacter)
	if len(labels) != 0 {
		t.Errorf("Expected no completions for a whitespace trigger, got %v", labels)
	}
}

// Asking explicitly on the same line still yields the statements.
func TestCompletionCtrlSpaceOnEmptyLineOffersStatements(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"    REWORK(2026259)\n" +
		".\n" +
		"   \n"
	labels := completionsAtKind(t, text, 3, 3, lsp.CompletionTriggerInvoked)
	if !containsLabel(labels, "++USERMOD") {
		t.Errorf("Expected statements on explicit request, got %v", labels)
	}
}

// Once a "+" is typed the prefix is no longer whitespace, so the list opens
// even though "+" is itself a trigger character.
func TestCompletionPlusAfterWhitespaceOffersStatements(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"    REWORK(2026259)\n" +
		".\n" +
		"   ++\n"
	labels := completionsAtKind(t, text, 3, 5, lsp.CompletionTriggerCharacter)
	if !containsLabel(labels, "++USERMOD") {
		t.Errorf("Expected statements while typing ++, got %v", labels)
	}
}

// Inside an open statement a space keeps triggering operand completion.
func TestCompletionSpaceInsideStatementOffersOperands(t *testing.T) {
	text := "++USERMOD(U1)\n" +
		"    \n"
	labels := completionsAtKind(t, text, 1, 4, lsp.CompletionTriggerCharacter)
	if !containsLabel(labels, "REWORK") {
		t.Errorf("Expected operands inside the open statement, got %v", labels)
	}
}
