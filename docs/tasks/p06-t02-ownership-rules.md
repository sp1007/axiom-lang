# p06-t02: Ownership Rules Checker

## Purpose
Enforce AXIOM's single-ownership rules at compile time: each value has exactly one owner, moving a value invalidates the source, borrowed values cannot be stored or moved, and mutable access is controlled. This is the core memory safety pass that eliminates use-after-free and data races at compile time.

## Context
AXIOM's ownership is simpler than Rust's borrow checker: there's no lifetime annotation on references. Instead, all borrows are local (cannot escape their scope), and the Generational References runtime catches any use-after-free that slips through (e.g., via unsafe blocks). The ownership checker enforces the compile-time rules; the runtime handles the escape valve.

## Inputs
- Typed, resolved AST
- ConnectionGraph (p06-t01) — populated by this pass
- SymbolTable — tracking symbol states

## Outputs
- Populated ConnectionGraph (Owns/Borrows/FlowsTo/EscapesTo edges added)
- `MovedSet{symIDs map[uint32]bool}` — symbols that have been moved
- `[]Diagnostic` — ownership violations

## Dependencies
- p06-t01: connection-graph — the graph to populate
- p04-t06: type-checker-statements — type info needed for ownership rules

## Subsystems Affected
- Memory safety: primary compile-time safety mechanism
- CTGC (p06-t05): reads ownership info to inject destroys
- Escape analysis (p06-t04): reads connection graph edges

## Detailed Requirements

1. `OwnershipChecker` struct: `cg *ConnectionGraph, moved map[uint32]bool, st *SymbolTable`
2. **Move rule**: when a value is used in a position that consumes it (assigned to another variable, passed as `!T` sink parameter, returned from function), add `FlowsTo` edge, mark source as moved. Any subsequent use of a moved value → error: `"use of moved value 'x'"`.
3. **Borrow rule** (`lent T` parameter): passing a value as `lent` → add `Borrows` edge from param to arg. The borrow cannot escape (no EscapesTo). Duration: limited to the call.
4. **Mut rule**: only `mut` symbols can be assigned to; reading is always OK.
5. **Isolated rule**: value declared as `Isolated[T]` → verify ConnectionGraph has no incoming EscapesTo edges from outside the value's subgraph.
6. **Escape tracking**: when a value is:
   - Returned from function → `EscapesTo` edge to return slot (heap)
   - Stored in a field of a heap-allocated struct → `EscapesTo` heap
   - Passed to `spawn` → `EscapesTo` actor heap (must be `Isolated[T]`)
   - Closed over in a closure → `EscapesTo` closure capture
7. Only values that escape are heap-allocated (escape analysis result, used in codegen).
8. `IsolatedVerifier`: called after building the graph — checks all `Isolated[T]` values have no external incoming edges.

## Implementation Steps

1. Create `compiler/sema/ownership.go`.
2. Implement `CheckOwnership(tree, st, tt)` — walk all reachable function bodies.
3. For each VarDecl: add ValueNode to ConnectionGraph.
4. For each AssignStmt `y = x`: add `FlowsTo(x_node, y_node)`, mark x as moved.
5. For each CallExpr with `!T` (sink) param: mark arg as moved.
6. For each CallExpr with `lent T` param: add `Borrows(param_node, arg_node)`.
7. For each return: add `EscapesTo(retval_node, RETURN_SLOT)`.
8. After each use of a symbol: check if it's in MovedSet → error.
9. Call `IsolatedVerifier` after building graph.
10. Write tests for each rule.

## Test Plan

- `TestMoveRule`: `let x = Foo{}; let y = x; use(x)` → "use of moved value 'x'"
- `TestBorrowRule`: pass to `lent T` param — no move, can use after
- `TestMutRule`: `let x = 5; x = 10` → "cannot assign to immutable 'x'"
- `TestEscapeReturn`: returned value → EscapesTo edge → heap allocated
- `TestIsolatedOk`: value with no external refs passed to spawn → OK
- `TestIsolatedFail`: value with external ref passed to spawn → error

## Validation Checklist

- [ ] Moved values cannot be used after move
- [ ] Borrowed values cannot be moved or stored
- [ ] Immutable variables cannot be assigned
- [ ] Isolated[T] verified at spawn sites
- [ ] ConnectionGraph populated with correct edges
- [ ] Error messages include variable names and locations

## Acceptance Criteria

- Compliance tests 031-040 (struct/ownership group) pass
- Use-after-move detected at compile time with clear error message

## Definition of Done

- [ ] `compiler/sema/ownership.go` implemented
- [ ] ConnectionGraph population complete
- [ ] Unit tests pass
- [ ] Integrated into compiler pipeline after type checking

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| False positives on complex ownership patterns | Start conservative; add escape hatches via `unsafe` block |
| Move detection misses indirect moves (via function calls) | Track !T sink params; all !T args are moves |

## Future Follow-up Tasks

- p06-t04: escape-analysis uses the populated ConnectionGraph
- p06-t05: CTGC inject destroy nodes based on ownership analysis
