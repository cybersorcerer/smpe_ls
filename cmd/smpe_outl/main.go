package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

var (
	version = "v1.3.3"
	commit  = "unknown"
)

// OutlineFile is the top-level JSON output entry per file
type OutlineFile struct {
	File    string          `json:"file"`
	Symbols []OutlineSymbol `json:"symbols"`
}

// OutlineSymbol is a DocumentSymbol-compatible structure with extra fields
type OutlineSymbol struct {
	Name           string          `json:"name"`
	ID             string          `json:"id,omitempty"`
	Kind           lsp.SymbolKind  `json:"kind"`
	HasInlineData  *bool           `json:"hasInlineData,omitempty"`
	Range          *lsp.Range      `json:"range,omitempty"`
	SelectionRange *lsp.Range      `json:"selectionRange,omitempty"`
	Children       []OutlineSymbol `json:"children,omitempty"`
}

func main() {
	jsonMode := flag.Bool("json", false, "Output results in JSON format")
	withRanges := flag.Bool("ranges", false, "Include range and selectionRange in JSON output (default: off)")
	withMeta := flag.Bool("meta", false, "Include hasInlineData field per statement in JSON output (default: off)")
	versionFlag := flag.Bool("version", false, "Show version information")
	shortVersionFlag := flag.Bool("v", false, "Show version information")
	dataFlag := flag.String("data", "", "Path to smpe.json data file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file-pattern>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nPrints the document outline (symbols) of SMP/E MCS files.\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --data <path>   Path to smpe.json data file\n")
		fmt.Fprintf(os.Stderr, "  --json          Output results in JSON format\n")
		fmt.Fprintf(os.Stderr, "  --meta          Include hasInlineData per statement in JSON output\n")
		fmt.Fprintf(os.Stderr, "  --ranges        Include range and selectionRange in JSON output\n")
		fmt.Fprintf(os.Stderr, "  --version, -v   Show version information\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s *.smpe\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --json *.smpe\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --json --meta *.smpe\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --json --ranges *.smpe\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s \"packages/**/*.smpe\"\n", os.Args[0])
	}

	flag.Parse()

	if *versionFlag || *shortVersionFlag {
		fmt.Printf("smpe_outl %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Copyright (c) 2025, 2026 Sir Tobi aka Cybersorcerer\n")
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Collect files
	var files []string
	for _, arg := range flag.Args() {
		matches, err := filepath.Glob(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding pattern '%s': %v\n", arg, err)
			os.Exit(1)
		}
		if len(matches) == 0 {
			files = append(files, arg)
		} else {
			files = append(files, matches...)
		}
	}
	files = uniqueFiles(files)

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No files found matching arguments\n")
		os.Exit(1)
	}

	// Load smpe.json
	dataPath := *dataFlag
	if dataPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error determining home directory: %v\n", err)
			os.Exit(1)
		}
		dataPath = homeDir + "/.local/share/smpe_ls/smpe.json"
	}
	store, err := data.Load(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading smpe.json from %s: %v\n", dataPath, err)
		os.Exit(1)
	}

	var outlines []OutlineFile
	hasError := false

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
			hasError = true
			continue
		}

		contentStr := string(content)
		p := parser.NewParser(store.Statements)
		doc := p.Parse(contentStr)
		lines := strings.Split(contentStr, "\n")

		symbols := buildSymbols(doc, lines, *withRanges, *withMeta)
		outlines = append(outlines, OutlineFile{
			File:    file,
			Symbols: symbols,
		})
	}

	if *jsonMode {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(outlines); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		printMarkdown(outlines)
	}

	if hasError {
		os.Exit(1)
	}
	os.Exit(0)
}

// buildSymbols converts a parsed document into OutlineSymbol entries
func buildSymbols(doc *parser.Document, lines []string, withRanges bool, withMeta bool) []OutlineSymbol {
	if doc == nil {
		return nil
	}

	var symbols []OutlineSymbol
	for _, stmt := range doc.Statements {
		if stmt == nil || stmt.Type != parser.NodeTypeStatement {
			continue
		}
		sym := buildStatementSymbol(stmt, lines, withRanges, withMeta)
		symbols = append(symbols, sym)
	}
	return symbols
}

