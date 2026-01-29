package formatting

import (
	"strings"
	"unicode/utf8"

	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

const (
	// MaxColumn is the maximum column for SMP/E content (columns 1-72)
	// Columns 73-80 are ignored by SMP/E
	MaxColumn = 72
)

// Config holds formatting configuration options
type Config struct {
	Enabled             bool
	IndentContinuation  int  // Number of spaces for continuation lines (default: 3)
	OneOperandPerLine   bool // Put each operand on its own line
	AlignOperands       bool // Align operands vertically
	PreserveComments    bool // Keep comments in their original position
	MoveLeadingComments bool // Move comments from before first statement into the statement
}

// DefaultConfig returns the default formatting configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:             true,
		IndentContinuation:  3,
		OneOperandPerLine:   true,
		AlignOperands:       false,
		PreserveComments:    true,
		MoveLeadingComments: false, // Default: don't move, just show diagnostic
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

// CommentInfo stores information about a comment in the original text
type CommentInfo struct {
	Text             string // The comment text including /* */
	Line             int    // Original line number
	Character        int    // Original character position
	AtEnd            bool   // Comment was at the end of a line (after content)
	AfterDot         bool   // Comment was after the terminator
	BeforeTerminator bool   // Multi-line comment that appears before the terminator line
}

// FormatDocument formats the entire document
func (p *Provider) FormatDocument(doc *parser.Document, text string) []lsp.TextEdit {
	if !p.config.Enabled || doc == nil {
		return nil
	}

	var edits []lsp.TextEdit
	lines := strings.Split(text, "\n")

	// Build a map of comments to move into each statement
	// Key: statement index, Value: comments to insert into that statement
	commentsForStatement := make(map[int][]CommentInfo)
	// Track the edit start line for each statement (may be earlier than statement line if comments are moved)
	editStartLine := make(map[int]int)

	if p.config.MoveLeadingComments && len(doc.Statements) > 0 {
		// 1. Comments before first statement -> move into first statement
		firstStmtLine := doc.Statements[0].Position.Line
		if firstStmtLine > 0 {
			comments := p.extractCommentsFromRange(lines, 0, firstStmtLine)
			if len(comments) > 0 {
				commentsForStatement[0] = append(commentsForStatement[0], comments...)
				// Extend the edit range to include the comment lines
				editStartLine[0] = 0
			}
		}

		// 2. Comments between statements -> move into next statement
		for i := 0; i < len(doc.Statements)-1; i++ {
			currentStmt := doc.Statements[i]
			nextStmt := doc.Statements[i+1]

			// Skip if current statement expects inline data
			// (lines after it are inline data, not standalone comments)
			if p.stmtExpectsInlineData(currentStmt) {
				continue
			}

			// Find the end line of current statement
			currentStmtEndLine := p.getStatementEndLine(currentStmt, lines)
			nextStmtStartLine := nextStmt.Position.Line

			// Check for comments between statements (starting from line AFTER terminator)
			gapStartLine := currentStmtEndLine + 1
			if gapStartLine < nextStmtStartLine {
				comments := p.extractCommentsFromRange(lines, gapStartLine, nextStmtStartLine)
				if len(comments) > 0 {
					// Move these comments into the NEXT statement (i+1)
					commentsForStatement[i+1] = append(commentsForStatement[i+1], comments...)
					// Extend the edit range to include the comment lines
					editStartLine[i+1] = gapStartLine
				}
			}
		}

		// Note: Comments after last statement cannot be moved (no next statement)
		// They will still show a diagnostic error
	}

	// Format each statement, passing any comments that should be moved into it
	for i, stmt := range doc.Statements {
		// Note: Statements expecting inline data ARE formatted (the statement itself),
		// but they don't get leading comments moved into them (see below)

		// Get comments to insert into this statement
		extraComments := commentsForStatement[i]

		// Determine the edit start line (may be earlier if we're absorbing comment lines)
		startLine := stmt.Position.Line
		if extStart, ok := editStartLine[i]; ok {
			startLine = extStart
		}

		stmtEdits := p.formatStatementWithLeadingCommentsAndRange(stmt, lines, doc.Comments, extraComments, startLine)
		edits = append(edits, stmtEdits...)
	}

	return edits
}

// stmtExpectsInlineData checks if a statement expects inline data
// A statement expects inline data if:
// 1. inline_data is true in smpe.json AND
// 2. NO external data source operands (FROMDS, RELFILE, TXLIB) AND
// 3. NO DELETE operand (DELETE means deletion mode, no inline data needed)
func (p *Provider) stmtExpectsInlineData(stmt *parser.Node) bool {
	// First check if statement definition indicates inline data
	if stmt.StatementDef == nil || !stmt.StatementDef.InlineData {
		return false
	}

	// Check if statement has operands that indicate data is NOT inline
	// FROMDS, RELFILE, TXLIB mean data comes from elsewhere
	// DELETE means the element is being deleted (no inline data needed)
	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeOperand {
			opName := child.Name
			if opName == "FROMDS" || opName == "RELFILE" || opName == "TXLIB" || opName == "DELETE" {
				return false
			}
		}
	}

	return true
}

