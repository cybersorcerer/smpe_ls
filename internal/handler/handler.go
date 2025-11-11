package handler

import (
	"sync"

	"github.com/cybersorcerer/smpe_ls/internal/completion"
	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/internal/hover"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/internal/semantic"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Handler implements the LSP handler interface
type Handler struct {
	version             string
	documents           map[string]string
	parsedDocuments     map[string]*parser.Document // AST cache
	documentsMutex      sync.RWMutex
	parser              *parser.Parser
	completionProvider  *completion.Provider
	hoverProvider       *hover.Provider
	diagnosticsProvider *diagnostics.Provider
	semanticProvider    *semantic.Provider
	server              *lsp.Server
}

// New creates a new handler
func New(version string, dataPath string) (*Handler, error) {
	// Load MCS data once and share it among all providers
	logger.Info("Loading MCS data from %s", dataPath)
	store, err := data.Load(dataPath)
	if err != nil {
		return nil, err
	}
	logger.Info("Loaded %d MCS statements", len(store.List))

	// Create parser
	parserInstance := parser.NewParser(store.Statements)

	// Create providers with shared data
	hoverProvider := hover.NewProvider(store)
	completionProvider := completion.NewProvider(store)
	diagnosticsProvider := diagnostics.NewProvider(store)
	semanticProvider := semantic.NewProvider(store.Statements)

	return &Handler{
		version:             version,
		documents:           make(map[string]string),
		parsedDocuments:     make(map[string]*parser.Document),
		parser:              parserInstance,
		completionProvider:  completionProvider,
		hoverProvider:       hoverProvider,
		diagnosticsProvider: diagnosticsProvider,
		semanticProvider:    semanticProvider,
	}, nil
}

// SetServer sets the LSP server (for sending notifications)
func (h *Handler) SetServer(server *lsp.Server) {
	h.server = server
}

// Initialize handles the initialize request
func (h *Handler) Initialize(params lsp.InitializeParams) (*lsp.InitializeResult, error) {
	logger.Info("Initializing LSP server")

	// Add all uppercase letters as trigger characters so completion triggers automatically when typing operand names
	triggerChars := []string{"+", "(", " "}
	for ch := 'A'; ch <= 'Z'; ch++ {
		triggerChars = append(triggerChars, string(ch))
	}

	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncFull,
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: triggerChars,
			},
			HoverProvider: true,
			SemanticTokensProvider: &lsp.SemanticTokensOptions{
				Legend: lsp.SemanticTokensLegend{
					TokenTypes: []string{
						"keyword",   // MCS statements
						"function",  // Operands
						"parameter", // Parameter values
						"comment",   // Comments
						"string",    // Quoted strings
						"number",    // Numbers
					},
					TokenModifiers: []string{},
				},
				Full:  true,
				Range: false,
			},
		},
		ServerInfo: &lsp.ServerInfo{
			Name:    "smpe_ls",
			Version: h.version,
		},
	}, nil
}

// TextDocumentDidOpen handles document open notification
func (h *Handler) TextDocumentDidOpen(params lsp.DidOpenTextDocumentParams) error {
	logger.Info("Document opened: %s", params.TextDocument.URI)

	h.documentsMutex.Lock()
	h.documents[params.TextDocument.URI] = params.TextDocument.Text

	// Parse document and cache AST
	doc := h.parser.Parse(params.TextDocument.Text)
	h.parsedDocuments[params.TextDocument.URI] = doc
	h.documentsMutex.Unlock()

	// Send diagnostics
	h.publishDiagnostics(params.TextDocument.URI)

	return nil
}

// TextDocumentDidChange handles document change notification
func (h *Handler) TextDocumentDidChange(params lsp.DidChangeTextDocumentParams) error {
	logger.Debug("Document changed: %s", params.TextDocument.URI)

	h.documentsMutex.Lock()
	// Full sync mode - replace entire document
	if len(params.ContentChanges) > 0 {
		h.documents[params.TextDocument.URI] = params.ContentChanges[0].Text
	}
	text := h.documents[params.TextDocument.URI]

	// Re-parse document and update cache
	doc := h.parser.Parse(text)
	h.parsedDocuments[params.TextDocument.URI] = doc
	h.documentsMutex.Unlock()

	// Send diagnostics
	h.publishDiagnostics(params.TextDocument.URI)

	return nil
}

