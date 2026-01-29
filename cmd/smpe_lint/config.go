package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cybersorcerer/smpe_ls/internal/diagnostics"
	"gopkg.in/yaml.v3"
)

// DiagnosticCode represents the code for each diagnostic type
type DiagnosticCode string

const (
	// Syntax Errors
	DiagUnknownStatement      DiagnosticCode = "unknown_statement"
	DiagInvalidLanguageID     DiagnosticCode = "invalid_language_id"
	DiagUnbalancedParentheses DiagnosticCode = "unbalanced_parentheses"
	DiagMissingTerminator     DiagnosticCode = "missing_terminator"
	DiagMissingParameter      DiagnosticCode = "missing_parameter"
	DiagContentBeyondCol72    DiagnosticCode = "content_beyond_column_72"

	// Operand Errors
	DiagUnknownOperand         DiagnosticCode = "unknown_operand"
	DiagDuplicateOperand       DiagnosticCode = "duplicate_operand"
	DiagEmptyOperandParameter  DiagnosticCode = "empty_operand_parameter"
	DiagMissingRequiredOperand DiagnosticCode = "missing_required_operand"
	DiagDependencyViolation    DiagnosticCode = "dependency_violation"
	DiagMutuallyExclusive      DiagnosticCode = "mutually_exclusive"
	DiagRequiredGroup          DiagnosticCode = "required_group"

	// Sub-Operand Errors
	DiagUnknownSubOperand    DiagnosticCode = "unknown_sub_operand"
	DiagSubOperandValidation DiagnosticCode = "sub_operand_validation"

	// Structural Errors
	DiagMissingInlineData           DiagnosticCode = "missing_inline_data"
	DiagStandaloneCommentBetweenMCS DiagnosticCode = "standalone_comment_between_mcs"
)

// LintConfig holds the linter configuration
// All diagnostics default to true (enabled)
type LintConfig struct {
	// WarningsAsErrors treats all warnings as errors (exit code 1)
	WarningsAsErrors bool `yaml:"warnings_as_errors" json:"warnings_as_errors"`

	// Diagnostics maps diagnostic codes to enabled/disabled (true/false)
	Diagnostics map[DiagnosticCode]bool `yaml:"diagnostics" json:"diagnostics"`
}

// DefaultLintConfig returns a config with all diagnostics enabled
func DefaultLintConfig() *LintConfig {
	return &LintConfig{
		WarningsAsErrors: false,
		Diagnostics:      make(map[DiagnosticCode]bool),
	}
}

// LoadConfig loads configuration from a file
// Supports both YAML and JSON formats
func LoadConfig(path string) (*LintConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultLintConfig()

	// Try YAML first (also handles JSON since YAML is a superset)
	ext := filepath.Ext(path)
	if ext == ".json" {
		err = json.Unmarshal(data, config)
	} else {
		err = yaml.Unmarshal(data, config)
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}

// FindConfigFile looks for a config file in standard locations
func FindConfigFile() string {
	// Check current directory first
	candidates := []string{
		".smpe_lint.yaml",
		".smpe_lint.yml",
		".smpe_lint.json",
	}

	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	// Check home directory
	home, err := os.UserHomeDir()
	if err == nil {
		for _, name := range candidates {
			path := filepath.Join(home, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// IsEnabled returns true if the diagnostic is enabled
// Defaults to true if not explicitly configured
func (c *LintConfig) IsEnabled(code DiagnosticCode) bool {
	if enabled, ok := c.Diagnostics[code]; ok {
		return enabled
	}
	return true // Default: enabled
}

// ToDiagnosticsConfig converts LintConfig to diagnostics.Config
func (c *LintConfig) ToDiagnosticsConfig() *diagnostics.Config {
	cfg := diagnostics.DefaultConfig()

	// Map lint config to diagnostics config
	cfg.UnknownStatement = c.IsEnabled(DiagUnknownStatement)
	cfg.InvalidLanguageId = c.IsEnabled(DiagInvalidLanguageID)
	cfg.UnbalancedParentheses = c.IsEnabled(DiagUnbalancedParentheses)
	cfg.MissingTerminator = c.IsEnabled(DiagMissingTerminator)
	cfg.MissingParameter = c.IsEnabled(DiagMissingParameter)
	cfg.UnknownOperand = c.IsEnabled(DiagUnknownOperand)
	cfg.DuplicateOperand = c.IsEnabled(DiagDuplicateOperand)
	cfg.EmptyOperandParameter = c.IsEnabled(DiagEmptyOperandParameter)
	cfg.MissingRequiredOperand = c.IsEnabled(DiagMissingRequiredOperand)
	cfg.DependencyViolation = c.IsEnabled(DiagDependencyViolation)
	cfg.MutuallyExclusive = c.IsEnabled(DiagMutuallyExclusive)
	cfg.RequiredGroup = c.IsEnabled(DiagRequiredGroup)
	cfg.MissingInlineData = c.IsEnabled(DiagMissingInlineData)
	cfg.UnknownSubOperand = c.IsEnabled(DiagUnknownSubOperand)
	cfg.SubOperandValidation = c.IsEnabled(DiagSubOperandValidation)
	cfg.ContentBeyondColumn72 = c.IsEnabled(DiagContentBeyondCol72)
	cfg.StandaloneCommentBetweenMCS = c.IsEnabled(DiagStandaloneCommentBetweenMCS)

	return cfg
}
