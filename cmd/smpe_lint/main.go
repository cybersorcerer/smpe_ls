package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cybersorcerer/smpe_ls/internal/data"
	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"github.com/cybersorcerer/smpe_ls/internal/parser"
	"github.com/cybersorcerer/smpe_ls/pkg/lsp"
)

var (
	version = "v0.8.2-alpha"
	commit  = "unknown"
)

// JSON Report Structures
type Report struct {
	Summary Summary      `json:"summary"`
	Files   []FileReport `json:"files"`
}

type Summary struct {
	TotalFiles      int  `json:"total_files"`
	FilesWithIssues int  `json:"files_with_issues"`
	TotalErrors     int  `json:"total_errors"`
	TotalWarnings   int  `json:"total_warnings"`
	Success         bool `json:"success"`
}

type FileReport struct {
	Path        string           `json:"path"`
	Status      string           `json:"status"` // "success", "warning", "failure"
	Diagnostics []DiagnosticItem `json:"diagnostics"`
}

type DiagnosticItem struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message"`
}

func main() {
	// Flags
	jsonMode := flag.Bool("json", false, "Output results in JSON format")
	versionFlag := flag.Bool("version", false, "Show version information")
	shortVersionFlag := flag.Bool("v", false, "Show version information")
	configFile := flag.String("config", "", "Path to configuration file (.smpe_lint.yaml or .smpe_lint.json)")
	warningsAsErrors := flag.Bool("warnings-as-errors", false, "Treat warnings as errors (exit code 1)")
	initConfig := flag.String("init", "", "Create a sample configuration file (yaml or json)")
	var disableFlags arrayFlags
	flag.Var(&disableFlags, "disable", "Disable specific diagnostic (can be used multiple times)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file-pattern>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nLints SMP/E MCS files and reports diagnostics.\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --config <path>       Path to configuration file (.smpe_lint.yaml or .smpe_lint.json)\n")
		fmt.Fprintf(os.Stderr, "  --disable <code>      Disable specific diagnostic (can be used multiple times)\n")
		fmt.Fprintf(os.Stderr, "  --init <format>       Create sample config file (yaml or json)\n")
		fmt.Fprintf(os.Stderr, "  --json                Output results in JSON format\n")
		fmt.Fprintf(os.Stderr, "  --version, -v         Show version information\n")
		fmt.Fprintf(os.Stderr, "  --warnings-as-errors  Treat warnings as errors (exit code 1)\n")
		fmt.Fprintf(os.Stderr, "\nDiagnostic Codes:\n")
		fmt.Fprintf(os.Stderr, "  Syntax:\n")
		fmt.Fprintf(os.Stderr, "    unknown_statement, invalid_language_id, unbalanced_parentheses,\n")
		fmt.Fprintf(os.Stderr, "    missing_terminator, missing_parameter, content_beyond_column_72\n")
		fmt.Fprintf(os.Stderr, "  Operands:\n")
		fmt.Fprintf(os.Stderr, "    unknown_operand, duplicate_operand, empty_operand_parameter,\n")
		fmt.Fprintf(os.Stderr, "    missing_required_operand, dependency_violation, mutually_exclusive,\n")
		fmt.Fprintf(os.Stderr, "    required_group\n")
		fmt.Fprintf(os.Stderr, "  Sub-Operands:\n")
		fmt.Fprintf(os.Stderr, "    unknown_sub_operand, sub_operand_validation\n")
		fmt.Fprintf(os.Stderr, "  Structural:\n")
		fmt.Fprintf(os.Stderr, "    missing_inline_data, standalone_comment_between_mcs\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s *.smpe\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --warnings-as-errors *.smpe\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --disable unknown_operand *.smpe\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config .smpe_lint.yaml *.smpe\n", os.Args[0])
	}

	flag.Parse()

	// Handle version
	if *versionFlag || *shortVersionFlag {
		fmt.Printf("smpe_lint %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Copyright (c) 2025 Sir Tobi aka Cybersorcerer\n")
		os.Exit(0)
	}

	// Handle --init to create sample config
	if *initConfig != "" {
		if err := createSampleConfig(*initConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Verify args
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Collect files from all arguments
	// This handles both:
	// 1. Shell-expanded: smpe_lint file1.smpe file2.smpe file3.smpe
	// 2. Quoted glob: smpe_lint "*.smpe"
	var files []string
	for _, arg := range flag.Args() {
		// Try to expand as glob pattern
		matches, err := filepath.Glob(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding pattern '%s': %v\n", arg, err)
			os.Exit(1)
		}
		if len(matches) == 0 {
			// No matches - could be an explicit file that doesn't exist
			// or a pattern with no matches. Add it anyway to get proper error later.
			files = append(files, arg)
		} else {
			files = append(files, matches...)
		}
	}

	// Remove duplicates (in case shell expansion and glob both match same files)
	files = uniqueFiles(files)

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No files found matching arguments\n")
		os.Exit(1)
	}

	// Load configuration
	lintConfig := DefaultLintConfig()

	// Try to find and load config file
	configPath := *configFile
	if configPath == "" {
		configPath = FindConfigFile()
	}
	if configPath != "" {
		loadedConfig, err := LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error loading config from %s: %v\n", configPath, err)
		} else {
			lintConfig = loadedConfig
		}
	}

	// Apply command-line overrides
	if *warningsAsErrors {
		lintConfig.WarningsAsErrors = true
	}

	// Apply --disable flags
	for _, code := range disableFlags {
		if lintConfig.Diagnostics == nil {
			lintConfig.Diagnostics = make(map[DiagnosticCode]bool)
		}
		lintConfig.Diagnostics[DiagnosticCode(code)] = false
	}

	// Load smpe.json
	dataPath := os.Getenv("HOME") + "/.local/share/smpe_ls/smpe.json"
	store, err := data.Load(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading smpe.json from %s: %v\n", dataPath, err)
		os.Exit(1)
	}

	diagProvider := diagnostics.NewProvider(store)
	diagConfig := lintConfig.ToDiagnosticsConfig()

	report := Report{
		Files: []FileReport{},
	}
	report.Summary.TotalFiles = len(files)
	report.Summary.Success = true

	hasErrors := false

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
			report.Summary.TotalErrors++
			hasErrors = true
			continue
		}

		// Create parser and parse
		p := parser.NewParser(store.Statements)
		doc := p.Parse(string(content))

		// Analyze with config
		diags := diagProvider.AnalyzeASTWithConfigAndText(doc, diagConfig, string(content))

		fileReport := FileReport{
			Path:        file,
			Status:      "success",
			Diagnostics: []DiagnosticItem{},
		}

		hasFileIssues := false
		for _, d := range diags {
			// Determine diagnostic code from message
			code := determineDiagnosticCode(d.Message)

			// Check if this diagnostic is enabled
			if !lintConfig.IsEnabled(code) {
				continue
			}

			// Only process errors and warnings
			if d.Severity == lsp.SeverityError || d.Severity == lsp.SeverityWarning {
				hasFileIssues = true

				item := DiagnosticItem{
					Line:    d.Range.Start.Line + 1,
					Column:  d.Range.Start.Character + 1,
					Code:    string(code),
					Message: cleanMessage(d.Message),
				}

				if d.Severity == lsp.SeverityError {
					item.Severity = "ERROR"
					report.Summary.TotalErrors++
					if fileReport.Status != "failure" {
						fileReport.Status = "failure"
					}
					hasErrors = true
				} else {
					item.Severity = "WARNING"
					report.Summary.TotalWarnings++
					if fileReport.Status == "success" {
						fileReport.Status = "warning"
					}
					// Warnings cause failure only if --warnings-as-errors is set
					if lintConfig.WarningsAsErrors {
						hasErrors = true
					}
				}

				fileReport.Diagnostics = append(fileReport.Diagnostics, item)
			}
		}

		if hasFileIssues {
			report.Summary.FilesWithIssues++
			report.Files = append(report.Files, fileReport)
		}
	}

	report.Summary.Success = !hasErrors

	if *jsonMode {
		// JSON Output
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Markdown Output
		if len(report.Files) > 0 {
			fmt.Println("# SMP/E Lint Report")
			fmt.Println()
			for _, file := range report.Files {
				fmt.Printf("## File: `%s`\n", file.Path)
				for _, d := range file.Diagnostics {
					severityIcon := "🔴"
					if d.Severity == "WARNING" {
						severityIcon = "⚠️"
					}

					// Format: - 🔴 **ERROR** [code] (Line X, Col Y): Message
					if d.Code != "" {
						fmt.Printf("- %s **%s** `%s` (Line %d, Col %d): %s\n",
							severityIcon, d.Severity, d.Code, d.Line, d.Column, d.Message)
					} else {
						fmt.Printf("- %s **%s** (Line %d, Col %d): %s\n",
							severityIcon, d.Severity, d.Line, d.Column, d.Message)
					}
				}
				fmt.Println()
			}
		}

		// Summary Footer
		fmt.Println("## Summary")
		fmt.Printf("- **Files checked**: %d\n", report.Summary.TotalFiles)

		if report.Summary.Success {
			fmt.Printf("- **Result**: ✅ SUCCESS\n")
		} else {
			fmt.Printf("- **Files with issues**: %d\n", report.Summary.FilesWithIssues)
			fmt.Printf("- **Total Errors**: %d\n", report.Summary.TotalErrors)
			fmt.Printf("- **Total Warnings**: %d\n", report.Summary.TotalWarnings)
			fmt.Printf("- **Result**: 🔴 FAILURE\n")
		}
	}

	if hasErrors {
		os.Exit(1)
	}
	os.Exit(0)
}

