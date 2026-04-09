# Module 03 — Syntax Highlighting

## Overview

The editor highlights all parts of an SMP/E file in different colors: statements,
operands, parameters, comments, and inline data. This makes it easy to grasp the
structure at a glance.

## Prerequisites

Extension installed — see [Module 01](01-installation.md).

## Statements

MCS statements such as `++USERMOD`, `++PTF`, `++FUNCTION`, or `++VER` are
highlighted in their own color:

```smpe
++USERMOD(LJS2024)
    DESC(Configuration parameter adjustment)
    REWORK(20240315)
    PRE(UJ12345).
```

## Operands and Parameters

Operands like `DESC`, `REWORK`, `PRE` and their parenthesized parameters are each
displayed in a distinct color.

## Comments

Inline comments in `/* ... */` format are shown in grey:

```smpe
++USERMOD(LJS2024)
    DESC(Configuration change) /* Change for customer XYZ */
    PRE(UJ12345).
```

## Statement Terminator

The terminator `.` at the end of each statement is highlighted separately.

## Inline Data

Statements that expect inline data (e.g. `++JCLIN`, `++MAC`, `++SRC`) show the
following data block in its own color:

```smpe
++JCLIN.
//SMPMCS   JOB (ACCT),'INSTALL',CLASS=A
//STEP1    EXEC PGM=IEFBR14
++JCLIN.
```

The JCL block between the two `++JCLIN.` statements is recognized as inline data
and colored differently from regular MCS syntax.

## Summary

- Statements, operands, parameters, comments, and terminators are visually distinguished
- Inline data (e.g. JCL after `++JCLIN`) is displayed separately
- Highlighting is language-specific and independent of the color theme
