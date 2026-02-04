# Changelog

All notable changes to this project are documented in this file.

## [0.8.1-alpha] - 2026

### Added

- **Command-Line Linter** - New `smpe_lint` tool for CI/CD integration
  - Markdown and JSON output formats
  - Configurable diagnostics via YAML/JSON config files
  - `--disable` flag to disable specific diagnostics
  - `--warnings-as-errors` for strict mode
  - `--init` to create sample configuration files

### Fixed

- **Comment Indentation** - Comments now start at column 3 (2 space indent) instead of column 1
- **Multi-line Comment Before Terminator** - Multi-line comments before terminator are now correctly placed before the `.` during formatting
- **SHSCRIPT Comma Preservation** - Commas in `SHSCRIPT(MYSCRIPT, POST)` are now preserved during formatting
- **Leading Comment Indentation** - Leading comments moved into statements now correctly start at column 3

## [0.8.0-alpha] - 2026

### Fixed

- **Multi-line Comment Parsing** - Parser now correctly captures full text of multi-line comments
- **Multi-line Comment Formatting** - Formatting preserves multi-line comments inside statements
- **Comment Line Wrapping** - Long lines within multi-line comments are now wrapped at column 72
- **Comment Before Terminator** - Comments before terminator on same line (e.g., `CALLLIBS /* comment */.`) are now preserved during formatting

## [0.7.8-alpha] - 2026

### Added

- **Column 72 Diagnostic** - Error diagnostic when content extends beyond column 72 (configurable via `smpe.diagnostics.contentBeyondColumn72`)
- **Standalone Comment Diagnostic** - Error diagnostic for comments outside MCS statements (configurable via `smpe.diagnostics.standaloneCommentBetweenMCS`)
  - Comments before first MCS statement
  - Comments between MCS statements (except for inline data)
  - Comments after last MCS statement (except for inline data)
- **Move Standalone Comments** - Optional formatting to automatically move standalone comments into the next MCS statement (configurable via `smpe.formatting.moveLeadingComments`)
  - Comments before first MCS statement → moved into first statement
  - Comments between MCS statements → moved into following statement
  - Comments after last MCS statement → cannot be moved (diagnostic only)

### Changed

- **Formatting** - Now enforces IBM SMP/E column 72 limit, wraps long lines automatically
- **Comment Preservation** - Formatting now preserves comments and places them correctly per IBM rules
- **Inline Data Protection** - Formatting no longer modifies statements that expect inline data

### Bug Fixes

- Added missing SYMPATH to ++HFS
- Add missing PRE/POST to SHSCRIPT
- Fixed "Missing inline data required" diag if SHSCRIPT is defined
- Fixed operand length Diagnostics

## [0.7.7-alpha] - 2026

### Added

- **Go to Definition** - Navigate to SYSMOD/FMID definitions within the same file (`F12` or `Cmd+Click`)
- **Find All References** - Find all references to a SYSMOD or FMID (`Shift+F12`)
- **Git Commit Hash** - Build includes commit hash for traceability (`smpe_ls --version`)

## [0.7.6] - 2026

### Added

- **Document Symbols / Outline View** - Navigate SMP/E files using the Outline panel or `Cmd+Shift+O`
  - Hierarchical view of all statements with their operands
  - Quick navigation to any statement in the document
  - Symbol icons based on statement type (SYSMOD, VER, MAC, etc.)
- **Whitespace Tolerance** - Parser now accepts spaces between `++` and statement name (e.g., `++ VER`)
  - Formatting automatically corrects this to proper `++VER` format

### Changed

- **Formatting** - Statement terminator (`.`) is now always placed on its own line at the beginning

### Fixed

- Highlighting for statements with spaces after `++` now covers the complete statement

## [0.7.5] - 2026

### Added

- **Document Formatting** - Format SMP/E documents with `Shift+Alt+F` (Windows/Linux) or `Shift+Option+F` (macOS)
  - Each operand on its own line (configurable)
  - Configurable continuation line indentation
  - Can be enabled/disabled via settings
- New formatting settings:
  - `smpe.formatting.enabled` - Enable/disable formatting (default: true)
  - `smpe.formatting.indentContinuation` - Spaces for continuation lines (default: 3)
  - `smpe.formatting.oneOperandPerLine` - Place each operand on its own line (default: true)
  - `smpe.formatting.formatOnSave` - Automatically format on save (default: false)
- **Column Rulers** - Visual guides at columns 72 and 80 for mainframe card boundaries
- Improved hover information formatting with separated required/optional operands
- `inline_data` support for all Data Element MCS statements with language variants

### Fixed

- UTF-8/UTF-16 character position calculation for non-ASCII characters (e.g., umlauts)
- Operand completion now works for all MCS statements
- Completion after inline data now correctly offers MCS statements when typing `++`
- DISTLIB operand correctly marked as required only for Data Element MCS, ++PROGRAM, and ++MOVE
- Output panel no longer opens automatically on extension startup

### Changed

- Diagnostics settings are now dynamically applied without restart

## [0.7.0] - 2025

### Added

- HFS MCS statements
- `++SHELLSCR` statement support
- Improved inline data parsing and diagnostics
- Various fixes in smpe.json

## [0.6.0] - 2025

### Changed

- Completion and diagnostics fully AST-based

## [0.5.1] - 2025

### Added

- `++MOVE` statement support
- `++MOD` statement support

## [0.4.0] - 2025

### Added

- `++MAC`, `++MACUPD`, `++SRC`, `++SRCUPD` statement support
- Inline data architecture with dynamic handling via smpe.json
- Improved syntax highlighting for inline data
- Visual diagnostic severity with Unicode symbols

### Fixed

- Dataset names with dots are handled correctly
- Boolean operand parsing fixed
- Completion and hover show correct statement names

## [0.3.0] - 2025

### Added

- Multiline parameter support
- Detection of missing closing parentheses
- Flexible whitespace handling
- `++JCLIN`, `++JAR`, `++JARUPD`, `++VER`, `++ZAP` statements
- `++JCLIN` inline JCL support

### Fixed

- Multiline parameters are correctly recognized
- Unbalanced parentheses are diagnosed

## [0.2.0] - 2025

### Added

- Basic diagnostics
- Hover information from JSON file
- Context-sensitive code completion

## [0.1.0] - 2025

### Added

- Initial release
- Syntax highlighting for SMP/E MCS statements
- VS Code extension framework
