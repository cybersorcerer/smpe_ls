# SMP/E MCS Language Server for VS Code

Language Server Extension for IBM SMP/E MCS (Modification Control Statements).

Release notes for every version are in the
[CHANGELOG](https://github.com/cybersorcerer/smpe_ls/blob/main/client/vscode-smpe/CHANGELOG.md),
or in the Changelog tab next to this page in VS Code.

## Features

- **Syntax Highlighting** - Color highlighting for SMP/E statements
- **Code Completion** - Context-sensitive completion for MCS statements and operands
- **Diagnostics** - Real-time validation with error and warning messages
- **Code Actions (Quick Fixes)** - Lightbulb fixes for missing terminators, missing required operands, and empty `REWORK()`
- **Signature Help** - Floating parameter hint for operands with description and type (toggle via `smpe.signatureHelp.enabled`)
- **Hover Information** - Documentation when hovering over statements and operands
- **Go to Definition** - Navigate to SYSMOD/FMID definitions (`F12` or `Cmd+Click`)
- **Find References** - Find all references to a SYSMOD or FMID (`Shift+F12`)
- **Document Symbols** - Outline view and quick navigation (`Cmd+Shift+O`)
- **Workspace Symbols** - Search for SYSMOD definitions across all `.smpe` files (`Cmd+T`)
- **Folding Ranges** - Collapse/expand MCS statements and multi-line comments
- **Check Missing Input Members** - Scan workspace for missing MCS input files with filterable Webview results
- **CodeLens** - Inline z/OSMF CSI queries for SYSMODs and DDDEFs
- **z/OSMF Integration** - Query CSI, browse USS directories and MVS datasets
- **Column Rulers** - Visual guides at columns 72 and 80 (mainframe card boundaries)
- **Document Formatting** - Auto-format SMP/E statements, and close a comment the moment it is opened

## Supported Statements

The extension supports all common SMP/E MCS statements, including:

- `++APAR`, `++PTF`, `++USERMOD`, `++FUNCTION`
- `++MAC`, `++MACUPD`, `++MOD`, `++SRC`, `++SRCUPD`
- `++JCLIN`, `++JAR`, `++JARUPD`
- `++VER`, `++ZAP`, `++DELETE`
- and many more...

## Configuration

### General

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.serverPath` | `""` | Path to the smpe_ls executable (uses bundled server if empty) |
| `smpe.dataPath` | `""` | Path to the smpe.json data file (uses bundled file if empty) |
| `smpe.outlPath` | `""` | Path to the smpe_outl executable (uses bundled binary if empty) |
| `smpe.debug` | `false` | Enable debug logging |
| `smpe.signatureHelp.enabled` | `true` | Enable Signature Help for operand parameters |

### Formatting

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.formatting.enabled` | `true` | Enable document formatting |
| `smpe.formatting.indentContinuation` | `4` | Spaces for continuation lines |
| `smpe.formatting.oneOperandPerLine` | `true` | Place each operand on its own line |
| `smpe.formatting.wrapListsAfterN` | `2` | Wrap comma-separated lists after N items per line (0 = disabled) |
| `smpe.formatting.formatOnSave` | `false` | Automatically format document when saving |
| `smpe.formatting.moveLeadingComments` | `true` | Move comments from before the first statement into the statement during formatting |

### Diagnostics

All diagnostics are enabled by default. Set to `false` to disable specific checks.

| Setting | Description |
|---------|-------------|
| `smpe.diagnostics.unknownStatement` | Report unknown statement types |
| `smpe.diagnostics.invalidLanguageId` | Report invalid 3-character language identifiers |
| `smpe.diagnostics.unbalancedParentheses` | Report unbalanced parentheses |
| `smpe.diagnostics.missingTerminator` | Report missing statement terminators (`.`) |
| `smpe.diagnostics.missingParameter` | Report missing required statement parameters |
| `smpe.diagnostics.unknownOperand` | Report unknown operands |
| `smpe.diagnostics.duplicateOperand` | Report duplicate operands |
| `smpe.diagnostics.emptyOperandParameter` | Report empty operand parameters |
| `smpe.diagnostics.missingRequiredOperand` | Report missing required operands |
| `smpe.diagnostics.dependencyViolation` | Report operand dependency violations |
| `smpe.diagnostics.mutuallyExclusive` | Report mutually exclusive operand conflicts |
| `smpe.diagnostics.requiredGroup` | Report missing required group operands |
| `smpe.diagnostics.missingInlineData` | Report missing inline data |
| `smpe.diagnostics.unknownSubOperand` | Report unknown sub-operands |
| `smpe.diagnostics.subOperandValidation` | Report sub-operand validation errors |
| `smpe.diagnostics.contentBeyondColumn72` | Report content that extends beyond column 72 |
| `smpe.diagnostics.standaloneCommentBetweenMCS` | Report standalone comments between MCS statements (causes SMP/E syntax error) |
| `smpe.diagnostics.commentInColumn1` | Report comments beginning in column 1 (a `/*` in column 1 indicates the end of an input data set); inline data is not checked |

### z/OSMF Integration

The z/OSMF integration requires a `.smpe-zosmf.yaml` configuration file. The extension
resolves the file in the following order:

1. **Workspace folders** — all open workspace folders are searched, first match wins
2. **Global fallback** — `~/.config/smpe_ls/.smpe-zosmf.yaml`

To create a configuration file in the workspace root, run:

`Ctrl+Shift+P` → **SMP/E: Create z/OSMF Config**

This creates `.smpe-zosmf.yaml` in the root of the first workspace folder and opens it
for editing. If you want a shared configuration across all projects, place the file at
`~/.config/smpe_ls/.smpe-zosmf.yaml` instead.

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.zosmf.queryTimeoutSeconds` | `300` | Maximum wait time for z/OSMF CSI queries (30–600s) |
| `smpe.zosDatasetsLlq` | `["MCS"]` | z/OS dataset last level qualifiers that trigger SMP/E language activation |

### Check Missing Input Members

| Setting                                       | Default             | Description                                                                                            |
|-----------------------------------------------|---------------------|--------------------------------------------------------------------------------------------------------|
| `smpe.checkMissingInputMembers.searchFolders` | `["customization"]` | Folders (relative to workspace root) to search for input member files. Use `"*"` for entire workspace. |

### Editor

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.editor.showColumnRulers` | `true` | Show column rulers at positions 72 and 80 for SMP/E files |
| `smpe.editor.autoDetectLanguage` | `true` | Automatically detect and set the SMP/E language mode for files with `++` statements or matching z/OS dataset LLQs. Disable to allow manual language mode overrides. |

## Commands

All commands are available via the Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`).

| Command | Description |
|---------|-------------|
| `SMP/E: Query SYSMOD via z/OSMF` | Query a SYSMOD against a CSI |
| `SMP/E: Query DDDEF via z/OSMF` | Query a DDDEF against a CSI |
| `SMP/E: List Zones via z/OSMF` | List zones of a CSI |
| `SMP/E: Free Form CSI Query` | Open the free form CSI query panel (saved queries + filter history) |
| `SMP/E: Create z/OSMF Config` | Create a `.smpe-zosmf.yaml` config file in the workspace root |
| `SMP/E: Clear Stored Password` | Clear the stored z/OSMF password |
| `SMP/E: Manage Filter History` | Edit, delete, or clear saved filter entries |
| `SMP/E: Check Missing Input Members` | Scan the workspace for missing MCS input member files |
| `SMP/E: Toggle Auto-Detect Language Mode` | Toggle the `smpe.editor.autoDetectLanguage` setting |
| `SMP/E: Open Training` | Open the bundled training documentation |
| `SMP/E: Query SYSMOD (CodeLens)` | Internal command triggered by the SYSMOD CodeLens link above `++PTF`/`++APAR`/etc. |
| `SMP/E: Query DDDEF (CodeLens)` | Internal command triggered by the DDDEF CodeLens link above `DISTLIB`/`SYSLIB`/etc. |

## File Extensions

The extension activates automatically for files with the following extensions:

- `.smpe`
- `.mcs`
- `.smp`

## Screenshots

**Coming soon**

## License

AGPL-3.0 - See [LICENSE](https://github.com/cybersorcerer/smpe_ls/blob/main/client/vscode-smpe/LICENSE) for details.

## Author

**Made with ❤️ by Sir Tobi aka Cybersorcerer**

---

**Note:** SMP/E is a registered trademark of IBM Corporation.
