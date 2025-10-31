package diagnostics

import (
	"github.com/cybersorcerer/smpe_ls/internal/lexer"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

// Provider provides diagnostics
type Provider struct{}

// NewProvider creates a new diagnostics provider
func NewProvider() *Provider {
	return &Provider{}
}

// Analyze analyzes the text and returns diagnostics
func (p *Provider) Analyze(text string) []lsp.Diagnostic {
	logger.Debug("Analyzing text for diagnostics")

	var diagnostics []lsp.Diagnostic

	// Lex and parse the text
	l := lexer.New(text)
	parser := parser.New(l)
	parser.Parse()

	// Get parser errors
	errors := parser.Errors()
	for _, err := range errors {
		diagnostic := lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 0},
			},
			Severity: lsp.SeverityError,
			Source:   "smpe_ls",
			Message:  err,
		}
		diagnostics = append(diagnostics, diagnostic)
	}

	logger.Debug("Found %d diagnostics", len(diagnostics))
	return diagnostics
}
