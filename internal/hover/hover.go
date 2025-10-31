package hover

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// MCSStatement represents a MCS statement from smpe.json
type MCSStatement struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameter   string     `json:"parameter"`
	Type        string     `json:"type"`
	Operands    []Operand  `json:"operands"`
}

// Operand represents an operand of a MCS statement
type Operand struct {
	Name        string   `json:"name"`
	Parameter   string   `json:"parameter"`
	Type        string   `json:"type"`
	Length      string   `json:"length,omitempty"`
	Description string   `json:"description"`
	AllowedValues []string `json:"allowedValues,omitempty"`
}

// Provider provides hover information
type Provider struct {
	statements map[string]MCSStatement
}

// NewProvider creates a new hover provider
func NewProvider(dataPath string) (*Provider, error) {
	p := &Provider{
		statements: make(map[string]MCSStatement),
	}

	if err := p.loadStatements(dataPath); err != nil {
		return nil, err
	}

	return p, nil
}

// loadStatements loads MCS statements from smpe.json
func (p *Provider) loadStatements(dataPath string) error {
	// If dataPath is relative, make it absolute from executable directory
	if !filepath.IsAbs(dataPath) {
		ex, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
		exDir := filepath.Dir(ex)
		dataPath = filepath.Join(exDir, dataPath)
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		logger.Error("Failed to read smpe.json: %v", err)
		return fmt.Errorf("failed to read smpe.json: %w", err)
	}

	var statements []MCSStatement
	if err := json.Unmarshal(data, &statements); err != nil {
		logger.Error("Failed to parse smpe.json: %v", err)
		return fmt.Errorf("failed to parse smpe.json: %w", err)
	}

	for _, stmt := range statements {
		p.statements[stmt.Name] = stmt
	}

	logger.Info("Loaded %d MCS statements", len(p.statements))
	return nil
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
	if stmt, ok := p.statements[word]; ok {
		return p.createStatementHover(stmt)
	}

	// Check if it's an operand
	// Find which MCS statement we're in
	for i := line; i >= 0; i-- {
		statementName := p.findMCSStatement(lines[i])
		if statementName != "" {
			if stmt, ok := p.statements[statementName]; ok {
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
	for name := range p.statements {
		if strings.Contains(line, name) {
			return name
		}
	}
	return ""
}

// createStatementHover creates hover info for a statement
func (p *Provider) createStatementHover(stmt MCSStatement) *lsp.Hover {
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
func (p *Provider) createOperandHover(operand Operand) *lsp.Hover {
	content := fmt.Sprintf("**%s**\n\n%s\n\n", operand.Name, operand.Description)

	if operand.Parameter != "" {
		content += fmt.Sprintf("**Parameter:** %s\n\n", operand.Parameter)
	}

	if operand.Type != "" {
		content += fmt.Sprintf("**Type:** %s\n\n", operand.Type)
	}

	if operand.Length != "" {
		content += fmt.Sprintf("**Length:** %s\n\n", operand.Length)
	}

	if len(operand.AllowedValues) > 0 {
		content += "**Allowed Values:**\n"
		for _, val := range operand.AllowedValues {
			content += fmt.Sprintf("- `%s`\n", val)
		}
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupKindMarkdown,
			Value: content,
		},
	}
}
