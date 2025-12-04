package completion

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/langid"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides code completion
type Provider struct {
	statements map[string]data.MCSStatement
}

// NewProvider creates a new completion provider with shared data
func NewProvider(store *data.Store) *Provider {
	return &Provider{
		statements: store.Statements,
	}
}

// GetCompletions returns completion items for the given position
// getReasonCompletions returns context-sensitive completions for ++HOLD REASON operand
// Behavior depends on which hold type is set: ERROR, SYSTEM, FIXCAT, or USER
func (p *Provider) getReasonCompletions(operandDef *data.Operand, fullText string, currentLine int) []lsp.CompletionItem {
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

	// Order statements for consistent display (Control MCS - alphabetically sorted)
	// TODO: This list should be generated dynamically from smpe.json instead of being hard-coded (see TODO.md)
	order := []string{"++APAR", "++ASSIGN", "++DELETE", "++FEATURE", "++FUNCTION", "++HOLD", "++IF", "++JAR", "++JARUPD", "++JCLIN", "++MAC", "++MOD", "++USERMOD", "++VER", "++ZAP"}

	// Add all Data Element MCS base names (will be expanded with language variants below)
	// TODO: This list should be generated dynamically from smpe.json instead of being hard-coded (see TODO.md)
	dataElementOrder := []string{
		"++BOOK", "++BSIND", "++CGM", "++CLIST", "++DATA", "++DATA1", "++DATA2",
		"++DATA3", "++DATA4", "++DATA5", "++DATA6", "++EXEC", "++FONT", "++GDF",
		"++HELP", "++IMG", "++MSG", "++PARM", "++PNL", "++PROBJ", "++PROC",
		"++PRODXML", "++PRSRC", "++PSEG", "++PUBLB", "++SAMP", "++SKL", "++TBL",
		"++TEXT", "++USER1", "++USER2", "++USER3", "++USER4", "++USER5",
		"++UTIN", "++UTOUT",
	}
	order = append(order, dataElementOrder...)

	for _, name := range order {
		stmt, ok := p.statements[name]
		if !ok {
			continue
		}

		// If this statement requires language variants, generate all variants
		if stmt.LanguageVariants {
			for _, langID := range langid.NationalLanguageIdentifiers {
				variantName := name + langID

				// Create insert text with parameter placeholder if needed
				insertText := variantName
				if stmt.Parameter != "" {
					insertText += "($1)" // $1 = cursor position inside parentheses
				}

				item := lsp.CompletionItem{
					Label:            variantName,
					Kind:             lsp.CompletionItemKindKeyword,
					Detail:           stmt.Type,
					Documentation:    stmt.Description,
					InsertTextFormat: lsp.InsertTextFormatSnippet,
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
		} else {
			// Normal statement without language variants
			// Create insert text with parameter placeholder if needed
			insertText := name
			if stmt.Parameter != "" {
				insertText += "($1)" // $1 = cursor position inside parentheses
			}

			item := lsp.CompletionItem{
				Label:            name,
				Kind:             lsp.CompletionItemKindKeyword,
				Detail:           stmt.Type,
				Documentation:    stmt.Description,
				InsertTextFormat: lsp.InsertTextFormatSnippet,
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
	}

	return items
}

// getOperandCompletions returns operand completions based on statement and context
func (p *Provider) getOperandCompletions(stmt data.MCSStatement, statementTextBefore string, currentLine string, line int, character int) []lsp.CompletionItem {
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

	// Special handling for ++MOD: context-sensitive completion based on DELETE operand
	// From syntax_diagrams/mod-delete.png: If DELETE is present, only CSECT, DISTLIB, VERSION are allowed
	// From syntax_diagrams/mod-add_replace.png: If DELETE is NOT present, all operands except DELETE are allowed
	isModDeleteMode := false
	if stmt.Name == "++MOD" {
		if _, hasDelete := presentOperands["DELETE"]; hasDelete {
			isModDeleteMode = true
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

		// Context-sensitive filtering for ++MOD based on DELETE mode
		if stmt.Name == "++MOD" {
			if isModDeleteMode {
				// DELETE mode: only CSECT, DISTLIB, VERSION allowed (DELETE already present)
				if primaryName != "CSECT" && primaryName != "DISTLIB" && primaryName != "VERSION" {
					continue
				}
			} else {
				// ADD/REPLACE mode: check if any non-DELETE operands are present
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
			Detail:           "Operand",
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
				Detail:           "Operand",
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

// isInsideInlineData checks if the cursor is inside inline data
// Inline data is present after statements with inline_data: true (++JCLIN, ++MAC, etc.)
// when no external data operands (FROMDS, RELFILE, SSI, TXLIB) are present
