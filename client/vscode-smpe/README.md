# SMP/E MCS Language Server for VS Code

Language Server Extension for IBM SMP/E MCS (Modification Control Statements).

## What's New in 0.9.6

- **Training Documentation** - 22 Markdown training modules (DE + EN) covering installation, all VSCode extension features, CLI tools (`smpe_lint`, `smpe_outl`), and z/OSMF integration. Available under `docs/training/`.

## What's New in 0.9.5

- **Check Missing Input Members** - New command (`SMP/E: Check Missing Input Members`) checks whether all referenced input member files exist in the workspace. Results displayed in a filterable/sortable Webview table. Available in Command Palette and via right-click on `.smpe` files.
- **`smpe.checkMissingInputMembers.searchFolders`** - Configures which folders are searched for input member files (default: `["customization"]`). Use `"*"` to search the entire workspace.

## What's New in 0.9.4

- **z/OSMF Config Lookup** - `.smpe-zosmf.yaml` is now resolved from workspace folders first, then `~/.config/smpe_ls/` as a global fallback
- **LLQ-based Language Detection** - z/OS datasets with configurable last level qualifiers (e.g. `.MCS`) are automatically recognized as SMP/E files (`smpe.zosDatasetsLlq` setting)

## What's New in 0.9.3

- **Workspace Symbols** - Search for SYSMOD definitions across all `.smpe` files (`Cmd+T`)
- **Dataset Member Attributes** - PDS member listing now shows full ISPF-style attributes (User, Created, Modified, Ver, Mod)

See the [CHANGELOG](https://github.com/cybersorcerer/smpe_ls/blob/main/client/vscode-smpe/CHANGELOG.md) for full details.

## Features

- **Syntax Highlighting** - Color highlighting for SMP/E statements
- **Code Completion** - Context-sensitive completion for MCS statements and operands
- **Diagnostics** - Real-time validation with error and warning messages
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
- **Document Formatting** - Auto-format SMP/E statements

## Supported Statements

The extension supports all common SMP/E MCS statements, including:

- `++APAR`, `++PTF`, `++USERMOD`, `++FUNCTION`
- `++MAC`, `++MACUPD`, `++MOD`, `++SRC`, `++SRCUPD`
- `++JCLIN`, `++JAR`, `++JARUPD`
- `++VER`, `++ZAP`, `++DELETE`
- and many more...

## Installation

### Download

Download the appropriate `.vsix` file for your platform from the [Release](https://github.com/cybersorcerer/smpe_ls/releases) page:

| Platform | File |
|----------|------|
| Windows x64 | `smpe-mcs-language-server-win32-x64-0.9.6.vsix` |
| Windows ARM64 | `smpe-mcs-language-server-win32-arm64-0.9.6.vsix` |
| macOS Apple Silicon | `smpe-mcs-language-server-darwin-arm64-0.9.6.vsix` |
| macOS Intel | `smpe-mcs-language-server-darwin-x64-0.9.6.vsix` |
| Linux x64 | `smpe-mcs-language-server-linux-x64-0.9.6.vsix` |
| Linux ARM64 | `smpe-mcs-language-server-linux-arm64-0.9.6.vsix` |

### Installation in VS Code

1. Open VS Code
2. `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)
3. Select "Extensions: Install from VSIX..."
4. Choose the downloaded `.vsix` file

Alternatively via terminal:

```bash
code --install-extension smpe-mcs-language-server-darwin-arm64-0.9.6.vsix
```

The Language Server is already included in the extension - no additional installation required.

## Configuration

### General

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.serverPath` | `""` | Path to the smpe_ls executable (uses bundled server if empty) |
| `smpe.dataPath` | `""` | Path to the smpe.json data file (uses bundled file if empty) |
| `smpe.debug` | `false` | Enable debug logging |

### Formatting

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.formatting.enabled` | `true` | Enable document formatting |
| `smpe.formatting.indentContinuation` | `3` | Spaces for continuation lines |
| `smpe.formatting.oneOperandPerLine` | `true` | Place each operand on its own line |
| `smpe.formatting.wrapListsAfterN` | `2` | Wrap comma-separated lists after N items per line (0 = disabled) |
| `smpe.formatting.formatOnSave` | `false` | Automatically format document when saving |

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

| Setting                           | Default   | Description                                                                 |
|-----------------------------------|-----------|-----------------------------------------------------------------------------|
| `smpe.zosmf.queryTimeoutSeconds`  | `300`     | Maximum wait time for z/OSMF CSI queries (30–600s)                          |
| `smpe.zosDatasetsLlq`             | `["MCS"]` | z/OS dataset last level qualifiers that trigger SMP/E language activation   |

### Check Missing Input Members

| Setting                                       | Default             | Description                                                                                             |
|-----------------------------------------------|---------------------|---------------------------------------------------------------------------------------------------------|
| `smpe.checkMissingInputMembers.searchFolders` | `["customization"]` | Folders (relative to workspace root) to search for input member files. Use `"*"` for entire workspace.  |

## File Extensions

The extension activates automatically for files with the following extensions:

- `.smpe`
- `.mcs`
- `.smp`

## Screenshots

**Coming soon**

## Known Limitations

- This is an alpha version
- There might be bugs

## License

AGPL-3.0 - See [LICENSE](LICENSE) for details.

## Author

**Made with ❤️ by Sir Tobi aka Cybersorcerer**

---

**Note:** SMP/E is a registered trademark of IBM Corporation.
