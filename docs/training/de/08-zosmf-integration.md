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
    user: USERID
    rejectUnauthorized: true
    csi:
      - SYS1.SMPCSI
      - SYS1.SMPCSI2
    defaultCsi: SYS1.SMPCSI
    zones:
      - GLOBAL
      - MVST100
      - MVST200
      - MVSD100
      - MVSD200
    defaultZones:
      - GLOBAL
      - MVST*

# Optionaler Standard-Server (wenn nicht gesetzt, wird immer nachgefragt)
# defaultServer: MVSDEV
```

### Felder der Konfigurationsdatei

| Feld | Pflicht | Beschreibung |
|------|---------|--------------|
| `name` | ja | Anzeigename des Servers |
| `host` | ja | Hostname oder IP-Adresse des z/OSMF-Servers |
| `port` | ja | Port des z/OSMF-Servers (typisch: 443 oder 10443) |
| `user` | ja | z/OS Benutzer-ID |
| `rejectUnauthorized` | nein | TLS-Zertifikat prüfen (Standard: `true`) |
| `csi` | ja | Liste der CSI-Datasets (mindestens eine Angabe) |
| `defaultCsi` | nein | Standard-CSI — wird bei mehreren CSIs ohne Rückfrage verwendet |
| `zones` | nein | Vollständige Liste aller Zonen im System — ermöglicht Wildcard-Auflösung |
| `defaultZones` | nein | Vorausgefüllte Zonen beim Abfrage-Dialog — unterstützt Wildcards (`*`, `?`) |
| `defaultServer` | nein | Name des Standard-Servers (auf oberster Ebene, außerhalb von `servers`) |

### Zonen und Wildcard-Auflösung

Die z/OSMF CSI Query API (GIMAPI) unterstützt kein Wildcard-Matching in Zonennamen —
jede Zone muss exakt und einzeln angegeben werden. Die Extension löst dieses Problem
clientseitig:

- In `zones` werden alle im System vorhandenen Zonennamen hinterlegt.
- In `defaultZones` (und im Abfrage-Dialog) können Wildcards verwendet werden:
  - `*` steht für beliebig viele Zeichen
  - `?` steht für genau ein Zeichen
- Die Extension löst Wildcards gegen die `zones`-Liste auf und übergibt der API
  die resultierenden exakten Zonennamen.

**Beispiel:** `defaultZones: [GLOBAL, MVST*]` mit `zones: [GLOBAL, MVST100, MVST200, MVSD100]`
ergibt bei der Abfrage die Zonen `GLOBAL`, `MVST100` und `MVST200`.

Wenn keine `zones`-Liste konfiguriert ist, werden Zonennamen unverändert an die API
übergeben — Wildcards funktionieren dann nicht.

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

Öffnet eine vollständige WebUI für freie CSI-Abfragen:

- **Entry Type** frei wählbar (SYSMOD, DDDEF, MOD, UNIX1, ...)
- **Subentries** per Picker auswählen
- **Filter** als CSI-Filterausdruck (z.B. `ENAME='UA12345'`)
- **Mehrere Zonen** gleichzeitig abfragen
- Ergebnisse exportierbar als **JSON** oder **CSV**
- Spaltenüberschriften bleiben beim Scrollen sichtbar (Sticky Header)

### HOLD Comments Viewer

In SYSMOD-Abfragen (GLOBAL Zone) erscheint in Zeilen mit HOLDDATA ein **HOLD**-Button.
Ein Klick lädt das PTF-Member aus SMPPTS via z/OSMF und zeigt den vollständigen
`++HOLD`-Block (inkl. COMMENT-Text) in einem Seitenpanel an.

Wenn der PTF bereits akzeptiert wurde und nicht mehr in SMPPTS vorhanden ist,
erscheint eine entsprechende Fehlermeldung.

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
- Free Form CSI Query ermöglicht beliebige CSI-Abfragen mit Export
- HOLD Comments Viewer zeigt den vollständigen `++HOLD`-Block aus SMPPTS
- CodeLens-Links ermöglichen Inline-Queries
- USS-Verzeichnisse und MVS-Datasets können durchsucht werden
- `SMP/E: Check Missing Input Members` prüft ob alle Input-Dateien vorhanden sind
