---
marp: true
theme: default
class: invert
paginate: true
footer: 'Mainframe TechTalk &nbsp;|&nbsp; SMP/E MCS Language Server &nbsp;|&nbsp; Ronny Funk &nbsp;|&nbsp; 2026'
style: |
  section {
    font-family: 'Segoe UI', 'IBM Plex Sans', Arial, sans-serif;
    background-color: #1e1e2e;
    color: #cdd6f4;
    font-size: 18px;
    padding: 40px 60px 30px 60px;
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
    align-items: flex-start;
    text-align: left;
  }
  section > * {
    width: 100%;
    text-align: left;
  }
  h1 {
    color: #89b4fa;
    border-bottom: 2px solid #313244;
    padding-bottom: 0.2em;
    margin-top: 0;
    margin-bottom: 0.6em;
    font-size: 1.5em;
    width: 100%;
  }
  h2 {
    color: #89dceb;
    font-size: 1.1em;
    margin-top: 0.5em;
    margin-bottom: 0.3em;
    width: 100%;
  }
  h3 {
    color: #a6e3a1;
    font-size: 1em;
    margin-top: 0.4em;
    margin-bottom: 0.2em;
  }
  p {
    margin: 0.3em 0;
  }
  code {
    background-color: #313244;
    color: #f38ba8;
    padding: 0.05em 0.35em;
    border-radius: 4px;
    font-size: 0.9em;
  }
  pre {
    background-color: #181825;
    border: 1px solid #313244;
    border-radius: 6px;
    padding: 0.6em 0.8em;
    margin: 0.4em 0;
    font-size: 0.78em;
    width: 100%;
    box-sizing: border-box;
  }
  pre code {
    background-color: transparent;
    color: #cdd6f4;
    padding: 0;
    font-size: 1em;
  }
  blockquote {
    border-left: 4px solid #f9e2af;
    background-color: #1e1e2e;
    color: #f9e2af;
    padding: 0.3em 0.8em;
    border-radius: 0 6px 6px 0;
    font-style: normal;
    margin: 0.4em 0;
    font-size: 0.9em;
    width: 100%;
    box-sizing: border-box;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.78em;
  }
  th {
    background-color: #313244;
    color: #89b4fa;
    padding: 0.3em 0.6em;
    text-align: left;
  }
  td {
    padding: 0.2em 0.6em;
    border-bottom: 1px solid #313244;
  }
  ul, ol {
    margin: 0.3em 0;
    padding-left: 1.4em;
  }
  ul li, ol li {
    margin-bottom: 0.2em;
  }
  footer {
    font-size: 0.65em;
    color: #6c7086;
    border-top: 1px solid #313244;
    padding-top: 4px;
    width: 100%;
  }
  section.title {
    justify-content: flex-start;
    align-items: flex-start;
    text-align: left;
  }
  section.title h1 {
    font-size: 2em;
    margin-top: 0.8em;
  }
  section.title h2 {
    font-size: 1.1em;
    color: #a6e3a1;
    font-weight: normal;
    margin-top: 0.3em;
  }
  section.title p {
    color: #6c7086;
    font-size: 0.85em;
    margin-top: 1em;
  }
---

<!-- _class: title -->

# SMP/E MCS Language Server
## Moderne Entwicklungsunterstützung für z/OS Systemprogrammierer

Version 1.3.8

Ronny Funk — Senior Mainframe Architect SVA

---

# Agenda

1. Was ist LSP?
2. Architektur-Überblick
3. Bestandteile der Extension
4. smpe_ls Language Server
5. smpe_lint — Der SMP/E Linter
6. smpe_outl — Der SMP/E Outliner
7. IntelliSense & Code Completion
8. Diagnostics & Formatierung
9. z/OSMF Integration & Commands
10. Outline View & Output Channel
11. Installation & Konfiguration
12. Settings-Referenz

---

# Was ist LSP?

## Language Server Protocol

Das **Language Server Protocol (LSP)** ist ein offener Standard von Microsoft, der die Kommunikation zwischen einem Editor und einem Sprachserver definiert.

