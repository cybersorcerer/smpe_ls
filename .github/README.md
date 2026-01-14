# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated building and testing.

**IMPORTANT:** The release packages automatically include the correct `smpe.json` data file for each platform in the correct location.

## Workflows

### 1. `build.yml` - Continuous Integration

**Trigger:**

- Push to `main` branch
- Pull requests to `main` branch

**Jobs:**

- **test** - Runs Go unit tests with coverage report
- **build** - Builds the binary for the current platform
- **lint** - Runs golangci-lint

### 2. `release.yml` - Release Build

**Trigger:**

- Push of version tags (e.g. `v0.7.0`)
- Push to `main` branch (build only, no release)

**Jobs:**

- **build** - Cross-compilation for all platforms:
  - Linux AMD64
  - Linux ARM64
  - macOS Apple Silicon (ARM64)
  - macOS Intel (AMD64)
  - Windows AMD64
  - Windows ARM64

- **build-vsix** - Builds VS Code extension packages for all platforms

- **release** - Automatically creates a GitHub Release with all binaries and VSIX packages (only for tags)

- **test** - Runs all tests (unit tests + test suite)

## Usage

### Creating an Automatic Release

1. Create and push a tag:

```bash
git tag -a v0.7.0 -m "Release version 0.7.0"
git push origin v0.7.0
```

2. GitHub Actions automatically builds:
   - All 6 platform binaries
   - All 6 platform-specific VSIX packages
   - Creates release packages (tar.gz for Linux/macOS, zip for Windows)
   - Creates a GitHub Release with all artifacts

3. Release is available at: `https://github.com/cybersorcerer/smpe_ls/releases`

### Pre-releases

Tags containing "alpha" or "beta" are automatically marked as pre-releases:

```bash
git tag v0.7.0-alpha
git tag v0.8.0-beta.1
```

### Local Build for All Platforms

```bash
# Build all binaries
make build-all

# Create release packages
make release

# Build all VSIX packages
make package-all
```

## Artifacts

After a successful build, binaries are stored as artifacts:

**For Tags (Release):**

- Packages available under GitHub Releases
- Format: `smpe_ls-VERSION-PLATFORM.tar.gz` or `.zip`
- VSIX format: `vscode-smpe-PLATFORM-VERSION.vsix`

**For Main Branch:**

- Artifacts available for 30 days
- Download via GitHub Actions UI

## Package Format

Each release package contains:

```text
smpe_ls-v0.7.0-linux-amd64/
├── smpe_ls           # Binary
├── smpe.json         # Statement definitions
└── README.md         # Documentation
```

## Platform Overview

| Platform | GOOS | GOARCH | Binary Name |
|----------|------|--------|-------------|
| Linux AMD64 | linux | amd64 | smpe_ls |
| Linux ARM64 | linux | arm64 | smpe_ls |
| macOS Apple Silicon | darwin | arm64 | smpe_ls |
| macOS Intel | darwin | amd64 | smpe_ls |
| Windows AMD64 | windows | amd64 | smpe_ls.exe |
| Windows ARM64 | windows | arm64 | smpe_ls.exe |

## VSIX Platform Overview

| Platform | Target | File |
|----------|--------|------|
| Windows x64 | win32-x64 | vscode-smpe-win32-x64-VERSION.vsix |
| Windows ARM64 | win32-arm64 | vscode-smpe-win32-arm64-VERSION.vsix |
| macOS Apple Silicon | darwin-arm64 | vscode-smpe-darwin-arm64-VERSION.vsix |
| macOS Intel | darwin-x64 | vscode-smpe-darwin-x64-VERSION.vsix |
| Linux x64 | linux-x64 | vscode-smpe-linux-x64-VERSION.vsix |
| Linux ARM64 | linux-arm64 | vscode-smpe-linux-arm64-VERSION.vsix |
