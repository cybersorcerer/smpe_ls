package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/completion"
	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/internal/hover"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

type Expectation struct {
	Line     int
	Severity string // "ERROR", "WARNING", "INFO", "NONE"
	Pattern  string // Optional: regex pattern to match diagnostic message
}

type TestResult struct {
	Name           string
	Passed         bool
	ErrorCount     int
	WarningCount   int
	InfoCount      int
	Diagnostics    []lsp.Diagnostic
	StatementCount int
	TestType       string // "syntax", "diagnostics", "completion", "hover"
	Expectations   []Expectation
	Mismatches     []string // List of expectation failures
}

func main() {
	// Load smpe.json
	dataPath := os.Getenv("HOME") + "/.local/share/smpe_ls/smpe.json"
	store, err := data.Load(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading statements: %v\n", err)
		os.Exit(1)
	}

	// Create providers
	p := parser.NewParser(store.Statements)
	dp := diagnostics.NewProvider(store)
	cp := completion.NewProvider(store)
	hp := hover.NewProvider(store)

	// Find test-files directory
	testDir := "./test-files"
	if len(os.Args) > 1 {
		testDir = os.Args[1]
	}

	// Get all .smpe test files
	files, err := filepath.Glob(filepath.Join(testDir, "*.smpe"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error finding test files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("❌ No test files found in %s\n", testDir)
		os.Exit(1)
	}

	printHeader(testDir, len(files))

	var results []TestResult
	totalPassed := 0
	totalFailed := 0

	for _, file := range files {
		basename := filepath.Base(file)

		// Read test file
		content, err := os.ReadFile(file)
		if err != nil {
			result := TestResult{
				Name:   basename,
				Passed: false,
			}
			results = append(results, result)
			totalFailed++
			continue
		}

		// Run all tests on this file
		result := runTestsOnFile(basename, string(content), p, dp, cp, hp)
		results = append(results, result)

		if result.Passed {
			totalPassed++
		} else {
			totalFailed++
		}
	}

	printResults(results, totalPassed, totalFailed)

	if totalFailed > 0 {
		os.Exit(1)
	}
}

func runTestsOnFile(filename, content string, p *parser.Parser, dp *diagnostics.Provider, cp *completion.Provider, hp *hover.Provider) TestResult {
	result := TestResult{
		Name:     filename,
		TestType: "diagnostics",
	}

	// Parse expectations from comments
	result.Expectations = parseExpectations(content)

	// Parse document
	doc := p.Parse(content)
	result.StatementCount = len(doc.Statements)

	// Run diagnostics
	diags := dp.AnalyzeAST(doc)
	result.Diagnostics = diags

	// Count by severity
	for _, diag := range diags {
		switch diag.Severity {
		case 1: // ERROR
			result.ErrorCount++
		case 2: // WARNING
			result.WarningCount++
		case 3: // INFO
			result.InfoCount++
		}
	}

	// Validate against expectations
	result.Passed = validateExpectations(&result)

	return result
}

// parseExpectations extracts EXPECT: annotations from comments
// Supported formats:
//   /* EXPECT: NONE */
//   /* EXPECT: ERROR - pattern */
//   /* EXPECT: WARNING - pattern */
//   /* EXPECT: INFO - pattern */
func parseExpectations(content string) []Expectation {
	var expectations []Expectation
	lines := strings.Split(content, "\n")

	// Regex to match: /* EXPECT: SEVERITY - optional message pattern */
	expectRegex := regexp.MustCompile(`/\*\s*EXPECT:\s*(NONE|ERROR|WARNING|INFO)(?:\s*-\s*(.+?))?\s*\*/`)

	for i, line := range lines {
		matches := expectRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			exp := Expectation{
				Line:     i + 1, // Line numbers are 1-based
				Severity: matches[1],
			}
			if len(matches) > 2 {
				exp.Pattern = strings.TrimSpace(matches[2])
			}
			expectations = append(expectations, exp)
		}
	}

	return expectations
}

// validateExpectations checks if actual diagnostics match expectations
func validateExpectations(result *TestResult) bool {
	// If no expectations defined, use old behavior (test- prefix should have no errors)
	if len(result.Expectations) == 0 {
		if strings.HasPrefix(result.Name, "test-") {
			return result.ErrorCount == 0
		}
		return true
	}

	// Validate each expectation
	result.Mismatches = []string{}
	allMatched := true

	for _, exp := range result.Expectations {
		matched := false

		// Find diagnostics on this line
		var diagsOnLine []lsp.Diagnostic
		for _, diag := range result.Diagnostics {
			// Diagnostics use 0-based line numbers, expectations use 1-based
			if diag.Range.Start.Line+1 == exp.Line {
				diagsOnLine = append(diagsOnLine, diag)
			}
		}

		if exp.Severity == "NONE" {
			// Expect no diagnostics on this line
			if len(diagsOnLine) > 0 {
				result.Mismatches = append(result.Mismatches,
					fmt.Sprintf("Line %d: Expected NONE, but got %d diagnostic(s)", exp.Line, len(diagsOnLine)))
				allMatched = false
			} else {
				matched = true
			}
		} else {
			// Expect specific severity
			expectedSeverity := severityStringToInt(exp.Severity)

			for _, diag := range diagsOnLine {
				if diag.Severity == expectedSeverity {
					// Check pattern if specified
					if exp.Pattern != "" {
						patternRegex, err := regexp.Compile(exp.Pattern)
						if err == nil && patternRegex.MatchString(diag.Message) {
							matched = true
							break
						}
					} else {
						matched = true
						break
					}
				}
			}

			if !matched {
				if exp.Pattern != "" {
					result.Mismatches = append(result.Mismatches,
						fmt.Sprintf("Line %d: Expected %s matching '%s', but not found", exp.Line, exp.Severity, exp.Pattern))
				} else {
					result.Mismatches = append(result.Mismatches,
						fmt.Sprintf("Line %d: Expected %s, but got %d diagnostic(s) with wrong severity",
							exp.Line, exp.Severity, len(diagsOnLine)))
				}
				allMatched = false
			}
		}
	}

	return allMatched
}

