package diagnostics

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/langid"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// MaxColumn is the maximum column for SMP/E content (columns 1-72)
// Columns 73-80 are ignored by SMP/E
const MaxColumn = 72

// Config holds the configuration for which diagnostics to enable/disable
type Config struct {
	UnknownStatement            bool
	InvalidLanguageId           bool
	UnbalancedParentheses       bool
	MissingTerminator           bool
	MissingParameter            bool
	UnknownOperand              bool
	DuplicateOperand            bool
	EmptyOperandParameter       bool
	MissingRequiredOperand      bool
	DependencyViolation         bool
	MutuallyExclusive           bool
	RequiredGroup               bool
	MissingInlineData           bool
	UnknownSubOperand           bool
	SubOperandValidation        bool
	ContentBeyondColumn72       bool
	StandaloneCommentBetweenMCS bool
}

// DefaultConfig returns a config with all diagnostics enabled
func DefaultConfig() *Config {
	return &Config{
		UnknownStatement:            true,
		InvalidLanguageId:           true,
		UnbalancedParentheses:       true,
		MissingTerminator:           true,
		MissingParameter:            true,
		UnknownOperand:              true,
		DuplicateOperand:            true,
		EmptyOperandParameter:       true,
		MissingRequiredOperand:      true,
		DependencyViolation:         true,
		MutuallyExclusive:           true,
		RequiredGroup:               true,
		MissingInlineData:           true,
		UnknownSubOperand:           true,
		SubOperandValidation:        true,
		ContentBeyondColumn72:       true,
		StandaloneCommentBetweenMCS: true,
	}
}

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

// AnalyzeAST analyzes an AST document and returns diagnostics with default config
// This replaces the old string-based Analyze() method
func (p *Provider) AnalyzeAST(doc *parser.Document) []lsp.Diagnostic {
	return p.AnalyzeASTWithConfig(doc, DefaultConfig())
}

// AnalyzeASTWithConfig analyzes an AST document and returns diagnostics based on config
func (p *Provider) AnalyzeASTWithConfig(doc *parser.Document, config *Config) []lsp.Diagnostic {
	return p.AnalyzeASTWithConfigAndText(doc, config, "")
}

// AnalyzeASTWithConfigAndText analyzes an AST document and returns diagnostics based on config
// The text parameter is needed for column 72 checking
func (p *Provider) AnalyzeASTWithConfigAndText(doc *parser.Document, config *Config, text string) []lsp.Diagnostic {
	logger.Debug("Analyzing AST for diagnostics")

	if config == nil {
		config = DefaultConfig()
	}

	// Initialize as empty array to ensure it serializes as [] not null in JSON
	diagnostics := make([]lsp.Diagnostic, 0)

	// Check for content beyond column 72 (needs original text)
	if config.ContentBeyondColumn72 && text != "" {
		diagnostics = append(diagnostics, p.checkContentBeyondColumn72(text)...)
	}

	// Analyze each statement in the AST
	for _, stmt := range doc.Statements {
		diagnostics = append(diagnostics, p.analyzeStatementWithConfig(stmt, config)...)
	}

	// Check for statements expecting inline data that might be missing it
	if config.MissingInlineData {
		diagnostics = append(diagnostics, p.checkMissingInlineData(doc)...)
	}

	// Check for standalone comments between MCS statements
	if config.StandaloneCommentBetweenMCS {
		diagnostics = append(diagnostics, p.checkStandaloneCommentsBetweenMCS(doc, text)...)
	}

	logger.Debug("Found %d diagnostics from AST", len(diagnostics))
	return diagnostics
}

