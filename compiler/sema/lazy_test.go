package sema_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func setupLazyTest() (*ast.InternPool, *sema.SymbolTable, *types.TypeTable) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	return pool, st, tt
}

func TestRegisterImport(t *testing.T) {
	pool, st, tt := setupLazyTest()
	lr := sema.NewLazyResolver(st, tt, nil)

	mathID := pool.Intern([]byte("math"))
	symIdx, diag := lr.RegisterImport(mathID, "math.ax", 0, 100)
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %v", diag)
	}

	sym := st.SymbolAt(symIdx)
	if sym.Kind != sema.SymModule {
		t.Errorf("expected SymModule, got %v", sym.Kind)
	}
}

func TestResolveField_Found(t *testing.T) {
	pool, st, tt := setupLazyTest()
	
	sqrtID := pool.Intern([]byte("sqrt"))
	
	loader := func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		// mock export
		idx, _ := st.Define(sqrtID, sema.SymFunc, 0, 1)
		m.Exports[sqrtID] = idx
		return nil
	}
	
	lr := sema.NewLazyResolver(st, tt, loader)

	mathID := pool.Intern([]byte("math"))
	lr.RegisterImport(mathID, "math.ax", 0, 100)

	idx, diag := lr.ResolveField(mathID, sqrtID, diagnostics.Pos{})
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %v", diag)
	}

	sym := st.SymbolAt(idx)
	if sym.Kind != sema.SymFunc {
		t.Errorf("expected SymFunc")
	}
}

func TestResolveField_NotFound(t *testing.T) {
	pool, st, tt := setupLazyTest()
	lr := sema.NewLazyResolver(st, tt, nil)

	mathID := pool.Intern([]byte("math"))
	lr.RegisterImport(mathID, "math.ax", 0, 100)

	missingID := pool.Intern([]byte("missing"))
	_, diag := lr.ResolveField(mathID, missingID, diagnostics.Pos{})
	
	if diag == nil {
		t.Fatalf("expected diagnostic for missing field")
	}
	if diag.Code != 2005 {
		t.Errorf("expected error code 2005, got %d", diag.Code)
	}
}

func TestLazyLoading_OnDemand(t *testing.T) {
	pool, st, tt := setupLazyTest()
	
	loaded := false
	loader := func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		loaded = true
		return nil
	}
	lr := sema.NewLazyResolver(st, tt, loader)

	mathID := pool.Intern([]byte("math"))
	lr.RegisterImport(mathID, "math.ax", 0, 100)

	if loaded {
		t.Errorf("module should not be loaded on register")
	}

	// Trigger load
	missingID := pool.Intern([]byte("missing"))
	lr.ResolveField(mathID, missingID, diagnostics.Pos{})

	if !loaded {
		t.Errorf("module should be loaded on field access")
	}
}

func TestLazyLoading_NoEagerLoad(t *testing.T) {
	pool, st, tt := setupLazyTest()
	loaded := false
	loader := func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		loaded = true
		return nil
	}
	lr := sema.NewLazyResolver(st, tt, loader)

	mathID := pool.Intern([]byte("math"))
	lr.RegisterImport(mathID, "math.ax", 0, 100)

	if loaded {
		t.Errorf("module should not load if never accessed")
	}
}

func TestCycleDetection(t *testing.T) {
	pool, st, tt := setupLazyTest()
	
	mathID := pool.Intern([]byte("math"))
	sqrtID := pool.Intern([]byte("sqrt"))

	// Create a loader that tries to resolve a field from its own module while loading
	var lr *sema.LazyResolver
	loader := func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		_, diag := lr.ResolveField(mathID, sqrtID, diagnostics.Pos{})
		if diag == nil {
			t.Errorf("expected cycle detection diagnostic")
		} else if diag.Code != 2003 {
			t.Errorf("expected code 2003, got %d", diag.Code)
		}
		return nil
	}
	lr = sema.NewLazyResolver(st, tt, loader)

	lr.RegisterImport(mathID, "math.ax", 0, 100)
	
	// This will trigger the loader, which will trigger the recursive resolve
	lr.ResolveField(mathID, sqrtID, diagnostics.Pos{})
}

func TestUnusedImportDetection(t *testing.T) {
	pool, st, tt := setupLazyTest()
	lr := sema.NewLazyResolver(st, tt, nil)

	mathID := pool.Intern([]byte("math"))
	lr.RegisterImport(mathID, "math.ax", 0, 100)

	diags := lr.CheckUnusedImports(pool)
	if len(diags) != 1 {
		t.Fatalf("expected 1 unused import warning, got %d", len(diags))
	}
	
	if !strings.Contains(diags[0].Message, "'math'") {
		t.Errorf("expected message to mention 'math'")
	}
}

