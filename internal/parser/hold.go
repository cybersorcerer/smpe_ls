package parser

import (
	"github.com/cybersorcerer/smpe_ls/internal/ast"
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
)

// parseHoldStatement parses a ++HOLD statement
// Syntax: ++HOLD(sysmod_id) FMID(fmid) [REASON(reason_id) | REASON(SYSTEM Reason IDs)]
//         [ERROR | FIXCAT | SYSTEM | USER]
//         [CLASS(class)] [DATE(yyddd)] [COMMENT(text)]
func (p *Parser) parseHoldStatement() *ast.HoldStatement {
	stmt := &ast.HoldStatement{
		Token:    p.curToken,
		Comments: []string{},
	}

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

	// Parse operands (can span multiple lines)
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
		case lexer.FMID:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Fmid = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.REASON:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Reason = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.ERROR:
			stmt.ErrorType = "ERROR"

		case lexer.FIXCAT:
			stmt.ErrorType = "FIXCAT"
			// Optional CATEGORY
			if p.peekTokenIs(lexer.CATEGORY) {
				p.nextToken()
				if !p.expectPeek(lexer.LPAREN) {
					return stmt
				}
				if !p.expectPeek(lexer.IDENT) {
					return stmt
				}
				stmt.Category = p.curToken.Literal
				if !p.expectPeek(lexer.RPAREN) {
					return stmt
				}
			}
			// Optional RESOLVER
			if p.peekTokenIs(lexer.RESOLVER) {
				p.nextToken()
				if !p.expectPeek(lexer.LPAREN) {
					return stmt
				}
				if !p.expectPeek(lexer.IDENT) {
					return stmt
				}
				stmt.Resolver = p.curToken.Literal
				if !p.expectPeek(lexer.RPAREN) {
					return stmt
				}
			}

		case lexer.SYSTEM:
			stmt.ErrorType = "SYSTEM"

		case lexer.USER:
			stmt.ErrorType = "USER"

		case lexer.CATEGORY:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Category = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.RESOLVER:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Resolver = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.CLASS:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Class = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.DATE:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.NUMBER) {
				return stmt
			}
			stmt.Date = p.curToken.Literal
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		case lexer.COMMENT_KW:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			// Comment can be IDENT or STRING
			p.nextToken()
			comment := p.curToken.Literal
			stmt.Comments = append(stmt.Comments, comment)
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}

		default:
			// Unknown token, might be end of statement
			return stmt
		}
	}

	return stmt
}
