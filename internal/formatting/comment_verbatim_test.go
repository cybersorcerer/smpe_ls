package formatting

import (
	"strings"
	"testing"
)

// Tests for comment text preservation: the formatter reproduces comments as
// written and only ever changes where they sit, never what they say.

// containsBlock reports whether all lines of want appear consecutively and
// byte-identically in got.
func containsBlock(got, want string) bool {
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")
	for i := 0; i+len(wantLines) <= len(gotLines); i++ {
		match := true
		for j, w := range wantLines {
			if gotLines[i+j] != w {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// The formatter must reproduce a multi-line comment verbatim - including lines
// the user left past column 72. Rewriting them is reported by diagnostics, not
// silently applied.
func TestMultiLineCommentIsPreservedVerbatim(t *testing.T) {
	p, fp := newTestFormatter(t)
	block := "  /* multi line comment start that is quite long and definitely goes past column seventy two\n" +
		"     second line\n" +
		"     third line */"
	input := "++VER(Z038)\n" +
		"    FMID(FXY1040)\n" +
		block + "\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !containsBlock(result, block) {
		t.Errorf("Comment block was rewritten.\nwant verbatim:\n%s\ngot:\n%s", block, result)
	}
}

// A multi-line comment placed before the first operand must survive verbatim
// instead of being collapsed into a single reflowed block.
func TestMultiLineLeadingCommentKeepsLineStructure(t *testing.T) {
	p, fp := newTestFormatter(t)
	block := "  /* block before first operand with a lot of text here\n" +
		"     second line here also carrying plenty of words\n" +
		"     third line closes it */"
	input := "++VER(Z038)\n" +
		block + "\n" +
		"    FMID(FXY1040)\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !containsBlock(result, block) {
		t.Errorf("Comment block was rewritten.\nwant verbatim:\n%s\ngot:\n%s", block, result)
	}
}

// A box-drawing header comment whose lines are laid out for exactly columns
// 1-72 must come through the formatter untouched.
func TestBoxCommentSurvivesFormatting(t *testing.T) {
	p, fp := newTestFormatter(t)
	block := "/*\n" +
		"+-- GITLAB-META-START --------------------------------------------------\n" +
		"| GIT_REPOSITORY_ID     : <<<CI_PROJECT_ID>>>\n" +
		"|   <<<CI_COMMIT_SHA>>>\n" +
		"+-- GITLAB-META-ENDE ---------------------------------------------------\n" +
		"*/"
	input := "++USERMOD(RIQJ005)\n" +
		block + "\n" +
		"    DESC('R+V IIQ Started Tasks JCL')\n" +
		"    REWORK(2026239)\n" +
		".\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !containsBlock(result, block) {
		t.Errorf("Box comment was rewritten.\nwant verbatim:\n%s\ngot:\n%s", block, result)
	}
}

// A comment the formatter relocates into the statement must not be left in
// column 1 - otherwise the formatter creates the very error the
// commentInColumn1 diagnostic reports. The whole block shifts by one amount so
// its relative layout is kept.
func TestMovedCommentIsShiftedOutOfColumn1(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "/* leading block in column 1\n" +
		"   second line of the block */\n" +
		"++VER(Z038) FMID(FXY1040) .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	want := "  /* leading block in column 1\n" +
		"     second line of the block */"
	if !containsBlock(result, want) {
		t.Errorf("Moved block not shifted as a unit.\nwant:\n%s\ngot:\n%s", want, result)
	}
	for i, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(line, "/*") {
			t.Errorf("Line %d still starts in column 1: %q", i, line)
		}
	}
}

// A single-line comment that sat behind an operand and no longer fits moves
// onto its own line. Its former column is meaningless there and must not be
// reproduced - doing so pads the line out for nothing.
func TestRelocatedSingleLineCommentGetsStandardIndent(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++VER(Z038)\n" +
		"    FMID(FXY1040) /* this single line comment is deliberately far too long for column seventy two */\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	want := "  /* this single line comment is deliberately far too long for column seventy two */"
	if !containsBlock(result, want) {
		t.Errorf("Relocated comment not placed at the standard indent:\n%s", result)
	}
}

// A comment that already stood on a line of its own keeps the column the user
// gave it.
func TestStandaloneCommentKeepsItsIndent(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++VER(Z038)\n" +
		"     /* comment indented by five */\n" +
		"    FMID(FXY1040)\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !containsBlock(result, "     /* comment indented by five */") {
		t.Errorf("Standalone comment lost its indentation:\n%s", result)
	}
}

// Formatting formatted output must not change it again, for every comment
// shape the formatter handles.
func TestCommentFormattingIsIdempotent(t *testing.T) {
	p, fp := newTestFormatter(t)

	cases := map[string]string{
		"header comment":     "++VER(Z038) /* on header */\n    FMID(FXY1040)\n  .\n",
		"operand comment":    "++VER(Z038)\n    FMID(FXY1040) /* on operand */\n  .\n",
		"overlong comment":   "++VER(Z038)\n    FMID(FXY1040) /* this single line comment is deliberately far too long for column seventy two */\n  .\n",
		"standalone comment": "++VER(Z038)\n     /* indented by five */\n    FMID(FXY1040)\n  .\n",
		"block comment":      "++VER(Z038)\n  /* block opens\n     second line\n     closes */\n    FMID(FXY1040)\n  .\n",
		"box comment":        "++USERMOD(RIQJ005)\n/*\n+-- META-START -----------------------------------------------------\n| KEY : VALUE\n+-- META-ENDE ------------------------------------------------------\n*/\n    REWORK(2026239)\n.\n",
		"after terminator":   "++VER(Z038) FMID(FXY1040) . /* after dot */\n",
		"moved comment":      "/* leading block\n   second line */\n++VER(Z038) FMID(FXY1040) .\n",
	}

	for name, input := range cases {
		first := formatOnce(t, p, fp, input)
		second := formatOnce(t, p, fp, first)
		if first != second {
			t.Errorf("%s: not idempotent.\nfirst:\n%s\nsecond:\n%s", name, first, second)
		}
	}
}
