# p10-t09: SoA Transform (AI-Annotated Layout Optimization)

## Purpose
Implement the Structure of Arrays (SoA) layout transformation triggered by the `@[ai::suggest_layout(layout="SoA")]` annotation, converting array-of-struct field access to separate per-field arrays for improved cache performance.

## Context
Array of Structs (AoS): `[{x:f32, y:f32, z:f32}]` — memory layout: xyzxyzxyz. Structure of Arrays (SoA): `{xs:[f32], ys:[f32], zs:[f32]}` — memory layout: xxxyyy zzz. For SIMD processing (vectorized loops that access one field at a time), SoA is 3× more cache-efficient. This transform is triggered by AI annotation and confirmed by escape analysis.

## Inputs
- `AirModule` with `@[ai::suggest_layout(layout="SoA")]` annotation on a struct
- Escape analysis confirming the annotated struct is only accessed in loops
- TypeTable for struct field information

## Outputs
- New TypeInfo for the SoA layout struct
- Modified AIR: field access patterns updated to use per-field arrays
- `ir/opt/soa_transform.go`

## Dependencies
- p10-t07: opt-loop-region — SoA only applied to loop-accessed structs
- p10-t01: opt-pipeline-manager — implements OptPass (O3)
- p06-t04: escape-analysis — escape info determines applicability

## Subsystems Affected
- Type system: new SoA type registered
- AIR: field access instructions transformed
- Memory layout: struct layout changed globally for the annotated type

## Detailed Requirements

1. `SoATransformPass` implements `OptPass`.
2. Trigger: struct has `AIHint{Kind:SuggestSoA}` in MetaTable.
3. Eligibility check:
   - All uses of the struct type are in loops (from LoopInfo)
   - The struct type is only accessed via field reads (no opaque uses)
   - Not passed to extern functions (layout must be known)
4. Transformation:
   - Create new TypeInfo `Foo_SoA` with fields: `xs: [T_x]`, `ys: [T_y]`, `zs: [T_z]`
   - Replace all `OpGEP %foo_arr, .x` patterns with `OpLoad %foo_soa.xs, [i]`
   - Update all allocation sites for `[Foo]` to allocate per-field arrays instead
5. Output `OptReport` summarizing which structs were transformed.
6. `--soa-report` flag: print which structs were SoA-transformed.

## Implementation Steps

1. Create `ir/opt/soa_transform.go`.
2. Find annotated structs via MetaTable AIHints scan.
3. Check eligibility.
4. Create SoA TypeInfo in TypeTable.
5. Rewrite all AIR field accesses.
6. Write tests.

## Test Plan

- `TestSoABasic`: `@[ai::suggest_layout(layout="SoA")] struct Vec3` in loop → transformed
- `TestSoAEligibility`: Vec3 used outside loop → NOT transformed
- `TestSoAFieldAccess`: `arr[i].x` → `arr_soa.xs[i]` in AIR
- `TestSoAPerf`: benchmark shows improved cache hit rate (measured with perf stat)

## Validation Checklist

- [ ] SoA transformation only applied to annotated structs
- [ ] Eligibility check prevents unsafe transformation
- [ ] All field accesses correctly rewritten
- [ ] AIR verifier passes after transformation
- [ ] Differential test: O0 == O3 output

## Acceptance Criteria

- `@[ai::suggest_layout(layout="SoA")]` on a particle struct leads to SoA layout in AIR
- Performance: 20%+ improvement on particle simulation benchmark

## Definition of Done

- [ ] `ir/opt/soa_transform.go` implemented
- [ ] Registered in O3 pipeline
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Incorrect field access rewriting causes UAF | Exhaustive test of all field access patterns; AIR verifier catches wrong types |
| SoA incompatible with FFI (C code expects AoS) | Check: structs passed to extern functions cannot be SoA-transformed |

## Future Follow-up Tasks

- p16-t16: std.compiler.ai provides the `@[ai::suggest_layout]` annotation runtime
