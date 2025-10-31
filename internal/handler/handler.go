package handler

import (
	"sync"

	"github.com/cybersorcerer/smpe_ls/internal/completion"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/internal/hover"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Handler implements the LSP handler interface
type Handler struct {
	version            string
	documents          map[string]string
	documentsMutex     sync.RWMutex
	completionProvider *completion.Provider
	hoverProvider      *hover.Provider
	diagnosticsProvider *diagnostics.Provider
	server             *lsp.Server
}

// New creates a new handler
func New(version string, dataPath string) (*Handler, error) {
	hoverProvider, err := hover.NewProvider(dataPath)
	if err != nil {
		return nil, err
	}

	return &Handler{
		version:             version,
		documents:           make(map[string]string),
		completionProvider:  completion.NewProvider(),
		hoverProvider:       hoverProvider,
		diagnosticsProvider: diagnostics.NewProvider(),
	}, nil
}

// SetServer sets the LSP server (for sending notifications)
func (h *Handler) SetServer(server *lsp.Server) {
	h.server = server
}

// Initialize handles the initialize request
func (h *Handler) Initialize(params lsp.InitializeParams) (*lsp.InitializeResult, error) {
	logger.Info("Initializing LSP server")

	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncFull,
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{"+", "("},
			},
			HoverProvider: true,
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
	h.documentsMutex.Unlock()

	// Send diagnostics
	h.publishDiagnostics(params.TextDocument.URI, params.TextDocument.Text)

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
	h.documentsMutex.Unlock()

	// Send diagnostics
	h.publishDiagnostics(params.TextDocument.URI, text)

	return nil
}

// TextDocumentDidClose handles document close notification
func (h *Handler) TextDocumentDidClose(params lsp.DidCloseTextDocumentParams) error {
	logger.Info("Document closed: %s", params.TextDocument.URI)

	h.documentsMutex.Lock()
	delete(h.documents, params.TextDocument.URI)
	h.documentsMutex.Unlock()

	return nil
}

// TextDocumentCompletion handles completion request
func (h *Handler) TextDocumentCompletion(params lsp.CompletionParams) ([]lsp.CompletionItem, error) {
	logger.Debug("Completion requested at %s:%d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	h.documentsMutex.RLock()
	text, ok := h.documents[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !ok {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	items := h.completionProvider.GetCompletions(text, params.Position.Line, params.Position.Character)
	logger.Debug("Returning %d completion items", len(items))

	return items, nil
}

// TextDocumentHover handles hover request
func (h *Handler) TextDocumentHover(params lsp.HoverParams) (*lsp.Hover, error) {
	logger.Debug("Hover requested at %s:%d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	h.documentsMutex.RLock()
	text, ok := h.documents[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !ok {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	hover := h.hoverProvider.GetHover(text, params.Position.Line, params.Position.Character)

	return hover, nil
}

// publishDiagnostics publishes diagnostics for a document
func (h *Handler) publishDiagnostics(uri string, text string) {
	if h.server == nil {
		return
	}

	diagnostics := h.diagnosticsProvider.Analyze(text)

	params := map[string]interface{}{
		"uri":         uri,
		"diagnostics": diagnostics,
	}

	if err := h.server.SendNotification("textDocument/publishDiagnostics", params); err != nil {
		logger.Error("Failed to publish diagnostics: %v", err)
	}
}
