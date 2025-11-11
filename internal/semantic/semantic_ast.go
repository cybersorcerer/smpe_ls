package semantic

import (
	"sort"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

// BuildTokensFromAST creates semantic tokens from an AST document
// This replaces the old string-based tokenization logic
func (p *Provider) BuildTokensFromAST(doc *parser.Document, text string) []int {
	tokens := []Token{}

	// Process comment nodes - expand them into per-line tokens
	commentTokens := p.expandComments(doc.Comments, text)
	tokens = append(tokens, commentTokens...)

	// Traverse AST and build tokens for statements
	for _, stmt := range doc.Statements {
		tokens = append(tokens, p.traverseNode(stmt)...)
	}

	// Sort tokens by line, then by character (required by LSP semantic tokens spec)
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Line != tokens[j].Line {
			return tokens[i].Line < tokens[j].Line
		}
		return tokens[i].StartChar < tokens[j].StartChar
	})

	// Convert to LSP delta-encoded format
	return p.encodeTokens(tokens)
}

// expandComments creates tokens from parser comment nodes
// For multi-line comments, it expands them into per-line tokens
// This uses parser nodes which already skip inline data sections
func (p *Provider) expandComments(comments []*parser.Node, text string) []Token {
	var tokens []Token

	if len(comments) == 0 {
		return tokens
	}

	// Split text into lines only once (needed for multi-line expansion)
	lines := strings.Split(text, "\n")

	for _, comment := range comments {
		startLine := comment.Position.Line
		startChar := comment.Position.Character

		// Check if this is a single-line comment by checking if comment end is on same line
		if startLine < len(lines) {
			line := lines[startLine]

			// Look for */ on the same line
			commentEndPos := strings.Index(line[startChar:], "*/")

			if commentEndPos != -1 {
				// Single-line comment (inline)
				tokens = append(tokens, Token{
					Line:      startLine,
					StartChar: startChar,
					Length:    commentEndPos + 2, // +2 for */
					Type:      TokenTypeComment,
					Modifiers: TokenModifierNone,
				})
			} else {
				// Multi-line comment - find end line and expand
				endLine := startLine + 1
				for endLine < len(lines) {
					if strings.Contains(lines[endLine], "*/") {
						break
					}
					endLine++
				}

				// Create token for first line
				firstLineLen := len(lines[startLine]) - startChar
				tokens = append(tokens, Token{
					Line:      startLine,
					StartChar: startChar,
					Length:    firstLineLen,
					Type:      TokenTypeComment,
					Modifiers: TokenModifierNone,
				})

				// Create tokens for middle lines
				for lineNum := startLine + 1; lineNum < endLine && lineNum < len(lines); lineNum++ {
					tokens = append(tokens, Token{
						Line:      lineNum,
						StartChar: 0,
						Length:    len(lines[lineNum]),
						Type:      TokenTypeComment,
						Modifiers: TokenModifierNone,
					})
				}

				// Create token for last line (if different from first)
				if endLine < len(lines) && endLine > startLine {
					endPos := strings.Index(lines[endLine], "*/")
					if endPos != -1 {
						tokens = append(tokens, Token{
							Line:      endLine,
							StartChar: 0,
							Length:    endPos + 2,
							Type:      TokenTypeComment,
							Modifiers: TokenModifierNone,
						})
					}
				}
			}
		}
	}

	logger.Debug("expandComments: Created %d comment tokens from %d parser nodes", len(tokens), len(comments))
	return tokens
}

// traverseNode recursively traverses an AST node and creates tokens
func (p *Provider) traverseNode(node *parser.Node) []Token {
	if node == nil {
		return []Token{}
	}

	tokens := []Token{}

	// Create token for this node based on its type
	switch node.Type {
	case parser.NodeTypeStatement:
		// Statement node -> Keyword token (blue)
		tokens = append(tokens, Token{
			Line:      node.Position.Line,
			StartChar: node.Position.Character,
			Length:    node.Position.Length,
			Type:      TokenTypeKeyword,
			Modifiers: TokenModifierNone,
		})
		logger.Debug("AST Token: Statement %s at line %d, char %d", node.Name, node.Position.Line, node.Position.Character)

	case parser.NodeTypeOperand:
		// Operand node -> Function token (purple/magenta)
		tokens = append(tokens, Token{
			Line:      node.Position.Line,
			StartChar: node.Position.Character,
			Length:    node.Position.Length,
			Type:      TokenTypeFunction,
			Modifiers: TokenModifierNone,
		})
		logger.Debug("AST Token: Operand %s at line %d, char %d", node.Name, node.Position.Line, node.Position.Character)

	case parser.NodeTypeParameter:
		// Parameter node -> Parameter token (orange)
		tokens = append(tokens, Token{
			Line:      node.Position.Line,
			StartChar: node.Position.Character,
			Length:    node.Position.Length,
			Type:      TokenTypeParameter,
			Modifiers: TokenModifierNone,
		})
		logger.Debug("AST Token: Parameter '%s' at line %d, char %d", node.Value, node.Position.Line, node.Position.Character)
	}

	// Recursively process children
	for _, child := range node.Children {
		tokens = append(tokens, p.traverseNode(child)...)
	}

	return tokens
}