func severityStringToInt(severity string) int {
	switch severity {
	case "ERROR":
		return 1
	case "WARNING":
		return 2
	case "INFO":
		return 3
	default:
		return 0
	}
}

func printHeader(testDir string, fileCount int) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           SMPE Language Server - Test Suite                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("📁 Test directory: %s\n", testDir)
	fmt.Printf("📄 Test files:     %d\n", fileCount)
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Println()
}

func printResults(results []TestResult, totalPassed, totalFailed int) {
	fmt.Println()
	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Println("                         TEST RESULTS                            ")
	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Group results by status
	var passed []TestResult
	var failed []TestResult

	for _, r := range results {
		if r.Passed {
			passed = append(passed, r)
		} else {
			failed = append(failed, r)
		}
	}

	// Print passed tests (compact)
	if len(passed) > 0 {
		fmt.Println("✅ PASSED TESTS")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for _, r := range passed {
			fmt.Printf("  ✓ %-30s  Statements: %2d  Errors: %d  Warnings: %d\n",
				r.Name, r.StatementCount, r.ErrorCount, r.WarningCount)
		}
		fmt.Println()
	}

	// Print failed tests (detailed)
	if len(failed) > 0 {
		fmt.Println("❌ FAILED TESTS (Detailed)")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for i, r := range failed {
			if i > 0 {
				fmt.Println()
			}
			printDetailedTestResult(r)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Println("                           SUMMARY                               ")
	fmt.Println("═════════════════════════════════════════════════════════════════")
	total := len(results)
	passRate := float64(totalPassed) / float64(total) * 100

	fmt.Printf("Total Tests:    %3d\n", total)
	fmt.Printf("Passed:         %3d  ✅\n", totalPassed)
	fmt.Printf("Failed:         %3d  ❌\n", totalFailed)
	fmt.Printf("Pass Rate:      %.1f%%\n", passRate)
	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Println()
}

func printDetailedTestResult(r TestResult) {
	fmt.Printf("  ⚠️  %s\n", r.Name)
	fmt.Printf("      ├─ Statements parsed: %d\n", r.StatementCount)
	fmt.Printf("      ├─ Expectations:     %d\n", len(r.Expectations))
	fmt.Printf("      ├─ Errors:   %d\n", r.ErrorCount)
	fmt.Printf("      ├─ Warnings: %d\n", r.WarningCount)
	fmt.Printf("      ├─ Info:     %d\n", r.InfoCount)

	// Print expectation mismatches first (most important)
	if len(r.Mismatches) > 0 {
		fmt.Printf("      ├─ Expectation Failures:\n")
		for _, mismatch := range r.Mismatches {
			fmt.Printf("         ❌ %s\n", mismatch)
		}
	}

	if len(r.Diagnostics) > 0 {
		fmt.Printf("      └─ Diagnostics:\n")

		// Group diagnostics by severity
		var errors, warnings, infos []lsp.Diagnostic
		for _, diag := range r.Diagnostics {
			switch diag.Severity {
			case 1:
				errors = append(errors, diag)
			case 2:
				warnings = append(warnings, diag)
			case 3:
				infos = append(infos, diag)
			}
		}

		// Print errors first
		if len(errors) > 0 {
			for _, diag := range errors {
				msg := strings.TrimPrefix(diag.Message, "🔴 ")
				msg = strings.TrimPrefix(msg, "⚠️ ")
				fmt.Printf("         🔴 Line %3d: %s\n", diag.Range.Start.Line+1, msg)
			}
		}

		// Then warnings
		if len(warnings) > 0 {
			for _, diag := range warnings {
				msg := strings.TrimPrefix(diag.Message, "🔴 ")
				msg = strings.TrimPrefix(msg, "⚠️ ")
				fmt.Printf("         ⚠️  Line %3d: %s\n", diag.Range.Start.Line+1, msg)
			}
		}

		// Then infos
		if len(infos) > 0 {
			for _, diag := range infos {
				msg := strings.TrimPrefix(diag.Message, "🔴 ")
				msg = strings.TrimPrefix(msg, "⚠️ ")
				fmt.Printf("         ℹ️  Line %3d: %s\n", diag.Range.Start.Line+1, msg)
			}
		}
	}
}
