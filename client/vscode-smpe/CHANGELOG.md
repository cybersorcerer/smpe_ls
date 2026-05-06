# Changelog

All notable changes to this project are documented in this file.

## [1.1.0] - 2026-05-05

### Added

- **HOLD Comments Viewer** - A dedicated HOLD button column appears in SYSMOD Free Form Query results (GLOBAL zone only, rows with HOLDDATA). Clicking the button fetches the PTF member from SMPPTS via z/OSMF and displays the complete `++HOLD` block text in a dedicated side panel. If the PTF has already been accepted and is no longer in SMPPTS, a clear warning message is shown instead.
- **Free Form Query: scrollable results table** - The query form and column header row now remain fixed while the result rows scroll independently, making it easy to navigate large result sets without losing context.

### Fixed

- **CRLF line endings** - z/OS Dataset Members opened via Zowe Explorer are transferred with CRLF line endings. The parser now normalizes `\r\n` to `\n` before processing, fixing false `unknown_operand` diagnostics for `COMMENT` content in `++HOLD` statements.
- **HOLD Comments: complete block displayed** - The viewer now shows the full `++HOLD` block up to the next MCS statement (`++`), instead of truncating at the first period in the text. This ensures LMOD lists and other multi-line content are fully visible.
- **HOLD button color** - The HOLD button in the Free Form Query results now uses the same primary button color as the Pick buttons, consistent with the active VSCode color theme.

## [1.0.1] - 2026-05-05

### Fixed

- **`++HOLD` COMMENT false diagnostics** - The COMMENT operand in `++HOLD` statements is now correctly treated as free text. Previously, the diagnostics engine incorrectly reported "Comment requires parameter TEXT" and "Comment requires parameter ENHANCED HOLD-DATA". The content within `COMMENT(...)` is now fully ignored during validation.
- **`++HOLD` missing terminator false positive** - English apostrophes inside `COMMENT(...)` free text (e.g. `site's`, `don't`) no longer corrupt statement terminator detection. The terminator scanner now blindly skips the content of free-text operands by paren-counting only, without quote tracking.
- **SMRTDATA operand added to `++HOLD`** - The enhanced HOLDDATA operand `SMRTDATA` (with sub-operands `CHGDT`, `FIX`, `SYMP`) is now correctly modelled as a separate structured operand and validated accordingly.

## [1.0.0] - 2026-04-28

### Added

- **Boilerplate Snippets** - All Control MCS statements now offer a second completion item `…` alongside the standard keyword completion (distinguishable by the snippet icon). Selecting it inserts a ready-to-use multi-line template with LSP tab stops for all required operands. The REWORK operand is pre-filled with the current date in `yyyyddd` format. Templates are defined in `smpe.json` and can be extended without code changes.

## [0.9.8] - 2026-04-17

### Fixed

- **Content beyond column 72 false positive in inline data** - The column 72 diagnostic no longer fires for lines that are inline data (e.g. JCL after `++JCLIN`, source after `++SRC`). Inline data regions are now excluded from the check, consistent with all other diagnostics.
- **Free Form CSI Query: default subentries** - When selecting an entry type, the subentries field is now pre-filled with a meaningful default set. Defaults are defined for SYSMOD, DDDEF, GLOBALZONE, TARGETZONE, DLIBZONE, and DLIB. The subentry picker also pre-selects the matching checkboxes. Entry types without a defined default still clear the field on change.
- **z/OS Dataset LLQ language detection** - Reworked LLQ detection for Zowe Explorer datasets: member name removal now uses a global regex flag, the LLQ is extracted robustly regardless of URI format, and the comparison is fully case-insensitive. Fixes cases where only the first configured LLQ triggered SMP/E language mode.
- **`firstLine` detection pattern** - Simplified from an incomplete hardcoded list of MCS statement names to `^\+\+[A-Z]`, which correctly matches any valid MCS statement.
- **Training documentation** - Corrected `.smpe-zosmf.yaml` example: `csi` is a flat string list, `defaultCsi` is a server-level field, removed non-existent `zosmfBase` field.

