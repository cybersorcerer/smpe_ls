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

	// Read test file
	content, err := os.ReadFile("examples/test-spaces.smpe")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading test file: %v\n", err)
		os.Exit(1)
	}

	// Parse
	doc := p.Parse(string(content))

	fmt.Printf("Total Statements: %d\n\n", len(doc.Statements))

	// Show all statements with their TO parameters
	for i, stmt := range doc.Statements {
		fmt.Printf("Statement %d: %s\n", i+1, stmt.Name)

		// Find TO operand
		for _, child := range stmt.Children {
			if child.Type == parser.NodeTypeOperand && child.Name == "TO" {
				fmt.Printf("  TO Operand found at line %d\n", child.Position.Line)

				// Show parameters
				if len(child.Children) > 0 {
					wrapper := child.Children[0]
					fmt.Printf("  Wrapper parameter: '%s'\n", wrapper.Value)
					fmt.Printf("  Number of sub-parameters: %d\n", len(wrapper.Children))

					for j, subParam := range wrapper.Children {
						if subParam.Type == parser.NodeTypeParameter {
							fmt.Printf("    %d: '%s' at line %d, char %d\n",
								j+1, subParam.Value, subParam.Position.Line, subParam.Position.Character)
						}
					}
				}
			}
		}
		fmt.Println()
	}
}
