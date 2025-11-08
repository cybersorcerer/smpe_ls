package hover

import (
	"fmt"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/langid"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides hover information
type Provider struct {
	statements map[string]data.MCSStatement
}

// NewProvider creates a new hover provider with shared data
func NewProvider(store *data.Store) *Provider {
	return &Provider{
		statements: store.Statements,
	}
}

// GetHover returns hover information for the given position
func (p *Provider) GetHover(text string, line, character int) *lsp.Hover {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return nil
	}

	// Find the word at the cursor position
	word := p.getWordAtPosition(currentLine, character)
	if word == "" {
		return nil
	}

	logger.Debug("Hover word: %s", word)

	// Check if it's a MCS statement
	// First try exact match
	stmt, ok := p.statements[word]
	if !ok {
		// Check if this is a language variant statement
		baseName, langID, hasLangID := langid.ExtractLanguageID(word)
		if hasLangID {
			// Try to get the base statement
			stmt, ok = p.statements[baseName]
			if ok {
				// Add language ID info to hover
				enhancedStmt := stmt
				enhancedStmt.Description = fmt.Sprintf("%s\n\n**Language:** %s", stmt.Description, langID)
				return p.createStatementHover(enhancedStmt)
			}
		}
	} else {
		return p.createStatementHover(stmt)
	}

	// Check if it's an operand
	// Find which MCS statement we're in
	for i := line; i >= 0; i-- {
		statementName := p.findMCSStatement(lines[i])
		if statementName != "" {
			// Try exact match first
			stmt, ok := p.statements[statementName]
			if !ok {
				// Check if this is a language variant
				baseName, _, hasLangID := langid.ExtractLanguageID(statementName)
				if hasLangID {
					stmt, ok = p.statements[baseName]
				}
			}

			if ok {
				for _, operand := range stmt.Operands {
					if operand.Name == word {
						return p.createOperandHover(operand)
					}
				}
			}
			break
		}
	}

	return nil
}

// getWordAtPosition gets the word at the given position
func (p *Provider) getWordAtPosition(line string, character int) string {
	// Find word boundaries
	start := character
	end := character

	// Move start backwards
	for start > 0 && (isWordChar(line[start-1])) {
		start--
	}

	// Move end forwards
	for end < len(line) && (isWordChar(line[end])) {
		end++
	}

	return line[start:end]
}

// isWordChar checks if a character is part of a word
func isWordChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
	       (ch >= '0' && ch <= '9') || ch == '_' || ch == '+' || ch == '-'
}

// findMCSStatement finds a MCS statement in a line
func (p *Provider) findMCSStatement(line string) string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "++") {
		return ""
	}

	// Try to extract the statement name
	// Format: ++STATEMENT(param) operands...
	// or:     ++STATEMENT operands...
	// or:     ++STATEMENT.

	// Find where the statement name ends
	endIdx := len(trimmed)
	for i := 2; i < len(trimmed); i++ {
		ch := trimmed[i]
		if ch == '(' || ch == ' ' || ch == '.' || ch == '\t' {
			endIdx = i
			break
		}
	}

	statementName := trimmed[:endIdx]

	// Check if this is a known statement
	if _, ok := p.statements[statementName]; ok {
		return statementName
	}

	// Check if it's a language variant
	baseName, _, hasLangID := langid.ExtractLanguageID(statementName)
	if hasLangID {
		if _, ok := p.statements[baseName]; ok {
			return baseName
		}
	}

	return ""
}

// createStatementHover creates hover info for a statement
func (p *Provider) createStatementHover(stmt data.MCSStatement) *lsp.Hover {
	content := fmt.Sprintf("**%s**\n\n%s\n\n", stmt.Name, stmt.Description)

	if stmt.Parameter != "" {
		content += fmt.Sprintf("**Parameter:** %s\n\n", stmt.Parameter)
	}

	if len(stmt.Operands) > 0 {
		content += "**Operands:**\n"
		for _, op := range stmt.Operands {
			content += fmt.Sprintf("- `%s`: %s\n", op.Name, op.Description)
		}
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupKindMarkdown,
			Value: content,
		},
	}
}

// createOperandHover creates hover info for an operand
func (p *Provider) createOperandHover(operand data.Operand) *lsp.Hover {
	content := fmt.Sprintf("**%s**\n\n%s\n\n", operand.Name, operand.Description)

	if operand.Parameter != "" {
		content += fmt.Sprintf("**Parameter:** %s\n\n", operand.Parameter)
	}

	if operand.Type != "" {
		content += fmt.Sprintf("**Type:** %s\n\n", operand.Type)
	}

	if operand.Length > 0 {
		content += fmt.Sprintf("**Length:** %d\n\n", operand.Length)
	}

	if operand.MutuallyExclusive != "" {
		content += fmt.Sprintf("**Mutually Exclusive with:** %s\n\n", operand.MutuallyExclusive)
	}

	if len(operand.Values) > 0 {
		content += "**Allowed Values:**\n"
		for _, val := range operand.Values {
			if val.Description != "" {
				content += fmt.Sprintf("- `%s`: %s\n", val.Name, val.Description)
			} else {
				content += fmt.Sprintf("- `%s`\n", val.Name)
			}
		}
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupKindMarkdown,
			Value: content,
		},
	}
}
