package lexer

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT   // identifiers (names, values)
	STRING  // quoted strings
	NUMBER  // numbers

	// MCS Statements
	MCS_APAR
	MCS_ASSIGN
	MCS_DELETE
	MCS_FEATURE
	MCS_FUNCTION
	MCS_HOLD

	// Common keywords/operands (alphabetically)
	ALIAS
	CATEGORY
	CLASS
	COMMENT_KW // COMMENT keyword
	DATE
	DELETE_KW  // DELETE keyword
	DESCRIPTION
	DISTLIB
	ERROR
	FESN
	FILES
	FIXCAT
	FMID
	FROMDS
	NUMBER_KW  // NUMBER keyword
	PRODUCT
	REASON
	RELFILE
	RESOLVER
	REWORK
	RFDSNPFX
	RMID
	SOURCEID
	SYSLIB
	SYSTEM
	TO
	TXLIB
	UNIT
	USER
	VERSION
	VOL

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	COMMA     // ,
	PERIOD    // .
	EQUAL     // =
	NEWLINE   // \n
)

var tokenNames = map[TokenType]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",
	IDENT:   "IDENT",
	STRING:  "STRING",
	NUMBER:  "NUMBER",

	MCS_APAR:     "++APAR",
	MCS_ASSIGN:   "++ASSIGN",
	MCS_DELETE:   "++DELETE",
	MCS_FEATURE:  "++FEATURE",
	MCS_FUNCTION: "++FUNCTION",
	MCS_HOLD:     "++HOLD",

	ALIAS:       "ALIAS",
	CATEGORY:    "CATEGORY",
	CLASS:       "CLASS",
	COMMENT_KW:  "COMMENT",
	DATE:        "DATE",
	DELETE_KW:   "DELETE",
	DESCRIPTION: "DESCRIPTION",
	DISTLIB:     "DISTLIB",
	ERROR:       "ERROR",
	FESN:        "FESN",
	FILES:       "FILES",
	FIXCAT:      "FIXCAT",
	FMID:        "FMID",
	FROMDS:      "FROMDS",
	NUMBER_KW:   "NUMBER",
	PRODUCT:     "PRODUCT",
	REASON:      "REASON",
	RELFILE:     "RELFILE",
	RESOLVER:    "RESOLVER",
	REWORK:      "REWORK",
	RFDSNPFX:    "RFDSNPFX",
	RMID:        "RMID",
	SOURCEID:    "SOURCEID",
	SYSLIB:      "SYSLIB",
	SYSTEM:      "SYSTEM",
	TO:          "TO",
	TXLIB:       "TXLIB",
	UNIT:        "UNIT",
	USER:        "USER",
	VERSION:     "VERSION",
	VOL:         "VOL",

	LPAREN:  "(",
	RPAREN:  ")",
	COMMA:   ",",
	PERIOD:  ".",
	EQUAL:   "=",
	NEWLINE: "NEWLINE",
}

// String returns the string representation of a TokenType
func (tt TokenType) String() string {
	if name, ok := tokenNames[tt]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", tt)
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// String returns a string representation of the token
func (t Token) String() string {
	return fmt.Sprintf("Token{Type:%s, Literal:%q, Line:%d, Col:%d}",
		t.Type, t.Literal, t.Line, t.Column)
}

// Keywords maps keywords to their token types
var keywords = map[string]TokenType{
	"++APAR":     MCS_APAR,
	"++ASSIGN":   MCS_ASSIGN,
	"++DELETE":   MCS_DELETE,
	"++FEATURE":  MCS_FEATURE,
	"++FUNCTION": MCS_FUNCTION,
	"++HOLD":     MCS_HOLD,

	"ALIAS":       ALIAS,
	"CATEGORY":    CATEGORY,
	"CLASS":       CLASS,
	"COMMENT":     COMMENT_KW,
	"DATE":        DATE,
	"DELETE":      DELETE_KW,
	"DESCRIPTION": DESCRIPTION,
	"DISTLIB":     DISTLIB,
	"ERROR":       ERROR,
	"FESN":        FESN,
	"FILES":       FILES,
	"FIXCAT":      FIXCAT,
	"FMID":        FMID,
	"FROMDS":      FROMDS,
	"NUMBER":      NUMBER_KW,
	"PRODUCT":     PRODUCT,
	"REASON":      REASON,
	"RELFILE":     RELFILE,
	"RESOLVER":    RESOLVER,
	"REWORK":      REWORK,
	"RFDSNPFX":    RFDSNPFX,
	"RMID":        RMID,
	"SOURCEID":    SOURCEID,
	"SYSLIB":      SYSLIB,
	"SYSTEM":      SYSTEM,
	"TO":          TO,
	"TXLIB":       TXLIB,
	"UNIT":        UNIT,
	"USER":        USER,
	"VERSION":     VERSION,
	"VOL":         VOL,
}

// LookupKeyword checks if an identifier is a keyword
func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
