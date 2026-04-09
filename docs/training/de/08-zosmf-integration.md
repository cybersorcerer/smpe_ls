# Modul 08 — z/OSMF Integration

## Überblick

Die Extension kann SYSMODs, DDDEFs und Zonen über die z/OSMF REST API abfragen,
USS-Verzeichnisse und MVS-Datasets durchsuchen, und fehlende Input-Member prüfen.
Voraussetzung ist eine z/OSMF-Konfigurationsdatei.

## Voraussetzungen

- Extension installiert — siehe [Modul 01](01-installation.md)
- Zugang zu einem z/OSMF-Server

## Konfiguration erstellen

`Cmd+Shift+P` → `SMP/E: Create z/OSMF Config`

Dies erstellt eine `.smpe-zosmf.yaml`-Datei im Workspace-Root und öffnet sie zur
Bearbeitung. Beispiel-Konfiguration:

```yaml
servers:
  - name: MVSDEV
    host: mvsdev.example.com
    port: 10443
    zosmfBase: /zosmf/services/v1
    zones:
      - TZONE
      - DZONE
    csis:
      - name: CSI
        dataset: SYS1.SMPCSI
        defaultCsi: true
```

Die Konfigurationsdatei wird in folgender Reihenfolge gesucht:

1. Alle offenen Workspace-Ordner (erster Treffer gewinnt)
2. `~/.config/smpe_ls/.smpe-zosmf.yaml` (globaler Fallback)

## Verfügbare Commands

### SYSMOD abfragen

`Cmd+Shift+P` → `SMP/E: Query SYSMOD via z/OSMF`

Gibt Informationen zu einem SYSMOD zurück: Status, FMID, PRE/REQ/SUP-Abhängigkeiten,
Holddata.

### DDDEF abfragen

`Cmd+Shift+P` → `SMP/E: Query DDDEF via z/OSMF`

Gibt DDDEF-Informationen zurück: Dataset-Name, Typ, Path, INITDISP. Klickbare
PATH-Links öffnen den USS-Verzeichnis-Browser, klickbare DATASET-Links öffnen
den MVS-Dataset-Browser.

### Zonen auflisten

`Cmd+Shift+P` → `SMP/E: List Zones via z/OSMF`

Listet alle konfigurierten Zonen mit ihren Eigenschaften.

### Free Form CSI Query

`Cmd+Shift+P` → `SMP/E: Free Form CSI Query`

Öffnet ein Eingabeformular für freie CSI-Abfragen: Zone, Entry Type, Subentries
und Filter können frei eingegeben werden.

## CodeLens

In `.smpe`-Dateien erscheinen über SYSMODs und FMIDs klickbare Links:

```smpe
++VER(Z038)
  [Query SYSMOD] [Query DDDEF]   ← CodeLens-Links
    FMID(HBB7790)
    PRE(UJ12345).
```

Ein Klick führt direkt eine z/OSMF-Abfrage aus.

## USS-Verzeichnis-Browser

Klickbare PATH-Links in DDDEF-Ergebnissen öffnen einen Verzeichnis-Browser:

- Navigation mit Breadcrumb-Leiste
- Dateien können direkt im Editor geöffnet werden (read-only)
- PATHPREFIX-Segmente werden automatisch aufgelöst

## MVS-Dataset-Browser

Klickbare DATASET-Links öffnen PDS-Member-Listen oder sequentielle Datasets:

- Member-Liste mit ISPF-Attributen (User, Created, Modified, Ver, Mod)
- Member können direkt im Editor geöffnet werden (read-only)

## Fehlende Input-Member prüfen

`Cmd+Shift+P` → `SMP/E: Check Missing Input Members`

Oder: Rechtsklick auf eine `.smpe`-Datei → `SMP/E: Check Missing Input Members`

Die Extension analysiert alle `.smpe`-Dateien im Workspace und prüft ob die
referenzierten Input-Member-Dateien vorhanden sind. Das Ergebnis erscheint in
einer Tabelle:

| SMP/E File | Statement | Member | Found |
|------------|-----------|--------|-------|
| mymod.smpe | ++PARM | RSSPRMI.parm | No |
| mymod.smpe | ++SRC | MYSRC.hlasm | Yes |

Die Tabelle kann nach allen Spalten sortiert und gefiltert werden (z.B. nur
fehlende Member anzeigen).

**Konfiguration:**

```json
{
  "smpe.checkMissingInputMembers.searchFolders": ["customization", "source"]
}
```

Standard: `["customization"]`. Mit `"*"` wird der gesamte Workspace durchsucht.

## Zusammenfassung

- `.smpe-zosmf.yaml` konfiguriert den z/OSMF-Zugang
- SYSMOD, DDDEF und Zonen können direkt aus VSCode abgefragt werden
- CodeLens-Links ermöglichen Inline-Queries
- USS-Verzeichnisse und MVS-Datasets können durchsucht werden
- `SMP/E: Check Missing Input Members` prüft ob alle Input-Dateien vorhanden sind
