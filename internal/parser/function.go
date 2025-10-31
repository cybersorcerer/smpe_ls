package parser

import (
	"github.com/cybersorcerer/smpe_ls/internal/ast"
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
)

// parseFunctionStatement parses a ++FUNCTION statement
// Syntax: ++FUNCTION(sysmod_id) [DESCRIPTION(description)]
//         [FESN(fe_service_number)]
//         [FILES(number) [RFDSNPFX(relfile_prefix)]]
//         [REWORK(level)]
func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}

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

	// Parse optional operands (can be on same or following lines)
	for {
		p.nextToken()

		if p.curTokenIs(lexer.EOF) {
			break
		}

		// Skip newlines between operands
		if p.curTokenIs(lexer.NEWLINE) {
			continue
		}

		// Check if we've reached another statement
		if p.curTokenIs(lexer.MCS_APAR) ||
			p.curTokenIs(lexer.MCS_ASSIGN) ||
			p.curTokenIs(lexer.MCS_DELETE) ||
			p.curTokenIs(lexer.MCS_FEATURE) ||
			p.curTokenIs(lexer.MCS_FUNCTION) ||
			p.curTokenIs(lexer.MCS_HOLD) {
			// Back up so the main parser can handle this statement
			return stmt
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

		case lexer.FESN:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) && !p.expectPeek(lexer.NUMBER) {
				return stmt
			}
			stmt.Fesn = p.curToken.Literal
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

			// Optional RFDSNPFX after FILES
			if p.peekTokenIs(lexer.RFDSNPFX) {
				p.nextToken()
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
			// REWORK is typically the last operand
			return stmt

		default:
			// Unknown token, might be end of statement
			return stmt
		}
	}

	return stmt
}
