# Module 05 — Diagnostics

## Overview

The language server validates SMP/E syntax in real time and shows errors, warnings,
and hints directly in the editor. All checks can be individually enabled or disabled
via settings.

## Prerequisites

Extension installed — see [Module 01](01-installation.md).

## Severity Levels

| Symbol | Severity | Description |
|--------|----------|-------------|
| 🔴 | Error | Syntax error — SMP/E would reject the file |
| ⚠️ | Warning | Potential issues that may occur at runtime |
| ℹ️ | Information | Hints for better readability |
| 💡 | Hint | Optional improvement suggestions |

## Available Checks

### Unknown Statement / Unknown Operand

```smpe
++USRMOD(LJS2024)    ← 🔴 Unknown statement "++USRMOD"
    DESCRIPTN(Text). ← 🔴 Unknown operand "DESCRIPTN"
```

### Missing Required Operand

```smpe
++PTF(UJ12345)
    FMID(HBB7790).   ← ⚠️ Missing required operand: VER
```

### Missing Terminator

```smpe
++USERMOD(LJS2024)
    DESC(My change)   ← ⚠️ Missing statement terminator (.)
++VER(Z038)
```

### Unbalanced Parentheses

```smpe
++USERMOD(LJS2024
    DESC(My change).  ← 🔴 Missing closing parenthesis
```

### Missing Inline Data

```smpe
++MAC(MYMACRO)
    SYSLIB(S$SYSM1)
    DISTLIB(A$SYSM1). ← ⚠️ Missing inline data after statement
```

### Content Beyond Column 72

Characters after column 72 are ignored by SMP/E:

```smpe
++USERMOD(LJS2024)
    DESC(This is a very long description that extends beyond column 72 limit!!). ← ⚠️
```

### Mutually Exclusive Operands

```smpe
++PARM(RSSPRMI)
    SYSLIB(S$IQPARM)
    TXLIB(PARM).     ← ⚠️ SYSLIB and TXLIB are mutually exclusive
```

### Standalone Comment Between Statements

```smpe
++USERMOD(LJS2024)
    DESC(Change).
/* This comment causes an SMP/E syntax error */   ← 🔴
++VER(Z038)
```

## Configuring Diagnostics

All checks can be individually disabled in VSCode settings:

```json
{
  "smpe.diagnostics.missingInlineData": false,
  "smpe.diagnostics.contentBeyondColumn72": false
}
```

Full list of settings:

| Setting | Default |
|---------|---------|
| `smpe.diagnostics.unknownStatement` | `true` |
| `smpe.diagnostics.unknownOperand` | `true` |
| `smpe.diagnostics.missingRequiredOperand` | `true` |
| `smpe.diagnostics.missingParameter` | `true` |
| `smpe.diagnostics.missingTerminator` | `true` |
| `smpe.diagnostics.unbalancedParentheses` | `true` |
| `smpe.diagnostics.missingInlineData` | `true` |
| `smpe.diagnostics.contentBeyondColumn72` | `true` |
| `smpe.diagnostics.mutuallyExclusive` | `true` |
| `smpe.diagnostics.standaloneCommentBetweenMCS` | `true` |
| `smpe.diagnostics.duplicateOperand` | `true` |
| `smpe.diagnostics.emptyOperandParameter` | `true` |
| `smpe.diagnostics.dependencyViolation` | `true` |
| `smpe.diagnostics.requiredGroup` | `true` |
| `smpe.diagnostics.invalidLanguageId` | `true` |
| `smpe.diagnostics.unknownSubOperand` | `true` |
| `smpe.diagnostics.subOperandValidation` | `true` |

## Summary

- Diagnostics run in real time as you type
- Four severity levels: Error 🔴, Warning ⚠️, Info ℹ️, Hint 💡
- All checks are individually configurable
- Setting changes take effect immediately without restart
