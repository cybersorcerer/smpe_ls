package main

import (
	"fmt"
	"os"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

func main() {
	// Load smpe.json
	dataPath := os.Getenv("HOME") + "/.local/share/smpe_ls/smpe.json"
	store, err := data.Load(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading statements: %v\n", err)
		os.Exit(1)
	}

	// Create parser and diagnostics provider
	p := parser.NewParser(store.Statements)
	dp := diagnostics.NewProvider(store)

	// Test different HFS-type statements
	tests := []struct {
		name    string
		content string
	}{
		{"++HFS", `++HFS(TEST1) DISTLIB(D1) SYSLIB(S1) TEXT.`},
		{"++SHELLSCR", `++SHELLSCR(SCRIPT1) DISTLIB(D1) SYSLIB(S1) TEXT.`},
		{"++AIX1", `++AIX1(FILE1) DISTLIB(D1) SYSLIB(S1) BINARY.`},
		{"++CLIENT1", `++CLIENT1(FILE1) DISTLIB(D1) SYSLIB(S1) TEXT.`},
	}

	for _, test := range tests {
		fmt.Printf("\n=== Testing %s ===\n", test.name)
		content := test.content
		fmt.Println(content)
		doc := p.Parse(content)

		fmt.Printf("\nStatements: %d\n", len(doc.Statements))
		fmt.Printf("StatementsExpectingInline: %d\n", len(doc.StatementsExpectingInline))

		if len(doc.StatementsExpectingInline) > 0 {
			stmt := doc.StatementsExpectingInline[0]
			fmt.Printf("Statement: %s at line %d\n", stmt.Name, stmt.Position.Line)
			fmt.Printf("HasInlineData: %v\n", stmt.HasInlineData)
			fmt.Printf("InlineDataLines: %d\n", stmt.InlineDataLines)

			if stmt.StatementDef != nil {
				fmt.Printf("\nOperands in StatementDef: %d\n", len(stmt.StatementDef.Operands))
			}
		}

		diags := dp.AnalyzeAST(doc)
		fmt.Printf("\nDiagnostics: %d\n", len(diags))
		for _, d := range diags {
			fmt.Printf("  Message: '%s'\n", d.Message)
		}
	}
}
