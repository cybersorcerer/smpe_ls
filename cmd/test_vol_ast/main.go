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

	// Test case from test-mac-simple.smpe line 9
	testContent := `++MAC(tetete) FROMDS(DSN(my.test) NUMBER(12) UNIT(SYSDA) VOL()) .`

	fmt.Printf("Testing AST-based diagnostics for:\n%s\n\n", testContent)

	// Parse
	doc := p.Parse(testContent)

	// Analyze using AST
	diags := dp.AnalyzeAST(doc)

	fmt.Printf("Found %d diagnostic(s):\n", len(diags))
	for i, diag := range diags {
		fmt.Printf("%d. [Line %d, Char %d-%d] %s\n",
			i+1,
			diag.Range.Start.Line+1,
			diag.Range.Start.Character,
			diag.Range.End.Character,
			diag.Message)
	}

	if len(diags) == 0 {
		fmt.Println("  (No diagnostics found)")
	}
}