// extractCommentsFromRange extracts comments from a range of lines (startLine inclusive, endLine exclusive)
func (p *Provider) extractCommentsFromRange(lines []string, startLine, endLine int) []CommentInfo {
	var comments []CommentInfo

	for lineNum := startLine; lineNum < endLine && lineNum < len(lines); lineNum++ {
		line := lines[lineNum]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Check if this line contains a comment
		if strings.Contains(trimmed, "/*") {
			// Extract the comment(s) from this line
			commentStart := strings.Index(line, "/*")
			commentEnd := strings.Index(line, "*/")

			if commentEnd != -1 {
				// Single-line comment
				commentText := line[commentStart : commentEnd+2]
				comments = append(comments, CommentInfo{
					Text:      commentText,
					Line:      lineNum,
					Character: commentStart,
					AtEnd:     false,
					AfterDot:  false,
				})
			} else {
				// Multi-line comment - find the end
				var commentText strings.Builder
				commentText.WriteString(line[commentStart:])

				for lineNum++; lineNum < endLine && lineNum < len(lines); lineNum++ {
					commentText.WriteString("\n")
					commentText.WriteString(lines[lineNum])
					if strings.Contains(lines[lineNum], "*/") {
						break
					}
				}
				comments = append(comments, CommentInfo{
					Text:      commentText.String(),
					Line:      lineNum,
					Character: commentStart,
					AtEnd:     false,
					AfterDot:  false,
				})
			}
		}
	}

	return comments
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
			stmtEdits := p.formatStatement(stmt, lines, doc.Comments)
			edits = append(edits, stmtEdits...)
		}
	}

	return edits
}

// formatStatement formats a single statement
func (p *Provider) formatStatement(stmt *parser.Node, lines []string, comments []*parser.Node) []lsp.TextEdit {
	return p.formatStatementWithLeadingComments(stmt, lines, comments, nil)
}

// formatStatementWithLeadingComments formats a single statement with optional leading comments
func (p *Provider) formatStatementWithLeadingComments(stmt *parser.Node, lines []string, comments []*parser.Node, leadingComments []CommentInfo) []lsp.TextEdit {
	return p.formatStatementWithLeadingCommentsAndRange(stmt, lines, comments, leadingComments, stmt.Position.Line)
}

// formatStatementWithLeadingCommentsAndRange formats a single statement with optional leading comments
// and allows specifying a custom edit start line (to absorb comment lines being moved into the statement)
func (p *Provider) formatStatementWithLeadingCommentsAndRange(stmt *parser.Node, lines []string, comments []*parser.Node, leadingComments []CommentInfo, editStartLine int) []lsp.TextEdit {
	if stmt == nil || stmt.Type != parser.NodeTypeStatement {
		return nil
	}

	// Get the original statement text (may span multiple lines)
	stmtStartLine := stmt.Position.Line
	endLine := p.getStatementEndLine(stmt, lines)

	originalText := p.getStatementText(stmt, lines)
	if originalText == "" {
		return nil
	}

	// Extract comments from the statement region
	stmtComments := p.extractCommentsInRange(comments, stmtStartLine, endLine, lines)

	// Merge leading comments with statement comments
	allComments := append(leadingComments, stmtComments...)

	// Build formatted text with comments preserved
	formattedText := p.buildFormattedStatementWithLeadingComments(stmt, allComments, lines, leadingComments)
	if formattedText == "" {
		return nil
	}
	// Skip edit if nothing changed, but always create edit if we have leading comments to insert
	if formattedText == originalText && len(leadingComments) == 0 {
		return nil
	}

	// Create a single edit that replaces from editStartLine to end of statement
	// This allows absorbing comment lines that are being moved into the statement
	edit := lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{Line: editStartLine, Character: 0},
			End:   lsp.Position{Line: endLine, Character: len(lines[endLine])},
		},
		NewText: formattedText,
	}

	return []lsp.TextEdit{edit}
}