func buildStatementSymbol(stmt *parser.Node, lines []string, withRanges bool, withMeta bool) OutlineSymbol {
	// Extract statement parameter (e.g. "LJS2012" from "++USERMOD(LJS2012)")
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

	endLine, endChar := getStatementEndPosition(stmt, lines)

	sym := OutlineSymbol{
		Name:     name,
		ID:       stmtParam,
		Kind:     getSymbolKind(stmt.Name),
		Children: buildOperandSymbols(stmt, withRanges),
	}
	if withMeta {
		hasInlineData := stmt.HasInlineData
		sym.HasInlineData = &hasInlineData
	}
	if withRanges {
		r := lsp.Range{
			Start: lsp.Position{Line: stmt.Position.Line, Character: stmt.Position.Character},
			End:   lsp.Position{Line: endLine, Character: endChar},
		}
		sr := lsp.Range{
			Start: lsp.Position{Line: stmt.Position.Line, Character: stmt.Position.Character},
			End:   lsp.Position{Line: stmt.Position.Line, Character: stmt.Position.Character + stmt.Position.Length},
		}
		sym.Range = &r
		sym.SelectionRange = &sr
	}

	return sym
}

func buildOperandSymbols(stmt *parser.Node, withRanges bool) []OutlineSymbol {
	var children []OutlineSymbol

	for _, child := range stmt.Children {
		if child.Type != parser.NodeTypeOperand {
			continue
		}

		paramValue := ""
		for _, grandchild := range child.Children {
			if grandchild.Type == parser.NodeTypeParameter {
				paramValue = grandchild.Value
				break
			}
		}

		if paramValue == "" {
			continue
		}

		name := child.Name + "(" + paramValue + ")"

		childSym := OutlineSymbol{
			Name: name,
			ID:   paramValue,
			Kind: lsp.SymbolKindProperty,
		}
		if withRanges {
			r := lsp.Range{
				Start: lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
				End:   lsp.Position{Line: child.Position.Line, Character: child.Position.Character + child.Position.Length + len(paramValue) + 2},
			}
			sr := lsp.Range{
				Start: lsp.Position{Line: child.Position.Line, Character: child.Position.Character},
				End:   lsp.Position{Line: child.Position.Line, Character: child.Position.Character + child.Position.Length},
			}
			childSym.Range = &r
			childSym.SelectionRange = &sr
		}

		children = append(children, childSym)
	}

	return children
}

func getSymbolKind(stmtName string) lsp.SymbolKind {
	switch stmtName {
	case "++FUNCTION", "++USERMOD", "++PTF", "++APAR":
		return lsp.SymbolKindClass
	case "++VER":
		return lsp.SymbolKindMethod
	case "++IF":
		return lsp.SymbolKindOperator
	case "++MAC", "++SRC", "++MOD":
		return lsp.SymbolKindStruct
	case "++MACUPD", "++SRCUPD":
		return lsp.SymbolKindEvent
	case "++JCLIN":
		return lsp.SymbolKindFile
	default:
		return lsp.SymbolKindFunction
	}
}

func getStatementEndPosition(stmt *parser.Node, lines []string) (int, int) {
	endLine := stmt.Position.Line
	endChar := stmt.Position.Character + stmt.Position.Length

	for _, child := range stmt.Children {
		if child.Position.Line > endLine {
			endLine = child.Position.Line
			endChar = child.Position.Character + child.Position.Length
		} else if child.Position.Line == endLine && child.Position.Character+child.Position.Length > endChar {
			endChar = child.Position.Character + child.Position.Length
		}
		for _, grandchild := range child.Children {
			if grandchild.Position.Line > endLine {
				endLine = grandchild.Position.Line
				endChar = grandchild.Position.Character + grandchild.Position.Length
			} else if grandchild.Position.Line == endLine && grandchild.Position.Character+grandchild.Position.Length > endChar {
				endChar = grandchild.Position.Character + grandchild.Position.Length
			}
		}
	}

	for i := endLine; i < len(lines); i++ {
		line := lines[i]
		for j := 0; j < len(line); j++ {
			if line[j] == '.' {
				return i, j + 1
			}
		}
		if i > stmt.Position.Line {
			for j := 0; j < len(line)-1; j++ {
				if line[j] == '+' && line[j+1] == '+' {
					return endLine, endChar
				}
			}
		}
	}

	return endLine, endChar
}

func printMarkdown(outlines []OutlineFile) {
	fmt.Println("# SMP/E Outline Report")
	fmt.Println()

	for _, outline := range outlines {
		fmt.Printf("## %s\n\n", outline.File)

		if len(outline.Symbols) == 0 {
			fmt.Println("_No symbols found._")
			fmt.Println()
			continue
		}

		for _, sym := range outline.Symbols {
			fmt.Printf("### %s\n", sym.Name)

			for _, child := range sym.Children {
				fmt.Printf("- `%s`\n", child.Name)
			}
			fmt.Println()
		}
	}
}

func uniqueFiles(files []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(files))
	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			result = append(result, file)
		}
	}
	return result
}
