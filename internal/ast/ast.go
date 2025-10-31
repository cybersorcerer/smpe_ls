package ast

import "github.com/cybersorcerer/smpe_ls/internal/lexer"

// Node is the interface that all AST nodes implement
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node
type Statement interface {
	Node
	statementNode()
}

// Program represents the root node of the AST
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	result := ""
	for _, stmt := range p.Statements {
		result += stmt.String() + "\n"
	}
	return result
}

// Operand represents an operand with optional value
type Operand struct {
	Name  string
	Value string
	Token lexer.Token
}

func (o *Operand) String() string {
	if o.Value != "" {
		return o.Name + "(" + o.Value + ")"
	}
	return o.Name
}

// AparStatement represents a ++APAR statement
type AparStatement struct {
	Token       lexer.Token // The ++APAR token
	SysmodID    string
	Description string
	Files       string
	Rfdsnpfx    string
	Rework      string
}

func (a *AparStatement) statementNode()       {}
func (a *AparStatement) TokenLiteral() string { return a.Token.Literal }
func (a *AparStatement) String() string {
	return "++APAR(" + a.SysmodID + ")"
}

// AssignStatement represents a ++ASSIGN statement
type AssignStatement struct {
	Token     lexer.Token // The ++ASSIGN token
	SourceID  string
	SysmodIDs []string // Can have multiple target sysmod IDs
}

func (a *AssignStatement) statementNode()       {}
func (a *AssignStatement) TokenLiteral() string { return a.Token.Literal }
func (a *AssignStatement) String() string {
	return "++ASSIGN SOURCEID TO SYSMODs"
}

// DeleteStatement represents a ++DELETE statement
type DeleteStatement struct {
	Token   lexer.Token // The ++DELETE token
	Name    string
	Syslib  string      // Can be "ALL"
	Aliases []string    // Multiple aliases possible
	Ddnames []string    // Multiple ddnames possible
}

func (d *DeleteStatement) statementNode()       {}
func (d *DeleteStatement) TokenLiteral() string { return d.Token.Literal }
func (d *DeleteStatement) String() string {
	return "++DELETE(" + d.Name + ")"
}

// FeatureStatement represents a ++FEATURE statement
type FeatureStatement struct {
	Token       lexer.Token // The ++FEATURE token
	Name        string
	Description string
	Fmids       []string // Multiple FMIDs possible
	Product     string
	Version     string // w.r.m.m format
	Rework      string
}

func (f *FeatureStatement) statementNode()       {}
func (f *FeatureStatement) TokenLiteral() string { return f.Token.Literal }
func (f *FeatureStatement) String() string {
	return "++FEATURE(" + f.Name + ")"
}

// FunctionStatement represents a ++FUNCTION statement
type FunctionStatement struct {
	Token       lexer.Token // The ++FUNCTION token
	SysmodID    string
	Description string
	Fesn        string
	Files       string
	Rfdsnpfx    string
	Rework      string
}

func (f *FunctionStatement) statementNode()       {}
func (f *FunctionStatement) TokenLiteral() string { return f.Token.Literal }
func (f *FunctionStatement) String() string {
	return "++FUNCTION(" + f.SysmodID + ")"
}

// HoldStatement represents a ++HOLD statement
type HoldStatement struct {
	Token      lexer.Token // The ++HOLD token
	SysmodID   string
	Fmid       string
	Reason     string // reason_id or SYSTEM reason
	ErrorType  string // ERROR, FIXCAT, SYSTEM, USER
	Category   string
	Resolver   string
	Class      string
	Date       string
	Comments   []string // Multiple comments possible
}

func (h *HoldStatement) statementNode()       {}
func (h *HoldStatement) TokenLiteral() string { return h.Token.Literal }
func (h *HoldStatement) String() string {
	return "++HOLD(" + h.SysmodID + ")"
}

// CommentStatement represents a comment line
type CommentStatement struct {
	Token lexer.Token
	Text  string
}

func (c *CommentStatement) statementNode()       {}
func (c *CommentStatement) TokenLiteral() string { return c.Token.Literal }
func (c *CommentStatement) String() string {
	return "* " + c.Text
}
