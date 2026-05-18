# p09-t10: AIR Builder — Ownership Operations

## Purpose
Lower ownership-specific operations from the typed AST to AIR: move semantics (OpMove), CTGC destroy (OpDestroy), generational reference creation (OpMakeRef), generational dereference with runtime check (OpDeref), and arena allocation (OpArenaAlloc).

## Context
Ownership operations are the unique AIR instructions that implement AXIOM's memory safety model. They have no analogue in LLVM IR or C IR. Getting these correct is critical — a wrong OpMakeRef or missing OpDestroy directly causes memory unsafety. The C-backend translates these to the runtime API (ax_alloc, ax_free, ax_make_ref, ax_deref).

## Inputs
- Ownership-analyzed AST: DestroyStmt, AliasStmt, MakeRef, Deref nodes
- AirFuncBuilder
- EscapesToHeap flags on VarDecl nodes

## Outputs
- `ir/air/builder/ownership.go` — ownership instruction lowering

## Dependencies
- p09-t07: air-builder-statements — statements call this for alloc/destroy
- p06-t05: ctgc-destroy-injection — DestroyStmt nodes in AST
- p07-t02: generational-ref-runtime — the runtime API these lower to

## Subsystems Affected
- AIR: OpMakeRef, OpDeref, OpDestroy, OpMove, OpArenaAlloc
- C-backend: these map to ax_make_ref, ax_deref, ax_free, ax_arena_alloc
- Memory safety: incorrect lowering = UAF at runtime

## Detailed Requirements

1. `lowerOwnershipOp(nodeIdx uint32) uint32` — called from lowerStmt/lowerExpr.
2. Heap allocation: when VarDecl has `EscapesToHeap`:
   ```
   %size = iconst sizeof(T)
   %raw_ptr = OpAlloc TypeID          ; calls ax_alloc
   %ref = OpMakeRef %raw_ptr          ; creates AxRef{ptr, gen_id}
   varMap[symID] = %ref               ; future uses go through OpDeref
   ```
3. Dereference (for heap vars): before each use of a heap var:
   ```
   %real_ptr = OpDeref %ref           ; calls ax_deref (gen_id check)
   %val = OpLoad %real_ptr, offset    ; then load
   ```
4. Move: `OpMove %dst, %src` — copies the AxRef value; source register invalidated by verifier.
5. Destroy (DestroyStmt): `OpFree %ref` (actually `OpDestroy %ref`) — calls ax_free on the underlying pointer (increments gen_id).
6. AliasStmt (reuse): `OpAliasReuse %new_ref, %old_ref` — reuse old pointer with incremented gen_id.
7. Arena alloc: when VarDecl has `FlagUsesArena`:
   ```
   %ptr = OpArenaAlloc %arena_ref, TypeID
   varMap[symID] = %ptr               ; no AxRef wrapper for arena allocs
   ```
8. No gen_id checks for arena-allocated vars (FlagUsesArena set) or unsafe-block vars.

## Implementation Steps

1. Create `ir/air/builder/ownership.go`.
2. Implement heap alloc sequence in VarDecl lowering.
3. Implement `emitDeref(refReg uint32) uint32` helper — emits OpDeref, returns raw ptr.
4. Update `lowerExpr` for Ident: if heap var, call `emitDeref` before load.
5. Implement `lowerDestroyStmt` → OpFree.
6. Implement `lowerAliasStmt` → OpAliasReuse.
7. Implement `lowerArenaAlloc` → OpArenaAlloc.
8. Write unit tests for each path.

## Test Plan

- `TestHeapAlloc`: heap VarDecl → OpAlloc + OpMakeRef sequence
- `TestStackAlloc`: stack VarDecl → no OpAlloc (just virtual slot)
- `TestDerefOnHeapRead`: reading heap var → OpDeref before OpLoad
- `TestDestroyInstr`: DestroyStmt → OpFree
- `TestAliasReuse`: AliasStmt → OpAliasReuse
- `TestArenaAlloc`: UsesArena var → OpArenaAlloc (no OpMakeRef)
- `TestCompliance031040`: compliance tests 031-040 (ownership group) produce correct AIR

## Validation Checklist

- [ ] Heap vars always accessed via OpDeref
- [ ] Arena vars never use OpMakeRef/OpDeref
- [ ] OpDestroy only on heap vars (not stack or arena)
- [ ] OpMove invalidates source (verifier enforces)
- [ ] AIR verifier passes on all ownership operations

## Acceptance Criteria

- Compliance tests 031-040 compile and run correctly (no runtime UAF with AddressSanitizer)
- Generated C code calls ax_deref before every heap variable access

## Definition of Done

- [ ] `ir/air/builder/ownership.go` implemented
- [ ] All ownership operations produce correct AIR
- [ ] Unit tests pass
- [ ] AIR verifier passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Missing OpDeref for heap var → UAF not caught | Build check: every load/store from a heap var must have OpDeref predecessor |
| OpDeref overhead in tight loops | Arena blocks and unsafe blocks bypass; use for safe code only |

## Future Follow-up Tasks

- p10-t05: opt-ctgc-air optimizes ownership operations at AIR level
- p08-t06: cgen-generational-checks lowered from these AIR instructions
