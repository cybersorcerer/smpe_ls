// Package codeactions provides LSP quick fixes for SMP/E diagnostics.
package codeactions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

const kindQuickFix = "quickfix"

// Provider builds code actions from diagnostics.
type Provider struct{}

// NewProvider creates a code actions provider.
func NewProvider() *Provider {
	return &Provider{}
}

// GetCodeActions returns quick fixes for the diagnostics in the request context,
// plus any cursor-based actions available at reqRange (e.g. refreshing a
// non-empty REWORK value). text is the full document text; it is used to
// locate parentheses for operand fixes. doc is the parsed AST for the same
// text, used for cursor-based lookups. It never returns nil (serializes as
// [] not null).
func (p *Provider) GetCodeActions(uri, text string, doc *parser.Document, reqRange lsp.Range, ctx lsp.CodeActionContext) []lsp.CodeAction {
	actions := []lsp.CodeAction{}
	var missingOperand []lsp.Diagnostic
	for _, d := range ctx.Diagnostics {
		switch d.Code {
		case diagnostics.CodeMissingTerminator:
			actions = append(actions, p.terminatorFix(uri, text, d))
		case diagnostics.CodeEmptyOperandParameter:
			if a, ok := p.reworkFix(uri, text, d); ok {
				actions = append(actions, a)
			}
		case diagnostics.CodeMissingRequiredOperand:
			if name := operandName(d); name != "" {
				actions = append(actions, p.operandFix(uri, text, d, name))
				missingOperand = append(missingOperand, d)
			}
		case diagnostics.CodeMoveInsertOperands:
			actions = append(actions, p.moveOperandFixes(uri, text, d)...)
		}
	}
	if len(missingOperand) >= 2 {
		actions = append(actions, p.insertAllOperandsFix(uri, text, missingOperand))
	}
	if a, ok := p.reworkUpdateFix(uri, doc, reqRange.Start); ok {
		actions = append(actions, a)
	}
	return actions
}

// operandInsertPos returns the position to insert a missing operand skeleton:
// the end of the statement header. The diagnostic Range only covers the
// statement name token (e.g. "++SRC"), so anchoring there would insert into
// the name operand (e.g. "++SRC(S1)"). The diagnostics provider attaches the
// statement's end line in Data; we append at that line's end. Falls back to the
// diagnostic range end when the payload is absent.
func operandInsertPos(text string, d lsp.Diagnostic) lsp.Position {
	if m, ok := d.Data.(map[string]interface{}); ok {
		if line, ok := jsonInt(m["endLine"]); ok {
			return lsp.Position{Line: line, Character: lineLength(text, line)}
		}
	}
	return d.Range.End
}

// operandFix inserts a single missing operand skeleton after the statement header.
func (p *Provider) operandFix(uri, text string, d lsp.Diagnostic, name string) lsp.CodeAction {
	pos := operandInsertPos(text, d)
	edit := lsp.TextEdit{
		Range:   lsp.Range{Start: pos, End: pos},
		NewText: "\n    " + name + "()",
	}
	return lsp.CodeAction{
		Title:       "Insert operand " + name,
		Kind:        kindQuickFix,
		Diagnostics: []lsp.Diagnostic{d},
		Edit:        &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
	}
}

// insertAllOperandsFix inserts skeletons for all missing required operands at once.
// All missing-operand diagnostics for one statement share that statement's range,
// so anchoring at the first diagnostic's header end inserts every skeleton at the statement.
func (p *Provider) insertAllOperandsFix(uri, text string, diags []lsp.Diagnostic) lsp.CodeAction {
	var sb strings.Builder
	for _, d := range diags {
		sb.WriteString("\n    " + operandName(d) + "()")
	}
	pos := operandInsertPos(text, diags[0])
	edit := lsp.TextEdit{
		Range:   lsp.Range{Start: pos, End: pos},
		NewText: sb.String(),
	}
	return lsp.CodeAction{
		Title:       "Insert all required operands",
		Kind:        kindQuickFix,
		Diagnostics: diags,
		Edit:        &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
	}
}

// moveOperandPayload mirrors the ++MOVE fix payload attached to the diagnostic.
// Numbers, slices and maps round-trip through JSON; re-marshaling d.Data and
// decoding into these structs accepts both native test payloads and the
// JSON-decoded runtime shape.
type moveOperandPayload struct {
	Fixes []moveFix `json:"fixes"`
}