## [0.9.7] - 2026-04-14

### Fixed

- **SYMPATH Operand not recognized in `++HFS`** - The parser incorrectly terminated multiline statements early when quoted path strings (e.g. `'../IVP/eqadrest.env'`) contained a dot before the closing parenthesis. The dot was mistakenly interpreted as a statement terminator. The terminator detection and parenthesis depth tracking are now quote-aware, correctly ignoring dots and parentheses inside `'...'` strings.
- **z/OS Dataset LLQ language detection** - When multiple LLQs were configured in `smpe.zosDatasetsLlq` (e.g. `["MCS", "USERMOD"]`), only the first entry triggered SMP/E language mode. The member name removal regex now uses a global flag to handle all URI formats reliably, and the LLQ comparison is fully case-insensitive. Also simplified the `firstLine` detection pattern from an incomplete statement list to `^\+\+[A-Z]`, covering all valid MCS statements.

## [0.9.6] - 2026-04-09

### Added

- **Training Documentation** - 22 Markdown training modules (DE + EN) covering installation, all VSCode extension features, CLI tools (`smpe_lint`, `smpe_outl`), and z/OSMF integration. Available under `docs/training/de/` and `docs/training/en/`.

## [0.9.5] - 2026-04-02

### Added

- **Check Missing Input Members** - New command `SMP/E: Check Missing Input Members` scans all `*.smpe` files in the workspace and checks whether the referenced input member files (e.g. `MYMOD.hlasm`, `MYSCRIPT.rexx`) exist in the configured search folders. Results are displayed in a filterable and sortable Webview table with Found/Missing status. Available in the Command Palette and via right-click on `.smpe` files in the editor and Explorer.
- **`smpe.checkMissingInputMembers.searchFolders` setting** - Configures which folders (relative to the workspace root) are searched for input member files. Default: `["customization"]`. Use `"*"` to search recursively in the entire workspace.

## [0.9.4] - 2026-03-31

### Added

- **smpe_outl** - New CLI tool that prints the document outline (symbols) of SMP/E MCS files as Markdown or JSON (`--json`), with optional LSP range output (`--ranges`). Part of the VSCode extension and release packages.
- **LLQ-based Language Detection** - z/OS datasets with configurable last level qualifiers (e.g. `.MCS`) are automatically activated as SMP/E files. Configurable via `smpe.zosDatasetsLlq` setting (default: `["MCS"]`).

### Changed

- **z/OSMF Config Lookup** - `.smpe-zosmf.yaml` is now resolved in this order: all open workspace folders (first match wins), then `~/.config/smpe_ls/` as a global fallback. Creating a config via command still creates it in the workspace root.
- **Dataset Member Listing** - `X-IBM-Attributes: base` header is now sent when listing PDS members, returning full ISPF-style attributes (User, Created, Modified, Ver, Mod).

## [0.9.3] - 2026-03-25

### Added

- **Workspace Symbols** - Search for SYSMOD definitions across all `.smpe` files in the workspace (`Cmd+T` or `workspace/symbol`)
- **Dataset Member Attributes** - PDS member listing via z/OSMF now requests `X-IBM-Attributes: base`, returning full ISPF-style attributes (User, Created, Modified, Ver, Mod)

## [0.9.2] - 2026-03-25

### Changed

- **Debug Logging Levels** - Separated `log()` (always visible) and `debugLog()` (only with `smpe.debug` enabled) across all extension modules (extension, client, configManager, queryProvider, freeFormPanel)

### Fixed

- **Logging When Debug Disabled** - Normal messages (startup, errors, query status) are now always shown in the output channel; only verbose debug details require `smpe.debug` to be enabled
- **Cell Tooltip in Result Tables** - Tooltip now also appears for cells with text longer than 40 characters, fixing cases where `scrollWidth` check alone missed truncated content
- **SYSMOD Subentry NPRE2/REQ2** - Removed invalid subentry names `NPRE2` and `REQ2` from the Free Form Query picker

## [0.9.1] - 2026-03-19

### Added

