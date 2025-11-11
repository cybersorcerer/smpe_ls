# REFACTORING-PLAN: Struktureller Parser für SMP/E Language Server

## Ziel
Ersetze String-Matching-basierte Parsing-Logik durch einen robusten, strukturellen Parser der auf Syntax-Diagrammen basiert und einen AST erstellt für alle LSP-Features.

---

## Phase 1: Parser & AST Grundlagen

### 1.1 Neue Package-Struktur
**Datei:** `internal/parser/parser.go`

**AST Node-Typen:**
```go
type NodeType int

const (
    NodeTypeDocument      // Gesamtes Dokument
    NodeTypeStatement     // MCS Statement (++USERMOD, ++JAR, etc.)
    NodeTypeOperand       // Operand (DESC, REWORK, FROMDS, etc.)
    NodeTypeParameter     // Parameter-Wert
    NodeTypeComment       // Kommentar
)

type Position struct {
    Line      int
    Character int
    Length    int
}

type Node struct {
    Type     NodeType
    Name     string           // Statement/Operand Name
    Value    string           // Parameter-Wert
    Position Position
    Parent   *Node
    Children []*Node

    // Semantische Referenzen
    StatementDef *data.MCSStatement  // Referenz zu smpe.json
    OperandDef   *data.Operand       // Referenz zu smpe.json
}

type Document struct {
    Statements []*Node
    Errors     []ParseError
}
```

### 1.2 Parser-Logik
**Basierend auf Syntax-Diagrammen:**

1. **Statement Parser:**
   - Erkennt `++STATEMENT`
   - Prüft in smpe.json ob Statement existiert
   - Parsed Statement-Parameter (falls `MCSStatement.Parameter` in smpe.json definiert)
   - Parsed Operanden basierend auf `MCSStatement.Operands`

2. **Operand Parser:**
   - Erkennt Operand-Name (kontext-basiert, nicht case-basiert)
   - Prüft in smpe.json ob Operand für aktuelles Statement existiert
   - Wenn `Operand.Values` existiert → parsed Sub-Operanden rekursiv
   - Sonst → parsed einfachen Parameter

3. **Data Element Parser (++MAC, ++SRC, etc.):**
   - Erkennt 3-stelligen Language-ID
   - Validiert Language-ID gegen `data.LanguageVariants`

---

## Phase 2: Migration - Semantic Tokens

### 2.1 Neue Implementierung
**Datei:** `internal/semantic/semantic.go` (Refactored)

**Änderungen:**
- Entferne `tokenizeLine`, `tokenizeOperands`
- Neue Funktion: `BuildTokensFromAST(doc *parser.Document) []Token`
- Traversiert AST und erstellt Tokens basierend auf `NodeType`:
  - `NodeTypeStatement` → `TokenTypeKeyword`
  - `NodeTypeOperand` → `TokenTypeFunction`
  - `NodeTypeParameter` → `TokenTypeParameter`

**Vorteile:**
- Keine Groß-/Kleinschreibungs-Logik
- Statement-Parameter werden korrekt klassifiziert
- Sub-Operanden (FROMDS) funktionieren automatisch

### 2.2 Testing
- Test mit `usermod.smpe`: `++USERMOD(LJS2012)` Parameter ist orange, nicht blau
- Test mit Data Elements: `++MAC(ASM)` Language-ID korrekt erkannt

---

## Phase 3: Migration - Completion

### 3.1 Aktuelle Implementierung analysieren
**Datei:** `internal/completion/completion.go`

**Was bleibt:**
- Context-Detection (cursor position → Statement/Operand/Parameter)
- Filtering-Logik (mutually_exclusive, allowed_if)

**Was wird ersetzt:**
- String-basiertes Statement-Erkennung → AST-basiert
- Position-basiertes Operand-Matching → AST Node Lookup

### 3.2 Neue Implementierung
```go
func (p *Provider) Provide(doc *parser.Document, position lsp.Position) []lsp.CompletionItem {
    // 1. Finde Node an Cursor-Position
    node := findNodeAtPosition(doc, position)

    // 2. Bestimme Kontext
    if node.Type == NodeTypeStatement && position in node.Parameter {
        return provideStatementParameterCompletion(node)
    }
    if node.Type == NodeTypeOperand && position in node.Parameter {
        if node.OperandDef.Values != nil {
            // Sub-Operanden vorschlagen
            return provideSubOperandCompletion(node)
        }
        // Einfacher Parameter, keine Completion
        return nil
    }
    // Operand-Level
    return provideOperandCompletion(node.Parent.StatementDef)
}
```