```
┌─────────────────────┐        JSON-RPC         ┌──────────────────────┐
│   Editor / Client   │ ◄─────────────────────► │   Language Server    │
│   (VS Code)         │  textDocument/completion│   (smpe_ls)          │
│                     │  textDocument/hover     │                      │
│                     │  textDocument/diagnostic│                      │
└─────────────────────┘                         └──────────────────────┘
```

### Vorteile für Systemprogrammierer
- **Ein Server** — viele Editoren (VS Code, Neovim, Emacs, ...)
- Echtzeit-Validierung direkt im Editor
- Keine separaten Tools nötig — alles integriert

---

# Architektur-Überblick

```
┌──────────────────────────────────────────────────────┐
│                 VS Code Extension                    │
│                                                      │
│  ┌────────────────┐  LSP  ┌───────────────────────┐  │
│  │  Extension     │◄─────►│  smpe_ls              │  │
│  │  (TypeScript)  │       │  Language Server (Go) │  │
│  └───────┬────────┘       └───────────────────────┘  │
│          │                                           │
│  ┌───────▼────────────────────────────────────────┐  │
│  │           z/OSMF Integration                   │  │
│  │  CSI-Queries · Dataset Browse · HOLD Comments  │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘

      ┌─────────────┐         ┌─────────────┐
      │  smpe_lint  │         │  smpe_outl  │
      │ (CLI Linter)│         │ (CLI Outl.) │
      └─────────────┘         └─────────────┘
```

Alle drei Binaries sind in **Go** implementiert und in der Extension gebündelt.

---

# Bestandteile der Extension

| Komponente | Typ | Zweck |
|-----------|-----|-------|
| **smpe_ls** | LSP Server (Go) | Sprachunterstützung im Editor |
| **smpe_lint** | CLI Tool (Go) | Standalone-Linting, CI/CD |
| **smpe_outl** | CLI Tool (Go) | Dokument-Outline als Text/JSON |
| **IntelliSense** | LSP Feature | Code Completion, Hover, Snippets |
| **Diagnostics** | LSP Feature | Echtzeit-Validierung |
| **Formatierung** | LSP Feature | Auto-Format von MCS Statements |
| **z/OSMF Integration** | Extension | CSI-Queries direkt aus VS Code |
| **Commands** | Extension | Aufrufbare Aktionen (Command Palette) |
| **Outline View** | LSP Feature | Navigation im Dokument |
| **Output Channel** | Extension | Debug-Logging & Server-Ausgabe |

---

# smpe_ls — Der Language Server

## Was er tut

`smpe_ls` ist ein LSP-kompatibler Language Server für **SMP/E MCS-Dateien** (`.smpe`, `.mcs`, `.smp`).

Er läuft als Hintergrundprozess und kommuniziert mit VS Code über **JSON-RPC via stdio**.

## LSP Capabilities (1/2)

| Capability | Beschreibung |
|-----------|-------------|
| `textDocument/completion` | Code Completion & Snippets |
| `textDocument/hover` | Hover-Dokumentation |
| `textDocument/publishDiagnostics` | Echtzeit-Fehlerprüfung |
| `textDocument/formatting` | Dokument-Formatierung |
| `textDocument/foldingRange` | Code Folding |

---

# smpe_ls — LSP Capabilities (2/2)

| Capability | Beschreibung |
|-----------|-------------|
| `textDocument/documentSymbol` | Outline (Dokumentsymbole) |
| `workspace/symbol` | Workspace-weite Symbolsuche |
| `textDocument/definition` | Gehe zur Definition |
| `textDocument/references` | Alle Referenzen finden |
| `textDocument/signatureHelp` | Signatur-Hilfe in Operanden-Klammern |
| `textDocument/codeAction` | Quick Fixes (Glühbirne) |

## Settings in VS Code

| Setting | Standard | Beschreibung |
|---------|----------|-------------|
| `smpe.serverPath` | `""` | Eigener Server-Pfad (leer = gebündelt) |
| `smpe.dataPath` | `""` | Eigene smpe.json (leer = gebündelt) |
| `smpe.debug` | `false` | Debug-Logging aktivieren |

---

# smpe_ls — CLI & Log

## Kommandozeilenargumente

```bash
smpe_ls [Optionen]

  --debug      Debug-Logging aktivieren
  --data       Pfad zur smpe.json Datendatei
               (Standard: ~/.local/share/smpe_ls/smpe.json)
  --version    Versionsinformation anzeigen
```