// extractCommentsInRange extracts all comments within a statement's line range
func (p *Provider) extractCommentsInRange(comments []*parser.Node, startLine, endLine int, lines []string) []CommentInfo {
	var result []CommentInfo

	for _, comment := range comments {
		if comment.Position.Line >= startLine && comment.Position.Line <= endLine {
			info := CommentInfo{
				Text:      comment.Value,
				Line:      comment.Position.Line,
				Character: comment.Position.Character,
			}

			// Check if comment is at end of line (after other content)
			if comment.Position.Line < len(lines) {
				line := lines[comment.Position.Line]
				beforeComment := strings.TrimSpace(line[:min(comment.Position.Character, len(line))])
				if beforeComment != "" && !strings.HasPrefix(beforeComment, "/*") {
					info.AtEnd = true
				}

				// Check if comment is after terminator (dot appears BEFORE the comment)
				// Note: If the dot appears AFTER the comment (e.g., "CALLLIBS /* comment */.")
				// then AfterDot should be false - the comment belongs to the statement
				if strings.Contains(beforeComment, ".") {
					info.AfterDot = true
				}
			}

			result = append(result, info)
		}
	}

	return result
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
// This includes any trailing comment after the terminator (even multi-line)
func (p *Provider) getStatementEndLine(stmt *parser.Node, lines []string) int {
	// Scan from statement start line to find the terminator
	// Track multi-line comment state to avoid treating dots inside comments as terminators
	terminatorLine := -1
	inMultiLineComment := false

	for i := stmt.Position.Line; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Stop if we hit another statement (but not on the first line)
		if strings.HasPrefix(trimmedLine, "++") && i > stmt.Position.Line {
			break
		}

		// Process this line character by character to track comment state
		// and find terminator dots that are outside comments
		foundTerminatorOnThisLine := false
		pos := 0
		for pos < len(line) {
			if inMultiLineComment {
				// Look for comment end
				endIdx := strings.Index(line[pos:], "*/")
				if endIdx == -1 {
					// Comment continues to next line
					break
				}
				pos += endIdx + 2
				inMultiLineComment = false
			} else {
				// Look for comment start, single-line comment, or dot
				remaining := line[pos:]

				// Find next interesting character
				commentStartIdx := strings.Index(remaining, "/*")
				dotIdx := strings.Index(remaining, ".")

				// If dot comes before comment start (or no comment), check if it's a terminator
				if dotIdx != -1 && (commentStartIdx == -1 || dotIdx < commentStartIdx) {
					// Found a dot outside of comment - this is a terminator
					foundTerminatorOnThisLine = true
					terminatorLine = i
					// Continue to check if a multi-line comment starts after this
					pos += dotIdx + 1
					continue
				}

				if commentStartIdx == -1 {
					// No more comments on this line
					break
				}

				// Found comment start
				pos += commentStartIdx + 2
				// Check if it closes on the same line
				endIdx := strings.Index(line[pos:], "*/")
				if endIdx == -1 {
					// Multi-line comment starts here
					inMultiLineComment = true
					break
				}
				pos += endIdx + 2
			}
		}

		// If we found a terminator and we're not in a multi-line comment, we're done
		if foundTerminatorOnThisLine && !inMultiLineComment {
			break
		}
		// If we found a terminator but are now in a multi-line comment,
		// the comment is after the terminator - continue to find where it ends
		if foundTerminatorOnThisLine && inMultiLineComment {
			// Comment after terminator - find where it ends
			for j := i + 1; j < len(lines); j++ {
				if strings.Contains(lines[j], "*/") {
					terminatorLine = j
					break
				}
				// Stop if we hit another statement (malformed comment)
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "++") {
					terminatorLine = j - 1
					break
				}
				terminatorLine = j
			}
			break
		}
	}

	if terminatorLine >= 0 {
		return terminatorLine
	}

	// No terminator found - return the furthest line from children
	endLine := stmt.Position.Line
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
	return endLine
}

