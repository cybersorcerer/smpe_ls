# SMP/E Language Server

A Language Server Protocol (LSP) implementation for SMP/E (System Modification Program/Extended) written in Go.

## Features

### Release 0.3.0 (Current)

**New Features:**

- **++MAC Statement Support** - Macro library management with context-sensitive validation
  - ADD/UPDATE mode: DISTLIB required, mutually exclusive operands validated
  - DELETE mode: Special validation for DELETE with FROMDS/RELFILE/SSI/TXLIB
- **++MACUPD Statement Support** - Macro update operations with inline data support
- **++MOD Statement Support** - Load module operations with inline data support
- **++SRC Statement Support** - Source code operations with inline Assembler/C/etc. support
- **++SRCUPD Statement Support** - Source update operations with inline data support
- **Inline Data Architecture** - Dynamic inline data handling via `inline_data` attribute in smpe.json
  - Supports ++JCLIN (JCL), ++MAC (Assembler), ++MACUPD, ++MOD, ++SRC, ++SRCUPD
  - No SMP/E diagnostics or completions inside inline data regions
  - External data operands (FROMDS, RELFILE, SSI, TXLIB) properly detected
- **Enhanced Syntax Highlighting** - Inline data (JCL/Assembler/REXX) displayed as plain text without false SMP/E highlighting
- **Visual Diagnostic Severity** - Unicode symbols in diagnostic messages for better distinction:
  - 🔴 ERROR
  - ⚠️ WARNING
  - ℹ️ INFORMATION
  - 💡 HINT

**Enhanced Diagnostics:**

- Fixed dataset name handling - dots in DSN(MY.DATA.SET) no longer treated as statement terminators
- Boolean operand parsing - DELETE, NOAPARS, etc. now correctly validated
- Improved terminator detection - only detects periods outside parentheses
- Completion now shows actual statement names instead of generic "MCS"
- Hover now shows correct statement/operand names instead of "MCS"

### Release 0.2.0

**New Features:**

- **Multiline Parameter Support** - Correctly parses operand parameters spanning multiple lines
- **Malformed Parenthesis Detection** - Diagnostics for missing closing parentheses in statement and operand parameters
- **Flexible Whitespace Handling** - Supports both `OPERAND(...)` and `OPERAND (...)` syntax
- **++JCLIN Statement Support** - Full support including inline JCL data handling
- **++JAR Statement Support** - JAR file management operations
- **++JARUPD Statement Support** - JAR update operations
- **++VER Statement Support** - Version specification
- **++ZAP Statement Support** - Superzap operations
- **Improved Completion** - mutually_exclusive operands now shown in completion (validated via diagnostics)

**Enhanced Diagnostics:**

- Detects unbalanced parentheses in statement parameters (e.g., `++APAR(A12345`)
- Detects unbalanced parentheses in operand parameters (e.g., `TO(A12345, A23456`)
- Validates mutually exclusive operands (e.g., ++JCLIN FROMDS vs RELFILE)
- Properly handles ++JCLIN inline data (skips diagnostics for JCL lines)

### Release 0.1.0

Core features:

- **Syntax Highlighting** - Color coding for MCS statements, operands, and comments (`/* */`)
- **Context-Aware Code Completion**
  - MCS statement completion with parameter placeholders
  - Operand completion based on statement type
  - Context-sensitive value completion (e.g., SYSMOD IDs from document)
  - Special handling for ++HOLD REASON operand based on hold type (ERROR, SYSTEM, FIXCAT, USER)
- **Diagnostics** - Real-time validation including:
  - Missing or malformed statement terminators (`.`)
  - Missing required statement parameters
  - Missing required operands (derived from syntax diagrams)
  - Empty operand parameters
  - Unknown operands
  - Duplicate operands
  - Dependency violations (e.g., RFDSNPFX requires FILES)
  - Unknown statement types
- **Hover Information** - Documentation and parameter details for statements and operands

## Supported MCS Statements

Version 0.3.0 supports 17 MCS statements with full diagnostics validation:

**Control MCS:**

- `++APAR` - Service SYSMOD (temporary fix)
- `++ASSIGN` - Source ID Assignment
- `++DELETE` - Delete Load Module
- `++FEATURE` - SYSMOD Set Description
- `++FUNCTION` - Function SYSMOD
- `++HOLD` - Exception Status
- `++IF` - Conditional Processing
- `++JAR` - JAR file management
- `++JARUPD` - JAR update operations
- `++JCLIN` - Job Control Language Input with inline JCL support
- `++MAC` - Macro library management (NEW in 0.3.0)
- `++MACUPD` - Macro update operations (NEW in 0.3.0)
- `++MOD` - Load module operations (NEW in 0.3.0)
- `++SRC` - Source code operations (NEW in 0.3.0)
- `++SRCUPD` - Source update operations (NEW in 0.3.0)
- `++VER` - Version specification
- `++ZAP` - Superzap operations

**Additional Statements in smpe.json (completion and hover available):**

