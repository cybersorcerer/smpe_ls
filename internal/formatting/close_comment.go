package formatting

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// closeCommentInsert is what is added after "/* ": a space, the closing marker,
// and nothing else. Typing then reads "/* text */" rather than "/* text*/".
const closeCommentInsert = " */"

// CloseComment answers the client's onTypeFormatting request for a typed space.
//
// After "/* " it inserts " */" on the same line, so a comment is closed the
// moment it is opened. The edit sits at the cursor, which stays where it is and
// writes into the comment that now exists around it.
//
// It stays out of the way in three cases:
//
//   - the space does not follow "/*", so nothing is being opened
//   - the comment is already closed further along the line, so a second "*/"
//     would end it in the wrong place
//   - the closing marker would reach past column 72, which SMP/E ignores - a
//     marker out there closes nothing and only looks like it does
//
// Inline data is left alone entirely. A "/*" there belongs to the element - a
// REXX program opens with one - and is none of the server's business.
func (p *Provider) CloseComment(doc *parser.Document, text string, pos lsp.Position, ch string) []lsp.TextEdit {
	if ch != " " {
		return nil
	}

	lines := strings.Split(text, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return nil
	}
	line := []rune(lines[pos.Line])
	col := pos.Character
	if col < 0 || col > len(line) {
		return nil
	}

	// The space just typed has to follow "/*" directly.
	if col < 3 || string(line[col-3:col]) != "/* " {
		return nil
	}

	// Already closed on this line: leave it alone.
	if strings.Contains(string(line[col:]), "*/") {
		return nil
	}

	// Columns are 1-based for the 72 limit, and the marker has to fit whole.
	if col+len([]rune(closeCommentInsert)) > MaxColumn {
		return nil
	}

	if isInsideInlineData(doc, pos.Line) {
		return nil
	}

	return []lsp.TextEdit{{
		Range:   lsp.Range{Start: pos, End: pos},
		NewText: closeCommentInsert,
	}}
}

// isInsideInlineData reports whether a line carries element data rather than
// MCS text. The parser has already worked out which statements expect inline
// data and where their data begins.
func isInsideInlineData(doc *parser.Document, line int) bool {
	if doc == nil {
		return false
	}
	for _, stmt := range doc.StatementsExpectingInline {
		start := stmt.Position.Line
		for _, child := range stmt.Children {
			if child.Position.Line > start {
				start = child.Position.Line
			}
		}
		if line <= start {
			continue
		}
		// Data runs until the next statement begins.
		next := -1
		for _, other := range doc.Statements {
			if other.Position.Line > start && (next == -1 || other.Position.Line < next) {
				next = other.Position.Line
			}
		}
		if next == -1 || line < next {
			return true
		}
	}
	return false
}
