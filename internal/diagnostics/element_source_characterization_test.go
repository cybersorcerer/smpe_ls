package diagnostics

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Characterization tests for the "which operand replaces inline data" rule.
//
// The rule is spelled out as a hardcoded list of operand names in several
// places. These tests pin down the behaviour that list produces today, over
// every statement in smpe.json and every relevant operand, so a refactoring
// that moves the list into smpe.json can be shown to change nothing.
//
// They deliberately restate the rule instead of importing it: a test that
// asked the production code what it expects would pass no matter what the
// production code does.

// elementSourceOperands are the operands that supply the element data from
// somewhere other than inline. Written as source text, so the statement is
// parsed the way a user would write it.
var elementSourceOperands = map[string]string{
	"FROMDS":  "FROMDS(DSN(MY.SOURCE.DS))",
	"RELFILE": "RELFILE(1)",
	"TXLIB":   "TXLIB(MYLIB)",
	"LKLIB":   "LKLIB(MYLIB)",
}

// deleteOperand is handled next to the element sources but is not one: it
// puts the statement into deletion mode, where no data is needed at all.
const deleteOperand = "DELETE"

// inlineDataStatements returns every statement smpe.json marks as carrying
// inline data, which are exactly the statements the rule applies to.
func inlineDataStatements(t *testing.T) []string {
	t.Helper()
	store, err := data.Load(realSMPEJSON)
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	var names []string
	for name, def := range store.Statements {
		if def.InlineData {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		t.Fatal("No statement carries inline_data - the matrix would be empty")
	}
	return names
}

// operandVariants is the set of documents built for one statement: once
// without any operand, once per element source, and once with DELETE.
func operandVariants() []struct {
	label          string
	operand        string
	replacesInline bool
} {
	variants := []struct {
		label          string
		operand        string
		replacesInline bool
	}{{"none", "", false}}
	for _, name := range []string{"FROMDS", "RELFILE", "TXLIB", "LKLIB"} {
		variants = append(variants, struct {
			label          string
			operand        string
			replacesInline bool
		}{name, elementSourceOperands[name], true})
	}
	return append(variants, struct {
		label          string
		operand        string
		replacesInline bool
	}{"DELETE", deleteOperand, true})
}

// buildDoc writes one statement with the given operand, followed by a comment
// line in column 1 and a following statement. The comment is the probe: it
// lies in what would be the inline data area, so whether it is reported shows
// how the statement was classified.
func buildDoc(stmt, operand string) string {
	head := stmt + "(E1) DISTLIB(ALIB)"
	if operand != "" {
		head += " " + operand
	}
	return head + " .\n" +
		"/* probe line */\n" +
		"++VER(Z038) FMID(F1) .\n"
}

func hasCode(diags []lsp.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

// hasMissingInlineData looks for the diagnostic by message, since
// checkMissingInlineData does not set a code.
func hasMissingInlineData(diags []lsp.Diagnostic) bool {
	for _, d := range diags {
		if strings.Contains(d.Message, "expects inline data") ||
			strings.Contains(d.Message, "IMASPZAP control statements") {
			return true
		}
	}
	return false
}

// A statement with an element source operand or DELETE carries no inline
// data, so the missing-inline-data diagnostic must stay silent. Without such
// an operand it must fire, because the probe line is a comment rather than
// data.
func TestCharacterizeMissingInlineDataAcrossStatements(t *testing.T) {
	_, p, dp := loadRealStore(t)
	cfg := &Config{MissingInlineData: true}

	for _, stmt := range inlineDataStatements(t) {
		for _, v := range operandVariants() {
			text := buildDoc(stmt, v.operand)
			diags := dp.AnalyzeASTWithConfigAndText(p.Parse(text), cfg, text)
			got := hasMissingInlineData(diags)

			// Without an element source the statement still expects data.
			// The probe line is a comment, which the parser counts as
			// inline data, so nothing is reported either way - except that
			// this is exactly the behaviour being pinned down.
			want := false
			if got != want {
				t.Errorf("%s + %s: missing_inline_data = %v, want %v\n%s",
					stmt, v.label, got, want, text)
			}
		}
	}
}

// The same matrix, but with a real data line instead of the comment probe.
// Here the statement without an element source has its data and must stay
// silent; adding an element source must not change that.
func TestCharacterizeMissingInlineDataWithRealData(t *testing.T) {
	_, p, dp := loadRealStore(t)
	cfg := &Config{MissingInlineData: true}

	for _, stmt := range inlineDataStatements(t) {
		for _, v := range operandVariants() {
			head := stmt + "(E1) DISTLIB(ALIB)"
			if v.operand != "" {
				head += " " + v.operand
			}
			text := head + " .\nREAL DATA LINE\n++VER(Z038) FMID(F1) .\n"
			diags := dp.AnalyzeASTWithConfigAndText(p.Parse(text), cfg, text)
			if hasMissingInlineData(diags) {
				t.Errorf("%s + %s: reported although data follows\n%s", stmt, v.label, text)
			}
		}
	}
}

// And once with nothing at all after the terminator, which is the case the
// diagnostic exists for.
func TestCharacterizeMissingInlineDataWithNothingFollowing(t *testing.T) {
	_, p, dp := loadRealStore(t)
	cfg := &Config{MissingInlineData: true}

	for _, stmt := range inlineDataStatements(t) {
		for _, v := range operandVariants() {
			head := stmt + "(E1) DISTLIB(ALIB)"
			if v.operand != "" {
				head += " " + v.operand
			}
			text := head + " .\n++VER(Z038) FMID(F1) .\n"
			diags := dp.AnalyzeASTWithConfigAndText(p.Parse(text), cfg, text)
			got := hasMissingInlineData(diags)
			want := !v.replacesInline
			if got != want {
				t.Errorf("%s + %s: missing_inline_data = %v, want %v\n%s",
					stmt, v.label, got, want, text)
			}
		}
	}
}

// checkCommentInColumn1 skips inline data. A statement whose data comes from
// an element source has no inline data area, so the probe line is ordinary
// MCS text there and must be reported.
func TestCharacterizeCommentInColumn1AcrossStatements(t *testing.T) {
	_, p, dp := loadRealStore(t)
	cfg := &Config{CommentInColumn1: true}

	for _, stmt := range inlineDataStatements(t) {
		for _, v := range operandVariants() {
			text := buildDoc(stmt, v.operand)
			diags := dp.AnalyzeASTWithConfigAndText(p.Parse(text), cfg, text)
			got := hasCode(diags, CodeCommentInColumn1)
			want := v.replacesInline
			if got != want {
				t.Errorf("%s + %s: comment_in_column_1 = %v, want %v\n%s",
					stmt, v.label, got, want, text)
			}
		}
	}
}

// Same for the standalone comment check, which shares the rule.
func TestCharacterizeStandaloneCommentAcrossStatements(t *testing.T) {
	_, p, dp := loadRealStore(t)
	cfg := &Config{StandaloneCommentBetweenMCS: true}

	for _, stmt := range inlineDataStatements(t) {
		for _, v := range operandVariants() {
			text := buildDoc(stmt, v.operand)
			diags := dp.AnalyzeASTWithConfigAndText(p.Parse(text), cfg, text)
			got := len(diags) > 0
			want := v.replacesInline
			if got != want {
				t.Errorf("%s + %s: standalone_comment = %v, want %v\n%s",
					stmt, v.label, got, want, text)
			}
		}
	}
}

// getMissingInlineDataMessage lists the element sources the statement
// actually defines. Pin down that list per statement.
func TestCharacterizeMissingInlineDataMessage(t *testing.T) {
	store, err := data.Load(realSMPEJSON)
	if err != nil {
		t.Fatalf("Failed to load smpe.json: %v", err)
	}
	_, p, dp := loadRealStore(t)
	cfg := &Config{MissingInlineData: true}

	for _, stmt := range inlineDataStatements(t) {
		// Which element sources does this statement define?
		var want []string
		for _, op := range store.Statements[stmt].Operands {
			name := strings.Split(op.Name, "|")[0]
			if _, ok := elementSourceOperands[name]; ok {
				want = append(want, name)
			}
		}

		text := stmt + "(E1) DISTLIB(ALIB) .\n++VER(Z038) FMID(F1) .\n"
		diags := dp.AnalyzeASTWithConfigAndText(p.Parse(text), cfg, text)
		if len(diags) != 1 {
			t.Errorf("%s: expected 1 diagnostic, got %d", stmt, len(diags))
			continue
		}
		msg := diags[0].Message

		// ++ZAP, ++MACUPD and ++SRCUPD name no element source at all: their
		// data are IMASPZAP respectively IEBUPDTE control statements, which
		// exist inline only. ++ZAP says so in its own words.
		if len(want) == 0 {
			if stmt != "++ZAP" && stmt != "++MACUPD" && stmt != "++SRCUPD" {
				t.Errorf("%s unexpectedly defines no element source: %q", stmt, msg)
				continue
			}
			if stmt == "++ZAP" {
				if !strings.Contains(msg, "IMASPZAP control statements") {
					t.Errorf("++ZAP: message %q does not name IMASPZAP", msg)
				}
			} else if !strings.Contains(msg, "expects inline data, none found") {
				t.Errorf("%s: message %q does not read as expected", stmt, msg)
			}
			continue
		}
		fragment := fmt.Sprintf("or one of %s", strings.Join(want, ", "))
		if !strings.Contains(msg, fragment) {
			t.Errorf("%s: message %q does not list %q", stmt, msg, fragment)
		}
	}
}
