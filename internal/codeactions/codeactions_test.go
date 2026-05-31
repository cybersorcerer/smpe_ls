package codeactions

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

func diag(code string, data any, line, char, length int) lsp.Diagnostic {
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: line, Character: char},
			End:   lsp.Position{Line: line, Character: char + length},
		},
		Code: code,
		Data: data,
	}
}

// applyEdit applies a single-line TextEdit to text and returns the result.
func applyEdit(text string, e lsp.TextEdit) string {
	lines := strings.Split(text, "\n")
	ln := e.Range.Start.Line
	line := lines[ln]
	line = line[:e.Range.Start.Character] + e.NewText + line[e.Range.End.Character:]
	lines[ln] = line
	return strings.Join(lines, "\n")
}

func TestMissingTerminatorFix(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	// No Data payload: falls back to the diagnostic range end (statement header).
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{diag(diagnostics.CodeMissingTerminator, nil, 0, 0, 6)},
	}
	actions := p.GetCodeActions(uri, "", ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Add statement terminator" {
		t.Errorf("unexpected title: %q", actions[0].Title)
	}
	edits := actions[0].Edit.Changes[uri]
	if len(edits) != 1 || edits[0].NewText != "\n." {
		t.Errorf("expected single newline '.' edit, got %+v", edits)
	}
	if edits[0].Range.Start != (lsp.Position{Line: 0, Character: 6}) {
		t.Errorf("expected fallback at line 0 char 6, got %+v", edits[0].Range.Start)
	}
}

// TestMissingTerminatorFixMultiline verifies the '.' is inserted on a new
// line at the end of the statement's last line (from Data), not after the
// statement header.
func TestMissingTerminatorFixMultiline(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	src := "++USERMOD(U1)\n    REWORK(2024366)"
	// Statement header on line 0; last operand is on line 1.
	data := map[string]interface{}{"endLine": float64(1)}
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{diag(diagnostics.CodeMissingTerminator, data, 0, 0, 6)},
	}
	actions := p.GetCodeActions(uri, src, ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	edits := actions[0].Edit.Changes[uri]
	if len(edits) != 1 || edits[0].NewText != "\n." {
		t.Fatalf("expected single newline '.' edit, got %+v", edits)
	}
	// Inserted at the end of line 1 (length of "    REWORK(2024366)" = 19).
	want := lsp.Position{Line: 1, Character: 19}
	if edits[0].Range.Start != want {
		t.Errorf("expected edit at %+v, got %+v", want, edits[0].Range.Start)
	}
}

func TestReworkFix(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	src := "++USERMOD(U1)\n    REWORK()\n    ."
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeEmptyOperandParameter, map[string]any{"operand": "REWORK"}, 1, 4, 6),
		},
	}
	actions := p.GetCodeActions(uri, src, ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Set REWORK to current date" {
		t.Errorf("unexpected title: %q", actions[0].Title)
	}
	edits := actions[0].Edit.Changes[uri]
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if !regexp.MustCompile(`^\d{7}$`).MatchString(edits[0].NewText) {
		t.Errorf("expected 7-digit Julian date as NewText, got %q", edits[0].NewText)
	}
}

func TestReworkFixAppliedNoStrayParens(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	src := "++USERMOD(U1)\n    REWORK()\n    ."
	// REWORK starts at line 1, char 4; name length 6 -> Range end char 10
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeEmptyOperandParameter, map[string]any{"operand": "REWORK"}, 1, 4, 6),
		},
	}
	actions := p.GetCodeActions(uri, src, ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	got := applyEdit(src, actions[0].Edit.Changes[uri][0])
	if !regexp.MustCompile(`(?m)^    REWORK\(\d{7}\)$`).MatchString(got) {
		t.Errorf("applied result has stray/missing parens:\n%s", got)
	}
	if strings.Contains(got, "()") {
		t.Errorf("stray empty parens remain:\n%s", got)
	}
}

func TestEmptyOperandNonReworkNoFix(t *testing.T) {
	p := NewProvider()
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeEmptyOperandParameter, map[string]any{"operand": "FMID"}, 1, 4, 4),
		},
	}
	if got := p.GetCodeActions("file:///t.smpe", "", ctx); len(got) != 0 {
		t.Errorf("expected no actions for non-REWORK empty operand, got %d", len(got))
	}
}

func TestSingleMissingOperandFix(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeMissingRequiredOperand, map[string]any{"operand": "SOURCEID"}, 0, 0, 8),
		},
	}
	actions := p.GetCodeActions(uri, "", ctx)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Insert operand SOURCEID" {
		t.Errorf("unexpected title: %q", actions[0].Title)
	}
	if got := actions[0].Edit.Changes[uri][0].NewText; got != "\n    SOURCEID()" {
		t.Errorf("unexpected edit text: %q", got)
	}
}

