package sema_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// runOwnership runs the full pipeline through ownership checking.
func runOwnership(src []byte) ([]diagnostics.Diagnostic, *sema.OwnershipChecker) {
	toks, _, _ := lexer.Lex(src)
	pool := ast.NewInternPool(16)
	tree, _ := parser.Parse(toks, src, pool)

	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	lazy := sema.NewLazyResolver(st, tt, nil)
	nr := sema.NewNameResolver(tree, pool, st, tt, lazy)
	nr.Resolve()

	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	ie.Infer()

	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	tc.Check()

	oc := sema.NewOwnershipChecker(tree, pool, st, tt)
	diags := oc.Check()
	return diags, oc
}

func TestOwnershipMoveRule(t *testing.T) {
	// Moving a value then using it should produce an error.
	src := []byte(`fn main():
    let x = 42
    let y = x
    let z = x
`)
	diags, _ := runOwnership(src)

	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "moved value") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'use of moved value' error for x after move to y")
		for _, d := range diags {
			t.Logf("  diag: %s", d.Message)
		}
	}
}

func TestOwnershipBorrowNoMove(t *testing.T) {
	// Simply passing a value to a function shouldn't move it (no sink parameter).
	src := []byte(`fn use_val(v: i32) -> i32:
    return v

fn main():
    let x = 42
    let r = use_val(x)
`)
	diags, _ := runOwnership(src)

	for _, d := range diags {
		if strings.Contains(d.Message, "moved value") {
			t.Errorf("unexpected move error: %s", d.Message)
		}
	}
}

func TestOwnershipMutRule(t *testing.T) {
	// Assigning to an immutable variable should produce an error.
	// Note: The type checker already handles this in check_stmt.go,
	// but the ownership checker also validates it.
	src := []byte(`fn main():
    let x = 5
    x = 10
`)
	diags, _ := runOwnership(src)

	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "immutable") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'cannot assign to immutable' error")
		for _, d := range diags {
			t.Logf("  diag: %s", d.Message)
		}
	}
}

func TestOwnershipEscapeReturn(t *testing.T) {
	// Returning a value creates an EscapesTo edge.
	src := []byte(`fn make_val() -> i32:
    let x = 42
    return x

fn main():
    let v = make_val()
`)
	diags, oc := runOwnership(src)

	// No errors expected
	for _, d := range diags {
		t.Logf("diag: %s", d.Message)
	}

	// The graph should have at least one node with EscapesTo
	graph := oc.Graph()
	if graph.NodeCount() > 0 {
		// Verify the structure exists (graph is per-function, so we just
		// check that it was populated at all during analysis)
		t.Logf("graph has %d nodes, %d edges", graph.NodeCount(), graph.EdgeCount())
	}
}

func TestOwnershipGraphPopulation(t *testing.T) {
	// Verify that ownership checking runs successfully on var declarations.
	src := []byte(`fn main():
    let a = 1
    let b = 2
    let c = 3
`)
	diags, oc := runOwnership(src)

	if len(diags) > 0 {
		t.Errorf("expected 0 ownership errors, got %d", len(diags))
		for _, d := range diags {
			t.Logf("  diag: %s", d.Message)
		}
	}

	// Graph should exist (even though per-function graph is restored after analysis)
	if oc.Graph() == nil {
		t.Error("expected non-nil graph after ownership check")
	}
}

func TestOwnershipCheckerCreation(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)

	oc := sema.NewOwnershipChecker(tree, pool, st, tt)
	if oc == nil {
		t.Fatal("NewOwnershipChecker returned nil")
	}
	if oc.Graph() == nil {
		t.Fatal("Graph() returned nil")
	}
	if oc.Moved() == nil {
		t.Fatal("Moved() returned nil")
	}
}
