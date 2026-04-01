# smpe_outl

A command-line outline tool for SMP/E MCS (Modification Control Statements) files.

## Overview

`smpe_outl` parses SMP/E MCS files and prints their document outline — the hierarchy of
MCS statements and their operands. It is designed to be integrated into CI/CD pipelines
for inventory reporting, change tracking, and programmatic processing of SMP/E packages.

## Installation

### From Source

```bash
cd cmd/smpe_outl
go build -o smpe_outl .
```

### Prerequisites

The tool requires `smpe.json` to be present at:

```bash
~/.local/share/smpe_ls/smpe.json
```

This file contains the SMP/E statement definitions and is shared with the smpe_ls language
server and smpe_lint.

## Usage

### Basic Usage

```bash
# Outline a single file
smpe_outl mypackage.smpe

# Outline multiple files using glob pattern
smpe_outl "*.smpe"
smpe_outl "packages/**/*.smpe"
```

### Output Formats

```bash
# Default: Markdown output
smpe_outl *.smpe

# JSON output (DocumentSymbol-compatible)
smpe_outl --json *.smpe

# JSON output with LSP range information
smpe_outl --json --ranges *.smpe
```

### Exit Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | Failure — file not found, unreadable, or invalid arguments |

### Command-Line Options

```text
Usage: smpe_outl [options] <file-pattern>

Options:
  --data <path>   Path to smpe.json data file
  --json          Output results in JSON format
  --ranges        Include range and selectionRange in JSON output
  --version, -v   Show version information
```

## Output Examples

### Markdown Output (Default)

```markdown
# SMP/E Outline Report

## mypackage.smpe

### ++USERMOD(LJS2012)
- `VER(Z038)`
- `PRE(UV12345)`
- `SUP(OY12345)`

### ++VER(Z038)
- `FMID(HBAS110)`
- `DISTLIB(SMPMTS)`
- `TXLIB(SMPPTS)`
```

### JSON Output (Default — no ranges)

```json
[
  {
    "file": "mypackage.smpe",
    "symbols": [
      {
        "name": "++USERMOD(LJS2012)",
        "id": "LJS2012",
        "kind": 5,
        "children": [
          { "name": "VER(Z038)",    "id": "Z038",    "kind": 7 },
          { "name": "PRE(UV12345)", "id": "UV12345", "kind": 7 },
          { "name": "SUP(OY12345)", "id": "OY12345", "kind": 7 }
        ]
      },
      {
        "name": "++VER(Z038)",
        "id": "Z038",
        "kind": 6,
        "children": [
          { "name": "FMID(HBAS110)",   "id": "HBAS110", "kind": 7 },
          { "name": "DISTLIB(SMPMTS)", "id": "SMPMTS",  "kind": 7 },
          { "name": "TXLIB(SMPPTS)",   "id": "SMPPTS",  "kind": 7 }
        ]
      }
    ]
  }
]
```

### JSON Output with `--ranges`

Adds LSP-compatible `range` and `selectionRange` fields to every symbol:

```json
{
  "name": "++USERMOD(LJS2012)",
  "id": "LJS2012",
  "kind": 5,
  "range": {
    "start": { "line": 0, "character": 0 },
    "end":   { "line": 10, "character": 1 }
  },
  "selectionRange": {
    "start": { "line": 0, "character": 0 },
    "end":   { "line": 0, "character": 18 }
  },
  "children": [...]
}
```

## Symbol Kinds

The `kind` field uses LSP `SymbolKind` values:

| Kind | Value | Statement |
|------|-------|-----------|
| Class | 5 | `++FUNCTION`, `++USERMOD`, `++PTF`, `++APAR` |
| Method | 6 | `++VER` |
| Property | 7 | Operands (children) |
| Struct | 23 | `++MAC`, `++SRC`, `++MOD` |
| Event | 24 | `++MACUPD`, `++SRCUPD` |
| File | 17 | `++JCLIN` |
| Function | 12 | All other statements |
| Operator | 25 | `++IF` |

## CI/CD Integration

### GitLab CI

```yaml
smpe-outline:
  stage: validate
  script:
    - smpe_outl "packages/**/*.smpe" > outline-report.md
    - smpe_outl --json "packages/**/*.smpe" > outline.json
  artifacts:
    paths:
      - outline-report.md
      - outline.json
    when: always
```

### GitHub Actions

```yaml
- name: Generate SMP/E Outline
  run: |
    smpe_outl --json packages/*.smpe > outline.json
```

### Jenkins

```groovy
stage('SMP/E Outline') {
    steps {
        sh 'smpe_outl --json packages/*.smpe > outline.json'
        archiveArtifacts artifacts: 'outline.json'
    }
}
```

## License

Copyright (c) 2025, 2026 Sir Tobi aka Cybersorcerer

See the main project LICENSE file for details.
