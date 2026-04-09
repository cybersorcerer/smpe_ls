# Module 06 — Hover and Navigation

## Overview

The language server displays IBM reference documentation when hovering over statements
and operands. It also enables navigation to definitions and references, and provides
an overview of all symbols in the document and workspace.

## Prerequisites

Extension installed — see [Module 01](01-installation.md).

## Hover Documentation

Move the mouse over a statement or operand — a tooltip shows the IBM documentation:

- **Statement hover** (`++USERMOD`): statement description, type, parameter info
- **Operand hover** (`DESC`): description, valid values, required or optional

Example: Hovering over `REWORK` shows:
> *REWORK — Specifies the date the SYSMOD was last reworked. Format: YYYYMMDD.*

## Go to Definition

Press `F12` or `Cmd+Click` (macOS) / `Ctrl+Click` (Windows/Linux) on a SYSMOD name
or FMID to jump to its definition within the same file:

```smpe
++VER(Z038)
    FMID(HBB7790)   ← Cmd+Click jumps to ++FUNCTION(HBB7790)
    PRE(UJ12345).   ← Cmd+Click jumps to ++USERMOD(UJ12345) or ++PTF(UJ12345)
```

## Find All References

`Shift+F12` on a SYSMOD name shows all places in the file where that SYSMOD is
referenced (e.g. in PRE, REQ, SUP, FMID).

## Document Symbols — Outline

`Cmd+Shift+O` (macOS) / `Ctrl+Shift+O` (Windows/Linux) opens the Outline view
with all statements in a hierarchy:

```
++USERMOD(LJS2024)
  ├── DESC(Configuration change)
  ├── REWORK(20240315)
  └── PRE(UJ12345)
++VER(Z038)
  ├── FMID(HBB7790)
  └── PRE(UJ12345)
```

Clicking an entry jumps directly to that location in the document.

## Workspace Symbols

`Cmd+T` (macOS) / `Ctrl+T` (Windows/Linux) searches all `.smpe` files in the
workspace for SYSMOD definitions:

- Type a SYSMOD name (e.g. `LJS2024`) to see all matches across all files
- Click a match to open the file and jump to the definition

## Folding

MCS statements and multi-line comments can be collapsed and expanded:

- Click the arrow to the left of the line number
- Or: `Cmd+K Cmd+[` / `Ctrl+K Ctrl+[` to collapse all

## Summary

- Hover shows IBM documentation for statements and operands
- `F12` / `Cmd+Click` jumps to definitions
- `Shift+F12` shows all references
- `Cmd+Shift+O` opens the outline of the current document
- `Cmd+T` searches all `.smpe` files in the workspace
- Statements and comments can be collapsed/expanded
