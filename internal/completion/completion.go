package completion

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// MCSStatement represents a MCS statement definition from smpe.json
type MCSStatement struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Parameter   string    `json:"parameter,omitempty"`
	Type        string    `json:"type"`
	Operands    []Operand `json:"operands,omitempty"`
}

// Operand represents an operand definition
type Operand struct {
	Name              string         `json:"name"`
	Parameter         string         `json:"parameter,omitempty"`
	Type              string         `json:"type,omitempty"`
	Length            int            `json:"length,omitempty"`
	Description       string         `json:"description"`
	AllowedIf         string         `json:"allowed_if,omitempty"`
	MutuallyExclusive string         `json:"mutually_exclusive,omitempty"`
	Values            []AllowedValue `json:"values,omitempty"`
}

// AllowedValue represents an allowed value for an operand
type AllowedValue struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Provider provides code completion
type Provider struct {
	statements map[string]MCSStatement
}

// NewProvider creates a new completion provider
func NewProvider(dataPath string) (*Provider, error) {
	// Load MCS definitions from smpe.json
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var statements []MCSStatement
	if err := json.Unmarshal(data, &statements); err != nil {
		return nil, err
	}

	// Build lookup map
	stmtMap := make(map[string]MCSStatement)
	for _, stmt := range statements {
		stmtMap[stmt.Name] = stmt
	}

	return &Provider{
		statements: stmtMap,
	}, nil
}

// GetCompletions returns completion items for the given position
func (p *Provider) GetCompletions(text string, line, character int) []lsp.CompletionItem {
	// Convert line/character to absolute position
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}

	currentLine := lines[line]
	if character < 0 || character > len(currentLine) {
		return nil
	}

	// DEBUG: Log every completion request
	logger.Debug("GetCompletions called - line: %d, character: %d, currentLine: %q", line, character, currentLine)

	// Calculate absolute cursor position
	cursorPos := 0
	for i := 0; i < line; i++ {
		cursorPos += len(lines[i]) + 1 // +1 for newline
	}
	cursorPos += character

	// Get cursor context using statement boundary detection
	ctx := parser.GetCursorContext(text, cursorPos)

	// DEBUG: Log what parser detected
	logger.Debug("Parser context - StatementType: %s, InParameter: %v, StatementText length: %d",
		ctx.StatementType, ctx.InParameter, len(ctx.StatementText))

	// Check if we're typing a MCS statement (at start or after +)
	textBefore := currentLine[:character]
	trimmedBefore := strings.TrimSpace(textBefore)

	// If we're typing + or ++, offer MCS completions with proper TextEdit
	if trimmedBefore == "" || (strings.HasPrefix(trimmedBefore, "+") && len(trimmedBefore) <= 2) {
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

	// If no statement type detected, offer MCS statement completions
	if ctx.StatementType == "" {
		return p.getMCSCompletions(nil)
	}

	// If statement type is incomplete (just +), offer MCS completions
	if strings.HasPrefix(ctx.StatementType, "+") && len(ctx.StatementType) < 3 {
		return p.getMCSCompletions(nil)
	}

	// Get statement definition
	stmt, ok := p.statements[ctx.StatementType]
	if !ok {
		// Unknown statement type, offer all MCS statements
		return p.getMCSCompletions(nil)
	}

	// Check cursor position within statement
	// IMPORTANT: Only use text UP TO the cursor, not the entire statement
	// This prevents seeing operands that come AFTER the cursor position
	statementTextBefore := ""
	if ctx.CursorOffset < len(ctx.StatementText) {
		statementTextBefore = ctx.StatementText[:ctx.CursorOffset]
	} else {
		statementTextBefore = ctx.StatementText
	}

	// For multiline statements, we need to use the full statement text from the parser
	// The parser already correctly handles statement boundaries across multiple lines
	// We should trust the parser's StatementText which includes all lines up to the cursor

	// DEBUG: Log what we're working with
	logger.Debug("Multiline completion - StatementType: %s, StatementTextBefore length: %d, CursorOffset: %d",
		ctx.StatementType, len(statementTextBefore), ctx.CursorOffset)
	logger.Debug("StatementTextBefore: %q", statementTextBefore)

	// If cursor is inside statement parameter parentheses, no completion
	// BUT: if cursor is inside OPERAND parameter parentheses, offer value completion
	if ctx.InParameter {
		// Check if we're inside an operand parameter (not statement parameter)
		// Pattern: OPERAND(cursor)
		operandName, inOperandParam := p.detectOperandParameter(currentLine, character)
		if inOperandParam {
			// Offer operand value completions
			return p.getOperandValueCompletions(stmt, operandName, text, line, character)
		}
		return nil
	}

	// Get operand completions based on statement type and context
	// This will be triggered when:
	// 1. After statement parameter: ++APAR(A12345) <cursor>
	// 2. After an operand: ++APAR(A12345) FILES(2) <cursor>
	// 3. While typing an operand name: ++APAR(A12345) FI<cursor>

	// Debug: Log the statement text being analyzed
	// logger.Debug("Statement text for completion: %q", statementTextBefore)

	return p.getOperandCompletions(stmt, statementTextBefore, currentLine, line, character)
}

