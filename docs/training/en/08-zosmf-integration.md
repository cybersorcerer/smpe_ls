# Module 08 — z/OSMF Integration

## Overview

The extension can query SYSMODs, DDDEFs, and zones via the z/OSMF REST API, browse
USS directories and MVS datasets, and check for missing input members. A z/OSMF
configuration file is required.

## Prerequisites

- Extension installed — see [Module 01](01-installation.md)
- Access to a z/OSMF server

## Creating the Configuration

`Cmd+Shift+P` → `SMP/E: Create z/OSMF Config`

This creates a `.smpe-zosmf.yaml` file in the workspace root and opens it for
editing. Example configuration:

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

The configuration file is resolved in this order:

1. All open workspace folders (first match wins)
2. `~/.config/smpe_ls/.smpe-zosmf.yaml` (global fallback)

## Available Commands

### Query SYSMOD

`Cmd+Shift+P` → `SMP/E: Query SYSMOD via z/OSMF`

Returns information about a SYSMOD: status, FMID, PRE/REQ/SUP dependencies, holddata.

### Query DDDEF

`Cmd+Shift+P` → `SMP/E: Query DDDEF via z/OSMF`

Returns DDDEF information: dataset name, type, path, INITDISP. Clickable PATH links
open the USS directory browser; clickable DATASET links open the MVS dataset browser.

### List Zones

`Cmd+Shift+P` → `SMP/E: List Zones via z/OSMF`

Lists all configured zones with their properties.

### Free Form CSI Query

`Cmd+Shift+P` → `SMP/E: Free Form CSI Query`

Opens an input form for free-form CSI queries: zone, entry type, subentries, and
filter can be entered freely.

## CodeLens

In `.smpe` files, clickable links appear above SYSMODs and FMIDs:

```smpe
++VER(Z038)
  [Query SYSMOD] [Query DDDEF]   ← CodeLens links
    FMID(HBB7790)
    PRE(UJ12345).
```

A click runs a z/OSMF query directly.

## USS Directory Browser

Clickable PATH links in DDDEF results open a directory browser:

- Navigation via breadcrumb bar
- Files can be opened directly in the editor (read-only)
- PATHPREFIX segments are resolved automatically

## MVS Dataset Browser

Clickable DATASET links open PDS member lists or sequential datasets:

- Member list with ISPF attributes (User, Created, Modified, Ver, Mod)
- Members can be opened directly in the editor (read-only)

## Check Missing Input Members

`Cmd+Shift+P` → `SMP/E: Check Missing Input Members`

Or: right-click on a `.smpe` file → `SMP/E: Check Missing Input Members`

The extension analyzes all `.smpe` files in the workspace and checks whether the
referenced input member files exist. The result appears in a table:

| SMP/E File | Statement | Member | Found |
|------------|-----------|--------|-------|
| mymod.smpe | ++PARM | RSSPRMI.parm | No |
| mymod.smpe | ++SRC | MYSRC.hlasm | Yes |

The table can be sorted and filtered by all columns (e.g. show only missing members).

**Configuration:**

```json
{
  "smpe.checkMissingInputMembers.searchFolders": ["customization", "source"]
}
```

Default: `["customization"]`. Use `"*"` to search the entire workspace.

## Summary

- `.smpe-zosmf.yaml` configures z/OSMF access
- SYSMODs, DDDEFs, and zones can be queried directly from VSCode
- CodeLens links enable inline queries
- USS directories and MVS datasets can be browsed
- `SMP/E: Check Missing Input Members` checks whether all input files are present