// analyzeStatementWithConfig analyzes a single statement node with config
func (p *Provider) analyzeStatementWithConfig(stmt *parser.Node, config *Config) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	// Validate statement exists in smpe.json
	if stmt.StatementDef == nil && (config.UnknownStatement || config.InvalidLanguageId) {
		// Check if this looks like a language variant statement with invalid language ID
		baseName, langID, hasLangID := langid.ExtractLanguageID(stmt.Name)
		if hasLangID {
			// Valid language ID but statement doesn't exist - shouldn't happen with proper validation
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				stmt,
				lsp.SeverityError,
				"Unknown statement type: "+baseName+" (with language ID "+langID+")",
			))
		} else {
			// Check if this could be an invalid language variant
			// Try to extract last 3 characters as potential language ID
			if len(stmt.Name) > 5 { // At least "++X" + 3 chars
				potentialLangID := stmt.Name[len(stmt.Name)-3:]
				potentialBase := stmt.Name[:len(stmt.Name)-3]

				// Check if base exists and supports language variants
				if langid.IsLanguageVariantStatement(potentialBase) {
					// This is a language variant statement with invalid language ID
					diagnostics = append(diagnostics, p.createDiagnosticFromNode(
						stmt,
						lsp.SeverityError,
						"Invalid language identifier '"+potentialLangID+"' for statement "+potentialBase,
					))
				} else {
					// Just an unknown statement
					diagnostics = append(diagnostics, p.createDiagnosticFromNode(
						stmt,
						lsp.SeverityError,
						"Unknown statement type: "+stmt.Name,
					))
				}
			} else {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityError,
					"Unknown statement type: "+stmt.Name,
				))
			}
		}
		return diagnostics
	}

	// Note: language_variants means the statement CAN have a language identifier suffix,
	// not that it MUST have one. Per syntax diagram, ++SAMP and ++SAMPENU are both valid.

	// Check for unbalanced parentheses first (more specific error)
	if config.UnbalancedParentheses {
		if stmt.UnbalancedParens > 0 {
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				stmt,
				lsp.SeverityError,
				"Missing closing parenthesis ')'",
			))
		} else if stmt.UnbalancedParens < 0 {
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				stmt,
				lsp.SeverityError,
				"Missing opening parenthesis '(' or extra closing parenthesis ')'",
			))
		}
	}

	// Check for missing terminator (only if parens are balanced)
	if config.MissingTerminator && !stmt.HasTerminator && stmt.UnbalancedParens == 0 {
		diagnostics = append(diagnostics, p.createDiagnosticFromNode(
			stmt,
			lsp.SeverityError,
			"Statement must be terminated with '.'",
		))
	}

	// Check for required statement parameter and its length
	if config.MissingParameter && stmt.StatementDef != nil && stmt.StatementDef.Parameter != "" {
		var paramNode *parser.Node
		for _, child := range stmt.Children {
			if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
				paramNode = child
				break
			}
		}

		if paramNode == nil {
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				stmt,
				lsp.SeverityError,
				"Missing required parameter: "+stmt.StatementDef.Parameter,
			))
		} else if stmt.StatementDef.Length > 0 && len(paramNode.Value) > stmt.StatementDef.Length {
			// Check parameter length
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				paramNode,
				lsp.SeverityWarning,
				fmt.Sprintf("⚠️ Parameter '%s' exceeds maximum length (%d > %d)",
					stmt.StatementDef.Parameter, len(paramNode.Value), stmt.StatementDef.Length),
			))
		}
	}

	// Collect operands from children
	operands := make(map[string]*parser.Node)
	var operandList []*parser.Node

	for _, child := range stmt.Children {
		if child.Type == parser.NodeTypeOperand {
			operandList = append(operandList, child)
			operands[child.Name] = child
		}
	}

	// Validate operands
	diagnostics = append(diagnostics, p.validateOperandsASTWithConfig(stmt, operands, operandList, config)...)

	return diagnostics
}

