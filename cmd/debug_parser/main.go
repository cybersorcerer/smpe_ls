package main

import (
	"fmt"
	"os"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/logger"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

func main() {
	// Initialize logger to stdout
	_ = logger.Init(true) // Enable debug mode

	// Load smpe.json
	dataPath := os.Getenv("HOME") + "/.local/share/smpe_ls/smpe.json"
	store, err := data.Load(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading statements: %v\n", err)
		os.Exit(1)
	}

	// Create parser
	p := parser.NewParser(store.Statements)

	// Test content
	content := `/* Minimal test for single-line vs multi-line HFS statements */

/* Single-line - should generate WARNING */
++HFS(TEST1) DISTLIB(D1) SYSLIB(S1) TEXT.

/* Multi-line - should generate WARNING */
++HFS(TEST2)
    DISTLIB(D1)
    SYSLIB(S1)
    TEXT.
`

	// Parse
	fmt.Println("=== Parsing test content ===")
	fmt.Println(content)
	fmt.Println("===")
	doc := p.Parse(content)

	fmt.Println("\n=== Results ===")
	fmt.Printf("Total statements parsed: %d\n", len(doc.Statements))
	fmt.Printf("Statements expecting inline: %d\n", len(doc.StatementsExpectingInline))

	fmt.Println("\n=== Statements ===")
	for i, stmt := range doc.Statements {
		fmt.Printf("%d: %s (Line %d)\n", i, stmt.Name, stmt.Position.Line+1)
	}

	fmt.Println("\n=== Statements Expecting Inline ===")
	for i, stmt := range doc.StatementsExpectingInline {
		fmt.Printf("%d: %s (Line %d, HasInlineData=%v)\n",
			i, stmt.Name, stmt.Position.Line+1, stmt.HasInlineData)
	}
}