// TextDocumentDidClose handles document close notification
func (h *Handler) TextDocumentDidClose(params lsp.DidCloseTextDocumentParams) error {
	logger.Info("Document closed: %s", params.TextDocument.URI)

	h.documentsMutex.Lock()
	delete(h.documents, params.TextDocument.URI)
	delete(h.parsedDocuments, params.TextDocument.URI) // Also clear cached AST
	h.documentsMutex.Unlock()

	return nil
}

// TextDocumentCompletion handles completion request
func (h *Handler) TextDocumentCompletion(params lsp.CompletionParams) ([]lsp.CompletionItem, error) {
	logger.Debug("Completion requested at %s:%d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	h.documentsMutex.RLock()
	text, ok := h.documents[params.TextDocument.URI]
	doc, hasDoc := h.parsedDocuments[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !ok {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	// Always ensure we have a parsed document (AST)
	if !hasDoc {
		logger.Debug("No parsed document found, parsing now: %s", params.TextDocument.URI)
		h.documentsMutex.Lock()
		doc = h.parser.Parse(text)
		h.parsedDocuments[params.TextDocument.URI] = doc
		h.documentsMutex.Unlock()
	}

	// Always use AST-based completion
	items := h.completionProvider.GetCompletionsAST(doc, text, params.Position.Line, params.Position.Character)
	logger.Debug("Using AST-based completion, returning %d items", len(items))

	return items, nil
}

// TextDocumentHover handles hover request
func (h *Handler) TextDocumentHover(params lsp.HoverParams) (*lsp.Hover, error) {
	logger.Debug("Hover requested at %s:%d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	h.documentsMutex.RLock()
	text, textExists := h.documents[params.TextDocument.URI]
	doc, hasDoc := h.parsedDocuments[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !textExists {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	// Always ensure we have a parsed document (AST)
	if !hasDoc {
		logger.Debug("No parsed document found for hover, parsing now: %s", params.TextDocument.URI)
		h.documentsMutex.Lock()
		doc = h.parser.Parse(text)
		h.parsedDocuments[params.TextDocument.URI] = doc
		h.documentsMutex.Unlock()
	}

	// Always use AST-based hover
	hover := h.hoverProvider.GetHoverAST(doc, params.Position.Line, params.Position.Character)
	logger.Debug("Using AST-based hover")

	return hover, nil
}

// TextDocumentSemanticTokensFull handles semantic tokens request
func (h *Handler) TextDocumentSemanticTokensFull(params lsp.SemanticTokensParams) (*lsp.SemanticTokens, error) {
	logger.Debug("Semantic tokens request for: %s", params.TextDocument.URI)

	h.documentsMutex.RLock()
	doc, exists := h.parsedDocuments[params.TextDocument.URI]
	text, textExists := h.documents[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !exists {
		logger.Debug("Document not found for semantic tokens: %s", params.TextDocument.URI)
		return &lsp.SemanticTokens{Data: []int{}}, nil
	}

	if !textExists {
		logger.Debug("Document text not found for semantic tokens: %s", params.TextDocument.URI)
		return &lsp.SemanticTokens{Data: []int{}}, nil
	}

	// Generate semantic tokens from AST
	data := h.semanticProvider.BuildTokensFromAST(doc, text)

	return &lsp.SemanticTokens{
		Data: data,
	}, nil
}

// publishDiagnostics publishes diagnostics for a document
func (h *Handler) publishDiagnostics(uri string) {
	if h.server == nil {
		return
	}

	// Get parsed document from cache
	h.documentsMutex.RLock()
	doc, exists := h.parsedDocuments[uri]
	h.documentsMutex.RUnlock()

	if !exists {
		logger.Debug("No parsed document found for diagnostics: %s", uri)
		return
	}

	// Generate diagnostics from AST
	diagnostics := h.diagnosticsProvider.AnalyzeAST(doc)

	params := map[string]interface{}{
		"uri":         uri,
		"diagnostics": diagnostics,
	}

	if err := h.server.SendNotification("textDocument/publishDiagnostics", params); err != nil {
		logger.Error("Failed to publish diagnostics: %v", err)
	}
}
