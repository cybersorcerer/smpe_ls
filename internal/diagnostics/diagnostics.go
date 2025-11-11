package diagnostics

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides diagnostics
type Provider struct {
	statements map[string]data.MCSStatement
}

// NewProvider creates a new diagnostics provider with shared data
func NewProvider(store *data.Store) *Provider {
	return &Provider{
		statements: store.Statements,
	}
}

// Analyze analyzes the text and returns diagnostics
func (p *Provider) Analyze(text string) []lsp.Diagnostic {
	logger.Debug("Analyzing text for diagnostics")

	// Initialize as empty array to ensure it serializes as [] not null in JSON
	diagnostics := make([]lsp.Diagnostic, 0)

	// Find all statements in the document
	statements := p.findAllStatements(text)

	for _, stmt := range statements {
		// Check for missing terminator
		if !stmt.HasTerminator {
			diagnostics = append(diagnostics, p.createDiagnostic(
				stmt.StartLine, stmt.StartCol,
				stmt.StartLine, stmt.StartCol+len(stmt.StatementType),
				lsp.SeverityError,
				"Statement must be terminated with '.'",
			))
		}

		// Check for missing or malformed statement parameter
		if stmt.StatementType != "" && stmt.StatementType != "++ASSIGN" {
			stmtDef, ok := p.statements[stmt.StatementType]
			if ok && stmtDef.Parameter != "" {
				if stmt.Parameter == "MALFORMED_MISSING_CLOSE_PAREN" {
					diagnostics = append(diagnostics, p.createDiagnostic(
						stmt.StartLine, stmt.StartCol,
						stmt.StartLine, stmt.StartCol+len(stmt.StatementType),
						lsp.SeverityError,
						"Statement parameter has malformed syntax: missing closing parenthesis ')'",
					))
				} else if stmt.Parameter == "" {
					diagnostics = append(diagnostics, p.createDiagnostic(
						stmt.StartLine, stmt.StartCol,
						stmt.StartLine, stmt.StartCol+len(stmt.StatementType),
						lsp.SeverityError,
						"Missing required parameter: "+stmtDef.Parameter,
					))
				}
			}
		}

		// Validate statement type
		if stmt.StatementType != "" && strings.HasPrefix(stmt.StatementType, "++") {
			if _, ok := p.statements[stmt.StatementType]; !ok {
				diagnostics = append(diagnostics, p.createDiagnostic(
					stmt.StartLine, stmt.StartCol,
					stmt.StartLine, stmt.StartCol+len(stmt.StatementType),
					lsp.SeverityError,
					"Unknown statement type: "+stmt.StatementType,
				))
				continue
			}

			// Validate operands for known statement types
			stmtDiags := p.validateOperands(stmt)
			diagnostics = append(diagnostics, stmtDiags...)
		}
	}

	logger.Debug("Found %d diagnostics", len(diagnostics))
	return diagnostics
}

// StatementInfo contains information about a parsed statement
type StatementInfo struct {
	StatementType  string
	Parameter      string
	Operands       map[string]string
	OperandLines   map[string]int // Track line number for each operand
	StartLine      int
	StartCol       int
	HasTerminator  bool
}

// findAllStatements finds all statements in the text
func (p *Provider) findAllStatements(text string) []StatementInfo {
	var statements []StatementInfo

	lines := strings.Split(text, "\n")
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			i++
			continue
		}

		// Check if line starts with ++
		if strings.HasPrefix(trimmed, "++") {
			stmt := p.parseStatement(lines, i)
			statements = append(statements, stmt)
			i = stmt.StartLine + 1 // Move past this statement

			// Special handling for statements with inline data
			// Check if this statement can have inline data based on smpe.json
			// Statements with inline_data: true can contain JCL, Assembler, REXX, etc.
			// Inline data is present when NO external data source operands (FROMDS, RELFILE, SSI, TXLIB) are specified
			stmtDef, stmtExists := p.statements[stmt.StatementType]
			if stmtExists && stmtDef.InlineData {
				// Check if inline data is present by checking for external data source operands
				hasExternalData := false
				externalDataOperands := []string{"FROMDS", "RELFILE", "SSI", "TXLIB"}
				for operandName := range stmt.Operands {
					for _, extOp := range externalDataOperands {
						if operandName == extOp {
							hasExternalData = true
							break
						}
					}
					if hasExternalData {
						break
					}
				}

				if !hasExternalData {
					// This statement has inline data - skip all lines until we find the next ++ statement or EOF
					// Inline data (JCL, Assembler, REXX, etc.) must not contain lines starting with ++
					for i < len(lines) {
						nextLine := strings.TrimSpace(lines[i])
						if strings.HasPrefix(nextLine, "++") {
							break // Found next MCS statement
						}
						i++
					}
					continue
				}
			}

			// Skip to next statement or end
			for i < len(lines) {
				nextLine := strings.TrimSpace(lines[i])
				if strings.HasPrefix(nextLine, "++") {
					break
				}
				if nextLine == "" || strings.HasPrefix(nextLine, "/*") {
					// Check if we've passed the terminator
					if stmt.HasTerminator {
						break
					}
				}
				i++
			}
		} else {
			i++
		}
	}

	return statements
}