// validateOperandsASTWithConfig validates operands for a statement using AST with config
func (p *Provider) validateOperandsASTWithConfig(stmt *parser.Node, operands map[string]*parser.Node, operandList []*parser.Node, config *Config) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	stmtDef := stmt.StatementDef
	if stmtDef == nil {
		return diagnostics
	}

	// Build valid operand map
	validOperands := make(map[string]bool)
	operandDefs := make(map[string]*parser.Node) // name -> first occurrence
	for _, opDef := range stmtDef.Operands {
		names := strings.Split(opDef.Name, "|")
		for _, name := range names {
			validOperands[name] = true
		}
	}

	// Check for unknown operands
	if config.UnknownOperand {
		for opName, opNode := range operands {
			if !validOperands[opName] {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					opNode,
					lsp.SeverityWarning,
					"Unknown operand '"+opName+"' for statement "+stmt.Name,
				))
			}
		}
	}

	// Check for duplicate operands
	seen := make(map[string]*parser.Node)
	for _, opNode := range operandList {
		if prevNode, exists := seen[opNode.Name]; exists {
			if config.DuplicateOperand {
				// Create diagnostic pointing to the duplicate
				msg := "Duplicate operand '" + opNode.Name + "'"
				if prevNode.Position.Line != opNode.Position.Line {
					// Only mention line if different
					msg += " (first occurrence at line " + strconv.Itoa(prevNode.Position.Line+1) + ")"
				}
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					opNode,
					lsp.SeverityHint,
					msg,
				))
			}
		} else {
			seen[opNode.Name] = opNode
			operandDefs[opNode.Name] = opNode
		}
	}

	// Check for empty operand parameters when required
	for _, op := range stmtDef.Operands {
		names := strings.Split(op.Name, "|")
		for _, name := range names {
			if opNode, exists := operands[name]; exists {
				// Check if this operand expects a parameter
				if config.EmptyOperandParameter && op.Parameter != "" {
					// Check if operand has children (either parameters or sub-operands)
					hasParam := false
					for _, child := range opNode.Children {
						// Accept either parameter nodes OR operand nodes (for sub-operands like FROMDS)
						if child.Type == parser.NodeTypeParameter && strings.TrimSpace(child.Value) != "" {
							hasParam = true
							break
						}
						if child.Type == parser.NodeTypeOperand {
							// This is a sub-operand (e.g., DSN in FROMDS)
							hasParam = true
							break
						}
					}

					if !hasParam {
						diagnostics = append(diagnostics, p.createDiagnosticFromNode(
							opNode,
							lsp.SeverityError,
							"Operand '"+name+"' requires a parameter: "+op.Parameter,
						))
					}
				}

				// Check length constraints for operand parameters
				// Only check if operand has a length defined and has a parameter value
				if op.Length > 0 {
					for _, child := range opNode.Children {
						if child.Type == parser.NodeTypeParameter {
							paramValue := strings.TrimSpace(child.Value)
							// Remove surrounding quotes if present (for string values)
							if len(paramValue) >= 2 && paramValue[0] == '\'' && paramValue[len(paramValue)-1] == '\'' {
								paramValue = paramValue[1 : len(paramValue)-1]
							}
							if paramValue != "" && len(paramValue) > op.Length {
								diagnostics = append(diagnostics, p.createDiagnosticFromNode(
									child,
									lsp.SeverityWarning,
									fmt.Sprintf("⚠️ Operand '%s' parameter exceeds maximum length (%d > %d)", name, len(paramValue), op.Length),
								))
							}
							break
						}
					}
				}

				// Check if this operand has sub-operands (values array) that need validation
				if (config.UnknownSubOperand || config.SubOperandValidation) && len(op.Values) > 0 && strings.Contains(op.Parameter, "(") {
					// This operand has sub-operands - validate them
					subDiags := p.validateSubOperandsASTWithConfig(opNode, op.Values, config)
					diagnostics = append(diagnostics, subDiags...)
				}
			}
		}
	}

	// Check for missing required operands
	// Note: Operands with required_group are handled separately below
	if config.MissingRequiredOperand {
		requiredOperands := getRequiredOperands(stmt.Name)
		for _, requiredOp := range requiredOperands {
			// Check if any alias of this operand is present
			found := false
			isRequiredGroup := false
			for _, op := range stmtDef.Operands {
				names := strings.Split(op.Name, "|")
				primaryName := names[0]

				if primaryName == requiredOp {
					// Skip if this is part of a required_group (handled separately)
					if op.RequiredGroup {
						isRequiredGroup = true
						break
					}

					// Check if this operand (or any of its aliases) is present
					for _, name := range names {
						if _, exists := operands[name]; exists {
							found = true
							break
						}
					}
					break
				}
			}

			if !found && !isRequiredGroup {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityWarning,
					"Missing required operand: "+requiredOp,
				))
			}
		}
	}

	// Check for dependency violations (allowed_if)
	if config.DependencyViolation {
		for _, op := range stmtDef.Operands {
			names := strings.Split(op.Name, "|")
			primaryName := names[0]

			if op.AllowedIf != "" {
				// Check if this operand is present
				operandPresent := false
				var operandNode *parser.Node
				for _, name := range names {
					if node, exists := operands[name]; exists {
						operandPresent = true
						operandNode = node
						break
					}
				}

				if operandPresent {
					// Check if dependency is met
					if _, exists := operands[op.AllowedIf]; !exists {
						diagnostics = append(diagnostics, p.createDiagnosticFromNode(
							operandNode,
							lsp.SeverityInformation,
							primaryName+" requires "+op.AllowedIf+" to be specified",
						))
					}
				}
			}
		}
	}

	// Check for mutually exclusive operands
	if config.MutuallyExclusive {
		for _, op := range stmtDef.Operands {
			names := strings.Split(op.Name, "|")
			primaryName := names[0]

			if op.MutuallyExclusive != "" {
				// Check if this operand is present
				operandPresent := false
				var operandNode *parser.Node
				for _, name := range names {
					if node, exists := operands[name]; exists {
						operandPresent = true
						operandNode = node
						break
					}
				}

				if operandPresent {
					// Check if any mutually exclusive operand is also present
					exclusiveOperands := strings.Split(op.MutuallyExclusive, "|")
					for _, exclusive := range exclusiveOperands {
						if _, exists := operands[exclusive]; exists {
							diagnostics = append(diagnostics, p.createDiagnosticFromNode(
								operandNode,
								lsp.SeverityError,
								primaryName+" is mutually exclusive with "+exclusive,
							))
						}
					}
				}
			}
		}
	}

	// Check for required_group: when multiple operands are marked as required + required_group,
	// at least one of them must be present
	if config.RequiredGroup {
		requiredGroups := make(map[string][]string) // required_group_id -> list of operand names
		for _, op := range stmtDef.Operands {
			if op.Required && op.RequiredGroup && op.RequiredGroupID != "" {
				// Use the required_group_id as the group key
				names := strings.Split(op.Name, "|")
				requiredGroups[op.RequiredGroupID] = append(requiredGroups[op.RequiredGroupID], names[0])
			}
		}

		// For each required group, check if at least one member is present
		for _, groupMembers := range requiredGroups {
			atLeastOnePresent := false
			for _, member := range groupMembers {
				// Check all aliases for this member
				for _, op := range stmtDef.Operands {
					names := strings.Split(op.Name, "|")
					if names[0] == member {
						// Check if any alias is present
						for _, name := range names {
							if _, exists := operands[name]; exists {
								atLeastOnePresent = true
								break
							}
						}
						break
					}
				}
				if atLeastOnePresent {
					break
				}
			}

			if !atLeastOnePresent {
				// Build a human-readable list of options
				optionsList := strings.Join(groupMembers, ", ")
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityError,
					"One of the following operands must be specified: "+optionsList,
				))
			}
		}
	}

	// Special validation for ++MOVE based on syntax diagrams
	// From syntax_diagrams/move-distlib.png and move-syslib.png
	if stmt.Name == "++MOVE" {
		hasDistlib := operands["DISTLIB"] != nil
		hasSyslib := operands["SYSLIB"] != nil

		// DISTLIB mode validation
		if hasDistlib {
			// TODISTLIB is required when DISTLIB is present
			if operands["TODISTLIB"] == nil {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityError,
					"TODISTLIB is required when DISTLIB is specified",
				))
			}

			// One of MAC, MOD, or SRC is required in DISTLIB mode
			if operands["MAC"] == nil && operands["MOD"] == nil && operands["SRC"] == nil {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityError,
					"One of MAC, MOD, or SRC is required when DISTLIB is specified",
				))
			}
		}

		// SYSLIB mode validation
		if hasSyslib {
			// TOSYSLIB is required when SYSLIB is present
			if operands["TOSYSLIB"] == nil {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityError,
					"TOSYSLIB is required when SYSLIB is specified",
				))
			}

			// One of MAC, SRC, LMOD, or FMID is required in SYSLIB mode
			if operands["MAC"] == nil && operands["SRC"] == nil && operands["LMOD"] == nil && operands["FMID"] == nil {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityError,
					"One of MAC, SRC, LMOD, or FMID is required when SYSLIB is specified",
				))
			}
		}

		// At least one mode must be specified (DISTLIB or SYSLIB)
		if !hasDistlib && !hasSyslib {
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				stmt,
				lsp.SeverityError,
				"Either DISTLIB or SYSLIB must be specified",
			))
		}
	}

	return diagnostics
}

