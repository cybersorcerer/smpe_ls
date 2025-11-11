package completion

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// GetCompletionsAST returns completion items using the AST
func (p *Provider) GetCompletionsAST(doc *parser.Document, text string, line, character int) []lsp.CompletionItem {
	// Convert line/character to absolute position
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return nil
	}

	logger.Debug("GetCompletionsAST - line: %d, character: %d", line, character)

	// Check if we're inside inline data - if so, don't provide SMP/E completions
	if p.isInsideInlineDataAST(doc, line) {
		logger.Debug("Cursor is inside inline data - no completions")
		return nil
	}

	// Calculate absolute cursor position
	cursorPos := 0
	for i := 0; i < line; i++ {
		cursorPos += len(lines[i]) + 1 // +1 for newline
	}
	cursorPos += character

	// Check if we're at line start or typing ++
	textBefore := currentLine[:character]
	trimmedBefore := strings.TrimSpace(textBefore)

	// Check if this is a continuation line (line > 0, starts with whitespace, has statement before)
	isContinuationLine := false
	if line > 0 && len(currentLine) > 0 && (currentLine[0] == ' ' || currentLine[0] == '\t') {
		// Check if there's a statement on a previous line
		for _, stmt := range doc.Statements {
			if stmt.Position.Line < line {
				isContinuationLine = true
				break
			}
		}
	}

	// If we're typing + or ++, offer MCS completions (but NOT on continuation lines)
	if !isContinuationLine && (trimmedBefore == "" || (strings.HasPrefix(trimmedBefore, "+") && len(trimmedBefore) <= 2)) {
		// Calculate how many + characters were typed
		plusCount := 0
		for i := len(textBefore) - 1; i >= 0 && textBefore[i] == '+'; i-- {
			plusCount++
		}

		// Calculate the range to replace (the + characters)
		startChar := character - plusCount
		replaceRange := lsp.Range{
			Start: lsp.Position{Line: line, Character: startChar},
			End:   lsp.Position{Line: line, Character: character},
		}

		return p.getMCSCompletions(&replaceRange)
	}

	// Find the node at cursor position
	node, context := p.findNodeAtPosition(doc, text, line, character)

	if node == nil {
		// No node found, offer MCS completions
		logger.Debug("No node found at position, offering MCS completions")
		return p.getMCSCompletions(nil)
	}

	logger.Debug("Found node type %v at position", node.Type)

	// Handle different node contexts
	switch context {
	case ContextStatementParameter:
		// Cursor inside statement parameter - no completion
		logger.Debug("Cursor in statement parameter - no completion")
		return nil

	case ContextOperandParameter:
		// Cursor inside operand parameter - offer value completions if available
		logger.Debug("Cursor in operand parameter: %s", node.Name)
		return p.getOperandValueCompletionsAST(node, text, line, character)

	case ContextOperandName:
		// Cursor in or after operand name - offer operand completions
		logger.Debug("Cursor in operand name context")
		if node.Type == parser.NodeTypeStatement {
			// After statement, offer operand completions
			stmt := node
			return p.getOperandCompletionsAST(stmt, text, line, character)
		}
		// Within an operand name, offer operand completions
		return p.getOperandCompletionsAST(node.Parent, text, line, character)

	case ContextBetweenOperands:
		// Cursor between operands - offer operand completions
		logger.Debug("Cursor between operands")
		if node.Type == parser.NodeTypeStatement {
			return p.getOperandCompletionsAST(node, text, line, character)
		}
		return nil

	default:
		// Unknown context
		logger.Debug("Unknown context")
		return nil
	}
}

// CompletionContext represents where the cursor is within the AST
type CompletionContext int

const (
	ContextUnknown CompletionContext = iota
	ContextStatementParameter
	ContextOperandParameter
	ContextOperandName
	ContextBetweenOperands
)

