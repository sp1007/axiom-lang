# p11-t05: Linear Scan Register Allocation

## Purpose
Implement Linear Scan Register Allocation to map virtual registers to physical x86-64 registers (or spill slots), completing the register assignment step of native code generation.

## Context
Linear Scan is an O(n log n) register allocation algorithm — fast enough for JIT compilers and used by Java HotSpot. It processes live intervals sorted by start position, maintaining an "active" set of currently-live intervals, assigning registers greedily. When no register is free, it spills the interval with the furthest end point.

## Inputs
- `[]LiveInterval` sorted by start (from p11-t04)
- Available register sets from p11-t08 (ABI-defined caller/callee-saved)
- LoopDepth per block (for spill cost estimation)

## Outputs
- `codegen/native/x86/regalloc.go` — LinearScanAllocator
- `VRegMap map[uint32]PhysOrSpill` (VReg → PhysReg or SpillSlot)

## Dependencies
- p11-t04: liveness-analysis — sorted LiveInterval list
- p11-t01: target-triple — architecture determines available registers

## Subsystems Affected
- Spill code (p11-t06): uses VRegMap to insert spills
- Machine code emitter (p11-t10): replaces VRegs with PhysRegs

## Detailed Requirements

```go
type PhysOrSpill struct {
    IsSpill  bool
    PhysReg  X86Reg      // if !IsSpill
    SpillIdx int         // if IsSpill, index into spill slot array
}

type LinearScanAllocator struct {
    Available  []X86Reg    // available integer registers
    FAvailable []X86Reg    // available float registers
    Active     []*LiveInterval  // sorted by End
    VRegMap    map[uint32]PhysOrSpill
}
```

Algorithm:
1. Sort intervals by Start.
2. For each interval I:
   a. Expire old intervals: remove intervals from Active where End < I.Start; free their registers.
   b. If no register available: spill. Spill the active interval with the largest End:
      - If that interval's End > I.End: spill that interval, give its register to I.
      - Else: spill I itself (assign to SpillSlot).
   c. Else: assign a free register to I, add to Active (sorted by End).
3. Spill slots: each spill gets an index; total spill count determines stack frame size.
4. Caller-saved vs callee-saved: prefer caller-saved for short-lived intervals (no save/restore), callee-saved for long-lived (must save in prologue).

## Implementation Steps

1. Create `codegen/native/x86/regalloc.go`.
2. Implement `Allocate(intervals []LiveInterval, target Target) VRegMap`.
3. Maintain Active as a sorted slice (insertion sort is fine for ≤16 active intervals).
4. Implement expire logic.
5. Implement spill selection (furthest End).
6. Write unit tests.

## Test Plan
- `TestAllocBasic`: 3 intervals, 2 registers → 2 assigned, 1 spilled
- `TestAllocNoSpill`: 16 intervals, 16 registers → all assigned, no spills
- `TestAllocSpillFurthest`: when spilling, the interval with furthest end is spilled
- `TestAllocCalleeSaved`: long-lived interval → prefers callee-saved register

## Validation Checklist
- [ ] Each VReg assigned exactly one PhysReg or SpillSlot
- [ ] Intervals that expire before overlap are correctly freed
- [ ] Spill selects furthest-end interval
- [ ] ABI constraints respected (RAX, RDX not callee-saved)

## Acceptance Criteria
- Simple functions (< 16 VRegs) produce no spills
- Spill count for a complex function is minimized

## Definition of Done
- [ ] `codegen/native/x86/regalloc.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| ABI-reserved registers accidentally allocated | Mark RAX, RDX, RSP, RBP as not available by default |

## Future Follow-up Tasks
- p11-t06: spill-code-generation inserts load/store for spilled registers
- p11-t10: emitter uses VRegMap to emit physical register encodings
