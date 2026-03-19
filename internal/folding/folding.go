package folding

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides folding range functionality
type Provider struct{}

// NewProvider creates a new folding provider
func NewProvider() *Provider {
	return &Provider{}
}

// GetFoldingRanges returns all folding ranges in a document
func (p *Provider) GetFoldingRanges(doc *parser.Document, lines []string) []lsp.FoldingRange {
	if doc == nil {
		return nil
	}

	var ranges []lsp.FoldingRange

	// Each MCS statement that spans multiple lines is a foldable region
	for _, stmt := range doc.Statements {
		if stmt == nil || stmt.Type != parser.NodeTypeStatement {
			continue
		}

		endLine := p.getStatementEndLine(stmt, lines)

		// Only create folding range if statement spans more than one line
		if endLine > stmt.Position.Line {
			ranges = append(ranges, lsp.FoldingRange{
				StartLine: stmt.Position.Line,
				EndLine:   endLine,
				Kind:      "region",
			})
		}
	}

	// Multi-line comments are foldable too
	for _, comment := range doc.Comments {
		if comment == nil {
			continue
		}

		endLine := comment.Position.Line
		// Check if this is a multi-line comment by scanning lines
		if comment.Position.Line < len(lines) {
			// Find the end of this comment block
			for i := comment.Position.Line; i < len(lines); i++ {
				trimmed := strings.TrimSpace(lines[i])
				if i == comment.Position.Line {
					// First line — check if it ends the comment on the same line
					if strings.Contains(trimmed, "/*") && strings.Contains(trimmed, "*/") {
						endLine = i
						break
					}
					continue
				}
				if strings.Contains(lines[i], "*/") {
					endLine = i
					break
				}
				// Stop if we hit a non-comment line
				if !strings.HasPrefix(trimmed, "*") && !strings.Contains(trimmed, "*/") {
					endLine = i - 1
					break
				}
			}
		}

		if endLine > comment.Position.Line {
			ranges = append(ranges, lsp.FoldingRange{
				StartLine: comment.Position.Line,
				EndLine:   endLine,
				Kind:      "comment",
			})
		}
	}

	return ranges
}

// getStatementEndLine finds the last line of a statement (including terminator)
func (p *Provider) getStatementEndLine(stmt *parser.Node, lines []string) int {
	endLine := stmt.Position.Line

	// Check children for the furthest line
	for _, child := range stmt.Children {
		if child.Position.Line > endLine {
			endLine = child.Position.Line
		}
		for _, grandchild := range child.Children {
			if grandchild.Position.Line > endLine {
				endLine = grandchild.Position.Line
			}
		}
	}

	// Look for terminator (.) after the last known position
	for i := endLine; i < len(lines); i++ {
		line := lines[i]
		for j := 0; j < len(line); j++ {
			if line[j] == '.' {
				return i
			}
		}
		// Stop if we hit another statement
		if i > stmt.Position.Line {
			for j := 0; j < len(line)-1; j++ {
				if line[j] == '+' && line[j+1] == '+' {
					return endLine
				}
			}
		}
	}

	return endLine
}
