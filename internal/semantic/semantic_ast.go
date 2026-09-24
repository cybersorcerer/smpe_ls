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

	// Traverse AST and build tokens for statements. The lines are needed to
	// split a value written across lines into one token per line.
	lines := strings.Split(text, "\n")
	for _, stmt := range doc.Statements {
		tokens = append(tokens, p.traverseNode(stmt, lines)...)
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
func (p *Provider) traverseNode(node *parser.Node, lines []string) []Token {
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
		tokens = append(tokens, parameterTokens(node, lines)...)
		logger.Debug("AST Token: Parameter '%s' at line %d, char %d", node.Value, node.Position.Line, node.Position.Character)
	}

	// Recursively process children
	for _, child := range node.Children {
		tokens = append(tokens, p.traverseNode(child, lines)...)
	}

	return tokens
}

// parameterTokens turns a parameter node into the tokens that describe it.
//
// A semantic token cannot span a line break: the LSP wire format encodes each
// token as a line plus a start column and a length, so a token whose length
// reaches past the end of its line is malformed and VS Code drops it. That is
// what happened to a value written across lines,
//
//	SUP(
//	    LBCP034
//	)
//
// where the node covers everything between the parentheses, newlines included.
// The value lost its colour while a value on one line kept it.
//
// A parameter that holds a list is already broken down into one child node per
// element, and those children carry their own tokens. Splitting the parent as
// well would put two tokens on the same text, so in that case the parent is
// left out and the children speak for it.
func parameterTokens(node *parser.Node, lines []string) []Token {
	line, start, length := node.Position.Line, node.Position.Character, node.Position.Length
	if line >= len(lines) {
		return nil
	}

	// Fits on its own line: one token, as before.
	if start+length <= len([]rune(lines[line])) {
		return []Token{{
			Line:      line,
			StartChar: start,
			Length:    length,
			Type:      TokenTypeParameter,
			Modifiers: TokenModifierNone,
		}}
	}

	// The children already cover the individual values.
	if len(node.Children) > 0 {
		return nil
	}

	var tokens []Token
	remaining := length
	for ln := line; ln < len(lines) && remaining > 0; ln++ {
		runes := []rune(lines[ln])
		from := 0
		if ln == line {
			from = start
		}
		if from > len(runes) {
			from = len(runes)
		}
		to := len(runes)
		if from+remaining < to {
			to = from + remaining
		}
		// What is left after this line, counting the newline that joins it
		// to the next one.
		remaining -= (to - from) + 1

		// Emit the value only, not the indentation around it.
		for from < to && isSpace(runes[from]) {
			from++
		}
		for to > from && isSpace(runes[to-1]) {
			to--
		}
		if to > from {
			tokens = append(tokens, Token{
				Line:      ln,
				StartChar: from,
				Length:    to - from,
				Type:      TokenTypeParameter,
				Modifiers: TokenModifierNone,
			})
		}
	}
	return tokens
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}