// TestSingleMissingOperandFixPosition verifies the operand skeleton is inserted
// after the full statement header (after the name operand, e.g. "(S1)"), not
// after the statement name token (e.g. "++SRC"). The diagnostic Range covers
// only the name token, so the fix must rely on the endLine Data payload. This
// is generic across all element MCS that take an inline name operand
// (++HELP(member), ++BOOK(member), …) and multi-line statements.
func TestSingleMissingOperandFixPosition(t *testing.T) {
	uri := "file:///t.smpe"
	tests := []struct {
		name     string
		src      string
		operand  string
		nameLen  int // length of the statement name token (diagnostic Range)
		endLine  int
		expected string
	}{
		{
			name:     "++SRC inline name operand",
			src:      "++SRC(S1)\n.",
			operand:  "DISTLIB",
			nameLen:  5,
			endLine:  0,
			expected: "++SRC(S1)\n    DISTLIB()\n.",
		},
		{
			name:     "++HELP element mcs",
			src:      "++HELP(MEMBER1)\n.",
			operand:  "DISTLIB",
			nameLen:  6,
			endLine:  0,
			expected: "++HELP(MEMBER1)\n    DISTLIB()\n.",
		},
		{
			name:     "++ASSIGN with operand on continuation line",
			src:      "++ASSIGN\n    SOURCEID(ID1)\n.",
			operand:  "TO",
			nameLen:  8,
			endLine:  1,
			expected: "++ASSIGN\n    SOURCEID(ID1)\n    TO()\n.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProvider()
			ctx := lsp.CodeActionContext{
				Diagnostics: []lsp.Diagnostic{
					diag(diagnostics.CodeMissingRequiredOperand,
						map[string]any{"operand": tc.operand, "endLine": float64(tc.endLine)},
						0, 0, tc.nameLen),
				},
			}
			actions := p.GetCodeActions(uri, tc.src, ctx)
			if len(actions) != 1 {
				t.Fatalf("expected 1 action, got %d", len(actions))
			}
			edit := actions[0].Edit.Changes[uri][0]
			if got := applyEdit(tc.src, edit); got != tc.expected {
				t.Errorf("operand inserted at wrong position\n got: %q\nwant: %q", got, tc.expected)
			}
		})
	}
}

func TestMissingOperandNilDataNoFix(t *testing.T) {
	p := NewProvider()
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeMissingRequiredOperand, nil, 0, 0, 8),
		},
	}
	if got := p.GetCodeActions("file:///t.smpe", "", ctx); len(got) != 0 {
		t.Errorf("expected no actions for missing-operand with nil Data, got %d", len(got))
	}
}

func TestMultipleMissingOperandsAddInsertAll(t *testing.T) {
	p := NewProvider()
	uri := "file:///t.smpe"
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeMissingRequiredOperand, map[string]any{"operand": "SOURCEID"}, 0, 0, 8),
			diag(diagnostics.CodeMissingRequiredOperand, map[string]any{"operand": "TO"}, 0, 0, 8),
		},
	}
	actions := p.GetCodeActions(uri, "", ctx)
	if len(actions) != 3 { // 2 single + 1 insert-all
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}
	var all *lsp.CodeAction
	for i := range actions {
		if actions[i].Title == "Insert all required operands" {
			all = &actions[i]
		}
	}
	if all == nil {
		t.Fatal("missing 'Insert all required operands' action")
	}
	if got := all.Edit.Changes[uri][0].NewText; got != "\n    SOURCEID()\n    TO()" {
		t.Errorf("unexpected insert-all text: %q", got)
	}
}

// TestMoveInsertOperandsJSONRoundtrip verifies the fix payload survives a JSON
// marshal/unmarshal cycle, which mirrors the real LSP wire path: the server
// sends diagnostics as JSON and the client echoes them back in the codeAction
// request, where Data is decoded into []interface{}/map[string]interface{}.
// The native-typed table tests do not exercise that decoded shape.
func TestMoveInsertOperandsJSONRoundtrip(t *testing.T) {
	uri := "file:///t.smpe"
	src := "++MOVE(M1)\n."
	// Build a diagnostic the way the diagnostics provider does (native types).
	d := diag(diagnostics.CodeMoveInsertOperands, map[string]any{
		"endLine": 0,
		"fixes": []map[string]any{
			{"title": "Insert DISTLIB mode operands", "operands": []map[string]any{
				{"name": "DISTLIB", "parens": true},
				{"name": "TODISTLIB", "parens": true},
				{"name": "MAC", "parens": false},
			}},
			{"title": "Insert SYSLIB mode operands", "operands": []map[string]any{
				{"name": "SYSLIB", "parens": true},
				{"name": "TOSYSLIB", "parens": true},
				{"name": "MAC", "parens": false},
			}},
		},
	}, 0, 0, 6)

	// Round-trip the diagnostic through JSON, as happens over the LSP wire.
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var wired lsp.Diagnostic
	if err := json.Unmarshal(raw, &wired); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	p := NewProvider()
	actions := p.GetCodeActions(uri, src, lsp.CodeActionContext{Diagnostics: []lsp.Diagnostic{wired}})
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions after JSON roundtrip, got %d", len(actions))
	}
	got := applyEdit(src, actions[0].Edit.Changes[uri][0])
	want := "++MOVE(M1)\n    DISTLIB()\n    TODISTLIB()\n    MAC\n."
	if got != want {
		t.Errorf("roundtrip edit wrong\n got: %q\nwant: %q", got, want)
	}
}

