# Changelog

All notable changes to this project are documented in this file.

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
