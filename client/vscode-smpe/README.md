# SMP/E Language Support for VS Code

Language Server Extension für IBM SMP/E MCS (Modification Control Statements).

## Features

- **Syntax Highlighting** - Farbliche Hervorhebung von SMP/E Statements
- **Code Completion** - Kontextsensitive Vervollständigung für MCS Statements und Operanden
- **Diagnostics** - Echtzeit-Validierung mit Fehler- und Warnmeldungen
- **Hover Information** - Dokumentation beim Überfahren von Statements und Operanden

## Unterstützte Statements

Die Extension unterstützt alle gängigen SMP/E MCS Statements, darunter:

- `++APAR`, `++PTF`, `++USERMOD`, `++FUNCTION`
- `++MAC`, `++MACUPD`, `++MOD`, `++SRC`, `++SRCUPD`
- `++JCLIN`, `++JAR`, `++JARUPD`
- `++VER`, `++ZAP`, `++DELETE`
- und viele mehr...

## Installation

### Download

Lade die passende `.vsix`-Datei für deine Plattform aus dem [Release](https://github.com/cybersorcerer/smpe_ls/releases) herunter:

| Plattform | Datei |
|-----------|-------|
| Windows x64 | `vscode-smpe-win32-x64-0.7.0.vsix` |
| Windows ARM64 | `vscode-smpe-win32-arm64-0.7.0.vsix` |
| macOS Apple Silicon | `vscode-smpe-darwin-arm64-0.7.0.vsix` |
| macOS Intel | `vscode-smpe-darwin-x64-0.7.0.vsix` |
| Linux x64 | `vscode-smpe-linux-x64-0.7.0.vsix` |
| Linux ARM64 | `vscode-smpe-linux-arm64-0.7.0.vsix` |

### Installation in VS Code

1. Öffne VS Code
2. `Cmd+Shift+P` (macOS) oder `Ctrl+Shift+P` (Windows/Linux)
3. Wähle "Extensions: Install from VSIX..."
4. Wähle die heruntergeladene `.vsix`-Datei

Alternativ via Terminal:

```bash
code --install-extension vscode-smpe-darwin-arm64-0.7.0.vsix
```

Der Language Server ist bereits in der Extension enthalten - keine weitere Installation notwendig.

## Konfiguration

| Einstellung | Standard | Beschreibung |
|-------------|----------|--------------|
| `smpe.debug` | `false` | Debug-Logging aktivieren |

## Dateiendungen

Die Extension aktiviert sich automatisch für Dateien mit folgenden Endungen:

- `.smpe`
- `.mcs`
- `.smp`

## Screenshots

**Coming soon**

## Bekannte Einschränkungen

- Dies ist eine Alpha-Version
- Nicht alle SMP/E Statements sind vollständig implementiert

## Lizenz

AGPL-3.0 - Siehe [LICENSE](LICENSE) für Details.

## Autor

Ronald Funk

---

**Hinweis:** SMP/E ist ein eingetragenes Warenzeichen von IBM Corporation.
