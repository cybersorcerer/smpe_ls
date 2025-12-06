# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated building and testing.

**IMPORTANT:** The release packages automatically include the correct `smpe.json` data file for each platform in the correct location.

## Workflows

### 1. `build.yml` - Continuous Integration

**Trigger:**
- Push auf `main` Branch
- Pull Requests auf `main` Branch

**Jobs:**
- **test** - Führt Go Unit-Tests aus mit Coverage-Report
- **build** - Baut die Binary für die aktuelle Plattform
- **lint** - Führt golangci-lint aus

### 2. `release.yml` - Release Build

**Trigger:**
- Push von Version-Tags (z.B. `v0.6.1`)
- Push auf `main` Branch (nur Build, kein Release)

**Jobs:**
- **build** - Cross-Compilation für alle Plattformen:
  - Linux AMD64
  - Linux ARM64
  - macOS Apple Silicon (ARM64)
  - macOS Intel (AMD64)
  - Windows AMD64
  - Windows ARM64

- **release** - Erstellt automatisch einen GitHub Release mit allen Binaries (nur bei Tags)

- **test** - Führt alle Tests aus (Unit-Tests + Test-Suite)

## Verwendung

### Automatisches Release erstellen

1. Tag erstellen und pushen:
```bash
git tag -a v0.6.1 -m "Release version 0.6.1"
git push origin v0.6.1
```

2. GitHub Actions baut automatisch:
   - Alle 6 Plattform-Binaries
   - Erstellt Release-Packages (tar.gz für Linux/macOS, zip für Windows)
   - Erstellt einen GitHub Release mit allen Artifacts

3. Release ist verfügbar unter: `https://github.com/DEIN-USER/smpe_ls/releases`

### Lokales Build für alle Plattformen

```bash
# Alle Binaries bauen
make build-all

# Release-Packages erstellen
make release
```

## Artifacts

Nach erfolgreichem Build werden die Binaries als Artifacts gespeichert:

**Bei Tags (Release):**
- Packages verfügbar unter GitHub Releases
- Format: `smpe_ls-VERSION-PLATFORM.tar.gz` oder `.zip`

**Bei Main-Branch:**
- Artifacts verfügbar für 30 Tage
- Download über GitHub Actions UI

## Package-Format

Jedes Release-Package enthält:
```
smpe_ls-v0.6.1-linux-amd64/
├── smpe_ls           # Binary
├── smpe.json         # Statement definitions
└── README.md         # Dokumentation
```

## Plattform-Übersicht

| Plattform | GOOS | GOARCH | Binary Name |
|-----------|------|--------|-------------|
| Linux AMD64 | linux | amd64 | smpe_ls |
| Linux ARM64 | linux | arm64 | smpe_ls |
| macOS Apple Silicon | darwin | arm64 | smpe_ls |
| macOS Intel | darwin | amd64 | smpe_ls |
| Windows AMD64 | windows | amd64 | smpe_ls.exe |
| Windows ARM64 | windows | arm64 | smpe_ls.exe |
