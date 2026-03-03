package data

import (
	"encoding/json"
	"fmt"
	"os"
)

// MCSStatement represents a MCS statement definition from smpe.json
type MCSStatement struct {
	Name             string    `json:"name"`
	LanguageVariants bool      `json:"language_variants,omitempty"`
	Description      string    `json:"description"`
	Parameter        string    `json:"parameter,omitempty"`
	Length           int       `json:"length,omitempty"`
	Type             string    `json:"type"`
	InlineData       bool      `json:"inline_data,omitempty"`
	Operands         []Operand `json:"operands,omitempty"`
}

// Operand represents an operand definition
type Operand struct {
	Name              string         `json:"name"`
	Parameter         string         `json:"parameter,omitempty"`
	Type              string         `json:"type,omitempty"`
	Length            int            `json:"length,omitempty"`
	Required          bool           `json:"required,omitempty"`
	RequiredGroup     bool           `json:"required_group,omitempty"`
	RequiredGroupID   string         `json:"required_group_id,omitempty"`
	Description       string         `json:"description,omitempty"`
	Values            []AllowedValue `json:"values,omitempty"`
	MutuallyExclusive string         `json:"mutually_exclusive,omitempty"`
	AllowedIf         string         `json:"allowed_if,omitempty"`
}

// AllowedValue represents an allowed value for an operand
// For sub-operands (e.g., DSN within FROMDS), this structure also includes type and length constraints
type AllowedValue struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameter   string `json:"parameter,omitempty"`   // Parameter syntax (e.g., "24|31|64" for AMODE)
	Type        string `json:"type,omitempty"`        // Type constraint (string, integer, etc.) for sub-operands
	Length      int    `json:"length,omitempty"`      // Maximum length constraint for sub-operands
}

// Store holds the shared MCS statement data
type Store struct {
	Statements map[string]MCSStatement
	List       []MCSStatement
}

// --- private types for new-format parsing ---

// smpeFileNew is the top-level structure of the new smpe.json format:
// { "templates": { ... }, "statements": [ ... ] }
type smpeFileNew struct {
	Templates  map[string][]Operand `json:"templates"`
	Statements []mcsStatementRaw    `json:"statements"`
}

// mcsStatementRaw mirrors MCSStatement but keeps Operands as raw JSON
// so $ref entries can be detected before full decoding.
type mcsStatementRaw struct {
	Name             string            `json:"name"`
	LanguageVariants bool              `json:"language_variants,omitempty"`
	Description      string            `json:"description"`
	Parameter        string            `json:"parameter,omitempty"`
	Length           int               `json:"length,omitempty"`
	Type             string            `json:"type"`
	InlineData       bool              `json:"inline_data,omitempty"`
	OperandsRaw      []json.RawMessage `json:"operands,omitempty"`
}

// refEntry is used to detect {"$ref": "template_name"} entries.
type refEntry struct {
	Ref string `json:"$ref"`
}

// Load reads and parses the smpe.json file, resolving any $ref template
// references at load time. Supports both the new object format and the
// legacy plain-array format for backwards compatibility.
func Load(dataPath string) (*Store, error) {
	fileBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	// Detect format by first non-whitespace character:
	// '{' => new format  |  '[' => legacy format
	for _, b := range fileBytes {
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			continue
		}
		if b == '{' {
			return loadNewFormat(fileBytes)
		}
		break
	}
	return loadLegacyFormat(fileBytes)
}

func loadNewFormat(fileBytes []byte) (*Store, error) {
	var wrapper smpeFileNew
	if err := json.Unmarshal(fileBytes, &wrapper); err != nil {
		return nil, fmt.Errorf("parsing smpe.json (new format): %w", err)
	}

	statements := make([]MCSStatement, 0, len(wrapper.Statements))
	for _, raw := range wrapper.Statements {
		stmt := MCSStatement{
			Name:             raw.Name,
			LanguageVariants: raw.LanguageVariants,
			Description:      raw.Description,
			Parameter:        raw.Parameter,
			Length:           raw.Length,
			Type:             raw.Type,
			InlineData:       raw.InlineData,
		}

		resolved, err := resolveOperands(raw.OperandsRaw, wrapper.Templates)
		if err != nil {
			return nil, fmt.Errorf("statement %q: %w", raw.Name, err)
		}
		stmt.Operands = resolved
		statements = append(statements, stmt)
	}

	return buildStore(statements), nil
}

func loadLegacyFormat(fileBytes []byte) (*Store, error) {
	var statements []MCSStatement
	if err := json.Unmarshal(fileBytes, &statements); err != nil {
		return nil, fmt.Errorf("parsing smpe.json: %w", err)
	}
	return buildStore(statements), nil
}

// resolveOperands processes a slice of raw JSON operand entries.
// If the slice has exactly one entry and it is a {"$ref": "name"}, the
// operands from the named template are returned. Otherwise each entry
// is decoded as a full Operand.
func resolveOperands(raw []json.RawMessage, templates map[string][]Operand) ([]Operand, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	// Single-entry $ref pattern: [{"$ref": "template_name"}]
	if len(raw) == 1 {
		var ref refEntry
		if err := json.Unmarshal(raw[0], &ref); err == nil && ref.Ref != "" {
			tpl, ok := templates[ref.Ref]
			if !ok {
				return nil, fmt.Errorf("unknown $ref %q (not defined in templates)", ref.Ref)
			}
			result := make([]Operand, len(tpl))
			copy(result, tpl)
			return result, nil
		}
	}

	// Normal case: decode each entry as a full Operand
	operands := make([]Operand, 0, len(raw))
	for _, r := range raw {
		var op Operand
		if err := json.Unmarshal(r, &op); err != nil {
			return nil, fmt.Errorf("parsing operand: %w", err)
		}
		operands = append(operands, op)
	}
	return operands, nil
}

func buildStore(statements []MCSStatement) *Store {
	stmtMap := make(map[string]MCSStatement, len(statements))
	for _, stmt := range statements {
		stmtMap[stmt.Name] = stmt
	}
	return &Store{
		Statements: stmtMap,
		List:       statements,
	}
}