// extractTrailingCommentAfterTerminator extracts the comment after the terminator from original text
// This handles multi-line comments that span multiple lines after the dot
func (p *Provider) extractTrailingCommentAfterTerminator(stmt *parser.Node, lines []string) string {
	if !stmt.HasTerminator {
		return ""
	}

	// Find the line with the terminator
	terminatorLine := -1
	for i := stmt.Position.Line; i < len(lines); i++ {
		line := lines[i]
		cleanLine := removeCommentsFromLine(line)
		cleanLine = strings.TrimSpace(cleanLine)
		if strings.HasSuffix(cleanLine, ".") {
			terminatorLine = i
			break
		}
		// Stop if we hit another statement
		if strings.HasPrefix(strings.TrimSpace(line), "++") && i > stmt.Position.Line {
			break
		}
	}

	if terminatorLine < 0 || terminatorLine >= len(lines) {
		return ""
	}

	// Find the position of the terminator in the line
	line := lines[terminatorLine]
	dotIdx := -1
	// Find the last dot that's not inside a comment or string
	inComment := false
	inString := false
	for i := 0; i < len(line); i++ {
		if !inString && i+1 < len(line) && line[i] == '/' && line[i+1] == '*' {
			inComment = true
			i++ // skip the *
			continue
		}
		if inComment && i+1 < len(line) && line[i] == '*' && line[i+1] == '/' {
			inComment = false
			i++ // skip the /
			continue
		}
		if !inComment && line[i] == '\'' {
			inString = !inString
			continue
		}
		if !inComment && !inString && line[i] == '.' {
			dotIdx = i
		}
	}

	if dotIdx < 0 {
		return ""
	}

	// Extract everything after the dot
	afterDot := ""
	if dotIdx+1 < len(line) {
		afterDot = strings.TrimSpace(line[dotIdx+1:])
	}

	// Check if there's a comment starting after the dot
	if afterDot == "" {
		return ""
	}

	// Check if the comment spans multiple lines
	if strings.Contains(afterDot, "/*") && !strings.Contains(afterDot, "*/") {
		// Multi-line comment - collect all lines until we find */
		var commentLines []string
		commentLines = append(commentLines, afterDot)
		for i := terminatorLine + 1; i < len(lines); i++ {
			commentLines = append(commentLines, lines[i])
			if strings.Contains(lines[i], "*/") {
				break
			}
			// Stop if we hit another statement
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "++") {
				// Remove the last line (which is a statement, not part of comment)
				commentLines = commentLines[:len(commentLines)-1]
				break
			}
		}
		return strings.Join(commentLines, "\n")
	}

	// Single-line comment or text after dot
	return afterDot
}