// arrayFlags allows multiple flag values
type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ", ")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

// determineDiagnosticCode maps diagnostic messages to codes
func determineDiagnosticCode(message string) DiagnosticCode {
	msg := strings.ToLower(message)

	// Syntax errors
	if strings.Contains(msg, "unknown statement") {
		return DiagUnknownStatement
	}
	if strings.Contains(msg, "invalid language identifier") {
		return DiagInvalidLanguageID
	}
	if strings.Contains(msg, "missing closing parenthesis") || strings.Contains(msg, "missing opening parenthesis") {
		return DiagUnbalancedParentheses
	}
	if strings.Contains(msg, "must be terminated") {
		return DiagMissingTerminator
	}
	if strings.Contains(msg, "missing required parameter") {
		return DiagMissingParameter
	}
	if strings.Contains(msg, "beyond column 72") {
		return DiagContentBeyondCol72
	}

	// Operand errors
	if strings.Contains(msg, "unknown operand") {
		return DiagUnknownOperand
	}
	if strings.Contains(msg, "duplicate operand") {
		return DiagDuplicateOperand
	}
	if strings.Contains(msg, "requires a parameter") {
		return DiagEmptyOperandParameter
	}
	if strings.Contains(msg, "missing required operand") {
		return DiagMissingRequiredOperand
	}
	if strings.Contains(msg, "requires") && strings.Contains(msg, "to be specified") {
		return DiagDependencyViolation
	}
	if strings.Contains(msg, "mutually exclusive") {
		return DiagMutuallyExclusive
	}
	if strings.Contains(msg, "one of the following operands must be specified") {
		return DiagRequiredGroup
	}

	// Sub-operand errors
	if strings.Contains(msg, "unknown sub-operand") {
		return DiagUnknownSubOperand
	}
	if strings.Contains(msg, "sub-operand") {
		return DiagSubOperandValidation
	}

	// Structural errors
	if strings.Contains(msg, "expects inline data") {
		return DiagMissingInlineData
	}
	if strings.Contains(msg, "comment not allowed") {
		return DiagStandaloneCommentBetweenMCS
	}

	return ""
}

