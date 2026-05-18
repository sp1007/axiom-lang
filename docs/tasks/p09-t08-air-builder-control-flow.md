# p09-t08: AIR Builder — Control Flow

## Purpose
Implement AIR lowering for control flow statements (if/elif/else, for, while, match), producing correct CFG structure with properly connected basic blocks, phi nodes at join points, and loop region markers.

## Context
Control flow lowering is the most complex part of AIR generation because it creates multiple basic blocks and requires phi nodes at join points. AXIOM's high-level loop regions must be preserved in AIR (via OpLoopBegin/OpLoopEnd markers) for the optimizer's loop detection pass. Correct phi placement is essential for SSA validity.

## Inputs
- Typed AST with IfStmt, ForStmt, WhileStmt, MatchStmt nodes
- AirFuncBuilder (p09-t02) — block creation and edge addition
- `lowerStmt`, `lowerExpr` from p09-t06, p09-t07

## Outputs
- `ir/air/builder/control.go` — control flow lowering

## Dependencies
- p09-t07: air-builder-statements — lowerStmt used for bodies
- p09-t02: air-basic-blocks — NewBlock, AddEdge, SwitchTo

## Subsystems Affected
- CFG structure: all non-trivial control flow creates multiple blocks
- SSA: phi nodes at merge points
- Loop optimizer (p10-t07): reads loop region markers

## Detailed Requirements

1. `lowerIfStmt(nodeIdx uint32)`:
   - Evaluate condition → %cond
   - Create then_block, else_block (if else exists), merge_block
   - Emit `OpBranch %cond, then_block, else_block` in current block
   - SwitchTo then_block, lowerBlock(then_body)
   - Emit `OpJump merge_block`
   - SwitchTo else_block, lowerBlock(else_body) if exists, else just OpJump merge_block
   - SwitchTo merge_block, emit phi for any value defined in both branches
2. `lowerWhileStmt(nodeIdx uint32)`:
   - Create header_block (condition), body_block, exit_block
   - Emit OpJump header_block from current
   - SwitchTo header_block: evaluate condition, `OpBranch %cond, body_block, exit_block`
   - SwitchTo body_block: emit `OpLoopBegin`, lowerBlock(body), `OpLoopEnd`, `OpJump header_block`
   - SwitchTo exit_block (continue after loop)
3. `lowerForStmt(nodeIdx uint32)`:
   - Lower iterable expression → %iter
   - Create init_block (create iterator), header_block (has_next?), body_block, exit_block
   - Pattern: call iter.next() → Option[T]; branch on Some/None
   - Emit `OpLoopBegin` at body start, `OpLoopEnd` at body end
4. `lowerMatchStmt(nodeIdx uint32)`:
   - Lower subject → %subject
   - Create arm_blocks: one per match arm
   - Emit tag extraction: `%tag = load %subject.tag`
   - Emit decision tree: chain of comparisons → jumps to arm blocks
   - Each arm block: bind payload variable, lower arm body, OpJump merge_block
   - merge_block: phi for any arm result value
5. Phi placement: when multiple blocks define the same symbol, add phi at merge points.

## Implementation Steps

1. Create `ir/air/builder/control.go`.
2. Implement `lowerIfStmt()` with phi placement.
3. Implement `lowerWhileStmt()` with loop markers.
4. Implement `lowerForStmt()` with iterator protocol.
5. Implement `lowerMatchStmt()` with tag extraction and decision tree.
6. Implement `emitPhi(sym uint32, preds []uint32) uint32` helper.
7. Write unit tests for each control flow kind.

## Test Plan

- `TestLowerIf`: simple if without else → then_block + merge_block
- `TestLowerIfElse`: if/else → then_block + else_block + merge_block + phi
- `TestLowerWhile`: while loop → header + body + exit + loop markers
- `TestLowerFor`: for in range → iterator blocks
- `TestLowerMatch`: 3-arm match → 3 arm blocks + merge + phi
- `TestLowerNestedIf`: if inside while → correct nested block structure
- Golden test: fibonacci AIR matches expected CFG structure

## Validation Checklist

- [ ] All blocks have terminators (verifier passes)
- [ ] Loop markers correctly nested (OpLoopBegin matched by OpLoopEnd)
- [ ] Phi nodes at all merge points
- [ ] Phi operand count matches predecessor count
- [ ] break/continue emit correct jumps (to exit_block/header_block)

## Acceptance Criteria

- Fibonacci function produces correct CFG (verified by AIR printer + golden test)
- AIR verifier produces 0 errors on all control flow patterns

## Definition of Done

- [ ] `ir/air/builder/control.go` implemented
- [ ] Unit tests pass
- [ ] AIR verifier passes on all lowered control flow

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Phi placement incorrect for complex nested control flow | Use standard SSA construction algorithm (Braun et al. or dominance frontier) |
| Break/continue target resolution when deeply nested | Maintain break_target/continue_target stack in builder state |

## Future Follow-up Tasks

- p10-t07: opt-loop-region uses OpLoopBegin/OpLoopEnd for loop detection
- p09-t09: async builder splits at await points (similar to block splitting)
