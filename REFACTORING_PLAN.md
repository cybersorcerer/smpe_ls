# Refactoring Plan: Elimination of completion.go and diagnostics.go

## Goal
Delete completion.go and diagnostics.go completely while maintaining full functionality.

## Current State Analysis

### completion.go Dependencies
**USED:**
- `type Provider struct` - used by handler.go and completion_ast.go methods
- `NewProvider()` - used by handler.go line 45

**DEAD CODE (not used anywhere):**
- `getMCSCompletions()` - ALREADY MOVED to completion_ast.go
- `getReasonCompletions()`
- `extractSysmodIDs()`
- `getOperandCompletions()`
- `isOperandChar()`
- `parseOperands()`
- `tokenize()`
- `isOperandName()`
- `isStatementName()`

### diagnostics.go Dependencies
**USED:**
- `type Provider struct` - used by handler.go and diagnostics_ast.go methods
- `NewProvider()` - used by handler.go line 46
- `getRequiredOperands()` - used by diagnostics_ast.go line 244

**DEAD CODE (not used anywhere):**
- `Analyze()` - old non-AST analyzer
- All other functions

## Refactoring Steps (in order)

### Phase 1: Refactor completion package
1. ✅ Move `getMCSCompletions()` to completion_ast.go (DONE)
2. Move `type Provider struct` to completion_ast.go
3. Move `NewProvider()` to completion_ast.go
4. Update handler.go imports (if needed)
5. Delete completion.go
6. Build and test

### Phase 2: Refactor diagnostics package
7. Move `type Provider struct` to diagnostics_ast.go
8. Move `NewProvider()` to diagnostics_ast.go
9. Move `getRequiredOperands()` to diagnostics_ast.go
10. Update handler.go imports (if needed)
11. Delete diagnostics.go
12. Build and test

### Phase 3: Final verification
13. Run all tests
14. Verify VSCode extension works
15. Check for any remaining references
16. Update documentation
