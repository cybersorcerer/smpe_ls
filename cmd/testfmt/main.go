package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/formatting"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

func main() {
	// Load MCS data
	store, err := data.Load("./data/smpe.json")
	if err != nil {
		fmt.Println("Error loading MCS data:", err)
		return
	}

	// Read file from command line or use default test case
	var text string
	if len(os.Args) > 1 {
		content, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		text = string(content)
	} else {
		// Default test case
		text = `++USERMOD(UMOD001)  REWORK(2023296)
                    DESC('Python script')
/**
 * Installs a python xxx script into z/OS Unix
 *
 * Rep: gitlab.com/path/xxxxxxxxxxxxxxxxxxxxxx/xx/yyyyyyyyyyyyyyyyyyyyyy.git
 *
 * Note to installer:
 *
 * This usermod clones a repo from gitlab. This means as applier you need to
 * have a full configured git environment.
 * Please make sure you saved a ssh-id-key to ~/.ssh and uploaded
 * the relating Pubkey to gitlab as trustworthy key. For questions:
 * ask someone with git/ssh experience or follow the guide on
 * https://docs.gitlab.com/user/ssh/
 *
 */
 .
++VER(Z038)       FMID(MYFUNC)
                  .
++SHELLSCR(MYSCRIPT)  SYSLIB(DD1)
                      SHSCRIPT(MYSCRIPT, POST)
                      DISTLIB(DD2)
                      TXLIB(DD3)
                      .`
	}

	lines := strings.Split(text, "\n")
	fmt.Println("=== Lines analysis ===")
	for i, line := range lines {
		fmt.Printf("Line %2d: %q\n", i, line)
	}

	// Find terminator location
	fmt.Println("\n=== Terminator analysis ===")
	for i, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if strings.HasSuffix(cleanLine, ".") || cleanLine == "." {
			fmt.Printf("Terminator found at line %d: %q\n", i, line)
		}
	}

	// Create parser and parse
	p := parser.NewParser(store.Statements)
	doc := p.Parse(text)

	fmt.Println("\n=== Parser results ===")
	fmt.Println("Number of statements:", len(doc.Statements))
	fmt.Println("Number of comments:", len(doc.Comments))

	for i, comment := range doc.Comments {
		fmt.Printf("\nComment %d:\n", i)
		fmt.Printf("  Line: %d, Char: %d\n", comment.Position.Line, comment.Position.Character)
		lineCount := strings.Count(comment.Value, "\n") + 1
		fmt.Printf("  Lines in comment: %d\n", lineCount)
	}

	for i, stmt := range doc.Statements {
		fmt.Printf("\nStatement %d: %s\n", i, stmt.Name)
		fmt.Printf("  Position: line %d\n", stmt.Position.Line)
		fmt.Printf("  HasTerminator: %v\n", stmt.HasTerminator)
		// Find where statement content ends
		lastChildLine := stmt.Position.Line
		for _, child := range stmt.Children {
			if child.Position.Line > lastChildLine {
				lastChildLine = child.Position.Line
			}
			// Debug: print child structure
			fmt.Printf("  Child: Type=%d Name=%q Value=%q\n", child.Type, child.Name, child.Value)
			for _, gc := range child.Children {
				fmt.Printf("    GC: Type=%d Name=%q Value=%q\n", gc.Type, gc.Name, gc.Value)
			}
		}
		fmt.Printf("  Last child on line: %d\n", lastChildLine)
	}

	// Create formatting provider
	fp := formatting.NewProvider()
	config := formatting.DefaultConfig()
	config.OneOperandPerLine = true
	fp.SetConfig(config)

	// Format
	edits := fp.FormatDocument(doc, text)

	fmt.Println("\n=== Edits ===")
	fmt.Println("Number of edits:", len(edits))
	for i, edit := range edits {
		fmt.Printf("\nEdit %d:\n", i)
		fmt.Printf("  Range: (%d,%d) to (%d,%d)\n",
			edit.Range.Start.Line, edit.Range.Start.Character,
			edit.Range.End.Line, edit.Range.End.Character)
		fmt.Printf("  NewText length: %d chars\n", len(edit.NewText))
		fmt.Printf("  NewText:\n---\n%s\n---\n", edit.NewText)
	}

	// Apply edits
	result := applyEdits(text, edits)
	fmt.Println("\n=== Formatted Output ===")
	fmt.Println(result)
}

func applyEdits(text string, edits []lsp.TextEdit) string {
	lines := strings.Split(text, "\n")

	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		startLine := edit.Range.Start.Line
		startChar := edit.Range.Start.Character
		endLine := edit.Range.End.Line
		endChar := edit.Range.End.Character

		var result []string
		result = append(result, lines[:startLine]...)

		if startLine < len(lines) && startChar < len(lines[startLine]) {
			before := lines[startLine][:startChar]
			after := ""
			if endLine < len(lines) && endChar <= len(lines[endLine]) {
				after = lines[endLine][endChar:]
			}
			newLines := strings.Split(edit.NewText, "\n")
			if len(newLines) > 0 {
				newLines[0] = before + newLines[0]
				newLines[len(newLines)-1] = newLines[len(newLines)-1] + after
			}
			result = append(result, newLines...)
		} else {
			result = append(result, strings.Split(edit.NewText, "\n")...)
		}

		if endLine+1 < len(lines) {
			result = append(result, lines[endLine+1:]...)
		}

		lines = result
	}

	return strings.Join(lines, "\n")
}
