# Modul 09 — Bonus: smpe_lint

## Überblick

`smpe_lint` ist ein Kommandozeilen-Tool das die gleichen Diagnostics wie der
Language Server anwendet — ideal für CI/CD-Pipelines und automatisierte Qualitätsprüfungen.

## Voraussetzungen

`smpe_lint`-Binary für die eigene Plattform aus den
[GitHub Releases](https://github.com/cybersorcerer/smpe_ls/releases/latest) herunterladen
und in den PATH legen (z.B. `~/.local/bin/smpe_lint`).

## Grundlegende Verwendung

```bash
# Eine oder mehrere Dateien prüfen
smpe_lint meinmod.smpe

# Mit Wildcard
smpe_lint *.smpe

# Version anzeigen
smpe_lint --version
```

## Ausgabeformate

**Markdown (Standard):**

```
# SMP/E Lint Report

## meinmod.smpe

🔴 **ERROR** Line 3: Unknown statement '++USRMOD'
⚠️ **WARNING** Line 7: Missing required operand: FMID

---
2 issues found (1 error, 1 warning)
```

**JSON (`--json`):**

```bash
smpe_lint --json *.smpe
```

```json
[
  {
    "file": "meinmod.smpe",
    "issues": [
      {
        "line": 3,
        "severity": "error",
        "code": "unknown_statement",
        "message": "Unknown statement '++USRMOD'"
      }
    ]
  }
]
```

## Strikter Modus

```bash
# Warnungen als Fehler behandeln (Exit-Code 1 auch bei Warnungen)
smpe_lint --warnings-as-errors *.smpe
```

## Spezifische Checks deaktivieren

```bash
smpe_lint --disable missing_inline_data --disable content_beyond_column72 *.smpe
```

## Konfigurationsdatei

```bash
# Beispiel-Konfiguration erstellen
smpe_lint --init

# Konfigurationsdatei verwenden
smpe_lint --config .smpe_lint.yaml *.smpe
```

Beispiel `.smpe_lint.yaml`:

```yaml
disable:
  - missing_inline_data
  - content_beyond_column72
warnings_as_errors: false
```

## Pfad zur smpe.json (`--data`)

`smpe_lint` liest die Statement-Definitionen standardmäßig aus
`~/.local/share/smpe_ls/smpe.json`. In Umgebungen ohne nutzbares
Home-Verzeichnis (Docker-Container, CI-Runner) den Pfad explizit angeben:

```bash
smpe_lint --data data/smpe.json *.smpe
```

## Exit-Codes

| Code | Bedeutung |
|------|-----------|
| `0` | Keine Fehler gefunden |
| `1` | Fehler oder Warnungen gefunden (mit `--warnings-as-errors`) |
| `2` | Tool-Fehler (Datei nicht gefunden, ungültige Argumente) |

## Einsatz in CI/CD-Pipelines

```yaml
# GitHub Actions Beispiel
- name: Lint SMP/E files
  run: |
    smpe_lint --data data/smpe.json --warnings-as-errors *.smpe
```

## Zusammenfassung

- `smpe_lint` prüft SMP/E-Dateien auf der Kommandozeile
- Markdown- und JSON-Ausgabe für unterschiedliche Anwendungsfälle
- `--warnings-as-errors` für strikte Pipelines
- `--disable` zum Ausblenden einzelner Checks
- `--data` für den smpe.json-Pfad in Containern/CI
- Exit-Code `0` = keine Fehler, `1` = Fehler gefunden
