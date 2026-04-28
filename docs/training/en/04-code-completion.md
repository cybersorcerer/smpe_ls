# Module 04 — Code Completion

## Overview

The language server provides context-sensitive completion for statements and operands.
Suggestions are always tailored to the current statement — only syntactically valid
operands are offered.

## Prerequisites

Extension installed — see [Module 01](01-installation.md).

## Statement Completion

Type `++` on an empty line and press `Ctrl+Space` (Windows/Linux) or `Cmd+Space`
(macOS). A list of all available MCS statements appears:

```
++APAR
++ASSIGN
++DELETE
++FUNCTION
++HOLD
++IF
++JCLIN
++MAC
...
```

Each entry shows a short description from the IBM documentation.

## Operand Completion

After opening a statement, only the operands valid for that statement are offered.
Example: after `++VER(`, completion offers:

```smpe
++VER(Z038)
    FMID(      ← Ctrl+Space here shows: FMID, PRE, REQ, SUP, ...
```

Operands already used are not suggested again (except list operands like `PRE`, `REQ`).

## Step by Step: Completing a ++VER Statement

How to build a complete `++VER` statement using completion:

1. New line, type `++VER(` — completion suggests known FMIDs
2. Select or type the FMID, close with `)` and press Enter
3. Type `FM` + `Ctrl+Space` → suggestion `FMID(`
4. Select `FMID(`, type the value, close with `)`
5. Type `PR` + `Ctrl+Space` → suggestion `PRE(`
6. Finish with `.` on a new line

Result:

```smpe
++VER(Z038)
    FMID(HBB7790)
    PRE(UJ12345
        UJ67890).
```

## Boilerplate Snippets

For all Control MCS statements, completion offers an additional snippet item.
It is recognizable by the `…` suffix in the label and the snippet icon:

```text
++PTF
++PTF …        ← Snippet
++USERMOD
++USERMOD …    ← Snippet
```

A snippet inserts a complete boilerplate template containing all required operands
with tab stops for each placeholder. Example for `++PTF …`:

```smpe
++PTF(UAnnnnn)
  DESC(description)
  REWORK(2026118)
  .
```

After selecting the snippet, the cursor jumps automatically to the first placeholder
(`UAnnnnn`). Press `Tab` to move to the next placeholder until all values are filled in.

The REWORK value is automatically pre-filled with the current date in `yyyyddd` format
(e.g. `2026118` for April 28, 2026).

## Mutually Exclusive Operands

Some operands are mutually exclusive (e.g. `SYSLIB` and `TXLIB` in `++PARM`).
Both are shown in completion — validation happens via Diagnostics (see Module 05).

## Summary

- `++` + `Ctrl+Space` shows all available statements
- Operand completion is context-sensitive — only valid operands are shown
- Already-used operands are not suggested again
- Completion works after inline data when `++` is typed