func TestMultipleImports(t *testing.T) {
	pool, st, tt := setupLazyTest()
	
	loadCounts := make(map[uint32]int)
	loader := func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		loadCounts[m.NameID]++
		return nil
	}
	lr := sema.NewLazyResolver(st, tt, loader)

	aID := pool.Intern([]byte("a"))
	bID := pool.Intern([]byte("b"))
	cID := pool.Intern([]byte("c"))

	lr.RegisterImport(aID, "a.ax", 0, 1)
	lr.RegisterImport(bID, "b.ax", 0, 2)
	lr.RegisterImport(cID, "c.ax", 0, 3)

	// Access only b
	fooID := pool.Intern([]byte("foo"))
	lr.ResolveField(bID, fooID, diagnostics.Pos{})

	if loadCounts[aID] != 0 { t.Errorf("module a should not be loaded") }
	if loadCounts[bID] != 1 { t.Errorf("module b should be loaded once") }
	if loadCounts[cID] != 0 { t.Errorf("module c should not be loaded") }
}

func TestMultipleFields_SameModule(t *testing.T) {
	pool, st, tt := setupLazyTest()
	
	loadCount := 0
	loader := func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		loadCount++
		// mock export
		idx, _ := st.Define(pool.Intern([]byte("f1")), sema.SymFunc, 0, 1)
		m.Exports[pool.Intern([]byte("f1"))] = idx
		
		idx2, _ := st.Define(pool.Intern([]byte("f2")), sema.SymFunc, 0, 2)
		m.Exports[pool.Intern([]byte("f2"))] = idx2
		return nil
	}
	lr := sema.NewLazyResolver(st, tt, loader)

	modID := pool.Intern([]byte("mod"))
	lr.RegisterImport(modID, "mod.ax", 0, 100)

	f1ID := pool.Intern([]byte("f1"))
	f2ID := pool.Intern([]byte("f2"))

	lr.ResolveField(modID, f1ID, diagnostics.Pos{})
	lr.ResolveField(modID, f2ID, diagnostics.Pos{})

	if loadCount != 1 {
		t.Errorf("module should be loaded exactly once, was loaded %d times", loadCount)
	}
}

func TestResolveField_ModuleNotImported(t *testing.T) {
	pool, st, tt := setupLazyTest()
	lr := sema.NewLazyResolver(st, tt, nil)

	modID := pool.Intern([]byte("unknown"))
	fID := pool.Intern([]byte("f"))

	_, diag := lr.ResolveField(modID, fID, diagnostics.Pos{})
	if diag == nil {
		t.Fatalf("expected diagnostic for unknown module")
	}
	if diag.Code != 2002 {
		t.Errorf("expected code 2002, got %d", diag.Code)
	}
}

func TestImportShadowing(t *testing.T) {
	pool, st, tt := setupLazyTest()
	lr := sema.NewLazyResolver(st, tt, nil)

	xID := pool.Intern([]byte("x"))
	
	// Global import
	idx1, _ := lr.RegisterImport(xID, "global_x.ax", 0, 10)
	
	// Inner scope import
	st.PushScope(sema.ScopeFunction)
	idx2, _ := lr.RegisterImport(xID, "local_x.ax", 0, 20)
	
	resolvedIdx, ok := st.Resolve(xID)
	if !ok {
		t.Fatalf("should resolve x")
	}
	
	if resolvedIdx != idx2 {
		t.Errorf("inner import should shadow global import")
	}
	
	st.PopScope()
	
	resolvedIdx2, _ := st.Resolve(xID)
	if resolvedIdx2 != idx1 {
		t.Errorf("global import should be visible after pop")
	}
}

func TestLazyDeterminism(t *testing.T) {
	pool, st1, tt1 := setupLazyTest()
	lr1 := sema.NewLazyResolver(st1, tt1, nil)
	
	_, st2, tt2 := setupLazyTest()
	lr2 := sema.NewLazyResolver(st2, tt2, nil)

	aID := pool.Intern([]byte("a"))
	
	idx1, _ := lr1.RegisterImport(aID, "a.ax", 0, 1)
	idx2, _ := lr2.RegisterImport(aID, "a.ax", 0, 1)
	
	if idx1 != idx2 {
		t.Errorf("import registration not deterministic")
	}
}