### 3.3 Testing
- Context-aware completion funktioniert weiterhin
- Nach `++USERMOD(` → keine Operand-Completion
- Innerhalb `FROMDS(` → DSN, VOLUME, UNIT, NUMBER vorgeschlagen

---

## Phase 4: Migration - Diagnostics

### 4.1 Aktuelle Implementierung analysieren
**Datei:** `internal/diagnostics/diagnostics.go`

**Was bleibt:**
- Required Operand Validation
- Parameter Validation
- Mutually Exclusive Validation

**Was wird ersetzt:**
- Statement Parsing → benutzt AST
- Operand Extraction → benutzt AST

### 4.2 Neue Implementierung
```go
func (d *DiagnosticsProvider) Analyze(doc *parser.Document) []lsp.Diagnostic {
    diagnostics := []lsp.Diagnostic{}

    for _, stmt := range doc.Statements {
        // Required Operands Check
        diagnostics = append(diagnostics, validateRequiredOperands(stmt)...)

        // Mutually Exclusive Check
        diagnostics = append(diagnostics, validateMutuallyExclusive(stmt)...)

        // Parameter Validation
        for _, operand := range stmt.Children {
            diagnostics = append(diagnostics, validateOperandParameter(operand)...)
        }
    }

    return diagnostics
}
```

### 4.3 Testing
- Required operands werden weiterhin erkannt
- Multiline comments erzeugen keine false positives

---

## Phase 5: Migration - Hover

### 5.1 Aktuelle Implementierung analysieren
**Datei:** `internal/hover/hover.go`

**Was wird ersetzt:**
- `extractWord()` → AST Node Lookup
- String-Matching gegen smpe.json → direkte Referenz via `Node.StatementDef` / `Node.OperandDef`

### 5.2 Neue Implementierung
```go
func (h *HoverProvider) Provide(doc *parser.Document, position lsp.Position) *lsp.Hover {
    node := findNodeAtPosition(doc, position)

    if node == nil {
        return nil
    }

    switch node.Type {
    case NodeTypeStatement:
        return buildStatementHover(node.StatementDef)
    case NodeTypeOperand:
        return buildOperandHover(node.OperandDef)
    case NodeTypeParameter:
        // Kein Hover für Parameter-Werte
        return nil
    }

    return nil
}
```

### 5.3 Testing
- Hover auf `++USERMOD` zeigt Statement-Info
- Hover auf `DESC` zeigt Operand-Info
- Hover auf Parameter-Wert zeigt nichts

---

## Phase 6: Code Cleanup

### 6.1 Zu entfernende Funktionen

**In `internal/semantic/semantic.go`:**
- `tokenizeLine()`
- `tokenizeOperands()`
- `isOperandChar()` (wird in Parser verschoben)

**In `internal/completion/completion.go`:**
- `extractStatementName()` (ersetzt durch AST)
- `findCurrentStatement()` (ersetzt durch AST)

**In `internal/diagnostics/diagnostics.go`:**
- `parseStatement()` (ersetzt durch Parser)
- `extractOperands()` (ersetzt durch AST)

**In `internal/hover/hover.go`:**
- `extractWord()` (ersetzt durch AST)

### 6.2 Verification
- `git grep "tokenizeLine"` → keine Treffer
- `git grep "extractWord"` → keine Treffer
- Alle Tests laufen durch

---

## Phase 7: Handler Integration

### 7.1 Zentraler Parser-Aufruf
**Datei:** `internal/handler/handler.go`

**Änderungen:**
```go
type Handler struct {
    parser              *parser.Parser
    semanticProvider    *semantic.Provider
    completionProvider  *completion.Provider
    diagnosticsProvider *diagnostics.Provider
    hoverProvider       *hover.Provider

    // Cache für parsed documents
    documents map[string]*parser.Document
}

func (h *Handler) TextDocumentDidOpen(params lsp.DidOpenTextDocumentParams) {
    // Parse document
    doc := h.parser.Parse(params.TextDocument.Text)
    h.documents[params.TextDocument.URI] = doc

    // Generate diagnostics from AST
    diagnostics := h.diagnosticsProvider.Analyze(doc)
    h.sendDiagnostics(params.TextDocument.URI, diagnostics)
}

func (h *Handler) TextDocumentSemanticTokensFull(params lsp.SemanticTokensParams) (*lsp.SemanticTokens, error) {
    doc := h.documents[params.TextDocument.URI]
    tokens := h.semanticProvider.BuildTokensFromAST(doc)
    return &lsp.SemanticTokens{Data: tokens}, nil
}
```

---

