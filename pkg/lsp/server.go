package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/logger"
)

// Server represents the LSP server
type Server struct {
	reader  *bufio.Reader
	writer  io.Writer
	handler Handler
}

// Handler interface for handling LSP requests
type Handler interface {
	Initialize(params InitializeParams) (*InitializeResult, error)
	TextDocumentDidOpen(params DidOpenTextDocumentParams) error
	TextDocumentDidChange(params DidChangeTextDocumentParams) error
	TextDocumentDidClose(params DidCloseTextDocumentParams) error
	TextDocumentCompletion(params CompletionParams) ([]CompletionItem, error)
	TextDocumentHover(params HoverParams) (*Hover, error)
	TextDocumentSemanticTokensFull(params SemanticTokensParams) (*SemanticTokens, error)
	WorkspaceDidChangeConfiguration(params DidChangeConfigurationParams) error
}

// NewServer creates a new LSP server
func NewServer(reader io.Reader, writer io.Writer, handler Handler) *Server {
	return &Server{
		reader:  bufio.NewReader(reader),
		writer:  writer,
		handler: handler,
	}
}

// Start starts the server
func (s *Server) Start() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				logger.Info("Client disconnected")
				return nil
			}
			logger.Error("Error reading message: %v", err)
			return err
		}

		if err := s.handleMessage(msg); err != nil {
			logger.Error("Error handling message: %v", err)
		}
	}
}

// readMessage reads a message from the client
func (s *Server) readMessage() ([]byte, error) {
	// Read headers
	headers := make(map[string]string)
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	// Get content length
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %v", err)
	}

	// Read content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(s.reader, content); err != nil {
		return nil, err
	}

	logger.Debug("Received message: %s", string(content))
	return content, nil
}

// handleMessage handles a message from the client
func (s *Server) handleMessage(msg []byte) error {
	// Parse as generic message to check for ID
	var genericMsg map[string]interface{}
	if err := json.Unmarshal(msg, &genericMsg); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	// If message has an "id" field, it's a request
	if _, hasID := genericMsg["id"]; hasID {
		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			return fmt.Errorf("invalid request: %v", err)
		}
		return s.handleRequest(&req)
	}

	// Otherwise it's a notification
	var notif Notification
	if err := json.Unmarshal(msg, &notif); err != nil {
		return fmt.Errorf("invalid notification: %v", err)
	}
	return s.handleNotification(&notif)
}

// handleRequest handles a request from the client
func (s *Server) handleRequest(req *Request) error {
	logger.Info("Handling request: method=%s, id=%v", req.Method, req.ID)
	logger.Debug("Handling request: %s", req.Method)

	switch req.Method {
	case "initialize":
		logger.Debug("Initialize params raw: %s", string(req.Params))
		var params InitializeParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.sendErrorResponse(req.ID, InvalidParams, "Invalid params")
		}

		result, err := s.handler.Initialize(params)
		if err != nil {
			return s.sendErrorResponse(req.ID, InternalError, err.Error())
		}

		return s.sendResponse(req.ID, result)

	case "textDocument/completion":
		var params CompletionParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.sendErrorResponse(req.ID, InvalidParams, "Invalid params")
		}

		result, err := s.handler.TextDocumentCompletion(params)
		if err != nil {
			return s.sendErrorResponse(req.ID, InternalError, err.Error())
		}

		return s.sendResponse(req.ID, result)

	case "textDocument/hover":
		var params HoverParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.sendErrorResponse(req.ID, InvalidParams, "Invalid params")
		}

		result, err := s.handler.TextDocumentHover(params)
		if err != nil {
			return s.sendErrorResponse(req.ID, InternalError, err.Error())
		}

		return s.sendResponse(req.ID, result)

	case "textDocument/semanticTokens/full":
		var params SemanticTokensParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.sendErrorResponse(req.ID, InvalidParams, "Invalid params")
		}

		result, err := s.handler.TextDocumentSemanticTokensFull(params)
		if err != nil {
			return s.sendErrorResponse(req.ID, InternalError, err.Error())
		}

		return s.sendResponse(req.ID, result)

	case "shutdown":
		return s.sendResponse(req.ID, nil)

	// Optional capabilities - respond with null to indicate not supported
	case "textDocument/documentSymbol",
		"textDocument/definition",
		"textDocument/references",
		"textDocument/formatting",
		"textDocument/rangeFormatting",
		"textDocument/onTypeFormatting",
		"textDocument/codeAction",
		"textDocument/codeLens",
		"textDocument/rename",
		"textDocument/signatureHelp",
		"textDocument/documentHighlight",
		"workspace/symbol",
		"workspace/executeCommand":
		logger.Debug("Unsupported method: %s", req.Method)
		return s.sendResponse(req.ID, nil)

	default:
		logger.Debug("Unknown method: %s", req.Method)
		return s.sendErrorResponse(req.ID, MethodNotFound, "Method not found")
	}
}

// handleNotification handles a notification from the client
func (s *Server) handleNotification(notif *Notification) error {
	logger.Debug("Handling notification: %s", notif.Method)

	switch notif.Method {
	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(notif.Params, &params); err != nil {
			return err
		}
		return s.handler.TextDocumentDidOpen(params)

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(notif.Params, &params); err != nil {
			return err
		}
		return s.handler.TextDocumentDidChange(params)

	case "textDocument/didClose":
		var params DidCloseTextDocumentParams
		if err := json.Unmarshal(notif.Params, &params); err != nil {
			return err
		}
		return s.handler.TextDocumentDidClose(params)

	case "exit":
		logger.Info("Received exit notification")
		return io.EOF

	case "initialized":
		// Nothing to do
		return nil

	case "workspace/didChangeConfiguration":
		var params DidChangeConfigurationParams
		if err := json.Unmarshal(notif.Params, &params); err != nil {
			return err
		}
		return s.handler.WorkspaceDidChangeConfiguration(params)

	default:
		logger.Debug("Unhandled notification: %s", notif.Method)
		return nil
	}
}

// sendResponse sends a response to the client
func (s *Server) sendResponse(id interface{}, result interface{}) error {
	resp := NewResponse(id, result)
	return s.writeMessage(resp)
}

// sendErrorResponse sends an error response to the client
func (s *Server) sendErrorResponse(id interface{}, code int, message string) error {
	resp := NewErrorResponse(id, code, message)
	return s.writeMessage(resp)
}

// SendNotification sends a notification to the client
func (s *Server) SendNotification(method string, params interface{}) error {
	notif := NewNotification(method, params)
	return s.writeMessage(notif)
}

// writeMessage writes a message to the client
func (s *Server) writeMessage(msg interface{}) error {
	data, err := EncodeMessage(msg)
	if err != nil {
		return err
	}

	logger.Debug("Sending message: %s", string(data))
	_, err = s.writer.Write(data)
	return err
}

// Notification parameter types

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}
