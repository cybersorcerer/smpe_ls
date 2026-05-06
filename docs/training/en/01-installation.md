# Module 01 — Installing the VSCode Extension

## Overview

This module shows how to install the SMP/E Language Server extension in Visual Studio
Code. After installation, VSCode automatically recognizes `.smpe` files and provides
all language server features.

## Prerequisites

- Visual Studio Code version 1.75 or later installed
- VSIX file for your platform downloaded (see below)

## Downloading the VSIX File

Download the correct VSIX file from the
[GitHub Releases page](https://github.com/cybersorcerer/smpe_ls/releases/latest):

| Platform | File |
|----------|------|
| Windows x64 | `smpe-mcs-language-server-win32-x64-1.1.0.vsix` |
| Windows ARM64 | `smpe-mcs-language-server-win32-arm64-1.1.0.vsix` |
| macOS Apple Silicon | `smpe-mcs-language-server-darwin-arm64-1.1.0.vsix` |
| macOS Intel | `smpe-mcs-language-server-darwin-x64-1.1.0.vsix` |
| Linux x64 | `smpe-mcs-language-server-linux-x64-1.1.0.vsix` |
| Linux ARM64 | `smpe-mcs-language-server-linux-arm64-1.1.0.vsix` |

The language server binary is already bundled inside the extension — no separate
installation required.

## Installation via VSCode UI

1. Open VSCode
2. Press `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)
3. Type `Extensions: Install from VSIX...` and confirm
4. Select the downloaded `.vsix` file

## Installation via Terminal

```bash
code --install-extension smpe-mcs-language-server-darwin-arm64-1.1.0.vsix
```

Then reload VSCode.

## Verifying the Installation

After reloading, a confirmation message appears in the Output panel:

1. Open `View` → `Output`
2. Select `SMP/E Language Server` from the dropdown on the right
3. You should see:

```
[2026-04-09T...] SMP/E Language Server extension activating...
[2026-04-09T...] Starting server: .../smpe_ls
```

## Key Settings

These settings can be configured in `.vscode/settings.json`:

| Setting | Default | Description |
|---------|---------|-------------|
| `smpe.serverPath` | `""` | Path to the `smpe_ls` binary (empty = bundled) |
| `smpe.dataPath` | `""` | Path to the `smpe.json` data file (empty = bundled) |
| `smpe.outlPath` | `""` | Path to the `smpe_outl` binary (empty = bundled) |
| `smpe.debug` | `false` | Enable debug logging |

## Summary

- Download the correct VSIX file for your platform
- Install via VSCode UI or terminal
- Do not forget to **Reload Window** after installation
- Verify installation via the Output panel `SMP/E Language Server`