## Log-Datei

Der Server schreibt bei jedem Start ein neues Log:

```
~/.local/share/smpe_ls/smpe_ls.log
```

Das Verzeichnis wird beim ersten Start automatisch angelegt.

---

# smpe_lint — Der SMP/E Linter

## Zweck

`smpe_lint` ist ein **standalone CLI-Tool** für die Validierung von SMP/E MCS-Dateien — unabhängig von VS Code. Ideal für **CI/CD Pipelines** und automatisierte Qualitätsprüfungen.

## Kommandozeilenargumente

```bash
smpe_lint [Optionen] <Dateimuster>

  --config <Pfad>       Konfigurationsdatei (.smpe_lint.yaml oder .json)
  --data <Pfad>         Pfad zur smpe.json Datendatei
  --disable <Code>      Bestimmte Diagnose deaktivieren (mehrfach nutzbar)
  --init <format>       Beispiel-Konfiguration erstellen (yaml oder json)
  --json                Ausgabe im JSON-Format
  --warnings-as-errors  Warnungen als Fehler behandeln (Exit-Code 1)
  --version, -v         Version anzeigen
```

> 🎬 **Live Demo**

---

# smpe_lint — Diagnostic Codes (Syntax)

## Syntax-Prüfungen

| Code | Beschreibung |
|------|-------------|
| `unknown_statement` | Unbekanntes MCS Statement |
| `invalid_language_id` | Ungültiger 3-stelliger Language-Identifier |
| `unbalanced_parentheses` | Unausgeglichene Klammern |
| `missing_terminator` | Fehlender Statement-Terminator (`.`) |
| `missing_parameter` | Fehlender erforderlicher Parameter |
| `content_beyond_column_72` | Inhalt jenseits Spalte 72 |
| `missing_inline_data` | Fehlende Inline-Daten (z.B. nach `++JCLIN`) |

## Konfiguration

Jeden Code einzeln abschalten:

```json
"smpe.diagnostics.unknownStatement": false
```

---

# smpe_lint — Diagnostic Codes (Operanden)

## Operanden-Prüfungen

| Code | Beschreibung |
|------|-------------|
| `unknown_operand` | Unbekannter Operand |
| `duplicate_operand` | Doppelter Operand |
| `missing_required_operand` | Fehlender Pflicht-Operand |
| `mutually_exclusive` | Sich ausschließende Operanden |
| `dependency_violation` | Operand-Abhängigkeit verletzt |
| `unknown_sub_operand` | Unbekannter Sub-Operand |
| `sub_operand_validation` | Sub-Operand Validierungsfehler |

## Konfiguration

Jeden Code einzeln abschalten:

```json
"smpe.diagnostics.unknownOperand": false
```

---

# smpe_lint — Konfigurationsdatei

Eine `.smpe_lint.yaml` erlaubt projektspezifische Einstellungen:

```bash
# Beispiel-Konfiguration erzeugen
smpe_lint --init yaml
```

```yaml
# .smpe_lint.yaml
disable:
  - unknown_operand
warnings_as_errors: false
```

## CI/CD Beispiele

```bash
# Einfaches Linting
smpe_lint *.smpe

# JSON-Ausgabe für Reporting
smpe_lint --json --warnings-as-errors packages/**/*.smpe

# Mit Konfigurationsdatei
smpe_lint --config .smpe_lint.yaml *.smpe
```

> 🎬 **Live Demo:** `smpe_lint --init yaml` und Ausgabe zeigen

---

# smpe_outl — Der SMP/E Outliner

## Zweck

`smpe_outl` gibt die **Dokument-Struktur** (Symbole) von SMP/E MCS-Dateien aus — als lesbaren Text oder maschinenlesbares JSON. Nützlich für Skripting, Reporting und Integration in eigene Werkzeuge.

## Kommandozeilenargumente

```bash
smpe_outl [Optionen] <Dateimuster>

  --json          Ausgabe im JSON-Format
  --ranges        Range und selectionRange in JSON aufnehmen
  --meta          hasInlineData-Feld pro Statement aufnehmen
  --data <Pfad>   Eigene smpe.json Datendatei
  --version, -v   Version anzeigen
```

