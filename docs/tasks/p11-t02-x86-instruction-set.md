# p11-t02: x86-64 Instruction Set Encoding Tables

## Purpose
Define x86-64 instruction encoding data for all instructions needed by the AXIOM native backend, enabling the machine code emitter to produce correct byte sequences.

## Context
x86-64 instruction encoding is complex: variable-length instructions, REX prefixes for 64-bit operands, ModRM/SIB bytes for memory addressing, VEX prefixes for AVX. Rather than implementing a general-purpose encoder, we define encoding tables for the specific instructions AXIOM needs.

## Inputs
- Intel 64-bit Software Developer's Manual (encoding reference)
- List of opcodes needed from p11-t03 (instruction selector)

## Outputs
- `codegen/native/x86/encoding.go` — instruction encoding tables and emit functions
- `codegen/native/x86/regs.go` — register definitions

## Dependencies
- p11-t01: target-triple — x86_64 target detected

## Subsystems Affected
- Machine code emitter (p11-t10): uses encoding tables
- Instruction selector (p11-t03): references instruction names

## Detailed Requirements

Registers:
```go
type X86Reg uint8
const (
    RAX X86Reg = 0; RCX=1; RDX=2; RBX=3; RSP=4; RBP=5; RSI=6; RDI=7
    R8=8; R9=9; R10=10; R11=11; R12=12; R13=13; R14=14; R15=15
    XMM0 X86Reg = 16; XMM1=17 // ... XMM15=31
)
```

Key instruction encoders (implemented as functions):
- `EncodeMovRR(dst, src X86Reg)` → REX.W + 0x89 + ModRM(3, src, dst)
- `EncodeMovRI(dst X86Reg, imm int64)` → REX.W + 0xB8+rd + imm64
- `EncodeMovMR(dst X86Reg, base X86Reg, disp int32)` → load from memory
- `EncodeAdd(dst, src X86Reg)` → REX.W + 0x01 + ModRM(3, src, dst)
- `EncodeSub(dst, src X86Reg)`, `EncodeMul`, `EncodeImul`, `EncodeIDiv`
- `EncodeCmp(dst, src X86Reg)` → REX.W + 0x3B + ModRM
- `EncodeJcc(condition, offset int32)` → 0x0F + 0x8X + rel32
- `EncodeJmp(offset int32)` → 0xE9 + rel32
- `EncodeCall(offset int32)` → 0xE8 + rel32
- `EncodeRet()` → 0xC3
- `EncodePush(reg X86Reg)` → 0x50+rd (REX if R8+)
- `EncodePop(reg X86Reg)` → 0x58+rd
- `EncodeLea(dst X86Reg, base X86Reg, disp int32)` → REX.W + 0x8D + ModRM+SIB+disp
- `EncodeVaddps(dst, src1, src2 X86Reg)` → VEX.256.0F.W0 + 0x58 (AVX2)
- `EncodeVmovdqu(dst X86Reg, base X86Reg, disp int32)` → load 256-bit SIMD

Helper functions:
- `makeREX(w, r, x, b bool) byte`
- `makeModRM(mod, reg, rm byte) byte`
- `makeSIB(scale, index, base byte) byte`

## Implementation Steps

1. Create `codegen/native/x86/regs.go` with register constants.
2. Create `codegen/native/x86/encoding.go` with encoder functions.
3. Implement each encoder function, emit bytes into `[]byte` buffer.
4. Test each encoder against known-correct byte sequences (from Intel manual or objdump).
5. Write `TestEncodeAdd`: verify byte output matches expected.

## Test Plan
- For each instruction: encode with specific registers, verify output bytes match expected (from `objdump -d` reference)
- `TestEncodeMovRI`: `MOV RAX, 42` → `48 B8 2A 00 00 00 00 00 00 00`
- `TestEncodeAddRR`: `ADD RAX, RBX` → `48 01 D8`
- `TestEncodeJmp`: forward jump with offset 100 → `E9 64 00 00 00`

## Validation Checklist
- [ ] Each encoder produces correct bytes (verified against objdump)
- [ ] REX prefix emitted for R8-R15 and 64-bit operands
- [ ] ModRM byte correct for all operand combinations
- [ ] VEX prefix correct for AVX2 instructions

## Acceptance Criteria
- Generated machine code passes `objdump -d` disassembly and matches expected assembly
- All 30+ instruction encoders produce correct bytes

## Definition of Done
- [ ] `codegen/native/x86/encoding.go` implemented
- [ ] Unit tests verify byte-level correctness
- [ ] Each instruction tested against objdump reference

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| REX prefix errors cause incorrect 32-bit vs 64-bit behavior | Test both with and without REX; verify register size in each test |

## Future Follow-up Tasks
- p11-t10: x86-machine-code-emitter uses these encoders
