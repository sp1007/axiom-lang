# p11-t17: ModRM/SIB Encoding Library

## Purpose
Implement a reusable x86-64 ModRM and SIB byte encoding library used by the machine code emitter. ModRM/SIB encoding is complex (memory addressing modes, register extensions, displacement sizes) and must be correct for every instruction. Centralizing this logic avoids bugs in p11-t10's emitter.

## Context
Every x86-64 instruction that references memory or registers uses ModRM (and optionally SIB) bytes. The encoding depends on: addressing mode (register-direct, [reg], [reg+disp8], [reg+disp32], [RIP+disp32], [base+index*scale+disp]), register numbers (0-15, with REX extensions), and special cases (RSP→SIB required, RBP→disp8 required).

## Inputs
- x86-64 instruction set definitions from p11-t02
- Register numbering (0=RAX, 1=RCX, ..., 15=R15)

## Outputs
- `codegen/native/x86/modrm.go` — ModRM/SIB encoding functions
- `codegen/native/x86/modrm_test.go` — exhaustive encoding tests

## Dependencies
- p11-t02: x86-instruction-set — register definitions

## Subsystems Affected
- x86 machine code emitter (p11-t10): uses this library for all memory/register operands

## Detailed Requirements

### 1. ModRM Byte Layout
```
Bits: [7:6] mod | [5:3] reg | [2:0] rm
mod=00: [rm]           (no displacement, except rm=101 → [RIP+disp32])
mod=01: [rm+disp8]
mod=10: [rm+disp32]
mod=11: rm (register direct)
rm=100: SIB byte follows
rm=101 (mod=00): [RIP+disp32] (x86-64 RIP-relative)
```

### 2. SIB Byte Layout
```
Bits: [7:6] scale | [5:3] index | [2:0] base
scale: 0=1, 1=2, 2=4, 3=8
index=100: no index (just base)
base=101 (mod=00): disp32 only (no base)
```

### 3. API

```go
// EncodeModRM returns the ModRM byte for register-to-register.
func EncodeModRM_RR(reg, rm PhysReg) byte

// EncodeModRM_RM returns ModRM (+ optional SIB + displacement) for memory operand.
func EncodeModRM_RM(reg PhysReg, base PhysReg, disp int32) []byte

// EncodeModRM_RIP returns ModRM for RIP-relative addressing.
func EncodeModRM_RIP(reg PhysReg, disp32 int32) []byte

// EncodeModRM_SIB returns ModRM+SIB for scaled index addressing.
func EncodeModRM_SIB(reg, base, index PhysReg, scale uint8, disp int32) []byte

// NeedsREX returns true if any operand requires REX prefix (R8-R15, 64-bit ops).
func NeedsREX(regs ...PhysReg) bool

// EncodeREX returns the REX prefix byte.
func EncodeREX(w, r, x, b bool) byte
```

### 4. Special Cases

- `RSP` (reg 4) as base → SIB byte always required
- `RBP` (reg 5) as base with no displacement → use `[RBP+0]` (mod=01, disp8=0)
- `R13` (reg 13) same rules as RBP
- `R12` (reg 12) same rules as RSP

## Implementation Steps

1. Create `codegen/native/x86/modrm.go`.
2. Implement `EncodeModRM_RR` — simple mod=11 encoding.
3. Implement `EncodeModRM_RM` with displacement size selection (0, 8, 32 bits).
4. Handle RSP/R12 special case (emit SIB with index=RSP).
5. Handle RBP/R13 special case (force disp8=0 when no displacement).
6. Implement `EncodeModRM_RIP` for RIP-relative.
7. Implement `EncodeModRM_SIB` for scaled index addressing.
8. Implement REX prefix encoding.
9. Write exhaustive tests for all register combinations.

## Test Plan

- `TestModRM_RR_AllRegs`: all 16×16 register pairs produce correct bytes
- `TestModRM_RM_NoDisp`: `[RAX]` → ModRM=0x00
- `TestModRM_RM_Disp8`: `[RAX+4]` → ModRM=0x40, disp=0x04
- `TestModRM_RM_Disp32`: `[RAX+1000]` → ModRM=0x80, disp=0xE8030000
- `TestModRM_RM_RSP`: `[RSP]` → ModRM+SIB (SIB=0x24)
- `TestModRM_RM_RBP`: `[RBP]` → ModRM=0x45, disp8=0x00
- `TestModRM_RIP`: RIP-relative → ModRM=0x05+reg
- `TestREX_R8`: R8-R15 set correct REX bits
- `TestModRM_SIB_Scale`: `[RAX+RCX*4]` → correct SIB byte

## Validation Checklist

- [ ] All 16 registers encode correctly without REX
- [ ] R8-R15 triggers REX prefix
- [ ] RSP base always emits SIB
- [ ] RBP base with no displacement uses disp8=0
- [ ] Displacement size selection is minimal (0 < disp8 < disp32)
- [ ] RIP-relative encoding correct

## Acceptance Criteria

- All 16×16 register pair encodings verified against Intel manual

## Definition of Done

- [ ] `codegen/native/x86/modrm.go` implemented
- [ ] All tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Encoding bugs cause silent wrong code | Exhaustive test matrix for all register combinations |
| 32-bit vs 64-bit operand size confusion | Always explicit about operand size in API |

## Future Follow-up Tasks

- p11-t10: Machine code emitter uses this library for all instructions
- p11-t03: Instruction selector uses addressing mode types that map to these functions
