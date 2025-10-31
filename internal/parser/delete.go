package parser

import (
	"github.com/cybersorcerer/smpe_ls/internal/ast"
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
)

// parseDeleteStatement parses a ++DELETE statement
// Syntax: ++DELETE(name) SYSLIB(ALL | ALIAS(alias)... | SYSLIB(ddname)...)
func (p *Parser) parseDeleteStatement() *ast.DeleteStatement {
	stmt := &ast.DeleteStatement{
		Token:   p.curToken,
		Aliases: []string{},
		Ddnames: []string{},
	}

	// Expect LPAREN
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Expect name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = p.curToken.Literal

	// Expect RPAREN
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Expect SYSLIB keyword
	if !p.expectPeek(lexer.SYSLIB) {
		return nil
	}

	// Expect LPAREN
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Check for ALL or specific aliases/ddnames
	p.nextToken()
	if p.curTokenIs(lexer.IDENT) && p.curToken.Literal == "ALL" {
		stmt.Syslib = "ALL"
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
		return stmt
	}

	// Parse ALIAS or SYSLIB operands
	for {
		if p.curTokenIs(lexer.ALIAS) {
			if !p.expectPeek(lexer.LPAREN) {
				return nil
			}
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			stmt.Aliases = append(stmt.Aliases, p.curToken.Literal)
			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}
		} else if p.curTokenIs(lexer.SYSLIB) {
			if !p.expectPeek(lexer.LPAREN) {
				return nil
			}
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			stmt.Ddnames = append(stmt.Ddnames, p.curToken.Literal)
			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}
		}

		// Check for more operands or end
		if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken() // consume closing paren
			break
		} else if p.peekTokenIs(lexer.ALIAS) || p.peekTokenIs(lexer.SYSLIB) {
			p.nextToken() // continue with next operand
		} else {
			break
		}
	}

	return stmt
}
