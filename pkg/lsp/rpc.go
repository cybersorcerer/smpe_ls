package lsp

import (
	"encoding/json"
	"fmt"
)

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Notification represents a JSON-RPC notification
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// NewResponse creates a new successful response
func NewResponse(id interface{}, result interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(id interface{}, code int, message string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}

// NewNotification creates a new notification
func NewNotification(method string, params interface{}) *Notification {
	paramsJSON, _ := json.Marshal(params)
	return &Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsJSON,
	}
}

// EncodeMessage encodes a message to JSON-RPC format
func EncodeMessage(msg interface{}) ([]byte, error) {
	content, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
	return append([]byte(header), content...), nil
}