## Setting in VS Code

| Setting | Standard | Beschreibung |
|---------|----------|-------------|
| `smpe.outlPath` | `""` | Eigener smpe_outl-Pfad (leer = gebündelt) |

---

# smpe_outl — Ausgabe-Beispiele

## Textausgabe

```
PTF UA12345  [line 1]
  VER 7.3    [line 5]
  MOD MYPROG [line 9]
APAR OA12345 [line 20]
  VER 7.3    [line 24]
```

## JSON-Ausgabe (mit `--json --ranges`)

```json
[
  {
    "name": "++PTF(UA12345)",
    "kind": 5,
    "range": { "start": {"line": 0, "character": 0},
               "end":   {"line": 18, "character": 1} },
    "hasInlineData": false
  }
]
```

> 🎬 **Live Demo**

---

# IntelliSense — Code Completion

## Was ist IntelliSense?

IntelliSense ist VS Codes Oberbegriff für kontextsensitive **Autovervollständigung**, **Hover-Dokumentation** und **Signatur-Hilfe**.

## Code Completion in smpe_ls

### 1. Statement-Completion
Beim Eintippen von `++` werden alle gültigen MCS Statements vorgeschlagen.

### 2. Operanden-Completion
Nach einem Statement-Namen werden alle gültigen Operanden vorgeschlagen — kontextabhängig, basierend auf `smpe.json`.

### 3. Boilerplate-Snippets (`…`)
Jedes Statement bietet ein Template mit Tab-Stops für alle Pflicht-Operanden. REWORK wird mit dem aktuellen Datum vorbefüllt.

> 🎬 **Live Demo:** Completion und Snippet zeigen

---

# IntelliSense — Hover & Datenbasis

## Hover-Dokumentation

Wenn der Cursor über einem **Statement** oder **Operanden** steht:

- Name und Kurzbeschreibung des Statements
- Erlaubte Operanden
- Beschreibung des Operanden unter dem Cursor
- Pflicht/Optional-Status

## smpe.json als zentrale Datenbasis

Alle Statements, Operanden, Abhängigkeiten und Dokumentationstexte kommen aus einer einzigen Datei:

```
smpe.json  ──►  smpe_ls  ──►  Completion
                         ──►  Hover
                         ──►  Diagnostics
                         ──►  Snippets
```

> 🎬 **Live Demo:** Hover über `++HOLD`, `++PTF`, Operanden zeigen

---

# Diagnostics — Echtzeit-Validierung

## Was wird geprüft?

Der Language Server validiert MCS-Dateien **beim Tippen** und markiert Probleme direkt im Editor.

## Beispiele

```smpe
++PTF (UA12345)  FMID(CTSA400)
  VER(7.3) PRODXT('IBM').
```
→ Fehler: `unknown_operand` (PRODXT)

```smpe
++HOLD (UA12345) SYSTEM
  FMID(CTSA400)
  REASON(DELETE
  COMMENT(Text).
```
→ Fehler: `unbalanced_parentheses`

> 🎬 **Live Demo:** Fehler eintippen und Diagnose zeigen

---

# Code Actions (Quick Fixes)

## 💡 Automatische Korrekturen direkt im Editor

Klick auf die Glühbirne (oder `Cmd+.` / `Ctrl+.`) öffnet das Quick-Fix-Menü.

## Verfügbare Fixes

| Fix | Wirkung |
|-----|---------|
| **Statement-Terminator einfügen** | Setzt den fehlenden `.` ans Ende des Statements |
| **Operand einfügen** / **Alle einfügen** | Fügt Gerüste wie `SOURCEID()` für fehlende Pflicht-Operanden ein |
| **REWORK auf aktuelles Datum setzen** | Füllt leeres `REWORK()` mit julianischem Datum (`jjjjddd`) |
| **REWORK aktualisieren** *(neu in v1.3.7)* | Frischt ein bereits gefülltes, veraltetes `REWORK()` auf — per Cursor, ganz ohne Diagnostic |

> Reine Texteinfügungen — keine Konfiguration nötig, automatisch aktiv.

---

# Formatierung

## Auto-Format von MCS Statements

