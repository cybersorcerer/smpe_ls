package parser

import (
	"strings"
	"unicode/utf8"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/langid"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
)

// byteOffsetToRuneOffset converts a byte offset in a string to a rune (character) offset
// This is needed because LSP uses UTF-16 code units for positions, not byte offsets
func byteOffsetToRuneOffset(s string, byteOffset int) int {
	return utf8.RuneCountInString(s[:byteOffset])
}

// runeCount returns the number of runes (characters) in a string
func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

// NodeType represents the type of AST node
type NodeType int

const (
	NodeTypeDocument  NodeType = iota // Gesamtes Dokument
	NodeTypeStatement                 // MCS Statement (++USERMOD, ++JAR, etc.)
	NodeTypeOperand                   // Operand (DESC, REWORK, FROMDS, etc.)
	NodeTypeParameter                 // Parameter-Wert
	NodeTypeComment                   // Kommentar
)

// Position represents the position of a node in the source text
type Position struct {
	Line      int
	Character int
	Length    int
}

// Node represents an AST node
type Node struct {
	Type     NodeType
	Name     string  // Statement/Operand Name
	Value    string  // Parameter-Wert (raw text)
	Position Position
	Parent   *Node
	Children []*Node

	// Semantische Referenzen zu smpe.json
	StatementDef *data.MCSStatement // Referenz für Statements
	OperandDef   *data.Operand      // Referenz für Operands

	// Statement-specific flags
	HasTerminator    bool   // Only for statement nodes - tracks if '.' terminator was found
	UnbalancedParens int    // Tracks parenthesis imbalance: positive = missing closing, negative = missing opening
	LanguageID       string // Language identifier for language variant statements (e.g., "ENU" from "++FONTENU")
	HasInlineData    bool   // True if actual inline data (non-empty, non-comment lines) was found
	InlineDataLines  int    // Number of actual inline data lines found
}

// ParseError represents a parsing error
type ParseError struct {
	Message  string
	Position Position
}

// Document represents the parsed document with AST
type Document struct {
	Statements                []*Node
	Comments                  []*Node // Top-level comments (block and line comments)
	Errors                    []ParseError
	StatementsExpectingInline []*Node // Statements that expect inline data but may not have it
}

// CollectedStatement represents a statement that may span multiple lines
type CollectedStatement struct {
	Text             string   // Combined text of all lines
	StartLine        int      // Line number where statement starts (0-indexed)
	Lines            []string // Original lines
	UnbalancedParens int      // Parenthesis imbalance: positive = missing closing, negative = missing opening
}

// Parser parses SMP/E MCS text into an AST
type Parser struct {
	statements map[string]data.MCSStatement
	operands   map[string]map[string]data.Operand // statement -> operand name -> operand
}

// NewParser creates a new parser instance
func NewParser(statements map[string]data.MCSStatement) *Parser {
	// Build operand lookup map
	operands := make(map[string]map[string]data.Operand)
	for stmtName, stmt := range statements {
		operands[stmtName] = make(map[string]data.Operand)
		for _, op := range stmt.Operands {
			// Split aliases (e.g., "DESCRIPTION|DESC" means DESC is an alias for DESCRIPTION)
			names := strings.Split(op.Name, "|")
			for _, name := range names {
				operands[stmtName][name] = op
			}
		}
	}

	return &Parser{
		statements: statements,
		operands:   operands,
	}
}

