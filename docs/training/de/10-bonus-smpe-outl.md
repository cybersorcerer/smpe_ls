# Modul 10 — Bonus: smpe_outl

## Überblick

`smpe_outl` gibt die Dokumentstruktur (Outline) von SMP/E-Dateien aus — als Markdown
für Reports oder als JSON für Skripte und Pipelines. Mit dem `--meta`-Flag kann es
auch fehlende Input-Member in CI/CD-Pipelines erkennen.

## Voraussetzungen

`smpe_outl`-Binary für die eigene Plattform aus den
[GitHub Releases](https://github.com/cybersorcerer/smpe_ls/releases/latest) herunterladen
und in den PATH legen (z.B. `~/.local/bin/smpe_outl`).

## Grundlegende Verwendung

```bash
# Outline als Markdown ausgeben
smpe_outl meinmod.smpe

# Mehrere Dateien
smpe_outl *.smpe
```

**Markdown-Ausgabe:**

```markdown
# SMP/E Outline Report

## meinmod.smpe

### ++USERMOD(LJS2024)
- `DESC(Konfigurationsänderung)`
- `REWORK(20240315)`
- `PRE(UJ12345)`

### ++VER(Z038)
- `FMID(HBB7790)`
- `PRE(UJ12345)`
```

## JSON-Ausgabe

```bash
smpe_outl --json *.smpe
```

```json
[
  {
    "file": "meinmod.smpe",
    "symbols": [
      {
        "name": "++USERMOD(LJS2024)",
        "id": "LJS2024",
        "kind": 5,
        "children": [
          { "name": "DESC(Konfigurationsänderung)", "id": "Konfigurationsänderung", "kind": 14 },
          { "name": "PRE(UJ12345)", "id": "UJ12345", "kind": 14 }
        ]
      }
    ]
  }
]
```

Das JSON-Format ist kompatibel mit der LSP-`DocumentSymbol`-Struktur.

## --meta Flag: Inline-Data-Information

```bash
smpe_outl --json --meta *.smpe
```

Fügt jedem Statement das Feld `hasInlineData` hinzu:

```json
{
  "name": "++MAC(MYMACRO)",
  "id": "MYMACRO",
  "kind": 23,
  "hasInlineData": true
}
```

## --ranges Flag: LSP-Positionsinformationen

```bash
smpe_outl --json --ranges *.smpe
```

Fügt `range` und `selectionRange` (Zeile/Spalte) zu jedem Symbol hinzu — nützlich
für Tools die auf Positionen im Quelltext zugreifen müssen.

## Pipeline Use Case: Fehlende Input-Member erkennen

Mit `--json --meta` können fehlende Input-Member in CI/CD-Pipelines erkannt werden.
Das folgende Shell-Skript zeigt das Prinzip:

```bash
#!/bin/bash
# Alle PARM-Statements ohne Inline-Data finden und prüfen ob die .parm-Datei existiert

smpe_outl --json --meta *.smpe | \
  jq -r '.[] | .file as $f | .symbols[] |
    select(.name | startswith("++PARM")) |
    select(.hasInlineData == false) |
    select((.children // []) | map(.name) | any(startswith("TXLIB(")) | not) |
    [$f, .id] | @tsv' | \
  while IFS=$'\t' read -r file member; do
    expected="customization/${member}.parm"
    if [ ! -f "$expected" ]; then
      echo "MISSING: $expected (referenced in $file)"
      exit_code=1
    fi
  done

exit ${exit_code:-0}
```

**Zuordnung Statement → Dateiendung:**

| Statement | Endung |
|-----------|--------|
| `++PARM` | `.parm` |
| `++SRC` | `.hlasm` |
| `++MAC` | `.hlasm` |
| `++EXEC` | `.rexx` |
| `++MOD` | `.mod` |
| `++ZAP` | `.zap` |
| `++PROC` | `.jcl` |
| `++CLIST` | `.clist` |
| `++MSG` | `.msg` |
| `++HELP` | `.help` |

## Zusammenfassung

- `smpe_outl` gibt die Dokumentstruktur als Markdown oder JSON aus
- `--json` erzeugt LSP-DocumentSymbol-kompatibles JSON
- `--meta` fügt `hasInlineData` pro Statement hinzu
- `--ranges` fügt LSP-Positionsinformationen hinzu
- Ideal für Inventarisierung, Reports und CI/CD-Pipelines
- Pipeline-Skripte können mit `--json --meta` fehlende Input-Member erkennen
