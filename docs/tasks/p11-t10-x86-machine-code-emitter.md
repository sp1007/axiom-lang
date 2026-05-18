# p11-t10: x86-64 Machine Code Emitter

## Purpose
Convert MachInst (machine instruction IR) to raw x86-64 binary bytes, producing the `.text` section content for ELF/PE-COFF/Mach-O object files.

## Context
The machine code emitter is the final step before object file emission. It encodes each MachInst into its x86-64 binary representation using the encoding tables from p11-t02, applying VRegMap from p11-t05 to resolve virtual registers to physical registers or RBP-relative stack addresses.

## Inputs
- `[]MachInst` per function (after register allocation and spill code)
- `VRegMap` from p11-t05 (VReg → PhysReg or SpillSlot)
- `StackFrame` from p11-t07 (for RBP offsets)
- Relocation targets (call targets, global refs) for back-patching

## Outputs
- `codegen/native/x86/emitter.go` — MachInst → byte sequence
- `[]byte` — raw machine code for `.text` section
- `[]Relocation` — unresolved references for p11-t11

## Dependencies
- p11-t02: x86-instruction-set — encoding tables
- p11-t05: linear-scan-regalloc — VRegMap
- p11-t07: x86-stack-frame — RBP offsets, prologue/epilogue bytes
- p11-t08/t09: ABI — call setup/teardown sequences

## Subsystems Affected
- Object file emitter (p11-t12): consumes raw `.text` bytes
- Relocation back-patcher (p11-t11): fixes up forward references

## Detailed Requirements

```go
type Emitter struct {
    Buf       []byte
    Relocs    []Relocation
    LabelMap  map[uint32]int  // block ID → byte offset in Buf
    Fixups    []Fixup         // {offset, targetBlockID} for branch targets
}

type Relocation struct {
    Offset   int     // byte offset in Buf needing fixup
    Symbol   string  // external symbol name
    Type     RelocType  // R_X86_64_PC32, R_X86_64_PLT32, etc.
    Addend   int32
}

type Fixup struct {
    PatchOffset int    // where to write the 32-bit relative offset
    TargetBlock uint32
}

func (e *Emitter) EmitFunc(f *MachFunc, vrmap VRegMap, frame StackFrame) ([]byte, []Relocation)
func (e *Emitter) emitInst(inst MachInst, vrmap VRegMap, frame StackFrame)
func (e *Emitter) emitPrologue(frame StackFrame)
func (e *Emitter) emitEpilogue(frame StackFrame)
func (e *Emitter) resolveFixups()
```

Encoding dispatch:
```go
switch inst.Op {
case MOV_RR: e.encodeMovRR(resolveReg(inst.Dst), resolveReg(inst.Src1))
case MOV_RI: e.encodeMovRI(resolveReg(inst.Dst), inst.Imm)
case MOV_RM: e.encodeMovRM(resolveReg(inst.Dst), rbpOffset(vrmap, inst.Src1))
case MOV_MR: e.encodeMovMR(rbpOffset(vrmap, inst.Dst), resolveReg(inst.Src1))
case ADD_RR: e.encodeAddRR(resolveReg(inst.Dst), resolveReg(inst.Src1))
case CALL:   e.encodeCall(inst.Target); e.addReloc(inst.Target)
case RET:    e.emit(0xC3)
case JMP:    e.encodeJmp(inst.Target); e.addFixup(inst.Target)
case JCC:    e.encodeJcc(inst.Cond, inst.Target); e.addFixup(inst.Target)
}
```

REX prefix handling:
- REX.W (0x48) for 64-bit operands
- REX.R for extended destination (R8-R15)
- REX.B for extended source (R8-R15)
- REX.X for extended SIB index

Branch fixup: emit 0x00000000 placeholder, record fixup; after all blocks emitted, patch relative offsets.

## Implementation Steps

1. Create `codegen/native/x86/emitter.go`.
2. Implement `EmitFunc()` — iterate blocks in RPO order, call emitInst per instruction.
3. Emit prologue before first block, epilogue before RET.
4. Implement register resolution: VReg → PhysReg encoding number (0=RAX, 1=RCX, ..., 7=RDI, 8=R8-15).
5. Implement spill slot resolution: VReg → `[rbp - offset]` ModRM encoding.
6. Implement REX prefix calculation for each instruction.
7. Implement branch fixup pass: after all blocks emitted, resolve intra-function jumps.
8. Emit relocations for CALL to external symbols.
9. Write unit tests comparing output bytes to expected encoding.

## Test Plan
- `TestEmitMovRR`: `MOV RAX, RBX` → bytes `48 89 D8`
- `TestEmitMovRI`: `MOV RAX, 42` → bytes `48 B8 2A 00 00 00 00 00 00 00`
- `TestEmitAddRR`: `ADD RDI, RSI` → bytes `48 01 F7`
- `TestEmitCall`: CALL produces E8 + 4-byte reloc placeholder
- `TestEmitJcc`: `JE label` produces 0F 84 + 4-byte offset (or patched)
- `TestEmitRet`: RET → byte `C3`
- `TestEmitR8Reg`: `MOV R8, R9` → REX prefix present
- `TestBranchFixup`: forward jump patched with correct offset after block layout

## Validation Checklist
- [ ] REX.W set for all 64-bit ops
- [ ] REX.R/REX.B set for R8-R15 operands
- [ ] Branch targets patched after all blocks emitted
- [ ] CALL relocations recorded with correct symbol names
- [ ] Spilled VRegs encoded as [rbp - N] memory operands

## Acceptance Criteria
- `add(a, b: i32) -> i32` function emits correct x86-64 bytes verifiable with `objdump -d`

## Definition of Done
- [ ] `codegen/native/x86/emitter.go` implemented
- [ ] Unit tests pass comparing bytes to hand-encoded expectations

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| REX prefix errors for R8-R15 registers | Encode register numbers 0-15, set REX bits automatically |
| Branch offset overflow (>2GB function) | Panic with "function too large" — not realistic in MVP |

## Future Follow-up Tasks
- p11-t11: relocation back-patcher handles external symbol fixups
- p11-t12: ELF64 emitter wraps .text bytes into object file
