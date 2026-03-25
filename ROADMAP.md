# SMP/E Language Server - Roadmap

This document lists potential improvements and extensions for the SMP/E Language Server,
prioritized by benefit and implementation effort.

---

## High Value, Moderate Effort

### 1. Rename (`textDocument/rename`)
Rename SYSMOD IDs and FMIDs across the entire document. All occurrences in `PRE`, `REQ`,
`SUP`, `FMID`, and `IF` operands are updated simultaneously. Very useful for large MCS files.

### 2. Inlay Hints (`textDocument/inlayHint`)
Show parameter labels directly in the editor, e.g. `FROMDS(` displays `DSN=` before the
value. Makes nested sub-operands significantly more readable.

### 3. Workspace Symbols (`workspace/symbol`)
Search for SYSMOD definitions across multiple `.smpe` files in a workspace. Useful for
projects that span many MCS files.

---

## Diagnostics Improvements

### 5. SYSMOD ID Format Validation
`++PTF(UA12345)` — the format `t + a + nnnnn` is not currently validated. Invalid IDs
will only surface on the mainframe.

### 6. Cross-Statement Consistency
`PRE`, `REQ`, and `SUP` operands reference SYSMOD IDs that may be defined in the same file.
Flag unknown references as a warning.

### 7. Numeric Parameter Format Validation
`REWORK` expects a date in format `yyyyddd`. Currently only the length is checked, not the
format itself.

### 8. Sub-Operand Required Field Validation
`FROMDS` requires `DSN` as a sub-operand. This is not currently validated by diagnostics.

---

## Data Quality (`smpe.json`)

### 9. Declarative `required` Flags in smpe.json
Required operands are currently hardcoded in `getRequiredOperands()` in
`internal/diagnostics/diagnostics_ast.go`. They should be declared directly in `smpe.json`
for maintainability and consistency.

### 10. Complete HFS and Data Element Operand Definitions
`++AIX1`–`++AIX5`, `++CLIENT1`–`++CLIENT5`, `++OS21`–`++OS25`, `++UNIX1`–`++UNIX5`,
`++WIN1`–`++WIN5` have incomplete operand definitions. Completion and hover are incomplete
for these statements.

---

## Higher Effort, High Value

### 11. Call Hierarchy
Visualize the SYSMOD dependency chain: which SYSMODs depend on this one, and what does
this one require? Corresponds to `SMP/E REPORT` on the mainframe.

### 12. Code Actions
Quick fixes triggered by diagnostics:
- Insert missing statement terminator `.`
- Remove conflicting mutually-exclusive operand
- Add missing required operand (e.g. `DISTLIB`)
- Convert standalone comment to inline comment

---

## Status

| Feature | Priority | Status |
|---------|----------|--------|
| Rename | High | Planned |
| Folding Ranges | High | Planned |
| Inlay Hints | Medium | Planned |
| Workspace Symbols | Medium | Planned |
| SYSMOD ID Format Validation | High | Planned |
| Cross-Statement Consistency | Medium | Planned |
| Numeric Parameter Format | Low | Planned |
| Sub-Operand Required Validation | Medium | Planned |
| Declarative `required` in smpe.json | Medium | Planned |
| Complete HFS Operand Definitions | Medium | Planned |
| Call Hierarchy | Low | Planned |
| Code Actions | High | Planned |