// parseStatement parses a single statement starting at the given line
func (p *Provider) parseStatement(lines []string, startLine int) StatementInfo {
	stmt := StatementInfo{
		Operands:     make(map[string]string),
		OperandLines: make(map[string]int),
		StartLine:    startLine,
		StartCol:     0,
	}

	line := strings.TrimSpace(lines[startLine])

	// Find start column
	for j, ch := range lines[startLine] {
		if ch == '+' {
			stmt.StartCol = j
			break
		}
	}

	// Remove comments before parsing to handle cases like: ++FUNCTION /* comment */ (param)
	cleanLine := removeComments(line)

	// Extract statement type and parameter
	// We need to carefully distinguish between:
	// 1. ++STATEMENT(param) - statement WITH parameter
	// 2. ++STATEMENT OPERAND(...) - statement WITHOUT parameter, followed by operand
	// The key is: the '(' must come immediately after the statement name (possibly with whitespace)

	// Find the statement name first (everything after ++)
	stmtStart := strings.Index(cleanLine, "++")
	if stmtStart != -1 {
		stmtStart += 2
		stmtEnd := stmtStart

		// Read statement name (uppercase letters/numbers)
		for stmtEnd < len(cleanLine) && isOperandChar(cleanLine[stmtEnd]) {
			stmtEnd++
		}

		if stmtEnd > stmtStart {
			stmtName := cleanLine[stmtStart:stmtEnd]
			stmt.StatementType = "++" + stmtName

			// Check if this statement expects a parameter according to smpe.json
			stmtDef, stmtExists := p.statements[stmt.StatementType]
			expectsParameter := stmtExists && stmtDef.Parameter != ""

			// Now check what follows: skip whitespace
			i := stmtEnd
			for i < len(cleanLine) && (cleanLine[i] == ' ' || cleanLine[i] == '\t') {
				i++
			}

			// Parse statement parameter ONLY if:
			// 1. Statement expects a parameter (according to smpe.json)
			// 2. '(' comes directly after statement name (with only whitespace between)
			//
			// Example distinction:
			// ++APAR(A12345)           -> A12345 is statement parameter (APAR expects param)
			// ++APAR (A12345)          -> A12345 is statement parameter (whitespace OK)
			// ++IF FMID(MYFMID)        -> FMID is operand (IF has no param)
			// ++ASSIGN SOURCEID(...)   -> SOURCEID is operand (ASSIGN has no param)

			if expectsParameter && i < len(cleanLine) && cleanLine[i] == '(' {
				// This statement expects a parameter and we found '(' - parse it
				parenCount := 1
				paramStart := i + 1
				paramEnd := paramStart

				for paramEnd < len(cleanLine) && parenCount > 0 {
					if cleanLine[paramEnd] == '(' {
						parenCount++
					} else if cleanLine[paramEnd] == ')' {
						parenCount--
					}
					if parenCount > 0 {
						paramEnd++
					}
				}

				// Check if parentheses are balanced
				if parenCount > 0 {
					// Missing closing parenthesis - mark as malformed
					stmt.Parameter = "MALFORMED_MISSING_CLOSE_PAREN"
				} else if paramEnd <= len(cleanLine) {
					stmt.Parameter = strings.TrimSpace(cleanLine[paramStart:paramEnd])
				}
			}
		}
	}

	// Collect all lines belonging to this statement into one string
	// This is needed to properly handle multi-line operand parameters
	var statementLines []string
	var lineNumbers []int // Track which source line each collected line came from
	currentLine := startLine
	inComment := false // Track if we're inside a multiline comment

	for currentLine < len(lines) {
		textLine := lines[currentLine]
		trimmedLine := strings.TrimSpace(textLine)

		// Check for comment start/end to properly skip multiline comments
		if strings.Contains(trimmedLine, "/*") {
			inComment = true
		}
		if strings.Contains(trimmedLine, "*/") {
			inComment = false
			// Don't skip this line - it might contain a terminator after the */
			// Fall through to process the rest of the line
		} else if inComment {
			// We're inside a comment and this line doesn't end it - skip
			currentLine++
			continue
		}

		// Skip empty lines
		if trimmedLine == "" {
			currentLine++
			continue
		}

		// If this is a new statement (and not the first line), stop collecting
		if currentLine > startLine && strings.HasPrefix(trimmedLine, "++") {
			break
		}

		// Remove comments from the line before adding to statement
		cleanedLine := removeComments(trimmedLine)
		if cleanedLine != "" {
			statementLines = append(statementLines, cleanedLine)
			lineNumbers = append(lineNumbers, currentLine)
		}

		// Check for terminator AFTER collecting the line
		// A terminator is a '.' that appears OUTSIDE of parentheses (after comments are removed)
		// This correctly handles dataset names like DSN(MY.DATA.SET) where dots appear inside parentheses
		if hasTerminatorOutsideParens(cleanedLine) {
			stmt.HasTerminator = true
			break
		}

		currentLine++
	}

	// Parse operands from the complete statement text
	// Join all lines with spaces to create a single text to parse
	completeStatement := strings.Join(statementLines, " ")
	p.parseOperandsFromText(completeStatement, lineNumbers, &stmt)

	// Debug: Log parsed operands for statements with FROMDS
	if _, hasFROMDS := stmt.Operands["FROMDS"]; hasFROMDS {
		logger.Debug("Statement %s at line %d has operands: %v", stmt.StatementType, stmt.StartLine+1, stmt.Operands)
	}

	return stmt
}

