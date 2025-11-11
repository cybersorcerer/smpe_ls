package main

import (
	"fmt"
	"os"
	"strings"

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

	// Create parser
	p := parser.NewParser(store.Statements)

	// Read test file
	content, err := os.ReadFile("examples/test.smpe")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading test file: %v\n", err)
		os.Exit(1)
	}

	// Parse
	doc := p.Parse(string(content))

	fmt.Printf("Total Statements parsed: %d\n\n", len(doc.Statements))

	// Show statements starting from line 180
	fmt.Println("Statements from line 180 onwards:")
	for i, stmt := range doc.Statements {
		if stmt.Position.Line >= 180 {
			// Get statement text preview
			lines := strings.Split(string(content), "\n")
			startLine := stmt.Position.Line
			preview := ""
			if startLine < len(lines) {
				preview = strings.TrimSpace(lines[startLine])
				if len(preview) > 60 {
					preview = preview[:60] + "..."
				}
			}

			fmt.Printf("  %d: %s at line %d (file line %d)\n", i, stmt.Name, stmt.Position.Line, stmt.Position.Line+1)
			fmt.Printf("      Preview: %s\n", preview)
			fmt.Printf("      HasTerminator: %v\n", stmt.HasTerminator)
			if stmt.StatementDef != nil {
				fmt.Printf("      InlineData capable: %v\n", stmt.StatementDef.InlineData)
			}
			fmt.Println()
		}
	}
}
