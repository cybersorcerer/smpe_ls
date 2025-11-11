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

	// Read test file
	content, err := os.ReadFile("examples/test-assign.smpe")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading test file: %v\n", err)
		os.Exit(1)
	}

	// Parse
	doc := p.Parse(string(content))

	fmt.Printf("Statements: %d\n", len(doc.Statements))
	fmt.Printf("Comments: %d\n", len(doc.Comments))
	fmt.Printf("Expecting inline: %d\n\n", len(doc.StatementsExpectingInline))

	// Debug: Show all statements with positions
	fmt.Println("All statements:")
	for i, stmt := range doc.Statements {
		fmt.Printf("  %d: %s at line %d, char %d\n", i, stmt.Name, stmt.Position.Line, stmt.Position.Character)
		// Show operands and their parameters with positions
		for _, child := range stmt.Children {
			if child.Type == parser.NodeTypeOperand {
				fmt.Printf("      Operand: %s at line %d, char %d\n", child.Name, child.Position.Line, child.Position.Character)
				for _, param := range child.Children {
					if param.Type == parser.NodeTypeParameter {
						fmt.Printf("        Parameter: '%s' at line %d, char %d\n", param.Value, param.Position.Line, param.Position.Character)
						// Show child parameters if any
						for _, subParam := range param.Children {
							if subParam.Type == parser.NodeTypeParameter {
								fmt.Printf("          SubParam: '%s' at line %d, char %d\n", subParam.Value, subParam.Position.Line, subParam.Position.Character)
							}
						}
					}
				}
			}
		}
	}
	fmt.Println()

	// Debug: Show all comments
	fmt.Println("All comments:")
	for i, comment := range doc.Comments {
		fmt.Printf("  %d: at line %d\n", i, comment.Position.Line)
	}
	fmt.Println()

	// Create diagnostics provider
	diagProvider := diagnostics.NewProvider(store)

	// Analyze
	diags := diagProvider.AnalyzeAST(doc)

	fmt.Printf("Total diagnostics: %d\n\n", len(diags))

	// Print diagnostics related to missing inline data
	fmt.Println("Diagnostics containing 'inline data':")
	for _, d := range diags {
		if contains(d.Message, "inline data") {
			fmt.Printf("  Line %d: %s\n", d.Range.Start.Line+1, d.Message)
		}
	}

	// Print first 10 diagnostics
	fmt.Println("\nFirst 10 diagnostics:")
	for i, d := range diags {
		if i >= 10 {
			break
		}
		fmt.Printf("  %d. Line %d: %s\n", i+1, d.Range.Start.Line+1, d.Message)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
