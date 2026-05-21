package cgen

// DeferEntry represents a single deferred call with its scope context.
type DeferEntry struct {
	CallExprIdx uint32 // AST node index of the deferred expression
	ScopeDepth  int    // depth at which this defer was registered
}

// EnhancedDeferStack extends the basic DeferStack with scope-depth tracking.
// It supports emitting defers for specific scope depths (e.g., break/continue
// should only emit defers from the current loop scope, not the entire function).
type EnhancedDeferStack struct {
	entries []DeferEntry
	depth   int
}

// NewEnhancedDeferStack creates an empty EnhancedDeferStack.
func NewEnhancedDeferStack() *EnhancedDeferStack {
	return &EnhancedDeferStack{}
}

// PushScope enters a new defer scope, increasing depth.
func (ds *EnhancedDeferStack) PushScope() {
	ds.depth++
}

// PopScope exits the current scope, returning all deferred entries for this depth
// in LIFO order. Depth is decremented.
func (ds *EnhancedDeferStack) PopScope() []DeferEntry {
	var result []DeferEntry
	i := len(ds.entries) - 1
	for i >= 0 && ds.entries[i].ScopeDepth == ds.depth {
		result = append(result, ds.entries[i])
		i--
	}
	ds.entries = ds.entries[:i+1]
	ds.depth--
	return result
}

// Push adds a deferred expression to the current scope.
func (ds *EnhancedDeferStack) Push(callExprIdx uint32) {
	ds.entries = append(ds.entries, DeferEntry{
		CallExprIdx: callExprIdx,
		ScopeDepth:  ds.depth,
	})
}

// CurrentScopeDefers returns the deferred entries for the current scope in LIFO order
// without popping the scope. Used for break/continue emission.
func (ds *EnhancedDeferStack) CurrentScopeDefers() []DeferEntry {
	var result []DeferEntry
	for i := len(ds.entries) - 1; i >= 0; i-- {
		if ds.entries[i].ScopeDepth == ds.depth {
			result = append(result, ds.entries[i])
		} else {
			break
		}
	}
	return result
}

// AllDefers returns ALL deferred entries across all scopes from deepest to shallowest
// in LIFO order. Used for return statements that must emit all defers.
func (ds *EnhancedDeferStack) AllDefers() []DeferEntry {
	result := make([]DeferEntry, len(ds.entries))
	for i, j := 0, len(ds.entries)-1; j >= 0; j, i = j-1, i+1 {
		result[i] = ds.entries[j]
	}
	return result
}

// Depth returns the current scope depth.
func (ds *EnhancedDeferStack) Depth() int {
	return ds.depth
}

// EmitDeferredCalls emits the C code for a list of deferred entries.
// Each deferred expression is emitted as a statement via the ExprGen.
func EmitDeferredCalls(w *IndentWriter, entries []DeferEntry, exprGen *ExprGen) {
	for _, entry := range entries {
		expr := exprGen.Emit(entry.CallExprIdx)
		w.Linef("%s;", expr)
	}
}
