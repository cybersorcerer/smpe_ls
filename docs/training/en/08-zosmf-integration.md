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

# Optional default server (if not set, user is always prompted)
# defaultServer: MVSDEV
```

### Configuration File Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Display name of the server |
| `host` | yes | Hostname or IP address of the z/OSMF server |
| `port` | yes | Port of the z/OSMF server (typically 443 or 10443) |
| `user` | yes | z/OS user ID |
| `rejectUnauthorized` | no | Verify TLS certificate (default: `true`) |
| `csi` | yes | List of CSI datasets (at least one required) |
| `defaultCsi` | no | Default CSI — used without prompting when multiple CSIs are configured |
| `zones` | no | Complete list of all zones in the system — enables wildcard resolution |
| `defaultZones` | no | Pre-filled zones in the query dialog — supports wildcards (`*`, `?`) |
| `defaultServer` | no | Name of the default server (top-level field, outside `servers`) |

### Zones and Wildcard Resolution

The z/OSMF CSI Query API (GIMAPI) does not support wildcard matching in zone names —
each zone must be specified exactly and individually. The extension solves this
client-side:

- `zones` holds the complete list of all zone names present in the system.
- `defaultZones` (and the query dialog) support wildcards:
  - `*` matches any number of characters
  - `?` matches exactly one character
- The extension resolves wildcards against the `zones` list and passes the resulting
  exact zone names to the API.

**Example:** `defaultZones: [GLOBAL, MVST*]` with `zones: [GLOBAL, MVST100, MVST200, MVSD100]`
results in the query being executed for zones `GLOBAL`, `MVST100`, and `MVST200`.

If no `zones` list is configured, zone names are passed to the API unchanged —
wildcards will not work in that case.

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

Opens a full WebUI for free-form CSI queries:

- **Entry type** freely selectable (SYSMOD, DDDEF, MOD, UNIX1, ...)
- **Subentries** selectable via picker
- **Filter** as a CSI filter expression (e.g. `ENAME='UA12345'`)
- **Multiple zones** can be queried simultaneously
- Results exportable as **JSON** or **CSV**
- Column headers remain visible while scrolling (sticky header)

### Filter History

Filter expressions are automatically saved (max. 20 entries, no duplicates). The **▼** button to the right of the filter input opens a dropdown with all saved entries — clicking an entry copies it into the input field.

To manage saved filters:

`Cmd+Shift+P` → `SMP/E: Manage Filter History`

- Select an entry → **Edit** or **Delete**
- **Clear All** removes the entire history
- Changes are immediately reflected in the open Free Form Query panel

### HOLD Comments Viewer

In SYSMOD queries (GLOBAL zone), rows with HOLDDATA show a **HOLD** button.
Clicking it fetches the PTF member from SMPPTS via z/OSMF and displays the complete
`++HOLD` block (including the COMMENT text) in a side panel.

If the PTF has already been accepted and is no longer present in SMPPTS, a clear
error message is shown instead.

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
- Free Form CSI Query enables arbitrary CSI queries with export
- Filter History automatically saves used filters (▼ button, max. 20 entries)
- HOLD Comments Viewer shows the complete `++HOLD` block from SMPPTS
- CodeLens links enable inline queries
- USS directories and MVS datasets can be browsed
- `SMP/E: Check Missing Input Members` checks whether all input files are present