// collectStatements preprocesses lines and collects complete statements
// Handles multiline statements
// Note: lines passed in are already cleaned of SMP/E comments by the first pass in Parse()
func (p *Parser) collectStatements(lines []string) []CollectedStatement {
	var collected []CollectedStatement
	var currentLines []string
	var startLine int
	inStatement := false
	parenDepth := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines when not in a statement
		if trimmed == "" && !inStatement {
			continue
		}

		// Check if this line starts a statement
		if strings.HasPrefix(trimmed, "++") {
			// Save previous statement if any
			if inStatement && len(currentLines) > 0 {
				collected = append(collected, CollectedStatement{
					Text:             strings.Join(currentLines, " "),
					StartLine:        startLine,
					Lines:            currentLines,
					UnbalancedParens: parenDepth,
				})
			}
			// Start new statement
			currentLines = []string{line}
			startLine = lineNum
			inStatement = true
			parenDepth = strings.Count(line, "(") - strings.Count(line, ")")
		} else if inStatement {
			// Continue collecting statement lines
			currentLines = append(currentLines, line)
			parenDepth += strings.Count(line, "(") - strings.Count(line, ")")
		}

		// Check if statement is complete (has terminator and all parens closed)
		if inStatement && hasTerminatorOutsideParens(trimmed) && parenDepth <= 0 {
			collected = append(collected, CollectedStatement{
				Text:             strings.Join(currentLines, " "),
				StartLine:        startLine,
				Lines:            currentLines,
				UnbalancedParens: parenDepth,
			})
			currentLines = nil
			inStatement = false
			parenDepth = 0
		}
	}

	// Handle unclosed statement at end
	if inStatement && len(currentLines) > 0 {
		collected = append(collected, CollectedStatement{
			Text:             strings.Join(currentLines, " "),
			StartLine:        startLine,
			Lines:            currentLines,
			UnbalancedParens: parenDepth,
		})
	}

	return collected
}


// parseLine parses a single line
func (p *Parser) parseLine(line string, lineNum int, doc *Document, currentStatement **Node) {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return
	}

	// Check if this is a statement line (++STATEMENT)
	if strings.HasPrefix(trimmed, "++") {
		idx := strings.Index(line, "++")
		stmtNode := p.parseStatement(line, lineNum, idx)
		if stmtNode != nil {
			doc.Statements = append(doc.Statements, stmtNode)
			*currentStatement = stmtNode
			logger.Debug("Parser: Line %d: Parsed statement %s", lineNum, stmtNode.Name)
		}
	} else if *currentStatement != nil {
		// Continuation line with operands
		operands := p.parseOperands(line, lineNum, 0, *currentStatement)
		(*currentStatement).Children = append((*currentStatement).Children, operands...)

		// Check for terminator on continuation line
		if hasTerminatorOutsideParens(trimmed) {
			(*currentStatement).HasTerminator = true
		}
	}
}

// parseStatement parses a statement starting at the given position
func (p *Parser) parseStatement(line string, lineNum int, startIdx int) *Node {
	// Find statement name (++STATEMENT)
	stmtEnd := startIdx + 2
	for stmtEnd < len(line) && isOperandChar(line[stmtEnd]) {
		stmtEnd++
	}

	if stmtEnd <= startIdx+2 {
		return nil
	}

	stmtName := line[startIdx:stmtEnd]

	// Check if this statement has a language ID suffix
	baseName, langID, hasLangID := langid.ExtractLanguageID(stmtName)

	// Create statement node
	// Convert byte offsets to rune offsets for LSP compatibility
	runeStart := byteOffsetToRuneOffset(line, startIdx)
	runeLength := runeCount(stmtName)

	stmtNode := &Node{
		Type: NodeTypeStatement,
		Name: stmtName,
		Position: Position{
			Line:      lineNum,
			Character: runeStart,
			Length:    runeLength,
		},
		Children:   []*Node{},
		LanguageID: langID,
	}

	// Link to statement definition from smpe.json
	// For language variant statements, use the base name
	lookupName := stmtName
	if hasLangID {
		lookupName = baseName
		logger.Debug("Parser: Statement %s has language ID %s, using base name %s", stmtName, langID, baseName)
	}

	if stmtDef, exists := p.statements[lookupName]; exists {
		stmtNode.StatementDef = &stmtDef
		logger.Debug("Parser: Statement %s linked to definition", lookupName)
	}

	// Check if statement has a parameter (e.g., ++USERMOD(LJS2012))
	i := stmtEnd
	// Skip whitespace before parenthesis
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}

	if i < len(line) && line[i] == '(' {
		// Parse statement parameter
		paramNode := p.parseParameter(line, lineNum, i, stmtNode)
		if paramNode != nil {
			stmtNode.Children = append(stmtNode.Children, paramNode)
			i = paramNode.Position.Character + paramNode.Position.Length + 1 // +1 for closing ')'
		}
	} else {
		i = stmtEnd
	}

	// Parse remaining operands on the same line
	if i < len(line) {
		// Convert byte offset to rune offset for LSP compatibility
		runeOffset := byteOffsetToRuneOffset(line, i)
		operands := p.parseOperands(line[i:], lineNum, runeOffset, stmtNode)
		stmtNode.Children = append(stmtNode.Children, operands...)
	}

	// Check for terminator on the statement line
	trimmed := strings.TrimSpace(line)
	if hasTerminatorOutsideParens(trimmed) {
		stmtNode.HasTerminator = true
	}

	return stmtNode
}

