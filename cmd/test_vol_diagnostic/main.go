package main

import (
	"fmt"
	"os"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
)

func main() {
	// Load smpe.json
	dataPath := os.Getenv("HOME") + "/.local/share/smpe_ls/smpe.json"
	store, err := data.Load(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading statements: %v\n", err)
		os.Exit(1)
	}

	// Create diagnostics provider
	dp := diagnostics.NewProvider(store)

	// Test case from test-mac-simple.smpe line 9
	testContent := `++MAC(tetete) FROMDS(DSN(my.test) NUMBER(12) UNIT(SYSDA) VOL()) .`

	fmt.Printf("Testing diagnostics for:\n%s\n\n", testContent)

	// Analyze
	diags := dp.Analyze(testContent)

	fmt.Printf("Found %d diagnostic(s):\n", len(diags))
	for i, diag := range diags {
		fmt.Printf("%d. [Line %d] %s: %s\n",
			i+1,
			diag.Range.Start.Line+1,
			getSeverityString(diag.Severity),
			diag.Message)
	}

	if len(diags) == 0 {
		fmt.Println("  (No diagnostics found)")
	}
}

func getSeverityString(severity int) string {
	switch severity {
	case 1:
		return "ERROR"
	case 2:
		return "WARNING"
	case 3:
		return "INFO"
	case 4:
		return "HINT"
	default:
		return "UNKNOWN"
	}
}
