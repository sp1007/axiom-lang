# p10-t05: CTGC Optimization at AIR Level

## Purpose
Implement CTGC (Compile-Time GC) optimizations at the AIR level: object reuse (OpAliasReuse), region allocation (grouping nearby allocations into a single bulk alloc), and elimination of unnecessary gen_id checks in performance-critical paths.

## Context
While the AST-level CTGC (Phase 06) handles the basic cases, the AIR-level CTGC pass has access to data flow information and can perform more aggressive optimizations. AIR-level analysis can identify reuse patterns that span basic blocks and can batch allocations that are provably alive simultaneously.

## Inputs
- `AirFunc` with OpAlloc, OpFree, OpMakeRef, OpDeref sequences
- Data flow information from the AIR CFG

## Outputs
- Modified `AirFunc` with OpFree+OpAlloc pairs replaced by OpAliasReuse
- Region allocation groups (single OpRegionAlloc replacing multiple OpAlloc)

## Dependencies
- p10-t01: opt-pipeline-manager — implements OptPass (O2+)
- p09-t10: air-builder-ownership — OpAliasReuse already defined
- p09-t04: air-verifier — validates transformations

## Subsystems Affected
- AIR: allocation patterns transformed
- Runtime: fewer ax_alloc/ax_free calls
- Performance: measurably reduces allocator pressure

## Detailed Requirements

1. `CTGCOptPass` implements `OptPass`.
2. **OpFree + OpAlloc reuse**: find pattern `OpFree %ref; ... OpAlloc TypeID` where same TypeID and no intervening use of `%ref` in a reachable path → replace with `OpAliasReuse %ref`.
   - "No intervening use": scan instructions between Free and Alloc; if `%ref` appears as Src operand → not safe to reuse.
3. **Region allocation**: find set of OpAlloc instructions all dominated by the same basic block entry, all with same-type allocations, all provably alive simultaneously → group into single `OpRegionAlloc {count: N, size: sizeof(T)*N}`.
4. **gen_id check elimination**: if a value's AxRef was just created by OpMakeRef in the same basic block (no escape possible) and is immediately dereferenced by OpDeref → eliminate the OpDeref (the gen_id check is redundant, the ref was just created).
5. Return true if any transformation applied.

## Implementation Steps

1. Create `ir/opt/ctgc_opt.go`.
2. Implement free/alloc pattern scanner.
3. Implement region allocation grouper.
4. Implement redundant OpDeref eliminator.
5. Register in O2+ pipeline.
6. Write benchmark: allocation-heavy loop before/after CTGC optimization.

## Test Plan

- `TestFreeAllocReuse`: `OpFree %ref; OpAlloc TypeID` → `OpAliasReuse %ref`
- `TestFreeAllocBlockedByUse`: use of %ref between Free and Alloc → NOT replaced
- `TestGenIdCheckElim`: `OpMakeRef %ptr; OpDeref %ref` in same block → OpDeref removed
- `TestRegionAlloc`: 3 same-type allocs in entry block → single region alloc
- `TestBenchmark`: allocation loop shows < 50% allocator calls after optimization

## Validation Checklist

- [ ] Reuse only when types match
- [ ] No use between Free and reuse AllcOc
- [ ] gen_id check eliminated only when provably safe (same block, no escape)
- [ ] AIR verifier passes after optimization
- [ ] Differential test: O0 == O2 output

## Acceptance Criteria

- A loop with alloc/free inside shows OpAliasReuse in optimized AIR
- Allocator pressure reduced by 50%+ on allocation-heavy benchmarks

## Definition of Done

- [ ] `ir/opt/ctgc_opt.go` implemented
- [ ] Registered in O2+ pipeline
- [ ] Unit tests pass
- [ ] Differential test passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Unsafe reuse (type size mismatch) | Strict TypeID equality check |
| gen_id check elimination creates UAF | Only eliminate in provably safe pattern (same block, just created) |

## Future Follow-up Tasks

- p14-t01: axalloc-size-classes makes reuse even cheaper