// TestMoveInsertOperandsFixes verifies move quick fixes: one CodeAction per
// payload "fixes" entry, each inserting its operands at the statement header end.
// Boolean flag operands (parens=false) insert without "()".
func TestMoveInsertOperandsFixes(t *testing.T) {
	uri := "file:///t.smpe"
	mkFix := func(title string, ops ...[2]any) map[string]any {
		var list []map[string]any
		for _, o := range ops {
			list = append(list, map[string]any{"name": o[0], "parens": o[1]})
		}
		return map[string]any{"title": title, "operands": list}
	}
	tests := []struct {
		name      string
		src       string
		endLine   int
		fixes     []map[string]any
		wantCount int
		applyIdx  int
		expected  string
	}{
		{
			name:    "no mode: two complete fixes",
			src:     "++MOVE(M1)\n.",
			endLine: 0,
			fixes: []map[string]any{
				mkFix("Insert DISTLIB mode operands",
					[2]any{"DISTLIB", true}, [2]any{"TODISTLIB", true}, [2]any{"MAC", false}),
				mkFix("Insert SYSLIB mode operands",
					[2]any{"SYSLIB", true}, [2]any{"TOSYSLIB", true}, [2]any{"MAC", false}),
			},
			wantCount: 2,
			applyIdx:  0,
			expected:  "++MOVE(M1)\n    DISTLIB()\n    TODISTLIB()\n    MAC\n.",
		},
		{
			name:    "missing TODISTLIB: single fix",
			src:     "++MOVE(M1)\n    DISTLIB(D1)\n.",
			endLine: 1,
			fixes: []map[string]any{
				mkFix("Insert operand TODISTLIB", [2]any{"TODISTLIB", true}),
			},
			wantCount: 1,
			applyIdx:  0,
			expected:  "++MOVE(M1)\n    DISTLIB(D1)\n    TODISTLIB()\n.",
		},
		{
			name:    "missing element: three flag fixes without parens",
			src:     "++MOVE(M1)\n    DISTLIB(D1) TODISTLIB(T1)\n.",
			endLine: 1,
			fixes: []map[string]any{
				mkFix("Insert operand MAC", [2]any{"MAC", false}),
				mkFix("Insert operand MOD", [2]any{"MOD", false}),
				mkFix("Insert operand SRC", [2]any{"SRC", false}),
			},
			wantCount: 3,
			applyIdx:  0,
			expected:  "++MOVE(M1)\n    DISTLIB(D1) TODISTLIB(T1)\n    MAC\n.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProvider()
			ctx := lsp.CodeActionContext{
				Diagnostics: []lsp.Diagnostic{
					diag(diagnostics.CodeMoveInsertOperands,
						map[string]any{"endLine": float64(tc.endLine), "fixes": tc.fixes},
						0, 0, 6),
				},
			}
			actions := p.GetCodeActions(uri, tc.src, ctx)
			if len(actions) != tc.wantCount {
				t.Fatalf("expected %d actions, got %d", tc.wantCount, len(actions))
			}
			edit := actions[tc.applyIdx].Edit.Changes[uri][0]
			if got := applyEdit(tc.src, edit); got != tc.expected {
				t.Errorf("wrong edit\n got: %q\nwant: %q", got, tc.expected)
			}
		})
	}
}

// TestMoveInsertOperandsEmptyNoFix verifies an empty/absent fixes payload
// produces no actions.
func TestMoveInsertOperandsEmptyNoFix(t *testing.T) {
	p := NewProvider()
	ctx := lsp.CodeActionContext{
		Diagnostics: []lsp.Diagnostic{
			diag(diagnostics.CodeMoveInsertOperands,
				map[string]any{"endLine": float64(0), "fixes": []map[string]any{}}, 0, 0, 6),
		},
	}
	if got := p.GetCodeActions("file:///t.smpe", "++MOVE(M1)\n.", ctx); len(got) != 0 {
		t.Errorf("expected no actions for empty fixes, got %d", len(got))
	}
}
