package sema

import (
	"testing"
)

func TestIsolatedFreshAlloc(t *testing.T) {
	// A freshly allocated value with no external refs → proven isolated.
	cg := NewConnectionGraph()
	x := cg.AddValueNode(1, 10, 1)

	iv := NewIsolatedVerifier(cg)

	ok, violators := iv.VerifyIsolated(x)
	if !ok {
		t.Errorf("fresh allocation should be isolated, got violators: %v", violators)
	}

	if !iv.IsFreshlyAllocated(x) {
		t.Error("fresh allocation should report IsFreshlyAllocated=true")
	}
}

func TestIsolatedWithBorrow(t *testing.T) {
	// Value x with an external borrow y → NOT isolated.
	cg := NewConnectionGraph()
	x := cg.AddValueNode(1, 10, 1) // the value we want to isolate
	y := cg.AddValueNode(2, 10, 1) // external reference holder

	// y borrows x (external incoming edge to x from y)
	cg.AddEdge(y, x, EdgeBorrows)

	iv := NewIsolatedVerifier(cg)

	ok, violators := iv.VerifyIsolated(x)
	if ok {
		t.Error("value with external borrow should NOT be isolated")
	}
	if len(violators) != 1 || violators[0] != y {
		t.Errorf("expected violator [%d], got %v", y, violators)
	}
}

func TestIsolatedTransitive(t *testing.T) {
	// Struct containing another struct, no external refs → OK.
	cg := NewConnectionGraph()
	outer := cg.AddValueNode(1, 10, 1)
	inner := cg.AddValueNode(2, 11, 1)

	// outer Owns inner — inner is part of the subgraph
	cg.AddEdge(outer, inner, EdgeOwns)

	iv := NewIsolatedVerifier(cg)

	ok, violators := iv.VerifyIsolated(outer)
	if !ok {
		t.Errorf("transitively owned subgraph with no external refs should be isolated, got violators: %v", violators)
	}
}

func TestIsolatedTransitiveViolation(t *testing.T) {
	// Struct containing another struct, but inner has external borrow → NOT OK.
	cg := NewConnectionGraph()
	outer := cg.AddValueNode(1, 10, 1)
	inner := cg.AddValueNode(2, 11, 1)
	external := cg.AddValueNode(3, 10, 1)

	cg.AddEdge(outer, inner, EdgeOwns)
	cg.AddEdge(external, inner, EdgeBorrows) // external borrow into the subgraph

	iv := NewIsolatedVerifier(cg)

	ok, violators := iv.VerifyIsolated(outer)
	if ok {
		t.Error("transitively owned subgraph with external borrow should NOT be isolated")
	}
	if len(violators) != 1 || violators[0] != external {
		t.Errorf("expected violator [%d], got %v", external, violators)
	}
}

func TestIsolatedAfterMove(t *testing.T) {
	// After moving x, the original is invalidated — no external refs remain.
	// For wrapper to contain x, we need wrapper Owns x.
	cg := NewConnectionGraph()
	val := cg.AddValueNode(1, 10, 1)
	iso := cg.AddValueNode(2, 10, 1)
	cg.AddEdge(iso, val, EdgeOwns) // isolated wrapper owns the value

	iv := NewIsolatedVerifier(cg)
	ok, _ := iv.VerifyIsolated(iso)
	if !ok {
		t.Error("isolated wrapper owning a value with no external refs should be isolated")
	}
}

func TestIsolatedSelfContainedCycle(t *testing.T) {
	// Self-contained cycle within the subgraph → still isolated.
	cg := NewConnectionGraph()
	a := cg.AddValueNode(1, 10, 1)
	b := cg.AddValueNode(2, 10, 1)

	cg.AddEdge(a, b, EdgeOwns)
	cg.AddEdge(b, a, EdgeFlowsTo) // cycle

	iv := NewIsolatedVerifier(cg)

	ok, violators := iv.VerifyIsolated(a)
	if !ok {
		t.Errorf("self-contained cycle should be isolated, got violators: %v", violators)
	}
}

func TestIsolatedMultipleExternalRefs(t *testing.T) {
	// Multiple external references → all reported.
	cg := NewConnectionGraph()
	target := cg.AddValueNode(1, 10, 1)
	ext1 := cg.AddValueNode(2, 10, 1)
	ext2 := cg.AddValueNode(3, 10, 1)

	cg.AddEdge(ext1, target, EdgeBorrows)
	cg.AddEdge(ext2, target, EdgeBorrows)

	iv := NewIsolatedVerifier(cg)

	ok, violators := iv.VerifyIsolated(target)
	if ok {
		t.Error("value with multiple external borrows should NOT be isolated")
	}
	if len(violators) != 2 {
		t.Errorf("expected 2 violators, got %d: %v", len(violators), violators)
	}
}

func TestIsolatedFreshlyAllocatedWithBorrow(t *testing.T) {
	// IsFreshlyAllocated returns false if borrows exist.
	cg := NewConnectionGraph()
	x := cg.AddValueNode(1, 10, 1)
	y := cg.AddValueNode(2, 10, 1)

	cg.AddEdge(y, x, EdgeBorrows)

	iv := NewIsolatedVerifier(cg)

	if iv.IsFreshlyAllocated(x) {
		t.Error("value with incoming borrow should NOT be freshly allocated")
	}
}
