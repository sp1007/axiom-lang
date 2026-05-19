package sema

import (
	"testing"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestAsyncFnType(t *testing.T) {
	pool := ast.NewInternPool(1024)
	tree := ast.NewTree([]byte(""), nil)
	tt := types.NewTypeTable()
	st := NewSymbolTable(pool)

	// Mock async func
	funcTypeID := tt.RegisterFunction([]types.TypeID{}, types.TypeI32, nil)
	funcInfo := tt.FuncInfo(funcTypeID)
	funcInfo.IsAsync = true // manually set

	// Create call expr
	callNode := tree.AddNode(ast.NodeCallExpr, 0)
	identNode := tree.AddNode(ast.NodeIdent, 0)
	tree.AppendChild(callNode, identNode)

	// Register in symtable
	nameID := pool.Intern([]byte("foo"))
	symIdx := uint32(len(st.Symbols))
	st.Symbols = append(st.Symbols, Symbol{
		NameID:   nameID,
		Kind:     SymFunc,
		Flags:    SymFlagPub,
		TypeID:   uint32(funcTypeID),
		DeclNode: 0,
		ScopeID:  0,
	})
	tree.SetPayload(identNode, symIdx)

	ie := NewInferenceEngine(tree, st, tt, nil)
	ie.Infer()

	resType := ie.TypeOf(callNode)
	isFuture, inner := IsFutureType(tt, resType)
	if !isFuture {
		t.Errorf("expected Future type, got %d", resType)
	}
	if inner != types.TypeI32 {
		t.Errorf("expected inner type i32, got %d", inner)
	}
}

func TestAwaitType(t *testing.T) {
	pool := ast.NewInternPool(1024)
	tree := ast.NewTree([]byte(""), nil)
	tt := types.NewTypeTable()
	st := NewSymbolTable(pool)

	futureType := CreateFutureType(tt, types.TypeI32)

	awaitNode := tree.AddNode(ast.NodeAwaitExpr, 0)
	exprNode := tree.AddNode(ast.NodeIdent, 0)
	tree.AppendChild(awaitNode, exprNode)

	nameID := pool.Intern([]byte("f"))
	symIdx := uint32(len(st.Symbols))
	st.Symbols = append(st.Symbols, Symbol{
		NameID:   nameID,
		Kind:     SymVar,
		Flags:    0,
		TypeID:   uint32(futureType),
		DeclNode: 0,
		ScopeID:  0,
	})
	tree.SetPayload(exprNode, symIdx)

	ie := NewInferenceEngine(tree, st, tt, nil)
	ie.Infer()

	resType := ie.TypeOf(awaitNode)
	if resType != types.TypeI32 {
		t.Errorf("expected i32, got %d", resType)
	}
}

func TestSpawnType(t *testing.T) {
	pool := ast.NewInternPool(1024)
	tree := ast.NewTree([]byte(""), nil)
	tt := types.NewTypeTable()
	st := NewSymbolTable(pool)

	spawnNode := tree.AddNode(ast.NodeSpawnExpr, 0)
	callNode := tree.AddNode(ast.NodeCallExpr, 0)
	tree.AppendChild(spawnNode, callNode)

	ie := NewInferenceEngine(tree, st, tt, nil)
	ie.Infer()

	resType := ie.TypeOf(spawnNode)
	if resType != types.TypeActorRef {
		t.Errorf("expected ActorRef, got %d", resType)
	}
}

func TestAwaitOutsideAsync(t *testing.T) {
	pool := ast.NewInternPool(1024)
	tree := ast.NewTree([]byte(""), nil)
	tt := types.NewTypeTable()
	st := NewSymbolTable(pool)

	futureType := CreateFutureType(tt, types.TypeI32)

	// fn main()
	funcNode := tree.AddNode(ast.NodeFuncDecl, 0)
	// it's not async!
	
	awaitNode := tree.AddNode(ast.NodeAwaitExpr, 0)
	exprNode := tree.AddNode(ast.NodeIdent, 0)
	tree.AppendChild(awaitNode, exprNode)
	tree.AppendChild(funcNode, awaitNode)

	nameID := pool.Intern([]byte("f"))
	symIdx := uint32(len(st.Symbols))
	st.Symbols = append(st.Symbols, Symbol{
		NameID:   nameID,
		Kind:     SymVar,
		Flags:    0,
		TypeID:   uint32(futureType),
		DeclNode: 0,
		ScopeID:  0,
	})
	tree.SetPayload(exprNode, symIdx)
	
	// Add root node manually
	tree.AppendChild(0, funcNode)

	ie := NewInferenceEngine(tree, st, tt, nil)
	ie.Infer()

	tc := NewTypeChecker(tree, pool, st, tt, ie)
	tc.Check()
	
	diags := tc.Errors()
	
	found := false
	for _, d := range diags {
		if d.Code == 3011 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error 3011 (await outside async)")
	}
}
