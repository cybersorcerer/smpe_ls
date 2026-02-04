# smpe_lint

A command-line linter for SMP/E MCS (Modification Control Statements) files.

## Overview

`smpe_lint` analyzes SMP/E MCS files and reports diagnostics including syntax errors, missing operands, invalid values, and structural issues. It's designed to be integrated into CI/CD pipelines to validate SMP/E packages before deployment.

## Installation

### From Source

```bash
cd cmd/smpe_lint
go build -o smpe_lint .
```

### Prerequisites

The linter requires `smpe.json` to be present at:

```bash
~/.local/share/smpe_ls/smpe.json
```

This file contains the SMP/E statement definitions and is shared with the smpe_ls language server.

## Usage

### Basic Usage

```bash
# Lint a single file
smpe_lint mypackage.smpe

# Lint multiple files using glob pattern
smpe_lint "*.smpe"
smpe_lint "packages/**/*.smpe"
```

### Output Formats

```bash
# Default: Markdown output (good for GitLab/GitHub CI reports)
smpe_lint *.smpe

# JSON output (good for programmatic processing)
smpe_lint --json *.smpe
```

### Exit Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success - no errors found |
| 1 | Failure - errors found or invalid arguments |

By default, only **errors** cause exit code 1. Warnings are reported but don't cause failure.

### Command-Line Options

```text
Usage: smpe_lint [options] <file-pattern>

Options:
  --config <path>       Path to configuration file (.smpe_lint.yaml or .smpe_lint.json)
  --disable <code>      Disable specific diagnostic (can be used multiple times)
  --init <format>       Create sample config file (yaml or json)
  --json                Output results in JSON format
  --version, -v         Show version information
  --warnings-as-errors  Treat warnings as errors (exit code 1)
```

### Examples

```bash
# Create a sample configuration file
smpe_lint --init yaml
smpe_lint --init json

# Treat all warnings as errors (strict mode for CI)
smpe_lint --warnings-as-errors *.smpe

# Disable specific diagnostics
smpe_lint --disable unknown_operand --disable duplicate_operand *.smpe

# Use a configuration file
smpe_lint --config .smpe_lint.yaml *.smpe
```

## Configuration File

Create a `.smpe_lint.yaml` (or `.smpe_lint.json`) file in your project root or home directory:

### YAML Format

```yaml
# .smpe_lint.yaml

# Treat all warnings as errors
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
  duplicate_operand: false        # Disable this diagnostic
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
```

### JSON Format

```json
{
  "warnings_as_errors": false,
  "diagnostics": {
    "unknown_statement": true,
    "duplicate_operand": false,
    "unknown_operand": true
  }
}
```

### Configuration Lookup Order

1. Path specified via `--config` flag
2. `.smpe_lint.yaml` in current directory
3. `.smpe_lint.yml` in current directory
4. `.smpe_lint.json` in current directory
5. `.smpe_lint.yaml` in home directory
6. `.smpe_lint.yml` in home directory
7. `.smpe_lint.json` in home directory

## Diagnostic Codes

### Syntax Errors

| Code | Description | Default Severity |
|------|-------------|------------------|
| `unknown_statement` | Unrecognized MCS statement type | Error |
| `invalid_language_id` | Invalid language identifier suffix | Error |
| `unbalanced_parentheses` | Missing opening or closing parenthesis | Error |
| `missing_terminator` | Statement not terminated with `.` | Error |
| `missing_parameter` | Required statement parameter missing | Error |
| `content_beyond_column_72` | Content extends past column 72 | Error |

### Operand Errors

| Code | Description | Default Severity |
|------|-------------|------------------|
| `unknown_operand` | Operand not valid for this statement | Warning |
| `duplicate_operand` | Same operand specified multiple times | Hint |
| `empty_operand_parameter` | Operand requires a parameter value | Error |
| `missing_required_operand` | Required operand not specified | Warning |
| `dependency_violation` | Operand requires another operand | Info |
| `mutually_exclusive` | Conflicting operands specified | Error |
| `required_group` | One of a group of operands required | Error |

### Sub-Operand Errors

| Code | Description | Default Severity |
|------|-------------|------------------|
| `unknown_sub_operand` | Sub-operand not valid | Warning |
| `sub_operand_validation` | Sub-operand value validation failed | Warning |

### Structural Errors

| Code | Description | Default Severity |
|------|-------------|------------------|
| `missing_inline_data` | Statement expects inline data | Warning |
| `standalone_comment_between_mcs` | Comment between MCS statements | Error |

## CI/CD Integration

### GitLab CI

```yaml
smpe-lint:
  stage: validate
  script:
    - smpe_lint --warnings-as-errors "packages/**/*.smpe" > lint-report.md
  artifacts:
    paths:
      - lint-report.md
    when: always
```

### GitHub Actions

```yaml
- name: Lint SMP/E Files
  run: |
    smpe_lint --json packages/*.smpe > lint-results.json
    smpe_lint packages/*.smpe
```

### Jenkins

```groovy
stage('SMP/E Lint') {
    steps {
        sh 'smpe_lint --warnings-as-errors packages/*.smpe'
    }
}
```

## Output Examples

### Markdown Output (Default)

```markdown
# SMP/E Lint Report

## File: `mypackage.smpe`
- 🔴 **ERROR** `missing_terminator` (Line 5, Col 1): Statement must be terminated with '.'
- ⚠️ **WARNING** `unknown_operand` (Line 8, Col 4): Unknown operand 'FOOBAR' for statement ++USERMOD

## Summary
- **Files checked**: 3
- **Files with issues**: 1
- **Total Errors**: 1
- **Total Warnings**: 1
- **Result**: 🔴 FAILURE
```

### JSON Output

```json
{
  "summary": {
    "total_files": 3,
    "files_with_issues": 1,
    "total_errors": 1,
    "total_warnings": 1,
    "success": false
  },
  "files": [
    {
      "path": "mypackage.smpe",
      "status": "failure",
      "diagnostics": [
        {
          "line": 5,
          "column": 1,
          "severity": "ERROR",
          "code": "missing_terminator",
          "message": "Statement must be terminated with '.'"
        }
      ]
    }
  ]
}
```

## License

Copyright (c) 2025, 2026 Sir Tobi aka Cybersorcerer

See the main project LICENSE file for details.
