# Module 04 — Code Completion

## Overview

The language server provides context-sensitive completion for statements and operands.
Suggestions are always tailored to the current statement — only syntactically valid
operands are offered.

## Prerequisites

Extension installed — see [Module 01](01-installation.md).

## Statement Completion

Type `++` on an empty line. The list of all available MCS statements opens while
you type and narrows down with every further character:

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

## When Completion Opens

Completion is triggered by `+` and by letters, upper or lower case. A space does
not trigger it: while indenting a continuation line or separating two operands
the list would only get in the way. It opens once a name is actually begun. To
see the statement list without typing, request completion explicitly with
`Ctrl+Space`.

Inside a comment completion stays quiet — that is prose, not MCS.

> On macOS `Ctrl+Space` is claimed by the system for switching the input source,
> so it may never reach the editor. Either free it up in System Settings →
> Keyboard → Keyboard Shortcuts → Input Sources, or simply type `+` — the list
> opens on its own.

## Operand Completion

After opening a statement, only the operands valid for that statement are offered.
Example: after `++VER(`, completion offers:

```smpe
++VER(Z038)
    FMID(      ← typing here shows: FMID, PRE, REQ, SUP, ...
```

Operands already used are not suggested again (except list operands like `PRE`, `REQ`).

## Step by Step: Completing a ++VER Statement

How to build a complete `++VER` statement using completion:

1. New line, type `++VER(` — completion suggests known FMIDs
2. Select or type the FMID, close with `)` and press Enter
3. Type `FM` → suggestion `FMID(`
4. Select `FMID(`, type the value, close with `)`
5. Type `PR` → suggestion `PRE(`
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

- Typing `++` shows all available statements; a space on an empty line does not
- Operand completion is context-sensitive — only valid operands are shown
- Already-used operands are not suggested again
- Completion works after inline data when `++` is typed