### Vor der Formatierung
```smpe
++PTF(UA12345) FMID(CTSA400) VER(7.3) REWORK(2026100) SUP(UA99999,UA88888).
```

### Nach der Formatierung
```smpe
++PTF(UA12345)
   FMID(CTSA400)
   VER(7.3)
   REWORK(2026100)
   SUP(UA99999,
       UA88888).
```

## Settings

| Setting | Standard | Beschreibung |
|---------|----------|-------------|
| `smpe.formatting.enabled` | `true` | Formatierung aktivieren |
| `smpe.formatting.indentContinuation` | `3` | Einrückung Fortsetzungszeilen |
| `smpe.formatting.oneOperandPerLine` | `true` | Jeder Operand in eigener Zeile |
| `smpe.formatting.wrapListsAfterN` | `2` | Listen umbrechen nach N Einträgen |
| `smpe.formatting.formatOnSave` | `false` | Beim Speichern formatieren |

> 🎬 **Live Demo**

---

# z/OSMF Integration

## Direkt aus VS Code auf das Mainframe zugreifen

Die Extension verbindet sich über die **z/OSMF REST API** mit dem z/OS System und ermöglicht CSI-Abfragen ohne ISPF.

## Konfigurationsdatei `.smpe-zosmf.yaml`

```yaml
host: zos.example.com
port: 443
user: SYSADM
basePath: /zosmf
defaultCsi: SMPE.GLOBAL.CSI
csi:
  - SMPE.GLOBAL.CSI
  - SMPE.TARGET.CSI
```

Anlegen per Command:
```
Cmd+Shift+P  →  SMP/E: Create z/OSMF Config
```

---

# z/OSMF — Settings & Spracherkennung

## Settings

| Setting | Standard | Beschreibung |
|---------|----------|-------------|
| `smpe.zosmf.queryTimeoutSeconds` | `300` | Timeout für CSI-Abfragen (30–600s) |
| `smpe.zosDatasetsLlq` | `["MCS"]` | LLQ für automatische Spracherkennung |

## Automatische Spracherkennung für z/OS Datasets

Datasets mit konfigurierbaren **Last Level Qualifiers** werden automatisch als SMP/E erkannt:

```json
"smpe.zosDatasetsLlq": ["MCS", "SMPE", "USERMOD"]
```

## Column Rulers

```json
"smpe.editor.showColumnRulers": true
```

Zeigt Linien bei Spalte 72 und 80 — die klassischen Mainframe-Kartengrenzen.

---

# z/OSMF Commands

## Verfügbare Commands (`Cmd+Shift+P`)

| Command | Beschreibung |
|---------|-------------|
| `SMP/E: Query SYSMOD via z/OSMF` | CSI-Abfrage für einen SYSMOD |
| `SMP/E: Query DDDEF via z/OSMF` | CSI-Abfrage für einen DDDEF |
| `SMP/E: List Zones via z/OSMF` | Alle Zonen auflisten |
| `SMP/E: Free Form CSI Query` | Freie CSI-Abfrage (alle Entry Types) |
| `SMP/E: Create z/OSMF Config` | Konfigurationsdatei anlegen |
| `SMP/E: Clear Stored Password` | Gespeichertes Passwort löschen |
| `SMP/E: Check Missing Input Members` | Fehlende Input-Member prüfen |
| `SMP/E: Manage Filter History` | Gespeicherte Filter verwalten |
| `SMP/E: Toggle Auto-Detect Language Mode` | Automatische Spracherkennung ein/aus |

## CodeLens

Klickbare Links direkt über SYSMOD- und DDDEF-Namen im Editor:

```smpe
++PTF(UA12345)       ← [Query SYSMOD] [Query DDDEF]
```

---

# Free Form CSI Query

## Leistungsfähige Abfragen direkt im Editor

- **Entry Type** per Picklist — alle CSI-Entry-Types inkl. HFS, Data-Elemente und `ELEMENT`-Pseudo-Entry *(neu in v1.3.5)*
- Sprachvarianten (z.B. `HFSESP`) erhalten automatisch die Subentries ihres Basistyps
- **Subentries** per Picker auswählen
- **Filter** als CSI-Filterausdruck (z.B. `ENAME='UA12345'`)
- **Mehrere Zonen** gleichzeitig abfragen
- Export als **JSON** oder **CSV**