type moveFix struct {
	Title    string         `json:"title"`
	Operands []moveOperandS `json:"operands"`
}

type moveOperandS struct {
	Name   string `json:"name"`
	Parens bool   `json:"parens"`
}

// moveOperandFixes builds one code action per payload "fixes" entry, each
// inserting its operands at the statement header end. Operands with parens
// insert "NAME()"; boolean flag operands insert "NAME". Fixes with an empty
// title, no operands or empty resulting text are skipped. Never returns nil.
func (p *Provider) moveOperandFixes(uri, text string, d lsp.Diagnostic) []lsp.CodeAction {
	actions := []lsp.CodeAction{}
	raw, err := json.Marshal(d.Data)
	if err != nil {
		return actions
	}
	var payload moveOperandPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return actions
	}
	pos := operandInsertPos(text, d)
	for _, fix := range payload.Fixes {
		if fix.Title == "" || len(fix.Operands) == 0 {
			continue
		}
		var sb strings.Builder
		for _, op := range fix.Operands {
			sb.WriteString("\n    " + op.Name)
			if op.Parens {
				sb.WriteString("()")
			}
		}
		if sb.Len() == 0 {
			continue
		}
		edit := lsp.TextEdit{
			Range:   lsp.Range{Start: pos, End: pos},
			NewText: sb.String(),
		}
		actions = append(actions, lsp.CodeAction{
			Title:       fix.Title,
			Kind:        kindQuickFix,
			Diagnostics: []lsp.Diagnostic{d},
			Edit:        &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
		})
	}
	return actions
}

// lineLength returns the rune length of the given 0-based line in text,
// used as the character offset to append at the line's end. Returns 0 if
// the line is out of range.
func lineLength(text string, line int) int {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return 0
	}
	return len([]rune(lines[line]))
}

// jsonInt extracts an int from a Data payload value. Numbers round-trip
// through JSON as float64, so we accept both float64 and int.
func jsonInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	}
	return 0, false
}

