package langid

// NationalLanguageIdentifiers contains all valid 3-character language identifiers
// as defined in the SMP/E Reference Manual Table 3
var NationalLanguageIdentifiers = []string{
	"ARA", // Arabic
	"CHS", // Simplified Chinese
	"CHT", // Traditional Chinese
	"DAN", // Danish
	"DES", // German (Switzerland)
	"DEU", // German (Germany)
	"ELL", // Greek
	"ENG", // English (United Kingdom)
	"ENP", // Uppercase English
	"ENU", // English (United States)
	"ESP", // Spanish
	"FIN", // Finnish
	"FRA", // French (France)
	"FRB", // French (Belgium)
	"FRC", // French (Canada)
	"FRS", // French (Switzerland)
	"HEB", // Hebrew
	"ISL", // Icelandic
	"ITA", // Italian (Italy)
	"ITS", // Italian (Switzerland)
	"JPN", // Japanese
	"KOR", // Korean
	"NLB", // Dutch (Belgium)
	"NLD", // Dutch (Netherlands)
	"NOR", // Norwegian
	"PTB", // Portuguese (Brazil)
	"PTG", // Portuguese (Portugal)
	"RMS", // Rhaeto-Romanic
	"RUS", // Russian
	"SVE", // Swedish
	"THA", // Thai
	"TRK", // Turkish
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

// LanguageVariantStatements contains all MCS statement base names that require
// a national language identifier suffix (e.g., ++BOOK becomes ++BOOKENU)
var LanguageVariantStatements = []string{
	"++BOOK",
	"++BSIND",
	"++CGM",
	"++DATA6",
	"++FONT",
	"++GDF",
	"++HELP",
	"++IMG",
	"++MSG",
	"++PNL",
	"++PROBJ",
	"++PRSRC",
	"++PSEG",
	"++PUBLB",
	"++SAMP",
	"++SKL",
	"++TBL",
	"++TEXT",
	"++UTIN",
	"++UTOUT",
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