// detectOperandParameter checks if cursor is inside an operand parameter
// Returns (operandName, true) if inside operand parameter, ("", false) otherwise
// Pattern: OPERAND(cursor) or OPERAND(partial_text|cursor)
func (p *Provider) detectOperandParameter(line string, character int) (string, bool) {
	if character <= 0 || character > len(line) {
		return "", false
	}

	// Search backwards from cursor to find opening (
	parenPos := -1
	for i := character - 1; i >= 0; i-- {
		if line[i] == '(' {
			parenPos = i
			break
		} else if line[i] == ')' {
			// We're outside or in a closed parameter
			return "", false
		}
	}

	if parenPos == -1 {
		return "", false
	}

	// Extract operand name before (
	nameEnd := parenPos
	nameStart := nameEnd - 1

	// Skip whitespace before (
	for nameStart >= 0 && (line[nameStart] == ' ' || line[nameStart] == '\t') {
		nameStart--
	}

	// Read operand name backwards
	for nameStart >= 0 && isOperandChar(line[nameStart]) {
		nameStart--
	}
	nameStart++ // Adjust to first character of name

	if nameStart >= nameEnd {
		return "", false
	}

	operandName := strings.TrimSpace(line[nameStart:nameEnd])

	// Verify it looks like an operand name (all uppercase)
	if !isOperandName(operandName) {
		return "", false
	}

	// Make sure we're not in the statement parameter
	// Check if this is after ++ at the start of line
	beforeOperand := strings.TrimSpace(line[:nameStart])
	if strings.HasPrefix(beforeOperand, "++") && !strings.Contains(beforeOperand, ")") {
		// This is the statement parameter, not an operand parameter
		return "", false
	}

	return operandName, true
}

// getOperandValueCompletions returns value completions for operand parameters
// This is highly context-sensitive, especially for ++HOLD REASON operand
func (p *Provider) getOperandValueCompletions(stmt MCSStatement, operandName string, fullText string, line int, character int) []lsp.CompletionItem {
	var items []lsp.CompletionItem

	// Find the operand definition
	var operandDef *Operand
	for i := range stmt.Operands {
		names := strings.Split(stmt.Operands[i].Name, "|")
		for _, name := range names {
			if name == operandName {
				operandDef = &stmt.Operands[i]
				break
			}
		}
		if operandDef != nil {
			break
		}
	}

	if operandDef == nil {
		return nil
	}

	// Special handling for ++HOLD REASON operand - context-dependent
	if stmt.Name == "++HOLD" && operandName == "REASON" {
		return p.getReasonCompletions(operandDef, fullText, line)
	}

	// Special handling for ++HOLD RESOLVER operand - show all SYSMOD IDs
	if stmt.Name == "++HOLD" && operandName == "RESOLVER" {
		sysmodIDs := extractSysmodIDs(fullText)
		for _, id := range sysmodIDs {
			items = append(items, lsp.CompletionItem{
				Label:         id,
				Kind:          lsp.CompletionItemKindValue,
				Detail:        "SYSMOD ID",
				Documentation: "SYSMOD identifier from document",
			})
		}
		return items
	}

	// For operands with predefined values, offer them
	if len(operandDef.Values) > 0 {
		for _, value := range operandDef.Values {
			items = append(items, lsp.CompletionItem{
				Label:         value.Name,
				Kind:          lsp.CompletionItemKindValue,
				Documentation: value.Description,
			})
		}
	}

	return items
}

