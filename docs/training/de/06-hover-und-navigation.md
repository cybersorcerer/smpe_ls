# Modul 06 — Hover und Navigation

## Überblick

Der Language Server zeigt beim Hovern über Statements und Operanden Dokumentation
direkt aus der IBM-Referenz. Zusätzlich ermöglicht er die Navigation zu Definitionen
und References sowie die Übersicht aller Symbole im Dokument und Workspace.

## Voraussetzungen

Extension installiert — siehe [Modul 01](01-installation.md).

## Hover-Dokumentation

Bewege den Mauszeiger über ein Statement oder einen Operanden — ein Tooltip zeigt
die IBM-Dokumentation:

- **Statement-Hover** (`++USERMOD`): Beschreibung des Statements, Typ, Parameter
- **Operanden-Hover** (`DESC`): Beschreibung, gültige Werte, Pflichtfeld oder optional

Beispiel: Hover über `REWORK` zeigt:
> *REWORK — Specifies the date the SYSMOD was last reworked. Format: YYYYMMDD.*

## Go to Definition

`F12` oder `Cmd+Click` (macOS) / `Ctrl+Click` (Windows/Linux) auf einen
SYSMOD-Namen oder eine FMID springt zur Definition innerhalb der gleichen Datei:

```smpe
++VER(Z038)
    FMID(HBB7790)   ← Cmd+Click springt zu ++FUNCTION(HBB7790)
    PRE(UJ12345).   ← Cmd+Click springt zu ++USERMOD(UJ12345) oder ++PTF(UJ12345)
```

## Find All References

`Shift+F12` auf einem SYSMOD-Namen zeigt alle Stellen in der Datei wo dieser
SYSMOD referenziert wird (z.B. in PRE, REQ, SUP, FMID).

## Document Symbols — Outline

`Cmd+Shift+O` (macOS) / `Ctrl+Shift+O` (Windows/Linux) öffnet die Outline-Ansicht
mit allen Statements des Dokuments in einer Hierarchie:

```
++USERMOD(LJS2024)
  ├── DESC(Konfigurationsänderung)
  ├── REWORK(20240315)
  └── PRE(UJ12345)
++VER(Z038)
  ├── FMID(HBB7790)
  └── PRE(UJ12345)
```

Ein Klick auf einen Eintrag springt direkt zur entsprechenden Stelle im Dokument.

## Workspace Symbols

`Cmd+T` (macOS) / `Ctrl+T` (Windows/Linux) durchsucht alle `.smpe`-Dateien im
Workspace nach SYSMOD-Definitionen:

- Eingabe eines SYSMOD-Namens (z.B. `LJS2024`) zeigt alle Treffer über alle Dateien
- Klick auf einen Treffer öffnet die Datei und springt zur Definition

## Folding

MCS-Statements und mehrzeilige Kommentare können ein- und ausgeklappt werden:

- Klick auf den Pfeil links neben der Zeilennummer
- Oder: `Cmd+K Cmd+[` / `Ctrl+K Ctrl+[` für alle zuklappen

## Zusammenfassung

- Hover zeigt IBM-Dokumentation für Statements und Operanden
- `F12` / `Cmd+Click` springt zu Definitionen
- `Shift+F12` zeigt alle Referenzen
- `Cmd+Shift+O` öffnet die Outline des aktuellen Dokuments
- `Cmd+T` durchsucht alle `.smpe`-Dateien im Workspace
- Statements und Kommentare können ein-/ausgeklappt werden