// cleanMessage removes emoji prefixes from diagnostic messages
func cleanMessage(message string) string {
	msg := message
	msg = strings.ReplaceAll(msg, "🔴 ", "")
	msg = strings.ReplaceAll(msg, "⚠️ ", "")
	msg = strings.ReplaceAll(msg, "ℹ️ ", "")
	msg = strings.ReplaceAll(msg, "💡 ", "")
	return msg
}

// uniqueFiles removes duplicate file paths from a slice
func uniqueFiles(files []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(files))
	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			result = append(result, file)
		}
	}
	return result
}

// createSampleConfig creates a sample configuration file
func createSampleConfig(format string) error {
	var filename string
	var content string

	switch strings.ToLower(format) {
	case "yaml", "yml":
		filename = ".smpe_lint.yaml"
		content = `# SMP/E Lint Configuration
# Generated by smpe_lint --init yaml

# Treat all warnings as errors (causes exit code 1)
warnings_as_errors: false

# Enable/disable individual diagnostics (true/false)
# All diagnostics are enabled by default
diagnostics:
  # Syntax Errors
  unknown_statement: true
  invalid_language_id: true
  unbalanced_parentheses: true
  missing_terminator: true
  missing_parameter: true
  content_beyond_column_72: true

  # Operand Validation
  unknown_operand: true
  duplicate_operand: true
  empty_operand_parameter: true
  missing_required_operand: true
  dependency_violation: true
  mutually_exclusive: true
  required_group: true

  # Sub-Operand Validation
  unknown_sub_operand: true
  sub_operand_validation: true

  # Structural Issues
  missing_inline_data: true
  standalone_comment_between_mcs: true
`
	case "json":
		filename = ".smpe_lint.json"
		content = `{
  "warnings_as_errors": false,
  "diagnostics": {
    "unknown_statement": true,
    "invalid_language_id": true,
    "unbalanced_parentheses": true,
    "missing_terminator": true,
    "missing_parameter": true,
    "content_beyond_column_72": true,
    "unknown_operand": true,
    "duplicate_operand": true,
    "empty_operand_parameter": true,
    "missing_required_operand": true,
    "dependency_violation": true,
    "mutually_exclusive": true,
    "required_group": true,
    "unknown_sub_operand": true,
    "sub_operand_validation": true,
    "missing_inline_data": true,
    "standalone_comment_between_mcs": true
  }
}
`
	default:
		return fmt.Errorf("unknown format '%s' (use 'yaml' or 'json')", format)
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("file '%s' already exists", filename)
	}

	// Write the file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("Created %s\n", filename)
	return nil
}
