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

// CompletionItem represents a completion item
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
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
	TextDocumentSync   int                        `json:"textDocumentSync,omitempty"`
	CompletionProvider *CompletionOptions         `json:"completionProvider,omitempty"`
	HoverProvider      bool                       `json:"hoverProvider,omitempty"`
	DiagnosticProvider *DiagnosticOptions         `json:"diagnosticProvider,omitempty"`
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

// InitializeParams represents the initialize request parameters
type InitializeParams struct {
	ProcessID int `json:"processId"`
	RootURI   string `json:"rootUri,omitempty"`
	Capabilities struct {
		// Client capabilities
	} `json:"capabilities"`
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