// getReasonCompletions returns context-sensitive completions for ++HOLD REASON operand
// Behavior depends on which hold type is set: ERROR, SYSTEM, FIXCAT, or USER
func (p *Provider) getReasonCompletions(operandDef *Operand, fullText string, currentLine int) []lsp.CompletionItem {
	var items []lsp.CompletionItem

	// Parse the current statement to find hold type
	// We need to look backwards from currentLine to find the statement start
	lines := strings.Split(fullText, "\n")

	// Find statement boundaries
	stmtStart := currentLine
	for stmtStart >= 0 {
		trimmed := strings.TrimSpace(lines[stmtStart])
		if strings.HasPrefix(trimmed, "++HOLD") {
			break
		}
		stmtStart--
	}

	if stmtStart < 0 {
		return nil
	}

	// Scan from statement start to current line to find hold type flags
	hasError := false
	hasSystem := false
	hasFixcat := false
	hasUser := false

	for i := stmtStart; i <= currentLine && i < len(lines); i++ {
		line := strings.ToUpper(strings.TrimSpace(lines[i]))

		// Check for hold type keywords
		if strings.Contains(line, "ERROR") && !strings.Contains(line, "/*") {
			hasError = true
		}
		if strings.Contains(line, "SYSTEM") && !strings.Contains(line, "/*") {
			hasSystem = true
		}
		if strings.Contains(line, "FIXCAT") && !strings.Contains(line, "/*") {
			hasFixcat = true
		}
		if strings.Contains(line, "USER") && !strings.Contains(line, "/*") {
			hasUser = true
		}
	}

	// Context-sensitive completion based on hold type
	if hasError {
		// ERROR hold type → show APAR IDs from document
		sysmodIDs := extractSysmodIDs(fullText)
		for _, id := range sysmodIDs {
			// Filter for APAR-like IDs (typically start with 'A' or similar pattern)
			if strings.HasPrefix(id, "A") {
				items = append(items, lsp.CompletionItem{
					Label:         id,
					Kind:          lsp.CompletionItemKindValue,
					Detail:        "APAR ID",
					Documentation: "APAR identifier from document (error reason)",
				})
			}
		}
	} else if hasSystem {
		// SYSTEM hold type → show predefined system reason IDs from JSON
		for _, value := range operandDef.Values {
			items = append(items, lsp.CompletionItem{
				Label:         value.Name,
				Kind:          lsp.CompletionItemKindValue,
				Detail:        "System Reason ID",
				Documentation: value.Description,
			})
		}
	} else if hasFixcat {
		// FIXCAT hold type → show all SYSMOD IDs from document
		sysmodIDs := extractSysmodIDs(fullText)
		for _, id := range sysmodIDs {
			items = append(items, lsp.CompletionItem{
				Label:         id,
				Kind:          lsp.CompletionItemKindValue,
				Detail:        "SYSMOD ID",
				Documentation: "SYSMOD identifier from document (fix category reason)",
			})
		}
	} else if hasUser {
		// USER hold type → no predefined suggestions (user-defined)
		// Return empty list
		return items
	} else {
		// No hold type specified yet - offer all possibilities as hints
		// Show predefined values from JSON (system reason IDs)
		for _, value := range operandDef.Values {
			items = append(items, lsp.CompletionItem{
				Label:         value.Name,
				Kind:          lsp.CompletionItemKindValue,
				Detail:        "System Reason ID",
				Documentation: value.Description,
			})
		}
	}

	return items
}

// extractSysmodIDs scans the entire document for SYSMOD IDs (APAR, FUNCTION, USERMOD, etc.)
func extractSysmodIDs(text string) []string {
	var ids []string
	seen := make(map[string]bool)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Look for MCS statements with IDs: ++APAR(ID), ++FUNCTION(ID), ++USERMOD(ID)
		if strings.HasPrefix(trimmed, "++") {
			// Find opening parenthesis
			if idx := strings.Index(trimmed, "("); idx != -1 {
				// Find matching closing parenthesis
				parenCount := 1
				paramStart := idx + 1
				paramEnd := paramStart

				for paramEnd < len(trimmed) && parenCount > 0 {
					if trimmed[paramEnd] == '(' {
						parenCount++
					} else if trimmed[paramEnd] == ')' {
						parenCount--
					}
					if parenCount > 0 {
						paramEnd++
					}
				}

				if paramEnd < len(trimmed) && paramEnd > paramStart {
					id := strings.TrimSpace(trimmed[paramStart:paramEnd])
					if id != "" && !seen[id] {
						ids = append(ids, id)
						seen[id] = true
					}
				}
			}
		}
	}

	return ids
}