// parseOperandsFromText extracts operands from the complete statement text
// lineNumbers maps the position in statementLines to the actual source line number
func (p *Provider) parseOperandsFromText(text string, lineNumbers []int, stmt *StatementInfo) {
	// Store original text before removing comments for line number calculation
	originalText := text

	// First, remove comments from the text
	text = removeComments(text)

	// Remove statement type and its parameter if present
	skipOffset := 0 // Track how much we've skipped from the beginning
	if strings.HasPrefix(text, "++") {
		// Skip the statement name
		i := 2 // skip ++
		for i < len(text) && isOperandChar(text[i]) {
			i++
		}

		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}

		// If this statement has a parameter, skip past it
		// We know it has a parameter if stmt.Parameter is not empty
		if stmt.Parameter != "" && i < len(text) && text[i] == '(' {
			// Find the matching closing parenthesis
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

		skipOffset = i
		text = text[i:]
	}

	i := 0
	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		if i >= len(text) {
			break
		}

		// Read operand name
		nameStart := i
		for i < len(text) && isOperandChar(text[i]) {
			i++
		}

		if i > nameStart {
			operandName := text[nameStart:i]

			// Only process if it looks like an operand (uppercase)
			// Also skip the statement type itself (e.g., "ASSIGN" in "++ASSIGN SOURCEID(...) TO(...)")
			if isOperandName(operandName) && !strings.HasPrefix("++"+operandName, stmt.StatementType) {
				operandValue := ""

				// Skip whitespace after operand name to check for parameter
				// This allows both TO(...) and TO (...) syntax
				for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
					i++
				}

				// Determine the operand's line number by finding which line it appears on
				// Calculate position in original text (before statement type was removed)
				posInOriginal := skipOffset + nameStart
				operandLineNum := p.findLineNumber(originalText, posInOriginal, lineNumbers, stmt.StartLine)

				// Check for parameter
				if i < len(text) && text[i] == '(' {
					parenCount := 1
					i++ // skip opening (
					paramStart := i
					for i < len(text) && parenCount > 0 {
						if text[i] == '(' {
							parenCount++
						} else if text[i] == ')' {
							parenCount--
						}
						if parenCount > 0 {
							i++
						}
					}

					// Check if parentheses are balanced
					if parenCount > 0 {
						// Missing closing parenthesis - store as malformed operand
						// The operand value will be everything from paramStart to end of text
						operandValue = "MALFORMED_MISSING_CLOSE_PAREN"
					} else {
						operandValue = strings.TrimSpace(text[paramStart:i])
						i++ // skip closing )
					}

					stmt.Operands[operandName] = operandValue
					stmt.OperandLines[operandName] = operandLineNum
				} else {
					// Boolean operand without parameter (e.g., DELETE, NOAPARS, etc.)
					// Store it with empty value to indicate it's present
					stmt.Operands[operandName] = ""
					stmt.OperandLines[operandName] = operandLineNum
				}
			}
		} else {
			i++
		}
	}
}

