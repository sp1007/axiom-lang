# p06-t06: CTGC Alias Reuse (Object Reuse Optimization)

## Purpose
Implement the CTGC alias reuse optimization: when a heap allocation immediately follows a destroy of the same type, reuse the existing memory rather than free+malloc. This eliminates allocator round-trips in common patterns like loop-body allocation, reducing allocation overhead to near-zero.

## Context
The pattern `let x = alloc(); ... =destroy(x); let y = alloc_same_type()` is common in loops. Instead of actually freeing x and then allocating y fresh, the compiler can keep x's memory and "rename" it to y, incrementing the generational ID to invalidate old references to x. This gives C++ placement-new-like performance without manual management.

## Inputs
- AST after CTGC destroy injection (p06-t05) — DestroyStmt nodes present
- ConnectionGraph — to verify no outstanding borrows of the destroyed value
- TypeTable — to verify same type

## Outputs
- New `NodeKind.AliasStmt` nodes that replace `DestroyStmt` + subsequent `VarDecl`
- Updated ConnectionGraph: `ReusedBy` edge from destroyed node to reuse node

## Dependencies
- p06-t05: ctgc-destroy-injection — DestroyStmt nodes must exist
- p06-t01: connection-graph — ReusedBy edge kind

## Subsystems Affected
- AST: AliasStmt replaces destroy+alloc pattern
- Code generation: AliasStmt → reuse pointer + increment gen_id
- Performance: eliminates allocator calls in loops

## Detailed Requirements

1. `AliasReuse` pass runs after CTGC destroy injection.
2. Pattern detection: for each `DestroyStmt{target: x}` followed by `VarDecl{y, typeID=T, EscapesToHeap}` where typeID of x == typeID of y:
   - Verify: no outstanding borrows of x at that point (check ConnectionGraph — no active Borrows edges)
   - If safe: replace with `AliasStmt{from: x, to: y}` (reuse x's memory for y)
3. `NodeKind.AliasStmt` contains `FromSym, ToSym uint32`.
4. C-backend for AliasStmt: `y_ptr = x_ptr; ((AxHeader*)y_ptr - 1)->gen_id++;`
5. This is O(1) vs O(malloc + free).
6. `ReusedBy` edge added to ConnectionGraph: from x_node to y_node.
7. Only apply when types are identical (same TypeID, same size).
8. Only apply when the destroy is unconditional (not inside an if branch that might not execute).

## Implementation Steps

1. Create `compiler/sema/alias_reuse.go`.
2. Add `NodeKind.AliasStmt` to `ast/node.go`.
3. Implement pattern scanner: walk block children looking for Destroy followed by VarDecl.
4. Check type match and borrow safety.
5. Replace matched pairs with AliasStmt node.
6. Add `ReusedBy` edge to ConnectionGraph.
7. In cgen: handle AliasStmt — emit pointer copy + gen_id increment.
8. Write tests: `TestAliasReuseBasic`, `TestAliasReuseTypeMismatch`, `TestAliasReuseWithBorrow`.

## Test Plan

- `TestAliasReuseBasic`: `=destroy(x); let y = alloc_Foo()` → replaced with `=alias(x, y)`
- `TestAliasReuseTypeMismatch`: destroy Foo, alloc Bar → NOT replaced (different types)
- `TestAliasReuseWithBorrow`: outstanding borrow of x when destroyed → NOT replaced (unsafe)
- `TestAliasReuseLoop`: loop with alloc/destroy pattern → reuse on every iteration after first
- `TestAliasReuseGenIDIncrement`: verify gen_id is incremented (old refs become invalid)

## Validation Checklist

- [ ] Type match required (same TypeID)
- [ ] No outstanding borrows required
- [ ] gen_id incremented in C output
- [ ] Old references to x invalidated after alias
- [ ] Only unconditional destroys reused

## Acceptance Criteria

- A loop allocating/freeing a value 1M times shows 0 malloc calls after first iteration (measured with malloc interposer)
- Generational reference check still works after alias reuse

## Definition of Done

- [ ] `compiler/sema/alias_reuse.go` implemented
- [ ] AliasStmt node kind added
- [ ] C-backend handles AliasStmt
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Conditional destroy (inside if) — reuse not safe | Only reuse unconditional destroys (not inside branches) |
| Type layout differs despite same TypeID | Same TypeID guarantees same layout; no risk |

## Future Follow-up Tasks

- p10-t05: opt-ctgc-air implements the same optimization at AIR level (more accurate)
- p08-t05: cgen-ownership emits the alias reuse pattern
