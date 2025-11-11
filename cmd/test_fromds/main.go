package main

import (
	"fmt"
	"os"

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

	// Test single statement
	content := `++JCLIN FROMDS(DSN(MY.DATA.SET) NUMBER(1)) RELFILE(2) .`

	// Parse
	doc := p.Parse(content)

	fmt.Printf("Total Statements: %d\n\n", len(doc.Statements))

	if len(doc.Statements) > 0 {
		stmt := doc.Statements[0]
		fmt.Printf("Statement: %s\n", stmt.Name)

		// Show all children
		for i, child := range stmt.Children {
			fmt.Printf("\nChild %d: Type=%v, Name='%s'\n", i, child.Type, child.Name)

			if child.Type == parser.NodeTypeOperand {
				fmt.Printf("  Operand: %s\n", child.Name)
				fmt.Printf("  Number of children: %d\n", len(child.Children))

				for j, subChild := range child.Children {
					fmt.Printf("    SubChild %d: Type=%v", j, subChild.Type)
					if subChild.Type == parser.NodeTypeParameter {
						fmt.Printf(", Value='%s'", subChild.Value)
					} else if subChild.Type == parser.NodeTypeOperand {
						fmt.Printf(", Name='%s'", subChild.Name)
					}
					fmt.Println()

					// Show grandchildren
					if len(subChild.Children) > 0 {
						for k, grandChild := range subChild.Children {
							fmt.Printf("      GrandChild %d: Type=%v", k, grandChild.Type)
							if grandChild.Type == parser.NodeTypeParameter {
								fmt.Printf(", Value='%s'", grandChild.Value)
							}
							fmt.Println()
						}
					}
				}
			}
		}
	}
}
