package parser

import (
	"fmt"
	"testing"
)

// nodePath renders a human-readable path to a node for failure messages,
// e.g. "++MOD > LEPARM > AC > [param]".
func nodePath(path []*Node) string {
	s := ""
	for i, n := range path {
		if i > 0 {
			s += " > "
		}
		label := n.Name
		if label == "" {
			label = fmt.Sprintf("[%s %q]", nodeTypeName(n.Type), n.Value)
		}
		s += label
	}
	return s
}

func nodeTypeName(t NodeType) string {
	switch t {
	case NodeTypeDocument:
		return "document"
	case NodeTypeStatement:
		return "statement"
	case NodeTypeOperand:
		return "operand"
	case NodeTypeParameter:
		return "param"
	case NodeTypeComment:
		return "comment"
	default:
		return "unknown"
	}
}

// checkNoDuplicateChildren verifies that no node has two children with the
// same Type, Value, and Position. This is the AST-level invariant that the
// LEPARM parameter-node-duplication bug violated: parseOperandParameter
// built a wrapper node whose own child had an identical Value/Position to
// itself, and the caller sometimes kept both.
func checkNoDuplicateChildren(t *testing.T, node *Node, path []*Node) {
	t.Helper()
	if node == nil {
		return
	}
	path = append(path, node)

	type key struct {
		typ   NodeType
		value string
		pos   Position
	}
	seen := map[key][]*Node{}
	for _, child := range node.Children {
		k := key{typ: child.Type, value: child.Value, pos: child.Position}
		seen[k] = append(seen[k], child)
	}
	for k, dupes := range seen {
		if len(dupes) > 1 {
			t.Errorf("duplicate children under %s: %d nodes with type=%s value=%q pos=%+v",
				nodePath(path), len(dupes), nodeTypeName(k.typ), k.value, k.pos)
		}
	}

	for _, child := range node.Children {
		checkNoDuplicateChildren(t, child, path)
	}
}

// checkNoSelfReferentialValue verifies that a NodeTypeParameter node never
// has a single child that is itself a NodeTypeParameter with the identical
// Value and Position — the exact shape of the LEPARM bug (a wrapper node
// wrapping a copy of itself), even in the case where there's only one
// child so checkNoDuplicateChildren's sibling-comparison wouldn't catch it.
func checkNoSelfReferentialValue(t *testing.T, node *Node, path []*Node) {
	t.Helper()
	if node == nil {
		return
	}
	path = append(path, node)

	if node.Type == NodeTypeParameter && len(node.Children) == 1 {
		child := node.Children[0]
		if child.Type == NodeTypeParameter && child.Value == node.Value && child.Position == node.Position {
			t.Errorf("parameter node wraps an identical copy of itself at %s: value=%q pos=%+v",
				nodePath(path), node.Value, node.Position)
		}
	}

	for _, child := range node.Children {
		checkNoSelfReferentialValue(t, child, path)
	}
}

// checkPositionsWithinLineBounds verifies every node's Position.Character
// is non-negative and, combined with Position.Length, does not obviously
// overflow (a cheap sanity check; it does not have access to source text
// here so it can't check against actual line length).
func checkPositionsNonNegative(t *testing.T, node *Node, path []*Node) {
	t.Helper()
	if node == nil {
		return
	}
	path = append(path, node)

	if node.Position.Line < 0 || node.Position.Character < 0 || node.Position.Length < 0 {
		t.Errorf("negative position at %s: %+v", nodePath(path), node.Position)
	}

	for _, child := range node.Children {
		checkPositionsNonNegative(t, child, path)
	}
}

// checkParentLinkage verifies every child's Parent pointer actually points
// back to the node that holds it as a child (catches wiring bugs where a
// node is appended to the wrong parent's Children slice).
func checkParentLinkage(t *testing.T, node *Node, path []*Node) {
	t.Helper()
	if node == nil {
		return
	}
	path = append(path, node)

	for _, child := range node.Children {
		if child.Parent != nil && child.Parent != node {
			t.Errorf("child at %s has Parent pointing elsewhere (wrong parent linkage)", nodePath(append(path, child)))
		}
	}

	for _, child := range node.Children {
		checkParentLinkage(t, child, path)
	}
}

// checkDocumentInvariants runs all structural invariant checks against every
// top-level statement in a parsed document. Call this from any test that
// parses real SMP/E MCS text, in addition to whatever value-specific
// assertions that test already makes.
func checkDocumentInvariants(t *testing.T, doc *Document) {
	t.Helper()
	if doc == nil {
		t.Fatal("checkDocumentInvariants: doc is nil")
	}
	for _, stmt := range doc.Statements {
		checkNoDuplicateChildren(t, stmt, nil)
		checkNoSelfReferentialValue(t, stmt, nil)
		checkPositionsNonNegative(t, stmt, nil)
		checkParentLinkage(t, stmt, nil)
	}
}
