package main

import (
	"fmt"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/formatting"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
)

func main() {
	// Load MCS data
	store, err := data.Load("./data/smpe.json")
	if err != nil {
		fmt.Println("Error loading MCS data:", err)
		return
	}

	// Test case: ronny.smpe file
	text := `/* FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF */
++EXEC(E1) DISTLIB(DD1) SYSLIB(DD2).
/* REXX */
++HFS(H0123456) DISTLIB(DD1) LINK(../testlink) SYSLIB(DD2) .
#!/bin/env bash
++VER(Z038)
.`

	// Create parser and parse
	p := parser.NewParser(store.Statements)
	doc := p.Parse(text)

	// Create formatting provider with OneOperandPerLine=true
	fp := formatting.NewProvider()
	config := formatting.DefaultConfig()
	config.OneOperandPerLine = true
	config.MoveLeadingComments = true
	fp.SetConfig(config)

	fmt.Println("Input:")
	fmt.Println(text)
	fmt.Println()

	fmt.Println("Config.OneOperandPerLine:", config.OneOperandPerLine)
	fmt.Println("Number of statements:", len(doc.Statements))
	fmt.Println("Number of comments:", len(doc.Comments))

	for i, comment := range doc.Comments {
		fmt.Printf("\nComment %d:\n", i)
		fmt.Printf("  Line: %d, Char: %d\n", comment.Position.Line, comment.Position.Character)
		fmt.Printf("  Value: %q\n", comment.Value)
	}

	for i, stmt := range doc.Statements {
		fmt.Printf("\nStatement %d: %s\n", i, stmt.Name)
		fmt.Println("  HasTerminator:", stmt.HasTerminator)
		for _, child := range stmt.Children {
			fmt.Printf("    Child: Type=%v, Name=%s, Value=%s\n", child.Type, child.Name, child.Value)
		}
	}
	fmt.Println()

	// Format
	edits := fp.FormatDocument(doc, text)

	fmt.Println("Number of edits:", len(edits))
	for i, edit := range edits {
		fmt.Printf("Edit %d:\n", i)
		fmt.Printf("  Range: (%d,%d) to (%d,%d)\n",
			edit.Range.Start.Line, edit.Range.Start.Character,
			edit.Range.End.Line, edit.Range.End.Character)
		fmt.Printf("  NewText:\n%s\n", indent(edit.NewText))
	}
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "    |" + line
	}
	return strings.Join(lines, "\n")
}
