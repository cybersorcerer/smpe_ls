# Modul 05 — Diagnostics

## Überblick

Der Language Server prüft SMP/E-Syntax in Echtzeit und zeigt Fehler, Warnungen
und Hinweise direkt im Editor an. Alle Checks können einzeln über die Einstellungen
aktiviert oder deaktiviert werden.

## Voraussetzungen

Extension installiert — siehe [Modul 01](01-installation.md).

## Schweregrade

| Symbol | Schweregrad | Beschreibung |
|--------|-------------|--------------|
| 🔴 | Fehler | Syntax-Fehler, SMP/E würde die Datei ablehnen |
| ⚠️ | Warnung | Mögliche Probleme, die zur Laufzeit auftreten können |
| ℹ️ | Information | Hinweise zur besseren Lesbarkeit |
| 💡 | Hinweis | Optionale Verbesserungsvorschläge |

## Verfügbare Checks

### Unbekanntes Statement / Unbekannter Operand

```smpe
++USRMOD(LJS2024)    ← 🔴 Unbekanntes Statement "++USRMOD"
    DESCRIPTN(Text). ← 🔴 Unbekannter Operand "DESCRIPTN"
```

### Fehlender Pflicht-Operand

```smpe
++PTF(UJ12345)
    FMID(HBB7790).   ← ⚠️ Fehlender Pflicht-Operand: VER fehlt
```

### Fehlender Terminator

```smpe
++USERMOD(LJS2024)
    DESC(Meine Änderung)   ← ⚠️ Fehlender Terminator (.)
++VER(Z038)
```

### Unbalancierte Klammern

```smpe
++USERMOD(LJS2024
    DESC(Meine Änderung).  ← 🔴 Fehlende schließende Klammer
```

### Fehlende Inline-Data

```smpe
++MAC(MYMACRO)
    SYSLIB(S$SYSM1)
    DISTLIB(A$SYSM1).      ← ⚠️ Fehlende Inline-Data nach Statement
```

### Inhalt nach Spalte 72

Zeichen nach Spalte 72 werden von SMP/E ignoriert:

```smpe
++USERMOD(LJS2024)
    DESC(Dies ist eine sehr lange Beschreibung die über Spalte 72 hinausgeht!!). ← ⚠️
```

### Gegenseitig ausschließende Operanden

```smpe
++PARM(RSSPRMI)
    SYSLIB(S$IQPARM)
    TXLIB(PARM).     ← ⚠️ SYSLIB und TXLIB schließen sich gegenseitig aus
```

### Alleinstehender Kommentar zwischen Statements

```smpe
++USERMOD(LJS2024)
    DESC(Änderung).
/* Dieser Kommentar verursacht einen SMP/E Syntaxfehler */   ← 🔴
++VER(Z038)
```

## Diagnostics konfigurieren

Alle Checks können in den VSCode-Einstellungen individuell deaktiviert werden:

```json
{
  "smpe.diagnostics.missingInlineData": false,
  "smpe.diagnostics.contentBeyondColumn72": false
}
```

Vollständige Liste aller Einstellungen:

| Einstellung | Standard |
|-------------|----------|
| `smpe.diagnostics.unknownStatement` | `true` |
| `smpe.diagnostics.unknownOperand` | `true` |
| `smpe.diagnostics.missingRequiredOperand` | `true` |
| `smpe.diagnostics.missingParameter` | `true` |
| `smpe.diagnostics.missingTerminator` | `true` |
| `smpe.diagnostics.unbalancedParentheses` | `true` |
| `smpe.diagnostics.missingInlineData` | `true` |
| `smpe.diagnostics.contentBeyondColumn72` | `true` |
| `smpe.diagnostics.mutuallyExclusive` | `true` |
| `smpe.diagnostics.standaloneCommentBetweenMCS` | `true` |
| `smpe.diagnostics.commentInColumn1` | `true` |
| `smpe.diagnostics.duplicateOperand` | `true` |
| `smpe.diagnostics.emptyOperandParameter` | `true` |
| `smpe.diagnostics.dependencyViolation` | `true` |
| `smpe.diagnostics.requiredGroup` | `true` |
| `smpe.diagnostics.invalidLanguageId` | `true` |
| `smpe.diagnostics.unknownSubOperand` | `true` |
| `smpe.diagnostics.subOperandValidation` | `true` |

## Code Actions (Quick Fixes)

💡 Wenn eine Diagnostic angezeigt wird, erscheint im Editor eine Glühbirne. Klick darauf
(oder `Cmd+.` auf macOS / `Ctrl+.` auf Windows/Linux) öffnet das Quick-Fix-Menü.

### Verfügbare Quick Fixes

- **Statement-Terminator einfügen** — setzt den fehlenden Punkt `.` ans Ende des Statements
  (für die Diagnostic „Fehlender Terminator").
- **Operand einfügen** / **Alle fehlenden Operanden einfügen** — fügt ein Gerüst wie
  `SOURCEID()` für jeden fehlenden Pflicht-Operanden ein. Sind zwei oder mehr Operanden
  gleichzeitig offen, erscheint zusätzlich die Aktion „Alle einfügen", die alle auf einmal ergänzt.
- **REWORK auf aktuelles Datum setzen** — füllt ein leeres `REWORK()` mit dem aktuellen
  julianischen Datum im Format `jjjjddd`.

### REWORK aktualisieren *(neu in v1.3.7)*

Steht der Cursor in einem bereits gefüllten `REWORK(...)`, dessen Wert nicht dem heutigen
Datum entspricht, bietet die Glühbirne zusätzlich **„REWORK aktualisieren"** an — unabhängig
von einer Diagnostic. Anders als die drei Fixes oben braucht dieser keinen Fehler als
Auslöser: er reagiert allein auf die Cursor-Position und ersetzt den vorhandenen Wert durch
das aktuelle julianische Datum. Ist der Wert bereits aktuell oder leer, erscheint diese
Aktion nicht (leeres `REWORK()` bleibt Sache des Fixes oben).

### Hinweise

- Die Fixes sind reine Texteinfügungen im Editor — keine externe Konfiguration nötig.
- Sie sind ohne weitere Einstellungen automatisch aktiv.

## Hinweis: z/OS Datasets via Zowe Explorer

z/OS Dataset Members die über Zowe Explorer geöffnet werden, haben CRLF-Zeilenenden.
Der Language Server normalisiert diese automatisch — es entstehen keine falschen
Diagnostics für `COMMENT`-Inhalte in `++HOLD`-Statements.

## Zusammenfassung

- Diagnostics laufen in Echtzeit beim Tippen
- Vier Schweregrade: Fehler 🔴, Warnung ⚠️, Info ℹ️, Hinweis 💡
- Alle Checks sind einzeln konfigurierbar
- Einstellungsänderungen wirken sofort ohne Neustart
- CRLF-Zeilenenden (Zowe Explorer) werden automatisch normalisiert
- Quick Fixes (💡) beheben fehlenden Terminator, fehlende Pflicht-Operanden und leeres REWORK automatisch
- Ein cursor-basierter Fix aktualisiert ein bereits gefülltes, veraltetes REWORK — ohne Diagnostic nötig
