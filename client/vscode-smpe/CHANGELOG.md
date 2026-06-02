# Changelog

All notable changes to this project are documented in this file.

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
