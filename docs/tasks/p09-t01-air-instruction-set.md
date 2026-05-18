# p09-t01: AIR Instruction Set Definition

## Purpose
Define the complete Axiom Intermediate Representation (AIR) instruction set — the fixed-width, SSA-form IR that bridges the semantic analysis frontend and the optimization/codegen backends. The instruction set is the contract between all passes.

## Context
AIR is a 3-address, SSA-form IR with 16 bytes per instruction. Unlike LLVM IR (which is text-based and pointer-heavy), AIR uses fixed-width structs in a flat array for cache-friendly traversal. Region-based control flow (`loop_region`) preserves high-level structure for optimizer passes. The instruction set must be stable before any optimization passes or backends are written.

## Inputs
- Spec: `AXIOM SPECIFICATION/05. IR thật sự.md`
- `compiler/ast/node.go` — NodeKinds to map to opcodes
- `compiler/types/types.go` — TypeIDs used in instructions

## Outputs
- `ir/air/opcodes.go` — AirOpcode enum with all opcodes
- `ir/air/inst.go` — AirInst struct (frozen 16-byte layout)
- Mnemonic strings for each opcode

## Dependencies
- p01-t03: struct-layout-definitions — AirInst already defined, extended here

## Subsystems Affected
- AIR builder (p09-t06 through p09-t10): emits these instructions
- AIR verifier (p09-t04): validates instruction invariants
- AIR printer (p09-t05): prints instructions using mnemonics
- Optimization passes (Phase 10): transform these instructions
- Native backend (Phase 11): lowers these to machine instructions

## Detailed Requirements

AirInst layout (FROZEN — 16 bytes):
```go
type AirInst struct {
    Opcode  uint16
    TypeID  uint16  // result type (0 for void instructions)
    Dest    uint32  // destination virtual register (0 for void)
    Src1    uint32  // first source operand
    Src2    uint32  // second source operand
}
```

Instruction classes with opcodes:

**Memory (0x01xx)**:
- `OpAlloc = 0x0101` — allocate heap object: Dest=ptr, Src1=TypeID, Src2=0
- `OpFree = 0x0102` — free heap object: Dest=0, Src1=ptr, Src2=0
- `OpLoad = 0x0103` — load from memory: Dest=val, Src1=addr, Src2=offset
- `OpStore = 0x0104` — store to memory: Dest=0, Src1=addr, Src2=val
- `OpGEP = 0x0105` — get element pointer: Dest=ptr, Src1=base_ptr, Src2=field_idx
- `OpCopy = 0x0106` — copy value: Dest=dst, Src1=src, Src2=size
- `OpMove = 0x0107` — move (invalidate src): Dest=dst, Src1=src
- `OpMakeRef = 0x0108` — create generational ref: Dest=AxRef, Src1=ptr
- `OpDeref = 0x0109` — dereference with gen check: Dest=ptr, Src1=AxRef
- `OpArenaAlloc = 0x010A` — arena allocate: Dest=ptr, Src1=arena, Src2=TypeID
- `OpDestroy = 0x010B` — CTGC destroy: Dest=0, Src1=ptr
- `OpAliasReuse = 0x010C` — reuse memory (alias): Dest=new_ptr, Src1=old_ptr

**ALU (0x02xx)**:
- `OpIConst = 0x0201` — integer constant: Dest=reg, Src1=value_lo, Src2=value_hi
- `OpFConst = 0x0202` — float constant: Dest=reg, Src1=bits_lo, Src2=bits_hi
- `OpIAdd = 0x0203`, `OpISub = 0x0204`, `OpIMul = 0x0205`, `OpIDiv = 0x0206`, `OpIMod = 0x0207`
- `OpFAdd = 0x0208`, `OpFSub = 0x0209`, `OpFMul = 0x020A`, `OpFDiv = 0x020B`
- `OpIPow = 0x020C`, `OpFPow = 0x020D`
- `OpICmpEq = 0x020E`, `OpICmpNe = 0x020F`, `OpICmpLt = 0x0210`, `OpICmpGt = 0x0211`, `OpICmpLe = 0x0212`, `OpICmpGe = 0x0213`
- `OpFCmpEq = 0x0214`, `OpFCmpLt = 0x0215`, `OpFCmpGt = 0x0216`
- `OpBAnd = 0x0217`, `OpBOr = 0x0218`, `OpBXor = 0x0219`, `OpBNot = 0x021A`, `OpBShl = 0x021B`, `OpBShr = 0x021C`
- `OpNeg = 0x021D` — arithmetic negate
- `OpIToF = 0x021E`, `OpFToI = 0x021F` — type casts
- `OpZExt = 0x0220`, `OpSExt = 0x0221`, `OpTrunc = 0x0222` — integer width casts

