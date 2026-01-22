package formatting

import (
	"strings"
	"unicode/utf8"

	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Config holds formatting configuration options
type Config struct {
	Enabled             bool
	IndentContinuation  int  // Number of spaces for continuation lines (default: 3)
	OneOperandPerLine   bool // Put each operand on its own line
	AlignOperands       bool // Align operands vertically
	PreserveComments    bool // Keep comments in their original position
}

// DefaultConfig returns the default formatting configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:            true,
		IndentContinuation: 3,
		OneOperandPerLine:  true,
		AlignOperands:      false,
		PreserveComments:   true,
	}
}

// Provider provides document formatting functionality
type Provider struct {
	config *Config
}

// NewProvider creates a new formatting provider
func NewProvider() *Provider {
	return &Provider{
		config: DefaultConfig(),
	}
}

// SetConfig updates the formatting configuration
func (p *Provider) SetConfig(config *Config) {
	if config != nil {
		p.config = config
	}
}

// GetConfig returns the current formatting configuration
func (p *Provider) GetConfig() *Config {
	return p.config
}

// FormatDocument formats the entire document
func (p *Provider) FormatDocument(doc *parser.Document, text string) []lsp.TextEdit {
	if !p.config.Enabled || doc == nil {
		return nil
	}

	var edits []lsp.TextEdit
	lines := strings.Split(text, "\n")

	for _, stmt := range doc.Statements {
		stmtEdits := p.formatStatement(stmt, lines)
		edits = append(edits, stmtEdits...)
	}

	return edits
}

// FormatRange formats a specific range in the document
func (p *Provider) FormatRange(doc *parser.Document, text string, startLine, endLine int) []lsp.TextEdit {
	if !p.config.Enabled || doc == nil {
		return nil
	}

	var edits []lsp.TextEdit
	lines := strings.Split(text, "\n")

	for _, stmt := range doc.Statements {
		// Check if statement is within the range
		if stmt.Position.Line >= startLine && stmt.Position.Line <= endLine {
			stmtEdits := p.formatStatement(stmt, lines)
			edits = append(edits, stmtEdits...)
		}
	}

	return edits
}

// formatStatement formats a single statement
func (p *Provider) formatStatement(stmt *parser.Node, lines []string) []lsp.TextEdit {
	if stmt == nil || stmt.Type != parser.NodeTypeStatement {
		return nil
	}

	// Get the original statement text (may span multiple lines)
	originalText := p.getStatementText(stmt, lines)
	if originalText == "" {
		return nil
	}

	// Build formatted text
	formattedText := p.buildFormattedStatement(stmt)
	if formattedText == "" || formattedText == originalText {
		return nil
	}

	// Calculate the range to replace
	startLine := stmt.Position.Line
	endLine := p.getStatementEndLine(stmt, lines)

	// Create a single edit that replaces the entire statement
	edit := lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{Line: startLine, Character: 0},
			End:   lsp.Position{Line: endLine, Character: len(lines[endLine])},
		},
		NewText: formattedText,
	}

	return []lsp.TextEdit{edit}
}

// getStatementText returns the full text of a statement including continuation lines
func (p *Provider) getStatementText(stmt *parser.Node, lines []string) string {
	if stmt.Position.Line >= len(lines) {
		return ""
	}

	startLine := stmt.Position.Line
	endLine := p.getStatementEndLine(stmt, lines)

	var parts []string
	for i := startLine; i <= endLine && i < len(lines); i++ {
		parts = append(parts, lines[i])
	}

	return strings.Join(parts, "\n")
}

// getStatementEndLine finds the last line of a statement
func (p *Provider) getStatementEndLine(stmt *parser.Node, lines []string) int {
	endLine := stmt.Position.Line

	// Check children for the furthest line
	for _, child := range stmt.Children {
		if child.Position.Line > endLine {
			endLine = child.Position.Line
		}
		// Also check grandchildren (e.g., operand parameters)
		for _, grandchild := range child.Children {
			if grandchild.Position.Line > endLine {
				endLine = grandchild.Position.Line
			}
		}
	}

	// Look for terminator on subsequent lines
	for i := endLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasSuffix(line, ".") {
			endLine = i
			break
		}
		// Stop if we hit another statement
		if strings.HasPrefix(line, "++") && i > stmt.Position.Line {
			break
		}
	}

	return endLine
}

// buildFormattedStatement builds the formatted text for a statement
func (p *Provider) buildFormattedStatement(stmt *parser.Node) string {
	var sb strings.Builder

	// Write statement name
	sb.WriteString(stmt.Name)

	// Write statement parameter if present
	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
			sb.WriteString("(")
			sb.WriteString(child.Value)
			sb.WriteString(")")
			break
		}
	}

	// Collect operands
	var operands []*parser.Node
	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeOperand {
			operands = append(operands, child)
		}
	}

	// Format operands
	indent := strings.Repeat(" ", p.config.IndentContinuation)

	if p.config.OneOperandPerLine && len(operands) > 0 {
		// Each operand on its own line
		for _, op := range operands {
			sb.WriteString("\n")
			sb.WriteString(indent)
			sb.WriteString(p.formatOperand(op))
		}
	} else {
		// All operands on the same line (with wrapping if needed)
		for i, op := range operands {
			if i == 0 {
				sb.WriteString(" ")
			} else {
				sb.WriteString(" ")
			}
			sb.WriteString(p.formatOperand(op))
		}
	}

	// Add terminator on its own line at the beginning
	if stmt.HasTerminator {
		sb.WriteString("\n.")
	}

	return sb.String()
}

// formatOperand formats a single operand
func (p *Provider) formatOperand(op *parser.Node) string {
	if op == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(op.Name)

	// Check for parameter
	for _, child := range op.Children {
		if child.Type == parser.NodeTypeParameter {
			sb.WriteString("(")
			sb.WriteString(p.formatOperandParameter(child))
			sb.WriteString(")")
			break
		}
	}

	return sb.String()
}

// formatOperandParameter formats the parameter value of an operand
func (p *Provider) formatOperandParameter(param *parser.Node) string {
	if param == nil {
		return ""
	}

	// Check if this parameter has sub-operands (nested structure)
	if len(param.Children) > 0 {
		var parts []string
		for _, child := range param.Children {
			if child.Type == parser.NodeTypeOperand {
				parts = append(parts, p.formatOperand(child))
			} else if child.Type == parser.NodeTypeParameter {
				parts = append(parts, child.Value)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, " ")
		}
	}

	return param.Value
}

// runeCount returns the number of runes in a string
func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}