// parseParameter parses a parameter value inside parentheses
func (p *Parser) parseParameter(line string, lineNum int, startIdx int, parent *Node) *Node {
	if startIdx >= len(line) || line[startIdx] != '(' {
		return nil
	}

	// Find matching closing parenthesis
	i := startIdx + 1
	parenCount := 1
	paramStart := i

	for i < len(line) && parenCount > 0 {
		if line[i] == '(' {
			parenCount++
		} else if line[i] == ')' {
			parenCount--
		}
		if parenCount > 0 {
			i++
		}
	}

	if i <= paramStart {
		return nil
	}

	paramValue := line[paramStart:i]

	// Convert byte offsets to rune offsets for LSP compatibility
	runeStart := byteOffsetToRuneOffset(line, paramStart)
	runeLength := runeCount(paramValue)

	paramNode := &Node{
		Type:   NodeTypeParameter,
		Value:  paramValue,
		Parent: parent,
		Position: Position{
			Line:      lineNum,
			Character: runeStart,
			Length:    runeLength,
		},
	}

	return paramNode
}

// parseOperands parses operands from the text
func (p *Parser) parseOperands(text string, lineNum int, offset int, parent *Node) []*Node {
	operands := []*Node{}
	i := 0

	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}

		if i >= len(text) {
			break
		}

		// Check for operand (uppercase word or lowercase if we're context-aware)
		if isOperandStartChar(text[i]) {
			start := i
			for i < len(text) && isOperandChar(text[i]) {
				i++
			}

			operandName := text[start:i]

			// Create operand node
			// Convert byte offsets to rune offsets for LSP compatibility
			// Note: offset is already in bytes, text[:start] gives us the prefix
			runeOffset := byteOffsetToRuneOffset(text, start)
			runeLength := runeCount(operandName)
			// Also convert offset from the parent context
			parentRuneOffset := offset // This may need adjustment if offset is also byte-based

			operandNode := &Node{
				Type:   NodeTypeOperand,
				Name:   operandName,
				Parent: parent,
				Position: Position{
					Line:      lineNum,
					Character: parentRuneOffset + runeOffset,
					Length:    runeLength,
				},
			}

			// Link to operand definition
			if parent.Type == NodeTypeStatement && parent.StatementDef != nil {
				// Top-level operand: link from statement's operands
				if stmtOps, exists := p.operands[parent.Name]; exists {
					if opDef, exists := stmtOps[operandName]; exists {
						operandNode.OperandDef = &opDef
						logger.Debug("Parser: Operand %s linked to definition", operandName)
					}
				}
			} else if parent.Type == NodeTypeOperand && parent.OperandDef != nil {
				// Sub-operand: link from parent operand's values
				// Support pipe-separated aliases (e.g., "AMODE|AMOD")
				for _, val := range parent.OperandDef.Values {
					names := strings.Split(val.Name, "|")
					for _, name := range names {
						if strings.TrimSpace(name) == operandName {
							// Convert AllowedValue to Operand for linking
							operandNode.OperandDef = &data.Operand{
								Name:        val.Name,
								Parameter:   val.Parameter,
								Type:        val.Type,
								Description: val.Description,
								Length:      val.Length,
							}
							logger.Debug("Parser: Sub-operand %s linked to definition", operandName)
							break
						}
					}
				}
			}

			// Skip whitespace before parenthesis
			for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
				i++
			}

			// Check if followed by parenthesis (parameter or sub-operands)
			if i < len(text) && text[i] == '(' {
				paramNode := p.parseOperandParameter(text, lineNum, offset, i, operandNode)
				if paramNode != nil {
					// Check if this operand has sub-operands by looking at the parameter definition
					// Sub-operands have parentheses in the parameter field (e.g., "DSN(dsname) VOL(volser)")
					// Simple parameters don't have parentheses (e.g., "REASON_ID")
					hasSubOperands := false
					if operandNode.OperandDef != nil && len(operandNode.OperandDef.Values) > 0 {
						// Check if the parameter contains parentheses indicating sub-operands
						hasSubOperands = strings.Contains(operandNode.OperandDef.Parameter, "(")
					}

					if hasSubOperands {
						// Parse sub-operands recursively
						subOperands := p.parseOperands(paramNode.Value, lineNum, paramNode.Position.Character, operandNode)
						operandNode.Children = subOperands
					} else {
						// Simple parameter
						operandNode.Children = []*Node{paramNode}
					}

					i = paramNode.Position.Character + paramNode.Position.Length + 1 - offset // +1 for closing ')'
				}
			}

			operands = append(operands, operandNode)
			continue
		}

		// Skip other characters
		i++
	}

	return operands
}