// checkContentBeyondColumn72 checks for content that extends beyond column 72
// Per IBM documentation, columns 73-80 are ignored by SMP/E
func (p *Provider) checkContentBeyondColumn72(text string) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic
	lines := strings.Split(text, "\n")

	for lineNum, line := range lines {
		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Check if line has content beyond column 72 (0-indexed: position 72+)
		// We need to count runes, not bytes, for proper Unicode support
		runes := []rune(line)
		if len(runes) > MaxColumn {
			// Check if the content beyond column 72 is just whitespace
			beyondContent := strings.TrimSpace(string(runes[MaxColumn:]))
			if beyondContent != "" {
				// There's actual content beyond column 72
				diagnostics = append(diagnostics, lsp.Diagnostic{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      lineNum,
							Character: MaxColumn,
						},
						End: lsp.Position{
							Line:      lineNum,
							Character: len(runes),
						},
					},
					Severity: lsp.SeverityError,
					Source:   "smpe_ls",
					Message:  "🔴 Content beyond column 72 will be ignored by SMP/E",
				})
			}
		}
	}

	return diagnostics
}

// checkStandaloneCommentsBetweenMCS checks for comments that stand alone between MCS statements
// Per IBM rules: Comments are allowed on the same line as the terminator (.) and can span multiple lines,
// but a comment that starts on a NEW line between statements causes a SMP/E syntax error.
// This also applies to:
// - Comments BEFORE the first statement in the file
// - Comments AFTER the last statement in the file
// Note: Statements with inline_data (++MAC, ++SRC, etc.) are excluded from this check
// because comments after them are part of inline data.
func (p *Provider) checkStandaloneCommentsBetweenMCS(doc *parser.Document, text string) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	if text == "" {
		return diagnostics
	}

	lines := strings.Split(text, "\n")

	// Helper function to find the end line of a statement (including terminator)
	findStmtEndLine := func(stmt *parser.Node) int {
		endLine := stmt.Position.Line
		for _, child := range stmt.Children {
			if child.Position.Line > endLine {
				endLine = child.Position.Line
			}
		}

		// Also check for the terminator line - it might be on a later line
		// Search from endLine forward until we find the terminator or next statement
		for i := endLine; i < len(lines); i++ {
			line := lines[i]
			// Remove comments to find the actual terminator
			cleanLine := line
			for {
				start := strings.Index(cleanLine, "/*")
				if start == -1 {
					break
				}
				end := strings.Index(cleanLine[start:], "*/")
				if end == -1 {
					cleanLine = cleanLine[:start]
					break
				}
				cleanLine = cleanLine[:start] + cleanLine[start+end+2:]
			}
			cleanLine = strings.TrimSpace(cleanLine)

			if strings.HasSuffix(cleanLine, ".") || cleanLine == "." {
				endLine = i
				break
			}
			// Stop if we hit another statement
			if strings.HasPrefix(strings.TrimSpace(line), "++") && i > stmt.Position.Line {
				break
			}
		}
		return endLine
	}

	// Helper function to check if a statement expects inline data
	// A statement expects inline data if:
	// 1. inline_data is true in smpe.json AND
	// 2. NO external data source operands (FROMDS, RELFILE, TXLIB) AND
	// 3. NO DELETE operand (DELETE means deletion mode, no inline data needed)
	stmtExpectsInlineData := func(stmt *parser.Node) bool {
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

	// Helper function to check for standalone comments in a line range
	// Returns the number of comment lines found (for tracking multi-line comments)
	checkLinesForStandaloneComments := func(startLine, endLine int, isBeforeFirstStmt bool) {
		for lineNum := startLine; lineNum < endLine; lineNum++ {
			if lineNum >= len(lines) {
				break
			}

			line := lines[lineNum]
			trimmed := strings.TrimSpace(line)

			// Skip empty lines
			if trimmed == "" {
				continue
			}

			// Check if this line starts a comment (/* at the beginning or after whitespace only)
			if strings.HasPrefix(trimmed, "/*") {
				// This is a standalone comment - ERROR
				var message string
				if isBeforeFirstStmt {
					message = "🔴 Comment not allowed before first MCS statement - SMP/E syntax error"
				} else {
					message = "🔴 Comment not allowed between MCS statements - SMP/E syntax error"
				}
				diagnostics = append(diagnostics, lsp.Diagnostic{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      lineNum,
							Character: strings.Index(line, "/*"),
						},
						End: lsp.Position{
							Line:      lineNum,
							Character: len(line),
						},
					},
					Severity: lsp.SeverityError,
					Source:   "smpe_ls",
					Message:  message,
				})
				// Skip remaining lines of this multi-line comment
				if !strings.Contains(trimmed, "*/") {
					for lineNum++; lineNum < endLine && lineNum < len(lines); lineNum++ {
						if strings.Contains(lines[lineNum], "*/") {
							break
						}
					}
				}
			}
			// Non-empty lines that are not comments and not statements are ignored
			// (could be continuation of multi-line comment from terminator line)
		}
	}

	// Check BEFORE the first statement (comments before any MCS statement are errors)
	if len(doc.Statements) > 0 {
		firstStmtStartLine := doc.Statements[0].Position.Line
		if firstStmtStartLine > 0 {
			checkLinesForStandaloneComments(0, firstStmtStartLine, true)
		}
	} else {
		// No statements at all - check entire file for comments
		checkLinesForStandaloneComments(0, len(lines), true)
	}

	// Check between consecutive statements
	for i := 0; i < len(doc.Statements)-1; i++ {
		currentStmt := doc.Statements[i]

		// Skip check if current statement expects inline data
		// (lines after it are inline data, not standalone comments)
		if stmtExpectsInlineData(currentStmt) {
			continue
		}

		currentStmtEndLine := findStmtEndLine(currentStmt)
		nextStmtStartLine := doc.Statements[i+1].Position.Line

		// Check lines between currentStmtEndLine and nextStmtStartLine
		// Start from line AFTER the terminator line
		checkLinesForStandaloneComments(currentStmtEndLine+1, nextStmtStartLine, false)
	}

	// Check after the last statement (comments after last statement are also errors)
	if len(doc.Statements) > 0 {
		lastStmt := doc.Statements[len(doc.Statements)-1]

		// Skip if last statement expects inline data
		if !stmtExpectsInlineData(lastStmt) {
			lastStmtEndLine := findStmtEndLine(lastStmt)

			// Check all lines after the last statement until end of file
			checkLinesForStandaloneComments(lastStmtEndLine+1, len(lines), false)
		}
	}

	return diagnostics
}

