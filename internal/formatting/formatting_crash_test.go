package formatting

import (
	"os"
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

func TestFormatIdempotent(t *testing.T) {
	input, err := os.ReadFile("../../test-files/test_required_validation.smpe")
	if err != nil {
		t.Fatal(err)
	}

	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatal(err)
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

	inputWithVer := string(input) + "++VER(Z038) PRE(U001 U002 U003) .\n"

	// First format
	doc1 := p.Parse(inputWithVer)
	edits1 := fp.FormatDocument(doc1, inputWithVer)
	if len(edits1) == 0 {
		t.Fatal("First format produced no edits")
	}
	firstFormat := edits1[0].NewText
	t.Logf("First format:\n%s", firstFormat)

	// Second format (should be idempotent)
	doc2 := p.Parse(firstFormat)
	t.Logf("Parsed %d statements from first format", len(doc2.Statements))
	edits2 := fp.FormatDocument(doc2, firstFormat)

	if len(edits2) == 0 {
		t.Log("PASS: Second format produced no edits (idempotent!)")
		return
	}

	secondFormat := edits2[0].NewText
	if firstFormat == secondFormat {
		t.Log("PASS: First and second format are identical (idempotent!)")
	} else {
		t.Errorf("FAIL: Formatting is NOT idempotent!\nFirst:\n%s\n\nSecond:\n%s", firstFormat, secondFormat)
	}
}

func TestFormatInlineCommentStaysWithOperand(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatal(err)
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

	// Multi-line input where comment is on the same line as FMID operand
	input := "++RELEASE(UZ12345)\n    FMID(FXY1040) /* my comment */\n    USER\n    REASON(CPU0A).\n"

	doc := p.Parse(input)
	edits := fp.FormatDocument(doc, input)
	if len(edits) == 0 {
		t.Fatal("No edits produced")
	}

	result := edits[0].NewText
	t.Logf("Formatted:\n%s", result)

	// The comment should stay with FMID, not move to the header
	if !strings.Contains(result, "FMID(FXY1040) /* my comment */") {
		t.Error("FAIL: Comment was not kept with its operand FMID")
	}
	if strings.Contains(result, "++RELEASE(UZ12345) /* my comment */") {
		t.Error("FAIL: Comment was wrongly moved to statement header")
	}

	// Second format should be stable
	doc2 := p.Parse(result)
	edits2 := fp.FormatDocument(doc2, result)
	if len(edits2) > 0 {
		result2 := edits2[0].NewText
		if result != result2 {
			t.Errorf("FAIL: Not idempotent!\nFirst:\n%s\n\nSecond:\n%s", result, result2)
		}
	}
}

func TestFormatMultilineCommentStaysWithOperand(t *testing.T) {
	store, err := data.Load("../../data/smpe.json")
	if err != nil {
		t.Fatal(err)
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

	// Multi-line comment on the same line as FMID operand
	input := "++RELEASE(UZ12345)\n    FMID(FXY1040) /* some comment\n    multiline */\n    USER\n    REASON(CPU0A).\n"

	doc := p.Parse(input)
	edits := fp.FormatDocument(doc, input)
	if len(edits) == 0 {
		t.Fatal("No edits produced")
	}

	result := edits[0].NewText
	t.Logf("Formatted:\n%s", result)

	// The multi-line comment should NOT be moved to before the terminator
	// It should stay near its operand FMID
	if strings.Contains(result, "REASON(CPU0A)\n  /* some comment") {
		t.Error("FAIL: Multi-line comment was moved away from FMID to before terminator")
	}

	// Second format should be stable
	doc2 := p.Parse(result)
	edits2 := fp.FormatDocument(doc2, result)
	if len(edits2) > 0 {
		result2 := edits2[0].NewText
		if result != result2 {
			t.Errorf("FAIL: Not idempotent!\nFirst:\n%s\n\nSecond:\n%s", result, result2)
		}
	}
}
