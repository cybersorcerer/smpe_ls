# SMP/E Language Server

A modern Language Server Protocol (LSP) implementation for IBM SMP/E (System Modification Program/Extended) written in Go.

[![Version](https://img.shields.io/badge/version-0.7.6-blue.svg)](https://github.com/cybersorcerer/smpe_ls/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)

## ✨ Features

- **🎨 Syntax Highlighting** - Color coding for MCS statements, operands, and comments
- **💡 Intelligent Code Completion** - Context-aware completion for statements and operands
- **🔍 Real-time Diagnostics** - Instant validation of SMP/E syntax and semantics
- **📖 Hover Documentation** - Inline documentation from IBM SMP/E Reference
- **📄 Document Symbols** - Outline view and quick navigation
- **📝 Document Formatting** - Auto-format SMP/E statements
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

### Version 0.7.6 (Current)

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

### Version 0.7.6 (Latest)

**New Features:**

- 📄 **Document Symbols** - Outline view and quick navigation (`Cmd+Shift+O`)
- 📝 **Document Formatting** - Auto-format SMP/E statements with configurable options
- 🔧 **Whitespace Tolerance** - Parser accepts `++ VER` syntax (formatting corrects it)
- 📏 **Column Rulers** - Visual guides at columns 72 and 80 for mainframe card boundaries
- ⚙️ **Format on Save** - Optionally auto-format when saving

**Improvements:**

- Statement terminator (`.`) now placed on its own line during formatting
- Highlighting correctly covers statements with spaces after `++`

### Version 0.7.0

**New Features:**
- ✨ **HFS Statement Support** - Added 10 missing HFS-type statements (++AIX1-5, ++CLIENT1-5)
- 🌍 **National Language Identifiers** - Full support for ++HFSxxx variants (32 language identifiers)
- ✨ **Hover Unit Tests** - Complete test coverage for hover functionality (11 tests)
- 📊 **Enhanced Test Suite** - Human-readable test output with detailed diagnostics
- 🤖 **GitHub Actions** - Automated multi-platform builds and releases
- 🔧 **Cross-compilation** - Native binaries for 6 platforms (Linux, macOS, Windows - AMD64 & ARM64)
- 📦 **Release Automation** - Automatic GitHub releases on version tags
- 🎯 **Makefile Improvements** - New targets: `test-suite`, `test-all`, `build-all`, `release`

**HFS Statements:**
- Added ++AIX1-5 (AIX client elements)
- Added ++CLIENT1-5 (generic client elements)
- ++HFS now supports all 32 national language identifier variants (++HFSENU, ++HFSDEU, etc.)
- Total HFS-type statements: 27 (2 base + 25 platform-specific variants)

**Test Infrastructure:**
- Central test runner with grouped results (passed/failed)
- Diagnostics organized by severity (🔴 Errors → ⚠️ Warnings → ℹ️ Info)
- Pass rate statistics and detailed failure reporting
- 19 comprehensive test files with valid and invalid test cases

**CI/CD:**
- Automatic builds on push to main branch
- Release builds on version tags with GitHub Releases
- Code coverage reporting to Codecov
- golangci-lint integration

### Version 0.6.0

**New Features:**
- Complete AST-based refactoring for completion and diagnostics
- All 25 Control MCS statements with full diagnostics validation
- Enhanced statement validation (++PRODUCT, ++PROGRAM, ++PTF, ++RELEASE)
- Improved required operand validation
- Better error messages for missing inline data

### Version 0.5.0

**New Features:**
- ++MOVE statement with complex DISTLIB/SYSLIB mode validation
- Enhanced inline data handling
- Improved diagnostics for mutually exclusive operands

### Version 0.4.0

**New Features:**
- ++NULL statement support
- Refactored diagnostics and completion to use AST
- Enhanced multiline statement handling

### Version 0.3.0

**New Features:**
- ++MAC, ++MACUPD, ++MOD, ++SRC, ++SRCUPD statement support
- Dynamic inline data handling via smpe.json
- Enhanced syntax highlighting for inline data
- Visual diagnostic severity with Unicode symbols (🔴 ⚠️ ℹ️ 💡)

### Version 0.2.0

**New Features:**
- Multiline parameter support
- Malformed parenthesis detection
- ++JCLIN, ++JAR, ++JARUPD, ++VER, ++ZAP statement support
- Improved completion for mutually exclusive operands

### Version 0.1.0

**Initial Release:**
- Syntax highlighting
- Context-aware code completion
- Real-time diagnostics
- Hover information

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

```
smpe_ls/
├── cmd/
│   ├── smpe_ls/        # Language server binary
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

- [ ] Neovim plugin
- [ ] Go to Definition
- [ ] Find References
- [x] Document Symbols
- [ ] Code Actions
- [x] Formatting
- [x] All MCS statements
- [ ] SMP/E Command language support

---

**Made with ❤️ by cybersorcerer**