- **Folding Ranges** - MCS statements and multi-line comments can be collapsed/expanded in the editor
- **Debug Logging Control** - All extension logging now respects the `smpe.debug` setting (default: on)
- **Global Zone SYSMOD Subentries** - Free Form Query SYSMOD subentry picker now includes Global Zone subentries (ACCID, APPID, DELETE, HOLDDATA, NPRE, PRE, REQ, SREL, TLIBPREFIX)

### Fixed

- **Free Form Query Subentries** - CSI queries now include TARGETZONE in entries array, fixing empty subentry results for SYSMOD and other entry types
- **SYSMOD Subentry DELETE2** - Corrected invalid subentry name `DELETE2` to `DELETE`
- **CHANGELOG Link** - Fixed broken relative link in extension README (now absolute GitHub URL)

## [0.9.0] - 2026-03-18

### Added

- **USS Directory Browsing** - Clickable PATH links in DDDEF query results open a USS directory browser via z/OSMF Files REST API with breadcrumb navigation, file viewing, and automatic PATHPREFIX stripping
- **MVS Dataset Browsing** - Clickable DATASET links in DDDEF query results open PDS member listings or sequential datasets via z/OSMF Dataset REST API
- **Read-Only File Viewing** - USS files and dataset members open as read-only virtual documents (`smpe-uss://` and `smpe-ds://` schemes)
- **Automatic PATHPREFIX Resolution** - USS paths with SMP/E PATHPREFIX segments (e.g. `/Z31TGT/usr/include`) are automatically resolved by stripping prefix segments on 404

### Changed

- **Webview Layout** - All result panels (SYSMOD, DDDEF, Zone, USS, Dataset) now open in the main editor area instead of beside

## [0.8.9] - 2026-03-08

### Added

- **Unified SYSMOD List Queries** - CodeLens now generates one query per operand covering all SYSMODs in a list, instead of one query per list element
- **Extended SYSMOD Reference Operands** - All 11 operands that accept SYSMOD lists are now recognized: DELETE, FMID, NPRE, PRE, REQ, RESOLVER, RMID, SUP, TO, UMID, VERSION
- **CSI List Support** - `.smpe-zosmf.yaml` now supports multiple CSIs per server with optional `defaultCsi`
- **CSI Selection in Free Form Query** - New dropdown to select the CSI in the Free Form Query panel
- **Missing Entry Display** - SYSMODs and DDDEFs not found by a query are shown in the result table with `.` in all columns

### Fixed

- **SYSMOD Query Filter** - Filter correctly generates `ENAME='SM1'|ENAME='SM2'|...` format for multiple SYSMODs
- **Space and Comma Separated Lists** - Both separators are now handled identically in CodeLens and query input

## [0.8.8] - 2026-03-03

### Added

- **smpe.json Template System** - Introduced `$ref` template mechanism reducing smpe.json size by ~42% while resolving all references at load time (zero consumer impact)
- **PATH and INITDISP in DDDEF Query** - DDDEF query result table now shows PATH (after DATASET) and INITDISP columns

### Fixed

- **SYSMOD Query Filter** - Filter now correctly generates `ENAME='SM1'|ENAME='SM2'|...` for all SYSMODs in list operands (PRE, REQ, SUP)
- **List Operand Parsing in CodeLens** - SMP/E list items separated by spaces (not only commas) are now correctly split into individual CodeLens entries
- **DESC Operand Formatting** - `DESC(This is a long description)` no longer incorrectly splits at spaces; only `list`-typed operands are split
- **List Display Format** - List operands (e.g. PRE, REQ, SUP) now render items on separate lines with correct indentation and closing `)` aligned to operand indent
- **Dot in Statement Parameter** - `++PRODUCT(PROD001,01.00.00)` no longer duplicates output; dots inside parenthesized statement parameters are no longer treated as terminators

## [0.8.7] - 2026-02-26

### Added

