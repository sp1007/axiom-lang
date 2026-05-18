# p11-t04: Liveness Analysis

## Purpose
Compute live intervals for all virtual registers — the ranges over which each register must hold a live value — enabling linear scan register allocation.

## Context
Liveness analysis determines which virtual registers are "alive" at each program point. A register is live if its value will be used in the future. Live intervals are represented as [start, end] instruction indices. The linear scan allocator uses these intervals to assign physical registers with minimal spilling.

## Inputs
- `[][]MachInst` per-block from instruction selector (p11-t03)
- CFG structure (Succs/Preds) from p09-t02

## Outputs
- `codegen/native/x86/liveness.go` — LiveInterval computation
- `[]LiveInterval` sorted by start position for linear scan

## Dependencies
- p11-t03: x86-instruction-selector — MachInst with VRegs
- p09-t02: air-basic-blocks — CFG for dataflow analysis

## Subsystems Affected
- Register allocator (p11-t05): primary consumer of live intervals
- Spill code (p11-t06): uses intervals for spill slot assignment

## Detailed Requirements

```go
type LiveInterval struct {
    VReg    uint32
    Start   uint32  // instruction number (linearized across all blocks)
    End     uint32  // last use instruction number
    Splits  []uint32 // split points for register-splitting optimization
}
```

Algorithm (standard reverse dataflow):
1. Linearize all MachInsts with consecutive instruction numbers (across all blocks in RPO).
2. For each block (in reverse post-order, processed in reverse — from exit to entry):
   - Compute `live_in[B]` = `live_out[B]` minus defs in B plus uses in B
   - `live_out[B]` = union of `live_in[S]` for all successors S
3. From live sets: for each VReg, interval starts at first definition, ends at last use (or end of last live block).
4. Handle loops: intervals that span a loop back-edge must extend to the end of the loop (even if not used in the loop body).
5. Sort intervals by `Start` for linear scan.

## Implementation Steps

1. Create `codegen/native/x86/liveness.go`.
2. Linearize MachInsts: assign consecutive numbers in RPO block order.
3. Compute live-in/live-out per block using iterative dataflow.
4. Build interval set: for each VReg, scan all blocks for first def and last use.
5. Extend intervals across loop back-edges.
6. Sort by start.
7. Write unit tests.

## Test Plan
- `TestLivenessSimple`: function with `%0 = iconst; %1 = iconst; %2 = add %0, %1; return %2` — intervals: %0=[0,2], %1=[1,2], %2=[2,3]
- `TestLivenessAcrossBlock`: value defined in block_0, used in block_2 — interval spans all blocks
- `TestLivenessLoop`: value used in loop → interval extends to loop end

## Validation Checklist
- [ ] Each VReg has exactly one interval (no holes in single-def SSA)
- [ ] Intervals sorted by Start
- [ ] Loop back-edges extend intervals correctly
- [ ] Live-in/live-out consistent at block boundaries

## Acceptance Criteria
- Liveness analysis produces correct intervals for fibonacci function

## Definition of Done
- [ ] `codegen/native/x86/liveness.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Incorrect live-out computation at loop back-edges | Use standard iterative algorithm with loop-aware extension |

## Future Follow-up Tasks
- p11-t05: linear-scan-regalloc uses sorted LiveInterval list