// findLineNumber determines which source line an operand appears on
// based on its position in the joined text
func (p *Provider) findLineNumber(text string, pos int, lineNumbers []int, defaultLine int) int {
	if len(lineNumbers) == 0 {
		return defaultLine
	}

	// Count how many space separators we've passed to determine which line we're on
	// The text was joined with spaces, so each space (that was added by Join) represents a line boundary
	// However, this is tricky because the original text also contains spaces
	// So we use a simpler approach: just return the first line for now
	// TODO: This could be improved by tracking exact character positions
	return lineNumbers[0]
}

// parseOperandsFromLine extracts operands from a single line (DEPRECATED - kept for reference)
// Use parseOperandsFromText instead for proper multi-line parameter support
func (p *Provider) parseOperandsFromLine(line string, lineNum int, stmt *StatementInfo) {
	// First, remove comments from the line
	text := removeComments(line)

	// Remove statement type and its parameter if present
	if strings.HasPrefix(text, "++") {
		// Skip the statement name
		i := 2 // skip ++
		for i < len(text) && isOperandChar(text[i]) {
			i++
		}

		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}

		// If this statement has a parameter, skip past it
		// We know it has a parameter if stmt.Parameter is not empty
		if stmt.Parameter != "" && i < len(text) && text[i] == '(' {
			// Find the matching closing parenthesis
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

	i := 0
	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		if i >= len(text) {
			break
		}

		// Read operand name
		nameStart := i
		for i < len(text) && isOperandChar(text[i]) {
			i++
		}

		if i > nameStart {
			operandName := text[nameStart:i]

			// Only process if it looks like an operand (uppercase)
			// Also skip the statement type itself (e.g., "ASSIGN" in "++ASSIGN SOURCEID(...) TO(...)")
			if isOperandName(operandName) && !strings.HasPrefix("++"+operandName, stmt.StatementType) {
				operandValue := ""

				// Check for parameter
				if i < len(text) && text[i] == '(' {
					parenCount := 1
					i++ // skip opening (
					paramStart := i
					for i < len(text) && parenCount > 0 {
						if text[i] == '(' {
							parenCount++
						} else if text[i] == ')' {
							parenCount--
						}
						if parenCount > 0 {
							i++
						}
					}
					operandValue = text[paramStart:i]
					i++ // skip closing )

					stmt.Operands[operandName] = operandValue
					stmt.OperandLines[operandName] = lineNum
				} else {
					// Boolean operand without parameter (e.g., DELETE, NOAPARS, etc.)
					// Store it with empty value to indicate it's present
					stmt.Operands[operandName] = ""
					stmt.OperandLines[operandName] = lineNum
				}
			}
		} else {
			i++
		}
	}
}

// hasTerminatorOutsideParens checks if a '.' exists outside of parentheses
// This function should be called AFTER removeComments() has been applied
// It correctly handles dataset names like DSN(MY.DATA.SET) where dots appear inside parentheses
func hasTerminatorOutsideParens(text string) bool {
	parenCount := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '(' {
			parenCount++
		} else if text[i] == ')' {
			parenCount--
		} else if text[i] == '.' && parenCount == 0 {
			// Found a '.' outside of parentheses - this is a statement terminator
			return true
		}
	}
	return false
}

