# Modul 03 — Syntax Highlighting

## Überblick

Der Editor hebt alle Bestandteile einer SMP/E-Datei farblich hervor: Statements,
Operanden, Parameter, Kommentare und Inline-Data. Dies macht es leichter die
Struktur auf einen Blick zu erfassen.

## Voraussetzungen

Extension installiert — siehe [Modul 01](01-installation.md).

## Statements

MCS-Statements wie `++USERMOD`, `++PTF`, `++FUNCTION` oder `++VER` werden in einer
eigenen Farbe hervorgehoben:

```smpe
++USERMOD(LJS2024)
    DESC(Anpassung der Konfigurationsparameter)
    REWORK(20240315)
    PRE(UJ12345).
```

## Operanden und Parameter

Operanden wie `DESC`, `REWORK`, `PRE` sowie ihre Parameter in Klammern werden
jeweils in einer eigenen Farbe dargestellt.

## Kommentare

Inline-Kommentare im Format `/* ... */` werden ausgegraut:

```smpe
++USERMOD(LJS2024)
    DESC(Konfigurationsanpassung) /* Änderung für Kunde XYZ */
    PRE(UJ12345).
```

## Statement-Terminator

Der Terminator `.` am Ende jedes Statements wird gesondert hervorgehoben.

## Inline-Data

Statements die Inline-Data erwarten (z.B. `++JCLIN`, `++MAC`, `++SRC`) zeigen den
nachfolgenden Datenblock in einer eigenen Farbe:

```smpe
++JCLIN.
//SMPMCS   JOB (ACCT),'INSTALL',CLASS=A
//STEP1    EXEC PGM=IEFBR14
++JCLIN.
```

Der JCL-Block zwischen den beiden `++JCLIN.`-Statements wird als Inline-Data
erkannt und anders eingefärbt als reguläre MCS-Syntax.

## Zusammenfassung

- Statements, Operanden, Parameter, Kommentare und Terminatoren werden farblich unterschieden
- Inline-Data (z.B. JCL nach `++JCLIN`) wird separat dargestellt
- Das Highlighting ist sprachspezifisch und unabhängig vom Farbschema