**Control (0x03xx)**:
- `OpJump = 0x0301` — unconditional jump: Dest=0, Src1=target_block_id
- `OpBranch = 0x0302` — conditional branch: Dest=0, Src1=cond, Src2=then_block | (else_block << 16) [using extra table]
- `OpCall = 0x0303` — function call: Dest=result, Src1=func_sym_id, Src2=args_start_idx [args in extra array]
- `OpReturn = 0x0304` — return: Dest=0, Src1=val (0 if void)
- `OpPhi = 0x0305` — SSA phi: Dest=result, Src1/Src2=incoming regs [extra: incoming block IDs]
- `OpLoopBegin = 0x0306` — loop region start marker
- `OpLoopEnd = 0x0307` — loop region end marker
- `OpSpawn = 0x0308` — spawn actor: Dest=ActorRef, Src1=func_sym, Src2=arg
- `OpSend = 0x0309` — send message: Dest=0, Src1=actor_ref, Src2=msg_ptr
- `OpRecv = 0x030A` — receive message: Dest=msg_ptr, Src1=0

**SIMD (0x04xx)**:
- `OpSIMDLoad = 0x0401`, `OpSIMDStore = 0x0402`
- `OpSIMDAdd = 0x0403`, `OpSIMDMul = 0x0404`, `OpSIMDFMA = 0x0405`

**Comptime (0x05xx)**:
- `OpComptime = 0x0501` — marks subgraph as compile-time evaluable

## Implementation Steps

1. Create `ir/air/opcodes.go` with all opcode constants.
2. Extend `ir/air/inst.go` (from p01-t03) — add mnemonic strings.
3. Add `AirOpcode.Mnemonic() string` method for printer.
4. Add `AirOpcode.IsTerminator() bool` — true for OpJump, OpBranch, OpReturn.
5. Add `AirOpcode.IsBinaryALU() bool`, `IsMemory() bool`, `IsControl() bool`.
6. Write `TestOpcodeCompleteness`: verify all opcodes have mnemonics.
7. Write `TestInstSize`: verify `unsafe.Sizeof(AirInst{}) == 16`.

## Test Plan

- `TestInstSize`: AirInst is exactly 16 bytes
- `TestOpcodeCount`: total opcodes ≤ 0xFFFF (fits in uint16)
- `TestOpcodeUniqueness`: no two opcodes have same value
- `TestMnemonics`: every opcode has non-empty mnemonic string

## Validation Checklist

- [ ] AirInst is exactly 16 bytes (verified with unsafe.Sizeof)
- [ ] All opcodes have unique uint16 values
- [ ] All opcodes have mnemonic strings
- [ ] IsTerminator() correct for jump/branch/return
- [ ] Opcode classes organized by prefix (0x01xx = memory, etc.)

## Acceptance Criteria

- `unsafe.Sizeof(AirInst{}) == 16` (test enforced)
- All 50+ opcodes defined with mnemonics

## Definition of Done

- [ ] `ir/air/opcodes.go` created
- [ ] `ir/air/inst.go` complete
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Need to add opcodes later, breaking existing passes | Use opcode class prefixes (0x01xx, 0x02xx) to leave room |
| Src1/Src2 not enough for multi-arg calls | Use extra array indexed by Src2 for variadic args |

## Future Follow-up Tasks

- p09-t02: air-basic-blocks uses these opcodes
- p09-t04: air-verifier validates these instructions
- p11-t03: x86-instruction-selector maps these to x86 instructions
