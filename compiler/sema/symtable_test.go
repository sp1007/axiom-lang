package sema_test

import (
	"fmt"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
)

func TestNewSymbolTable_BuiltinsPresent(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	builtins := []string{"i32", "bool", "string", "void", "f64"}
	for _, name := range builtins {
		nameID := pool.Intern([]byte(name))
		idx, found := st.Resolve(nameID)
		if !found {
			t.Errorf("builtin %s not found", name)
			continue
		}
		sym := st.SymbolAt(idx)
		if sym.Kind != sema.SymBuiltinType {
			t.Errorf("builtin %s has wrong kind %v", name, sym.Kind)
		}
	}
}

func TestDefine_SingleScope(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	xID := pool.Intern([]byte("x"))
	idx, err := st.Define(xID, sema.SymVar, 0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolvedIdx, found := st.Resolve(xID)
	if !found {
		t.Fatalf("could not resolve x")
	}
	if resolvedIdx != idx {
		t.Fatalf("resolved to wrong symbol: got %d, want %d", resolvedIdx, idx)
	}
	
	sym := st.SymbolAt(idx)
	if sym.Kind != sema.SymVar || sym.DeclNode != 100 {
		t.Errorf("symbol fields incorrect: %+v", sym)
	}
}

func TestDefine_DuplicateError(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	xID := pool.Intern([]byte("x"))
	_, err := st.Define(xID, sema.SymVar, 0, 100)
	if err != nil {
		t.Fatalf("unexpected error on first define: %v", err)
	}

	_, err = st.Define(xID, sema.SymVar, 0, 101)
	if err == nil {
		t.Fatalf("expected error on duplicate define")
	}
	if err.Code != 2001 {
		t.Errorf("expected error code 2001, got %d", err.Code)
	}
}

func TestResolve_InnerToOuter(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	xID := pool.Intern([]byte("x"))
	globalIdx, _ := st.Define(xID, sema.SymVar, 0, 100)

	st.PushScope(sema.ScopeFunction)

	resolvedIdx, found := st.Resolve(xID)
	if !found || resolvedIdx != globalIdx {
		t.Errorf("could not resolve outer x from inner scope")
	}
}

func TestResolve_Shadowing(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	xID := pool.Intern([]byte("x"))
	globalIdx, _ := st.Define(xID, sema.SymVar, 0, 100)

	st.PushScope(sema.ScopeBlock)
	innerIdx, _ := st.Define(xID, sema.SymVar, 0, 200)

	resolvedIdx, found := st.Resolve(xID)
	if !found || resolvedIdx != innerIdx {
		t.Errorf("expected to resolve inner shadowed x")
	}

	st.PopScope()

	resolvedIdxAfter, foundAfter := st.Resolve(xID)
	if !foundAfter || resolvedIdxAfter != globalIdx {
		t.Errorf("expected to resolve outer x after pop")
	}
}

func TestResolve_NotFound(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	yID := pool.Intern([]byte("y"))
	_, found := st.Resolve(yID)
	if found {
		t.Errorf("should not find y")
	}
}

func TestPopScope_HidesInnerSymbols(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	st.PushScope(sema.ScopeBlock)
	xID := pool.Intern([]byte("x"))
	st.Define(xID, sema.SymVar, 0, 100)
	
	_, foundBefore := st.Resolve(xID)
	if !foundBefore {
		t.Fatalf("x should be found before pop")
	}

	st.PopScope()

	_, foundAfter := st.Resolve(xID)
	if foundAfter {
		t.Errorf("x should not be found after popping its scope")
	}
}

func TestPopScope_GlobalPanics(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when popping global scope")
		}
	}()

	st.PopScope()
}

func TestScopeKind_Preserved(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	idx := st.PushScope(sema.ScopeFunction)
	if st.CurrentScope() != idx {
		t.Errorf("CurrentScope mismatch")
	}
	if st.Scopes[st.CurrentScope()].Kind != sema.ScopeFunction {
		t.Errorf("ScopeKind mismatch")
	}
}

func TestScopeDepth_Increments(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	if st.CurrentDepth() != 0 {
		t.Errorf("expected depth 0, got %d", st.CurrentDepth())
	}

	st.PushScope(sema.ScopeBlock)
	st.PushScope(sema.ScopeBlock)
	st.PushScope(sema.ScopeBlock)

	if st.CurrentDepth() != 3 {
		t.Errorf("expected depth 3, got %d", st.CurrentDepth())
	}

	st.PopScope()
	
	if st.CurrentDepth() != 2 {
		t.Errorf("expected depth 2, got %d", st.CurrentDepth())
	}
}

