package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func setupEffectsTest() (*ast.AstTree, *ast.InternPool, *sema.SymbolTable, *types.TypeTable, *sema.EffectChecker) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	tree := ast.NewTree(nil, nil)
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeProgram})
	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	ec := sema.NewEffectChecker(tree, pool, st, tt, ie)
	return tree, pool, st, tt, ec
}

func TestEffectPropagation_Unhandled(t *testing.T) {
	tree, pool, st, _, ec := setupEffectsTest()
	
	// Impure function B that raises an error (TypeID 100)
	funcBID := pool.Intern([]byte("B"))
	symIdxB, _ := st.Define(funcBID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxB] = sema.EffectSet{
		Raises:  []types.TypeID{100},
		IsPure:  false,
		IsAsync: false,
	}
	
	// Function A that calls B but doesn't declare raises 100
	funcAID := pool.Intern([]byte("A"))
	symIdxA, _ := st.Define(funcAID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxA] = sema.EffectSet{
		Raises:  []types.TypeID{}, // empty
		IsPure:  false,
		IsAsync: false,
	}
	
	fnA := addNode(tree, 0, ast.NodeFuncDecl, symIdxA, 0)
	callB := addNode(tree, fnA, ast.NodeCallExpr, 0, 0)
	addNode(tree, callB, ast.NodeIdent, symIdxB, 0) // callee B
	
	diags := ec.Check()
	if len(diags) == 0 || diags[0].Code != 3041 {
		t.Errorf("expected unhandled effect error, got %v", diags)
	}
}

func TestEffectPropagation_HandledOrDeclared(t *testing.T) {
	tree, pool, st, _, ec := setupEffectsTest()
	
	// Impure function B that raises an error (TypeID 100)
	funcBID := pool.Intern([]byte("B"))
	symIdxB, _ := st.Define(funcBID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxB] = sema.EffectSet{
		Raises:  []types.TypeID{100},
		IsPure:  false,
		IsAsync: false,
	}
	
	// Function A that calls B and DECLARES it raises 100
	funcAID := pool.Intern([]byte("A"))
	symIdxA, _ := st.Define(funcAID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxA] = sema.EffectSet{
		Raises:  []types.TypeID{100}, // declared!
		IsPure:  false,
		IsAsync: false,
	}
	
	fnA := addNode(tree, 0, ast.NodeFuncDecl, symIdxA, 0)
	callB := addNode(tree, fnA, ast.NodeCallExpr, 0, 0)
	addNode(tree, callB, ast.NodeIdent, symIdxB, 0) // callee B
	
	diags := ec.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected error: %v", diags)
	}
}

func TestEffectPure_ImpureCallError(t *testing.T) {
	tree, pool, st, _, ec := setupEffectsTest()
	
	// Impure function B
	funcBID := pool.Intern([]byte("B"))
	symIdxB, _ := st.Define(funcBID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxB] = sema.EffectSet{
		IsPure:  false,
	}
	
	// Pure function A
	funcAID := pool.Intern([]byte("A"))
	symIdxA, _ := st.Define(funcAID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxA] = sema.EffectSet{
		IsPure:  true,
	}
	
	fnA := addNode(tree, 0, ast.NodeFuncDecl, symIdxA, 0)
	callB := addNode(tree, fnA, ast.NodeCallExpr, 0, 0)
	addNode(tree, callB, ast.NodeIdent, symIdxB, 0) // callee B
	
	diags := ec.Check()
	if len(diags) == 0 || diags[0].Code != 3040 {
		t.Errorf("expected pure function error, got %v", diags)
	}
}

func TestEffectPure_PureCall(t *testing.T) {
	tree, pool, st, _, ec := setupEffectsTest()
	
	// Pure function B
	funcBID := pool.Intern([]byte("B"))
	symIdxB, _ := st.Define(funcBID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxB] = sema.EffectSet{
		IsPure:  true,
	}
	
	// Pure function A
	funcAID := pool.Intern([]byte("A"))
	symIdxA, _ := st.Define(funcAID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxA] = sema.EffectSet{
		IsPure:  true,
	}
	
	fnA := addNode(tree, 0, ast.NodeFuncDecl, symIdxA, 0)
	callB := addNode(tree, fnA, ast.NodeCallExpr, 0, 0)
	addNode(tree, callB, ast.NodeIdent, symIdxB, 0) // callee B
	
	diags := ec.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected error: %v", diags)
	}
}

func TestEffectAsync_AwaitInNonAsyncError(t *testing.T) {
	tree, pool, st, _, ec := setupEffectsTest()
	
	// Non-async function A
	funcAID := pool.Intern([]byte("A"))
	symIdxA, _ := st.Define(funcAID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxA] = sema.EffectSet{
		IsAsync: false,
	}
	
	fnA := addNode(tree, 0, ast.NodeFuncDecl, symIdxA, 0)
	addNode(tree, fnA, ast.NodeAwaitExpr, 0, 0)
	
	diags := ec.Check()
	if len(diags) == 0 || diags[0].Code != 3042 {
		t.Errorf("expected await in non-async error, got %v", diags)
	}
}

func TestEffectAsync_AwaitInAsync(t *testing.T) {
	tree, pool, st, _, ec := setupEffectsTest()
	
	// Async function A
	funcAID := pool.Intern([]byte("A"))
	symIdxA, _ := st.Define(funcAID, sema.SymFunc, 0, 0)
	ec.FuncEffects[symIdxA] = sema.EffectSet{
		IsAsync: true,
	}
	
	fnA := addNode(tree, 0, ast.NodeFuncDecl, symIdxA, 0)
	addNode(tree, fnA, ast.NodeAwaitExpr, 0, 0)
	
	diags := ec.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected error: %v", diags)
	}
}
