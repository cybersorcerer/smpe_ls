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

	// Test problematic statement
	content := `++JCLIN FROMDS(DSN(MY.DATA.SET) NUMBER(1)) RELFILE(2) .`

	// Parse
	doc := p.Parse(content)

	fmt.Printf("Parsed %d statements\n\n", len(doc.Statements))

	// Create diagnostics provider
	diagProvider := diagnostics.NewProvider(store)

	// Analyze
	diags := diagProvider.AnalyzeAST(doc)

	fmt.Printf("Total diagnostics: %d\n\n", len(diags))

	if len(diags) > 0 {
		fmt.Println("Diagnostics:")
		for i, d := range diags {
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
			fmt.Printf("  %d. [%s] %s\n", i+1, severity, d.Message)
		}
	} else {
		fmt.Println("No diagnostics - all good!")
	}
}