// findNodeAtPosition finds the AST node at the given position and determines the context
func (p *Provider) findNodeAtPosition(doc *parser.Document, text string, line, character int) (*parser.Node, CompletionContext) {
	// Find the statement that contains this position
	var targetStmt *parser.Node
	for _, stmt := range doc.Statements {
		// Check if position is within this statement's range
		if line == stmt.Position.Line {
			targetStmt = stmt
			break
		}
		// For multiline statements, check if line is between statement start and last child
		if line > stmt.Position.Line {
			// Find the last line of this statement
			lastLine := stmt.Position.Line
			for _, child := range stmt.Children {
				lastLine = p.getLastLine(child, lastLine)
			}
			if line <= lastLine {
				targetStmt = stmt
				break
			}
		}
	}

	// If no statement found and we're on a continuation line (line > 0, starts with whitespace),
	// check if there's a statement on a previous line that we should continue
	if targetStmt == nil && line > 0 {
		lines := strings.Split(text, "\n")
		if line < len(lines) {
			currentLine := lines[line]
			// Check if line starts with whitespace (continuation line)
			if len(currentLine) > 0 && (currentLine[0] == ' ' || currentLine[0] == '\t') {
				// Find the last statement before this line
				for _, stmt := range doc.Statements {
					if stmt.Position.Line < line {
						targetStmt = stmt
					}
				}
			}
		}
	}

	if targetStmt == nil {
		return nil, ContextUnknown
	}

	// Check if cursor is in statement parameter
	// Statement parameter is right after statement name: ++STMT(param)
	if len(targetStmt.Children) > 0 {
		firstChild := targetStmt.Children[0]
		if firstChild.Type == parser.NodeTypeParameter && firstChild.Parent == targetStmt {
			// Check if cursor is on same line and within parameter range
			if line == firstChild.Position.Line {
				paramStart := firstChild.Position.Character - 1 // -1 for opening (
				paramEnd := firstChild.Position.Character + firstChild.Position.Length + 1 // +1 for closing )
				if character >= paramStart && character <= paramEnd {
					return targetStmt, ContextStatementParameter
				}
			}
		}
	}

	// Check if cursor is within an operand's parameter
	operandNode := p.findOperandAtPosition(targetStmt, line, character, text)
	if operandNode != nil {
		// Check if cursor is inside operand's parameter (either with existing parameter node or empty ())
		hasParameterNode := false
		for _, child := range operandNode.Children {
			if child.Type == parser.NodeTypeParameter {
				hasParameterNode = true
				// Check if cursor is within parameter range
				if line == child.Position.Line {
					paramStart := child.Position.Character - 1 // -1 for opening (
					paramEnd := child.Position.Character + child.Position.Length + 1 // +1 for closing )
					if character >= paramStart && character <= paramEnd {
						return operandNode, ContextOperandParameter
					}
				}
			}
		}

		// If no parameter node found, check if cursor is between parentheses after operand name
		// This handles cases like FROMDS() and FROMDS(DSN(...) ...) where sub-operands exist
		if !hasParameterNode && line == operandNode.Position.Line {
			nameEnd := operandNode.Position.Character + operandNode.Position.Length
			// Check if there's an opening paren right after the operand name
			lines := strings.Split(text, "\n")
			if line < len(lines) {
				currentLine := lines[line]
				// Look for ( after operand name
				for i := nameEnd; i < len(currentLine); i++ {
					if currentLine[i] == '(' {
						// Found opening paren, now find MATCHING closing paren with depth tracking
						closingParen := -1
						parenDepth := 1
						for j := i + 1; j < len(currentLine); j++ {
							if currentLine[j] == '(' {
								parenDepth++
							} else if currentLine[j] == ')' {
								parenDepth--
								if parenDepth == 0 {
									closingParen = j
									break
								}
							}
						}

						// Check if cursor is after (
						if character >= i {
							// If there's a closing paren, check if cursor is before or on it
							if closingParen != -1 && character <= closingParen {
								return operandNode, ContextOperandParameter
							}
							// If no closing paren and cursor is after (, we're still in parameter
							if closingParen == -1 {
								return operandNode, ContextOperandParameter
							}
						}
						break
					} else if currentLine[i] != ' ' && currentLine[i] != '\t' {
						// Non-whitespace that's not (, stop looking
						break
					}
				}
			}
		}

		// Check if this operand is actually a sub-operand (child of another operand)
		// If so, we're in the parent operand's parameter context
		if operandNode.Parent != nil && operandNode.Parent.Type == parser.NodeTypeOperand {
			// This is a sub-operand, treat cursor as being in parent's parameter
			return operandNode.Parent, ContextOperandParameter
		}

		// Cursor is on the operand name itself
		if line == operandNode.Position.Line {
			nameStart := operandNode.Position.Character
			nameEnd := nameStart + operandNode.Position.Length
			if character >= nameStart && character <= nameEnd {
				return operandNode, ContextOperandName
			}
		}
	}

	// Cursor is somewhere in the statement but not in a specific parameter
	// Offer operand completions
	return targetStmt, ContextBetweenOperands
}

