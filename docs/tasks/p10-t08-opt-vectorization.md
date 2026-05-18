# p10-t08: Auto-Vectorization (SIMD)

## Purpose
Implement auto-vectorization for simple loop patterns that operate on arrays of f32/f64, replacing scalar operations with SIMD instructions (x86 AVX2 8-wide floats or SSE 4-wide).

## Context
Vectorization can give 4-8× speedup for numerical code. AXIOM's vectorizer targets the most common pattern: a loop that processes elements of an array independently (no cross-iteration dependencies). The vectorizer emits `OpSIMD` AIR instructions that the native backend translates to `_mm256_*` AVX2 intrinsics or falls back to scalar.

## Inputs
- `AirFunc` with identified loops (LoopInfo from p10-t07)
- Array/slice access patterns within the loop body

## Outputs
- Modified `AirFunc` with OpSIMD instructions replacing scalar loop body
- Scalar tail loop for non-multiple-of-8 elements

## Dependencies
- p10-t07: opt-loop-region — loop identification required first
- p10-t01: opt-pipeline-manager — implements OptPass (O3)
- p09-t01: air-instruction-set — OpSIMD* opcodes

## Subsystems Affected
- AIR: scalar ops replaced with SIMD ops
- Native backend (p11-t08): translates OpSIMD to AVX2 intrinsics
- Performance: 4-8× speedup on vectorizable loops

## Detailed Requirements

1. `VectorizationPass` implements `OptPass`.
2. Target pattern: simple countable loop over arrays (for i in 0..n):
   - Loop body has no memory-carried dependencies (element i does not read element i-1)
   - All operations are on f32 or f64 arrays
   - Loop body is a sequence of element-wise ops (add, mul, sub, div, neg)
3. Legality check:
   - Verify no loop-carried dependencies: for each array store, verify no array load with potentially aliasing address in the same iteration
   - Verify element type is f32 or f64 (not integers — overflow semantics differ)
   - Verify loop is countable (known bounds at compile time or runtime count variable)
4. Transformation: replace N scalar iterations with N/8 SIMD iterations + remainder:
   ```
   for i in 0..(n / 8):           ; SIMD loop
       %v = simd_load arr + i*8   ; 8 floats
       %result = simd_add %v, %v2
       simd_store result + i*8, %result
   for i in (n/8)*8..n:           ; scalar tail
       ... original scalar body ...
   ```
5. `OpSIMDLoad{width:8, TypeID:TypeF32}`, `OpSIMDAdd{width:8}`, `OpSIMDStore{width:8}`.
6. Guard: `if target.SupportsAVX2 { vectorize } else { scalar }`.

## Implementation Steps

1. Create `ir/opt/vectorize.go`.
2. Implement loop pattern recognizer.
3. Implement dependency analysis (conservative: any aliasing → don't vectorize).
4. Implement SIMD loop generation with scalar tail.
5. Add target feature check (AVX2 available).
6. Write tests.

## Test Plan

- `TestVectorizeAddArrays`: `for i in 0..n: c[i] = a[i] + b[i]` → vectorized with OpSIMDAdd
- `TestNoVectorizeDependence`: `for i in 1..n: a[i] = a[i-1] + 1` → NOT vectorized (dependency)
- `TestNoVectorizeInteger`: `for i in 0..n: c[i] = a[i] + b[i]` with i32 → NOT vectorized
- `TestVectorizeTail`: n=10, SIMD loop for i=0..8, scalar for i=8..10
- `TestVectorizeWithFMA`: `a[i] + b[i] * c[i]` → `simd_fma` (fused multiply-add)

## Validation Checklist

- [ ] Simple element-wise float loops vectorized
- [ ] Loop-carried dependency prevents vectorization
- [ ] Scalar tail correctly handles non-multiple-of-8 counts
- [ ] AVX2 feature gate respected
- [ ] AIR verifier passes on vectorized code

## Acceptance Criteria

- A simple array addition loop runs 4-8× faster after vectorization (measured with benchmark)
- Non-vectorizable loops fall back to scalar correctly

## Definition of Done

- [ ] `ir/opt/vectorize.go` implemented
- [ ] Registered in O3 pipeline
- [ ] Unit tests pass
- [ ] Differential test: O0 == O3 output for all programs

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Vectorized results differ from scalar (FP ordering) | Document that O3 may reorder float ops; match IEEE 754 behavior |
| AVX2 not available on all CI machines | Test under `GOARCH=amd64 GOFLAGS=-v` with AVX2 detection |

## Future Follow-up Tasks

- p11-t02: x86-instruction-set includes AVX2 instruction encodings
- p11-t03: x86-instruction-selector lowers OpSIMD to actual AVX2
