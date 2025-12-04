# TODO: GROSSES REFACTORING - Alte nicht-AST Dateien löschen

## Ziel
Alle alten nicht-AST-basierten Dateien löschen, die nicht mehr verwendet werden:
- **completion.go** - Nur Provider struct wird noch gebraucht, alle Completion-Funktionen sind tot
- **diagnostics.go** - Nur Provider struct und getRequiredOperands() werden noch gebraucht, alle Diagnostics-Funktionen sind tot

## Problem
Diese Dateien enthalten:
1. Provider structs (noch gebraucht von *_ast.go)
2. Hilfsfunktionen (teilweise noch gebraucht)
3. **TOTE CODE**: Alte Logik die nicht mehr verwendet wird

## Lösung - Completion Refactoring
- [x] Provider struct nach completion_ast.go verschieben
- [x] NewProvider() nach completion_ast.go verschieben
- [x] Prüfen welche Hilfsfunktionen noch verwendet werden und nach completion_ast.go verschieben
- [x] **completion.go löschen**

## Lösung - Diagnostics Refactoring
- [ ] Provider struct nach diagnostics_ast.go verschieben
- [ ] NewProvider() nach diagnostics_ast.go verschieben
- [ ] getRequiredOperands() nach diagnostics_ast.go verschieben (wird noch von diagnostics_ast.go aufgerufen)
- [ ] Prüfen welche Hilfsfunktionen noch verwendet werden und nach diagnostics_ast.go verschieben
- [ ] **diagnostics.go löschen**

## Zusätzlich zu prüfen
- [ ] Gibt es weitere alte nicht-AST Dateien die gelöscht werden können?
- [ ] cmd/test_* Programme die noch alte .Analyze() oder alte Completion-Funktionen verwenden → löschen oder auf AST umstellen

## Notizen
- handler.go verwendet nur GetCompletionsAST() und AnalyzeAST()
- Keine aktive Logik verwendet die alten Funktionen mehr
- **NACH ++MOVE FERTIGSTELLUNG DURCHFÜHREN**

