# Module 10 — Bonus: smpe_outl

## Overview

`smpe_outl` outputs the document structure (outline) of SMP/E files — as Markdown
for reports or as JSON for scripts and pipelines. With the `--meta` flag it can also
detect missing input members in CI/CD pipelines.

## Prerequisites

Download the `smpe_outl` binary for your platform from
[GitHub Releases](https://github.com/cybersorcerer/smpe_ls/releases/latest) and
place it in your PATH (e.g. `~/.local/bin/smpe_outl`).

## Basic Usage

```bash
# Output outline as Markdown
smpe_outl mymod.smpe

# Multiple files
smpe_outl *.smpe
```

**Markdown output:**

```markdown
# SMP/E Outline Report

## mymod.smpe

### ++USERMOD(LJS2024)
- `DESC(Configuration change)`
- `REWORK(20240315)`
- `PRE(UJ12345)`

### ++VER(Z038)
- `FMID(HBB7790)`
- `PRE(UJ12345)`
```

## JSON Output

```bash
smpe_outl --json *.smpe
```

```json
[
  {
    "file": "mymod.smpe",
    "symbols": [
      {
        "name": "++USERMOD(LJS2024)",
        "id": "LJS2024",
        "kind": 5,
        "children": [
          { "name": "DESC(Configuration change)", "id": "Configuration change", "kind": 14 },
          { "name": "PRE(UJ12345)", "id": "UJ12345", "kind": 14 }
        ]
      }
    ]
  }
]
```

The JSON format is compatible with the LSP `DocumentSymbol` structure.

## --meta Flag: Inline Data Information

```bash
smpe_outl --json --meta *.smpe
```

Adds the `hasInlineData` field to each statement:

```json
{
  "name": "++MAC(MYMACRO)",
  "id": "MYMACRO",
  "kind": 23,
  "hasInlineData": true
}
```

## --ranges Flag: LSP Position Information

```bash
smpe_outl --json --ranges *.smpe
```

Adds `range` and `selectionRange` (line/character) to each symbol — useful for
tools that need to access positions in the source text.

## Pipeline Use Case: Detecting Missing Input Members

With `--json --meta`, missing input members can be detected in CI/CD pipelines.
The following shell script illustrates the approach:

```bash
#!/bin/bash
# Find all PARM statements without inline data and check if the .parm file exists

smpe_outl --json --meta *.smpe | \
  jq -r '.[] | .file as $f | .symbols[] |
    select(.name | startswith("++PARM")) |
    select(.hasInlineData == false) |
    select((.children // []) | map(.name) | any(startswith("TXLIB(")) | not) |
    [$f, .id] | @tsv' | \
  while IFS=$'\t' read -r file member; do
    expected="customization/${member}.parm"
    if [ ! -f "$expected" ]; then
      echo "MISSING: $expected (referenced in $file)"
      exit_code=1
    fi
  done

exit ${exit_code:-0}
```

**Statement to file extension mapping:**

| Statement | Extension |
|-----------|-----------|
| `++PARM` | `.parm` |
| `++SRC` | `.hlasm` |
| `++MAC` | `.hlasm` |
| `++EXEC` | `.rexx` |
| `++MOD` | `.mod` |
| `++ZAP` | `.zap` |
| `++PROC` | `.jcl` |
| `++CLIST` | `.clist` |
| `++MSG` | `.msg` |
| `++HELP` | `.help` |

## Summary

- `smpe_outl` outputs the document structure as Markdown or JSON
- `--json` produces LSP DocumentSymbol-compatible JSON
- `--meta` adds `hasInlineData` per statement
- `--ranges` adds LSP position information
- Ideal for inventory, reports, and CI/CD pipelines
- Pipeline scripts can use `--json --meta` to detect missing input members
