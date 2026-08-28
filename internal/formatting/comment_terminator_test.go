package formatting

import (
	"strings"
	"testing"
)

// Tests for terminator detection around comments: a dot inside an operand
// value must not be mistaken for the statement terminator.

// A dot inside an operand value (e.g. a dataset name) must not make the
// formatter treat a following comment as a post-terminator comment, which
// would drop it from the output entirely.
func TestCommentAfterDottedValueIsPreserved(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++MAC(ABC)\n" +
		"    FROMDS(DSN(HLQ.MID.LLQ)) /* keep me */\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, "/* keep me */") {
		t.Errorf("Comment after dotted operand value was dropped:\n%s", result)
	}
}

// A comment that really follows the terminator must still be recognized and
// stay attached to the terminator line.
func TestCommentAfterTerminatorStaysOnTerminatorLine(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++VER(Z038) FMID(FXY1040) . /* trailing after dot */\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	lastLine := lines[len(lines)-1]
	if !strings.HasPrefix(lastLine, ". ") || !strings.Contains(lastLine, "/* trailing after dot */") {
		t.Errorf("Expected trailing comment on terminator line, got %q", lastLine)
	}
	if strings.Count(result, "/* trailing after dot */") != 1 {
		t.Errorf("Trailing comment duplicated:\n%s", result)
	}
}

// A dot inside a single-quoted operand value is not a terminator, so a comment
// behind it must survive.
func TestCommentAfterQuotedValueWithDotIsPreserved(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++USERMOD(RIQJ005)\n" +
		"    DESC('R+V IIQ. Started Tasks') /* keep me */\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, "/* keep me */") {
		t.Errorf("Comment after quoted value containing a dot was dropped:\n%s", result)
	}
}

// The same for a dot inside a list value that spans lines - the parenthesis
// depth carries across the line break.
func TestCommentAfterDottedValueAcrossLinesIsPreserved(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++VER(Z038)\n" +
		"    SUP(RIQJ000,\n" +
		"        RIQJ0.1) /* keep me too */\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	if !strings.Contains(result, "/* keep me too */") {
		t.Errorf("Comment after multi-line dotted value was dropped:\n%s", result)
	}
}

// The SMP/E reference shows a comment that opens before the period and is
// closed by "*/." - all of its lines must come through.
func TestCommentClosedRightBeforeTerminator(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++VER(Z038) FMID(FXY1040) /* comment after operand\n" +
		"   continued on subsequent\n" +
		"   records is okay.        */.\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	for _, want := range []string{"comment after operand", "continued on subsequent", "records is okay."} {
		if !strings.Contains(result, want) {
			t.Errorf("Lost comment text %q:\n%s", want, result)
		}
	}
	if !strings.Contains(result, "FMID(FXY1040)") {
		t.Errorf("Operand lost:\n%s", result)
	}
}

// A comment standing before the terminator must stay inside the statement and
// must not be pushed behind the period, where a second comment would be a
// syntax error.
func TestCommentBeforeTerminatorStaysBeforeIt(t *testing.T) {
	p, fp := newTestFormatter(t)
	input := "++VER(Z038)\n" +
		"    FMID(FXY1040)\n" +
		"    /* comment before the period */\n" +
		"  .\n"
	result := formatOnce(t, p, fp, input)
	t.Logf("Result:\n%s", result)

	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	if last != "." {
		t.Errorf("Expected bare terminator on the last line, got %q", last)
	}
	if !strings.Contains(result, "/* comment before the period */") {
		t.Errorf("Comment lost:\n%s", result)
	}
}