// removeComments removes /* ... */ style comments from a line
func removeComments(line string) string {
	result := strings.Builder{}
	i := 0
	for i < len(line) {
		if i < len(line)-1 && line[i] == '/' && line[i+1] == '*' {
			// Start of comment - skip until we find */
			i += 2
			for i < len(line) {
				if i < len(line)-1 && line[i] == '*' && line[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
		} else {
			result.WriteByte(line[i])
			i++
		}
	}
	return result.String()
}

// validateOperands validates operands for a statement
func (p *Provider) validateOperands(stmt StatementInfo) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	stmtDef, ok := p.statements[stmt.StatementType]
	if !ok {
		return diagnostics
	}

	// Check for unknown operands
	validOperands := make(map[string]bool)
	for _, op := range stmtDef.Operands {
		names := strings.Split(op.Name, "|")
		for _, name := range names {
			validOperands[name] = true
		}
	}

	for opName, opLine := range stmt.OperandLines {
		if !validOperands[opName] {
			diagnostics = append(diagnostics, p.createDiagnostic(
				opLine, 0,
				opLine, len(opName),
				lsp.SeverityWarning,
				"Unknown operand '"+opName+"' for statement "+stmt.StatementType,
			))
		}
	}

	// Check for empty operand parameters and malformed parentheses
	// If an operand has a parameter defined in smpe.json, it should not be empty
	for _, op := range stmtDef.Operands {
		names := strings.Split(op.Name, "|")
		for _, name := range names {
			if opValue, exists := stmt.Operands[name]; exists {
				lineNum := stmt.OperandLines[name]

				// Check for malformed parentheses (missing closing paren)
				if opValue == "MALFORMED_MISSING_CLOSE_PAREN" {
					diagnostics = append(diagnostics, p.createDiagnostic(
						lineNum, 0,
						lineNum, len(name),
						lsp.SeverityError,
						"Operand '"+name+"' has malformed parameter: missing closing parenthesis ')'",
					))
					continue // Don't check for empty parameter if it's malformed
				}

				// Check if this operand expects a parameter
				if op.Parameter != "" && strings.TrimSpace(opValue) == "" {
					diagnostics = append(diagnostics, p.createDiagnostic(
						lineNum, 0,
						lineNum, len(name),
						lsp.SeverityError,
						"Operand '"+name+"' requires a parameter: "+op.Parameter,
					))
				}

				// Check if this operand has sub-operands (values array) that need validation
				if len(op.Values) > 0 && strings.Contains(op.Parameter, "(") {
					// This operand has sub-operands - validate them
					subDiags := p.validateSubOperands(opValue, lineNum, name, op.Values)
					diagnostics = append(diagnostics, subDiags...)
				}
			}
		}
	}

	// Check for missing required operands based on syntax diagrams
	// These rules are derived from the syntax diagrams in syntax_diagrams/
	requiredOperands := getRequiredOperands(stmt.StatementType)
	for _, requiredOp := range requiredOperands {
		// Check if any alias of this operand is present
		found := false
		for _, op := range stmtDef.Operands {
			names := strings.Split(op.Name, "|")
			primaryName := names[0]

			if primaryName == requiredOp {
				// Check if this operand (or any of its aliases) is present
				for _, name := range names {
					if _, exists := stmt.Operands[name]; exists {
						found = true
						break
					}
				}
				break
			}
		}

		if !found {
			diagnostics = append(diagnostics, p.createDiagnostic(
				stmt.StartLine, stmt.StartCol,
				stmt.StartLine, stmt.StartCol+len(stmt.StatementType),
				lsp.SeverityWarning,
				"Missing required operand: "+requiredOp,
			))
		}
	}

	// Check for dependency violations
	for _, op := range stmtDef.Operands {
		names := strings.Split(op.Name, "|")
		primaryName := names[0]

		if op.AllowedIf != "" {
			// Check if this operand is present
			operandPresent := false
			for _, name := range names {
				if _, exists := stmt.Operands[name]; exists {
					operandPresent = true
					break
				}
			}

			if operandPresent {
				// Check if dependency is met
				if _, exists := stmt.Operands[op.AllowedIf]; !exists {
					lineNum := stmt.OperandLines[primaryName]
					diagnostics = append(diagnostics, p.createDiagnostic(
						lineNum, 0,
						lineNum, len(primaryName),
						lsp.SeverityInformation,
						primaryName+" requires "+op.AllowedIf+" to be specified",
					))
				}
			}
		}
	}

	// Check for mutually exclusive operands
	for _, op := range stmtDef.Operands {
		names := strings.Split(op.Name, "|")
		primaryName := names[0]

		if op.MutuallyExclusive != "" {
			// Check if this operand is present
			operandPresent := false
			var operandLine int
			for _, name := range names {
				if _, exists := stmt.Operands[name]; exists {
					operandPresent = true
					operandLine = stmt.OperandLines[name]
					break
				}
			}

			if operandPresent {
				// Check if any mutually exclusive operand is also present
				exclusiveOperands := strings.Split(op.MutuallyExclusive, "|")
				for _, exclusive := range exclusiveOperands {
					if _, exists := stmt.Operands[exclusive]; exists {
						diagnostics = append(diagnostics, p.createDiagnostic(
							operandLine, 0,
							operandLine, len(primaryName),
							lsp.SeverityError,
							primaryName+" is mutually exclusive with "+exclusive,
						))
					}
				}
			}
		}
	}

	// Check for duplicate operands
	seen := make(map[string]int)
	for opName, opLine := range stmt.OperandLines {
		if prevLine, exists := seen[opName]; exists {
			diagnostics = append(diagnostics, p.createDiagnostic(
				opLine, 0,
				opLine, len(opName),
				lsp.SeverityHint,
				"Duplicate operand '"+opName+"' (first defined at line "+string(rune(prevLine+1))+")",
			))
		}
		seen[opName] = opLine
	}

	return diagnostics
}

// createDiagnostic creates a diagnostic with proper range
func (p *Provider) createDiagnostic(startLine, startCol, endLine, endCol, severity int, message string) lsp.Diagnostic {
	// Add severity prefix with Unicode symbols for better visual distinction
	var prefix string
	switch severity {
	case lsp.SeverityError:
		prefix = "🔴 "
	case lsp.SeverityWarning:
		prefix = "⚠️ "
	case lsp.SeverityInformation:
		prefix = "ℹ️ "
	case lsp.SeverityHint:
		prefix = "💡 "
	}

	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: startLine, Character: startCol},
			End:   lsp.Position{Line: endLine, Character: endCol},
		},
		Severity: severity,
		Source:   "smpe_ls",
		Message:  prefix + message,
	}
}

