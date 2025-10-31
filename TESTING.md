# Testing Guide for SMP/E Language Server

## Prerequisites

Before testing, ensure you have:

1. Built and installed the language server:
   ```bash
   make install
   ```

2. Installed VSCode extension dependencies:
   ```bash
   make vscode-deps
   ```

3. Compiled the VSCode extension:
   ```bash
   make vscode-compile
   ```

Or simply run:
```bash
make vscode
```

## Testing in VSCode

### Step 1: Open VSCode Extension

```bash
cd client/vscode-smpe
code .
```

### Step 2: Launch Extension Development Host

1. In VSCode, press `F5` or:
   - Go to Run → Start Debugging
   - Select "Launch Extension"

2. A new VSCode window will open titled "[Extension Development Host]"

### Step 3: Open Example File

In the Extension Development Host window:

1. Open the examples folder from the project root:
   ```
   File → Open Folder → smpe_ls/examples
   ```

2. Open `sample.smpe`

### Step 4: Test Features

#### 1. Syntax Highlighting

The example file should show color coding for:
- MCS statements (++APAR, ++FUNCTION, etc.) in **keyword color**
- Operands (DESCRIPTION, FILES, etc.) in **property color**
- Strings in quotes in **string color**
- Comments (lines starting with *) in **comment color**

#### 2. Code Completion

Test completion by typing:

```smpe
++
```

You should see a completion list with all MCS statements.

After selecting a statement, type space and you should see operand completions:

```smpe
++APAR(TEST)
```

Type to trigger operand completions (DESCRIPTION, FILES, RFDSNPFX, REWORK).

#### 3. Hover Information

Hover your mouse over:
- A MCS statement (e.g., `++APAR`)
- An operand (e.g., `DESCRIPTION`)

You should see a popup with:
- Statement/operand name
- Description from smpe.json
- Parameter information
- For operands: type, length, allowed values

#### 4. Diagnostics

Create syntax errors to test diagnostics:

```smpe
++APAR
```
(Missing required parentheses)

```smpe
++ASSIGN SOURCEID(TEST)
```
(Missing required TO operand)

You should see:
- Red squiggly lines under errors
- Error messages in the Problems panel (View → Problems)

### Step 5: Check Server Logs

The server logs to `~/.local/share/smpe_ls/log`

View logs in real-time:
```bash
tail -f ~/.local/share/smpe_ls/log
```

Or with debug enabled:
1. In Extension Development Host, open settings:
   - Cmd+, (Mac) or Ctrl+, (Windows/Linux)
   - Search for "smpe"
   - Enable "Smpe: Debug"

2. Reload window:
   - Cmd+Shift+P → "Developer: Reload Window"

3. Check log file for detailed debug output

## Testing Checklist

### Basic Functionality

- [ ] Server starts without errors
- [ ] Extension activates when opening .smpe file
- [ ] Log file is created at ~/.local/share/smpe_ls/log

### Syntax Highlighting

- [ ] MCS statements are highlighted
- [ ] Operands are highlighted
- [ ] Strings are highlighted
- [ ] Comments are highlighted
- [ ] Numbers are highlighted

### Code Completion

- [ ] `++` triggers MCS statement completion
- [ ] Selecting statement shows operands
- [ ] Operands are context-specific to statement
- [ ] Completion inserts proper syntax (with parentheses)

### Hover Information

- [ ] Hovering over ++APAR shows description
- [ ] Hovering over ++ASSIGN shows description
- [ ] Hovering over ++DELETE shows description
- [ ] Hovering over ++FEATURE shows description
- [ ] Hovering over ++FUNCTION shows description
- [ ] Hovering over ++HOLD shows description
- [ ] Hovering over operands shows their description
- [ ] Hover shows parameter types and lengths

### Diagnostics

- [ ] Missing parentheses shows error
- [ ] Missing required operands shows error
- [ ] Errors appear in Problems panel
- [ ] Errors have correct line/column positions

## Common Issues

### Extension Not Activating

**Problem:** Extension doesn't activate when opening .smpe file

**Solutions:**
1. Check file extension is .smpe, .mcs, or .smp
2. Reload VSCode window: Cmd+Shift+P → "Developer: Reload Window"
3. Check Output panel: View → Output → "SMP/E Language Server"

### Server Not Found

**Problem:** "Cannot find server executable"

**Solutions:**
1. Ensure server is installed:
   ```bash
   which smpe_ls
   ```

2. If not in PATH, specify full path in settings:
   ```json
   {
     "smpe.serverPath": "/Users/YOUR_USER/.local/bin/smpe_ls"
   }
   ```

### No Completions

**Problem:** Code completion not working

**Solutions:**
1. Check server is running:
   - View → Output → "SMP/E Language Server"
   - Look for "Server started" message

2. Check cursor position (must be after ++ or operand name)

3. Try triggering manually: Ctrl+Space

### No Hover

**Problem:** Hover information not showing

**Solutions:**
1. Verify smpe.json exists in data/ directory
2. Check server log for errors loading smpe.json
3. Ensure hovering directly over keyword (not spaces)

### Data File Not Found

**Problem:** Server can't find data/smpe.json

**Solutions:**
1. Ensure data/smpe.json exists in project root
2. Run server with absolute path to data file:
   ```json
   {
     "smpe.serverPath": "smpe_ls --data /absolute/path/to/data/smpe.json"
   }
   ```

## Debugging Tips

### Enable Debug Logging

1. VSCode Settings:
   ```json
   {
     "smpe.debug": true
   }
   ```

2. Reload window

3. Check log:
   ```bash
   tail -f ~/.local/share/smpe_ls/log
   ```

### View LSP Communication

1. In Extension Development Host:
   - Help → Toggle Developer Tools
   - Console tab

2. Look for LSP messages (if using debug build)

### Test Parser Manually

Test the parser standalone:

```bash
cd /path/to/smpe_ls

# Create test file
cat > test.smpe << 'EOF'
++APAR(TEST123) DESCRIPTION('Test')
EOF

# Run server manually (will read from stdin)
./smpe_ls --debug
```

### Rebuild Everything

If things are broken:

```bash
# Clean everything
make clean
rm -rf client/vscode-smpe/node_modules
rm -rf client/vscode-smpe/out

# Rebuild from scratch
make vscode

# Reinstall server
make install

# Restart VSCode completely
```

## Performance Testing

### Large Files

Create a large test file:

```bash
for i in {1..1000}; do
  echo "++APAR(TEST$i) DESCRIPTION('Test $i')" >> large.smpe
done
```

Test:
- Opening large file (should be instant)
- Completion response time (should be < 100ms)
- Diagnostics update time (should be < 500ms)

### Multiple Files

Open multiple .smpe files and verify:
- Each gets syntax highlighting
- Completion works in all files
- Diagnostics are file-specific

## Success Criteria

The testing is successful if:

1. ✅ All items in Testing Checklist pass
2. ✅ No errors in server log (except expected parse errors)
3. ✅ Extension works without crashes
4. ✅ Features respond quickly (< 500ms)
5. ✅ Server handles invalid input gracefully

## Next Steps

After successful testing:

1. Document any issues found
2. Create GitHub issues for bugs
3. Update DEVELOPMENT_PLAN.md with lessons learned
4. Consider adding automated tests
