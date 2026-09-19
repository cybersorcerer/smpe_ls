package data

// Element source operands.
//
// A statement that carries element data gets that data either inline, right
// after its terminator, or from somewhere else - a data set, a relative file,
// a library. The operand naming that other place replaces the inline data, so
// every check that asks "does this statement expect inline data?" has to know
// which operands those are.
//
// The list lives in smpe.json under "element_source" so it is stated once.
// Adding a new one there is enough; no Go code needs to change. Missing it in
// one of the checks is what made ++MOD(X) LKLIB(DD1) report absent inline
// data before v1.3.13.
//
// DELETE is deliberately not part of this list. It names no source at all: it
// puts the statement into deletion mode, where no data is needed. The checks
// treat it next to the element sources but for a different reason.

// defaultElementSourceOperands is what the checks fall back to when smpe.json
// carries no "element_source" list, which is the case for a file written
// before the list was introduced. Without it every element source would go
// unrecognized and every such statement would be reported as missing its
// inline data, so the fallback fails towards the established behaviour
// rather than towards silence.
var defaultElementSourceOperands = []string{"FROMDS", "RELFILE", "TXLIB", "LKLIB"}

// elementSourceOperands is filled from smpe.json when the store is loaded.
var elementSourceOperands = defaultElementSourceOperands

// SetElementSourceOperands replaces the list. Called by Load after reading
// smpe.json; an empty list keeps the built-in default.
func SetElementSourceOperands(names []string) {
	if len(names) == 0 {
		elementSourceOperands = defaultElementSourceOperands
		return
	}
	elementSourceOperands = names
}

// ElementSourceOperands returns the operand names that supply element data
// from outside the SYSMOD.
func ElementSourceOperands() []string {
	return elementSourceOperands
}

// IsElementSource reports whether an operand supplies the element data from
// somewhere other than inline. The name is checked on its own, without asking
// whether the statement defines that operand: an operand that does not belong
// to the statement is reported separately as unknown, and a statement is not
// asked for inline data it was never going to carry.
func IsElementSource(operandName string) bool {
	for _, name := range elementSourceOperands {
		if name == operandName {
			return true
		}
	}
	return false
}
