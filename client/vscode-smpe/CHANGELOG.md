# Changelog

All notable changes to this project are documented in this file.

## [1.3.9] - 2026-08-25

### Added

- **FETCHOPT parameter values** - `data/smpe.json` now declares `FETCHOPT(PACK|NOPACK)` correctly (it was missing its parameter entirely), so it gets hover, completion, and formatting support like the rest of `LEPARM`'s attributes.

### Fixed

- **LEPARM and other sub-operand containers (parsing, completion, outline, formatting)** - A parser bug duplicated every operand's parameter value into two identical AST nodes; harmless for most operands, but it corrupted semantic-highlighting tokens and, combined with a second bug, silently dropped `LEPARM`/`FROMDS` entirely from the Outline view and `smpe_outl` output. Completion inside a nested sub-operand (e.g. `LEPARM(AC(│)`) leaked the parent's suggestion list instead of the sub-operand's own; already-used sub-operands (including alias forms like `AMOD`/`AMODE`) were still re-offered; and enumerated pipe-value operands (`AMODE`, `UPCASE`, `FETCHOPT`, …) offered no completions at all inside their parentheses. Formatting of comma- or space-separated sub-operand lists (`LEPARM`, `FROMDS`) dropped the user's original separator and never wrapped long lists onto multiple lines, regardless of separator.

## [1.3.8] - 2026-08-25

### Fixed

- **Outline / document symbol range past comment dates** - A `.` inside a `/* ... */` block comment (e.g. a German-style date like `19.11.25`) was mistaken for the statement terminator, cutting the Outline view, folding range, and `smpe_outl` symbol range short at the comment instead of the real terminator. Fixed in the shared symbol logic and in `smpe_outl`, which had its own duplicated copy of the same bug.

### Changed

- **`smpe_outl` no longer duplicates symbol logic** - Now shares `internal/symbols.Provider` instead of carrying its own copy of the end-position and symbol-kind logic, so future fixes only need to happen in one place.

## [1.3.7] - 2026-07-08

### Added

- **New Quick Fix: "Update REWORK to current date"** - Refreshes an already-filled `REWORK()` value to today's date when it is stale, available right from the cursor without needing a diagnostic first. The existing "Set REWORK to current date" fix (for an empty `REWORK()`) is unchanged and keeps handling that case.

## [1.3.6] - 2026-07-07

### Added

- **smpe_lint `--data` flag** - New `--data <path>` option to set the smpe.json location explicitly, for Docker containers and CI runners without a usable home directory (matching `smpe_outl`).

### Changed

- **smpe_lint default path resolution** - The default smpe.json lookup now uses the operating system's home directory resolution, so it also works on Windows and fails with a clear error when no home directory exists.

## [1.3.5] - 2026-07-06

### Added

- **Free Form Query entry types complete** - The entry type picklist now covers all SMP/E CSI entry types: added `HFS`, `SHELLSCR`, the `ELEMENT` pseudo-entry and all data element types (BOOK, CLIST, EXEC, MSG, PARM, PROC, SAMP, USER1-USER5 and more), each with its valid subentries (46 → 86 entry types).
- **National language variants** - Entry types with a language suffix (e.g. `HFSESP`, `MSGENU`) automatically resolve to the subentries of their base type.

### Fixed

- **Free Form Query HFS entries** - `HFS` was missing from the entry type picklist, so its subentries could not be selected.

## [1.3.4] - 2026-07-01

### Fixed

- **Free Form Query subentry picker** - Already selected subentries now show a checkmark again when the picker is reopened, and newly picked subentries are merged into the list in alphabetical order instead of being appended to the end.

## [1.3.3] - 2026-06-02

### Added

- **Signature Help** - The language server now implements `textDocument/signatureHelp`. When the cursor is inside an operand's parentheses (`DISTLIB(│)`), a floating box shows the expected parameter, a short description and the type, sourced from `smpe.json`. It appears automatically while typing `(` and after accepting an operand from the completion list. Boolean flag operands (no parameter) show no box. Toggle with the new setting `smpe.signatureHelp.enabled` (default `true`).

### Fixed

- **Fix language discrepancies** - Free Form Query now has only english button labels

## [1.3.2] - 2026-05-31

### Added

- **Code Actions (Quick Fixes)** - The language server now implements `textDocument/codeAction`. The editor lightbulb (`Cmd+.` / `Ctrl+.`) offers one-click fixes for diagnostics:
  - **Add statement terminator** - inserts the missing `.` for a statement without a terminator.
  - **Insert operand X** / **Insert all required operands** - inserts a skeleton (e.g. `SOURCEID()`) for each missing required operand; when two or more are missing, an additional aggregate action inserts them all at once.
  - **Set REWORK to current date** - fills an empty `REWORK()` operand with the current Julian date in `yyyyddd` format, inserted between the existing parentheses.

## [1.3.1] - 2026-05-20

### Added

- **Training material bundled with the extension** - The full training (DE and EN, 11 modules) is now shipped inside the VSIX. A new command `SMP/E: Open Training` opens a module picker; the selected module renders in VSCode's built-in Markdown preview. Language is auto-detected from the VSCode UI language (`vscode.env.language`); falls back to English.

### Fixed

- **MCS completion menu stays open while typing `++STATEMENT` prefix** - Typing `++S`, `++SR`, `++SRC`, … no longer dismisses the completion list. The completion menu remains open and continues to filter MCS statements as more characters are typed.
- **Snippet completion items carry `filterText`** - Boilerplate snippet items now include an explicit `filterText` so VSCode and blink-cmp (Neovim) apply their prefix filter correctly. Snippets are no longer hidden when typing `++P`, `++PT`, etc.
- **Saved Query subentries now reflected in subentry picker** - Loading a saved query in the Free Form Query panel now rebuilds the subentry checkbox grid based on the stored subentries. Previously the picker remained empty (or showed stale defaults) because programmatic value changes did not trigger the `input` event listener on the entry type field.

## [1.3.0] - 2026-05-13

### Added

- **Saved Queries in Free Form Query** - Save and reuse complete CSI queries directly from the Free Form Query panel. Queries are persisted in `.smpe-saved-queries.yaml` in the workspace root. A collapsible saved queries section appears below the input form with load and delete actions.
- **`smpe.editor.autoDetectLanguage` setting** - New boolean setting (default: `true`) to control whether the extension automatically sets the SMP/E language mode based on file content or z/OS dataset LLQ. Set to `false` to allow manual language mode overrides (e.g. switching a `.smpe` buffer to REXX) without the extension reverting the change.
- **`SMP/E: Toggle Auto-Detect Language Mode` command** - New command in the Command Palette to toggle `smpe.editor.autoDetectLanguage` on/off without opening settings.
