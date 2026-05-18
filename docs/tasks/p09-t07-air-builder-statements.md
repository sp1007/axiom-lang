# p09-t07: AIR Builder — Statements

## Purpose
Implement the AIR builder for statements — the component that lowers AXIOM statement AST nodes to AIR instructions, including variable declarations, assignments, returns, defers, and CTGC destroy/alias nodes.

## Context
Statement lowering builds on expression lowering (p09-t06). The key statement operations are: allocating storage for local variables (OpAlloc for heap, stack slots for stack vars), storing expression results, handling defer accumulation at function exit, and emitting the CTGC-injected destroy and alias nodes.

## Inputs
- Typed, ownership-analyzed AST with DestroyStmt and AliasStmt nodes injected by CTGC
- AirFuncBuilder (p09-t02)
- `lowerExpr` from p09-t06

## Outputs
- `ir/air/builder/stmt.go` — `lowerStmt(nodeIdx uint32)` function

## Dependencies
- p09-t06: air-builder-expressions — `lowerExpr` used for initializers and RHS
- p06-t05: ctgc-destroy-injection — DestroyStmt nodes in AST
- p06-t06: ctgc-alias-reuse — AliasStmt nodes in AST

## Subsystems Affected
- AIR builder: statements drive the top-level lowering
- Memory management: alloc/free/alias instructions emitted here

## Detailed Requirements

1. `lowerStmt(nodeIdx uint32)` — no return value (statements are void).
2. Statement lowering by NodeKind:
   - `VarDecl (let x: T = expr)`:
     - If `EscapesToHeap`: emit `%addr = OpAlloc TypeID`; emit `%val = lowerExpr(initExpr)`; emit `OpStore %addr, %val`; map symID → %addr in builder's varMap
     - If stack: reserve virtual slot `%addr = allocaSlot(TypeID)`; same pattern
   - `VarDecl (mut x := expr)`: same as let, but with IsMut flag noted in varMap
   - `AssignStmt (x = expr)`: look up %addr from varMap; emit `%val = lowerExpr(expr)`; emit `OpStore %addr, %val`
   - `ReturnStmt`: emit deferred statements (reverse order), emit `OpReturn %val` or `OpReturn 0`
   - `DeferStmt`: accumulate deferred expressions in a stack (not emitted yet); emitted at each return site
   - `DestroyStmt`: emit `OpFree %ptr` (or `OpDestroy %ptr`)
   - `AliasStmt (=alias(x, y))`: emit `OpAliasReuse %x_ptr → %y_ptr`; update varMap: y → x's register
   - `SpawnExpr as stmt`: lower the spawn expression, discard result
3. `varMap map[uint32]uint32` (SymID → virtual register holding the address).
4. Defer: use a `[]uint32` stack of deferred node indices; pop and lower at each return.
5. For params: at function entry, emit `param_N` pseudo-instructions and store to local slot.

## Implementation Steps

1. Create `ir/air/builder/stmt.go`.
2. Implement `lowerStmt()` dispatch on NodeKind.
3. Implement VarDecl with heap/stack branching on EscapesToHeap flag.
4. Implement defer accumulation and flush-at-return.
5. Implement DestroyStmt → OpFree.
6. Implement AliasStmt → OpAliasReuse.
7. Write unit tests.

## Test Plan

- `TestLowerVarDecl`: `let x: i32 = 5` → OpAlloc or stack slot + OpStore
- `TestLowerAssign`: `x = 10` → OpStore to x's address
- `TestLowerReturn`: `return x` → load x + OpReturn
- `TestLowerDefer`: `defer f()` — f() emitted before return
- `TestLowerDestroy`: DestroyStmt → OpFree emitted
- `TestLowerAlias`: AliasStmt → OpAliasReuse emitted

## Validation Checklist

- [ ] Heap vars allocated with OpAlloc
- [ ] Stack vars use virtual slot (no OpAlloc)
- [ ] Deferred statements executed in reverse order
- [ ] DestroyStmt → OpFree
- [ ] AliasStmt → OpAliasReuse
- [ ] varMap correctly tracks symbol → register mapping

## Acceptance Criteria

- AIR for a function with let/assign/return produces correct OpAlloc/OpStore/OpLoad/OpReturn sequence
- AIR verifier produces 0 errors on statement-lowered code

## Definition of Done

- [ ] `ir/air/builder/stmt.go` implemented
- [ ] Unit tests pass
- [ ] Verifier passes on all lowered statements

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Defer in loop body — each iteration defers | Deferred stmts are flushed at every return/break/continue |
| Multiple return sites with different defer states | Track defer stack at each return site independently |

## Future Follow-up Tasks

- p09-t08: air-builder-control-flow handles if/for/while which call lowerStmt
