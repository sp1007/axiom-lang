package sema

import (
	"encoding/json"
	"testing"
)

func TestCGAddNodes(t *testing.T) {
	cg := NewConnectionGraph()

	// Add value nodes
	n0 := cg.AddValueNode(1, 10, 0) // global scope
	n1 := cg.AddValueNode(2, 11, 1) // function scope
	n2 := cg.AddValueNode(3, 12, 2) // inner block scope

	if cg.NodeCount() != 3 {
		t.Fatalf("NodeCount() = %d, want 3", cg.NodeCount())
	}

	if n0 != 0 || n1 != 1 || n2 != 2 {
		t.Errorf("node IDs: got %d,%d,%d, want 0,1,2", n0, n1, n2)
	}

	// Lookup by symbol ID
	nodeID, ok := cg.NodeOfSym(2)
	if !ok || nodeID != 1 {
		t.Errorf("NodeOfSym(2) = %d, %v; want 1, true", nodeID, ok)
	}

	_, ok = cg.NodeOfSym(99)
	if ok {
		t.Error("NodeOfSym(99) should return false")
	}
}

func TestCGAddRefNode(t *testing.T) {
	cg := NewConnectionGraph()

	val := cg.AddValueNode(1, 10, 0)
	ref := cg.AddRefNode(val)

	if cg.NodeCount() != 2 {
		t.Fatalf("NodeCount() = %d, want 2", cg.NodeCount())
	}

	node := cg.Nodes[ref]
	if !node.IsRef {
		t.Error("ref node should have IsRef=true")
	}

	// Should have an automatic Borrows edge
	borrows := cg.OutEdges(ref, EdgeBorrows)
	if len(borrows) != 1 || borrows[0] != val {
		t.Errorf("ref node should borrow val; got %v", borrows)
	}
}

func TestCGAddEdges(t *testing.T) {
	cg := NewConnectionGraph()

	n0 := cg.AddValueNode(1, 10, 0)
	n1 := cg.AddValueNode(2, 11, 1)
	n2 := cg.AddValueNode(3, 12, 1)

	cg.AddEdge(n0, n1, EdgeOwns)
	cg.AddEdge(n0, n2, EdgeOwns)
	cg.AddEdge(n1, n2, EdgeFlowsTo)

	if cg.EdgeCount() != 3 {
		t.Fatalf("EdgeCount() = %d, want 3", cg.EdgeCount())
	}

	// Out edges of n0 with kind Owns
	owns := cg.OutEdges(n0, EdgeOwns)
	if len(owns) != 2 {
		t.Errorf("OutEdges(n0, Owns) = %v; want 2 targets", owns)
	}

	// In edges of n2 with kind Owns
	ownedBy := cg.InEdges(n2, EdgeOwns)
	if len(ownedBy) != 1 || ownedBy[0] != n0 {
		t.Errorf("InEdges(n2, Owns) = %v; want [n0]", ownedBy)
	}

	// FlowsTo
	flows := cg.OutEdges(n1, EdgeFlowsTo)
	if len(flows) != 1 || flows[0] != n2 {
		t.Errorf("OutEdges(n1, FlowsTo) = %v; want [n2]", flows)
	}
}

func TestCGEscapeDetection(t *testing.T) {
	cg := NewConnectionGraph()

	local := cg.AddValueNode(1, 10, 1)
	global := cg.AddValueNode(0, 10, 0)
	escapingLocal := cg.AddValueNode(2, 10, 1)

	// escapingLocal escapes to global
	cg.AddEdge(escapingLocal, global, EdgeEscapesTo)

	if cg.Escapes(local) {
		t.Error("local should NOT escape (no EscapesTo edges)")
	}

	if !cg.Escapes(escapingLocal) {
		t.Error("escapingLocal SHOULD escape (has EscapesTo edge)")
	}
}