// isOperandChar checks if a character is valid in an operand name
func isOperandChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
}

// isOperandName checks if a token looks like an operand name
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

// validateSubOperands validates sub-operands within an operand's parameter
// For example, validates DSN, NUMBER, VOL, UNIT within FROMDS(...)
func (p *Provider) validateSubOperands(paramValue string, lineNum int, parentOperand string, subOperandDefs []data.AllowedValue) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	// Parse sub-operands from the parameter value
	// Example: "DSN(my.test) NUMBER(12) UNIT(SYSDA) VOL()"
	subOperands := p.parseSubOperands(paramValue)

	// Validate each sub-operand
	for subOpName, subOpValue := range subOperands {
		// Find the definition for this sub-operand
		var subOpDef *data.AllowedValue
		for i := range subOperandDefs {
			if subOperandDefs[i].Name == subOpName {
				subOpDef = &subOperandDefs[i]
				break
			}
		}

		if subOpDef == nil {
			// Unknown sub-operand
			diagnostics = append(diagnostics, p.createDiagnostic(
				lineNum, 0,
				lineNum, len(parentOperand),
				lsp.SeverityWarning,
				"Unknown sub-operand '"+subOpName+"' for "+parentOperand,
			))
			continue
		}

		// Check if sub-operand has an empty parameter when it shouldn't
		// Sub-operands with type "string" or "integer" and length > 0 should not be empty
		// Exception: length=0 means no length restriction (can be empty)
		if subOpValue == "" && subOpDef.Length > 0 && (subOpDef.Type == "string" || subOpDef.Type == "integer") {
			diagnostics = append(diagnostics, p.createDiagnostic(
				lineNum, 0,
				lineNum, len(parentOperand),
				lsp.SeverityWarning,
				"Sub-operand '"+subOpName+"' of "+parentOperand+" has empty parameter (expected "+subOpDef.Type+")",
			))
		}

		// Check length constraints for non-empty values
		if subOpValue != "" && subOpDef.Length > 0 && len(subOpValue) > subOpDef.Length {
			diagnostics = append(diagnostics, p.createDiagnostic(
				lineNum, 0,
				lineNum, len(parentOperand),
				lsp.SeverityWarning,
				"Sub-operand '"+subOpName+"' of "+parentOperand+" exceeds maximum length",
			))
		}
	}

	return diagnostics
}

