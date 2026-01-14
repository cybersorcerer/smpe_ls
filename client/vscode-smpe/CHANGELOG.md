# Changelog

Alle wesentlichen Änderungen an diesem Projekt werden in dieser Datei dokumentiert.

## [0.7.0] - 2025

### Neu
- HFS MCS Statements hinzugefügt
- `++SHELLSCR` Statement Support
- Verbessertes Inline Data Parsing und Diagnostics
- Diverse Korrekturen in smpe.json

## [0.6.0] - 2025

### Geändert
- Completion und Diagnostics vollständig AST-basiert

## [0.5.1] - 2025

### Neu
- `++MOVE` Statement Support
- `++MOD` Statement Support

## [0.4.0] - 2025

### Neu
- `++MAC`, `++MACUPD`, `++SRC`, `++SRCUPD` Statement Support
- Inline Data Architektur mit dynamischer Behandlung via smpe.json
- Verbessertes Syntax Highlighting für Inline Data
- Visuelle Diagnostic Severity mit Unicode Symbolen

### Behoben
- Dataset-Namen mit Punkten werden korrekt behandelt
- Boolean Operand Parsing korrigiert
- Completion und Hover zeigen korrekte Statement-Namen

## [0.3.0] - 2025

### Neu
- Multiline Parameter Support
- Erkennung von fehlenden schließenden Klammern
- Flexible Whitespace Behandlung
- `++JCLIN`, `++JAR`, `++JARUPD`, `++VER`, `++ZAP` Statements
- `++JCLIN` Inline JCL Support

### Behoben
- Multiline Parameter werden korrekt erkannt
- Unbalancierte Klammern werden diagnostiziert

## [0.2.0] - 2025

### Neu
- Grundlegende Diagnostics
- Hover Information aus JSON-Datei
- Kontextsensitive Code Completion

## [0.1.0] - 2025

### Neu
- Initiales Release
- Syntax Highlighting für SMP/E MCS Statements
- VS Code Extension Grundgerüst