## Filter History

- Filter-Ausdrücke werden **automatisch gespeichert** (max. 20, keine Duplikate)
- **▼**-Button neben dem Filter-Feld öffnet Dropdown mit gespeicherten Einträgen
- `SMP/E: Manage Filter History` → Edit / Delete / Clear All
- Änderungen wirken **sofort** im offenen Panel

## Saved Queries *(neu in v1.3.0)*

- Vollständige CSI-Abfragen (Zone, Entry Type, Subentries, Filter) **speichern und wiederverwenden**
- Queries werden in `.smpe-saved-queries.yaml` im Workspace-Root gespeichert
- Aufklappbarer **Saved Queries**-Bereich unterhalb des Eingabeformulars
- Klick auf Eintrag lädt die Abfrage zurück ins Formular

> 🎬 **Live Demo:** Free Form Query, Filter History, Saved Queries

---

# HOLD Comments Viewer

## PTF HOLD-Texte direkt im Editor

In SYSMOD-Abfragen (GLOBAL Zone) mit HOLDDATA:

1. HOLD-Button in der Ergebniszeile klicken
2. PTF-Member wird aus SMPPTS via z/OSMF geladen
3. Vollständiger `++HOLD`-Block wird im Seitenpanel angezeigt
4. Bei bereits akzeptierten PTFs: klare Fehlermeldung

> 🎬 **Live Demo:** HOLDDATA und HOLD Comments Viewer

---

# Check Missing Input Members

## Zweck

Prüft ob alle in MCS-Dateien referenzierten Input-Member-Dateien im Workspace vorhanden sind — bevor SMP/E den Fehler meldet.

## Aufruf

```
Cmd+Shift+P  →  SMP/E: Check Missing Input Members
```

Oder per Rechtsklick auf eine `.smpe`-Datei im Explorer.

## Ergebnis

Eine filterbare und sortierbare Tabelle mit Status **Found** / **Missing** für jeden referenzierten Member.

Die Spalte **Source** zeigt, wie der Member ermittelt wurde: `convention` (Elementname + Endung unterhalb der Suchordner) oder `placeholder` (eine `{{ pfad }}`-Zeile, relativ zum Repository-Root).

## Settings

| Setting | Standard | Beschreibung |
|---------|----------|-------------|
| `smpe.checkMissingInputMembers.searchFolders` | `["customization"]` | Suchordner. `"*"` = gesamter Workspace |

> 🎬 **Live Demo**

---

# Outline View

## Was ist die Outline View?

Die **Outline View** in VS Code zeigt die Struktur des aktuell geöffneten Dokuments als navigierbaren Baum.

```
OUTLINE
├── ++PTF(UA12345)
│   ├── ++VER(7.3)
│   └── ++MOD(MYPROG)
├── ++APAR(OA99999)
│   └── ++VER(7.3)
└── ++HOLD(UA12345)
```

## Wie smpe_ls die Outline verwendet

Der Language Server implementiert `textDocument/documentSymbol` — jedes MCS Statement wird als Symbol gemeldet, inkl. hierarchischer Struktur.

## Workspace Symbol Search

`Cmd+T` → Suche nach SYSMOD-Namen über **alle** `.smpe`-Dateien im Workspace.

> 🎬 **Live Demo**

---

# Output Channel

## Was ist der Output Channel?

Der **Output Channel** ist ein dediziertes Log-Panel in VS Code:

```
View → Output → "SMP/E Language Server"
```

## Was wird dort ausgegeben?

- Server-Start und -Stop Meldungen
- LSP-Verbindungsstatus
- z/OSMF Anfragen und Antwort-Codes
- Fehler bei der Konfiguration
- Debug-Ausgaben (wenn `smpe.debug: true`)

## Debug-Logging aktivieren

```json
"smpe.debug": true
```

Schreibt zusätzlich in: `~/.local/share/smpe_ls/smpe_ls.log`

> 🎬 **Live Demo:** Output Channel und Log-Datei zeigen

---

# Installation

## 1. VSIX-Datei herunterladen

