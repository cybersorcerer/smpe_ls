# Design: SMP/E Language Server Training Documentation

**Date:** 2026-04-09  
**Status:** Approved  
**Scope:** End-user training for the SMP/E Language Server VSCode Extension, smpe_lint, and smpe_outl

---

## 1. Goal

Create a modular, demonstrative training in Markdown format for end users of the SMP/E Language Server toolchain. The training is written in both German and English as parallel, independent documents.

---

## 2. Target Audience

- **SMP/E developers** who are familiar with SMP/E MCS syntax
- **New to VSCode** or the SMP/E Language Server extension
- No VSCode or Language Server prior knowledge assumed
- No exercises — the training is demonstrative (trainer-led or self-study)

---

## 3. Structure

Two parallel directory trees, one per language:

```
docs/training/
├── de/
│   ├── 00-index.md
│   ├── 01-installation.md
│   ├── 02-erste-schritte.md
│   ├── 03-syntax-highlighting.md
│   ├── 04-code-completion.md
│   ├── 05-diagnostics.md
│   ├── 06-hover-und-navigation.md
│   ├── 07-formatting.md
│   ├── 08-zosmf-integration.md
│   ├── 09-bonus-smpe-lint.md
│   └── 10-bonus-smpe-outl.md
└── en/
    ├── 00-index.md
    ├── 01-installation.md
    ├── 02-first-steps.md
    ├── 03-syntax-highlighting.md
    ├── 04-code-completion.md
    ├── 05-diagnostics.md
    ├── 06-hover-and-navigation.md
    ├── 07-formatting.md
    ├── 08-zosmf-integration.md
    ├── 09-bonus-smpe-lint.md
    └── 10-bonus-smpe-outl.md
```

---

## 4. Module Content

### Module 00 — Index (DE: `00-index.md` / EN: `00-index.md`)
- Overview of all modules with links
- Prerequisites (VSCode installed, VSIX available)
- Version reference (which version of the extension is covered)
- Link to GitHub releases page

### Module 01 — Installation (DE: `01-installation.md` / EN: `01-installation.md`)
- Download the correct VSIX for the platform (table: Windows x64/ARM64, macOS Apple Silicon/Intel, Linux x64/ARM64)
- Install via VSCode UI (`Extensions: Install from VSIX...`) and via terminal (`code --install-extension`)
- Verify installation: check Output panel "SMP/E Language Server"
- Key settings overview: `smpe.serverPath`, `smpe.dataPath`, `smpe.debug`

### Module 02 — First Steps (DE: `02-erste-schritte.md` / EN: `02-first-steps.md`)
- Open or create a `.smpe` file
- Automatic language detection (file extension `.smpe`, `.mcs`, `.smp`)
- Automatic detection for files without extension (first line starts with `++`)
- z/OS dataset detection via LLQ (`smpe.zosDatasetsLlq` setting)
- Column rulers at 72 and 80

### Module 03 — Syntax Highlighting (DE: `03-syntax-highlighting.md` / EN: `03-syntax-highlighting.md`)
- Color coding for MCS statements (`++USERMOD`, `++PTF`, etc.)
- Operands and their parameters
- Inline comments (`/* ... */`)
- Inline data blocks (e.g. after `++JCLIN`, `++MAC`)
- Code example: `++USERMOD` with multiple operands

### Module 04 — Code Completion (DE: `04-code-completion.md` / EN: `04-code-completion.md`)
- Trigger completion with `Ctrl+Space`
- Statement completion (type `++` to get all statements)
- Operand completion (context-sensitive: only valid operands for the current statement)
- Parameter completion where applicable
- Mutually exclusive operands
- Code example: completing a `++VER` statement step by step

### Module 05 — Diagnostics (DE: `05-diagnostics.md` / EN: `05-diagnostics.md`)
- Real-time validation as you type
- Severity levels: Error (🔴), Warning (⚠️), Info (ℹ️), Hint (💡)
- Overview of all configurable diagnostic checks:
  - Unknown statement / operand
  - Missing required operand
  - Missing parameter
  - Missing terminator (`.`)
  - Unbalanced parentheses
  - Missing inline data
  - Content beyond column 72
  - Mutually exclusive operands
  - Standalone comment between MCS statements
