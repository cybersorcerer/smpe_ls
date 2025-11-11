package main

import (
	"fmt"
	"os"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
)

func main() {
	dataPath := os.Getenv("HOME") + "/.local/share/smpe_ls/smpe.json"
	store, err := data.Load(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	dp := diagnostics.NewProvider(store)

	testCases := []struct {
		name string
		content string
	}{
		{
			name: "Empty VOL() - should warn",
			content: `++MAC(test) FROMDS(DSN(my.test) VOL()) .`,
		},
		{
			name: "Valid VOL(ABC123) - should pass",
			content: `++MAC(test) FROMDS(DSN(my.test) VOL(ABC123)) .`,
		},
		{
			name: "Empty NUMBER() and UNIT() - should warn",
			content: `++MAC(test) FROMDS(DSN(my.test) NUMBER() UNIT()) .`,
		},
		{
			name: "All sub-operands valid - should pass",
			content: `++MAC(test) FROMDS(DSN(my.test) NUMBER(12) UNIT(SYSDA) VOL(ABC123)) .`,
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n========== Test %d: %s ==========\n", i+1, tc.name)
		fmt.Printf("Content: %s\n", tc.content)

		diags := dp.Analyze(tc.content)
		fmt.Printf("Diagnostics: %d\n", len(diags))
		for _, diag := range diags {
			fmt.Printf("  - %s\n", diag.Message)
		}
	}
}
