package cgen_test

import (
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
)

func TestEnhancedDeferStack_LIFO(t *testing.T) {
	ds := cgen.NewEnhancedDeferStack()
	ds.PushScope()
	ds.Push(10)
	ds.Push(20)
	ds.Push(30)

	result := ds.PopScope()
	if len(result) != 3 {
		t.Fatalf("expected 3 defers, got %d", len(result))
	}
	// LIFO: 30, 20, 10
	if result[0].CallExprIdx != 30 || result[1].CallExprIdx != 20 || result[2].CallExprIdx != 10 {
		t.Errorf("LIFO order incorrect: %v", result)
	}
}

func TestEnhancedDeferStack_NestedScopes(t *testing.T) {
	ds := cgen.NewEnhancedDeferStack()
	ds.PushScope() // depth 1
	ds.Push(1)     // outer

	ds.PushScope() // depth 2
	ds.Push(2)     // inner
	ds.Push(3)     // inner

	// Pop inner scope — should only get entries from depth 2
	inner := ds.PopScope()
	if len(inner) != 2 {
		t.Fatalf("inner scope: expected 2 defers, got %d", len(inner))
	}
	if inner[0].CallExprIdx != 3 || inner[1].CallExprIdx != 2 {
		t.Errorf("inner defers incorrect: %v", inner)
	}

	// Pop outer scope — should only get entries from depth 1
	outer := ds.PopScope()
	if len(outer) != 1 {
		t.Fatalf("outer scope: expected 1 defer, got %d", len(outer))
	}
	if outer[0].CallExprIdx != 1 {
		t.Errorf("outer defer = %d, want 1", outer[0].CallExprIdx)
	}
}

func TestEnhancedDeferStack_CurrentScopeDefers(t *testing.T) {
	ds := cgen.NewEnhancedDeferStack()
	ds.PushScope()
	ds.Push(1) // outer

	ds.PushScope()
	ds.Push(2) // inner
	ds.Push(3) // inner

	// CurrentScopeDefers should return inner scope only, without popping
	current := ds.CurrentScopeDefers()
	if len(current) != 2 {
		t.Fatalf("expected 2 current scope defers, got %d", len(current))
	}
	// Should be LIFO: 3, 2
	if current[0].CallExprIdx != 3 || current[1].CallExprIdx != 2 {
		t.Errorf("current scope defers incorrect: %v", current)
	}

	// Verify scope was NOT popped
	if ds.Depth() != 2 {
		t.Errorf("depth should still be 2, got %d", ds.Depth())
	}
}

func TestEnhancedDeferStack_AllDefers(t *testing.T) {
	ds := cgen.NewEnhancedDeferStack()
	ds.PushScope() // depth 1
	ds.Push(1)

	ds.PushScope() // depth 2
	ds.Push(2)
	ds.Push(3)

	// AllDefers returns everything in LIFO order (3, 2, 1) for return statements
	all := ds.AllDefers()
	if len(all) != 3 {
		t.Fatalf("expected 3 total defers, got %d", len(all))
	}
	if all[0].CallExprIdx != 3 || all[1].CallExprIdx != 2 || all[2].CallExprIdx != 1 {
		t.Errorf("all defers order incorrect: %v", all)
	}
}

func TestEnhancedDeferStack_EmptyScope(t *testing.T) {
	ds := cgen.NewEnhancedDeferStack()
	ds.PushScope()

	result := ds.PopScope()
	if len(result) != 0 {
		t.Errorf("empty scope should return 0 defers, got %d", len(result))
	}
}

func TestEnhancedDeferStack_DepthTracking(t *testing.T) {
	ds := cgen.NewEnhancedDeferStack()
	if ds.Depth() != 0 {
		t.Errorf("initial depth should be 0, got %d", ds.Depth())
	}

	ds.PushScope()
	if ds.Depth() != 1 {
		t.Errorf("after push, depth should be 1, got %d", ds.Depth())
	}

	ds.PushScope()
	if ds.Depth() != 2 {
		t.Errorf("after second push, depth should be 2, got %d", ds.Depth())
	}

	ds.PopScope()
	if ds.Depth() != 1 {
		t.Errorf("after pop, depth should be 1, got %d", ds.Depth())
	}
}

func TestEnhancedDeferStack_LoopDeferPattern(t *testing.T) {
	// Simulates defer inside a loop body:
	// fn example():
	//     defer a()
	//     while cond:
	//         defer b()
	//         // break should emit b() only, not a()
	ds := cgen.NewEnhancedDeferStack()
	ds.PushScope() // function scope
	ds.Push(100)   // defer a()

	// Simulate loop iteration 1
	ds.PushScope() // loop body scope
	ds.Push(200)   // defer b()

	// Break: emit only current scope defers
	breakDefers := ds.CurrentScopeDefers()
	if len(breakDefers) != 1 {
		t.Fatalf("break should see 1 defer, got %d", len(breakDefers))
	}
	if breakDefers[0].CallExprIdx != 200 {
		t.Errorf("break defer = %d, want 200", breakDefers[0].CallExprIdx)
	}

	// End of iteration: pop loop body
	ds.PopScope()

	// Return: emit ALL defers
	returnDefers := ds.AllDefers()
	if len(returnDefers) != 1 {
		t.Fatalf("return should see 1 defer (a), got %d", len(returnDefers))
	}
	if returnDefers[0].CallExprIdx != 100 {
		t.Errorf("return defer = %d, want 100", returnDefers[0].CallExprIdx)
	}
}