// findSubOperandEnd finds the end position of a sub-operand including its parameters
func (p *Provider) findSubOperandEnd(subOp *parser.Node, currentLine string) int {
	opStart := subOp.Position.Character
	opEnd := opStart + subOp.Position.Length
	maxEnd := opEnd

	// Check for parameter children of this sub-operand
	for _, child := range subOp.Children {
		if child.Type == parser.NodeTypeParameter {
			paramEnd := child.Position.Character + child.Position.Length + 1 // +1 for )
			if paramEnd > maxEnd {
				maxEnd = paramEnd
			}
		}
	}

	// If no parameter child found, check if there are parentheses after sub-operand name
	if maxEnd == opEnd && opEnd < len(currentLine) {
		for i := opEnd; i < len(currentLine); i++ {
			if currentLine[i] == '(' {
				// Found opening paren, look for closing paren
				parenDepth := 1
				for j := i + 1; j < len(currentLine); j++ {
					if currentLine[j] == '(' {
						parenDepth++
					} else if currentLine[j] == ')' {
						parenDepth--
						if parenDepth == 0 {
							maxEnd = j
							break
						}
					}
				}
				break
			} else if currentLine[i] != ' ' && currentLine[i] != '\t' {
				// Non-whitespace that's not (, stop looking
				break
			}
		}
	}

	return maxEnd
}

// findOperandAtPosition finds the operand node at the given position
func (p *Provider) findOperandAtPosition(stmt *parser.Node, line, character int, text string) *parser.Node {
	lines := strings.Split(text, "\n")

	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeOperand {
			// Check if cursor is on this operand's line
			if line == child.Position.Line && line < len(lines) {
				currentLine := lines[line]
				opStart := child.Position.Character
				opEnd := opStart + child.Position.Length

				// Check if cursor is within operand name or after it (for parameter)
				if character >= opStart {
					// Find the end of this operand (including its parameter and sub-operands)
					// We need to find the matching closing paren after the operand name
					opEndWithParam := opEnd

					// Look for ( after operand name to find the parameter block
					hasParam := false
					for i := opEnd; i < len(currentLine); i++ {
						if currentLine[i] == '(' {
							hasParam = true
							// Found opening paren, find matching closing paren with depth tracking
							parenDepth := 1
							for j := i + 1; j < len(currentLine); j++ {
								if currentLine[j] == '(' {
									parenDepth++
								} else if currentLine[j] == ')' {
									parenDepth--
									if parenDepth == 0 {
										// This is the matching closing paren
										opEndWithParam = j
										break
									}
								}
							}
							// If no closing paren found (unclosed parenthesis), extend to end of line
							if opEndWithParam == opEnd {
								opEndWithParam = len(currentLine)
							}
							break
						} else if currentLine[i] != ' ' && currentLine[i] != '\t' {
							// Non-whitespace that's not (, stop looking
							break
						}
					}

					// If we have sub-operands but didn't find the parameter opening above,
					// try to infer it from the sub-operand positions
					if !hasParam && len(child.Children) > 0 {
						for _, paramChild := range child.Children {
							if paramChild.Type == parser.NodeTypeParameter {
								paramEnd := paramChild.Position.Character + paramChild.Position.Length + 1 // +1 for )
								if paramEnd > opEndWithParam {
									opEndWithParam = paramEnd
								}
							} else if paramChild.Type == parser.NodeTypeOperand {
								// This is a sub-operand (e.g., DSN inside FROMDS)
								// Find its end position including its own parameters
								subOpEnd := p.findSubOperandEnd(paramChild, currentLine)
								if subOpEnd > opEndWithParam {
									opEndWithParam = subOpEnd
								}
							}
						}
					}

					if character <= opEndWithParam {
						return child
					}
				}
			}

			// Check children recursively (for sub-operands)
			if subOp := p.findOperandAtPosition(child, line, character, text); subOp != nil {
				return subOp
			}
		}
	}
	return nil
}

// getLastLine recursively finds the last line number in a node tree
func (p *Provider) getLastLine(node *parser.Node, currentMax int) int {
	if node.Position.Line > currentMax {
		currentMax = node.Position.Line
	}
	for _, child := range node.Children {
		currentMax = p.getLastLine(child, currentMax)
	}
	return currentMax
}