func TestCGTransitiveEscape(t *testing.T) {
	cg := NewConnectionGraph()

	a := cg.AddValueNode(1, 10, 1)
	b := cg.AddValueNode(2, 10, 1)
	global := cg.AddValueNode(0, 10, 0)

	// a Owns b, b EscapesTo global
	cg.AddEdge(a, b, EdgeOwns)
	cg.AddEdge(b, global, EdgeEscapesTo)

	if !cg.Escapes(a) {
		t.Error("a should escape transitively (a Owns b, b EscapesTo global)")
	}

	if !cg.Escapes(b) {
		t.Error("b should escape directly")
	}
}

func TestCGFlowsToEscape(t *testing.T) {
	cg := NewConnectionGraph()

	a := cg.AddValueNode(1, 10, 1)
	b := cg.AddValueNode(2, 10, 1)
	retval := cg.AddValueNode(0, 10, 0)

	// a FlowsTo b, b EscapesTo retval (return value)
	cg.AddEdge(a, b, EdgeFlowsTo)
	cg.AddEdge(b, retval, EdgeEscapesTo)

	if !cg.Escapes(a) {
		t.Error("a should escape transitively via FlowsTo chain")
	}
}

func TestCGCycleHandling(t *testing.T) {
	cg := NewConnectionGraph()

	a := cg.AddValueNode(1, 10, 1)
	b := cg.AddValueNode(2, 10, 1)

	// Cycle: a Owns b, b FlowsTo a
	cg.AddEdge(a, b, EdgeOwns)
	cg.AddEdge(b, a, EdgeFlowsTo)

	// Should not infinite loop — neither escapes
	if cg.Escapes(a) {
		t.Error("a should NOT escape (cycle but no EscapesTo)")
	}
	if cg.Escapes(b) {
		t.Error("b should NOT escape (cycle but no EscapesTo)")
	}
}

func TestCGDominatedBy(t *testing.T) {
	cg := NewConnectionGraph()

	n0 := cg.AddValueNode(1, 10, 0) // global
	n1 := cg.AddValueNode(2, 10, 2) // nested scope

	if !cg.DominatedBy(n0, 0) {
		t.Error("n0 (lifetime=0) should be dominated by scope 0")
	}
	if cg.DominatedBy(n0, 1) {
		t.Error("n0 (lifetime=0) should NOT be dominated by scope 1")
	}
	if !cg.DominatedBy(n1, 2) {
		t.Error("n1 (lifetime=2) should be dominated by scope 2")
	}
	if !cg.DominatedBy(n1, 1) {
		t.Error("n1 (lifetime=2) should be dominated by scope 1")
	}
}

func TestCGReset(t *testing.T) {
	cg := NewConnectionGraph()
	cg.AddValueNode(1, 10, 0)
	cg.AddValueNode(2, 11, 1)
	cg.AddEdge(0, 1, EdgeOwns)

	cg.Reset()

	if cg.NodeCount() != 0 {
		t.Errorf("after Reset(), NodeCount() = %d, want 0", cg.NodeCount())
	}
	if cg.EdgeCount() != 0 {
		t.Errorf("after Reset(), EdgeCount() = %d, want 0", cg.EdgeCount())
	}
	if _, ok := cg.NodeOfSym(1); ok {
		t.Error("after Reset(), NodeOfSym should not find sym 1")
	}
}

