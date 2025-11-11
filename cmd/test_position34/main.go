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

	// Test the exact case from test-mac-simple.smpe line 9 position 34
	content := `++MAC(tetete) FROMDS(DSN(my.test) )`

	fmt.Printf("Testing: '%s'\n", content)
	fmt.Printf("Position 34 is here: %s\n", content[:34]+"^"+content[34:])
	fmt.Println()

	// Parse
	doc := p.Parse(content)

	// Debug: Show parsed structure
	if len(doc.Statements) > 0 {
		stmt := doc.Statements[0]
		fmt.Printf("Statement: %s (Char=%d, Len=%d)\n", stmt.Name, stmt.Position.Character, stmt.Position.Length)
		for i, child := range stmt.Children {
			fmt.Printf("  Child %d: Type=%v Name='%s' (Char=%d, Len=%d)\n", i, child.Type, child.Name, child.Position.Character, child.Position.Length)
			for j, subChild := range child.Children {
				fmt.Printf("    SubChild %d.%d: Type=%v Name='%s' (Char=%d, Len=%d)\n", i, j, subChild.Type, subChild.Name, subChild.Position.Character, subChild.Position.Length)
				for k, subSubChild := range subChild.Children {
					fmt.Printf("      SubSubChild %d.%d.%d: Type=%v Name='%s'\n", i, j, k, subSubChild.Type, subSubChild.Name)
				}
			}
		}
	}
	fmt.Println()

	// Get completions at position 34 (after DSN(my.test) )
	items := cp.GetCompletionsAST(doc, content, 0, 34)

	fmt.Printf("Completions at position 34: %d\n", len(items))
	if len(items) == 0 {
		fmt.Println("  No completions returned!")
	} else {
		for i, item := range items {
			fmt.Printf("  %d. %s - %s\n", i+1, item.Label, item.Detail)
		}
	}
}
