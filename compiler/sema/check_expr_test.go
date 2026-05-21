package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// Re-use setupCheckTest from check_stmt_test.go, so we just declare it here if needed,
// but since they are in the same package (sema_test) we can just call setupCheckTest().
// However, the node appending is also re-used.

func TestCheckExpr_BinaryHandledByInfer(t *testing.T) {
	// Binary exprs are handled by InferenceEngine, so we just verify they pass without errors.
	tree, _, _, _, ie, tc := setupCheckTest()
	bin := addNode(tree, 0, ast.NodeBinaryExpr, 0, 0)
	addNode(tree, bin, ast.NodeIntLit, 1, 0)
	addNode(tree, bin, ast.NodeIntLit, 2, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheckExpr_UnaryNumeric(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	unary := addNode(tree, 0, ast.NodeUnaryExpr, 0, 1) // 1 = -
	addNode(tree, unary, ast.NodeIntLit, 42, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheckExpr_UnaryNumericError(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	unary := addNode(tree, 0, ast.NodeUnaryExpr, 0, 1) // 1 = -
	addNode(tree, unary, ast.NodeBoolLit, 0, 0) // bool
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3020 {
		t.Errorf("expected unary - error for bool")
	}
}

func TestCheckExpr_UnaryBool(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	unary := addNode(tree, 0, ast.NodeUnaryExpr, 0, 2) // 2 = not
	addNode(tree, unary, ast.NodeBoolLit, 0, 0)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheckExpr_IndexExprValid(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	arrID := pool.Intern([]byte("arr"))
	symIdx, _ := st.Define(arrID, sema.SymVar, 0, 0)
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeString) // string can be indexed
	
	idx := addNode(tree, 0, ast.NodeIndexExpr, 0, 0)
	addNode(tree, idx, ast.NodeIdent, symIdx, 0) // arr
	addNode(tree, idx, ast.NodeIntLit, 0, 0)     // 0
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheckExpr_IndexExprInvalidCol(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	arrID := pool.Intern([]byte("arr"))
	symIdx, _ := st.Define(arrID, sema.SymVar, 0, 0)
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeI32) // i32 cannot be indexed
	
	idx := addNode(tree, 0, ast.NodeIndexExpr, 0, 0)
	addNode(tree, idx, ast.NodeIdent, symIdx, 0) // arr
	addNode(tree, idx, ast.NodeIntLit, 0, 0)     // 0
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3023 {
		t.Errorf("expected index collection error")
	}
}

func TestCheckExpr_IndexExprInvalidIdx(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	arrID := pool.Intern([]byte("arr"))
	symIdx, _ := st.Define(arrID, sema.SymVar, 0, 0)
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeString)
	
	idx := addNode(tree, 0, ast.NodeIndexExpr, 0, 0)
	addNode(tree, idx, ast.NodeIdent, symIdx, 0)    // arr
	addNode(tree, idx, ast.NodeStringLit, 0, 0)     // "0" (not integer)
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3024 {
		t.Errorf("expected index type error")
	}
}

func TestCheckExpr_FieldExprValid(t *testing.T) {
	tree, pool, st, tt, ie, tc := setupCheckTest()
	objID := pool.Intern([]byte("obj"))
	symIdx, _ := st.Define(objID, sema.SymVar, 0, 0)
	
	fieldID := pool.Intern([]byte("f"))
	structType := tt.RegisterStruct(0, []types.FieldEntry{{NameID: fieldID, TypeID: types.TypeI32}}, nil)
	st.SymbolAt(symIdx).TypeID = uint32(structType)
	
	field := addNode(tree, 0, ast.NodeFieldExpr, fieldID, 0)
	addNode(tree, field, ast.NodeIdent, symIdx, 0) // obj
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheckExpr_FieldExprInvalidObj(t *testing.T) {
	tree, pool, st, _, ie, tc := setupCheckTest()
	objID := pool.Intern([]byte("obj"))
	symIdx, _ := st.Define(objID, sema.SymVar, 0, 0)
	st.SymbolAt(symIdx).TypeID = uint32(types.TypeI32) // i32 has no fields
	
	field := addNode(tree, 0, ast.NodeFieldExpr, 0, 0)
	addNode(tree, field, ast.NodeIdent, symIdx, 0) // obj
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3025 {
		t.Errorf("expected field object error")
	}
}

func TestCheckExpr_CastLegal(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	cast := addNode(tree, 0, ast.NodeCastExpr, uint32(types.TypeF64), 0)
	addNode(tree, cast, ast.NodeIntLit, 42, 0) // int -> float is legal
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) > 0 {
		t.Errorf("unexpected errors: %v", diags)
	}
}

func TestCheckExpr_CastIllegal(t *testing.T) {
	tree, _, _, _, ie, tc := setupCheckTest()
	cast := addNode(tree, 0, ast.NodeCastExpr, uint32(types.TypeI32), 0)
	addNode(tree, cast, ast.NodeStringLit, 0, 0) // string -> int is illegal
	
	ie.Infer()
	diags := tc.Check()
	if len(diags) == 0 || diags[0].Code != 3026 {
		t.Errorf("expected illegal cast error")
	}
}

// Await test removed pending Future type implementation
