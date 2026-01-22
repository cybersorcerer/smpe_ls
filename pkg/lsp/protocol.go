package lsp

// LSP Protocol types and structures
// Based on Language Server Protocol Specification

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a text document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic represents a diagnostic (error, warning, etc.)
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

// DiagnosticSeverity levels
const (
	SeverityError       = 1
	SeverityWarning     = 2
	SeverityInformation = 3
	SeverityHint        = 4
)

// TextDocumentIdentifier identifies a text document
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier identifies a versioned text document
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

// TextDocumentItem represents a text document
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// TextDocumentContentChangeEvent describes a change to a text document
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// TextEdit represents a text edit
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label            string    `json:"label"`
	Kind             int       `json:"kind,omitempty"`
	Detail           string    `json:"detail,omitempty"`
	Documentation    string    `json:"documentation,omitempty"`
	InsertText       string    `json:"insertText,omitempty"`
	InsertTextFormat int       `json:"insertTextFormat,omitempty"`
	TextEdit         *TextEdit `json:"textEdit,omitempty"`
}

// CompletionItemKind values
const (
	CompletionItemKindText        = 1
	CompletionItemKindMethod      = 2
	CompletionItemKindFunction    = 3
	CompletionItemKindConstructor = 4
	CompletionItemKindField       = 5
	CompletionItemKindVariable    = 6
	CompletionItemKindClass       = 7
	CompletionItemKindInterface   = 8
	CompletionItemKindModule      = 9
	CompletionItemKindProperty    = 10
	CompletionItemKindUnit        = 11
	CompletionItemKindValue       = 12
	CompletionItemKindEnum        = 13
	CompletionItemKindKeyword     = 14
	CompletionItemKindSnippet     = 15
)

// InsertTextFormat values
const (
	InsertTextFormatPlainText = 1
	InsertTextFormatSnippet   = 2
)

// Hover represents hover information
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents marked up content
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// MarkupKind values
const (
	MarkupKindPlainText = "plaintext"
	MarkupKindMarkdown  = "markdown"
)

// ServerCapabilities describes the capabilities of the server
type ServerCapabilities struct {
	TextDocumentSync                int                    `json:"textDocumentSync,omitempty"`
	CompletionProvider              *CompletionOptions     `json:"completionProvider,omitempty"`
	HoverProvider                   bool                   `json:"hoverProvider,omitempty"`
	DiagnosticProvider              *DiagnosticOptions     `json:"diagnosticProvider,omitempty"`
	SemanticTokensProvider          *SemanticTokensOptions `json:"semanticTokensProvider,omitempty"`
	DocumentFormattingProvider      bool                   `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider bool                   `json:"documentRangeFormattingProvider,omitempty"`
	DocumentSymbolProvider          bool                   `json:"documentSymbolProvider,omitempty"`
	DefinitionProvider              bool                   `json:"definitionProvider,omitempty"`
	ReferencesProvider              bool                   `json:"referencesProvider,omitempty"`
}

// TextDocumentSyncKind values
const (
	TextDocumentSyncNone        = 0
	TextDocumentSyncFull        = 1
	TextDocumentSyncIncremental = 2
)

// CompletionOptions describes completion options
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// DiagnosticOptions describes diagnostic options
type DiagnosticOptions struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

// SemanticTokensOptions describes semantic tokens options
type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
	Range  bool                 `json:"range,omitempty"`
}

// SemanticTokensLegend describes the legend for semantic tokens
type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// SemanticToken represents a single semantic token (delta-encoded)
type SemanticToken struct {
	DeltaLine      int `json:"deltaLine"`
	DeltaStart     int `json:"deltaStart"`
	Length         int `json:"length"`
	TokenType      int `json:"tokenType"`
	TokenModifiers int `json:"tokenModifiers"`
}

// SemanticTokensParams represents parameters for semantic tokens request
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokens represents the result of a semantic tokens request
type SemanticTokens struct {
	Data []int `json:"data"`
}

// InitializationOptions represents client-provided initialization options
type InitializationOptions struct {
	Diagnostics *DiagnosticsOptions `json:"diagnostics,omitempty"`
	Formatting  *FormattingOptions  `json:"formatting,omitempty"`
}

// DiagnosticsOptions configures which diagnostics are enabled
type DiagnosticsOptions struct {
	UnknownStatement       bool `json:"unknownStatement"`
	InvalidLanguageId      bool `json:"invalidLanguageId"`
	UnbalancedParentheses  bool `json:"unbalancedParentheses"`
	MissingTerminator      bool `json:"missingTerminator"`
	MissingParameter       bool `json:"missingParameter"`
	UnknownOperand         bool `json:"unknownOperand"`
	DuplicateOperand       bool `json:"duplicateOperand"`
	EmptyOperandParameter  bool `json:"emptyOperandParameter"`
	MissingRequiredOperand bool `json:"missingRequiredOperand"`
	DependencyViolation    bool `json:"dependencyViolation"`
	MutuallyExclusive      bool `json:"mutuallyExclusive"`
	RequiredGroup          bool `json:"requiredGroup"`
	MissingInlineData      bool `json:"missingInlineData"`
	UnknownSubOperand      bool `json:"unknownSubOperand"`
	SubOperandValidation   bool `json:"subOperandValidation"`
}

// InitializeParams represents the initialize request parameters
type InitializeParams struct {
	ProcessID             int                    `json:"processId"`
	RootURI               string                 `json:"rootUri,omitempty"`
	Capabilities          struct{}               `json:"capabilities"`
	InitializationOptions *InitializationOptions `json:"initializationOptions,omitempty"`
}

// InitializeResult represents the initialize response
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerInfo contains server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// DidChangeConfigurationParams represents workspace/didChangeConfiguration params
type DidChangeConfigurationParams struct {
	Settings *SettingsPayload `json:"settings"`
}

// SettingsPayload represents the settings sent from the client
type SettingsPayload struct {
	Smpe *SmpeSettings `json:"smpe,omitempty"`
}

// SmpeSettings represents the smpe.* settings from VSCode
type SmpeSettings struct {
	Diagnostics *DiagnosticsOptions `json:"diagnostics,omitempty"`
	Formatting  *FormattingOptions  `json:"formatting,omitempty"`
}

// FormattingOptions configures document formatting behavior
type FormattingOptions struct {
	Enabled            bool `json:"enabled"`
	IndentContinuation int  `json:"indentContinuation"`
	OneOperandPerLine  bool `json:"oneOperandPerLine"`
}

// DocumentFormattingParams represents textDocument/formatting request params
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingRequestOptions `json:"options"`
}

// FormattingRequestOptions contains formatting options from the client
type FormattingRequestOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// DocumentRangeFormattingParams represents textDocument/rangeFormatting request params
type DocumentRangeFormattingParams struct {
	TextDocument TextDocumentIdentifier   `json:"textDocument"`
	Range        Range                    `json:"range"`
	Options      FormattingRequestOptions `json:"options"`
}

// DocumentSymbolParams represents textDocument/documentSymbol request params
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DefinitionParams represents textDocument/definition request params
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// ReferenceParams represents textDocument/references request params
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

// ReferenceContext contains additional context for reference requests
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// DocumentSymbol represents a symbol in a document (hierarchical)
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolKind represents the kind of a symbol
type SymbolKind int

// SymbolKind values
const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)
