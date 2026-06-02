# SMP/E Language Server

A modern Language Server Protocol (LSP) implementation for IBM SMP/E (System Modification Program/Extended) written in Go.

[![Version](https://img.shields.io/badge/version-1.3.3-blue.svg)](https://github.com/cybersorcerer/smpe_ls/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)

The project provides a Go-based language server plus a **VSCode extension** (*IBM z/OS SMP/E MCS Tools*) that delivers rich editing support for SMP/E MCS files: syntax highlighting, context-aware completion, real-time diagnostics, hover documentation, navigation, formatting, and integrated z/OSMF CSI queries. Two companion CLI tools (`smpe_lint`, `smpe_outl`) bring the same validation and outline capabilities to CI/CD pipelines.

## 🧩 VSCode Extension

The extension bundles the language server — no separate installation required.

**Install from the VSCode Marketplace:**

[**IBM z/OS SMP/E MCS Tools** on the Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=Cybersorcerer66.cybersorcerer66-smpe-mcs)

Or from within VSCode: open the Extensions view (`Cmd+Shift+X` / `Ctrl+Shift+X`), search for **SMP/E MCS**, and click **Install**.

Alternatively, install a platform-specific `.vsix` from the [releases page](https://github.com/cybersorcerer/smpe_ls/releases), or build from source (see [Quick Start](#-quick-start)).

## ✨ Features

- **🎨 Syntax Highlighting** - Color coding for MCS statements, operands, and comments
- **💡 Intelligent Code Completion** - Context-aware completion for statements and operands
- **🔍 Real-time Diagnostics** - Instant validation of SMP/E syntax and semantics
- **💡 Code Actions (Quick Fixes)** - One-click fixes for missing terminators, missing required operands, and empty `REWORK()` (current date)
- **🖋️ Signature Help** - Floating parameter hint for operands (`DISTLIB(│)`) with description and type, toggleable via `smpe.signatureHelp.enabled`
- **🔧 Command-Line Linter** - CI/CD-ready linter with configurable diagnostics (`smpe_lint`)
- **🗂️ Command-Line Outline** - Document outline tool for CI/CD inventory and reporting (`smpe_outl`)
- **📖 Hover Documentation** - Inline documentation from IBM SMP/E Reference
- **🔗 Go to Definition** - Navigate to SYSMOD/FMID definitions
- **🔎 Find References** - Find all references to a SYSMOD or FMID
- **📄 Document Symbols** - Outline view and quick navigation (`Cmd+Shift+O`)
- **🔍 Workspace Symbols** - Search for SYSMOD definitions across all `.smpe` files (`Cmd+T`)
- **📐 Folding Ranges** - Collapse/expand MCS statements and multi-line comments
- **📝 Document Formatting** - Auto-format SMP/E statements
- **🔭 CodeLens** - Inline z/OSMF CSI queries for SYSMODs and DDDEFs
- **🌐 z/OSMF Integration** - Query CSI, browse USS directories and MVS datasets via z/OSMF REST API
- **📥 Check Missing Input Members** - Scan the workspace for missing MCS input files, results shown in a filterable/sortable Webview table
- **💾 Saved & Reusable Queries** - Save, manage and reuse complete CSI queries plus a persistent filter history in the Free Form Query panel
- **📏 Column Rulers** - Visual guides at columns 72 and 80 for mainframe card boundaries
- **📚 Training Documentation** - 22 Markdown training modules (DE + EN) covering installation, all extension features, CLI tools, and z/OSMF integration (`SMP/E: Open Training`)
- **🌍 Multi-platform** - Native binaries for Linux, macOS, and Windows (AMD64 & ARM64)
- **⚡ Fast & Lightweight** - Written in Go with zero external dependencies

## 📦 Installation

### Platform-Specific Installation

Download pre-built binaries for your platform from the [latest release](https://github.com/cybersorcerer/smpe_ls/releases/latest).

**Available platforms:**

- Linux (AMD64, ARM64)
- macOS (Apple Silicon, Intel)
- Windows (AMD64, ARM64)

**Installation paths:**

- **Linux/macOS:** Binary in `/usr/local/bin/`, data in `~/.local/share/smpe_ls/`
- **Windows:** Binary and data in `%LOCALAPPDATA%\smpe_ls\`

📋 For detailed installation instructions, see [INSTALL.md](INSTALL.md)

### Build from Source

```bash
git clone https://github.com/cybersorcerer/smpe_ls.git
cd smpe_ls
make install
```

## 🚀 Quick Start

### VSCode Extension

1. Install the language server:

   ```bash
   make install
   ```

2. Build the VSCode extension:

   ```bash
   make vscode
   ```

3. Open `client/vscode-smpe` in VSCode and press F5

4. Create a `.smpe` file and start coding!

### Command Line

```bash
# Show version
smpe_ls --version

# Enable debug logging
smpe_ls --debug

# Use custom data file
smpe_ls --data /path/to/smpe.json
```

### Command-Line Linter

For CI/CD integration, use the `smpe_lint` tool:

```bash
# Lint SMP/E files
smpe_lint *.smpe

# Strict mode (warnings cause failure)
smpe_lint --warnings-as-errors *.smpe

# JSON output for programmatic processing
smpe_lint --json *.smpe

# Ignore specific diagnostics
smpe_lint --ignore unknown_operand *.smpe

# Use configuration file
smpe_lint --config .smpe_lint.yaml *.smpe
```

See [cmd/smpe_lint/README.md](cmd/smpe_lint/README.md) for full documentation.

### Command-Line Outline Tool

For CI/CD inventory and reporting, use the `smpe_outl` tool:

```bash
# Print outline as Markdown
smpe_outl *.smpe

# JSON output (DocumentSymbol-compatible)
smpe_outl --json *.smpe

# JSON output with LSP range information
smpe_outl --json --ranges *.smpe
```

See [cmd/smpe_outl/README.md](cmd/smpe_outl/README.md) for full documentation.

## 📝 Example

```smpe
/* Sample SMP/E MCS statements */
++APAR(AB12345)
    DESCRIPTION('Fix for security vulnerability')
    FILES(5)
    RFDSNPFX(APARA12)
    REWORK(20250101).

++FUNCTION(HBB7790)
    DESCRIPTION('Base function for product XYZ')
    FMID(HBB7780)
    VERSION(01.00.00).

++HOLD(AB12345)
    FMID(HBB7790)
    REASON(B12345)
    ERROR
    COMMENT('Critical security fix required').

/* Inline JCL data support */
++JCLIN.
//SMPMCS   JOB (ACCT),'INSTALL',CLASS=A
//STEP1    EXEC PGM=IEWL
//SYSLMOD  DD DSN=SYS1.LINKLIB,DISP=SHR
++JCLIN.
```

## 🎯 Supported MCS Statements

**Control MCS (25 statements with full diagnostics):**

| Statement | Description | Diagnostics |
|-----------|-------------|-------------|
| `++APAR` | Service SYSMOD (temporary fix) | ✅ |
| `++ASSIGN` | Source ID Assignment | ✅ |
| `++DELETE` | Delete Load Module | ✅ |
| `++FEATURE` | SYSMOD Set Description | ✅ |
| `++FUNCTION` | Function SYSMOD | ✅ |
| `++HOLD` | Exception Status | ✅ |
| `++IF` | Conditional Processing | ✅ |
| `++JAR` | JAR file management | ✅ |
| `++JARUPD` | JAR update operations | ✅ |
| `++JCLIN` | Job Control Language Input | ✅ |
| `++MAC` | Macro library management | ✅ |
| `++MACUPD` | Macro update operations | ✅ |
| `++MOD` | Load module operations | ✅ |
| `++MOVE` | Move module operations | ✅ |
| `++NULL` | Null SYSMOD | ✅ |
| `++PRODUCT` | Product definition | ✅ |
| `++PROGRAM` | Program/module definition | ✅ |
| `++PTF` | Program Temporary Fix | ✅ |
| `++RELEASE` | Release from hold status | ✅ |
| `++RENAME` | Rename operations | ✅ |
| `++SRC` | Source code operations | ✅ |
| `++SRCUPD` | Source update operations | ✅ |
| `++USERMOD` | User modification | ✅ |
| `++VER` | Version specification | ✅ |
| `++ZAP` | Superzap operations | ✅ |

**HFS MCS (27 statements with full diagnostics):**

All HFS-type statements share the same syntax and validation rules:

| Statement | Description | Diagnostics |
|-----------|-------------|-------------|
| `++HFS` | HFS file operations (supports 32 language variants: ++HFSxxx*) | ✅ |
| `++SHELLSCR` | Shell script operations | ✅ |
| `++AIX1` - `++AIX5` | AIX client file operations (5 variants) | ✅ |
| `++CLIENT1` - `++CLIENT5` | Generic client file operations (5 variants) | ✅ |
| `++OS21` - `++OS25` | OS/2 client file operations (5 variants) | ✅ |
| `++UNIX1` - `++UNIX5` | UNIX client file operations (5 variants) | ✅ |
| `++WIN1` - `++WIN5` | Windows client file operations (5 variants) | ✅ |

*`++HFS` can also be coded as `++HFSxxx` where xxx is one of 32 national language identifiers (e.g., `++HFSENU`, `++HFSDEU`, `++HFSJPN`). All language variants are supported in completion, diagnostics, and hover.

**Data Element MCS:**

- All data element statements with language variants (++BOOK, ++CLIST, ++EXEC, ++FONT, ++HELP, ++MSG, ++PARM, etc.)
- Completion and hover available for all statements

## 🧪 Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
make test-all

# Run unit tests only (completion, diagnostics, hover, parser)
make test

# Run central test suite (all .smpe test files)
make test-suite
```

**Test Coverage:**

- 57 unit tests across 4 modules
- 27 integration test files with 24 passing
- Tests for completion, diagnostics, hover, and parser

## 🏗️ Building for All Platforms

```bash
# Build for all platforms
make build-all

# Create release packages
make release

# Results in:
# - dist/smpe_ls-linux-amd64
# - dist/smpe_ls-linux-arm64
# - dist/smpe_ls-macos-arm64
# - dist/smpe_ls-macos-amd64
# - dist/smpe_ls-windows-amd64.exe
# - dist/smpe_ls-windows-arm64.exe
```

## 📋 What's New

### Version 1.3.3

**New Features**

- 🖋️ **Signature Help** - When the cursor is inside an operand's parentheses (`DISTLIB(│)`), a floating box shows the expected parameter, a short description and the type (from `smpe.json`). It triggers automatically while typing `(` and after accepting an operand from the completion list; boolean flag operands show no box. Toggle with `smpe.signatureHelp.enabled` (default `true`).

**Bug Fixes**

- **Fix language discrepancies** - Free Form Query now has only english button labels

### Version 1.3.2

**New Features:**

- 💡 **Code Actions (Quick Fixes)** - The editor lightbulb (`Cmd+.` / `Ctrl+.`) now offers one-click fixes for diagnostics:
  - **Add statement terminator** - inserts the missing `.`
  - **Insert operand X** / **Insert all required operands** - inserts skeletons for missing required operands
  - **Set REWORK to current date** - fills an empty `REWORK()` with today's Julian date (`yyyyddd`)

### Version 1.3.1

**Bug Fixes:**

- 🐛 **MCS completion menu stays open while typing `++STATEMENT` prefix** - Typing `++S`, `++SR`, `++SRC`, … no longer dismisses the completion list. The completion menu remains open and continues to filter MCS statements as more characters are typed.
- 🐛 **Snippet items respect VSCode/blink prefix filter** - Boilerplate snippet completion items now carry an explicit `filterText` so that VSCode (and blink-cmp in Neovim) match them against the typed prefix. Snippets are no longer hidden when typing `++P`, `++PT`, etc.

### Version 1.3.0

**New Features:**

- ✨ **Saved Queries in Free Form Query** - Save, manage and reuse CSI queries. Queries are stored in `.smpe-saved-queries.yaml` in the workspace root. The Free Form Query panel shows a collapsible saved queries section below the input form.
- ⚙️ **Auto-Detect Language Mode Toggle** - New setting `smpe.editor.autoDetectLanguage` (default: `true`) and command `SMP/E: Toggle Auto-Detect Language Mode`. When disabled, manual language mode changes (e.g. switching a `.smpe` buffer to REXX) are preserved.

## 🔧 Configuration

### VSCode Settings

Add to `.vscode/settings.json`:

```json
{
  "smpe.serverPath": "smpe_ls",
  "smpe.debug": false,
  "smpe.dataPath": "~/.local/share/smpe_ls/smpe.json"
}
```

### Logging

Logs are written to:

- **Linux/macOS:** `~/.local/share/smpe_ls/smpe_ls.log`
- **Windows:** `%LOCALAPPDATA%\smpe_ls\smpe_ls.log`

Enable debug logging:

```bash
smpe_ls --debug
```

Or in VSCode:

```json
{
  "smpe.debug": true
}
```

## 🏛️ Architecture

### Parser Strategy

**Recursive Descent Parser** with AST generation:

- Statement-specific parser functions
- Grammar derived from IBM SMP/E Reference documentation
- Zero external parser dependencies

### Data Sources

**smpe.json** (`data/smpe.json`)

- Statement and operand descriptions
- Grammar rules and validation logic
- Completion and hover information
- Required operands and mutually exclusive operands

### Components

```text
smpe_ls/
├── cmd/
│   ├── smpe_ls/        # Language server binary
│   ├── smpe_lint/      # Command-line linter for CI/CD
│   └── smpe_test/      # Central test suite
├── internal/
│   ├── completion/     # Code completion provider
│   ├── diagnostics/    # Syntax validation
│   ├── hover/          # Documentation provider
│   ├── parser/         # AST parser
│   └── handler/        # LSP protocol handler
├── client/
│   └── vscode-smpe/    # VSCode extension
└── data/
    └── smpe.json       # Statement definitions
```

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Backward Compatibility** - Don't break existing functionality
2. **Minimal Changes** - Keep changes focused and targeted
3. **Test Coverage** - Add tests for new features
4. **Documentation** - Update README and inline docs

### Development Workflow

```bash
# Install development dependencies
make install

# Run tests
make test-all

# Build for all platforms
make build-all

# Create release packages
make release

# Clean build artifacts
make clean-all
```

## 📚 Resources

- **IBM z/OS SMP/E Documentation:** https://www.ibm.com/docs/en/zos/3.1.0?topic=smpe-zos-reference
- **Language Server Protocol:** https://microsoft.github.io/language-server-protocol/
- **VSCode Extension API:** https://code.visualstudio.com/api

## 🙏 Acknowledgments

Statement and operand descriptions are derived from:

**IBM z/OS 3.1 SMP/E Reference**
© Copyright IBM Corporation
https://www.ibm.com/docs/en/zos/3.1.0?topic=smpe-zos-reference

SMP/E is a registered trademark of International Business Machines Corporation.

## 📄 License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

Copyright (C) 2025 Ronald Funk

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.

**Commercial Licensing:** This software is also available under a commercial license for organizations that wish to use it without the restrictions of the AGPL-3.0 license. For commercial licensing inquiries, please contact Ronald Funk.

See [LICENSE](LICENSE) file for the full license text.

## 🐛 Troubleshooting

### Server Not Starting

1. Verify installation:

   ```bash
   which smpe_ls
   smpe_ls --version
   ```

2. Check log file:

   ```bash
   tail -f ~/.local/share/smpe_ls/smpe_ls.log
   ```

3. Verify data file:

   ```bash
   ls -la ~/.local/share/smpe_ls/smpe.json
   ```

### VSCode Extension Issues

1. Reload window: `Cmd+Shift+P` → "Developer: Reload Window"

2. Check Output panel: View → Output → "SMP/E Language Server"

3. Reinstall server:

   ```bash
   make clean install
   ```

### Build Issues

1. Verify Go version:

   ```bash
   go version  # Should be 1.21+
   ```

2. Clean and rebuild:

   ```bash
   make clean-all
   make build
   ```

## 🗺️ Roadmap

- [x] Code Actions (Quick Fixes)
- [x] Signature Help
- [ ] Rename (SYSMOD/FMID)

---

**Made with ❤️ by Sir Tobi aka Cybersorcerer**
