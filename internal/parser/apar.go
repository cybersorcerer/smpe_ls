package parser

import (
	"github.com/cybersorcerer/smpe_ls/internal/ast"
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
)

// parseAparStatement parses a ++APAR statement
// Syntax: ++APAR(sysmod_id) [DESCRIPTION(description)] [FILES(number)] [RFDSNPFX(relfile_prefix)]
//         [REWORK(level)]
func (p *Parser) parseAparStatement() *ast.AparStatement {
	stmt := &ast.AparStatement{Token: p.curToken}

	// Expect LPAREN
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Expect sysmod_id
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.SysmodID = p.curToken.Literal

	// Expect RPAREN
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Parse optional operands
	for {
		p.nextToken()

		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.EOF) {
			break
		}

		switch p.curToken.Type {
		case lexer.DESCRIPTION:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Description = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.FILES:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.NUMBER) {
				return stmt
			}
			stmt.Files = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.RFDSNPFX:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Rfdsnpfx = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.REWORK:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) && !p.expectPeek(lexer.NUMBER) {
				return stmt
			}
			stmt.Rework = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}
		}
	}

	return stmt
}
