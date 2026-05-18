# p10-t03: Dead Code Elimination Pass

## Purpose
Implement the Dead Code Elimination (DCE) optimization pass that removes instructions whose results are never used and eliminates unreachable basic blocks.

## Context
DCE runs after constant folding (which leaves NOP'd instructions) and after inlining (which may leave unreachable blocks). Together, constant folding + DCE is the core of AXIOM's O1 optimization. Two forms: instruction-level DCE (remove unused computations) and block-level DCE (remove unreachable blocks via CFG analysis). Early DCE also leverages Lazy Field Analysis to eliminate entire unreachable functions.

## Inputs
- `AirFunc` with potentially dead instructions and unreachable blocks
- `ReachableSet` from Lazy Field Analysis (p04-t03) for early DCE

## Outputs
- Modified `AirFunc` with dead instructions marked as NOP, unreachable blocks removed

## Dependencies
- p10-t01: opt-pipeline-manager — implements OptPass
- p09-t02: air-basic-blocks — CFG structure for block DCE
- p04-t03: lazy-field-analysis — for function-level DCE

## Subsystems Affected
- AIR: instructions and blocks removed
- Code size: fewer instructions = smaller output

## Detailed Requirements

1. `DCEPass` implements `OptPass`.
2. **Instruction-level DCE**: for each instruction, if Dest register has no uses (scan all Src1/Src2 across all instructions), mark as NOP. Repeat until fixpoint.
   - Exception: instructions with side effects (OpStore, OpFree, OpCall, OpReturn, OpBranch, OpJump, OpSpawn, OpSend) are never DCE'd.
3. **Block-level DCE**: BFS from entry block; any block not reachable from entry → remove.
   - Update Preds/Succs of adjacent blocks.
   - Remove instructions in dead blocks.
4. **Function-level DCE**: functions not in LFA ReachableSet → remove from AirModule entirely.
5. After removing dead blocks: remove any phi operands referencing removed blocks.
6. Return true if any change made.

## Implementation Steps

1. Create `ir/opt/dce.go`.
2. Implement `buildUseCount(func *AirFunc) map[uint32]int` — count uses of each register.
3. Mark instructions with use_count[dest] == 0 and no side effects as NOP.
4. BFS from entry block, collect reachable set.
5. Remove unreachable blocks (not in reachable set).
6. Fix up phi nodes after block removal.
7. Remove unreachable functions from AirModule using LFA set.
8. Write tests.

## Test Plan

- `TestDCEUnusedInstruction`: `%r = iadd %a, %b` where %r never used → NOP'd
- `TestDCESideEffect`: `store %addr, %val` — never DCE'd even if result unused
- `TestDCEDeadBlock`: unreachable block (no predecessor) → removed
- `TestDCEDeadFunction`: function not reachable from main → removed from module
- `TestDCEPhiCleanup`: phi with operand from removed block → phi updated
- `TestDCEChained`: dead chain: A defines B, B defines C, C unused → all three NOP'd

## Validation Checklist

- [ ] Pure instructions with unused results are NOP'd
- [ ] Side-effecting instructions never removed
- [ ] Unreachable blocks removed with correct edge cleanup
- [ ] Phi nodes updated after block removal
- [ ] Unreachable functions removed from module
- [ ] AIR verifier passes after DCE

## Acceptance Criteria

- A function with `let x = expensive_pure_computation(); return 0` — x computation removed at O1
- All compliance tests still pass after DCE

## Definition of Done

- [ ] `ir/opt/dce.go` implemented
- [ ] Registered in OptPipeline O1+
- [ ] Unit tests pass
- [ ] Differential test: O0 == O1 output for all programs

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Removing side-effectful instruction (false positive) | Build exhaustive list of side-effecting opcodes; default to NOT removing |
| Breaking SSA after block removal | Only remove entire blocks; fix phi operands carefully |

## Future Follow-up Tasks

- p10-t04: inlining creates new DCE opportunities (run DCE after inlining)
