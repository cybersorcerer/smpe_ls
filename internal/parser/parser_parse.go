package parser

import (
	"strings"
)

// Parse parses the given text and returns a Document with AST
func (p *Parser) Parse(text string) *Document {
	doc := &Document{
		Statements:                []*Node{},
		Comments:                  []*Node{},
		Errors:                    []ParseError{},
		StatementsExpectingInline: []*Node{},
	}

	lines := strings.Split(text, "\n")

	// First pass: Extract all comments and create clean lines for parsing
	cleanLines := make([]string, len(lines))
	inBlockComment := false
	var commentStartLine, commentStartChar int

	// Track if we're inside a statement region (from ++ to terminator)
	inStatement := false

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line starts a statement
		if strings.HasPrefix(trimmed, "++") {
			inStatement = true
		}

		// Check if this line ends a statement (has terminator)
		if inStatement && hasTerminatorOutsideParens(trimmed) {
			// Statement ends on this line, but keep inStatement=true for this line
			// It will be reset after processing
		}

		hasCommentStart := strings.Contains(line, "/*")
		hasCommentEnd := strings.Contains(line, "*/")

		// Only process comments if we're in a statement region
		// Comments outside statements (e.g., in inline data) are ignored
		if inStatement {
			// Inline comment (/* ... */ on same line)
			if hasCommentStart && hasCommentEnd {
				commentStart := strings.Index(line, "/*")
				commentEnd := strings.Index(line, "*/")

				commentNode := &Node{
					Type:  NodeTypeComment,
					Value: line[commentStart : commentEnd+2],
					Position: Position{
						Line:      lineNum,
						Character: commentStart,
						Length:    commentEnd + 2 - commentStart,
					},
				}
				doc.Comments = append(doc.Comments, commentNode)

				// Create clean line (before + after comment)
				before := line[:commentStart]
				after := ""
				if commentEnd+2 < len(line) {
					after = line[commentEnd+2:]
				}
				cleanLines[lineNum] = before + " " + after

				// If statement ends on this line, reset inStatement
				if hasTerminatorOutsideParens(trimmed) {
					inStatement = false
				}
				continue
			}

			// Block comment start
			if hasCommentStart && !inBlockComment {
				inBlockComment = true
				commentStartLine = lineNum
				commentStartChar = strings.Index(line, "/*")
				// Preserve content before the comment start
				before := ""
				if commentStartChar > 0 {
					before = line[:commentStartChar]
				}
				cleanLines[lineNum] = before
				continue
			}

			// Block comment end
			if hasCommentEnd && inBlockComment {
				inBlockComment = false
				commentEnd := strings.Index(line, "*/")

				commentNode := &Node{
					Type: NodeTypeComment,
					Position: Position{
						Line:      commentStartLine,
						Character: commentStartChar,
						Length:    commentEnd + 2,
					},
				}
				doc.Comments = append(doc.Comments, commentNode)

				// Keep any content after the comment end (e.g., terminator)
				after := ""
				if commentEnd+2 < len(line) {
					after = line[commentEnd+2:]
				}
				cleanLines[lineNum] = after

				// If statement ends on this line, reset inStatement
				if hasTerminatorOutsideParens(trimmed) {
					inStatement = false
				}
				continue
			}

			// Inside block comment
			if inBlockComment {
				cleanLines[lineNum] = ""
				continue
			}
		}

		// Regular line (or line outside statement with potential comment)
		cleanLines[lineNum] = line

		// If we were in a statement and it ended, reset the flag
		if inStatement && hasTerminatorOutsideParens(trimmed) {
			inStatement = false
		}
	}

	// Second pass: Collect complete statements using clean lines
	statements := p.collectStatements(cleanLines)

	// Third pass: Parse statements and track inline data
	for i, stmt := range statements {
		lineNum := stmt.StartLine

		// Parse using joined text (for correct multiline handling)
		var currentStatement *Node
		p.parseLine(stmt.Text, lineNum, doc, &currentStatement)

		// Fix positions in the AST nodes for this statement (only needed for multiline)
		if currentStatement != nil {
			if len(stmt.Lines) > 1 {
				p.fixPositionsForMultilineStatement(currentStatement, stmt)
			}

			// Set unbalanced parentheses flag
			currentStatement.UnbalancedParens = stmt.UnbalancedParens

			// Check if this statement expects inline data
			if currentStatement.StatementDef != nil && currentStatement.StatementDef.InlineData {
				trimmed := strings.TrimSpace(stmt.Text)
				if strings.HasSuffix(trimmed, ".") {
					hasDelete := false
					hasExternalSource := false

					for _, child := range currentStatement.Children {
						if child.Type == NodeTypeOperand {
							if child.Name == "DELETE" {
								hasDelete = true
								break
							}
							if child.Name == "TXLIB" || child.Name == "RELFILE" || child.Name == "FROMDS" {
								hasExternalSource = true
							}
						}
					}

					if !hasDelete && !hasExternalSource {
						doc.StatementsExpectingInline = append(doc.StatementsExpectingInline, currentStatement)

						// Track inline data: check lines between this statement and next
						endLine := len(cleanLines)
						if i+1 < len(statements) {
							endLine = statements[i+1].StartLine
						}

						// Count non-empty lines after this statement as inline data
						for lineIdx := stmt.StartLine + len(stmt.Lines); lineIdx < endLine; lineIdx++ {
							if strings.TrimSpace(cleanLines[lineIdx]) != "" {
								currentStatement.HasInlineData = true
								currentStatement.InlineDataLines++
							}
						}
					}
				}
			}
		}
	}

	return doc
}

// originalPos represents a position in the original multi-line text
type originalPos struct {
	Line int
	Char int
}

// fixPositionsForMultilineStatement corrects position information for nodes
// that were parsed from joined text back to original line/char positions
func (p *Parser) fixPositionsForMultilineStatement(stmt *Node, collected CollectedStatement) {
	// Build position mapping: char position in joined text -> (line, char) in original
	posMap := make(map[int]originalPos)
	joinedPos := 0

	for lineIdx, line := range collected.Lines {
		lineNum := collected.StartLine + lineIdx

		// Map each character in this line
		for charIdx := range line {
			posMap[joinedPos] = originalPos{Line: lineNum, Char: charIdx}
			joinedPos++
		}

		// Account for the space added by Join
		if lineIdx < len(collected.Lines)-1 {
			posMap[joinedPos] = originalPos{Line: lineNum, Char: len(line)}
			joinedPos++
		}
	}

	// Recursively fix positions in all nodes
	p.fixNodePosition(stmt, posMap)
}

// fixNodePosition recursively updates position information for a node and its children
func (p *Parser) fixNodePosition(node *Node, posMap map[int]originalPos) {
	if node == nil {
		return
	}

	// Look up the original position for this node's character offset
	// All positions in the parsed tree are relative to the joined text
	if origPos, ok := posMap[node.Position.Character]; ok {
		node.Position.Line = origPos.Line
		node.Position.Character = origPos.Char
	}

	// Recursively fix children
	for _, child := range node.Children {
		p.fixNodePosition(child, posMap)
	}
}
