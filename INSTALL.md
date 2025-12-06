# Installation Guide

This guide explains how to install the SMP/E Language Server on different platforms.

## Table of Contents

- [Linux Installation](#linux-installation)
- [macOS Installation](#macos-installation)
- [Windows Installation](#windows-installation)
- [VSCode Extension Setup](#vscode-extension-setup)

---

## Linux Installation

### Download and Install Binary

**AMD64 (Intel/AMD):**

```bash
# Download latest release
curl -L https://github.com/cybersorcerer/smpe_ls/releases/latest/download/smpe_ls-v0.7.0-linux-amd64-full.tar.gz | tar xz

# Install binary
cd smpe_ls-v0.7.0-linux-amd64
sudo install -m 755 smpe_ls /usr/local/bin/

# Install data file
mkdir -p ~/.local/share/smpe_ls
cp smpe.json ~/.local/share/smpe_ls/

# Verify installation
smpe_ls --version
```

**ARM64 (Raspberry Pi, ARM servers):**

```bash
# Download latest release
curl -L https://github.com/cybersorcerer/smpe_ls/releases/latest/download/smpe_ls-v0.7.0-linux-arm64-full.tar.gz | tar xz

# Install binary
cd smpe_ls-v0.7.0-linux-arm64
sudo install -m 755 smpe_ls /usr/local/bin/

# Install data file
mkdir -p ~/.local/share/smpe_ls
cp smpe.json ~/.local/share/smpe_ls/

# Verify installation
smpe_ls --version
```

### Expected File Locations

- **Binary:** `/usr/local/bin/smpe_ls`
- **Data file:** `~/.local/share/smpe_ls/smpe.json`
- **Log file:** `~/.local/share/smpe_ls/smpe_ls.log`

---

## macOS Installation

### Download and Install Binary

**Apple Silicon (M1/M2/M3):**

```bash
# Download latest release
curl -L https://github.com/cybersorcerer/smpe_ls/releases/latest/download/smpe_ls-v0.7.0-macos-arm64-full.tar.gz | tar xz

# Install binary
cd smpe_ls-v0.7.0-macos-arm64
sudo install -m 755 smpe_ls /usr/local/bin/

# Install data file
mkdir -p ~/.local/share/smpe_ls
cp smpe.json ~/.local/share/smpe_ls/

# Verify installation
smpe_ls --version
```

**Intel (x86_64):**

```bash
# Download latest release
curl -L https://github.com/cybersorcerer/smpe_ls/releases/latest/download/smpe_ls-v0.7.0-macos-amd64-full.tar.gz | tar xz

# Install binary
cd smpe_ls-v0.7.0-macos-amd64
sudo install -m 755 smpe_ls /usr/local/bin/

# Install data file
mkdir -p ~/.local/share/smpe_ls
cp smpe.json ~/.local/share/smpe_ls/

# Verify installation
smpe_ls --version
```

### Expected File Locations

- **Binary:** `/usr/local/bin/smpe_ls`
- **Data file:** `~/.local/share/smpe_ls/smpe.json`
- **Log file:** `~/.local/share/smpe_ls/smpe_ls.log`

### macOS Security Note

If macOS blocks the binary:

```bash
# Allow the binary to run
xattr -d com.apple.quarantine /usr/local/bin/smpe_ls
```

Or go to **System Preferences → Security & Privacy → General** and click "Allow Anyway".

---

## Windows Installation

### Download and Install Binary

**AMD64 (Intel/AMD) - PowerShell:**

```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/cybersorcerer/smpe_ls/releases/latest/download/smpe_ls-v0.7.0-windows-amd64-full.zip" -OutFile "smpe_ls.zip"

# Extract archive
Expand-Archive smpe_ls.zip -DestinationPath $env:TEMP\smpe_ls

# Create installation directory
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\smpe_ls"

# Install binary
Copy-Item "$env:TEMP\smpe_ls\smpe_ls-v0.7.0-windows-amd64\smpe_ls.exe" -Destination "$env:LOCALAPPDATA\smpe_ls\"

# Install data file
Copy-Item "$env:TEMP\smpe_ls\smpe_ls-v0.7.0-windows-amd64\smpe.json" -Destination "$env:LOCALAPPDATA\smpe_ls\"

# Add to PATH
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:LOCALAPPDATA\smpe_ls", [EnvironmentVariableTarget]::User)

# Refresh PATH in current session
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User")

# Verify installation
smpe_ls --version
```

**ARM64 (Surface Pro X, ARM laptops) - PowerShell:**

```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/cybersorcerer/smpe_ls/releases/latest/download/smpe_ls-v0.7.0-windows-arm64-full.zip" -OutFile "smpe_ls.zip"

# Extract archive
Expand-Archive smpe_ls.zip -DestinationPath $env:TEMP\smpe_ls

# Create installation directory
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\smpe_ls"

# Install binary
Copy-Item "$env:TEMP\smpe_ls\smpe_ls-v0.7.0-windows-arm64\smpe_ls.exe" -Destination "$env:LOCALAPPDATA\smpe_ls\"

# Install data file
Copy-Item "$env:TEMP\smpe_ls\smpe_ls-v0.7.0-windows-arm64\smpe.json" -Destination "$env:LOCALAPPDATA\smpe_ls\"

# Add to PATH
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:LOCALAPPDATA\smpe_ls", [EnvironmentVariableTarget]::User)

# Refresh PATH in current session
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User")

# Verify installation
smpe_ls --version
```

### Expected File Locations

- **Binary:** `%LOCALAPPDATA%\smpe_ls\smpe_ls.exe` (e.g., `C:\Users\YourName\AppData\Local\smpe_ls\smpe_ls.exe`)
- **Data file:** `%LOCALAPPDATA%\smpe_ls\smpe.json`
- **Log file:** `%LOCALAPPDATA%\smpe_ls\smpe_ls.log`

### Manual Installation (Alternative)

1. Download the ZIP file for your platform
2. Extract to `C:\Program Files\smpe_ls\` or `%LOCALAPPDATA%\smpe_ls\`
3. Add the directory to your system PATH
4. Copy `smpe.json` to `%LOCALAPPDATA%\smpe_ls\`

---

## VSCode Extension Setup

After installing the language server, set up the VSCode extension:

### 1. Build the Extension

```bash
cd client/vscode-smpe
npm install
npm run compile
```

### 2. Configure VSCode Settings

Add to your VSCode `settings.json`:

**Linux/macOS:**
```json
{
  "smpe.serverPath": "/usr/local/bin/smpe_ls",
  "smpe.debug": false
}
```

**Windows:**
```json
{
  "smpe.serverPath": "C:\\Users\\YourName\\AppData\\Local\\smpe_ls\\smpe_ls.exe",
  "smpe.debug": false
}
```

### 3. Launch Extension Development Host

1. Open `client/vscode-smpe` in VSCode
2. Press `F5` to launch Extension Development Host
3. Open or create a `.smpe` file
4. Enjoy language features!

---

## Troubleshooting

### Server Not Found

**Linux/macOS:**
```bash
which smpe_ls
# Should output: /usr/local/bin/smpe_ls
```

**Windows:**
```powershell
where smpe_ls
# Should output: C:\Users\...\AppData\Local\smpe_ls\smpe_ls.exe
```

### Data File Not Found

Check if data file exists:

**Linux/macOS:**
```bash
ls -la ~/.local/share/smpe_ls/smpe.json
```

**Windows:**
```powershell
Test-Path "$env:LOCALAPPDATA\smpe_ls\smpe.json"
```

### Check Server Logs

**Linux/macOS:**
```bash
tail -f ~/.local/share/smpe_ls/smpe_ls.log
```

**Windows:**
```powershell
Get-Content "$env:LOCALAPPDATA\smpe_ls\smpe_ls.log" -Tail 20 -Wait
```

### Permission Issues (Linux/macOS)

```bash
sudo chmod +x /usr/local/bin/smpe_ls
```

---

## Uninstall

### Linux/macOS

```bash
sudo rm /usr/local/bin/smpe_ls
rm -rf ~/.local/share/smpe_ls
```

### Windows

```powershell
Remove-Item "$env:LOCALAPPDATA\smpe_ls" -Recurse -Force
# Then remove from PATH via System Environment Variables
```

---

## Build from Source

If you prefer to build from source:

```bash
# Clone repository
git clone https://github.com/cybersorcerer/smpe_ls.git
cd smpe_ls

# Install (builds and installs to ~/.local/bin)
make install

# Or build manually
go build -o smpe_ls ./cmd/smpe_ls

# Install data file
mkdir -p ~/.local/share/smpe_ls
cp data/smpe.json ~/.local/share/smpe_ls/
```

---

For more information, see the [README.md](README.md).
