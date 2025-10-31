package parser

import (
	"github.com/cybersorcerer/smpe_ls/internal/ast"
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
)

// parseAssignStatement parses a ++ASSIGN statement
// Syntax: ++ASSIGN SOURCEID(source_id) TO(sysmod_id [,sysmod_id]...)
func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{
		Token:     p.curToken,
		SysmodIDs: []string{},
	}

	// Expect SOURCEID keyword
	if !p.expectPeek(lexer.SOURCEID) {
		return nil
	}

	// Expect LPAREN
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Expect source_id
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.SourceID = p.curToken.Literal

	// Expect RPAREN
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Expect TO keyword
	if !p.expectPeek(lexer.TO) {
		return nil
	}

	// Expect LPAREN
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Parse sysmod IDs (can be multiple, comma-separated)
	for {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.SysmodIDs = append(stmt.SysmodIDs, p.curToken.Literal)

		// Check for comma (more IDs to come) or closing paren
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
			continue
		} else if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken() // consume closing paren
			break
		} else {
			p.peekError(lexer.RPAREN)
			return nil
		}
	}

	return stmt
}