func TestCGSerialization(t *testing.T) {
	cg := NewConnectionGraph()

	n0 := cg.AddValueNode(1, 10, 0)
	n1 := cg.AddValueNode(2, 11, 1)
	n2 := cg.AddValueNode(3, 12, 2)
	cg.AddEdge(n0, n1, EdgeOwns)
	cg.AddEdge(n1, n2, EdgeFlowsTo)
	cg.AddEdge(n2, n0, EdgeEscapesTo)

	// Marshal to JSON
	data, err := json.Marshal(cg)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal into a new graph
	cg2 := NewConnectionGraph()
	if err := json.Unmarshal(data, cg2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify node count
	if cg2.NodeCount() != 3 {
		t.Errorf("deserialized NodeCount() = %d, want 3", cg2.NodeCount())
	}

	// Verify edge count
	if cg2.EdgeCount() != 3 {
		t.Errorf("deserialized EdgeCount() = %d, want 3", cg2.EdgeCount())
	}

	// Verify escape detection works on deserialized graph
	if !cg2.Escapes(n0) {
		t.Error("deserialized: n0 should escape (n0 Owns n1, n1 FlowsTo n2, n2 EscapesTo n0)")
	}

	// Verify sym lookup
	nodeID, ok := cg2.NodeOfSym(2)
	if !ok || nodeID != 1 {
		t.Errorf("deserialized: NodeOfSym(2) = %d, %v; want 1, true", nodeID, ok)
	}
}

func TestCGString(t *testing.T) {
	cg := NewConnectionGraph()
	cg.AddValueNode(1, 10, 0)
	cg.AddValueNode(2, 11, 1)
	cg.AddEdge(0, 1, EdgeOwns)

	s := cg.String()
	if len(s) == 0 {
		t.Error("String() returned empty string")
	}
	// Just check it doesn't panic and contains key info
	if !containsSubstring(s, "2 nodes") || !containsSubstring(s, "1 edges") {
		t.Errorf("String() doesn't contain expected summary: %s", s)
	}
}

func TestCGTenVariableFunction(t *testing.T) {
	// Acceptance criteria: correctly models a 10-variable function.
	cg := NewConnectionGraph()

	// Simulate 10 local variables
	var nodes [10]uint32
	for i := 0; i < 10; i++ {
		nodes[i] = cg.AddValueNode(uint32(i+1), 10, 1)
	}

	// Chain: 0 Owns 1, 1 Owns 2, ..., 8 Owns 9
	for i := 0; i < 9; i++ {
		cg.AddEdge(nodes[i], nodes[i+1], EdgeOwns)
	}

	// Only last node escapes
	cg.AddEdge(nodes[9], 0, EdgeEscapesTo) // escapes to "global" node 0

	// Wait, node 0 is one of our nodes. Let's add a separate global node.
	cg2 := NewConnectionGraph()
	global := cg2.AddValueNode(0, 10, 0)
	for i := 0; i < 10; i++ {
		nodes[i] = cg2.AddValueNode(uint32(i+1), 10, 1)
	}
	for i := 0; i < 9; i++ {
		cg2.AddEdge(nodes[i], nodes[i+1], EdgeOwns)
	}
	cg2.AddEdge(nodes[9], global, EdgeEscapesTo)

	if cg2.NodeCount() != 11 {
		t.Errorf("10-var function: NodeCount() = %d, want 11", cg2.NodeCount())
	}

	// All nodes should escape transitively
	for i := 0; i < 10; i++ {
		if !cg2.Escapes(nodes[i]) {
			t.Errorf("10-var: node %d should escape transitively", i)
		}
	}

	// Global itself doesn't escape (no outgoing EscapesTo)
	if cg2.Escapes(global) {
		t.Error("10-var: global node should not escape")
	}
}

func TestCGAllOutEdges(t *testing.T) {
	cg := NewConnectionGraph()
	n0 := cg.AddValueNode(1, 10, 0)
	n1 := cg.AddValueNode(2, 11, 1)
	n2 := cg.AddValueNode(3, 12, 1)

	cg.AddEdge(n0, n1, EdgeOwns)
	cg.AddEdge(n0, n2, EdgeBorrows)

	edges := cg.AllOutEdges(n0)
	if len(edges) != 2 {
		t.Errorf("AllOutEdges(n0) = %d edges, want 2", len(edges))
	}
}

func TestCGEdgeKindString(t *testing.T) {
	kinds := []EdgeKind{EdgeOwns, EdgeBorrows, EdgeFlowsTo, EdgeEscapesTo, EdgeReusedBy}
	expected := []string{"Owns", "Borrows", "FlowsTo", "EscapesTo", "ReusedBy"}

	for i, k := range kinds {
		if k.String() != expected[i] {
			t.Errorf("EdgeKind(%d).String() = %q, want %q", k, k.String(), expected[i])
		}
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
