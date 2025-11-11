package main

import (
	"fmt"
	"os"

	"github.com/cybersorcerer/smpe_ls/internal/completion"
	"github.com/cybersorcerer/smpe_ls/internal/data"
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

	// Create parser and completion provider
	p := parser.NewParser(store.Statements)
	cp := completion.NewProvider(store)

	// Test multiple scenarios
	testCases := []struct {
		name      string
		content   string
		line      int
		character int
	}{
		{
			name:      "Empty FROMDS() - cursor on (",
			content:   `++JCLIN FROMDS() RELFILE(2) .`,
			line:      0,
			character: 14, // On (
		},
		{
			name:      "Empty FROMDS() - cursor on )",
			content:   `++JCLIN FROMDS() RELFILE(2) .`,
			line:      0,
			character: 15, // On )
		},
		{
			name:      "FROMDS(D typed",
			content:   `++JCLIN FROMDS(D) RELFILE(2) .`,
			line:      0,
			character: 15, // After D
		},
		{
			name:      "FROMDS( without closing",
			content:   `++JCLIN FROMDS(`,
			line:      0,
			character: 14, // On the '(' character
		},
		{
			name:      "Real-world case from log",
			content:   `++MAC(tetete) FROMDS()`,
			line:      0,
			character: 21, // On ) - this was failing
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n========== %s ==========\n", tc.name)
		fmt.Printf("Content: '%s'\n", tc.content)
		fmt.Printf("Position: line=%d, char=%d\n\n", tc.line, tc.character)

		content := tc.content

		// Parse
		doc := p.Parse(content)

		// Debug: Show parsed structure
		if len(doc.Statements) > 0 {
			stmt := doc.Statements[0]
			fmt.Printf("Statement: %s (Char=%d, Len=%d)\n", stmt.Name, stmt.Position.Character, stmt.Position.Length)
			for i, child := range stmt.Children {
				fmt.Printf("  Child %d: Type=%v Name='%s' (Char=%d, Len=%d)\n", i, child.Type, child.Name, child.Position.Character, child.Position.Length)
				for j, subChild := range child.Children {
					fmt.Printf("    SubChild %d.%d: Type=%v Name='%s'\n", i, j, subChild.Type, subChild.Name)
				}
			}
		}
		fmt.Println()

		// Get completions
		items := cp.GetCompletionsAST(doc, content, tc.line, tc.character)

		fmt.Printf("Completions: %d\n", len(items))
		if len(items) == 0 {
			fmt.Println("  No completions returned!")
		} else {
			for i, item := range items {
				fmt.Printf("  %d. %s\n", i+1, item.Label)
			}
		}
	}
}