// getOperandCompletionsAST returns operand completions for a statement using AST
func (p *Provider) getOperandCompletionsAST(stmt *parser.Node, text string, line, character int) []lsp.CompletionItem {
	if stmt.StatementDef == nil {
		return nil
	}

	// Collect already present operands
	presentOperands := make(map[string]bool)
	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeOperand {
			presentOperands[child.Name] = true
		}
	}

	logger.Debug("Present operands: %v", presentOperands)

	// Build completion items for operands
	var items []lsp.CompletionItem
	for _, op := range stmt.StatementDef.Operands {
		// Split aliases
		names := strings.Split(op.Name, "|")
		primaryName := names[0]

		// Check if already present
		alreadyPresent := false
		for _, name := range names {
			if presentOperands[name] {
				alreadyPresent = true
				break
			}
		}

		// Filter based on mutually_exclusive and allowed_if
		if alreadyPresent {
			continue
		}

		// Check mutually_exclusive
		if op.MutuallyExclusive != "" {
			exclusiveOps := strings.Split(op.MutuallyExclusive, "|")
			skip := false
			for _, exOp := range exclusiveOps {
				if presentOperands[exOp] {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		// Check allowed_if
		if op.AllowedIf != "" {
			requiredOps := strings.Split(op.AllowedIf, "|")
			allowed := false
			for _, reqOp := range requiredOps {
				if presentOperands[reqOp] {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		// Build completion item
		item := lsp.CompletionItem{
			Label:  primaryName,
			Kind:   lsp.CompletionItemKindProperty,
			Detail: op.Description,
		}

		// Add parameter hint if available
		if op.Parameter != "" {
			item.InsertText = primaryName + "($0)"
			item.InsertTextFormat = lsp.InsertTextFormatSnippet
		} else {
			item.InsertText = primaryName
		}

		items = append(items, item)
	}

	return items
}

// getOperandValueCompletionsAST returns value completions for operand parameters using AST
func (p *Provider) getOperandValueCompletionsAST(operandNode *parser.Node, fullText string, line int, character int) []lsp.CompletionItem {
	if operandNode.OperandDef == nil {
		return nil
	}

	// Check if this operand has sub-operands
	// Sub-operands are identified by the parent operand having a Parameter field with parentheses
	// e.g., "DSN(dsname) NUMBER(number)" indicates DSN and NUMBER are sub-operands
	if len(operandNode.OperandDef.Values) > 0 && operandNode.OperandDef.Parameter != "" && strings.Contains(operandNode.OperandDef.Parameter, "(") {
		// These are sub-operands (e.g., DSN, NUMBER for FROMDS)
		var items []lsp.CompletionItem
		for _, subOp := range operandNode.OperandDef.Values {
			item := lsp.CompletionItem{
				Label:  subOp.Name,
				Kind:   lsp.CompletionItemKindProperty,
				Detail: subOp.Description,
			}

			// Sub-operands always have parameters (indicated by parent's Parameter field)
			item.InsertText = subOp.Name + "($0)"
			item.InsertTextFormat = lsp.InsertTextFormatSnippet

			items = append(items, item)
		}
		return items
	}

	// Otherwise, these are simple enumerated value completions
	if len(operandNode.OperandDef.Values) > 0 {
		var items []lsp.CompletionItem
		for _, value := range operandNode.OperandDef.Values {
			item := lsp.CompletionItem{
				Label:  value.Name,
				Kind:   lsp.CompletionItemKindValue,
				Detail: value.Description,
			}
			items = append(items, item)
		}
		return items
	}

	return nil
}

// isInsideInlineDataAST checks if the cursor line is inside inline data using AST
func (p *Provider) isInsideInlineDataAST(doc *parser.Document, line int) bool {
	// Find if there's a statement expecting inline data that contains this line
	for _, stmt := range doc.StatementsExpectingInline {
		// Find the line where this statement ends (its terminator)
		stmtEndLine := stmt.Position.Line
		for _, child := range stmt.Children {
			childEnd := p.getLastLine(child, stmtEndLine)
			if childEnd > stmtEndLine {
				stmtEndLine = childEnd
			}
		}

		// Find the start of next statement (if any)
		nextStmtLine := -1
		for j, nextStmt := range doc.Statements {
			if nextStmt.Position.Line > stmtEndLine {
				// Make sure this is actually after our statement in document order
				stmtIdx := -1
				for k, s := range doc.Statements {
					if s == stmt {
						stmtIdx = k
						break
					}
				}
				if j > stmtIdx {
					nextStmtLine = nextStmt.Position.Line
					break
				}
			}
		}

		// If cursor is between statement end and next statement start, it's in inline data
		if line > stmtEndLine {
			if nextStmtLine == -1 || line < nextStmtLine {
				logger.Debug("Cursor at line %d is in inline data after statement at line %d", line, stmt.Position.Line)
				return true
			}
		}
	}

	return false
}
