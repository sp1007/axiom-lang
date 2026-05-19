package sema

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
)

func TestAliasReuseBasic(t *testing.T) {
	// DestroyStmt(x) followed by VarDecl(y) of same type → AliasStmt
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 50

	nameY := pool.Intern([]byte("y"))
	symY, _ := st.Define(nameY, SymVar, 0, 0)
	st.SymbolAt(symY).TypeID = 50 // same type

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	// Add DestroyStmt for x
	destroy := tree.AddNode(ast.NodeDestroyStmt, 0)
	tree.SetPayload(destroy, symX)
	tree.AppendChild(block, destroy)

	// Add VarDecl for y (heap-allocated)
	varY := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varY, symY)
	tree.SetFlags(varY, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varY)

	cg := NewConnectionGraph()
	cg.AddValueNode(symX, 50, 1)
	cg.AddValueNode(symY, 50, 1)

	ar := NewAliasReuse(tree, st, cg)
	count := ar.Optimize(fn)

	if count != 1 {
		t.Fatalf("expected 1 alias reuse, got %d", count)
	}

	// Verify the DestroyStmt was replaced with AliasStmt
	destroyNode := tree.Node(destroy)
	if destroyNode.Kind != ast.NodeAliasStmt {
		t.Errorf("expected AliasStmt, got %s", destroyNode.Kind)
	}
	if destroyNode.Payload != symX {
		t.Errorf("AliasStmt.Payload (from) = %d, want %d", destroyNode.Payload, symX)
	}
	if destroyNode.ExtraIdx != symY {
		t.Errorf("AliasStmt.ExtraIdx (to) = %d, want %d", destroyNode.ExtraIdx, symY)
	}

	// Verify ReusedBy edge was added
	reused := cg.OutEdges(0, EdgeReusedBy) // node 0 = symX
	if len(reused) != 1 {
		t.Errorf("expected 1 ReusedBy edge, got %d", len(reused))
	}
}

func TestAliasReuseTypeMismatch(t *testing.T) {
	// Destroy Foo, alloc Bar → NOT replaced (different types)
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 50 // Foo

	nameY := pool.Intern([]byte("y"))
	symY, _ := st.Define(nameY, SymVar, 0, 0)
	st.SymbolAt(symY).TypeID = 51 // Bar (different!)

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	destroy := tree.AddNode(ast.NodeDestroyStmt, 0)
	tree.SetPayload(destroy, symX)
	tree.AppendChild(block, destroy)

	varY := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varY, symY)
	tree.SetFlags(varY, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varY)

	ar := NewAliasReuse(tree, st, nil)
	count := ar.Optimize(fn)

	if count != 0 {
		t.Errorf("expected 0 alias reuse for type mismatch, got %d", count)
	}

	// Verify DestroyStmt was NOT replaced
	if tree.Node(destroy).Kind != ast.NodeDestroyStmt {
		t.Error("DestroyStmt should not be replaced for type mismatch")
	}
}

func TestAliasReuseWithBorrow(t *testing.T) {
	// Outstanding borrow of x when destroyed → NOT replaced (unsafe)
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 50

	nameY := pool.Intern([]byte("y"))
	symY, _ := st.Define(nameY, SymVar, 0, 0)
	st.SymbolAt(symY).TypeID = 50

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	destroy := tree.AddNode(ast.NodeDestroyStmt, 0)
	tree.SetPayload(destroy, symX)
	tree.AppendChild(block, destroy)

	varY := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varY, symY)
	tree.SetFlags(varY, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varY)

	// CG with active borrow on x
	cg := NewConnectionGraph()
	xNode := cg.AddValueNode(symX, 50, 1)
	cg.AddValueNode(symY, 50, 1)
	borrower := cg.AddValueNode(99, 50, 1) // external borrower
	cg.AddEdge(borrower, xNode, EdgeBorrows)

	ar := NewAliasReuse(tree, st, cg)
	count := ar.Optimize(fn)

	if count != 0 {
		t.Errorf("expected 0 alias reuse with active borrow, got %d", count)
	}
}

func TestAliasReuseNoCG(t *testing.T) {
	// Without a ConnectionGraph, alias reuse still works if types match.
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 50

	nameY := pool.Intern([]byte("y"))
	symY, _ := st.Define(nameY, SymVar, 0, 0)
	st.SymbolAt(symY).TypeID = 50

	prog := tree.AddNode(ast.NodeProgram, 0)
	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(prog, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	destroy := tree.AddNode(ast.NodeDestroyStmt, 0)
	tree.SetPayload(destroy, symX)
	tree.AppendChild(block, destroy)

	varY := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varY, symY)
	tree.SetFlags(varY, ast.FlagEscapesToHeap)
	tree.AppendChild(block, varY)

	ar := NewAliasReuse(tree, st, nil) // no CG
	count := ar.Optimize(fn)

	if count != 1 {
		t.Errorf("expected 1 alias reuse without CG, got %d", count)
	}
}