// wrapCommentAt72 wraps a comment to fit within column 72
// If the comment (with indent) exceeds column 72, it's converted to a multi-line comment
// Input: indent (e.g. "   ") and comment text (e.g. "/* This is a long comment */")
// Returns: slice of lines, each fitting within column 72
func (p *Provider) wrapCommentAt72(indent string, comment string) []string {
	fullLine := indent + comment
	if runeCount(fullLine) <= MaxColumn {
		return []string{fullLine}
	}

	// Extract the comment content (without /* and */)
	content := strings.TrimPrefix(comment, "/*")
	content = strings.TrimSuffix(content, "*/")
	content = strings.TrimSpace(content)

	// Calculate available space per line
	// Format: indent + "/* " + content + " */"
	// For first line (if also last): indent + "/* " + text + " */"
	// For first line (if not last): indent + "/* " + text
	// For middle lines: indent + "   " + text (continuation)
	// For last line: indent + "   " + text + " */"
	prefixFirst := indent + "/* "
	prefixCont := indent + "   "
	suffix := " */"

	// First line must also account for suffix when determining if we need to wrap
	availableFirst := MaxColumn - runeCount(prefixFirst) - runeCount(suffix)
	availableCont := MaxColumn - runeCount(prefixCont) - runeCount(suffix)

	if availableFirst <= 10 || availableCont <= 10 {
		// Not enough space, just return the original (will show diagnostic)
		return []string{fullLine}
	}

	// Split content into words
	words := strings.Fields(content)
	if len(words) == 0 {
		return []string{indent + "/* */"}
	}

	var resultLines []string
	var currentLine strings.Builder
	isFirstLine := true

	// Helper to split a long word into chunks that fit
	splitLongWord := func(word string, available int) []string {
		var chunks []string
		runes := []rune(word)
		for len(runes) > 0 {
			chunkSize := available
			if chunkSize > len(runes) {
				chunkSize = len(runes)
			}
			chunks = append(chunks, string(runes[:chunkSize]))
			runes = runes[chunkSize:]
		}
		return chunks
	}

	for i, word := range words {
		available := availableCont
		if isFirstLine {
			available = availableFirst
		}

		isLastWord := i == len(words)-1

		if currentLine.Len() == 0 {
			// Starting a new line
			if runeCount(word) > available {
				// Word itself is too long - split it into chunks
				chunks := splitLongWord(word, available)
				for j, chunk := range chunks {
					isLastChunk := j == len(chunks)-1 && isLastWord
					if isFirstLine {
						if isLastChunk {
							resultLines = append(resultLines, prefixFirst+chunk+suffix)
						} else {
							resultLines = append(resultLines, prefixFirst+chunk)
						}
						isFirstLine = false
					} else {
						if isLastChunk {
							resultLines = append(resultLines, prefixCont+chunk+suffix)
						} else {
							resultLines = append(resultLines, prefixCont+chunk)
						}
					}
				}
			} else {
				currentLine.WriteString(word)
				// If this is the last word, flush immediately
				if isLastWord {
					if isFirstLine {
						resultLines = append(resultLines, prefixFirst+currentLine.String()+suffix)
					} else {
						resultLines = append(resultLines, prefixCont+currentLine.String()+suffix)
					}
					currentLine.Reset()
				}
			}
		} else {
			// Adding to existing line
			newLen := currentLine.Len() + 1 + runeCount(word)
			if newLen <= available {
				currentLine.WriteString(" ")
				currentLine.WriteString(word)
				// If this is the last word, flush immediately
				if isLastWord {
					if isFirstLine {
						resultLines = append(resultLines, prefixFirst+currentLine.String()+suffix)
					} else {
						resultLines = append(resultLines, prefixCont+currentLine.String()+suffix)
					}
					currentLine.Reset()
				}
			} else {
				// Flush current line and start new one
				if isFirstLine {
					resultLines = append(resultLines, prefixFirst+currentLine.String())
					isFirstLine = false
				} else {
					resultLines = append(resultLines, prefixCont+currentLine.String())
				}
				currentLine.Reset()

				// Check if word fits on new line
				if runeCount(word) > availableCont {
					// Word is too long - split it
					chunks := splitLongWord(word, availableCont)
					for j, chunk := range chunks {
						isLastChunk := j == len(chunks)-1 && isLastWord
						if isLastChunk {
							resultLines = append(resultLines, prefixCont+chunk+suffix)
						} else {
							resultLines = append(resultLines, prefixCont+chunk)
						}
					}
				} else {
					currentLine.WriteString(word)
					// If this is the last word, flush immediately
					if isLastWord {
						resultLines = append(resultLines, prefixCont+currentLine.String()+suffix)
						currentLine.Reset()
					}
				}
			}
		}
	}

	return resultLines
}

// removeCommentsFromLine removes /* */ comments from a line
func removeCommentsFromLine(line string) string {
	result := line
	for {
		start := strings.Index(result, "/*")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "*/")
		if end == -1 {
			// Unterminated comment - remove from start to end
			result = result[:start]
			break
		}
		result = result[:start] + result[start+end+2:]
	}
	return result
}

