package symbols

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

const maxWorkspaceSymbols = 500

// GetWorkspaceSymbols searches for symbols across all .smpe files in the workspace
func (p *Provider) GetWorkspaceSymbols(
	query string,
	parsedDocuments map[string]*parser.Document,
	documentTexts map[string]string,
	rootURI string,
	parserInstance *parser.Parser,
) []lsp.SymbolInformation {
	query = strings.ToUpper(query)
	var results []lsp.SymbolInformation

	// 1. Search already-opened documents
	for uri, doc := range parsedDocuments {
		text := documentTexts[uri]
		lines := strings.Split(text, "\n")
		syms := p.extractSymbolInformation(doc, uri, lines)
		for _, sym := range syms {
			if query == "" || strings.Contains(strings.ToUpper(sym.Name), query) {
				results = append(results, sym)
				if len(results) >= maxWorkspaceSymbols {
					return results
				}
			}
		}
	}

	// 2. Search workspace directory for .smpe files not yet opened
	rootPath := uriToPath(rootURI)
	if rootPath == "" {
		return results
	}

	_ = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible directories
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".smpe") {
			return nil
		}

		fileURI := pathToURI(path)
		// Skip files already processed from open documents
		if _, opened := parsedDocuments[fileURI]; opened {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			logger.Debug("workspace/symbol: cannot read %s: %v", path, err)
			return nil
		}

		doc := parserInstance.Parse(string(content))
		lines := strings.Split(string(content), "\n")
		syms := p.extractSymbolInformation(doc, fileURI, lines)
		for _, sym := range syms {
			if query == "" || strings.Contains(strings.ToUpper(sym.Name), query) {
				results = append(results, sym)
				if len(results) >= maxWorkspaceSymbols {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	return results
}

// extractSymbolInformation converts a parsed document into flat SymbolInformation entries
func (p *Provider) extractSymbolInformation(doc *parser.Document, uri string, lines []string) []lsp.SymbolInformation {
	if doc == nil {
		return nil
	}

	var result []lsp.SymbolInformation
	for _, stmt := range doc.Statements {
		if stmt == nil || stmt.Type != parser.NodeTypeStatement {
			continue
		}

		// Build name with parameter (e.g. "++USERMOD(LJS2012)")
		stmtParam := ""
		for _, child := range stmt.Children {
			if child.Type == parser.NodeTypeParameter && child.Parent == stmt {
				stmtParam = child.Value
				break
			}
		}
		name := stmt.Name
		if stmtParam != "" {
			name = stmt.Name + "(" + stmtParam + ")"
		}

		endLine, endChar := p.getStatementEndPosition(stmt, lines)

		result = append(result, lsp.SymbolInformation{
			Name: name,
			Kind: p.getSymbolKind(stmt.Name),
			Location: lsp.Location{
				URI: uri,
				Range: lsp.Range{
					Start: lsp.Position{Line: stmt.Position.Line, Character: stmt.Position.Character},
					End:   lsp.Position{Line: endLine, Character: endChar},
				},
			},
		})
	}
	return result
}

// uriToPath converts a file:// URI to an OS path
func uriToPath(uri string) string {
	if uri == "" {
		return ""
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	path := parsed.Path
	// On Windows, the path starts with / before the drive letter (e.g. /C:/...)
	if runtime.GOOS == "windows" && len(path) > 2 && path[0] == '/' && path[2] == ':' {
		path = path[1:]
	}
	return path
}

// pathToURI converts an OS path to a file:// URI
func pathToURI(path string) string {
	// Normalize to forward slashes
	path = filepath.ToSlash(path)
	if runtime.GOOS == "windows" {
		// Windows: file:///C:/path
		return "file:///" + path
	}
	// Unix: file:///path (path already starts with /)
	return "file://" + path
}
