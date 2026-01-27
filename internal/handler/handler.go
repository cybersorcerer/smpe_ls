package handler

import (
	"strings"
	"sync"

	"github.com/cybersorcerer/smpe_ls/internal/completion"
	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/internal/formatting"
	"github.com/cybersorcerer/smpe_ls/internal/hover"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/internal/references"
	"github.com/cybersorcerer/smpe_ls/internal/semantic"
	"github.com/cybersorcerer/smpe_ls/internal/symbols"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// DiagnosticsConfig holds the configuration for which diagnostics to enable/disable
type DiagnosticsConfig struct {
	UnknownStatement            bool `json:"unknownStatement"`
	InvalidLanguageId           bool `json:"invalidLanguageId"`
	UnbalancedParentheses       bool `json:"unbalancedParentheses"`
	MissingTerminator           bool `json:"missingTerminator"`
	MissingParameter            bool `json:"missingParameter"`
	UnknownOperand              bool `json:"unknownOperand"`
	DuplicateOperand            bool `json:"duplicateOperand"`
	EmptyOperandParameter       bool `json:"emptyOperandParameter"`
	MissingRequiredOperand      bool `json:"missingRequiredOperand"`
	DependencyViolation         bool `json:"dependencyViolation"`
	MutuallyExclusive           bool `json:"mutuallyExclusive"`
	RequiredGroup               bool `json:"requiredGroup"`
	MissingInlineData           bool `json:"missingInlineData"`
	UnknownSubOperand           bool `json:"unknownSubOperand"`
	SubOperandValidation        bool `json:"subOperandValidation"`
	ContentBeyondColumn72       bool `json:"contentBeyondColumn72"`
	StandaloneCommentBetweenMCS bool `json:"standaloneCommentBetweenMCS"`
}

// DefaultDiagnosticsConfig returns a config with all diagnostics enabled
func DefaultDiagnosticsConfig() *DiagnosticsConfig {
	return &DiagnosticsConfig{
		UnknownStatement:            true,
		InvalidLanguageId:           true,
		UnbalancedParentheses:       true,
		MissingTerminator:           true,
		MissingParameter:            true,
		UnknownOperand:              true,
		DuplicateOperand:            true,
		EmptyOperandParameter:       true,
		MissingRequiredOperand:      true,
		DependencyViolation:         true,
		MutuallyExclusive:           true,
		RequiredGroup:               true,
		MissingInlineData:           true,
		UnknownSubOperand:           true,
		SubOperandValidation:        true,
		ContentBeyondColumn72:       true,
		StandaloneCommentBetweenMCS: true,
	}
}

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
	formattingProvider  *formatting.Provider
	symbolProvider      *symbols.Provider
	referencesProvider  *references.Provider
	server              *lsp.Server
	diagnosticsConfig   *DiagnosticsConfig
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
	formattingProvider := formatting.NewProvider()
	symbolProvider := symbols.NewProvider()
	referencesProvider := references.NewProvider()

	return &Handler{
		version:             version,
		documents:           make(map[string]string),
		parsedDocuments:     make(map[string]*parser.Document),
		parser:              parserInstance,
		completionProvider:  completionProvider,
		hoverProvider:       hoverProvider,
		diagnosticsProvider: diagnosticsProvider,
		semanticProvider:    semanticProvider,
		formattingProvider:  formattingProvider,
		symbolProvider:      symbolProvider,
		referencesProvider:  referencesProvider,
		diagnosticsConfig:   DefaultDiagnosticsConfig(),
	}, nil
}

// SetServer sets the LSP server (for sending notifications)
func (h *Handler) SetServer(server *lsp.Server) {
	h.server = server
}