// buildFormattedStatementWithLeadingComments builds the formatted text for a statement,
// with support for inserting leading comments (comments from before the first statement)
func (p *Provider) buildFormattedStatementWithLeadingComments(stmt *parser.Node, comments []CommentInfo, lines []string, leadingComments []CommentInfo) string {
	var outputLines []string
	indent := strings.Repeat(" ", p.config.IndentContinuation)

	// Start building the first line: statement name + parameter
	firstLine := stmt.Name

	// Add statement parameter if present
	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
			firstLine += "(" + child.Value + ")"
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

	// Separate leading comments from inline comments
	// Leading comments should be inserted after the statement name line
	var inlineComments []CommentInfo
	for _, c := range comments {
		// Skip leading comments - they are handled separately
		isLeading := false
		for _, lc := range leadingComments {
			if lc.Text == c.Text && lc.Line == c.Line {
				isLeading = true
				break
			}
		}
		if isLeading {
			continue
		}

		// Skip after-dot comments - they are handled via extractTrailingCommentAfterTerminator
		if !c.AfterDot {
			inlineComments = append(inlineComments, c)
		}
	}

	// Extract trailing comment after terminator directly from original text
	// This handles multi-line comments that the parser doesn't capture correctly
	trailingComment := p.extractTrailingCommentAfterTerminator(stmt, lines)

	if p.config.OneOperandPerLine && len(operands) > 0 {
		// First line is just the statement header
		outputLines = append(outputLines, p.wrapLineAt72(firstLine, ""))

		// Insert leading comments after the statement header line
		// Comments start at column 3 (2 space indent), not at operand indent
		commentIndent := "  "
		for _, lc := range leadingComments {
			// Wrap comment if it exceeds column 72
			wrappedLines := p.wrapCommentAt72(commentIndent, lc.Text)
			outputLines = append(outputLines, wrappedLines...)
		}

		// Each operand on its own line
		for _, op := range operands {
			opText := indent + p.formatOperand(op)
			outputLines = append(outputLines, p.wrapLineAt72(opText, indent))
		}

		// Insert multi-line inline comments BEFORE the terminator
		for _, c := range inlineComments {
			if strings.Contains(c.Text, "\n") {
				commentLines := strings.Split(c.Text, "\n")
				wrappedLines := p.wrapMultiLineCommentAt72(commentLines)
				outputLines = append(outputLines, wrappedLines...)
			}
		}

		// Terminator on its own line
		if stmt.HasTerminator {
			termLine := "."
			// Add trailing comment after terminator (may be multi-line)
			if trailingComment != "" {
				termLine += " " + trailingComment
			}
			outputLines = append(outputLines, termLine)
		}
	} else {
		// All on fewer lines - need to track column position
		currentLine := firstLine

		// If we have leading comments and no operands, insert them after header
		if len(leadingComments) > 0 {
			outputLines = append(outputLines, currentLine)
			// Comments start at column 3 (2 space indent)
			commentIndent := "  "
			for _, lc := range leadingComments {
				// Wrap comment if it exceeds column 72
				wrappedLines := p.wrapCommentAt72(commentIndent, lc.Text)
				outputLines = append(outputLines, wrappedLines...)
			}
			currentLine = ""
		}

		for i, op := range operands {
			opText := p.formatOperand(op)

			// Check if adding this operand would exceed column 72
			if currentLine == "" {
				currentLine = indent + opText
			} else {
				newLen := runeCount(currentLine) + 1 + runeCount(opText)
				if newLen > MaxColumn && currentLine != firstLine {
					// Start a new line
					outputLines = append(outputLines, currentLine)
					currentLine = indent + opText
				} else {
					if i == 0 || currentLine == firstLine {
						currentLine += " " + opText
					} else {
						currentLine += " " + opText
					}
				}
			}
		}

		// Flush current line before multi-line comments
		if currentLine != "" {
			outputLines = append(outputLines, currentLine)
			currentLine = ""
		}

		// Insert multi-line inline comments BEFORE the terminator
		for _, c := range inlineComments {
			if strings.Contains(c.Text, "\n") {
				commentLines := strings.Split(c.Text, "\n")
				wrappedLines := p.wrapMultiLineCommentAt72(commentLines)
				outputLines = append(outputLines, wrappedLines...)
			}
		}

		// Add terminator
		if stmt.HasTerminator {
			termLine := "."
			// Add trailing comment after terminator (may be multi-line)
			if trailingComment != "" {
				termLine += " " + trailingComment
			}
			outputLines = append(outputLines, termLine)
		}
	}

	// Insert single-line inline comments into the output
	// Multi-line comments are already handled above (before terminator)
	// Single-line comments are added at the end of lines if they fit
	if len(inlineComments) > 0 && len(outputLines) > 0 {
		for _, c := range inlineComments {
			// Skip multi-line comments - already handled
			if strings.Contains(c.Text, "\n") {
				continue
			}
			// Single-line comment: try to add at end of a line if it fits
			added := false
			for i := range outputLines {
				lineLen := runeCount(outputLines[i])
				commentLen := runeCount(c.Text)
				if lineLen+1+commentLen <= MaxColumn {
					outputLines[i] += " " + c.Text
					added = true
					break
				}
			}
			// If it didn't fit anywhere, add it as a separate line before terminator
			if !added {
				insertIdx := len(outputLines)
				if insertIdx > 0 && strings.TrimSpace(outputLines[insertIdx-1]) == "." {
					insertIdx = insertIdx - 1
				}
				newOutput := make([]string, 0, len(outputLines)+1)
				newOutput = append(newOutput, outputLines[:insertIdx]...)
				newOutput = append(newOutput, c.Text)
				newOutput = append(newOutput, outputLines[insertIdx:]...)
				outputLines = newOutput
			}
		}
	}

	return strings.Join(outputLines, "\n")
}

