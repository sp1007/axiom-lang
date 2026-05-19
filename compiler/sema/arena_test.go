package sema

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
)

func TestArenaBasic(t *testing.T) {
	// in [arena]: { let x = Foo{} } → x marked with FlagUsesArena
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	// Node 0 is the root Program (created by NewTree)

	// Create arena ident
	arenaName := pool.Intern([]byte("arena"))
	arenaSym, _ := st.Define(arenaName, SymVar, 0, 0)

	// Create ArenaBlock as child of root
	arenaBlock := tree.AddNode(ast.NodeArenaBlock, 0)
	tree.AppendChild(0, arenaBlock) // append to root

	// First child: arena ident
	arenaIdent := tree.AddNode(ast.NodeIdent, 0)
	tree.SetPayload(arenaIdent, arenaSym)
	tree.AppendChild(arenaBlock, arenaIdent)

	// VarDecl inside arena block
	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	tree.AppendChild(arenaBlock, varX)

	ap := NewArenaPass(tree, pool, st)
	diags := ap.Process()

	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}

	// Verify x has FlagUsesArena
	xNode := tree.Node(varX)
	if xNode.Flags&ast.FlagUsesArena == 0 {
		t.Error("VarDecl inside ArenaBlock should have FlagUsesArena set")
	}
}

func TestArenaDestroy(t *testing.T) {
	// At ArenaBlock exit, DestroyStmt for the arena is injected.
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	arenaName := pool.Intern([]byte("arena"))
	arenaSym, _ := st.Define(arenaName, SymVar, 0, 0)

	arenaBlock := tree.AddNode(ast.NodeArenaBlock, 0)
	tree.AppendChild(0, arenaBlock)

	arenaIdent := tree.AddNode(ast.NodeIdent, 0)
	tree.SetPayload(arenaIdent, arenaSym)
	tree.AppendChild(arenaBlock, arenaIdent)

	ap := NewArenaPass(tree, pool, st)
	ap.Process()

	// Walk children of arenaBlock to find DestroyStmt
	foundDestroy := false
	child := tree.Node(arenaBlock).FirstChild
	for child != 0 {
		n := tree.Node(child)
		if n.Kind == ast.NodeDestroyStmt && n.Payload == arenaSym {
			foundDestroy = true
		}
		child = n.NextSibling
	}

	if !foundDestroy {
		t.Error("expected DestroyStmt for arena at ArenaBlock exit")
	}
}

func TestArenaCTGCSkip(t *testing.T) {
	// CTGC should skip values with FlagUsesArena.
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	fn := tree.AddNode(ast.NodeFuncDecl, 0)
	tree.AppendChild(0, fn)

	block := tree.AddNode(ast.NodeBlock, 0)
	tree.AppendChild(fn, block)

	// Add VarDecl with both EscapesToHeap and UsesArena
	nameX := pool.Intern([]byte("x"))
	symX, _ := st.Define(nameX, SymVar, 0, 0)
	st.SymbolAt(symX).TypeID = 50

	varX := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varX, symX)
	tree.SetFlags(varX, ast.FlagEscapesToHeap|ast.FlagUsesArena)
	tree.AppendChild(block, varX)

	moved := make(map[uint32]bool)
	ctgc := NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(fn)

	count := ctgc.DestroyCount(block)
	if count != 0 {
		t.Errorf("CTGC should skip UsesArena values, got %d DestroyStmt", count)
	}
}

func TestArenaNestedBlocks(t *testing.T) {
	// Nested arena blocks: inner uses inner arena.
	pool := ast.NewInternPool(16)
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)
	st := NewSymbolTable(pool)

	// Outer arena block
	outerArenaName := pool.Intern([]byte("outer_arena"))
	outerArenaSym, _ := st.Define(outerArenaName, SymVar, 0, 0)

	outerBlock := tree.AddNode(ast.NodeArenaBlock, 0)
	tree.AppendChild(0, outerBlock) // child of root

	outerIdent := tree.AddNode(ast.NodeIdent, 0)
	tree.SetPayload(outerIdent, outerArenaSym)
	tree.AppendChild(outerBlock, outerIdent)

	// VarDecl in outer
	nameA := pool.Intern([]byte("a"))
	symA, _ := st.Define(nameA, SymVar, 0, 0)
	varA := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varA, symA)
	tree.AppendChild(outerBlock, varA)

	// Inner arena block
	innerArenaName := pool.Intern([]byte("inner_arena"))
	innerArenaSym, _ := st.Define(innerArenaName, SymVar, 0, 0)
	_ = innerArenaSym

	innerBlock := tree.AddNode(ast.NodeArenaBlock, 0)
	tree.AppendChild(outerBlock, innerBlock)

	innerIdent := tree.AddNode(ast.NodeIdent, 0)
	tree.SetPayload(innerIdent, innerArenaSym)
	tree.AppendChild(innerBlock, innerIdent)

	// VarDecl in inner
	nameB := pool.Intern([]byte("b"))
	symB, _ := st.Define(nameB, SymVar, 0, 0)
	varB := tree.AddNode(ast.NodeVarDecl, 0)
	tree.SetPayload(varB, symB)
	tree.AppendChild(innerBlock, varB)

	ap := NewArenaPass(tree, pool, st)
	ap.Process()

	// Both vars should have FlagUsesArena
	if tree.Node(varA).Flags&ast.FlagUsesArena == 0 {
		t.Error("outer var should have FlagUsesArena")
	}
	if tree.Node(varB).Flags&ast.FlagUsesArena == 0 {
		t.Error("inner var should have FlagUsesArena")
	}

	// Check ArenaVarDeclCount
	outerCount := ap.ArenaVarDeclCount(outerBlock)
	if outerCount != 2 { // a + b (nested)
		t.Errorf("expected 2 arena vars in outer block, got %d", outerCount)
	}
}