// Initialize handles the initialize request
func (h *Handler) Initialize(params lsp.InitializeParams) (*lsp.InitializeResult, error) {
	logger.Info("Initializing LSP server")

	// Process initialization options for diagnostics configuration
	if params.InitializationOptions != nil && params.InitializationOptions.Diagnostics != nil {
		opts := params.InitializationOptions.Diagnostics
		h.diagnosticsConfig = &DiagnosticsConfig{
			UnknownStatement:       opts.UnknownStatement,
			InvalidLanguageId:      opts.InvalidLanguageId,
			UnbalancedParentheses:  opts.UnbalancedParentheses,
			MissingTerminator:      opts.MissingTerminator,
			MissingParameter:       opts.MissingParameter,
			UnknownOperand:         opts.UnknownOperand,
			DuplicateOperand:       opts.DuplicateOperand,
			EmptyOperandParameter:  opts.EmptyOperandParameter,
			MissingRequiredOperand: opts.MissingRequiredOperand,
			DependencyViolation:    opts.DependencyViolation,
			MutuallyExclusive:      opts.MutuallyExclusive,
			RequiredGroup:          opts.RequiredGroup,
			MissingInlineData:      opts.MissingInlineData,
			UnknownSubOperand:      opts.UnknownSubOperand,
			SubOperandValidation:        opts.SubOperandValidation,
			ContentBeyondColumn72:       opts.ContentBeyondColumn72,
			StandaloneCommentBetweenMCS: opts.StandaloneCommentBetweenMCS,
		}
		logger.Info("Diagnostics config received from client: MissingRequiredOperand=%v, UnknownOperand=%v, ContentBeyondColumn72=%v",
			opts.MissingRequiredOperand, opts.UnknownOperand, opts.ContentBeyondColumn72)
	} else {
		logger.Info("Using default diagnostics config (all enabled)")
	}

	// Process initialization options for formatting configuration
	if params.InitializationOptions != nil && params.InitializationOptions.Formatting != nil {
		opts := params.InitializationOptions.Formatting
		h.formattingProvider.SetConfig(&formatting.Config{
			Enabled:             opts.Enabled,
			IndentContinuation:  opts.IndentContinuation,
			OneOperandPerLine:   opts.OneOperandPerLine,
			MoveLeadingComments: opts.MoveLeadingComments,
		})
		logger.Info("Formatting config received from client: Enabled=%v, IndentContinuation=%d, OneOperandPerLine=%v, MoveLeadingComments=%v",
			opts.Enabled, opts.IndentContinuation, opts.OneOperandPerLine, opts.MoveLeadingComments)
	} else {
		logger.Info("Using default formatting config")
	}

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
			HoverProvider:                   true,
			DocumentFormattingProvider:      true,
			DocumentRangeFormattingProvider: true,
			DocumentSymbolProvider:          true,
			DefinitionProvider:              true,
			ReferencesProvider:              true,
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

	// Get parsed document and text from cache
	h.documentsMutex.RLock()
	doc, exists := h.parsedDocuments[uri]
	text, textExists := h.documents[uri]
	h.documentsMutex.RUnlock()

	if !exists {
		logger.Debug("No parsed document found for diagnostics: %s", uri)
		return
	}

	if !textExists {
		logger.Debug("No document text found for diagnostics: %s", uri)
		return
	}

	// Convert handler config to diagnostics config
	diagConfig := &diagnostics.Config{
		UnknownStatement:            h.diagnosticsConfig.UnknownStatement,
		InvalidLanguageId:           h.diagnosticsConfig.InvalidLanguageId,
		UnbalancedParentheses:       h.diagnosticsConfig.UnbalancedParentheses,
		MissingTerminator:           h.diagnosticsConfig.MissingTerminator,
		MissingParameter:            h.diagnosticsConfig.MissingParameter,
		UnknownOperand:              h.diagnosticsConfig.UnknownOperand,
		DuplicateOperand:            h.diagnosticsConfig.DuplicateOperand,
		EmptyOperandParameter:       h.diagnosticsConfig.EmptyOperandParameter,
		MissingRequiredOperand:      h.diagnosticsConfig.MissingRequiredOperand,
		DependencyViolation:         h.diagnosticsConfig.DependencyViolation,
		MutuallyExclusive:           h.diagnosticsConfig.MutuallyExclusive,
		RequiredGroup:               h.diagnosticsConfig.RequiredGroup,
		MissingInlineData:           h.diagnosticsConfig.MissingInlineData,
		UnknownSubOperand:           h.diagnosticsConfig.UnknownSubOperand,
		SubOperandValidation:        h.diagnosticsConfig.SubOperandValidation,
		ContentBeyondColumn72:       h.diagnosticsConfig.ContentBeyondColumn72,
		StandaloneCommentBetweenMCS: h.diagnosticsConfig.StandaloneCommentBetweenMCS,
	}

	// Generate diagnostics from AST with config and text (for column 72 checking)
	diags := h.diagnosticsProvider.AnalyzeASTWithConfigAndText(doc, diagConfig, text)

	params := map[string]interface{}{
		"uri":         uri,
		"diagnostics": diags,
	}

	if err := h.server.SendNotification("textDocument/publishDiagnostics", params); err != nil {
		logger.Error("Failed to publish diagnostics: %v", err)
	}
}