// getMCSCompletions returns MCS statement completions
func (p *Provider) getMCSCompletions(replaceRange *lsp.Range) []lsp.CompletionItem {
	var items []lsp.CompletionItem

	// Order statements for consistent display
	order := []string{"++APAR", "++ASSIGN", "++DELETE", "++FEATURE", "++FUNCTION", "++HOLD", "++IF", "++JAR", "++JARUPD", "++VER", "++ZAP"}

	for _, name := range order {
		stmt, ok := p.statements[name]
		if !ok {
			continue
		}

		// Create insert text with parameter placeholder if needed
		// Use snippet format to position cursor inside parentheses
		insertText := name
		if stmt.Parameter != "" {
			insertText += "($1)" // $1 = cursor position inside parentheses
		}

		item := lsp.CompletionItem{
			Label:            name,
			Kind:             lsp.CompletionItemKindKeyword,
			Detail:           stmt.Type,
			Documentation:    stmt.Description,
			InsertTextFormat: lsp.InsertTextFormatSnippet, // Enable snippet support
		}

		// If we have a range to replace (typed + characters), use TextEdit
		if replaceRange != nil {
			item.TextEdit = &lsp.TextEdit{
				Range:   *replaceRange,
				NewText: insertText,
			}
		} else {
			item.InsertText = insertText
		}

		items = append(items, item)
	}

	return items
}

// getOperandCompletions returns operand completions based on statement and context
func (p *Provider) getOperandCompletions(stmt MCSStatement, statementTextBefore string, currentLine string, line int, character int) []lsp.CompletionItem {
	var items []lsp.CompletionItem

	// Find the start of the word being typed in the current line (for TextEdit range calculation)
	currentLineText := currentLine[:character]
	wordStart := character
	for wordStart > 0 && isOperandChar(currentLineText[wordStart-1]) {
		wordStart--
	}

	// Remove the incomplete word at cursor position from statement text before parsing
	// This prevents treating "F" as a complete operand name when we're typing "FROMDS"
	statementTextForParsing := statementTextBefore
	if wordStart < character {
		// Calculate how many characters to remove from the end of statementTextBefore
		incompleteWordLen := character - wordStart
		if len(statementTextForParsing) >= incompleteWordLen {
			statementTextForParsing = statementTextForParsing[:len(statementTextForParsing)-incompleteWordLen]
		}
	}

	// Parse which operands are already present (excluding the incomplete word at cursor)
	presentOperands := p.parseOperands(statementTextForParsing)

	// Debug: log which operands were found
	// DEBUG: Uncomment to see what operands are detected
	var opNames []string
	for op := range presentOperands {
		opNames = append(opNames, op)
	}
	logger.Debug("Present operands in %q: %v", statementTextForParsing, opNames)

	// Special handling for ++JAR: context-sensitive completion based on DELETE operand
	// From syntax_diagrams/jar-delete.png: If DELETE is present, only DISTLIB and VERSION are allowed
	// From syntax_diagrams/jar-add.png: If DELETE is NOT present, all operands except DELETE are allowed
	isJarDeleteMode := false
	if stmt.Name == "++JAR" {
		if _, hasDelete := presentOperands["DELETE"]; hasDelete {
			isJarDeleteMode = true
		}
	}

	// Calculate range to replace if we're in the middle of typing
	var replaceRange *lsp.Range
	if wordStart < character {
		replaceRange = &lsp.Range{
			Start: lsp.Position{Line: line, Character: wordStart},
			End:   lsp.Position{Line: line, Character: character},
		}
	}

	for _, operand := range stmt.Operands {
		// Get operand names (may have aliases like "DESCRIPTION|DESC")
		names := strings.Split(operand.Name, "|")
		primaryName := names[0]

		// Skip if operand already present (avoid duplicates)
		if _, exists := presentOperands[primaryName]; exists {
			continue
		}

		// Context-sensitive filtering for ++JAR based on DELETE mode
		if stmt.Name == "++JAR" {
			if isJarDeleteMode {
				// DELETE mode: only DISTLIB and VERSION allowed (DELETE already present)
				if primaryName != "DISTLIB" && primaryName != "VERSION" {
					continue
				}
			} else {
				// CREATE mode: check if any non-DELETE operands are present
				hasNonDeleteOperands := false
				for op := range presentOperands {
					if op != "DELETE" {
						hasNonDeleteOperands = true
						break
					}
				}

				// If non-DELETE operands are present, don't offer DELETE anymore
				if hasNonDeleteOperands && primaryName == "DELETE" {
					continue
				}
			}
		}

		// Check if operand has dependency (e.g., RFDSNPFX depends on FILES)
		if operand.AllowedIf != "" {
			// Check if dependency is present
			if _, exists := presentOperands[operand.AllowedIf]; !exists {
				continue // Skip this operand as its dependency is not met
			}
		}

		// Note: mutually_exclusive operands are NOT filtered from completion
		// They should be offered as suggestions, and diagnostics will flag the error if used incorrectly
		// This allows users to see all available options and make informed choices

		// Create insert text with parameter placeholder if needed
		// Use snippet format to position cursor inside parentheses
		insertText := primaryName
		if operand.Parameter != "" {
			insertText += "($1)" // $1 = first tab stop (cursor position)
		}

		// Create documentation with type and length info
		doc := operand.Description
		if operand.Type != "" {
			doc = "Type: " + operand.Type + "\n\n" + doc
		}

		item := lsp.CompletionItem{
			Label:            primaryName,
			Kind:             lsp.CompletionItemKindProperty,
			Detail:           operand.Parameter,
			Documentation:    doc,
			InsertTextFormat: lsp.InsertTextFormatSnippet, // Enable snippet support
		}

		// Use TextEdit if we're replacing typed characters
		if replaceRange != nil {
			item.TextEdit = &lsp.TextEdit{
				Range:   *replaceRange,
				NewText: insertText,
			}
		} else {
			item.InsertText = insertText
		}

		items = append(items, item)

		// Add aliases as separate items
		for i := 1; i < len(names); i++ {
			aliasName := names[i]
			aliasInsertText := aliasName
			if operand.Parameter != "" {
				aliasInsertText += "($1)" // Cursor inside parentheses
			}

			aliasItem := lsp.CompletionItem{
				Label:            aliasName,
				Kind:             lsp.CompletionItemKindProperty,
				Detail:           "alias for " + primaryName,
				Documentation:    doc,
				InsertTextFormat: lsp.InsertTextFormatSnippet, // Enable snippet support
			}

			if replaceRange != nil {
				aliasItem.TextEdit = &lsp.TextEdit{
					Range:   *replaceRange,
					NewText: aliasInsertText,
				}
			} else {
				aliasItem.InsertText = aliasInsertText
			}

			items = append(items, aliasItem)
		}
	}

	return items
}