// checkMissingInlineData checks if statements expecting inline data actually have it
func (p *Provider) checkMissingInlineData(doc *parser.Document) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	// If a statement expecting inline data is followed by another statement (or comment + statement),
	// it means the inline data is missing
	for _, stmt := range doc.StatementsExpectingInline {
		// Check if statement has operands that indicate data is NOT inline
		// FROMDS, RELFILE, TXLIB mean data comes from elsewhere
		// DELETE is a special case for HFS that removes files (no inline data needed)
		hasExternalDataSource := false
		for _, child := range stmt.Children {
			if child.Type == parser.NodeTypeOperand {
				opName := child.Name
				if opName == "FROMDS" || opName == "RELFILE" || opName == "TXLIB" || opName == "DELETE" {
					hasExternalDataSource = true
					break
				}
			}
		}

		// Skip inline data check if data comes from external source
		if hasExternalDataSource {
			continue
		}

		// Find the next statement after this one in the main statements list
		stmtIndex := -1
		for j, s := range doc.Statements {
			if s == stmt {
				stmtIndex = j
				break
			}
		}

		if !stmt.HasInlineData {
			if stmtIndex != -1 && stmtIndex < len(doc.Statements)-1 {
				// Statement is not the last one - inline data should come before next statement
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityWarning,
					p.getMissingInlineDataMessage(stmt)+" before next statement",
				))
			} else if stmtIndex == len(doc.Statements)-1 {
				// This is the last statement in the document - inline data is missing
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityWarning,
					p.getMissingInlineDataMessage(stmt),
				))
			}
		}
	}

	return diagnostics
}