Von der GitHub Release-Seite die passende Plattform-Version laden:

| Plattform | Datei |
|-----------|-------|
| macOS Apple Silicon | `...-darwin-arm64-1.3.8.vsix` |
| macOS Intel | `...-darwin-x64-1.3.8.vsix` |
| Windows x64 | `...-win32-x64-1.3.8.vsix` |
| Linux x64 | `...-linux-x64-1.3.8.vsix` |

## 2. In VS Code installieren

```
Cmd+Shift+P  →  Extensions: Install from VSIX...
```

## 3. z/OSMF konfigurieren (optional)

```
Cmd+Shift+P  →  SMP/E: Create z/OSMF Config
```

> 🎬 **Live Demo:** Installation und erste Schritte

---

# Settings-Referenz — Allgemein

| Setting | Standard | Beschreibung |
|---------|----------|-------------|
| `smpe.serverPath` | `""` | Eigener smpe_ls Pfad |
| `smpe.dataPath` | `""` | Eigene smpe.json |
| `smpe.outlPath` | `""` | Eigener smpe_outl Pfad |
| `smpe.debug` | `false` | Debug-Logging |
| `smpe.editor.showColumnRulers` | `true` | Spaltenlinien bei 72 und 80 |
| `smpe.zosDatasetsLlq` | `["MCS"]` | LLQ für Spracherkennung |
| `smpe.zosmf.queryTimeoutSeconds` | `300` | Query-Timeout (30–600s) |
| `smpe.checkMissingInputMembers.searchFolders` | `["customization"]` | Suchordner |
| `smpe.signatureHelp.enabled` | `true` | Signatur-Hilfe in Operanden-Klammern |
| `smpe.editor.autoDetectLanguage` | `true` | Automatische Spracherkennung (deaktivieren für manuelle Sprachauswahl) |

---

# Settings-Referenz — Formatierung

| Setting | Standard | Beschreibung |
|---------|----------|-------------|
| `smpe.formatting.enabled` | `true` | Formatierung aktiv |
| `smpe.formatting.indentContinuation` | `3` | Einrückung Folgezeilen |
| `smpe.formatting.oneOperandPerLine` | `true` | Ein Operand pro Zeile |
| `smpe.formatting.wrapListsAfterN` | `2` | Listen umbrechen |
| `smpe.formatting.formatOnSave` | `false` | Format beim Speichern |
| `smpe.formatting.moveLeadingComments` | `false` | Kommentare in Statement verschieben |

---

# Settings-Referenz — Diagnostics

| Setting | Standard |
|---------|----------|
| `smpe.diagnostics.unknownStatement` | `true` |
| `smpe.diagnostics.invalidLanguageId` | `true` |
| `smpe.diagnostics.unbalancedParentheses` | `true` |
| `smpe.diagnostics.missingTerminator` | `true` |
| `smpe.diagnostics.missingParameter` | `true` |
| `smpe.diagnostics.unknownOperand` | `true` |
| `smpe.diagnostics.duplicateOperand` | `true` |
| `smpe.diagnostics.missingRequiredOperand` | `true` |
| `smpe.diagnostics.mutuallyExclusive` | `true` |
| `smpe.diagnostics.dependencyViolation` | `true` |
| `smpe.diagnostics.contentBeyondColumn72` | `true` |
| `smpe.diagnostics.missingInlineData` | `true` |
| `smpe.diagnostics.commentInColumn1` | `true` |

---

<!-- _class: title -->

# Zusammenfassung

## smpe_ls bringt moderne IDE-Unterstützung für z/OS SMP/E MCS Statements

- **Echtzeit-Validierung** — Fehler sofort sehen, nicht erst beim SMP/E APPLY
- **IntelliSense** — Kein Nachschlagen im SMP/E Reference Manual
- **z/OSMF Integration** — CSI-Abfragen direkt aus VS Code
- **CLI Tools** — smpe_lint für CI/CD, smpe_outl für Scripting
- **Plattformunabhängig** — Windows, macOS, Linux
- **Open Source** — AGPL-3.0

---

<!-- _class: title -->

# Fragen?

**GitHub:** https://github.com/cybersorcerer/smpe_ls

**Ronny Funk — Senior Mainframe Architect SVA**
