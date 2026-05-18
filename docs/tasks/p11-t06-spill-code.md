# p11-t06: Spill Code Generation

## Purpose
Insert load/store instructions for spilled virtual registers, converting spill slots (assigned by register allocator) to actual stack memory accesses.

## Context
When a virtual register is spilled, its value must be stored to a stack slot after each definition and loaded from the stack slot before each use. Efficient spill code minimizes the number of loads/stores (coalescing non-overlapping spills into the same slot).

## Inputs
- `VRegMap` from register allocator (p11-t05) — which VRegs are spilled and their slot index
- `[][]MachInst` — machine instruction list to modify
- Stack frame layout (p11-t07) — spill slot addresses

## Outputs
- `codegen/native/x86/spill.go` — spill code inserter
- Modified `[][]MachInst` with spill stores/loads inserted

## Dependencies
- p11-t05: linear-scan-regalloc — VRegMap with spill decisions
- p11-t07: x86-stack-frame — spill slot addresses (RBP + offset)

## Subsystems Affected
- Machine instruction stream: spill loads/stores inserted
- Stack frame: spill slots contribute to frame size

## Detailed Requirements

1. For each spilled VReg (IsSpill=true in VRegMap):
   - After each instruction that DEFINES the VReg: insert `MOV [rbp - spillOffset], physReg` (store to spill slot)
   - Before each instruction that USES the VReg: insert `MOV physReg, [rbp - spillOffset]` (load from spill slot)
2. Spill slot size: 8 bytes (all values stored as 64-bit in spill slot, regardless of actual type).
3. Spill offset: `rbp - 8 - 8*spillIdx` (below saved rbp, one slot per spill index).
4. Use a fresh temporary physical register for spill load/stores (usually R10, R11 — caller-saved scratch).
5. Coalescing: if two spilled VRegs have non-overlapping live intervals, they can share the same spill slot (reduces frame size).
6. Update stack frame total spill size: `frame.SpillBytes = totalSpillSlots * 8`.

## Implementation Steps

1. Create `codegen/native/x86/spill.go`.
2. Implement `InsertSpillCode(blocks [][]MachInst, vrmap VRegMap, frame *StackFrame)`.
3. Scan each instruction for uses/defs of spilled VRegs.
4. Insert load before uses, store after defs.
5. Implement coalescing (optional, O2+): compute non-overlapping spill slot assignment.
6. Write unit tests.

## Test Plan
- `TestSpillInsert`: spilled VReg → store after def + load before use
- `TestSpillCoalesce`: two non-overlapping spills → share same stack slot
- `TestSpillNoCoalesce`: overlapping spills → separate slots
- `TestSpillCorrect`: compiled program with forced spill runs correctly

## Validation Checklist
- [ ] Store inserted after each def of spilled VReg
- [ ] Load inserted before each use of spilled VReg
- [ ] Coalesced spills use same RBP offset
- [ ] Scratch registers (R10/R11) used for spill moves
- [ ] Frame size accounts for all spill slots

## Acceptance Criteria
- Spilled program compiles and runs correctly
- Functions with > 16 live values correctly spill extras to stack

## Definition of Done
- [ ] `codegen/native/x86/spill.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Scratch register used for spill is itself live | R10/R11 reserved as scratch (not allocatable by regalloc) |

## Future Follow-up Tasks
- p11-t07: stack frame finalizes layout with spill slots included
