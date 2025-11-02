package parser

import (
	"strings"
)

// StatementBoundary represents the start and end positions of a statement
type StatementBoundary struct {
	Start int // Character offset from start of file
	End   int // Character offset from start of file (-1 if not terminated)
}

// FindCurrentStatement finds the statement boundary containing the given position
// Statements are delimited by '.' (period) which can appear on any line
func FindCurrentStatement(text string, position int) StatementBoundary {
	if position < 0 || position > len(text) {
		return StatementBoundary{Start: 0, End: -1}
	}

	// Find previous statement terminator (scanning backwards from position)
	start := 0
	for i := position - 1; i >= 0; i-- {
		if text[i] == '.' {
			// Check if this is really a statement terminator
			// (not inside a comment or string)
			if isStatementTerminator(text, i) {
				start = i + 1
				break
			}
		}
	}

	// Skip leading whitespace after statement terminator
	for start < len(text) && (text[start] == ' ' || text[start] == '\t' || text[start] == '\n' || text[start] == '\r') {
		start++
	}

	// Find next statement terminator (scanning forwards from position)
	end := -1
	for i := position; i < len(text); i++ {
		if text[i] == '.' {
			// Check if this is really a statement terminator
			if isStatementTerminator(text, i) {
				end = i
				break
			}
		}
	}

	return StatementBoundary{Start: start, End: end}
}

// isStatementTerminator checks if the '.' at the given position is a real statement terminator
// Returns false if the '.' is inside a block comment /* ... */
func isStatementTerminator(text string, dotPos int) bool {
	// Check if we're inside a block comment
	inComment := false

	for i := 0; i <= dotPos; i++ {
		if i < len(text)-1 {
			// Check for comment start
			if text[i] == '/' && text[i+1] == '*' {
				inComment = true
				i++ // skip next char
				continue
			}
			// Check for comment end
			if text[i] == '*' && text[i+1] == '/' {
				if inComment {
					inComment = false
				}
				i++ // skip next char
				continue
			}
		}
	}

	// If we're still in a comment at dotPos, this is not a terminator
	return !inComment
}

// ExtractStatement extracts the statement text from the boundary
// Returns empty string if the statement is not terminated
func ExtractStatement(text string, boundary StatementBoundary) string {
	if boundary.End == -1 {
		// Statement not terminated, return up to end of text
		if boundary.Start >= len(text) {
			return ""
		}
		return text[boundary.Start:]
	}

	if boundary.Start >= boundary.End {
		return ""
	}

	return text[boundary.Start:boundary.End]
}

// RemoveComments removes all block comments /* ... */ from the text
// This simplifies parsing by removing comment noise
func RemoveComments(text string) string {
	var result strings.Builder
	inComment := false

	for i := 0; i < len(text); i++ {
		if i < len(text)-1 {
			// Check for comment start
			if text[i] == '/' && text[i+1] == '*' {
				inComment = true
				i++ // skip next char
				continue
			}
			// Check for comment end
			if text[i] == '*' && text[i+1] == '/' {
				if inComment {
					inComment = false
					i++ // skip next char
					continue
				}
			}
		}

		// Add character if not in comment
		if !inComment {
			result.WriteByte(text[i])
		}
	}

	return result.String()
}

// GetStatementType extracts the statement type (e.g., "++APAR") from the statement text
func GetStatementType(statementText string) string {
	// Remove comments first
	cleanText := RemoveComments(statementText)
	cleanText = strings.TrimSpace(cleanText)

	// Statement type is everything up to the first '(' or whitespace
	for i, ch := range cleanText {
		if ch == '(' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			return cleanText[:i]
		}
	}

	return cleanText
}

// GetCursorContext returns information about where the cursor is within the statement
type CursorContext struct {
	StatementType string // e.g., "++APAR"
	StatementText string // Full statement text (without comments)
	CursorOffset  int    // Cursor position within the statement
	InParameter   bool   // True if cursor is inside a parameter (within parentheses)
	InOperand     bool   // True if cursor is in an operand name
	OperandName   string // Name of the operand the cursor is in (if any)
}

// GetCursorContext analyzes the cursor position within a statement
func GetCursorContext(text string, cursorPos int) CursorContext {
	boundary := FindCurrentStatement(text, cursorPos)
	statementText := ExtractStatement(text, boundary)

	// Calculate cursor offset within the statement
	cursorOffset := cursorPos - boundary.Start
	if cursorOffset < 0 {
		cursorOffset = 0
	}

	// Remove comments for analysis
	cleanText := RemoveComments(statementText)

	ctx := CursorContext{
		StatementType: GetStatementType(cleanText),
		StatementText: cleanText,
		CursorOffset:  cursorOffset,
	}

	// Determine if cursor is inside parentheses
	// Count opening and closing parens up to cursor position
	openParens := 0
	for i := 0; i < cursorOffset && i < len(cleanText); i++ {
		if cleanText[i] == '(' {
			openParens++
		} else if cleanText[i] == ')' {
			openParens--
		}
	}
	ctx.InParameter = openParens > 0

	// Try to extract operand name if cursor is in or after one
	if cursorOffset > 0 && cursorOffset <= len(cleanText) {
		ctx.OperandName = extractOperandAtCursor(cleanText, cursorOffset)
		ctx.InOperand = ctx.OperandName != ""
	}

	return ctx
}

// extractOperandAtCursor finds the operand name at or before the cursor position
func extractOperandAtCursor(text string, cursorPos int) string {
	if cursorPos <= 0 || cursorPos > len(text) {
		return ""
	}

	// Scan backwards to find the start of the current word
	start := cursorPos - 1
	for start > 0 && isIdentifierChar(text[start]) {
		start--
	}
	if !isIdentifierChar(text[start]) {
		start++
	}

	// Scan forwards to find the end of the word
	end := cursorPos
	for end < len(text) && isIdentifierChar(text[end]) {
		end++
	}

	// Check if this looks like an operand (contains letters)
	word := text[start:end]
	if len(word) > 0 && !strings.HasPrefix(word, "++") {
		// Make sure it's not the statement parameter
		// (statement parameters are right after the statement type)
		return word
	}

	return ""
}

// isIdentifierChar returns true if the character is valid in an identifier
func isIdentifierChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') ||
		(ch >= 'a' && ch <= 'z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '+' || ch == '-'
}
