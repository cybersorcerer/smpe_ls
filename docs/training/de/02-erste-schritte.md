# Modul 02 — Erste Schritte

## Überblick

Dieses Modul zeigt wie eine SMP/E-Datei geöffnet oder erstellt wird, wie die
automatische Spracherkennung funktioniert und welche visuellen Hilfen der Editor
bereitstellt.

## Voraussetzungen

Extension installiert und VSCode neu geladen — siehe [Modul 01](01-installation.md).

## Eine .smpe-Datei öffnen oder erstellen

Öffne eine bestehende Datei mit der Endung `.smpe`, `.mcs` oder `.smp`, oder
erstelle eine neue Datei und speichere sie mit einer dieser Endungen.

Die Extension aktiviert sich automatisch — in der unteren Statusleiste erscheint `SMP/E`.

## Automatische Spracherkennung

Die Extension erkennt SMP/E-Dateien auf drei Wegen:

**1. Dateiendung** — `.smpe`, `.mcs`, `.smp` werden immer als SMP/E behandelt.

**2. Erste Zeile** — Dateien ohne Endung werden erkannt wenn die erste nicht-leere
Zeile mit `++` beginnt:

```smpe
++USERMOD(MYMOD001)
    DESC(Meine erste Änderung).
```

**3. z/OS Datasets** — Datasets mit konfigurierbaren Last Level Qualifiers (LLQ)
werden automatisch aktiviert. Standard-LLQ ist `MCS`:

```json
"smpe.zosDatasetsLlq": ["MCS", "SMPE"]
```

Ein Dataset wie `SYS1.MYPKG.MCS` wird damit automatisch als SMP/E erkannt.

## Spaltenlineale

Der Editor zeigt automatisch visuelle Lineale bei Spalte 72 und 80 — die
IBM-Kartengrenzen für SMP/E-Syntax:

- **Spalte 72** — Inhalt dahinter wird von SMP/E ignoriert (grau markiert)
- **Spalte 80** — Ende der Kartenzeile (dunkelgrau markiert)

Die Lineale können über die Einstellung `smpe.editor.showColumnRulers` deaktiviert
werden.

## Wortumbruch deaktivieren

Für SMP/E-Dateien ist der Wortumbruch standardmäßig deaktiviert (`editor.wordWrap: off`),
da die Spaltenposition syntaktisch relevant ist.

## Zusammenfassung

- Dateiendungen `.smpe`, `.mcs`, `.smp` aktivieren die Extension automatisch
- Dateien ohne Endung werden anhand der ersten Zeile (`++`) erkannt
- z/OS Datasets werden über konfigurierbare LLQs erkannt
- Spaltenlineale bei 72 und 80 sind standardmäßig aktiv