- **Zowe Explorer Integration** - Language Server now activates for dataset members opened via Zowe Explorer (`zowe-ds://`, `zowe-uss://`, `zowe-jobs://` URI schemes)
- **Auto Language Detection** - Files without extension are automatically detected as SMP/E when the first non-empty line starts with `++`; detection runs before and after Language Server startup to also catch already-open documents
- **firstLine Detection** - `firstLine` pattern added for fast language detection without file extension
- **Server Version in Output** - Language Server version and commit hash are now shown in the VSCode Output panel on startup

### Fixed

- **Nested Parentheses in List Wrapping** - `PARM(PATHMODE(0,7,5,5))` no longer gets incorrectly split at commas inside nested parentheses, preventing duplicate output after formatting
- **Dot in Single-Quoted String** - `LINK('../path')` no longer causes the formatter to treat `.` inside the string literal as a statement terminator, preventing duplicate operand output
- **Dot in Inline Comment** - A `.` inside an inline comment (e.g. `/* version 12.7 */`) no longer terminates the statement prematurely, fixing broken syntax highlighting for subsequent comment lines

## [0.8.6] - 2026-02-25

### Added

- **Extended CSI Entry Types in Free Form Query** - Added DLIBZONE, FEATURE, FMIDSET, HOLDDATA, JAR, JARUPD, OPTIONS, ORDER, PRODUCT, PROGRAM, UTILITY, ZONESET with their correct subentries

### Fixed

- **Duplicate Emoji in Diagnostics** - Parameter length exceeded diagnostics no longer show two warning emojis
- **activationEvents** - Removed redundant `onLanguage:smpe` activation event (VSCode generates it automatically)

## [0.8.5] - 2026-02-24

### Added

- **HFS Entry Types in Free Form Query** - Added AIX1-5, CLIENT1-5, OS21-5, UNIX1-5, WIN1-5 with their subentries to the Free Form CSI Query
- **Entry Type Picker** - New radio button picker for Entry Type selection (alphabetically sorted, no groups)

### Fixed

- **Export JSON/CSV** - Export buttons in non-Free Form query panels (SYSMOD, DDDEF, Zone) now work correctly (CSP inline handler fix)

## [0.8.4-alpha] - 2026

### Added

- **List Wrapping** - Comma-separated operand lists (e.g., `PRE`, `REQ`, `SUP`) are automatically wrapped when exceeding a configurable threshold (default: 2 items per line)
- **CodeLens for z/OSMF Queries** - Inline CodeLens actions for querying SYSMODs and DDDEFs via z/OSMF
- New formatting setting: `smpe.formatting.wrapListsAfterN` - Number of list items before wrapping (default: 2, 0 = disabled)

### Fixed

- **Parser Crash** - Fixed `slice bounds out of range` panic when formatting files containing `*/` before `/*` on the same line
- **Formatting Idempotency** - Second format pass no longer destroys comments; formatting is now stable/idempotent
- **Inline Comment Handling** - Single-line and multi-line inline comments stay with their operand during formatting instead of being moved to the statement header
- **DELETE Mode DISTLIB Diagnostic** - `++MOD DELETE`, `++SRC DELETE`, and `++PROGRAM DELETE` no longer produce false "Missing required operand: DISTLIB" warnings

## [0.8.3-alpha] - 2026

### Added

- **z/OSMF Free Form CSI Query** - New Webview-based query interface (`SMP/E: Free Form CSI Query`)
  - Combined input form and result table in a single Webview panel
  - Server selection from `.smpe-zosmf.yaml` configuration
  - Zone input with wildcard pattern matching (`*` and `?`)
  - Entry Type dropdown (SYSMOD, DDDEF, TARGETZONE, DLIB, GLOBALZONE, etc.)
  - Free-form Subentries and Filter input
  - Dynamic result table with column headers derived from specified subentries
  - JSON and CSV export
- **Zone Pattern Matching** - Wildcard support (`*`, `?`) for zone parameters in z/OSMF queries
  - Matches against `zones` list defined in `.smpe-zosmf.yaml` server configuration

### Fixed

- **List-type Parameter Validation** - `PRE`, `REQ`, `SUP` operands with multiple values no longer produce false length warnings
- **golangci-lint** - Fixed `ineffassign` error in `formatting.go`

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