// WorkspaceDidChangeConfiguration handles configuration changes from the client
func (h *Handler) WorkspaceDidChangeConfiguration(params lsp.DidChangeConfigurationParams) error {
	logger.Info("Configuration changed")

	// Update diagnostics config if provided
	if params.Settings != nil && params.Settings.Smpe != nil && params.Settings.Smpe.Diagnostics != nil {
		opts := params.Settings.Smpe.Diagnostics
		h.diagnosticsConfig = &DiagnosticsConfig{
			UnknownStatement:       opts.UnknownStatement,
			InvalidLanguageId:      opts.InvalidLanguageId,
			UnbalancedParentheses:  opts.UnbalancedParentheses,
			MissingTerminator:      opts.MissingTerminator,
			MissingParameter:       opts.MissingParameter,
			UnknownOperand:         opts.UnknownOperand,
			DuplicateOperand:       opts.DuplicateOperand,
			EmptyOperandParameter:  opts.EmptyOperandParameter,
			MissingRequiredOperand: opts.MissingRequiredOperand,
			DependencyViolation:    opts.DependencyViolation,
			MutuallyExclusive:      opts.MutuallyExclusive,
			RequiredGroup:          opts.RequiredGroup,
			MissingInlineData:      opts.MissingInlineData,
			UnknownSubOperand:      opts.UnknownSubOperand,
			SubOperandValidation:        opts.SubOperandValidation,
			ContentBeyondColumn72:       opts.ContentBeyondColumn72,
			StandaloneCommentBetweenMCS: opts.StandaloneCommentBetweenMCS,
		}
		logger.Info("Updated diagnostics config: MissingRequiredOperand=%v, ContentBeyondColumn72=%v",
			opts.MissingRequiredOperand, opts.ContentBeyondColumn72)

		// Republish diagnostics for all open documents
		h.republishAllDiagnostics()
	}

	// Update formatting config if provided
	if params.Settings != nil && params.Settings.Smpe != nil && params.Settings.Smpe.Formatting != nil {
		opts := params.Settings.Smpe.Formatting
		h.formattingProvider.SetConfig(&formatting.Config{
			Enabled:             opts.Enabled,
			IndentContinuation:  opts.IndentContinuation,
			OneOperandPerLine:   opts.OneOperandPerLine,
			MoveLeadingComments: opts.MoveLeadingComments,
		})
		logger.Info("Updated formatting config: Enabled=%v, IndentContinuation=%d, OneOperandPerLine=%v, MoveLeadingComments=%v",
			opts.Enabled, opts.IndentContinuation, opts.OneOperandPerLine, opts.MoveLeadingComments)
	}

	return nil
}

// republishAllDiagnostics republishes diagnostics for all open documents
func (h *Handler) republishAllDiagnostics() {
	h.documentsMutex.RLock()
	uris := make([]string, 0, len(h.documents))
	for uri := range h.documents {
		uris = append(uris, uri)
	}
	h.documentsMutex.RUnlock()

	logger.Info("Republishing diagnostics for %d open documents", len(uris))
	for _, uri := range uris {
		h.publishDiagnostics(uri)
	}
}

