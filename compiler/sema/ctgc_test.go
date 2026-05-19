package sema

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
)

func buildCTGCTree(pool *ast.InternPool) (*ast.AstTree, *SymbolTable) {
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)
	return tree, st
}

func TestDestroyAtBlockEnd(t *testing.T) {
	// let x = alloc_heap() → DestroyStmt for x at block end
	pool := ast.NewInternPool(16)
	tree, st := buildCTGCTree(pool)

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	// Add VarDecl for x (heap-allocated, non-primitive type)
	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	sym := st.SymbolAt(symX)
	sym.TypeID = 50 // non-primitive type ID

	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	tree.SetFlags(varX, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varX)

	moved := make(map[uint32]bool)
	ctgc := NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(fn)

	count := ctgc.DestroyCount(block)
	if count != 1 {
		t.Errorf("expected 1 DestroyStmt, got %d", count)
	}
}

func TestNoDestroyMoved(t *testing.T) {
	// let x = Foo{}; consume(x) — x moved → no destroy for x
	pool := ast.NewInternPool(16)
	tree, st := buildCTGCTree(pool)

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	sym := st.SymbolAt(symX)
	sym.TypeID = 50

	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	tree.SetFlags(varX, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varX)

	// Mark as moved
	moved := map[uint32]bool{symX: true}
	ctgc := NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(fn)

	count := ctgc.DestroyCount(block)
	if count != 0 {
		t.Errorf("expected 0 DestroyStmt for moved value, got %d", count)
	}
}

func TestDestroyOrder(t *testing.T) {
	// let x = A{}; let y = B{}; → destroy y before x (LIFO)
	pool := ast.NewInternPool(16)
	tree, st := buildCTGCTree(pool)

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 50

	nameY := pool.Intern([]byte("y"))
	symY, _ := st.Define(nameY, SymVar, 0, 0)
	st.SymbolAt(symY).TypeID = 51

	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	tree.SetFlags(varX, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varX)

	varY := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varY, symY)
	tree.SetFlags(varY, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varY)

	moved := make(map[uint32]bool)
	ctgc := NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(fn)

	count := ctgc.DestroyCount(block)
	if count != 2 {
		t.Fatalf("expected 2 DestroyStmt, got %d", count)
	}

	// Verify LIFO order: y destroyed before x
	// Walk block children to find DestroyStmt nodes
	var destroySyms []uint32
	child := tree.Node(block).FirstChild
	for child != 0 {
		n := tree.Node(child)
		if n.Kind == ast.NodeDestroyStmt {
			destroySyms = append(destroySyms, n.Payload)
		}
		child = n.NextSibling
	}

	if len(destroySyms) != 2 {
		t.Fatalf("expected 2 destroy syms, got %d", len(destroySyms))
	}
	// LIFO: y first (symY), then x (symX)
	if destroySyms[0] != symY || destroySyms[1] != symX {
		t.Errorf("LIFO order wrong: got [%d, %d], want [%d, %d]", destroySyms[0], destroySyms[1], symY, symX)
	}
}

func TestNoDestroyStack(t *testing.T) {
	// Stack-allocated value → no DestroyStmt
	pool := ast.NewInternPool(16)
	tree, st := buildCTGCTree(pool)

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 50

	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	// NO FlagEscapesToHeap set
	tree.AppendChild(block, varX)

	moved := make(map[uint32]bool)
	ctgc := NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(fn)

	count := ctgc.DestroyCount(block)
	if count != 0 {
		t.Errorf("expected 0 DestroyStmt for stack-allocated value, got %d", count)
	}
}

func TestNoDestroyPrimitive(t *testing.T) {
	// let x: i32 = 5 → no DestroyStmt (primitive type)
	pool := ast.NewInternPool(16)
	tree, st := buildCTGCTree(pool)

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 3 // i32 (primitive)

	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	tree.SetFlags(varX, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varX)

	moved := make(map[uint32]bool)
	ctgc := NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(fn)

	count := ctgc.DestroyCount(block)
	if count != 0 {
		t.Errorf("expected 0 DestroyStmt for primitive type, got %d", count)
	}
}

func TestDestroyMultipleBlocks(t *testing.T) {
	// Nested blocks: each block gets its own destroys.
	pool := ast.NewInternPool(16)
	tree, st := buildCTGCTree(pool)

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	outerBlock := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, outerBlock)

	// Outer var
	nameA := pool.Intern([]byte("a"))
	symA, _ := st.Define(nameA, SymVar, 0, 0)
	st.SymbolAt(symA).TypeID = 50

	varA := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varA, symA)
	tree.SetFlags(varA, ast.FlagEscapesToHeap)
	tree.AppendChild(outerBlock, varA)

	// Inner block with its own var
	innerBlock := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(outerBlock, innerBlock)

	nameB := pool.Intern([]byte("b"))
	symB, _ := st.Define(nameB, SymVar, 0, 0)
	st.SymbolAt(symB).TypeID = 51

	varB := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varB, symB)
	tree.SetFlags(varB, ast.FlagEscapesToHeap)
	tree.AppendChild(innerBlock, varB)

	moved := make(map[uint32]bool)
	ctgc := NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(fn)

	outerCount := ctgc.DestroyCount(outerBlock)
	innerCount := ctgc.DestroyCount(innerBlock)

	// Inner block has 1 destroy for b
	if innerCount != 1 {
		t.Errorf("inner block: expected 1 DestroyStmt, got %d", innerCount)
	}
	// Outer block has 1 destroy for a + the inner block's destroy (nested)
	// Actually DestroyCount counts recursively, so outer includes inner
	if outerCount != 2 {
		t.Errorf("outer block (recursive): expected 2 DestroyStmt total, got %d", outerCount)
	}
}
