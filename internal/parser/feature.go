package parser

import (
	"github.com/cybersorcerer/smpe_ls/internal/ast"
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
)

// parseFeatureStatement parses a ++FEATURE statement
// Syntax: ++FEATURE(name) DESCRIPTION(description) [FMID(fmid)...]
//         PRODUCT(prodid) w.r.m.m [REWORK(level)]
func (p *Parser) parseFeatureStatement() *ast.FeatureStatement {
	stmt := &ast.FeatureStatement{
		Token: p.curToken,
		Fmids: []string{},
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

	// Parse operands on same line
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

		case lexer.FMID:
			if !p.expectPeek(lexer.LPAREN) {
				return stmt
			}
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			stmt.Fmids = append(stmt.Fmids, p.curToken.Literal)
			if !p.expectPeek(lexer.RPAREN) {
				return stmt
			}
		}
	}

	// Optional: Parse PRODUCT line (next line)
	p.consumeOptionalNewlines()

	if p.curTokenIs(lexer.PRODUCT) {
		if !p.expectPeek(lexer.LPAREN) {
			return stmt
		}
		if !p.expectPeek(lexer.IDENT) {
			return stmt
		}
		stmt.Product = p.curToken.Literal
		if !p.expectPeek(lexer.RPAREN) {
			return stmt
		}

		// Parse version (w.r.m.m format)
		if p.expectPeek(lexer.NUMBER) {
			version := p.curToken.Literal
			// Parse dots and remaining numbers
			for p.peekTokenIs(lexer.PERIOD) {
				p.nextToken() // consume period
				version += "."
				if p.expectPeek(lexer.NUMBER) {
					version += p.curToken.Literal
				}
			}
			stmt.Version = version
		}

		// Optional REWORK
		if p.peekTokenIs(lexer.REWORK) {
			p.nextToken()
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
