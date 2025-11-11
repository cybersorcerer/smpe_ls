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

	// Create parser
	p := parser.NewParser(store.Statements)

	// Test cases
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "Missing closing paren",
			code: `++APAR(A44444
    DESCRIPTION(Missing closing paren)
    .`,
		},
		{
			name: "Extra closing paren",
			code: `++APAR(A55555))
    DESCRIPTION(Extra closing)
    .`,
		},
		{
			name: "Correct parens",
			code: `++APAR(A66666)
    DESCRIPTION(All good)
    .`,
		},
		{
			name: "Missing terminator",
			code: `++APAR(A77777)
    DESCRIPTION(No terminator)`,
		},
	}

	diagProvider := diagnostics.NewProvider(store)

	for _, tc := range testCases {
		fmt.Printf("\n=== %s ===\n", tc.name)
		fmt.Printf("Code: %s\n\n", tc.code)

		doc := p.Parse(tc.code)
		diags := diagProvider.AnalyzeAST(doc)

		if len(doc.Statements) > 0 {
			stmt := doc.Statements[0]
			fmt.Printf("HasTerminator: %v\n", stmt.HasTerminator)
			fmt.Printf("UnbalancedParens: %d\n", stmt.UnbalancedParens)
		}

		if len(diags) > 0 {
			fmt.Println("\nDiagnostics:")
			for _, d := range diags {
				severity := "INFO"
				switch d.Severity {
				case 1:
					severity = "ERROR"
				case 2:
					severity = "WARNING"
				case 3:
					severity = "INFO"
				case 4:
					severity = "HINT"
				}
				fmt.Printf("  [%s] %s\n", severity, d.Message)
			}
		} else {
			fmt.Println("\nNo diagnostics - all good!")
		}
	}
}