// operandName reads the operand name from the diagnostic Data payload.
// Data round-trips through JSON as map[string]interface{}.
func operandName(d lsp.Diagnostic) string {
	m, ok := d.Data.(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := m["operand"].(string)
	return name
}

// reworkFix inserts the current Julian date inside the existing REWORK() parens.
// The diagnostic Range covers only the operand name; we locate the '(' and ')'
// in the document text at d.Range.End.Line to build a precise TextEdit.
// Returns false when the empty operand is not REWORK or parens cannot be found.
func (p *Provider) reworkFix(uri, text string, d lsp.Diagnostic) (lsp.CodeAction, bool) {
	if !strings.EqualFold(operandName(d), "REWORK") {
		return lsp.CodeAction{}, false
	}

	// Locate the '(' and ')' on the operand's line in the document.
	lines := strings.Split(text, "\n")
	lineIdx := d.Range.End.Line
	if lineIdx >= len(lines) {
		return lsp.CodeAction{}, false
	}
	ln := lines[lineIdx]
	endChar := d.Range.End.Character

	openIdx := strings.Index(ln[endChar:], "(")
	if openIdx < 0 {
		return lsp.CodeAction{}, false
	}
	openCol := endChar + openIdx
	closeIdx := strings.Index(ln[openCol:], ")")
	if closeIdx < 0 {
		return lsp.CodeAction{}, false
	}
	closeCol := openCol + closeIdx

	// Build an edit that replaces the content between '(' and ')' with the Julian date.
	julian := currentJulianDate()
	editRange := lsp.Range{
		Start: lsp.Position{Line: lineIdx, Character: openCol + 1},
		End:   lsp.Position{Line: lineIdx, Character: closeCol},
	}
	edit := lsp.TextEdit{Range: editRange, NewText: julian}
	return lsp.CodeAction{
		Title:       "Set REWORK to current date",
		Kind:        kindQuickFix,
		Diagnostics: []lsp.Diagnostic{d},
		Edit:        &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
		IsPreferred: true,
	}, true
}

// currentJulianDate returns today's date in SMP/E REWORK format (yyyyddd).
func currentJulianDate() string {
	now := time.Now()
	return fmt.Sprintf("%d%03d", now.Year(), now.YearDay())
}

// findNodeAtPosition finds the deepest AST node at the given position.
// This is a local copy of the same lookup used by hover/completion/signature;
// see the astutil refactor tech-debt note for consolidating these.
func findNodeAtPosition(doc *parser.Document, line, character int) *parser.Node {
	for _, stmt := range doc.Statements {
		if node := findNodeInTree(stmt, line, character); node != nil {
			return node
		}
	}
	return nil
}

func findNodeInTree(node *parser.Node, line, character int) *parser.Node {
	if node == nil {
		return nil
	}
	if line == node.Position.Line {
		nodeEnd := node.Position.Character + node.Position.Length
		if character >= node.Position.Character && character < nodeEnd {
			for _, child := range node.Children {
				if childNode := findNodeInTree(child, line, character); childNode != nil {
					return childNode
				}
			}
			return node
		}
	}
	for _, child := range node.Children {
		if childNode := findNodeInTree(child, line, character); childNode != nil {
			return childNode
		}
	}
	return nil
}

// reworkUpdateFix offers to refresh a non-empty REWORK value that is not
// today's date. It complements reworkFix, which only handles the empty-value
// case (CodeEmptyOperandParameter diagnostic); this one is cursor-based and
// does not depend on any diagnostic being present. Returns false when the
// cursor is not inside a statement's REWORK operand, when REWORK has no
// parameter node (empty — reworkFix's job), or when the value already
// matches today's date.
func (p *Provider) reworkUpdateFix(uri string, doc *parser.Document, pos lsp.Position) (lsp.CodeAction, bool) {
	if doc == nil {
		return lsp.CodeAction{}, false
	}

	node := findNodeAtPosition(doc, pos.Line, pos.Character)
	if node == nil {
		return lsp.CodeAction{}, false
	}

	// The cursor node is either the REWORK operand itself or one of its
	// parameter children; walk up until we find it or hit the enclosing
	// statement (in which case the cursor is on some other part of it,
	// e.g. the statement name or a different operand).
	reworkOp := node
	for reworkOp != nil && !(reworkOp.Type == parser.NodeTypeOperand && strings.EqualFold(reworkOp.Name, "REWORK")) {
		if reworkOp.Type == parser.NodeTypeStatement {
			reworkOp = nil
			break
		}
		reworkOp = reworkOp.Parent
	}
	if reworkOp == nil {
		return lsp.CodeAction{}, false
	}

	var paramNode *parser.Node
	for _, child := range reworkOp.Children {
		if child.Type == parser.NodeTypeParameter {
			paramNode = child
			break
		}
	}
	if paramNode == nil || strings.TrimSpace(paramNode.Value) == "" {
		// Empty REWORK is reworkFix's job, not ours.
		return lsp.CodeAction{}, false
	}

	julian := currentJulianDate()
	if strings.TrimSpace(paramNode.Value) == julian {
		return lsp.CodeAction{}, false
	}

	editRange := lsp.Range{
		Start: lsp.Position{Line: paramNode.Position.Line, Character: paramNode.Position.Character},
		End:   lsp.Position{Line: paramNode.Position.Line, Character: paramNode.Position.Character + paramNode.Position.Length},
	}
	edit := lsp.TextEdit{Range: editRange, NewText: julian}
	return lsp.CodeAction{
		Title: "Update REWORK to current date",
		Kind:  kindQuickFix,
		Edit:  &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
	}, true
}

// terminatorFix inserts a '.' on its own line after the end of the statement.
// The statement may span multiple lines, so the end line is read from the
// diagnostic Data payload (set by the diagnostics provider) and the '.' is
// inserted at that line's end. If the line is absent, it falls back to the
// diagnostic range end (statement header).
func (p *Provider) terminatorFix(uri, text string, d lsp.Diagnostic) lsp.CodeAction {
	pos := d.Range.End
	if m, ok := d.Data.(map[string]interface{}); ok {
		if line, ok := jsonInt(m["endLine"]); ok {
			pos = lsp.Position{Line: line, Character: lineLength(text, line)}
		}
	}
	edit := lsp.TextEdit{
		Range:   lsp.Range{Start: pos, End: pos},
		NewText: "\n.",
	}
	return lsp.CodeAction{
		Title:       "Add statement terminator",
		Kind:        kindQuickFix,
		Diagnostics: []lsp.Diagnostic{d},
		Edit:        &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{uri: {edit}}},
		IsPreferred: true,
	}
}
