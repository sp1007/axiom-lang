# p11-t07: x86-64 Stack Frame Layout

## Purpose
Implement stack frame layout computation and prologue/epilogue emission for x86-64 functions, ensuring correct frame structure for debugging, unwinding, and ABI compliance.

## Context
The x86-64 stack frame must be laid out precisely: saved RBP at the top, callee-saved registers below, spill slots below those, local allocations at the bottom, with 16-byte alignment maintained at all CALL sites.

## Inputs
- `VRegMap` — which physical registers are callee-saved and need saving
- Spill slot count from spill code generation
- Local variable sizes from VarDecl escape analysis

## Outputs
- `codegen/native/x86/frame.go` — StackFrame computation
- Prologue/epilogue MachInst sequences

## Dependencies
- p11-t06: spill-code — spill slot count known
- p11-t08: x86-abi-sysv — callee-saved register list

## Subsystems Affected
- Machine code emitter: prologue/epilogue emitted at function entry/exit
- Stack frame: RBP-relative addresses for all locals and spill slots

## Detailed Requirements

```go
type StackFrame struct {
    CalleeSaved    []X86Reg  // registers to save/restore
    SpillSlots     int       // number of spill slots (each 8 bytes)
    LocalBytes     int       // stack space for local allocations
    AlignPadding   int       // padding to achieve 16-byte alignment
    TotalSize      int       // callee_saved*8 + spill*8 + local + padding
}
```

Layout (high → low address):
```
[old RBP]          ← RBP points here after prologue
[callee-saved regs] (8 bytes each, in order)
[spill slots]      (8 bytes each)
[local allocs]     (variable size)
[alignment pad]
[return addr]      ← RSP on function entry
```

Prologue sequence:
```asm
PUSH RBP
MOV RBP, RSP
PUSH callee_saved_reg1
PUSH callee_saved_reg2
...
SUB RSP, (spill_slots*8 + local_bytes + align_padding)
```

Epilogue (before RET):
```asm
MOV RSP, RBP  ; or ADD RSP, frame_size
POP callee_saved_regN  ; in reverse order
...
POP RBP
RET
```

16-byte alignment rule: `(frame_size + 8) % 16 == 0` (the +8 is the return address on stack).

## Implementation Steps

1. Create `codegen/native/x86/frame.go`.
2. Implement `ComputeFrame(calleeSaved []X86Reg, spillCount, localBytes int) StackFrame`.
3. Implement `EmitPrologue(frame StackFrame) []MachInst`.
4. Implement `EmitEpilogue(frame StackFrame) []MachInst`.
5. Compute spill slot addresses: `rbpOffset(slotIdx) = -8 - 8*slotIdx - len(calleeSaved)*8`.
6. Write tests.

## Test Plan
- `TestFrameNoSpill`: 0 callee-saved, 0 spills → frame size = 0 (just push/pop RBP)
- `TestFrameWithSpills`: 3 spill slots → frame_size = 24 + alignment
- `TestFrameAlignment`: verify frame maintains 16-byte RSP alignment before CALL
- `TestFrameCalleeSaved`: R12 used → saved in prologue, restored in epilogue

## Validation Checklist
- [ ] 16-byte alignment maintained
- [ ] Callee-saved registers saved/restored in correct order (reverse in epilogue)
- [ ] Spill slot addresses are valid RBP-relative offsets
- [ ] Frame size accounts for all components

## Acceptance Criteria
- Function with 5 spills produces correct prologue reducing RSP by 40 bytes
- `gdb` can unwind through AXIOM stack frames

## Definition of Done
- [ ] `codegen/native/x86/frame.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Off-by-8 alignment error (forget return address in RSP calculation) | Add explicit test for alignment at CALL boundary |

## Future Follow-up Tasks
- p11-t10: machine code emitter emits prologue/epilogue bytes
- p11-t13: dwarf-line-info emits frame info for debugging