// parseOperandParameter parses a parameter for an operand
// Returns a wrapper node containing individual parameter nodes for comma-separated values
func (p *Parser) parseOperandParameter(text string, lineNum int, offset int, startIdx int, parent *Node) *Node {
	if startIdx >= len(text) || text[startIdx] != '(' {
		return nil
	}

	// Find matching closing parenthesis
	i := startIdx + 1
	parenCount := 1
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

	if i <= paramStart {
		return nil
	}

	paramValue := text[paramStart:i]

	// Convert byte offsets to rune offsets for LSP compatibility
	runeParamStart := byteOffsetToRuneOffset(text, paramStart)
	runeParamLength := runeCount(paramValue)

	// Create a wrapper node for the entire parameter section
	wrapperNode := &Node{
		Type:   NodeTypeParameter,
		Value:  paramValue,
		Parent: parent,
		Position: Position{
			Line:      lineNum,
			Character: offset + runeParamStart,
			Length:    runeParamLength,
		},
		Children: []*Node{},
	}

	// Split comma-separated parameters and create individual nodes
	// This enables proper highlighting for multiline parameter lists
	params := p.splitParameters(paramValue)
	currentPos := paramStart

	for _, param := range params {
		// Find where this parameter actually starts in the text
		// (accounting for whitespace and commas)
		paramIdx := strings.Index(text[currentPos:i], param)
		if paramIdx == -1 {
			continue
		}
		actualStart := currentPos + paramIdx

		// Convert byte offsets to rune offsets
		runeActualStart := byteOffsetToRuneOffset(text, actualStart)
		runeParamLen := runeCount(param)

		paramNode := &Node{
			Type:   NodeTypeParameter,
			Value:  param,
			Parent: wrapperNode,
			Position: Position{
				Line:      lineNum,
				Character: offset + runeActualStart,
				Length:    runeParamLen,
			},
		}
		wrapperNode.Children = append(wrapperNode.Children, paramNode)

		// Move currentPos to after this parameter (in bytes for string indexing)
		currentPos = actualStart + len(param)
	}

	return wrapperNode
}

// splitParameters splits a parameter string into individual trimmed parameters
// According to SMP/E reference, parameters can be separated by commas OR one or more blanks
func (p *Parser) splitParameters(paramValue string) []string {
	var params []string
	var current strings.Builder
	parenDepth := 0

	for i := 0; i < len(paramValue); i++ {
		ch := paramValue[i]

		if ch == '(' {
			parenDepth++
			current.WriteByte(ch)
		} else if ch == ')' {
			parenDepth--
			current.WriteByte(ch)
		} else if (ch == ',' || ch == ' ' || ch == '\t') && parenDepth == 0 {
			// Found a separator (comma or whitespace) at the top level - split here
			param := strings.TrimSpace(current.String())
			if param != "" {
				params = append(params, param)
			}
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	// Add the last parameter
	param := strings.TrimSpace(current.String())
	if param != "" {
		params = append(params, param)
	}

	return params
}

// isOperandChar checks if a character is valid in an operand name
func isOperandChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

// isOperandStartChar checks if a character can start an operand
func isOperandStartChar(ch byte) bool {
	return ch >= 'A' && ch <= 'Z'
}

// hasTerminatorOutsideParens checks if a '.' exists outside of parentheses
// This correctly handles dataset names like DSN(MY.DATA.SET) where dots appear inside parentheses
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
