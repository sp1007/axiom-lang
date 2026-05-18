# p10-t02: Constant Folding Pass

## Purpose
Implement the constant folding optimization pass that evaluates constant expressions at compile time, replacing runtime arithmetic with precomputed values.

## Context
Constant folding is the simplest and most universally beneficial optimization. `iconst 2 + iconst 3` becomes `iconst 5` with zero runtime cost. In AXIOM, constant folding also handles `#run` compile-time expressions (via the comptime interpreter in p10-t06) and folding through function inlining results. This pass is in every optimization tier (O1+).

## Inputs
- `AirFunc` with OpIConst/OpFConst followed by arithmetic instructions
- TypeTable for type checking fold results

## Outputs
- Modified AirFunc with constant arithmetic folded
- Dead instructions (original arithmetic ops) marked for DCE

## Dependencies
- p10-t01: opt-pipeline-manager — implements OptPass interface
- p09-t01: air-instruction-set — opcodes being folded

## Subsystems Affected
- AIR: instructions replaced/removed
- DCE (p10-t03): works together with constant folding (fold first, DCE removes dead)

## Detailed Requirements

1. `ConstantFoldingPass` implements `OptPass`.
2. For each function: scan all instructions, detect foldable patterns.
3. Foldable patterns (both operands are OpIConst/OpFConst):
   - `OpIAdd`: fold `iconst a + iconst b → iconst (a+b)`
   - `OpISub`, `OpIMul`, `OpIMod`: same pattern
   - `OpIDiv`: fold, but guard against division by zero (emit diagnostic, don't fold if divisor = 0)
   - `OpIPow`: fold integer power
   - `OpFAdd`, `OpFSub`, `OpFMul`, `OpFDiv`, `OpFPow`: fold floats (IEEE 754)
   - `OpICmpEq`, `OpICmpNe`, `OpICmpLt`, etc.: fold comparisons → `iconst 1` or `iconst 0`
   - `OpBAnd`, `OpBOr`, `OpBXor`, `OpBShl`, `OpBShr`: fold bitwise ops
   - `OpBNot`: fold boolean not
   - `OpNeg`: fold negation
4. Replace the arithmetic instruction with a new `OpIConst/OpFConst` with the folded value.
5. Mark the old arithmetic instruction as NOP (opcode = `OpNop`).
6. Return `true` if any folds were made.
7. Handle integer overflow: wrapping arithmetic (don't fold away overflow — it's defined behavior in AXIOM).
8. Handle `#run`: when `OpComptime` subgraph is all-constant, evaluate with comptime interpreter (p10-t06), replace with result.

## Implementation Steps

1. Create `ir/opt/constant_fold.go`.
2. Implement constant value tracking: `type ConstValue struct{IsConst bool; IVal int64; FVal float64}`.
3. Implement `evalConst(inst AirInst, vals map[uint32]ConstValue) ConstValue` for each arithmetic opcode.
4. Pass over all instructions: track which registers are known constants.
5. When arithmetic instruction has all-constant operands: fold → emit OpIConst, NOP original.
6. Return true if any change made.
7. Write golden test: `arithmetic_const.ax` folds to single constant.

## Test Plan

- `TestFoldAddInts`: `iconst 2 + iconst 3` → `iconst 5`
- `TestFoldDivByZero`: `iconst 5 / iconst 0` → diagnostic, no fold
- `TestFoldComparison`: `iconst 5 == iconst 5` → `iconst 1` (true)
- `TestFoldBitwise`: `iconst 0b1010 & iconst 0b1100` → `iconst 0b1000`
- `TestFoldFloat`: `fconst 1.5 + fconst 2.5` → `fconst 4.0`
- `TestFoldChain`: `1 + 2 + 3` → `iconst 6` (two folds, requires fixpoint)
- `TestNoFoldNonConst`: `x + iconst 3` where x is runtime → not folded

## Validation Checklist

- [ ] All arithmetic op pairs folded correctly
- [ ] Division by zero handled (no fold)
- [ ] Float folding uses IEEE 754 semantics
- [ ] Folded value replaces original, old instruction is NOP'd
- [ ] AIR verifier passes after folding
- [ ] Returns true only when changes made

## Acceptance Criteria

- `#run 2 ** 10` folds to `iconst 1024` at compile time
- All compliance tests still pass after O1 optimization

## Definition of Done

- [ ] `ir/opt/constant_fold.go` implemented
- [ ] Registered in OptPipeline O1+
- [ ] Unit tests pass
- [ ] Differential test: O0 == O1 output for all programs

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Float folding diverges from runtime (different FPU mode) | Use Go's float64 math which matches IEEE 754 |
| Folding through function calls (interprocedural) | Defer to comptime interpreter; only fold within single function |

## Future Follow-up Tasks

- p10-t06: comptime-interpreter handles #run expression folding
- p10-t03: DCE removes NOP'd instructions after folding