// TextDocumentFormatting handles document formatting request
func (h *Handler) TextDocumentFormatting(params lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	logger.Debug("Formatting requested for: %s", params.TextDocument.URI)

	// Check if formatting is enabled
	if !h.formattingProvider.GetConfig().Enabled {
		logger.Debug("Formatting is disabled")
		return nil, nil
	}

	h.documentsMutex.RLock()
	text, textExists := h.documents[params.TextDocument.URI]
	doc, hasDoc := h.parsedDocuments[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !textExists {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	// Ensure we have a parsed document
	if !hasDoc {
		logger.Debug("No parsed document found for formatting, parsing now: %s", params.TextDocument.URI)
		h.documentsMutex.Lock()
		doc = h.parser.Parse(text)
		h.parsedDocuments[params.TextDocument.URI] = doc
		h.documentsMutex.Unlock()
	}

	edits := h.formattingProvider.FormatDocument(doc, text)
	logger.Debug("Formatting returned %d edits", len(edits))

	return edits, nil
}

// TextDocumentRangeFormatting handles range formatting request
func (h *Handler) TextDocumentRangeFormatting(params lsp.DocumentRangeFormattingParams) ([]lsp.TextEdit, error) {
	logger.Debug("Range formatting requested for: %s (lines %d-%d)",
		params.TextDocument.URI, params.Range.Start.Line, params.Range.End.Line)

	// Check if formatting is enabled
	if !h.formattingProvider.GetConfig().Enabled {
		logger.Debug("Formatting is disabled")
		return nil, nil
	}

	h.documentsMutex.RLock()
	text, textExists := h.documents[params.TextDocument.URI]
	doc, hasDoc := h.parsedDocuments[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !textExists {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	// Ensure we have a parsed document
	if !hasDoc {
		logger.Debug("No parsed document found for range formatting, parsing now: %s", params.TextDocument.URI)
		h.documentsMutex.Lock()
		doc = h.parser.Parse(text)
		h.parsedDocuments[params.TextDocument.URI] = doc
		h.documentsMutex.Unlock()
	}

	edits := h.formattingProvider.FormatRange(doc, text, params.Range.Start.Line, params.Range.End.Line)
	logger.Debug("Range formatting returned %d edits", len(edits))

	return edits, nil
}

// UpdateFormattingConfig updates the formatting configuration
func (h *Handler) UpdateFormattingConfig(config *formatting.Config) {
	h.formattingProvider.SetConfig(config)
}

// TextDocumentDocumentSymbol handles document symbol request
func (h *Handler) TextDocumentDocumentSymbol(params lsp.DocumentSymbolParams) ([]lsp.DocumentSymbol, error) {
	logger.Debug("Document symbols requested for: %s", params.TextDocument.URI)

	h.documentsMutex.RLock()
	text, textExists := h.documents[params.TextDocument.URI]
	doc, hasDoc := h.parsedDocuments[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !textExists {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	// Ensure we have a parsed document
	if !hasDoc {
		logger.Debug("No parsed document found for symbols, parsing now: %s", params.TextDocument.URI)
		h.documentsMutex.Lock()
		doc = h.parser.Parse(text)
		h.parsedDocuments[params.TextDocument.URI] = doc
		h.documentsMutex.Unlock()
	}

	lines := strings.Split(text, "\n")
	symbols := h.symbolProvider.GetDocumentSymbols(doc, lines)
	logger.Debug("Document symbols returned %d symbols", len(symbols))

	return symbols, nil
}

// TextDocumentDefinition handles go-to-definition request
func (h *Handler) TextDocumentDefinition(params lsp.DefinitionParams) (*lsp.Location, error) {
	logger.Debug("Definition requested at %s:%d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	h.documentsMutex.RLock()
	text, textExists := h.documents[params.TextDocument.URI]
	doc, hasDoc := h.parsedDocuments[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !textExists {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	// Ensure we have a parsed document
	if !hasDoc {
		logger.Debug("No parsed document found for definition, parsing now: %s", params.TextDocument.URI)
		h.documentsMutex.Lock()
		doc = h.parser.Parse(text)
		h.parsedDocuments[params.TextDocument.URI] = doc
		h.documentsMutex.Unlock()
	}

	location := h.referencesProvider.GetDefinition(doc, text, params.Position.Line, params.Position.Character)
	if location != nil {
		location.URI = params.TextDocument.URI
		logger.Debug("Definition found at line %d", location.Range.Start.Line)
	} else {
		logger.Debug("No definition found")
	}

	return location, nil
}

// TextDocumentReferences handles find-references request
func (h *Handler) TextDocumentReferences(params lsp.ReferenceParams) ([]lsp.Location, error) {
	logger.Debug("References requested at %s:%d:%d (includeDeclaration=%v)",
		params.TextDocument.URI, params.Position.Line, params.Position.Character, params.Context.IncludeDeclaration)

	h.documentsMutex.RLock()
	text, textExists := h.documents[params.TextDocument.URI]
	doc, hasDoc := h.parsedDocuments[params.TextDocument.URI]
	h.documentsMutex.RUnlock()

	if !textExists {
		logger.Debug("Document not found: %s", params.TextDocument.URI)
		return nil, nil
	}

	// Ensure we have a parsed document
	if !hasDoc {
		logger.Debug("No parsed document found for references, parsing now: %s", params.TextDocument.URI)
		h.documentsMutex.Lock()
		doc = h.parser.Parse(text)
		h.parsedDocuments[params.TextDocument.URI] = doc
		h.documentsMutex.Unlock()
	}

	locations := h.referencesProvider.GetReferences(doc, text, params.Position.Line, params.Position.Character, params.Context.IncludeDeclaration)

	// Set URI for all locations
	for i := range locations {
		locations[i].URI = params.TextDocument.URI
	}

	logger.Debug("Found %d references", len(locations))

	return locations, nil
}
