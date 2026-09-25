# Changelog

All notable changes to this project are documented in this file.

## [1.3.17] - 2026-09-25

### Added

- **A comment is closed when it is opened** - Typing `/*` followed by a space appends the closing marker on the same line and leaves the cursor between the two, so you can write straight on:

  ```smpe
  ++USERMOD(U1) /* | */
  ```

  It is left out in three cases: when the marker would come to rest beyond column 72, where SMP/E stops reading and a marker would only look as if it closed the comment; when the line already carries one further to the right; and in inline data, where a `/*` opens a REXX program and belongs to the element rather than to the MCS text. The insertion rides on `editor.formatOnType`, which the extension now turns on for `.smpe` by default - switching it off in your own settings switches the closing off as well.

### Fixed

- **Accepting a suggestion replaces what was typed** - Typing `s` and accepting `SUP` produced `sSUP`, while `S` worked. Suggestions now state which text they replace instead of leaving it to the editor, whose fallback only recognises the typed fragment when it matches the item case-sensitively. Filtering is case insensitive, so the item was offered either way and only the replacement went wrong. Operand values were affected in the same manner - `CLASS(e` followed by `ERREL` would have produced `eERREL`. Shipped in 1.3.16.

## [1.3.16] - 2026-09-24

### Fixed

- **No completions inside comments** - Writing prose in a `/* ... */` comment no longer opens the suggestion list, and every space no longer reopens it. This covers a finished comment as well as one still being typed, where no closing `*/` exists yet and the parser therefore records nothing. Element data is unaffected: a `/* REXX */` opening an inline program is not an MCS comment, and the check keeps them apart.
- **Multiline parameter values keep their highlighting** - A value written across lines lost its colour entirely:

  ```
  SUP(
      LBCP034
  )
  ```

  A semantic token cannot span a line break - the protocol encodes it as a line plus a column and a length - so the token covering everything between the parentheses was malformed and the editor dropped it. Such a value now gets one token per line, on the value itself rather than the indentation. Lists were unaffected because they are already broken down into one node per element.
- **SYSMOD and DDDEF results show every requested subentry** - The queries behind the CodeLens asked for fifteen respectively nine subentries but displayed only a handful; the rest were fetched and thrown away. Table columns, CSV and JSON export now follow the requested list, the way the Free Form Query has always built its table. Values too long for a column are reachable through the cell tooltip, in both result views.

### Changed

- **The space bar no longer opens the completion list** - Indenting a continuation line or separating operands stays quiet. The list opens once an operand name is actually begun, which every letter triggers - upper and lower case alike, since the offered names are upper case either way. `+` still opens the statement list and Ctrl+Space still asks explicitly.
- **SYSMOD and DDDEF queries use the Free Form defaults** - Both now ask for the same subentries the Free Form Query offers for that entry type, so the two ways of querying a CSI return the same fields.

## [1.3.15] - 2026-09-19

### Fixed

- **Statements carrying element data were not marked as expecting inline data** - `++CLIST`, `++DATA`, `++DATA1` to `++DATA5`, `++ZAP`, `++PROGRAM`, `++JAR` and `++JARUPD` all carry element data, but `smpe.json` did not mark them with `inline_data`. Nothing that depends on that flag applied to them: the `missingInlineData` diagnostic never fired, the formatter treated their data lines as statement text, and the comment-in-column-1 and standalone-comment checks reported the data as if it were MCS text. `++ZAP` was the clearest case - its IMASPZAP control statements can only follow inline, since it names no external source at all. `++DATA1` to `++DATA5` were additionally skipped by Check Missing Input Members, which knew no extension for them.
- **`++ZAP` now names what it is missing** - It is the one statement with no `FROMDS`, `RELFILE`, `TXLIB` or `LKLIB` to offer, so the diagnostic no longer suggests an alternative that does not exist. It states what the SMP/E reference states: the IMASPZAP control statements follow the `++ZAP` MCS immediately.
- **Numbered statement families keep their digit in the expected file name** - `++AIX1` now expects `<element>.aix1`. The extension was derived with the trailing digits removed, so `++AIX1` through `++AIX5` all expected `<element>.aix`, and one file answered for all five. The same applied to `++CLIENT1`-`5`, `++OS21`-`25`, `++UNIX1`-`5`, `++USER1`-`5` and `++WIN1`-`5`. `++DATA1` to `++DATA6` keep sharing `.data` and are listed explicitly.
- **A comment before an operand hid the element source** - In `++MAC(A) /* c */ FROMDS(DSN(X)) .` the operand was reported without its parameter, so Check Missing Input Members did not recognize `FROMDS` as supplying the data and demanded an input member. Operand names are now matched on their own.
- **`make install` no longer kills a running server** - Copying over a binary that is currently running invalidates the pages macOS has mapped, and the process dies with SIGKILL. The target is removed before the copy.

