package langid

import (
	"testing"
)

// The language lists are populated from smpe.json at load time, so a unit test
// of this package has to supply them itself. Importing the data package here
// would be an import cycle (data depends on langid).
func setupLanguageData() {
	SetNationalLanguageIdentifiers([]string{
		"ARA", "CHS", "CHT", "DAN", "DES", "DEU", "ELL", "ENG",
		"ENP", "ENU", "ESP", "FIN", "FRA", "FRB", "FRC", "FRS",
		"HEB", "ISL", "ITA", "ITS", "JPN", "KOR", "NLB", "NLD",
		"NOR", "PTB", "PTG", "RMS", "RUS", "SVE", "THA", "TRK",
	})
	SetLanguageVariantStatements([]string{
		"++BOOK", "++BSIND", "++CGM", "++DATA6", "++FONT", "++GDF",
		"++HFS", "++HELP", "++IMG", "++MSG", "++PNL", "++PROBJ",
		"++PRSRC", "++PSEG", "++PUBLB", "++SAMP", "++SKL", "++TBL",
		"++TEXT", "++UTIN", "++UTOUT",
	})
}

func TestHFSLanguageID(t *testing.T) {
	setupLanguageData()

	// Test that ++HFS is recognized as a language variant statement
	if !IsLanguageVariantStatement("++HFS") {
		t.Error("++HFS should be a language variant statement")
	}

	// Test that ++HFSENU is correctly parsed
	base, langID, hasLang := ExtractLanguageID("++HFSENU")
	if !hasLang {
		t.Error("++HFSENU should have a language ID")
	}
	if base != "++HFS" {
		t.Errorf("Expected base '++HFS', got '%s'", base)
	}
	if langID != "ENU" {
		t.Errorf("Expected langID 'ENU', got '%s'", langID)
	}

	// Test that ++HFSDEU is correctly parsed
	base, langID, hasLang = ExtractLanguageID("++HFSDEU")
	if !hasLang {
		t.Error("++HFSDEU should have a language ID")
	}
	if base != "++HFS" {
		t.Errorf("Expected base '++HFS', got '%s'", base)
	}
	if langID != "DEU" {
		t.Errorf("Expected langID 'DEU', got '%s'", langID)
	}

	// Test that ++HFS alone does not have a language ID
	_, _, hasLang = ExtractLanguageID("++HFS")
	if hasLang {
		t.Error("++HFS should not have a language ID")
	}

	// Test GenerateAllVariants for ++HFS
	variants := GenerateAllVariants("++HFS")
	expectedCount := len(NationalLanguageIdentifiers)
	if len(variants) != expectedCount {
		t.Errorf("Expected %d variants for ++HFS, got %d", expectedCount, len(variants))
	}

	// Check that first variant is ++HFSARA
	if len(variants) > 0 && variants[0] != "++HFSARA" {
		t.Errorf("Expected first variant '++HFSARA', got '%s'", variants[0])
	}
}