## Phase 8: Data Element Support

### 8.1 Language-ID Parser
**Datei:** `internal/langid/langid.go` (NEU)

```go
type LanguageID struct {
    Code     string    // z.B. "ASM", "C", "PLI"
    Position Position
}

func ParseLanguageID(text string, pos int) (*LanguageID, error) {
    // Extrahiert 3-stelligen Language-Code
    // Validiert gegen data.LanguageVariants
}
```

### 8.2 Parser Integration
- Data Element Statements (++MAC, ++SRC, ++MACUPD, ++SRCUPD) erhalten spezielle Behandlung
- Language-ID wird als Teil des Statement-Namens oder als Statement-Parameter geparsed
- AST Node enthält Referenz zur Language-ID

---

## Phasen-Reihenfolge & Abhängigkeiten

```
Phase 1 (Parser & AST)
    ↓
Phase 2 (Semantic Tokens) ← Testing Feedback Loop
    ↓
Phase 3 (Completion) ← Testing Feedback Loop
    ↓
Phase 4 (Diagnostics) ← Testing Feedback Loop
    ↓
Phase 5 (Hover) ← Testing Feedback Loop
    ↓
Phase 7 (Handler Integration)
    ↓
Phase 8 (Data Elements)
    ↓
Phase 6 (Code Cleanup) ← Nur NACH allen Features!
```

---

## Erfolgskriterien

### Funktional:
- ✅ Semantic Tokens: Statement-Parameter orange (nicht blau)
- ✅ Completion: Context-aware wie vorher
- ✅ Diagnostics: Required operands, keine false positives
- ✅ Hover: Statement/Operand Info korrekt
- ✅ Data Elements: Language-ID korrekt erkannt

### Code-Qualität:
- ✅ Kein toter Code
- ✅ Alle Tests grün
- ✅ Parser basiert auf Syntax-Diagrammen
- ✅ AST ist wiederverwendbar für zukünftige Features (Go to Definition, etc.)

---

## Geschätzte Komplexität

- **Phase 1:** 🔴 Hoch (Neue Architektur)
- **Phase 2:** 🟡 Mittel (Semantic Tokens Refactor)
- **Phase 3:** 🟡 Mittel (Completion Refactor)
- **Phase 4:** 🟢 Niedrig (Diagnostics Refactor)
- **Phase 5:** 🟢 Niedrig (Hover Refactor)
- **Phase 6:** 🟢 Niedrig (Cleanup)
- **Phase 7:** 🟡 Mittel (Integration)
- **Phase 8:** 🟡 Mittel (Data Elements)

**Gesamt:** ~3-5 Sessions (abhängig von Testing-Feedback)

---

## Wichtige Designentscheidungen

### 1. Warum AST statt direktes Token-Building?
- **Wiederverwendbarkeit:** AST kann für Go to Definition, Find References, Document Symbols verwendet werden
- **Testbarkeit:** Parser kann isoliert getestet werden
- **Wartbarkeit:** Änderungen an Syntax-Diagrammen nur im Parser, nicht in allen Features

### 2. Warum keine Groß-/Kleinschreibungs-Abhängigkeit?
- **Strukturelles Parsing:** Position und Kontext bestimmen Token-Typ, nicht Character-Klasse
- **Robustheit:** Funktioniert auch mit lowercase/mixed-case Input
- **Syntax-Diagramm-konform:** Diagramme spezifizieren Struktur, nicht Case

### 3. Warum Document-Cache im Handler?
- **Performance:** Parsing ist teuer, cache vermeidet Re-Parse
- **Konsistenz:** Alle Features arbeiten auf dem gleichen AST
- **Incremental Updates:** Später kann Delta-Parsing implementiert werden

---

## Risiken & Mitigation

### Risiko 1: Parser zu komplex
**Mitigation:**
- Start mit einfachen Statements (++USERMOD, ++VER)
- Iterativ erweitern
- Extensive Tests für jedes Statement

### Risiko 2: Performance-Probleme bei großen Files
**Mitigation:**
- Lazy Parsing (nur sichtbarer Bereich)
- Incremental Updates
- Profiling vor Optimierung

### Risiko 3: Breaking Changes während Migration
**Mitigation:**
- Feature Flags für neuen Parser
- Parallel-Betrieb alter/neuer Code
- Schrittweise Migration pro Feature

---

## Nächste Schritte

1. **Dieser Plan** als `REFACTORING_PLAN.md` committen
2. **Phase 1** in separater Branch starten
3. **Regelmäßiges Testing** mit realen SMP/E Files
4. **Feedback-Loop** nach jeder Phase
