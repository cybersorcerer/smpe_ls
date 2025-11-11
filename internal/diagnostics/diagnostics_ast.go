package diagnostics

import (
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/langid"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// AnalyzeAST analyzes an AST document and returns diagnostics
// This replaces the old string-based Analyze() method
func (p *Provider) AnalyzeAST(doc *parser.Document) []lsp.Diagnostic {
	logger.Debug("Analyzing AST for diagnostics")

	// Initialize as empty array to ensure it serializes as [] not null in JSON
	diagnostics := make([]lsp.Diagnostic, 0)

	// Analyze each statement in the AST
	for _, stmt := range doc.Statements {
		diagnostics = append(diagnostics, p.analyzeStatement(stmt)...)
	}

	// Check for statements expecting inline data that might be missing it
	diagnostics = append(diagnostics, p.checkMissingInlineData(doc)...)

	logger.Debug("Found %d diagnostics from AST", len(diagnostics))
	return diagnostics
}

// analyzeStatement analyzes a single statement node
func (p *Provider) analyzeStatement(stmt *parser.Node) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	// Validate statement exists in smpe.json
	if stmt.StatementDef == nil {
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

	// Check if statement requires language ID but doesn't have one
	if stmt.StatementDef.LanguageVariants && stmt.LanguageID == "" {
		diagnostics = append(diagnostics, p.createDiagnosticFromNode(
			stmt,
			lsp.SeverityError,
			"Statement "+stmt.Name+" requires a 3-character language identifier suffix (e.g., "+stmt.Name+"ENU)",
		))
	}

	// Check for unbalanced parentheses first (more specific error)
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

	// Check for missing terminator (only if parens are balanced)
	if !stmt.HasTerminator && stmt.UnbalancedParens == 0 {
		diagnostics = append(diagnostics, p.createDiagnosticFromNode(
			stmt,
			lsp.SeverityError,
			"Statement must be terminated with '.'",
		))
	}

	// Check for required statement parameter
	if stmt.StatementDef.Parameter != "" {
		hasParameter := false
		for _, child := range stmt.Children {
			if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
				hasParameter = true
				break
			}
		}

		if !hasParameter {
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				stmt,
				lsp.SeverityError,
				"Missing required parameter: "+stmt.StatementDef.Parameter,
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
	diagnostics = append(diagnostics, p.validateOperandsAST(stmt, operands, operandList)...)

	return diagnostics
}

// validateOperandsAST validates operands for a statement using AST
func (p *Provider) validateOperandsAST(stmt *parser.Node, operands map[string]*parser.Node, operandList []*parser.Node) []lsp.Diagnostic {
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
	for opName, opNode := range operands {
		if !validOperands[opName] {
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				opNode,
				lsp.SeverityWarning,
				"Unknown operand '"+opName+"' for statement "+stmt.Name,
			))
		}
	}

	// Check for duplicate operands
	seen := make(map[string]*parser.Node)
	for _, opNode := range operandList {
		if prevNode, exists := seen[opNode.Name]; exists {
			// Create diagnostic pointing to the duplicate
			msg := "Duplicate operand '" + opNode.Name + "'"
			if prevNode.Position.Line != opNode.Position.Line {
				// Only mention line if different
				msg += " (first occurrence at line " + string(rune(prevNode.Position.Line+1)) + ")"
			}
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				opNode,
				lsp.SeverityHint,
				msg,
			))
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
				if op.Parameter != "" {
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

				// Check if this operand has sub-operands (values array) that need validation
				if len(op.Values) > 0 && strings.Contains(op.Parameter, "(") {
					// This operand has sub-operands - validate them
					subDiags := p.validateSubOperandsAST(opNode, op.Values)
					diagnostics = append(diagnostics, subDiags...)
				}
			}
		}
	}

	// Check for missing required operands
	requiredOperands := getRequiredOperands(stmt.Name)
	for _, requiredOp := range requiredOperands {
		// Check if any alias of this operand is present
		found := false
		for _, op := range stmtDef.Operands {
			names := strings.Split(op.Name, "|")
			primaryName := names[0]

			if primaryName == requiredOp {
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

		if !found {
			diagnostics = append(diagnostics, p.createDiagnosticFromNode(
				stmt,
				lsp.SeverityWarning,
				"Missing required operand: "+requiredOp,
			))
		}
	}

	// Check for dependency violations (allowed_if)
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

	// Check for mutually exclusive operands
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

	return diagnostics
}

// checkMissingInlineData checks if statements expecting inline data actually have it
func (p *Provider) checkMissingInlineData(doc *parser.Document) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	// If a statement expecting inline data is followed by another statement (or comment + statement),
	// it means the inline data is missing
	for i, stmt := range doc.StatementsExpectingInline {
		// Find the next statement after this one in the main statements list
		stmtIndex := -1
		for j, s := range doc.Statements {
			if s == stmt {
				stmtIndex = j
				break
			}
		}

		if stmtIndex != -1 && stmtIndex < len(doc.Statements)-1 {
			// Check if the statement actually has inline data
			// The parser tracked whether actual non-empty, non-comment lines were found
			if !stmt.HasInlineData {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityWarning,
					"Statement "+stmt.Name+" expects inline data but none found before next statement",
				))
			}
		}

		// Also check if this is the last statement expecting inline data
		// and there's no more content after it - that's also missing data
		if i == len(doc.StatementsExpectingInline)-1 && stmtIndex == len(doc.Statements)-1 {
			// This is the last statement in the document and it expects inline data
			if !stmt.HasInlineData {
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					stmt,
					lsp.SeverityWarning,
					"Statement "+stmt.Name+" expects inline data but none found",
				))
			}
		}
	}

	return diagnostics
}

// validateSubOperandsAST validates sub-operands within an operand's parameter using AST
// For example, validates DSN, NUMBER, VOL, UNIT within FROMDS(...)
func (p *Provider) validateSubOperandsAST(operandNode *parser.Node, subOperandDefs []data.AllowedValue) []lsp.Diagnostic {
	var diagnostics []lsp.Diagnostic

	// Build a map of sub-operand definitions for quick lookup
	subOpDefMap := make(map[string]*data.AllowedValue)
	for i := range subOperandDefs {
		subOpDefMap[subOperandDefs[i].Name] = &subOperandDefs[i]
	}

	// Iterate through the children of the operand node to find sub-operands
	for _, child := range operandNode.Children {
		if child.Type == parser.NodeTypeOperand {
			// This is a sub-operand (e.g., DSN, VOL, UNIT inside FROMDS)
			subOpDef, exists := subOpDefMap[child.Name]
			if !exists {
				// Unknown sub-operand
				diagnostics = append(diagnostics, p.createDiagnosticFromNode(
					child,
					lsp.SeverityWarning,
					"Unknown sub-operand '"+child.Name+"' for "+operandNode.Name,
				))
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
