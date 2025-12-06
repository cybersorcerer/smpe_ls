package data

import (
	"encoding/json"
	"os"
)

// MCSStatement represents a MCS statement definition from smpe.json
type MCSStatement struct {
	Name             string    `json:"name"`
	LanguageVariants bool      `json:"language_variants,omitempty"`
	Description      string    `json:"description"`
	Parameter        string    `json:"parameter,omitempty"`
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

// Load reads and parses the smpe.json file once
func Load(dataPath string) (*Store, error) {
	// Load MCS definitions from smpe.json
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var statements []MCSStatement
	if err := json.Unmarshal(data, &statements); err != nil {
		return nil, err
	}

	// Build lookup map
	stmtMap := make(map[string]MCSStatement, len(statements))
	for _, stmt := range statements {
		stmtMap[stmt.Name] = stmt
	}

	return &Store{
		Statements: stmtMap,
		List:       statements,
	}, nil
}
