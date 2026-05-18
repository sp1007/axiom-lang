# p10-t07: Loop Region Detection & Annotation

## Purpose
Identify natural loops in the CFG using dominance analysis and back-edge detection, annotate basic blocks with loop depth, and enable loop-invariant code motion as a first loop optimization.

## Context
Loop optimization is critical for performance. Before vectorization (p10-t08) or SoA transforms (p10-t09) can run, loops must be identified and annotated. AXIOM's AIR already has `OpLoopBegin`/`OpLoopEnd` markers from the builder — this pass validates and augments them with precise loop membership information.

## Inputs
- `AirFunc` CFG with OpLoopBegin/OpLoopEnd markers from AIR builder
- Dominator tree (computed from CFG in p09-t02)

## Outputs
- `BasicBlock.LoopDepth` correctly set for all blocks
- `LoopInfo{header, body, exit blocks}` per loop
- Loop-invariant instructions hoisted to loop preheader

## Dependencies
- p10-t01: opt-pipeline-manager — implements OptPass (O2+)
- p09-t02: air-basic-blocks — CFG and dominator computation

## Subsystems Affected
- Loop optimizations (p10-t08, p10-t09): require LoopInfo
- Register allocator (p11-t05): uses LoopDepth for spill cost estimation

## Detailed Requirements

1. `LoopInfo` struct: `{Header uint32, Blocks []uint32, ExitBlocks []uint32, Depth uint8}`
2. Loop detection using back-edges: a back-edge exists from B to A if A dominates B and there's a CFG edge B→A. The natural loop of back-edge (B,A) is the set of nodes that can reach B without going through A.
3. Validate OpLoopBegin/OpLoopEnd markers — warn if they don't match detected back-edges.
4. Set `BasicBlock.LoopDepth`: blocks inside 1 loop have depth=1, inside nested loop depth=2, etc.
5. Loop-Invariant Code Motion (LICM): for each loop, find instructions whose operands are all loop-invariant (defined outside the loop or constants). Hoist them to the loop preheader block.
6. Create preheader block if it doesn't exist: a new block inserted before the loop header with a single edge from all outside predecessors to the preheader, and preheader→header.
7. Return true if any hoisting occurred.

## Implementation Steps

1. Create `ir/opt/loop_region.go`.
2. Implement back-edge detection from dominator tree.
3. Implement natural loop computation (BFS backward from back-edge source).
4. Set LoopDepth on all BasicBlocks.
5. Implement LICM: identify invariant instructions, hoist to preheader.
6. Write tests: `TestLoopDetect`, `TestLoopDepth`, `TestLICM`.

## Test Plan

- `TestSimpleWhileLoop`: single while loop → 1 loop, depth 1
- `TestNestedLoop`: for inside while → 2 loops, inner depth 2
- `TestLICMHoist`: `let x = expensive_pure(); while cond: use(x)` → x hoisted out
- `TestLICMNotHoist`: loop-dependent computation → NOT hoisted
- `TestLoopDepthUnrolled`: block depths correctly assigned for complex nested loops

## Validation Checklist

- [ ] All back-edges detected
- [ ] LoopDepth set correctly (0 = outside all loops)
- [ ] Hoisted instructions compute same value as before
- [ ] Preheader inserted when needed
- [ ] AIR verifier passes after hoisting

## Acceptance Criteria

- Fibonacci's recursive call not hoisted (not loop-invariant)
- Pure constant computation inside loop correctly hoisted to preheader

## Definition of Done

- [ ] `ir/opt/loop_region.go` implemented
- [ ] Registered in O2+ pipeline
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Hoisting instruction with side effects | Only hoist pure instructions (no OpStore, OpCall, OpSpawn, etc.) |
| Incorrect loop membership affects downstream optimizations | Validate against OpLoopBegin/End markers |

## Future Follow-up Tasks

- p10-t08: vectorization uses loop info for SIMD substitution
- p10-t09: SoA transform uses loop info to identify vectorizable field accesses
- p11-t05: register allocator uses LoopDepth for spill cost
