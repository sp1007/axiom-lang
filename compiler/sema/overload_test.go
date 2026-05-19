package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func setupOverloadTest() (*ast.InternPool, *sema.SymbolTable, *types.TypeTable, *sema.OverloadResolver) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	or := sema.NewOverloadResolver(st, tt)
	return pool, st, tt, or
}

func TestOverloadExact(t *testing.T) {
	pool, st, tt, or := setupOverloadTest()
	fooID := pool.Intern([]byte("foo"))
	
	// Create fn foo(x: i32)
	symIdx, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	funcTypeID := tt.RegisterFunction([]types.TypeID{types.TypeI32}, types.TypeVoid, nil)
	st.SymbolAt(symIdx).TypeID = uint32(funcTypeID)
	
	res, err := or.Resolve(fooID, []types.TypeID{types.TypeI32})
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Message)
	}
	if res.SymbolID != symIdx {
		t.Errorf("expected symbol %d, got %d", symIdx, res.SymbolID)
	}
	if res.Score != 4 {
		t.Errorf("expected score 4 for exact match, got %d", res.Score)
	}
}

func TestOverloadCoercible(t *testing.T) {
	pool, st, tt, or := setupOverloadTest()
	fooID := pool.Intern([]byte("foo"))
	
	// Create fn foo(x: i64)
	symIdx, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	funcTypeID := tt.RegisterFunction([]types.TypeID{types.TypeI64}, types.TypeVoid, nil)
	st.SymbolAt(symIdx).TypeID = uint32(funcTypeID)
	
	res, err := or.Resolve(fooID, []types.TypeID{types.TypeI32})
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Message)
	}
	if res.SymbolID != symIdx {
		t.Errorf("expected symbol %d, got %d", symIdx, res.SymbolID)
	}
	if res.Score != 3 {
		t.Errorf("expected score 3 for coercible match, got %d", res.Score)
	}
}

func TestOverloadAmbiguous(t *testing.T) {
	pool, st, tt, or := setupOverloadTest()
	fooID := pool.Intern([]byte("foo"))
	
	// Create fn foo(x: i64) in global scope
	symIdx1, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	funcTypeID1 := tt.RegisterFunction([]types.TypeID{types.TypeI64}, types.TypeVoid, nil)
	st.SymbolAt(symIdx1).TypeID = uint32(funcTypeID1)
	
	// Push scope, create fn foo(x: f64)
	st.PushScope(sema.ScopeBlock)
	symIdx2, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	funcTypeID2 := tt.RegisterFunction([]types.TypeID{types.TypeF64}, types.TypeVoid, nil)
	st.SymbolAt(symIdx2).TypeID = uint32(funcTypeID2)
	
	// Call with i32. Both i64 and f64 are coercible targets, score 3.
	_, err := or.Resolve(fooID, []types.TypeID{types.TypeI32})
	if err == nil || err.Code != 3031 {
		t.Errorf("expected ambiguity error, got %v", err)
	}
}

func TestOverloadNoMatch(t *testing.T) {
	pool, st, tt, or := setupOverloadTest()
	fooID := pool.Intern([]byte("foo"))
	
	// Create fn foo(x: string)
	symIdx, _ := st.Define(fooID, sema.SymFunc, 0, 0)
	funcTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString}, types.TypeVoid, nil)
	st.SymbolAt(symIdx).TypeID = uint32(funcTypeID)
	
	// Call with i32
	_, err := or.Resolve(fooID, []types.TypeID{types.TypeI32})
	if err == nil || err.Code != 3030 {
		t.Errorf("expected no match error, got %v", err)
	}
}

func TestBuiltinPlus(t *testing.T) {
	_, _, _, or := setupOverloadTest()
	
	res := or.ResolveBuiltinOp("+", types.TypeI32, types.TypeI32)
	if res != types.TypeI32 {
		t.Errorf("expected i32")
	}
	
	res = or.ResolveBuiltinOp("+", types.TypeI32, types.TypeI64)
	if res != types.TypeI64 {
		t.Errorf("expected i64 (widening)")
	}
}

func TestBuiltinStringPlus(t *testing.T) {
	_, _, _, or := setupOverloadTest()
	
	res := or.ResolveBuiltinOp("+", types.TypeString, types.TypeString)
	if res != types.TypeString {
		t.Errorf("expected string")
	}
}
