# Modul 07 — Formatierung

## Überblick

Der Language Server kann SMP/E-Dokumente automatisch formatieren: Operanden werden
auf einzelne Zeilen verteilt, Einrückungen vereinheitlicht und der Terminator auf
eine eigene Zeile gesetzt.

## Voraussetzungen

Extension installiert — siehe [Modul 01](01-installation.md).

## Dokument formatieren

Tastenkombination:
- macOS: `Shift+Option+F`
- Windows/Linux: `Shift+Alt+F`

Oder: `Cmd+Shift+P` → `Format Document`

## Was die Formatierung macht

**Vorher:**

```smpe
++USERMOD(LJS2024) DESC(Konfigurationsänderung) REWORK(20240315) PRE(UJ12345 UJ67890).
```

**Nachher:**

```smpe
++USERMOD(LJS2024)
    DESC(Konfigurationsänderung)
    REWORK(20240315)
    PRE(UJ12345
        UJ67890).
```

- Jeder Operand auf einer eigenen Zeile
- Continuation-Zeilen mit konfigurierter Einrückung (Standard: 4 Leerzeichen)
- Terminator `.` auf eigener Zeile am Zeilenanfang
- Listen (PRE, REQ, SUP) werden nach N Einträgen umgebrochen

## Format on Save

Automatisches Formatieren beim Speichern aktivieren:

```json
{
  "smpe.formatting.formatOnSave": true
}
```

## Einstellungen

| Einstellung | Standard | Beschreibung |
|-------------|----------|--------------|
| `smpe.formatting.enabled` | `true` | Formatierung aktivieren |
| `smpe.formatting.indentContinuation` | `4` | Einrückung für Continuation-Zeilen |
| `smpe.formatting.oneOperandPerLine` | `true` | Jeden Operanden auf eigene Zeile |
| `smpe.formatting.wrapListsAfterN` | `2` | Listen nach N Einträgen umbrechen (0 = nie) |
| `smpe.formatting.formatOnSave` | `false` | Beim Speichern automatisch formatieren |
| `smpe.formatting.moveLeadingComments` | `false` | Kommentare vor dem ersten Statement ins Statement verschieben |

## Hinweis zu Inline-Data

Statements mit Inline-Data (z.B. `++JCLIN`, `++MAC`) werden von der Formatierung
nicht verändert — der Inline-Data-Block bleibt unberührt.

## Zusammenfassung

- `Shift+Option+F` (macOS) / `Shift+Alt+F` (Windows/Linux) formatiert das Dokument
- Operanden werden auf einzelne Zeilen verteilt, Terminator auf eigene Zeile
- Formatierung ist über mehrere Einstellungen konfigurierbar
- `formatOnSave` aktiviert automatisches Formatieren beim Speichern
- Inline-Data-Blöcke werden nicht verändert
