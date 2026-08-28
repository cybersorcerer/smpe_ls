package langid

// NationalLanguageIdentifiers holds all valid 3-character language
// identifiers. It is populated from smpe.json ("language_identifiers") when
// the data store is loaded, so the definitions live in one place only.
var NationalLanguageIdentifiers []string

// SetNationalLanguageIdentifiers replaces the identifier list. Called by the
// data package after reading smpe.json.
func SetNationalLanguageIdentifiers(ids []string) {
	NationalLanguageIdentifiers = ids
}

// IsValidLanguageID checks if the given string is a valid national language identifier
func IsValidLanguageID(id string) bool {
	for _, valid := range NationalLanguageIdentifiers {
		if id == valid {
			return true
		}
	}
	return false
}

// LanguageVariantStatements holds the MCS statement base names that may carry
// a national language identifier suffix (++BOOK -> ++BOOKENU). It is populated
// from the "language_variants" flag in smpe.json when the store is loaded.
var LanguageVariantStatements []string

// SetLanguageVariantStatements replaces the base name list. Called by the data
// package after reading smpe.json.
func SetLanguageVariantStatements(names []string) {
	LanguageVariantStatements = names
}

// IsLanguageVariantStatement checks if the given statement base name requires
// a language identifier suffix
func IsLanguageVariantStatement(baseName string) bool {
	for _, variant := range LanguageVariantStatements {
		if baseName == variant {
			return true
		}
	}
	return false
}

// ExtractLanguageID extracts the language identifier from a statement name
// Returns the base name and language ID separately
// E.g., "++FONTENU" returns ("++FONT", "ENU", true)
// E.g., "++APAR" returns ("++APAR", "", false)
func ExtractLanguageID(statementName string) (baseName string, langID string, hasLangID bool) {
	// Check if statement name is long enough to have a language ID
	if len(statementName) < 8 {
		return statementName, "", false
	}

	// Check each language variant statement
	for _, baseStmt := range LanguageVariantStatements {
		if len(statementName) >= len(baseStmt)+3 {
			// Check if statement starts with base name
			if statementName[:len(baseStmt)] == baseStmt {
				// Extract potential language ID (last 3 characters)
				potentialLangID := statementName[len(baseStmt):]
				if len(potentialLangID) == 3 && IsValidLanguageID(potentialLangID) {
					return baseStmt, potentialLangID, true
				}
			}
		}
	}

	return statementName, "", false
}

// GenerateAllVariants generates all language variants for a given base statement name
// E.g., "++FONT" returns ["++FONTARA", "++FONTCHS", ..., "++FONTTRK"]
func GenerateAllVariants(baseName string) []string {
	if !IsLanguageVariantStatement(baseName) {
		return []string{baseName}
	}

	variants := make([]string, 0, len(NationalLanguageIdentifiers))
	for _, langID := range NationalLanguageIdentifiers {
		variants = append(variants, baseName+langID)
	}
	return variants
}
