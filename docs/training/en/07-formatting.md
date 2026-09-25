# Module 07 — Formatting

## Overview

The language server can automatically format SMP/E documents: operands are placed on
individual lines, indentation is normalized, and the terminator is placed on its own
line.

## Prerequisites

Extension installed — see [Module 01](01-installation.md).

## Formatting a Document

Keyboard shortcut:
- macOS: `Shift+Option+F`
- Windows/Linux: `Shift+Alt+F`

Or: `Cmd+Shift+P` → `Format Document`

## What Formatting Does

**Before:**

```smpe
++USERMOD(LJS2024) DESC(Configuration change) REWORK(20240315) PRE(UJ12345 UJ67890).
```

**After:**

```smpe
++USERMOD(LJS2024)
    DESC(Configuration change)
    REWORK(20240315)
    PRE(UJ12345
        UJ67890).
```

- Each operand on its own line
- Continuation lines with configurable indentation (default: 4 spaces)
- Terminator `.` on its own line at the beginning
- Lists (PRE, REQ, SUP) are wrapped after N items

## Format on Save

Enable automatic formatting on save:

```json
{
  "smpe.formatting.formatOnSave": true
}
```

## Comments Close Themselves

Typing `/*` followed by a space gets you the closing `*/` as well — on the same
line, with the cursor left in between:

```smpe
++USERMOD(U1) /* │ */
```

You can write on without having to add the end yourself. Three cases are left
alone:

- The `*/` would come to rest beyond column 72. SMP/E stops reading there, so a
  marker out in that area would only look as if it closed the comment.
- The line already carries a `*/` further to the right.
- The line belongs to inline data. A `/*` opens a REXX program there and
  belongs to the element, not to the MCS syntax.

The insertion depends on `editor.formatOnType`, which the extension turns on
for `.smpe` by default. Turning it off in your own settings turns off the
closing as well.

## Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.formatting.enabled` | `true` | Enable formatting |
| `smpe.formatting.indentContinuation` | `4` | Indentation for continuation lines |
| `smpe.formatting.oneOperandPerLine` | `true` | Place each operand on its own line |
| `smpe.formatting.wrapListsAfterN` | `2` | Wrap lists after N items (0 = never) |
| `smpe.formatting.formatOnSave` | `false` | Automatically format when saving |
| `smpe.formatting.moveLeadingComments` | `true` | Move comments before first statement into the statement |

## Note on Inline Data

Statements with inline data (e.g. `++JCLIN`, `++MAC`) are not modified by formatting
— the inline data block is left untouched.

## Summary

- `Shift+Option+F` (macOS) / `Shift+Alt+F` (Windows/Linux) formats the document
- Operands are placed on individual lines, terminator on its own line
- Formatting is configurable via several settings
- `formatOnSave` enables automatic formatting on save
- Inline data blocks are not modified
