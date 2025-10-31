package parser

import (
	"fmt"

	"github.com/cybersorcerer/smpe_ls/internal/ast"
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
)

// Parser represents the parser
type Parser struct {
	lexer     *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
}

// New creates a new Parser instance
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// curTokenIs checks if current token is of given type
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if peek token is of given type
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if peek token is of given type and advances if true
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError adds an error for unexpected peek token
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("Line %d, Col %d: expected next token to be %s, got %s instead",
		p.peekToken.Line, p.peekToken.Column, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
	logger.Debug("Parser error: %s", msg)
}

// addError adds a custom error message
func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.errors = append(p.errors, msg)
	logger.Debug("Parser error: %s", msg)
}

// skipToNextStatement skips tokens until the next statement or EOF
func (p *Parser) skipToNextStatement() {
	for !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.MCS_APAR) ||
			p.curTokenIs(lexer.MCS_ASSIGN) ||
			p.curTokenIs(lexer.MCS_DELETE) ||
			p.curTokenIs(lexer.MCS_FEATURE) ||
			p.curTokenIs(lexer.MCS_FUNCTION) ||
			p.curTokenIs(lexer.MCS_HOLD) {
			return
		}
		p.nextToken()
	}
}

// Parse parses the input and returns an AST
func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EOF) {
		// Skip comments and newlines at top level
		if p.curTokenIs(lexer.COMMENT) {
			stmt := &ast.CommentStatement{
				Token: p.curToken,
				Text:  p.curToken.Literal,
			}
			program.Statements = append(program.Statements, stmt)
			p.nextToken()
			continue
		}

		if p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	return program
}

// parseStatement parses a statement based on current token
func (p *Parser) parseStatement() ast.Statement {
	logger.Debug("Parsing statement: %s", p.curToken.Type)

	switch p.curToken.Type {
	case lexer.MCS_APAR:
		return p.parseAparStatement()
	case lexer.MCS_ASSIGN:
		return p.parseAssignStatement()
	case lexer.MCS_DELETE:
		return p.parseDeleteStatement()
	case lexer.MCS_FEATURE:
		return p.parseFeatureStatement()
	case lexer.MCS_FUNCTION:
		return p.parseFunctionStatement()
	case lexer.MCS_HOLD:
		return p.parseHoldStatement()
	default:
		p.addError("Line %d, Col %d: unknown statement type: %s",
			p.curToken.Line, p.curToken.Column, p.curToken.Literal)
		p.skipToNextStatement()
		return nil
	}
}

// consumeOptionalNewlines consumes zero or more newlines
func (p *Parser) consumeOptionalNewlines() {
	for p.curTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}
}