- How to disable specific diagnostics via settings
- Code example: a `++USERMOD` with several intentional errors

### Module 06 — Hover and Navigation (DE: `06-hover-und-navigation.md` / EN: `06-hover-and-navigation.md`)
- Hover over a statement → IBM documentation tooltip
- Hover over an operand → description and valid values
- Go to Definition (`F12` / `Cmd+Click`) for SYSMOD/FMID references
- Find All References (`Shift+F12`)
- Document Symbols: Outline panel (`Cmd+Shift+O`) — hierarchical view of all statements
- Workspace Symbols: search across all `.smpe` files (`Cmd+T`)
- Folding: collapse/expand statements and comments

### Module 07 — Formatting (DE: `07-formatting.md` / EN: `07-formatting.md`)
- Format document: `Shift+Alt+F` (Windows/Linux) / `Shift+Option+F` (macOS)
- What formatting does: one operand per line, continuation indent, terminator on own line
- Format on Save option (`smpe.formatting.formatOnSave`)
- Configurable settings:
  - `smpe.formatting.indentContinuation`
  - `smpe.formatting.oneOperandPerLine`
  - `smpe.formatting.wrapListsAfterN`
  - `smpe.formatting.moveLeadingComments`
- Code example: before and after formatting

### Module 08 — z/OSMF Integration (DE: `08-zosmf-integration.md` / EN: `08-zosmf-integration.md`)
- What is z/OSMF integration and what does it require
- Create config: `SMP/E: Create z/OSMF Config` command → `.smpe-zosmf.yaml`
- Config file lookup order (workspace → `~/.config/smpe_ls/`)
- Available commands:
  - `SMP/E: Query SYSMOD via z/OSMF`
  - `SMP/E: Query DDDEF via z/OSMF`
  - `SMP/E: List Zones via z/OSMF`
  - `SMP/E: Free Form CSI Query`
- CodeLens: inline query links on SYSMOD/FMID references
- USS directory browser and MVS dataset browser
- **Check Missing Input Members**: `SMP/E: Check Missing Input Members` command
  - What it checks: input member files referenced by MCS statements
  - `smpe.checkMissingInputMembers.searchFolders` setting
  - Result table: filter and sort by file, statement, member, found status

### Module 09 — Bonus: smpe_lint (DE: `09-bonus-smpe-lint.md` / EN: `09-bonus-smpe-lint.md`)
- What is `smpe_lint` and when to use it (CI/CD pipelines)
- Installation / path to binary
- Basic usage: `smpe_lint *.smpe`
- Markdown output (default) and JSON output (`--json`)
- `--warnings-as-errors` for strict mode
- `--disable` to skip specific checks
- Config file (`--init`, `--config`)
- Exit codes

### Module 10 — Bonus: smpe_outl (DE: `10-bonus-smpe-outl.md` / EN: `10-bonus-smpe-outl.md`)
- What is `smpe_outl` and when to use it (inventory, reporting)
- Installation / path to binary
- Basic usage: `smpe_outl *.smpe` (Markdown output)
- JSON output (`--json`) — DocumentSymbol-compatible
- `--meta` flag: includes `hasInlineData` per statement
- `--ranges` flag: includes LSP range information
- Example: using JSON output in a script

---

## 5. Document Format per Module

Each module follows this structure:

```markdown
# Module N — Title

## Overview
One paragraph: what this module covers, why it matters.

## Prerequisites
What the user needs before starting this module (e.g. "Extension installed — see Module 01").

## [Feature/Topic 1]
Short intro sentence.
[Code block or description]
Step-by-step demonstration narrative.

## [Feature/Topic 2]
...

## Summary
Bullet list of the key points covered in this module.
```

---

## 6. Conventions

- All code examples use realistic SMP/E content (not placeholder lorem ipsum)
- Settings are always shown as full dotted names (e.g. `smpe.formatting.enabled`)
- Keyboard shortcuts shown for both macOS and Windows/Linux
- Screenshots are not included (Markdown only — trainers use live demo)
- Version covered: **v0.9.5**

---

## 7. Delivery

- Location: `docs/training/de/` and `docs/training/en/`
- Format: Markdown (`.md`)
- No build step required — plain Markdown files
- 10 modules per language = 20 files + 2 index files = **22 files total**
