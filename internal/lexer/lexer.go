package lexer

import (
	"unicode"
)

// Lexer represents the lexical analyzer
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// New creates a new Lexer instance
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar advances the position and reads the next character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for "NUL"
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar returns the next character without advancing the position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '(':
		tok = l.newToken(LPAREN, string(l.ch))
	case ')':
		tok = l.newToken(RPAREN, string(l.ch))
	case ',':
		tok = l.newToken(COMMA, string(l.ch))
	case '.':
		tok = l.newToken(PERIOD, string(l.ch))
	case '=':
		tok = l.newToken(EQUAL, string(l.ch))
	case '\n':
		tok = l.newToken(NEWLINE, "\\n")
		l.line++
		l.column = 0
	case '*':
		// Check for comment (line starting with *)
		if l.column == 1 {
			comment := l.readComment()
			tok = Token{Type: COMMENT, Literal: comment, Line: l.line, Column: 1}
			return tok
		}
		tok = l.newToken(ILLEGAL, string(l.ch))
	case '\'':
		// Read quoted string
		str := l.readString()
		tok = Token{Type: STRING, Literal: str, Line: tok.Line, Column: tok.Column}
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if l.ch == '+' && l.peekChar() == '+' {
			// MCS Statement (++KEYWORD)
			line := tok.Line
			col := tok.Column
			literal := l.readMCSStatement()
			tok = Token{
				Type:    LookupKeyword(literal),
				Literal: literal,
				Line:    line,
				Column:  col,
			}
			return tok
		} else if isLetter(l.ch) {
			// Identifier or keyword
			line := tok.Line
			col := tok.Column
			literal := l.readIdentifier()
			tok = Token{
				Type:    LookupKeyword(literal),
				Literal: literal,
				Line:    line,
				Column:  col,
			}
			return tok
		} else if isDigit(l.ch) {
			// Number
			line := tok.Line
			col := tok.Column
			literal := l.readNumber()
			tok = Token{
				Type:    NUMBER,
				Literal: literal,
				Line:    line,
				Column:  col,
			}
			return tok
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

// newToken creates a new token
func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	}
}

// readIdentifier reads an identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readMCSStatement reads a MCS statement (++KEYWORD)
func (l *Lexer) readMCSStatement() string {
	position := l.position
	// Read ++
	l.readChar() // first +
	l.readChar() // second +
	// Read the keyword part
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readString reads a quoted string
func (l *Lexer) readString() string {
	// Skip opening quote
	l.readChar()
	position := l.position

	for l.ch != '\'' && l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}

	str := l.input[position:l.position]

	// Skip closing quote if present
	if l.ch == '\'' {
		l.readChar()
	}

	return str
}

// readComment reads a comment line (starting with *)
func (l *Lexer) readComment() string {
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

// skipWhitespace skips whitespace characters (but not newlines)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// isLetter checks if a character is a letter
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

// Tokenize returns all tokens from the input
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}
