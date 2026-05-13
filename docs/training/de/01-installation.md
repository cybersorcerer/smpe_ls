# Modul 01 — Installation der VSCode Extension

## Überblick

Dieses Modul zeigt wie die SMP/E Language Server Extension in Visual Studio Code
installiert wird. Nach der Installation erkennt VSCode `.smpe`-Dateien automatisch
und stellt alle Language-Server-Features bereit.

## Voraussetzungen

- Visual Studio Code ab Version 1.75 installiert
- VSIX-Datei für die eigene Plattform heruntergeladen (siehe unten)

## VSIX-Datei herunterladen

Die passende VSIX-Datei befindet sich auf der
[GitHub Releases Seite](https://github.com/cybersorcerer/smpe_ls/releases/latest):

| Plattform | Datei |
|-----------|-------|
| Windows x64 | `smpe-mcs-language-server-win32-x64-1.3.0.vsix` |
| Windows ARM64 | `smpe-mcs-language-server-win32-arm64-1.3.0.vsix` |
| macOS Apple Silicon | `smpe-mcs-language-server-darwin-arm64-1.3.0.vsix` |
| macOS Intel | `smpe-mcs-language-server-darwin-x64-1.3.0.vsix` |
| Linux x64 | `smpe-mcs-language-server-linux-x64-1.3.0.vsix` |
| Linux ARM64 | `smpe-mcs-language-server-linux-arm64-1.3.0.vsix` |

Der Language Server ist bereits in der Extension enthalten — keine separate
Installation notwendig.

## Installation über die VSCode-Oberfläche

1. VSCode öffnen
2. `Cmd+Shift+P` (macOS) bzw. `Ctrl+Shift+P` (Windows/Linux)
3. `Extensions: Install from VSIX...` eingeben und bestätigen
4. Die heruntergeladene `.vsix`-Datei auswählen

## Installation über das Terminal

```bash
code --install-extension smpe-mcs-language-server-darwin-arm64-1.3.0.vsix
```

Danach VSCode neu laden.

## Installation prüfen

Nach dem Laden erscheint im Output-Panel eine Bestätigungsmeldung:

1. `View` → `Output` öffnen
2. Rechts im Dropdown `SMP/E Language Server` auswählen
3. Dort erscheint beim Start:

```
[2026-04-09T...] SMP/E Language Server extension activating...
[2026-04-09T...] Starting server: .../smpe_ls
```

## Wichtige Einstellungen

Diese Einstellungen können in der VSCode-Einstellungsdatei (`.vscode/settings.json`)
konfiguriert werden:

| Einstellung | Standard | Beschreibung |
|-------------|----------|--------------|
| `smpe.serverPath` | `""` | Pfad zur `smpe_ls`-Binary (leer = bundled) |
| `smpe.dataPath` | `""` | Pfad zur `smpe.json`-Datei (leer = bundled) |
| `smpe.outlPath` | `""` | Pfad zur `smpe_outl`-Binary (leer = bundled) |
| `smpe.debug` | `false` | Debug-Logging aktivieren |

## Zusammenfassung

- Die korrekte VSIX-Datei für die eigene Plattform herunterladen
- Installation über VSCode-UI oder Terminal
- Nach der Installation **Reload Window** nicht vergessen
- Installation über den Output-Panel `SMP/E Language Server` prüfen
