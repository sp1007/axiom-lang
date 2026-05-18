# p06-t05: CTGC Destroy Injection

## Purpose
Implement Compile-Time Garbage Collection (CTGC) by automatically injecting `=destroy` operations at the exit of each scope for all owned heap values that have not been moved or returned. This provides automatic memory management without a garbage collector's runtime overhead.

## Context
CTGC is analogous to C++ destructors or Rust's `Drop` trait, but fully automatic and compiler-injected. When a scope exits (function return, if/else branch end, for loop end), the compiler identifies all heap-allocated values declared in that scope that are still alive (not moved, not returned) and injects `=destroy` nodes into the AST. These become `ax_free()` calls in the C backend, with the generational ID incremented to invalidate any outstanding references.

## Inputs
- Ownership-analyzed AST with MovedSet and escape annotations
- SymbolTable — scope membership of each symbol
- ConnectionGraph — for destroy ordering (avoid double-free)

## Outputs
- New `NodeKind.DestroyStmt` nodes injected into AST at scope exits
- Modified AST with automatic memory management
- `compiler/sema/ctgc.go`

## Dependencies
- p06-t04: escape-analysis — only heap values (EscapesToHeap flag) need destroys
- p06-t02: ownership-rules — MovedSet consulted to skip moved values
- p06-t01: connection-graph — ordering of destroys

## Subsystems Affected
- AST: new DestroyStmt nodes injected
- Code generation: DestroyStmt → ax_free() in C backend
- Memory safety: CTGC eliminates memory leaks without GC

## Detailed Requirements

1. `CTGCPass` struct: `tree *AstTree, st *SymbolTable, moved map[uint32]bool`
2. `InjectDestroys(funcNodeIdx uint32)`:
   - Walk all Block nodes in the function (scope boundaries)
   - At each block exit: collect all VarDecl nodes in that block with `EscapesToHeap` set AND NOT in MovedSet AND TypeID is not primitive
   - Inject `DestroyStmt{target: symID}` nodes in REVERSE declaration order (LIFO)
3. LIFO ordering: destroy last-declared first — prevents use-after-destroy.
4. `NodeKind.DestroyStmt` contains the SymID of the value to destroy.
5. Skip injection for:
   - Stack-allocated values (no EscapesToHeap flag)
   - Values already moved (in MovedSet)
   - Returned values (their ownership transferred)
   - Primitive types (i32, f64, bool — no heap allocation)
   - Values in `Arena` blocks (arena handles bulk free)
6. For values returned from a function: no destroy in the returning scope (ownership transferred to caller). Caller's scope will destroy it at its scope exit.
7. Handle early returns: inject destroys for all still-alive heap values before each `return` statement.

## Implementation Steps

1. Create `compiler/sema/ctgc.go` with `CTGCPass`.
2. Add `NodeKind.DestroyStmt` to `ast/node.go`.
3. Implement `InjectDestroys()`: walk blocks, at each block's last position inject destroys.
4. Handle early returns: find all `ReturnStmt` nodes, before each inject destroys for values alive at that point.
5. Handle if/else branches: each branch gets its own destroy list for values declared within it.
6. In AST tree: `AppendToBlock(blockIdx, nodeIdx)` — insert node at end of a block's child list.
7. Write tests: `TestDestroyAtBlockEnd`, `TestDestroyEarlyReturn`, `TestNoDestroyMoved`, `TestDestroyOrder`.

## Test Plan

- `TestDestroyAtBlockEnd`: `{ let x = alloc_heap(); }` → DestroyStmt for x at block end
- `TestDestroyEarlyReturn`: `if cond: return val` → destroys injected before return
- `TestNoDestroyMoved`: `let x = Foo{}; consume(x)` — x moved → no destroy for x
- `TestDestroyOrder`: `let x = A{}; let y = B{}; }` → destroy y before x (LIFO)
- `TestNoDestroyStack`: stack-allocated value → no DestroyStmt
- `TestNoDestroyPrimitive`: `let x: i32 = 5` → no DestroyStmt (primitive)

## Validation Checklist

- [ ] DestroyStmt injected for all non-moved heap values at scope exit
- [ ] LIFO ordering maintained
- [ ] Early returns get destroy injections
- [ ] Moved values skipped
- [ ] Stack values skipped
- [ ] Primitive types skipped

## Acceptance Criteria

- Running the compliance test suite through the C backend shows no memory leaks (verified with valgrind or AddressSanitizer)
- DestroyStmt injection produces valid, non-redundant destroys

## Definition of Done

- [ ] `compiler/sema/ctgc.go` implemented
- [ ] DestroyStmt injected correctly in all test cases
- [ ] Unit tests pass
- [ ] Integrated into compiler pipeline after escape analysis

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Double-free if destroy injected for both outer and inner scope | Only inject at the DECLARING scope; inner blocks don't see outer vars |
| Missing destroy for complex control flow (break/continue) | Handle break/continue same as return — inject destroys before |

## Future Follow-up Tasks

- p06-t06: CTGC alias reuse optimization
- p08-t03: cgen-statements emits ax_free() for DestroyStmt nodes
- p10-t05: opt-ctgc-air does the same analysis at the AIR level
