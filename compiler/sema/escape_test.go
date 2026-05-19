package sema

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
)

func TestEscapeNoEscapeSimple(t *testing.T) {
	// Local variable used only locally → stack allocated.
	cg := NewConnectionGraph()
	cg.AddValueNode(0, 0, 0) // return slot
	x := cg.AddValueNode(1, 10, 1)

	pool := ast.NewInternPool(16)
	st := NewSymbolTable(pool)
	tree := ast.NewTree(nil, []lexer.Token{{Kind: lexer.TokenEOF}})

	ea := NewEscapeAnalysis(tree, pool, st, nil)
	report := ea.AnalyzeFunction(0, cg)

	_ = x
	// x should NOT escape (no EscapesTo edges)
	if cg.Escapes(x) {
		t.Error("simple local variable should NOT escape")
	}
	_ = report
}

func TestEscapeReturn(t *testing.T) {
	// Returned value escapes to heap.
	cg := NewConnectionGraph()
	retSlot := cg.AddValueNode(0, 0, 0) // return slot
	x := cg.AddValueNode(1, 10, 1)

	// x EscapesTo return slot
	cg.AddEdge(x, retSlot, EdgeEscapesTo)

	if !cg.Escapes(x) {
		t.Error("returned value should escape")
	}
}

func TestEscapeHeapStore(t *testing.T) {
	// Value stored in a heap structure → escapes.
	cg := NewConnectionGraph()
	cg.AddValueNode(0, 0, 0)          // return slot
	x := cg.AddValueNode(1, 10, 1)    // local value
	heap := cg.AddValueNode(0, 10, 0) // heap node

	cg.AddEdge(x, heap, EdgeEscapesTo)

	if !cg.Escapes(x) {
		t.Error("heap-stored value should escape")
	}
}

func TestEscapeBorrowNoEscape(t *testing.T) {
	// Borrowing a value does not cause escape.
	cg := NewConnectionGraph()
	cg.AddValueNode(0, 0, 0) // return slot
	x := cg.AddValueNode(1, 10, 1)
	borrow := cg.AddValueNode(2, 10, 1)

	// borrow Borrows x — but Borrows is not followed by Escapes()
	cg.AddEdge(borrow, x, EdgeBorrows)

	if cg.Escapes(x) {
		t.Error("borrowed value should NOT escape (Borrows is not transitive for escape)")
	}
}

func TestEscapeTransitiveOwns(t *testing.T) {
	// If a owns b, and b escapes, then a escapes.
	cg := NewConnectionGraph()
	cg.AddValueNode(0, 0, 0) // return slot
	a := cg.AddValueNode(1, 10, 1)
	b := cg.AddValueNode(2, 10, 1)
	global := cg.AddValueNode(0, 10, 0)

	cg.AddEdge(a, b, EdgeOwns)
	cg.AddEdge(b, global, EdgeEscapesTo)

	if !cg.Escapes(a) {
		t.Error("a should escape transitively (a Owns b, b EscapesTo)")
	}
}

func TestEscapeReport(t *testing.T) {
	// Verify EscapeReport tracks heap vs stack allocations.
	cg := NewConnectionGraph()
	cg.AddValueNode(0, 0, 0)       // return slot
	cg.AddValueNode(1, 10, 1)      // x (local)
	escaping := cg.AddValueNode(2, 10, 1) // y (escapes)
	global := cg.AddValueNode(0, 10, 0)

	cg.AddEdge(escaping, global, EdgeEscapesTo)

	// Build a minimal AST with two VarDecls
	pool := ast.NewInternPool(16)
	st := NewSymbolTable(pool)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)

	// Add program node
	prog := tree.AddNode(ast.NodeProgram, 0)

	// Add function node
	funcNode := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, funcNode)

	// Add VarDecl for x (sym 1 → does not escape)
	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	tree.AppendChild(funcNode, varX)

	// Add VarDecl for y (sym 2 → escapes)
	nameY := pool.Intern([]byte("y"))
	symY, _ := st.Define(nameY, SymVar, 0, 0)
	varY := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varY, symY)
	tree.AppendChild(funcNode, varY)

	// Register syms in CG
	// CG nodes: 0=retslot, 1=x, 2=y(escaping), 3=global
	// symX and symY are the symbol table indices, not CG node IDs
	// We need to re-create the CG with correct symIDs
	cg2 := NewConnectionGraph()
	cg2.AddValueNode(0, 0, 0)      // return slot
	cg2.AddValueNode(symX, 10, 1)  // x
	cg2.AddValueNode(symY, 10, 1)  // y
	globalNode := cg2.AddValueNode(0, 10, 0)
	cg2.AddEdge(2, globalNode, EdgeEscapesTo) // y escapes

	ea := NewEscapeAnalysis(tree, pool, st, nil)
	report := ea.AnalyzeFunction(funcNode, cg2)

	_ = escaping

	if len(report.AllocatesOnHeap) != 1 {
		t.Errorf("expected 1 heap allocation, got %d", len(report.AllocatesOnHeap))
	}
	if len(report.AllocatesOnStack) != 1 {
		t.Errorf("expected 1 stack allocation, got %d", len(report.AllocatesOnStack))
	}

	// Verify flag is set on escaping VarDecl
	yNode := tree.Node(varY)
	if yNode.Flags&ast.FlagEscapesToHeap == 0 {
		t.Error("escaping VarDecl should have FlagEscapesToHeap set")
	}

	xNode := tree.Node(varX)
	if xNode.Flags&ast.FlagEscapesToHeap != 0 {
		t.Error("non-escaping VarDecl should NOT have FlagEscapesToHeap set")
	}
}

func TestEscapeSizeThreshold(t *testing.T) {
	// Value above size threshold always goes to heap.
	pool := ast.NewInternPool(16)
	st := NewSymbolTable(pool)
	// No types table needed for this test since we don't set TypeID
	ea := NewEscapeAnalysis(nil, pool, st, nil)
	ea.SetSizeThreshold(512) // 512 byte threshold
	if ea.sizeThreshold != 512 {
		t.Errorf("SetSizeThreshold(512) didn't take effect: got %d", ea.sizeThreshold)
	}
}