func TestSymbolFlags_MoveTracking(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	xID := pool.Intern([]byte("x"))
	idx, _ := st.Define(xID, sema.SymVar, 0, 100)

	if st.IsMoved(idx) {
		t.Errorf("symbol should not be moved initially")
	}

	st.MarkMoved(idx)

	if !st.IsMoved(idx) {
		t.Errorf("symbol should be marked moved")
	}
}

func TestSymbolFlags_Preserved(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	xID := pool.Intern([]byte("x"))
	st.Define(xID, sema.SymVar, sema.SymFlagPub|sema.SymFlagMut, 100)
	
	idx, _ := st.Resolve(xID)
	sym := st.SymbolAt(idx)
	if sym.Flags&(sema.SymFlagPub|sema.SymFlagMut) != (sema.SymFlagPub | sema.SymFlagMut) {
		t.Errorf("flags not preserved: %v", sym.Flags)
	}
}

func TestResolveInScope_SpecificScope(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)

	scopeA := st.PushScope(sema.ScopeBlock)
	xID := pool.Intern([]byte("x"))
	st.Define(xID, sema.SymVar, 0, 100)

	st.PushScope(sema.ScopeBlock) // nested scope
	
	st.PopScope() // pop back to A
	st.PopScope() // pop back to global
	
	scopeB := st.PushScope(sema.ScopeBlock) // sibling scope to A
	yID := pool.Intern([]byte("y"))
	st.Define(yID, sema.SymVar, 0, 200)

	_, foundXA := st.ResolveInScope(xID, scopeA)
	if !foundXA { t.Errorf("expected x in scopeA") }
	
	_, foundXB := st.ResolveInScope(xID, scopeB)
	if foundXB { t.Errorf("did not expect x in scopeB") }
}

func TestHashMapGrowth(t *testing.T) {
	pool := ast.NewInternPool(1024)
	st := sema.NewSymbolTable(pool)
	
	st.PushScope(sema.ScopeFunction)

	// Insert 100 symbols to force multiple resize operations
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("var_%d", i)
		nameID := pool.Intern([]byte(name))
		st.Define(nameID, sema.SymVar, 0, uint32(i))
	}

	// Verify all are resolvable
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("var_%d", i)
		nameID := pool.Intern([]byte(name))
		_, found := st.Resolve(nameID)
		if !found {
			t.Errorf("failed to resolve %s after map growth", name)
		}
	}
}

func TestDeterminism(t *testing.T) {
	pool := ast.NewInternPool(16)
	st1 := sema.NewSymbolTable(pool)
	st2 := sema.NewSymbolTable(pool)

	xID := pool.Intern([]byte("x"))
	yID := pool.Intern([]byte("y"))

	idx1X, _ := st1.Define(xID, sema.SymVar, 0, 1)
	idx1Y, _ := st1.Define(yID, sema.SymVar, 0, 2)
	
	idx2X, _ := st2.Define(xID, sema.SymVar, 0, 1)
	idx2Y, _ := st2.Define(yID, sema.SymVar, 0, 2)

	if idx1X != idx2X || idx1Y != idx2Y {
		t.Errorf("symbol indices are not deterministic")
	}
}

func TestResolve_NeverPanics(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	
	st.PushScope(sema.ScopeFunction)
	st.PushScope(sema.ScopeBlock)
	xID := pool.Intern([]byte("x"))
	st.Define(xID, sema.SymVar, 0, 100)
	st.PopScope()
	st.Resolve(xID) // this is totally fine, just returns false
}

func BenchmarkResolve_Depth10(b *testing.B) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	
	for i := 0; i < 10; i++ {
		st.PushScope(sema.ScopeBlock)
	}
	
	xID := pool.Intern([]byte("x"))
	st.Define(xID, sema.SymVar, 0, 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Resolve(xID)
	}
}

func BenchmarkDefine_1000Symbols(b *testing.B) {
	pool := ast.NewInternPool(2048)
	names := make([]uint32, 1000)
	for i := 0; i < 1000; i++ {
		names[i] = pool.Intern([]byte(fmt.Sprintf("v%d", i)))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st := sema.NewSymbolTable(pool)
		st.PushScope(sema.ScopeFunction)
		for j := 0; j < 1000; j++ {
			st.Define(names[j], sema.SymVar, 0, uint32(j))
		}
	}
}