// wrapLineAt72 wraps a line if it exceeds column 72
// Returns the wrapped line(s) as a single string with newlines
func (p *Provider) wrapLineAt72(line string, continuationIndent string) string {
	if runeCount(line) <= MaxColumn {
		return line
	}

	var result []string
	remaining := line

	for runeCount(remaining) > MaxColumn {
		// Find a good break point before column 72
		breakPoint := p.findBreakPoint(remaining, MaxColumn)
		if breakPoint <= 0 {
			// No good break point found, force break at MaxColumn
			breakPoint = MaxColumn
		}

		// Extract the part that fits
		runes := []rune(remaining)
		result = append(result, string(runes[:breakPoint]))

		// Continue with the rest, adding continuation indent
		remaining = continuationIndent + strings.TrimLeft(string(runes[breakPoint:]), " ")
	}

	if remaining != "" {
		result = append(result, remaining)
	}

	return strings.Join(result, "\n")
}

// findBreakPoint finds a good position to break a line (at a space or after a comma)
func (p *Provider) findBreakPoint(line string, maxCol int) int {
	runes := []rune(line)
	if len(runes) <= maxCol {
		return len(runes)
	}

	// Look for the last space or comma before maxCol
	lastBreak := -1
	for i := maxCol - 1; i > 0; i-- {
		if runes[i] == ' ' || runes[i] == ',' {
			lastBreak = i + 1 // Break after the space/comma
			break
		}
	}

	return lastBreak
}

// wrapMultiLineCommentAt72 ensures multi-line comments start at column 3 (2 space indent)
// Comments must not start in column 1 in SMP/E
func (p *Provider) wrapMultiLineCommentAt72(commentLines []string) []string {
	const commentIndent = "  " // 2 spaces = column 3
	result := make([]string, len(commentLines))
	for i, line := range commentLines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			result[i] = ""
		} else {
			result[i] = commentIndent + trimmed
		}
	}
	return result
}

// formatOperand formats a single operand
func (p *Provider) formatOperand(op *parser.Node) string {
	if op == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(op.Name)

	// Check for parameter or sub-operands
	hasContent := false
	for _, child := range op.Children {
		if child.Type == parser.NodeTypeParameter {
			sb.WriteString("(")
			sb.WriteString(p.formatOperandParameter(child))
			sb.WriteString(")")
			hasContent = true
			break
		}
	}

	// If no parameter found, check for sub-operands (e.g., FROMDS(DSN(...)))
	if !hasContent {
		var subOps []string
		for _, child := range op.Children {
			if child.Type == parser.NodeTypeOperand {
				subOps = append(subOps, p.formatOperand(child))
			}
		}
		if len(subOps) > 0 {
			sb.WriteString("(")
			sb.WriteString(strings.Join(subOps, " "))
			sb.WriteString(")")
		}
	}

	return sb.String()
}

// formatOperandParameter formats the parameter value of an operand
func (p *Provider) formatOperandParameter(param *parser.Node) string {
	if param == nil {
		return ""
	}

	// Preserve the original parameter value - don't reformat it
	// This keeps commas, spaces, and other separators as the user wrote them
	return param.Value
}

// runeCount returns the number of runes in a string
func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