- All Data Element MCS with language variants (++BOOK, ++CLIST, ++EXEC, ++FONT, ++HELP, ++MSG, ++PARM, etc.)
- Additional Control MCS (++MOVE, ++NULL, ++PRODUCT, ++PROGRAM, ++PTF, ++RELEASE, ++RENAME, ++USERMOD)

## Installation

### Prerequisites

- Go 1.19 or later
- VSCode (for testing)
- Node.js and npm (for VSCode extension)

### Build and Install

```bash
# Build the language server
make build

# Install to ~/.local/bin
make install

# Build VSCode extension
make vscode
```

## Usage

### VSCode

1. Build the VSCode extension:

   ```bash
   make vscode
   ```

2. Open `client/vscode-smpe` in VSCode

3. Press F5 to launch Extension Development Host

4. Create or open a `.smpe`, `.mcs`, or `.smp` file

5. Start typing SMP/E statements and enjoy the language features!

### Configuration

VSCode settings (`.vscode/settings.json`):

```json
{
  "smpe.serverPath": "smpe_ls",
  "smpe.debug": false
}
```

- `smpe.serverPath`: Path to the smpe_ls executable (default: searches in ~/.local/bin)
- `smpe.debug`: Enable debug logging (logs to ~/.local/share/smpe_ls/log)

### Command Line Options

```bash
smpe_ls [options]

Options:
  --debug          Enable debug logging
  --version        Show version
  --data <path>    Path to smpe.json data file (default: data/smpe.json)
```

### Building

```bash
# Build only
make build

# Build and install
make install

# Clean build artifacts
make clean

# Run tests
make test
```

### Logging

The server logs to `~/.local/share/smpe_ls/log`

Enable debug logging:

```bash
smpe_ls --debug
```

Or in VSCode settings:

```json
{
  "smpe.debug": true
}
```

## Architecture

### Parser Strategy

Uses a **Recursive Descent Parser** with no external dependencies:

- One parser function per MCS statement type
- Grammar derived from syntax diagrams
- Semantic information from smpe.json

### Data Sources

1. **Syntax Diagrams** (`syntax_diagrams/*.png`)
   - Define grammar and structure
   - Used for parser implementation
   - Show mandatory vs. optional operands

2. **smpe.json** (`data/smpe.json`)
   - Statement descriptions
   - Operand details and allowed values
   - Used for hover info and completion

## Examples

### Example SMP/E File

```smpe
/* Sample SMP/E MCS statements */
++APAR(AB12345)
    DESCRIPTION(Fix for security vulnerability)
    FILES(5)
    RFDSNPFX(APARA12)
    REWORK(20250101)
    .

++FUNCTION(HBB7790)
    DESCRIPTION(Base function for product XYZ)
    FILES(100)
    RFDSNPFX(FUNC123)
    .

++HOLD(AB12345)
    FMID(HBB7790)
    REASON(B12345)
    ERROR
    COMMENT(Critical security fix required before apply)
    .

++IF FMID(HBB7790)
    THEN
    REQ(FEATURE)
    .

/* NEW in 0.2.0: Multiline parameters */
++ASSIGN
    SOURCEID(PROD2025)
    TO(
        AB12345,
        AB23456,
        AB34567
    )
    .

/* NEW in 0.2.0: ++JCLIN with inline JCL data */
++JCLIN.
//SMPMCS   JOB (ACCT),'INSTALL',CLASS=A,MSGCLASS=X
//STEP1    EXEC PGM=IEWL
//SYSLMOD  DD DSN=SYS1.LINKLIB,DISP=SHR
/*

/* NEW in 0.2.0: ++JAR statement */
++JAR(MYJAR) DISTLIB(AJARLIB) SYSLIB(SJARLIB) RELFILE(2)
    PARM(PATHMODE(0,6,4,4))
    LINK('../myapp.jar')
    .
```

## Troubleshooting

### Server Not Starting

1. Check that smpe_ls is in PATH or ~/.local/bin:

   ```bash
   which smpe_ls
   ```

2. Check log file:

   ```bash
   tail -f ~/.local/share/smpe_ls/log
   ```

3. Verify data file exists:

   ```bash
   ls -la data/smpe.json
   ```

### VSCode Extension Not Working

1. Ensure server is installed:

   ```bash
   make install
   ```

2. Check VSCode Output panel:
   - View � Output
   - Select "SMP/E Language Server" from dropdown

3. Reload VSCode window:
   - Cmd+Shift+P � "Developer: Reload Window"

## Future Enhancements

- Additional MCS statements (++VER, ++MOD, ++MAC, etc.)
- Go to Definition
- Find References
- Document Symbols
- Code Actions
- Formatting
- Neovim plugin

## Acknowledgments

Statement and operand descriptions, hover information, and documentation content are derived from:

**IBM z/OS 3.1 SMP/E Reference**
© Copyright IBM Corporation
<https://www.ibm.com/docs/en/zos/3.1.0?topic=smpe-zos-reference>

SMP/E is a registered trademark of International Business Machines Corporation.

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Maintain backward compatibility
2. Make minimal, targeted changes
3. Test thoroughly before submitting
4. Update documentation
