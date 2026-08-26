package main

import (
	"fmt"
	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"os"
	"path/filepath"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	store, err := data.Load(filepath.Join(home, ".local", "share", "smpe_ls", "smpe.json"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	p := parser.NewParser(store.Statements)

	content, err := os.ReadFile("test-files/test-mac.smpe")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	doc := p.Parse(string(content))

	fmt.Printf("Statements: %d\n", len(doc.Statements))
	fmt.Printf("Comments: %d\n", len(doc.Comments))
	fmt.Printf("Expecting inline: %d\n\n", len(doc.StatementsExpectingInline))

	fmt.Printf("First 15 comments:\n")
	for i, c := range doc.Comments {
		if i >= 15 {
			break
		}
		fmt.Printf("  %d: Line %d (0-indexed)\n", i+1, c.Position.Line)
	}

	if len(doc.StatementsExpectingInline) > 0 {
		fmt.Printf("\nStatements expecting inline:\n")
		for i, s := range doc.StatementsExpectingInline {
			if i >= 5 {
				break
			}
			fmt.Printf("  %s at line %d\n", s.Name, s.Position.Line)
		}
	}
}
