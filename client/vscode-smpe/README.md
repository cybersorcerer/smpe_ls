# SMP/E Language Support for VS Code

Language Server Extension for IBM SMP/E MCS (Modification Control Statements).

## Features

- **Syntax Highlighting** - Color highlighting for SMP/E statements
- **Code Completion** - Context-sensitive completion for MCS statements and operands
- **Diagnostics** - Real-time validation with error and warning messages
- **Hover Information** - Documentation when hovering over statements and operands

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
| Windows x64 | `vscode-smpe-win32-x64-0.7.0.vsix` |
| Windows ARM64 | `vscode-smpe-win32-arm64-0.7.0.vsix` |
| macOS Apple Silicon | `vscode-smpe-darwin-arm64-0.7.0.vsix` |
| macOS Intel | `vscode-smpe-darwin-x64-0.7.0.vsix` |
| Linux x64 | `vscode-smpe-linux-x64-0.7.0.vsix` |
| Linux ARM64 | `vscode-smpe-linux-arm64-0.7.0.vsix` |

### Installation in VS Code

1. Open VS Code
2. `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)
3. Select "Extensions: Install from VSIX..."
4. Choose the downloaded `.vsix` file

Alternatively via terminal:

```bash
code --install-extension vscode-smpe-darwin-arm64-0.7.0.vsix
```

The Language Server is already included in the extension - no additional installation required.

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.debug` | `false` | Enable debug logging |

## File Extensions

The extension activates automatically for files with the following extensions:

- `.smpe`
- `.mcs`
- `.smp`

## Screenshots

**Coming soon**

## Known Limitations

- This is an alpha version
- Not all SMP/E statements are fully implemented

## License

AGPL-3.0 - See [LICENSE](LICENSE) for details.

## Author

Ronald Funk

---

**Note:** SMP/E is a registered trademark of IBM Corporation.
