package semantic

import (
	"github.com/cybersorcerer/smpe_ls/internal/data"
)

// TokenType represents the type of semantic token
type TokenType int

const (
	TokenTypeKeyword TokenType = iota // MCS statements like ++USERMOD
	TokenTypeFunction                 // Operands like DESC, REWORK
	TokenTypeParameter                // Parameter values inside parentheses
	TokenTypeComment                  // Comments
	TokenTypeString                   // Quoted strings
	TokenTypeNumber                   // Numbers
)

// TokenModifier represents modifiers for semantic tokens
type TokenModifier int

const (
	TokenModifierNone TokenModifier = 0
)

// Provider provides semantic tokens for syntax highlighting
type Provider struct {
	statements map[string]data.MCSStatement
	operands   map[string]map[string]data.Operand // statement -> operand name -> operand
}

// NewProvider creates a new semantic token provider
func NewProvider(statements map[string]data.MCSStatement) *Provider {
	// Build operand lookup map for faster access
	operands := make(map[string]map[string]data.Operand)
	for stmtName, stmt := range statements {
		operands[stmtName] = make(map[string]data.Operand)
		for _, op := range stmt.Operands {
			operands[stmtName][op.Name] = op
		}
	}

	return &Provider{
		statements: statements,
		operands:   operands,
	}
}

// Token represents a single semantic token
type Token struct {
	Line      int
	StartChar int
	Length    int
	Type      TokenType
	Modifiers TokenModifier
}

// encodeTokens converts tokens to LSP delta-encoded format (flat integer array)
// Each token is represented by 5 integers: [deltaLine, deltaStart, length, tokenType, tokenModifiers]
func (p *Provider) encodeTokens(tokens []Token) []int {
	if len(tokens) == 0 {
		return []int{}
	}

	data := make([]int, 0, len(tokens)*5)
	prevLine := 0
	prevChar := 0

	for _, token := range tokens {
		deltaLine := token.Line - prevLine
		deltaChar := token.StartChar
		if deltaLine == 0 {
			deltaChar = token.StartChar - prevChar
		}

		data = append(data,
			deltaLine,
			deltaChar,
			token.Length,
			int(token.Type),
			int(token.Modifiers),
		)

		prevLine = token.Line
		prevChar = token.StartChar
	}

	return data
}

// isOperandChar checks if a character is valid in an operand name
func isOperandChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}