### Changed

- **Input member file extensions moved into `smpe.json`** - The mapping of statements whose extension deviates from the default rule now lives in the `file_ext` object rather than in the extension's source, so it is maintained in one place with the rest of the statement data.
- **The operands that replace inline data moved into `smpe.json`** - `FROMDS`, `RELFILE`, `TXLIB` and `LKLIB` are listed once under `element_source`. The rule was spelled out as a hardcoded list in eight places, which is how `LKLIB` came to be missing from some of them before 1.3.13. Adding another source operand is now a single edit.

## [1.3.14] - 2026-09-16

### Fixed

- **A dot in a multi-line comment ended the statement** - The standalone-comment check had its own terminator search that stripped comments only within a single line, so a `.` inside a `/* ... */` block (a sentence ending in a period, for instance) closed the statement and the comment that followed was reported as standing between MCS statements. It now shares `GetStatementEndPosition` with the Outline. That shared search had two flaws of its own, fixed here as well: a `++` anywhere in a line ended it, so `++APAR` inside a comment cut the statement short - only a `++` as the first non-blank character outside a block comment counts now, checked before the line is scanned so the next statement's terminator is never mistaken for this one's. And with unbalanced parentheses the depth made the terminator unfindable, producing follow-up diagnostics on a statement that is already reported as malformed.
- **Completion continued statements that were already finished** - An indented line after a terminated statement offered that statement's operands, and typing `++` there did not switch to the statement list because the line still counted as a continuation. A line now continues a statement only while that statement is open.
- **Typing a space no longer pops up the statement list** - The server evaluates the LSP trigger kind: a space on a free line stays quiet, while an explicit request still lists every statement. Inside an open statement a space keeps triggering operand completion, and `+` opens the statement list as before.
- **Completion never answers with `null`** - A JSON null result makes the client treat the response as malformed. Code actions and diagnostics already guarded against this; completion now does too.

## [1.3.13] - 2026-09-11

### Fixed

- **A dot in an operand value ended the statement in the Outline** - The symbol range of a statement stopped at the first `.` outside a block comment, so a free-text `DESC(R+V IIQ. SMF Exit 83)`, a dotted dataset name or a quoted value cut the range short at that operand instead of reaching the terminator. `++USERMOD(LIIQ101)` ended at the dot inside `DESC` rather than on the terminator line, truncating the Outline view, folding ranges, breadcrumbs, workspace symbols and `smpe_outl` output. Terminator detection now tracks parenthesis depth and single-quoted strings as well, and scans from the statement's first line so an operand opening its parenthesis on an earlier line is accounted for. The 1.3.8 fix for dots inside comments is unaffected and covered by a regression test.

## [1.3.12] - 2026-09-11

### Fixed

- **`LKLIB` counts as an external data source** - Per the ++MOD usage notes a module in the data set named by the `LKLIB` ddname is not packaged inline; `LKLIB` is mutually exclusive with inline packaging just like `FROMDS`, `RELFILE` and `TXLIB`. It was missing from every check that decides whether a statement expects inline data, so `++MOD(X) DISTLIB(Y) LKLIB(DD1) .` was reported as missing its inline data. This affected the `missingInlineData` diagnostic, the column 72 and standalone comment checks, the comment-in-column-1 check, the formatter, and Check Missing Input Members, which demanded an input member for such a statement. `LKLIB` now also appears among the alternatives listed in the diagnostic message. The `mutuallyExclusive` diagnostic was already correct, since it reads the relations from `smpe.json`.

## [1.3.11] - 2026-09-01

### Fixed

