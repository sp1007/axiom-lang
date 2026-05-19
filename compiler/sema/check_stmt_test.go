package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func setupCheckTest() (*ast.AstTree, *ast.InternPool, *sema.SymbolTable, *types.TypeTable, *sema.InferenceEngine, *sema.TypeChecker) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	tree := ast.NewTree(nil, nil)
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeProgram})
	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	return tree, pool, st, tt, ie, tc
}

func TestCheck_LetWithAnnotation(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	xID := pool.Intern([]byte("x"))
	symIdx, _ := st.Define(xID, sema.SymVar, 0, 0)
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeI32)
	
	decl := addNode(tree, 0, ast.NodeVarDecl, symIdx, 0)
	addNode(tree, decl, ast.NodeIntLit, 42, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheck_LetTypeMismatch(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	xID := pool.Intern([]byte("x"))
	symIdx, _ := st.Define(xID, sema.SymVar, 0, 0)
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeBool) // mismatch
	
	decl := addNode(tree, 0, ast.NodeVarDecl, symIdx, 0)
	addNode(tree, decl, ast.NodeIntLit, 42, 0)
	
	ie.Infer() // This actually emits 3001
	// The type checker doesn't check Let mismatch (inference engine does it)
	// But it traverses.
	diags := ie.Infer()
	if len(diags) == 0 {
		t.Errorf("expected mismatch error from infer")
	}
	
	tcDiags := tc.Check()
	_ = tcDiags
}

func TestCheck_MutAssign(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	xID := pool.Intern([]byte("x"))
	symIdx, _ := st.Define(xID, sema.SymVar, sema.SymFlagMut, 0)
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeI32)
	
	assign := addNode(tree, 0, ast.NodeAssignStmt, 0, 0)
	addNode(tree, assign, ast.NodeIdent, symIdx, 0) // LHS
	addNode(tree, assign, ast.NodeIntLit, 42, 0)    // RHS
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors for mut assign: %v", diags)
	}
}

func TestCheck_ImmutableAssign(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	xID := pool.Intern([]byte("x"))
	symIdx, _ := st.Define(xID, sema.SymVar, 0, 0) // not mut
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeI32)
	
	assign := addNode(tree, 0, ast.NodeAssignStmt, 0, 0)
	addNode(tree, assign, ast.NodeIdent, symIdx, 0) // LHS
	addNode(tree, assign, ast.NodeIntLit, 42, 0)    // RHS
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3010 {
		t.Errorf("expected immutable assign error")
	}
}

func TestCheck_IfCondBool(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	ifStmt := addNode(tree, 0, ast.NodeIfStmt, 0, 0)
	addNode(tree, ifStmt, ast.NodeBoolLit, 0, 0)
	addNode(tree, ifStmt, ast.NodeBlock, 0, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheck_IfCondNotBool(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	ifStmt := addNode(tree, 0, ast.NodeIfStmt, 0, 0)
	addNode(tree, ifStmt, ast.NodeIntLit, 42, 0) // not bool
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3011 {
		t.Errorf("expected bool condition error")
	}
}

func TestCheck_ForRange(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	iID := pool.Intern([]byte("i"))
	symIdx, _ := st.Define(iID, sema.SymVar, 0, 0)
	
	forStmt := addNode(tree, 0, ast.NodeForStmt, 0, 0)
	addNode(tree, forStmt, ast.NodeIdent, symIdx, 0) // loop var
	addNode(tree, forStmt, ast.NodeBinaryExpr, 0, 0) // 0..10
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheck_ReturnMatch(t *testing.T) {
	tree, pool, st, tt, ie, tc := setupCheckTest()
	fooID := pool.Intern([]byte("foo"))
	symIdx, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	
	funcTypeID := tt.RegisterFunction([]types.TypeID{}, types.TypeI32, nil)
	st.SymbolAt(symIdx).TypeID = uint32(funcTypeID)
	
	fn := addNode(tree, 0, ast.NodeFuncDecl, symIdx, 0)
	ret := addNode(tree, fn, ast.NodeReturnStmt, 0, 0)
	addNode(tree, ret, ast.NodeIntLit, 42, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheck_ReturnMismatch(t *testing.T) {
	tree, pool, st, tt, ie, tc := setupCheckTest()
	fooID := pool.Intern([]byte("foo"))
	symIdx, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	
	funcTypeID := tt.RegisterFunction([]types.TypeID{}, types.TypeI32, nil)
	st.SymbolAt(symIdx).TypeID = uint32(funcTypeID)
	
	fn := addNode(tree, 0, ast.NodeFuncDecl, symIdx, 0)
	ret := addNode(tree, fn, ast.NodeReturnStmt, 0, 0)
	addNode(tree, ret, ast.NodeStringLit, 0, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3005 {
		t.Errorf("expected return type mismatch error")
	}
}

func TestCheck_ReturnVoid(t *testing.T) {
	tree, pool, st, tt, ie, tc := setupCheckTest()
	fooID := pool.Intern([]byte("foo"))
	symIdx, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	
	funcTypeID := tt.RegisterFunction([]types.TypeID{}, types.TypeVoid, nil)
	st.SymbolAt(symIdx).TypeID = uint32(funcTypeID)
	
	fn := addNode(tree, 0, ast.NodeFuncDecl, symIdx, 0)
	addNode(tree, fn, ast.NodeReturnStmt, 0, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheck_BreakInLoop(t *testing.T) {
	tree, pool, _, _, ie, tc := setupCheckTest()
	breakID := pool.Intern([]byte("break"))
	
	forStmt := addNode(tree, 0, ast.NodeForStmt, 0, 0)
	addNode(tree, forStmt, ast.NodeIdent, breakID, 0) // mock break
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheck_BreakOutsideLoop(t *testing.T) {
	tree, pool, _, _, ie, tc := setupCheckTest()
	breakID := pool.Intern([]byte("break"))
	addNode(tree, 0, ast.NodeIdent, breakID, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3013 {
		t.Errorf("expected break outside loop error")
	}
}

func TestCheck_DeferCall(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	deferStmt := addNode(tree, 0, ast.NodeDeferStmt, 0, 0)
	addNode(tree, deferStmt, ast.NodeCallExpr, 0, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheck_DeferNotCall(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	deferStmt := addNode(tree, 0, ast.NodeDeferStmt, 0, 0)
	addNode(tree, deferStmt, ast.NodeIntLit, 0, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3015 {
		t.Errorf("expected defer call error")
	}
}

func TestCheck_NestedScopes(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	// Just test deep traversal doesn't panic
	addNode(tree, 0, ast.NodeBlock, 0, 0)
	ie.Infer()
	tc.Check()
}