// isOperandChar checks if a character is valid in an operand name
func isOperandChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
}

// parseOperands extracts operand names from statement text
// Returns a map of operand names that are present or being typed
func (p *Provider) parseOperands(text string) map[string]bool {
	operands := make(map[string]bool)

	// Remove statement name and its parameter (if present)
	// Pattern: ++STATEMENTNAME or ++STATEMENTNAME(param)
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "++"); idx >= 0 {
		text = text[idx+2:]

		// Skip the statement name itself (e.g., ASSIGN, APAR, etc.)
		i := 0
		for i < len(text) && (isOperandChar(text[i])) {
			i++
		}

		// If we found a '(' immediately after the statement name,
		// this is the statement's parameter, skip it
		if i < len(text) && text[i] == '(' {
			parenCount := 1
			i++ // skip opening (
			for i < len(text) && parenCount > 0 {
				if text[i] == '(' {
					parenCount++
				} else if text[i] == ')' {
					parenCount--
				}
				i++
			}
		}

		text = text[i:]
	}

	// Use a more sophisticated parser that tracks operand name + optional parameter
	// Pattern: OPERAND_NAME or OPERAND_NAME(...)
	i := 0
	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}
		if i >= len(text) {
			break
		}

		// Read operand name (uppercase letters/numbers)
		nameStart := i
		for i < len(text) && isOperandChar(text[i]) {
			i++
		}

		if i > nameStart {
			operandName := text[nameStart:i]
			// Only add if it looks like an operand (all uppercase)
			// Allow DELETE as both operand and statement name
			if isOperandName(operandName) && (operandName == "DELETE" || !p.isStatementName(operandName)) {
				operands[operandName] = true
			}

			// Skip optional parameter in parentheses
			if i < len(text) && text[i] == '(' {
				parenCount := 1
				i++ // skip opening (
				for i < len(text) && parenCount > 0 {
					if text[i] == '(' {
						parenCount++
					} else if text[i] == ')' {
						parenCount--
					}
					i++
				}
			}
		} else {
			// Not an operand character, skip it
			i++
		}
	}

	return operands
}

// tokenize splits text into tokens, respecting parentheses
func tokenize(text string) []string {
	var tokens []string
	var current strings.Builder
	inParens := 0

	for i := 0; i < len(text); i++ {
		ch := text[i]

		switch ch {
		case '(':
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			inParens++
		case ')':
			inParens--
		case ' ', '\t', '\n', '\r':
			if inParens == 0 && current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			if inParens == 0 {
				current.WriteByte(ch)
			}
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// isOperandName checks if a token looks like an operand name
// isOperandName checks if a token is a valid operand name
// It must be uppercase AND not a statement name
func isOperandName(token string) bool {
	if len(token) == 0 {
		return false
	}

	// Operand names are typically all uppercase letters
	for _, ch := range token {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}

	return true
}

// isStatementName checks if a token is a known MCS statement name
func (p *Provider) isStatementName(token string) bool {
	_, exists := p.statements["++"+token]
	return exists
}