// getMissingInlineDataMessage returns a statement-specific message for missing inline data
func (p *Provider) getMissingInlineDataMessage(stmt *parser.Node) string {
	// Build list of alternative operands based on statement type
	var alternatives []string

	// Check which operands are available for this statement
	if stmt.StatementDef != nil {
		for _, op := range stmt.StatementDef.Operands {
			opNames := strings.Split(op.Name, "|")
			primaryName := opNames[0]

			// These operands indicate external data sources
			if primaryName == "FROMDS" || primaryName == "RELFILE" ||
				primaryName == "TXLIB" {
				alternatives = append(alternatives, primaryName)
			}
		}
	}

	// Build the message
	baseMsg := stmt.Name + " expects inline data"

	if len(alternatives) > 0 {
		return baseMsg + " or one of " + strings.Join(alternatives, ", ")
	}

	return baseMsg + " but none found"
}

// validateSubOperandsASTWithConfig validates sub-operands within an operand's parameter using AST
// For example, validates DSN, NUMBER, VOL, UNIT within FROMDS(...)
func (p *Provider) validateSubOperandsASTWithConfig(operandNode *parser.Node, subOperandDefs []data.AllowedValue, config *Config) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	// Build a map of sub-operand definitions for quick lookup
	// Support pipe-separated aliases (e.g., "AMODE|AMOD")
	subOpDefMap := make(map[string]*data.AllowedValue)
	for i := range subOperandDefs {
		names := strings.Split(subOperandDefs[i].Name, "|")
		for _, name := range names {
			subOpDefMap[strings.TrimSpace(name)] = &subOperandDefs[i]
		}
	}

	// Iterate through the children of the operand node to find sub-operands
	for _, child := range operandNode.Children {
		if child.Type == parser.NodeTypeOperand {
			// This is a sub-operand (e.g., DSN, VOL, UNIT inside FROMDS)
			subOpDef, exists := subOpDefMap[child.Name]
			if !exists {
				// Unknown sub-operand
				if config.UnknownSubOperand {
					diagnostics = append(diagnostics, p.createDiagnosticFromNode(
						child,
						lsp.SeverityWarning,
						"Unknown sub-operand '"+child.Name+"' for "+operandNode.Name,
					))
				}
				continue
			}

			// Check if sub-operand has a parameter when it should
			// Sub-operands with type "string" or "integer" and length > 0 should not be empty
			hasParam := false
			var paramValue string
			for _, subChild := range child.Children {
				if subChild.Type == parser.NodeTypeParameter {
					hasParam = true
					paramValue = strings.TrimSpace(subChild.Value)
					break
				}
			}

			if config.SubOperandValidation {
				// Check if parameter is empty when it shouldn't be
				if subOpDef.Length > 0 && (subOpDef.Type == "string" || subOpDef.Type == "integer") {
					if !hasParam || paramValue == "" {
						diagnostics = append(diagnostics, p.createDiagnosticFromNode(
							child,
							lsp.SeverityWarning,
							"Sub-operand '"+child.Name+"' of "+operandNode.Name+" has empty parameter (expected "+subOpDef.Type+")",
						))
					}
				}

				// Check length constraints for non-empty values
				if hasParam && paramValue != "" && subOpDef.Length > 0 && len(paramValue) > subOpDef.Length {
					diagnostics = append(diagnostics, p.createDiagnosticFromNode(
						child,
						lsp.SeverityWarning,
						"Sub-operand '"+child.Name+"' of "+operandNode.Name+" exceeds maximum length",
					))
				}
			}
		}
	}

	return diagnostics
}

