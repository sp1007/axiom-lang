# p10-t10: C-Backend v2 — Lowering from AIR

## Purpose
Rewrite the C-backend to lower from optimized AIR instead of directly from the TypedAST. This enables the optimization pipeline's improvements to reach the final binary and produces better C code by operating on a simpler, more regular IR.

## Context
The Phase 08 C-backend lowered directly from the TypedAST, which was necessary to get a working compiler quickly. Now that AIR is stable and optimized, the C-backend should consume AIR for cleaner code and to benefit from optimization passes. The AIR→C mapping is more regular than AST→C and easier to maintain.

## Inputs
- Optimized `AirModule` (after O0/O1/O2/O3 passes)
- TypeTable for C type name generation
- InternPool for name resolution

## Outputs
- `codegen/cgen/air_cgen.go` — AIR→C code generator (replaces or augments `codegen/cgen/`)
- Valid C11 source file

## Dependencies
- p10-t01: opt-pipeline-manager — AIR arrives post-optimization
- p09-t01: air-instruction-set — AIR opcodes being lowered
- p08-t01: cgen-type-mapping — C type names (reuse existing)

## Subsystems Affected
- C-backend: major rewrite from AST-based to AIR-based
- Build pipeline: optimization now feeds into codegen
- Test: all existing C-backend tests must pass with new backend

## Detailed Requirements

1. `AirCGen` struct: `tt *TypeTable, pool *InternPool, buf *bytes.Buffer`
2. `Generate(module *AirModule) string` — returns C source string.
3. AIR → C instruction mapping:
   - `%r: i32 = iconst 42` → `ax_i32 r42 = 42;`
   - `%r: i32 = iadd %a, %b` → `ax_i32 r_N = r_a + r_b;`
   - `%r = load %addr, 0` → `ax_i32 r_N = *((ax_i32*)r_addr);`
   - `store %addr, %val` → `*((ax_T*)r_addr) = r_val;`
   - `%r = alloc TypeID` → `ax_T* r_N = (ax_T*)ax_alloc(sizeof(ax_T));`
   - `free %ref` → `ax_free(r_ref.ptr);`
   - `%r = makeref %ptr` → `AxRef r_N = ax_make_ref(r_ptr);`
   - `%r = deref %ref` → `void* r_N = ax_deref(r_ref);`
   - `%r = call @sym, %a1, %a2` → `ax_T r_N = _AX_module_sym(r_a1, r_a2);`
   - `return %val` → `return r_val;`
   - `jump block_3` → `goto block_3;`
   - `branch %cond, block_2, block_3` → `if (r_cond) goto block_2; else goto block_3;`
   - Block labels: `block_N: ;`
4. Phi nodes: insert copies at predecessor blocks before jump to merge block.
5. Register naming: `r_{N}` where N is the virtual register index.
6. Function header: `static ax_T _AX_module_funcname(ax_T r_param0, ax_T r_param1) {`
7. Generate forward declarations for all functions first, then bodies.

## Implementation Steps

1. Create `codegen/cgen/air_cgen.go`.
2. Implement `Generate()` main loop: forward decls, then function bodies.
3. For each function: emit function header, then iterate blocks in block order.
4. For each instruction in a block: call `emitInst(inst)`.
5. Implement phi lowering: insert `r_phi = r_incoming;` in predecessor blocks.
6. Update `cmd/axc/build.go` to run optimization pipeline then call AirCGen.
7. Run all existing C-backend compliance tests (001-070) with the new backend.

## Test Plan

- `TestAirCGenHello`: generate C for hello world AIR → valid C that compiles with gcc
- `TestAirCGenFibonacci`: recursive fibonacci → correct C with recursive call
- `TestAirCGenAlloc`: heap allocation → `ax_alloc` + `ax_make_ref` in output
- `TestAirCGenOptimized`: optimized AIR (constant-folded) → simplified C output
- `TestDifferentialO0O2`: O0 and O2 compiled programs produce same runtime output

## Validation Checklist

- [ ] Generated C compiles with `gcc -Wall -Wextra` with no warnings
- [ ] All 100 compliance tests pass with new backend
- [ ] Phi nodes correctly lowered to copies
- [ ] Block labels emitted before each block's instructions
- [ ] Forward declarations prevent use-before-define errors in C

## Acceptance Criteria

- All 100 compliance tests pass using AIR C-backend
- Generated C is simpler and shorter than AST C-backend output for the same program

## Definition of Done

- [ ] `codegen/cgen/air_cgen.go` implemented
- [ ] Integrated into build pipeline
- [ ] All compliance tests pass
- [ ] Differential test passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| C goto-based control flow rejected by some compilers | Use `goto` which is standard C11; test with gcc and clang |
| Variable shadowing in goto-based C | Assign all variables at function top (declare before blocks) |

## Future Follow-up Tasks

- p11-t15: native-backend-integration replaces C-backend for production builds
- p10-t11: differential tests validate correctness
