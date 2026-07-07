# Module 09 — Bonus: smpe_lint

## Overview

`smpe_lint` is a command-line tool that applies the same diagnostics as the language
server — ideal for CI/CD pipelines and automated quality checks.

## Prerequisites

Download the `smpe_lint` binary for your platform from
[GitHub Releases](https://github.com/cybersorcerer/smpe_ls/releases/latest) and
place it in your PATH (e.g. `~/.local/bin/smpe_lint`).

## Basic Usage

```bash
# Check one or more files
smpe_lint mymod.smpe

# With wildcard
smpe_lint *.smpe

# Show version
smpe_lint --version
```

## Output Formats

**Markdown (default):**

```
# SMP/E Lint Report

## mymod.smpe

🔴 **ERROR** Line 3: Unknown statement '++USRMOD'
⚠️ **WARNING** Line 7: Missing required operand: FMID

---
2 issues found (1 error, 1 warning)
```

**JSON (`--json`):**

```bash
smpe_lint --json *.smpe
```

```json
[
  {
    "file": "mymod.smpe",
    "issues": [
      {
        "line": 3,
        "severity": "error",
        "code": "unknown_statement",
        "message": "Unknown statement '++USRMOD'"
      }
    ]
  }
]
```

## Strict Mode

```bash
# Treat warnings as errors (exit code 1 also for warnings)
smpe_lint --warnings-as-errors *.smpe
```

## Disabling Specific Checks

```bash
smpe_lint --disable missing_inline_data --disable content_beyond_column72 *.smpe
```

## Configuration File

```bash
# Generate a sample configuration
smpe_lint --init

# Use a configuration file
smpe_lint --config .smpe_lint.yaml *.smpe
```

Example `.smpe_lint.yaml`:

```yaml
disable:
  - missing_inline_data
  - content_beyond_column72
warnings_as_errors: false
```

## Path to smpe.json (`--data`)

By default `smpe_lint` reads the statement definitions from
`~/.local/share/smpe_ls/smpe.json`. In environments without a usable
home directory (Docker containers, CI runners), pass the path explicitly:

```bash
smpe_lint --data data/smpe.json *.smpe
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No issues found |
| `1` | Issues found (or warnings with `--warnings-as-errors`) |
| `2` | Tool error (file not found, invalid arguments) |

## Use in CI/CD Pipelines

```yaml
# GitHub Actions example
- name: Lint SMP/E files
  run: |
    smpe_lint --data data/smpe.json --warnings-as-errors *.smpe
```

## Summary

- `smpe_lint` checks SMP/E files on the command line
- Markdown and JSON output for different use cases
- `--warnings-as-errors` for strict pipelines
- `--disable` to suppress individual checks
- `--data` to set the smpe.json path in containers/CI
- Exit code `0` = no issues, `1` = issues found