// createDiagnosticFromNode creates a diagnostic from an AST node
func (p *Provider) createDiagnosticFromNode(node *parser.Node, severity int, message string) lsp.Diagnostic {
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
			Start: lsp.Position{
				Line:      node.Position.Line,
				Character: node.Position.Character,
			},
			End: lsp.Position{
				Line:      node.Position.Line,
				Character: node.Position.Character + node.Position.Length,
			},
		},
		Severity: severity,
		Source:   "smpe_ls",
		Message:  prefix + message,
	}
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
	case "++MOD":
		// From syntax_diagrams/mod_add_replace.png: DISTLIB is required for ADD/REPLACE mode
		// From syntax_diagrams/mod_delete.png: DELETE mode has no required operands beyond DELETE itself
		// Note: We return DISTLIB as required, but mutually_exclusive validation in smpe.json
		// will handle the case where DELETE is present (which makes DISTLIB optional)
		return []string{"DISTLIB"}
	case "++MOVE":
		// From syntax_diagrams/move-distlib.png and move-syslib.png:
		// DISTLIB mode: DISTLIB + TODISTLIB + one of (MAC|MOD|SRC) required
		// SYSLIB mode: SYSLIB + TOSYSLIB + one of (MAC|SRC|LMOD|FMID) required
		// Complex conditional validation is now handled in validateOperandsAST for ++MOVE
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
	case "++SRC":
		// From syntax_diagrams/src-add-replace.png and src-delete.png:
		// DISTLIB is required in ADD/REPLACE mode (when DELETE is not specified)
		// In DELETE mode, DELETE is specified and makes DISTLIB optional (via mutually_exclusive)
		// We return DISTLIB as required, but mutually_exclusive validation in smpe.json
		// will handle the case where DELETE is present (which makes DISTLIB optional)
		return []string{"DISTLIB"}
	case "++RENAME":
		// From syntax_diagrams/rename.png:
		// TONAME is required (must follow old_name parameter)
		return []string{"TONAME"}
	case "++USERMOD":
		// From syntax_diagrams/usermod.png:
		// No operands are strictly required, all are optional
		// RFDSNPFX has allowed_if dependency on FILES (handled automatically)
		return []string{}
	case "++PRODUCT":
		// From syntax_diagrams/product.png:
		// DESCRIPTION (or DESC) and SREL are required operands
		return []string{"DESCRIPTION", "SREL"}
	case "++PROGRAM":
		// From syntax_diagrams/program-add-replace.png and program-delete.png:
		// DISTLIB is required for ADD/REPLACE mode
		// In DELETE mode, DELETE is specified and makes DISTLIB optional (via mutually_exclusive)
		// We return DISTLIB as required, but mutually_exclusive validation in smpe.json
		// will handle the case where DELETE is present (which makes DISTLIB optional)
		return []string{"DISTLIB"}
	case "++PTF":
		// From syntax_diagrams/ptf.png:
		// No operands are strictly required, all are optional
		// RFDSNPFX has allowed_if dependency on FILES (handled automatically)
		return []string{}
	case "++FEATURE":
		// From syntax_diagrams/feature.png:
		// No operands are strictly required, all are optional
		// The diagram shows DESCRIPTION, FMID, PRODUCT, and REWORK can all be bypassed
		return []string{}
	case "++RELEASE":
		// From syntax_diagrams/release.png:
		// FMID and REASON are required
		// One of ERROR/FIXCAT/SYSTEM/USER is required (mutually_exclusive group)
		// Note: We don't list ERROR/FIXCAT/SYSTEM/USER here because they are handled
		// by the mutually_exclusive required group validation in smpe.json
		return []string{"FMID", "REASON"}
	default:
		// No required operands for other statements (based on current syntax diagrams)
		return []string{}
	}
}
