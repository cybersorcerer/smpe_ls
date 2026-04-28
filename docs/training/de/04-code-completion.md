# Modul 04 — Code Completion

## Überblick

Der Language Server bietet kontextsensitive Vervollständigung für Statements und
Operanden. Die Vorschläge sind immer auf das aktuell getippte Statement abgestimmt —
es werden nur syntaktisch gültige Operanden angeboten.

## Voraussetzungen

Extension installiert — siehe [Modul 01](01-installation.md).

## Statement-Vervollständigung

Tippe `++` in einer leeren Zeile und drücke `Ctrl+Space` (Windows/Linux) bzw.
`Cmd+Space` (macOS). Es erscheint eine Liste aller verfügbaren MCS-Statements:

```
++APAR
++ASSIGN
++DELETE
++FUNCTION
++HOLD
++IF
++JCLIN
++MAC
...
```

Jeder Eintrag zeigt eine kurze Beschreibung aus der IBM-Dokumentation.

## Operanden-Vervollständigung

Nach dem Öffnen eines Statements werden nur die für dieses Statement gültigen
Operanden angeboten. Beispiel: Nach `++VER(` bietet die Completion folgende
Operanden an:

```smpe
++VER(Z038)
    FMID(      ← Ctrl+Space hier zeigt: FMID, PRE, REQ, SUP, ...
```

Operanden die bereits verwendet wurden, werden nicht erneut vorgeschlagen
(außer bei List-Operanden wie `PRE`, `REQ`).

## Schritt für Schritt: ++VER Statement vervollständigen

So entsteht ein vollständiges `++VER`-Statement mit Completion:

1. Neue Zeile, tippe `++VER(` — Completion schlägt bekannte FMIDs vor
2. Wähle oder tippe die FMID, schließe mit `)` und drücke Enter
3. Tippe `FM` + `Ctrl+Space` → Vorschlag `FMID(`
4. Wähle `FMID(`, tippe den Wert, schließe mit `)`
5. Tippe `PR` + `Ctrl+Space` → Vorschlag `PRE(`
6. Abschließen mit `.` auf einer neuen Zeile

Ergebnis:

```smpe
++VER(Z038)
    FMID(HBB7790)
    PRE(UJ12345
        UJ67890).
```

## Boilerplate-Snippets

Für alle Control MCS Statements bietet die Completion zusätzlich ein Snippet-Item an.
Es ist am `…`-Suffix im Label und am Snippet-Icon erkennbar:

```text
++PTF
++PTF …        ← Snippet
++USERMOD
++USERMOD …    ← Snippet
```

Ein Snippet fügt ein vollständiges Boilerplate-Template ein, das alle required Operanden
enthält und Tab Stops für die Platzhalter setzt. Beispiel für `++PTF …`:

```smpe
++PTF(UAnnnnn)
  DESC(description)
  REWORK(2026118)
  .
```

Nach der Auswahl springt der Cursor automatisch zum ersten Platzhalter (`UAnnnnn`).
Mit `Tab` gelangt man zum nächsten Platzhalter, bis alle Werte ausgefüllt sind.

Der REWORK-Wert wird automatisch mit dem aktuellen Datum im Format `yyyyddd` vorbesetzt
(z.B. `2026118` für den 28. April 2026).

## Gegenseitig ausschließende Operanden

Manche Operanden schließen sich gegenseitig aus (z.B. `SYSLIB` und `TXLIB` bei
`++PARM`). Beide werden in der Completion angezeigt — die Validierung erfolgt
über Diagnostics (siehe Modul 05).

## Zusammenfassung

- `++` + `Ctrl+Space` zeigt alle verfügbaren Statements
- Operanden-Completion ist kontextsensitiv — nur gültige Operanden werden gezeigt
- Bereits verwendete Operanden werden nicht erneut vorgeschlagen
- Completion funktioniert auch nach Inline-Data wenn `++` getippt wird