- **`commentInColumn1` no longer fires inside inline data** - Inline data is not MCS text and is now excluded from the check. In JCLIN a `/*` in column 1 is the regular JCL delimiter closing a `//... DD *` stream and has to sit exactly there, so the diagnostic introduced in 1.3.10 reported it as an error. Element data is passed through verbatim as well, so only the statement regions are inspected. A `/*` in column 1 within a statement region is still reported, including in files that contain inline data further down.

## [1.3.10] - 2026-08-28

### Added

- **Check Missing Input Members resolves `{{ path }}` placeholders** - A statement can point at its input member with a `{{ ./path }}` line in its inline data area, which the build pipeline replaces with the contents of that file. The path is relative to the repository root and is checked exactly as written, so it may point outside the configured search folders. Such statements were skipped entirely before, because a placeholder line counts as inline data. Results now carry a **Source** column telling the two checks apart: `placeholder` for a resolved `{{ path }}`, `convention` for the established `<element name><extension>` lookup below `smpe.checkMissingInputMembers.searchFolders`.
- **New diagnostic: comment beginning in column 1** - A `/*` in column 1 marks the end of an input data set, so SMP/E stops reading the member at that line. Reported as an error on every affected line, including continuation lines of a block comment and lines inside inline data, where the same truncation applies. Two quick fixes are offered: indent only the reported line, or shift the whole comment block by the same amount so a box drawing keeps its layout. Toggle with `smpe.diagnostics.commentInColumn1` (default `true`).

### Fixed

- **The formatter no longer rewrites comment text** - Comments were reflowed, re-indented and re-wrapped at column 72, which destroyed box drawings, tables and column-aligned metadata blocks such as generated GITLAB-META headers. The formatter now decides only where a comment sits, never what it says: the text is reproduced exactly as written, including its original indentation. Lines past column 72 and comments in column 1 are reported by diagnostics instead of being silently rewritten. The one exception is a comment the formatter relocates itself (`smpe.formatting.moveLeadingComments`), which is shifted as a whole block so it does not end up in column 1.
- **Comments after operand values containing dots are kept** - A dot inside an operand value, such as the dataset name in `FROMDS(DSN(HLQ.MID.LLQ))` or a quoted `DESC('R+V IIQ. Started')`, was treated as the statement terminator, so a comment following that operand was classified as a post-terminator comment and dropped from the formatted output. Terminator detection now tracks parenthesis depth, quoted strings and comments, and works across line breaks.
- **Formatting no longer reaches into inline data** - An unclosed comment after the terminator made the formatter search on for `*/` into the following lines, pulling inline data into the statement text. For statements expecting inline data the scan now stops at the terminator line.
- **No input member demanded for FROMDS, RELFILE or DELETE** - These operands mean the element data comes from elsewhere or the element is being deleted, so no member file is expected. Only `TXLIB` was excluded before, making every such statement a false positive.
- **Input member extensions for all statements** - Statements without an explicit mapping now derive their extension from the statement name (`++BOOK` expects `<element>.book`), provided `smpe.json` marks them as expecting inline data. Language variants resolve to their base statement, so `++PNLDEU` expects `<element>.pnl` and `++SKLDEU` keeps `.skl.jcl`. Three entries never matched a statement and were corrected: `++SHELLSRC` was a typo for `++SHELLSCR` (its `.sh` mapping is preserved), `++SHSCRIPT` is an operand rather than a statement, and `++PNLENU` covered exactly one of 32 languages. Existing mappings keep their extensions unchanged.
- **`smpe_outl` reports parameterless flag operands** - The sub-operand fallback searched forward for the next `(`, so an operand without a value absorbed the value of a later one (`DELETE DISTLIB(AMACLIB)` became `DELETE(AMACLIB)`) or vanished when no parenthesized operand followed. Operands such as `DELETE` and `USER` now appear under their bare name.
- **CHANGELOG link in the extension README** - Relative links in the extension README resolve against the repository root rather than `client/vscode-smpe/`, so the link returned a 404.

### Changed

- **Language definitions come from `smpe.json`** - The 32 national language identifiers were hardcoded in Go, and the base statements accepting a language suffix existed twice: in Go and as `language_variants` in `smpe.json`. Both now live in `smpe.json` alone. `++HFS` was missing its `language_variants` flag although its own description names `++HFSxxx`.

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
