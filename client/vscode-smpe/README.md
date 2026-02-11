# SMP/E Language Support for VS Code

Language Server Extension for IBM SMP/E MCS (Modification Control Statements).

## What's New in 0.8.3

- **z/OSMF Free Form CSI Query** - New Webview with input form and dynamic result table for free-form z/OSMF CSI queries
- **Zone Pattern Matching** - Wildcard support (`*`, `?`) for zone parameters in z/OSMF queries
- **List-type Parameter Fix** - `PRE`, `REQ`, `SUP` with multiple values no longer produce false warnings

See the [CHANGELOG](CHANGELOG.md) for full details.

## Features

- **Syntax Highlighting** - Color highlighting for SMP/E statements
- **Code Completion** - Context-sensitive completion for MCS statements and operands
- **Diagnostics** - Real-time validation with error and warning messages
- **Hover Information** - Documentation when hovering over statements and operands
- **Go to Definition** - Navigate to SYSMOD/FMID definitions (`F12` or `Cmd+Click`)
- **Find References** - Find all references to a SYSMOD or FMID (`Shift+F12`)
- **Document Symbols** - Outline view and quick navigation (`Cmd+Shift+O`)
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
| Windows x64 | `vscode-smpe-win32-x64-0.8.3.vsix` |
| Windows ARM64 | `vscode-smpe-win32-arm64-0.8.3.vsix` |
| macOS Apple Silicon | `vscode-smpe-darwin-arm64-0.8.3.vsix` |
| macOS Intel | `vscode-smpe-darwin-x64-0.8.3.vsix` |
| Linux x64 | `vscode-smpe-linux-x64-0.8.3.vsix` |
| Linux ARM64 | `vscode-smpe-linux-arm64-0.8.3.vsix` |

### Installation in VS Code

1. Open VS Code
2. `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)
3. Select "Extensions: Install from VSIX..."
4. Choose the downloaded `.vsix` file

Alternatively via terminal:

```bash
code --install-extension vscode-smpe-darwin-arm64-0.8.3.vsix
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
