# Module 02 — First Steps

## Overview

This module shows how to open or create an SMP/E file, how automatic language
detection works, and what visual aids the editor provides.

## Prerequisites

Extension installed — see [Module 01](01-installation.md).

## Opening or Creating a .smpe File

Open an existing file with the extension `.smpe`, `.mcs`, or `.smp`, or create a
new file and save it with one of these extensions.

The extension activates automatically — the status bar at the bottom shows `SMP/E`.

## Automatic Language Detection

The extension recognizes SMP/E files in three ways:

**1. File extension** — `.smpe`, `.mcs`, `.smp` are always treated as SMP/E.

**2. First line** — Files without an extension are detected when the first non-empty
line starts with `++`:

```smpe
++USERMOD(MYMOD001)
    DESC(My first change).
```

**3. z/OS Datasets** — Datasets with configurable Last Level Qualifiers (LLQ) are
automatically activated. The default LLQ is `MCS`:

```json
"smpe.zosDatasetsLlq": ["MCS", "SMPE"]
```

A dataset like `SYS1.MYPKG.MCS` is automatically recognized as SMP/E.

## Column Rulers

The editor automatically shows visual rulers at columns 72 and 80 — the IBM card
boundaries for SMP/E syntax:

- **Column 72** — Content beyond this is ignored by SMP/E (shown in grey)
- **Column 80** — End of the card line (shown in dark grey)

The rulers can be disabled via `smpe.editor.showColumnRulers`.

## Word Wrap

Word wrap is disabled by default for SMP/E files (`editor.wordWrap: off`) because
column position is syntactically significant.

## Summary

- File extensions `.smpe`, `.mcs`, `.smp` activate the extension automatically
- Files without extension are detected by their first line (`++`)
- z/OS datasets are detected via configurable LLQs
- Column rulers at 72 and 80 are active by default
