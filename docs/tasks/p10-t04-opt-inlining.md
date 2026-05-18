# p10-t04: Function Inlining Pass

## Purpose
Implement function inlining — replacing a function call with the body of the called function — to eliminate call overhead and enable further optimizations (constant folding across call boundaries, better register allocation).

## Context
Inlining is one of the most impactful optimizations. Inlining `abs(x)` eliminates a function call and enables constant folding if x is constant. AXIOM's inliner uses a cost heuristic: inline if the callee is small (< 30 AIR instructions) and the call site benefits (inner loop or only one caller). Guard against infinite recursion.

## Inputs
- `AirModule` — call graph accessible via function SymIDs
- Inlining heuristics configuration (inline threshold: 30 instructions)

## Outputs
- Modified `AirFunc` with call sites replaced by inlined instruction sequences
- New virtual registers and blocks created from cloned callee

## Dependencies
- p10-t01: opt-pipeline-manager — implements OptPass (O2+)
- p09-t02: air-basic-blocks — AirFuncBuilder used for inlining
- p09-t01: air-instruction-set — OpCall being replaced

## Subsystems Affected
- AIR: calls replaced with inlined bodies
- Register allocation: inlined code shares register space with caller

## Detailed Requirements

1. `InliningPass` implements `OptPass`.
2. For each function, scan for `OpCall` instructions.
3. For each call site: check inlining eligibility:
   - Callee has ≤ 30 AIR instructions (configurable)
   - Callee is not recursive (no call to itself in callee body)
   - Callee is not `extern` or variadic
   - Callee is reachable (not a function pointer — known at compile time)
4. Inline procedure:
   - Clone callee's `AirInst` slice (fresh registers: offset all VReg IDs by caller's max VReg)
   - Map callee's param registers to caller's arg registers (add OpMove instructions for each param)
   - Replace callee's `OpReturn %val` with `OpJump continue_block` (and store return value in a fresh register)
   - Replace `OpCall` in caller with: jump to inlined body's entry, then continue from `continue_block`
5. Register remapping: `newReg = cloneReg + caller_max_reg_offset`.
6. Return true if any inlining occurred.
7. After inlining, run constant folding and DCE for maximum benefit.

## Implementation Steps

1. Create `ir/opt/inline.go`.
2. Implement `costOf(callee *AirFunc) int` — count non-NOP instructions.
3. Implement `isInlineable(callee *AirFunc) bool`.
4. Implement `inlineCall(caller *AirFunc, callInstIdx uint32, callee *AirFunc)`.
5. Clone callee instructions with remapped registers.
6. Replace OpCall with inlined code.
7. Write unit tests.

## Test Plan

- `TestInlineSmallFunc`: `fn identity(x: i32) -> i32: return x` → inlined at call site
- `TestNoInlineRecursive`: `fn fib(n: i32) -> i32` — recursive, not inlined
- `TestNoInlineExtern`: `extern "C" fn printf(...)` — extern, not inlined
- `TestNoInlineLarge`: function with > 30 instructions → not inlined
- `TestInlineThenFold`: after inlining `abs(-5)`, constant fold → `iconst 5`

## Validation Checklist

- [ ] Small pure functions inlined
- [ ] Recursive functions not inlined
- [ ] Extern functions not inlined
- [ ] Register remapping correct (no register collision)
- [ ] AIR verifier passes after inlining
- [ ] Returns true only when inlining occurred

## Acceptance Criteria

- `abs(x)` (3-instruction function) inlined at all call sites
- No infinite loop when inlining function that calls abs() which calls nothing

## Definition of Done

- [ ] `ir/opt/inline.go` implemented
- [ ] Registered in OptPipeline O2+
- [ ] Unit tests pass
- [ ] Differential test: O0 == O2 output for all programs

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Code size explosion (too much inlining) | Strict threshold (30 insts); configurable with --inline-threshold |
| Incorrect register remapping causing SSA violation | AIR verifier catches this immediately after inlining |

## Future Follow-up Tasks

- p10-t02: constant folding benefits greatly from inlining
- p10-t03: DCE removes any dead code exposed after inlining