// parseSubOperands parses sub-operands from a parameter value
// Example: "DSN(my.test) NUMBER(12) UNIT(SYSDA) VOL()"
// Returns: map{"DSN": "my.test", "NUMBER": "12", "UNIT": "SYSDA", "VOL": ""}
func (p *Provider) parseSubOperands(paramValue string) map[string]string {
	result := make(map[string]string)
	text := strings.TrimSpace(paramValue)

	i := 0
	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		if i >= len(text) {
			break
		}

		// Read sub-operand name
		nameStart := i
		for i < len(text) && isOperandChar(text[i]) {
			i++
		}

		if i > nameStart {
			subOpName := text[nameStart:i]

			// Skip whitespace before parameter
			for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
				i++
			}

			// Check for parameter
			if i < len(text) && text[i] == '(' {
				parenDepth := 1
				i++ // skip opening (
				paramStart := i

				// Find matching closing paren with depth tracking
				for i < len(text) && parenDepth > 0 {
					if text[i] == '(' {
						parenDepth++
					} else if text[i] == ')' {
						parenDepth--
					}
					if parenDepth > 0 {
						i++
					}
				}

				// Extract parameter value
				paramValue := ""
				if i > paramStart {
					paramValue = strings.TrimSpace(text[paramStart:i])
				}
				result[subOpName] = paramValue

				if parenDepth == 0 {
					i++ // skip closing )
				}
			} else {
				// Sub-operand without parameter (should be rare for sub-operands)
				result[subOpName] = ""
			}
		} else {
			i++
		}
	}

	return result
}

// getRequiredOperands returns the list of required operands for a statement
// These requirements are derived from the syntax diagrams in syntax_diagrams/
func getRequiredOperands(statementType string) []string {
	switch statementType {
	case "++ASSIGN":
		// From syntax_diagrams/assign.png: SOURCEID and TO are required
		return []string{"SOURCEID", "TO"}
	case "++IF":
		// From syntax_diagrams/if.png: FMID and REQ are required, THEN is optional
		return []string{"FMID", "REQ"}
	case "++DELETE":
		// From syntax_diagrams/delete.png: SYSLIB is required
		return []string{"SYSLIB"}
	case "++JAR":
		// From syntax_diagrams/jar-add.png and jar-delete.png:
		// No operands are strictly required beyond the name parameter
		// DISTLIB and SYSLIB are important but not enforced as required here
		// The statement itself requires either add mode (DISTLIB/SYSLIB) or delete mode (DELETE)
		return []string{}
	case "++JARUPD":
		// From syntax_diagrams/jar-upd.png:
		// No operands are strictly required beyond the name parameter
		return []string{}
	case "++VER":
		// From syntax_diagrams/ver.png:
		// No operands are strictly required, all are optional
		return []string{}
	case "++ZAP":
		// From syntax_diagrams/zap.png:
		// No operands are strictly required beyond the name parameter
		// DALIAS and TALIAS are mutually exclusive
		return []string{}
	case "++MAC":
		// From syntax_diagrams/mac.png and mac-delete.png:
		// DISTLIB is required in ADD/UPDATE mode (when DELETE is not specified)
		// In DELETE mode (when DELETE is specified), DISTLIB is optional
		// Since this is conditional, we don't enforce it here
		// The mutually_exclusive validation handles the DELETE operand constraints
		return []string{}
	case "++USERMOD":
		// From syntax_diagrams/usermod.png:
		// No operands are strictly required, all are optional
		// RFDSNPFX has allowed_if dependency on FILES (handled automatically)
		return []string{}
	default:
		// No required operands for other statements (based on current syntax diagrams)
		return []string{}
	}
}
